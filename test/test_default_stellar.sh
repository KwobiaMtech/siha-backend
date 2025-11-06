#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "🌟 Testing Default Stellar Wallet Creation"
echo "=========================================="

# Create new user to test default wallet creation
TEST_EMAIL="default_stellar_$(date +%s)@example.com"
TEST_PASSWORD="password123"

echo "📝 Step 1: Registering new user (should create Stellar wallet automatically)..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"firstName\": \"Default\",
    \"lastName\": \"Stellar\"
  }")

echo "Register Response: $REGISTER_RESPONSE"

# Extract user ID and verify user
USER_ID=$(echo $REGISTER_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
if [ ! -z "$USER_ID" ]; then
    echo "Verifying user $USER_ID..."
    mongosh "mongodb://user:1v4HU0gkvkScT2n@localhost:27017/healthy_pay?authSource=admin" --eval "
    db.users.updateOne(
      {_id: ObjectId('$USER_ID')}, 
      {\$set: {is_verified: true, kyc_status: 'approved'}}
    );
    print('User verified');
    " > /dev/null
fi

echo -e "\n🔐 Step 2: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

JWT_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$JWT_TOKEN" ]; then
    echo "❌ Failed to get JWT token"
    exit 1
fi

echo "✅ JWT Token obtained"

echo -e "\n📊 Step 3: Getting default wallet (should be Stellar)..."
WALLET_RESPONSE=$(curl -s -X GET "$BASE_URL/wallet" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Wallet Response: $WALLET_RESPONSE"

echo -e "\n🏦 Step 4: Getting Stellar wallet directly..."
STELLAR_WALLET=$(curl -s -X GET "$BASE_URL/stellar/wallet" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Stellar Wallet: $STELLAR_WALLET"

echo -e "\n✅ Default Stellar wallet test completed!"
