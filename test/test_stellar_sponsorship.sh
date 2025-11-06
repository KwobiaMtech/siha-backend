#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "🌟 Testing Real Stellar Testnet Sponsorship"
echo "==========================================="

# Create unique test user
TEST_EMAIL="stellar_testnet_$(date +%s)@example.com"
TEST_PASSWORD="password123"

echo "📝 Step 1: Creating new user with real Stellar sponsorship..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"firstName\": \"Stellar\",
    \"lastName\": \"Testnet\"
  }")

echo "Register Response: $REGISTER_RESPONSE"

# Extract and verify user
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

echo -e "\n🏦 Step 3: Getting Stellar wallet (should be created with real sponsorship)..."
WALLET_RESPONSE=$(curl -s -X GET "$BASE_URL/stellar/wallet" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Wallet Response: $WALLET_RESPONSE"

# Extract wallet details
PUBLIC_KEY=$(echo $WALLET_RESPONSE | grep -o '"publicKey":"[^"]*' | cut -d'"' -f4)

if [ ! -z "$PUBLIC_KEY" ]; then
    echo -e "\n🔍 Step 4: Verifying account exists on Stellar testnet..."
    echo "Public Key: $PUBLIC_KEY"
    
    # Check account on Stellar testnet using Horizon API
    HORIZON_RESPONSE=$(curl -s "https://horizon-testnet.stellar.org/accounts/$PUBLIC_KEY")
    
    if echo "$HORIZON_RESPONSE" | grep -q "account_id"; then
        echo "✅ Account found on Stellar testnet!"
        
        # Extract account details
        ACCOUNT_ID=$(echo $HORIZON_RESPONSE | grep -o '"account_id":"[^"]*' | cut -d'"' -f4)
        SEQUENCE=$(echo $HORIZON_RESPONSE | grep -o '"sequence":"[^"]*' | cut -d'"' -f4)
        
        echo "Account ID: $ACCOUNT_ID"
        echo "Sequence: $SEQUENCE"
        
        # Check balances
        echo -e "\n💰 Account Balances:"
        echo "$HORIZON_RESPONSE" | grep -o '"balance":"[^"]*' | head -5
        
        # Check if account has USDC trustline
        if echo "$HORIZON_RESPONSE" | grep -q "USDC"; then
            echo "✅ USDC trustline found!"
        else
            echo "⚠️ USDC trustline not found (may still be processing)"
        fi
        
    else
        echo "❌ Account not found on Stellar testnet"
        echo "Error response: $HORIZON_RESPONSE"
    fi
else
    echo "❌ No public key found in wallet response"
fi

echo -e "\n✅ Stellar Testnet Sponsorship Test Completed!"
