#!/bin/bash

echo "🧪 Testing Backend Deposit Flow"
echo "================================"

BASE_URL="http://localhost:8080/api/v1"

# Test 1: Health check
echo "1. Testing health endpoint..."
HEALTH=$(curl -s "$BASE_URL/health")
echo "Health: $HEALTH"

# Test 2: Try to register a test user
echo -e "\n2. Registering test user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@healthypay.com",
    "password": "password123",
    "firstName": "Test",
    "lastName": "User"
  }')
echo "Register: $REGISTER_RESPONSE"

# Test 3: Try to login
echo -e "\n3. Logging in test user..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@healthypay.com",
    "password": "password123"
  }')
echo "Login: $LOGIN_RESPONSE"

# Extract token if login successful
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ ! -z "$TOKEN" ]; then
  echo -e "\n4. Testing deposit initiation with token..."
  DEPOSIT_RESPONSE=$(curl -s -X POST "$BASE_URL/deposits/initiate" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "amount": 10.0,
      "paymentMethodId": "69166645b371c954f6aa0c04",
      "investmentPercentage": 0.0,
      "donationChoice": "none"
    }')
  echo "Deposit: $DEPOSIT_RESPONSE"
else
  echo -e "\n❌ No token received, cannot test deposit"
fi

echo -e "\n✅ Backend validation complete"
