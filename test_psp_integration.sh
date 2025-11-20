#!/bin/bash

echo "🔄 Testing PSP Integration for Live Deposits"
echo "============================================"

BASE_URL="http://localhost:8080/api/v1"

# Step 1: Health check
echo "1. Checking backend health..."
HEALTH=$(curl -s "$BASE_URL/health")
if [[ $HEALTH == *"healthy"* ]]; then
    echo "✅ Backend is running"
else
    echo "❌ Backend is not running"
    exit 1
fi

# Step 2: Login to get token
echo -e "\n2. Getting authentication token..."
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "e2e_test@healthypay.com", "password": "password123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [[ $TOKEN ]]; then
    echo "✅ Login successful"
else
    echo "❌ Login failed"
    exit 1
fi

# Step 3: Test deposit with live PSP
echo -e "\n3. Testing live PSP deposit initiation..."
DEPOSIT_RESPONSE=$(curl -s -X POST "$BASE_URL/deposits/initiate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "amount": 1.0,
    "paymentMethodId": "691f8f78b371c954f6aa0c05",
    "investmentPercentage": 0.0,
    "donationChoice": "none"
  }')

echo "Deposit Response: $DEPOSIT_RESPONSE"

# Check if deposit was initiated
if [[ $DEPOSIT_RESPONSE == *"id"* && $DEPOSIT_RESPONSE == *"initiated"* ]]; then
    echo "✅ Deposit initiated with PSP"
    
    # Extract deposit ID
    DEPOSIT_ID=$(echo $DEPOSIT_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    
    # Step 4: Check deposit status
    echo -e "\n4. Checking deposit status..."
    STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/deposits/$DEPOSIT_ID/status" \
      -H "Authorization: Bearer $TOKEN")
    
    echo "Status Response: $STATUS_RESPONSE"
    
    if [[ $STATUS_RESPONSE == *"status"* ]]; then
        echo "✅ Status check successful"
    else
        echo "❌ Status check failed"
    fi
    
else
    echo "❌ Deposit initiation failed"
    echo "Response: $DEPOSIT_RESPONSE"
fi

echo -e "\n🏁 PSP Integration test complete"
