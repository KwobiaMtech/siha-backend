#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
TEST_EMAIL="stellar_test@example.com"
TEST_PASSWORD="password123"

echo "🌟 Testing Stellar Wallet Integration"
echo "===================================="

# Step 1: Register test user
echo "📝 Step 1: Registering test user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"firstName\": \"Stellar\",
    \"lastName\": \"Test\"
  }")

echo "Register Response: $REGISTER_RESPONSE"

# Step 2: Login to get JWT token
echo -e "\n🔐 Step 2: Logging in..."
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
    echo "❌ Failed to get JWT token"
    exit 1
fi

echo "✅ JWT Token obtained: ${JWT_TOKEN:0:20}..."

# Step 3: Create Stellar wallet
echo -e "\n🏦 Step 3: Creating Stellar wallet..."
WALLET_RESPONSE=$(curl -s -X POST "$BASE_URL/stellar/wallet" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "{
    \"network\": \"testnet\"
  }")

echo "Wallet Response: $WALLET_RESPONSE"

# Step 4: Get wallet details
echo -e "\n📊 Step 4: Getting wallet details..."
WALLET_DETAILS=$(curl -s -X GET "$BASE_URL/stellar/wallet" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Wallet Details: $WALLET_DETAILS"

# Step 5: Get trustlines
echo -e "\n🔗 Step 5: Getting trustlines..."
TRUSTLINES=$(curl -s -X GET "$BASE_URL/stellar/trustlines" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Trustlines: $TRUSTLINES"

# Step 6: Get asset info
echo -e "\n💰 Step 6: Getting USDC asset info..."
ASSET_INFO=$(curl -s -X GET "$BASE_URL/stellar/asset-info?asset=USDC" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Asset Info: $ASSET_INFO"

# Step 7: Test USDC send (will fail without real balance)
echo -e "\n💸 Step 7: Testing USDC send..."
SEND_RESPONSE=$(curl -s -X POST "$BASE_URL/stellar/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "{
    \"toAddress\": \"GDESTINATION123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ\",
    \"amount\": 5.0,
    \"memo\": \"Test payment\"
  }")

echo "Send Response: $SEND_RESPONSE"

# Step 8: Get transaction history
echo -e "\n📜 Step 8: Getting transaction history..."
TRANSACTIONS=$(curl -s -X GET "$BASE_URL/stellar/transactions" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Transactions: $TRANSACTIONS"

echo -e "\n✅ Stellar wallet test completed!"
