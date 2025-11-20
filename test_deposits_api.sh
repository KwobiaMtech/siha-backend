#!/bin/bash

echo "🧪 Testing Deposits API Flow"
echo "============================"

BASE_URL="http://localhost:8080/api/v1"

# Step 1: Login
echo "1. Logging in..."
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "e2e_test@healthypay.com", "password": "password123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [[ $TOKEN ]]; then
    echo "✅ Login successful"
else
    echo "❌ Login failed"
    exit 1
fi

# Step 2: Get existing deposits
echo -e "\n2. Getting user deposits..."
DEPOSITS_RESPONSE=$(curl -s -X GET "$BASE_URL/deposits/" \
  -H "Authorization: Bearer $TOKEN")

echo "Deposits Response: $DEPOSITS_RESPONSE"

# Count deposits
DEPOSIT_COUNT=$(echo $DEPOSITS_RESPONSE | grep -o '"id":"[^"]*"' | wc -l)
echo "Found $DEPOSIT_COUNT deposits"

# Step 3: Create new deposit
echo -e "\n3. Creating new deposit..."
NEW_DEPOSIT=$(curl -s -X POST "$BASE_URL/deposits/initiate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "amount": 5.0,
    "paymentMethodId": "691f916309592f22d6ff813a",
    "investmentPercentage": 10.0,
    "donationChoice": "none"
  }')

echo "New Deposit: $NEW_DEPOSIT"

# Step 4: Get updated deposits list
echo -e "\n4. Getting updated deposits list..."
UPDATED_DEPOSITS=$(curl -s -X GET "$BASE_URL/deposits/" \
  -H "Authorization: Bearer $TOKEN")

NEW_COUNT=$(echo $UPDATED_DEPOSITS | grep -o '"id":"[^"]*"' | wc -l)
echo "Updated count: $NEW_COUNT deposits"

if [[ $NEW_COUNT -gt $DEPOSIT_COUNT ]]; then
    echo "✅ New deposit added successfully"
else
    echo "⚠️ Deposit count unchanged"
fi

echo -e "\n🏁 Deposits API test complete"
