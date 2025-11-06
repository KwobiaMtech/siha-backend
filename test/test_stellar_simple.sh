#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "🌟 Testing Stellar Wallet Integration (Direct)"
echo "=============================================="

# Use existing test user credentials
TEST_EMAIL="test@example.com"
TEST_PASSWORD="password123"

# Step 1: Login to get JWT token
echo "🔐 Step 1: Logging in with existing user..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

echo "Login Response: $LOGIN_RESPONSE"

# Extract JWT token
JWT_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$JWT_TOKEN" ]; then
    echo "❌ Failed to get JWT token. Trying to create new user..."
    
    # Register new user
    REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
      -H "Content-Type: application/json" \
      -d "{
        \"email\": \"stellar_new@example.com\",
        \"password\": \"$TEST_PASSWORD\",
        \"firstName\": \"Stellar\",
        \"lastName\": \"User\"
      }")
    
    echo "Register Response: $REGISTER_RESPONSE"
    
    # For testing, let's manually verify the user
    USER_ID=$(echo $REGISTER_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    if [ ! -z "$USER_ID" ]; then
        echo "Manually verifying user $USER_ID..."
        mongosh "mongodb://user:1v4HU0gkvkScT2n@localhost:27017/healthy_pay?authSource=admin" --eval "
        db.users.updateOne(
          {_id: ObjectId('$USER_ID')}, 
          {\$set: {is_verified: true, kyc_status: 'approved'}}
        );
        print('User verified');
        "
        
        # Try login again
        LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
          -H "Content-Type: application/json" \
          -d "{
            \"email\": \"stellar_new@example.com\",
            \"password\": \"$TEST_PASSWORD\"
          }")
        
        JWT_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    fi
fi

if [ -z "$JWT_TOKEN" ]; then
    echo "❌ Still failed to get JWT token"
    exit 1
fi

echo "✅ JWT Token obtained: ${JWT_TOKEN:0:20}..."

# Step 2: Create Stellar wallet
echo -e "\n🏦 Step 2: Creating Stellar wallet..."
WALLET_RESPONSE=$(curl -s -X POST "$BASE_URL/stellar/wallet" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "{
    \"network\": \"testnet\"
  }")

echo "Wallet Response: $WALLET_RESPONSE"

# Step 3: Get wallet details
echo -e "\n📊 Step 3: Getting wallet details..."
WALLET_DETAILS=$(curl -s -X GET "$BASE_URL/stellar/wallet" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Wallet Details: $WALLET_DETAILS"

# Step 4: Get trustlines
echo -e "\n🔗 Step 4: Getting trustlines..."
TRUSTLINES=$(curl -s -X GET "$BASE_URL/stellar/trustlines" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Trustlines: $TRUSTLINES"

# Step 5: Get asset info
echo -e "\n💰 Step 5: Getting USDC asset info..."
ASSET_INFO=$(curl -s -X GET "$BASE_URL/stellar/asset-info?asset=USDC" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Asset Info: $ASSET_INFO"

echo -e "\n✅ Stellar wallet integration test completed!"
