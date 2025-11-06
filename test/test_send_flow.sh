#!/bin/bash

echo "🧪 Testing Send Flow Integration"
echo "================================"

BASE_URL="http://localhost:8080/api/v1"

# Step 1: Login to get token
echo -e "\n1️⃣ Logging in..."
login_response=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

echo "Login response: $login_response"
token=$(echo "$login_response" | jq -r '.token // empty')

if [ -z "$token" ]; then
    echo "❌ Login failed, cannot test send flow"
    exit 1
fi

echo "✅ Login successful, token: ${token:0:20}..."

# Step 2: Get payment methods
echo -e "\n2️⃣ Getting payment methods..."
payment_methods_response=$(curl -s -X GET $BASE_URL/send/payment-methods \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token")

echo "Payment methods: $payment_methods_response"
echo "Note: Mobile money methods don't show balance for security reasons"

# Step 3: Get recipients
echo -e "\n3️⃣ Getting recipients..."
recipients_response=$(curl -s -X GET $BASE_URL/send/recipients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token")

echo "Recipients: $recipients_response"

# Step 4: Send money
echo -e "\n4️⃣ Sending money..."
send_response=$(curl -s -X POST $BASE_URL/send/money \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d '{
    "paymentMethod": "wallet_balance",
    "recipientName": "John Doe",
    "recipientAccount": "0244123456",
    "recipientType": "mobile",
    "amount": 50.00,
    "investmentPercentage": 5.0,
    "donationChoice": "profit",
    "description": "Test send money"
  }')

echo "Send money response: $send_response"

# Step 5: Get transactions
echo -e "\n5️⃣ Getting transactions..."
transactions_response=$(curl -s -X GET $BASE_URL/transactions/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token")

echo "Transactions: $transactions_response"

echo -e "\n✅ Send flow test completed!"
