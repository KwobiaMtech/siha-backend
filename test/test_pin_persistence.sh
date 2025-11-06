#!/bin/bash

echo "=== PIN Persistence Fix Test ==="
echo

# Test 1: Register new user
echo "1. Registering new user for PIN persistence test..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pinpersist@test.com",
    "password": "password123",
    "firstName": "PIN",
    "lastName": "Persist"
  }')

echo "Registration Response:"
echo $REGISTER_RESPONSE | jq '.'
echo

# Test 2: Get verification code
echo "2. Getting verification code..."
CODE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/test-email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pinpersist@test.com"
  }')

VERIFICATION_CODE=$(echo $CODE_RESPONSE | jq -r '.code')
echo "Verification Code: $VERIFICATION_CODE"

# Test 3: Verify email to get token
echo "3. Verifying email to get auth token..."
VERIFY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"pinpersist@test.com\",
    \"code\": \"$VERIFICATION_CODE\"
  }")

echo "Verification Response:"
echo $VERIFY_RESPONSE | jq '.'

TOKEN=$(echo $VERIFY_RESPONSE | jq -r '.token // empty')
echo "Token: $TOKEN"

if [ "$TOKEN" != "" ] && [ "$TOKEN" != "null" ]; then
  # Test 4: Setup PIN with valid token
  echo ""
  echo "4. Setting up PIN with valid token..."
  PIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/setup-pin \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "pin": "1234"
    }')

  echo "PIN Setup Response:"
  echo $PIN_RESPONSE | jq '.'
  
  PIN_MESSAGE=$(echo $PIN_RESPONSE | jq -r '.message // empty')
  
  if [ "$PIN_MESSAGE" = "PIN set successfully" ]; then
    echo "✅ PIN setup API call successful"
    
    # Test 5: Login to verify PIN was persisted
    echo ""
    echo "5. Testing login to verify PIN persistence..."
    LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
      -H "Content-Type: application/json" \
      -d '{
        "email": "pinpersist@test.com",
        "password": "password123"
      }')

    echo "Login Response:"
    echo $LOGIN_RESPONSE | jq '.'
    
    HAS_PIN=$(echo $LOGIN_RESPONSE | jq -r '.hasPIN // false')
    echo "Has PIN after setup: $HAS_PIN"
    
    if [ "$HAS_PIN" = "true" ]; then
      echo ""
      echo "✅ PIN PERSISTENCE TEST PASSED!"
      echo "✅ PIN was successfully saved to database"
      echo "✅ Login correctly returns hasPIN: true"
      echo "✅ Database update is working correctly"
    else
      echo ""
      echo "❌ PIN PERSISTENCE TEST FAILED!"
      echo "❌ PIN was not saved to database"
      echo "❌ Login returns hasPIN: $HAS_PIN (expected: true)"
    fi
  else
    echo "❌ PIN setup failed: $PIN_MESSAGE"
  fi
else
  echo "❌ Could not get valid token for PIN setup test"
  echo "   This might be due to email verification code issues"
  echo "   But we can still test the PIN setup endpoint structure"
  
  # Test PIN setup without valid token (should fail)
  echo ""
  echo "Testing PIN setup endpoint structure..."
  NO_TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/setup-pin \
    -H "Content-Type: application/json" \
    -d '{
      "pin": "1234"
    }')
  
  echo "Response without token:"
  echo $NO_TOKEN_RESPONSE | jq '.'
  
  if [ "$(echo $NO_TOKEN_RESPONSE | jq -r '.error')" = "Authorization header required" ]; then
    echo "✅ PIN setup endpoint is properly protected"
  fi
fi

echo ""
echo "=== PIN Persistence Test Complete ==="
