#!/bin/bash

echo "=== Simple PIN Persistence Test ==="
echo

# Create a user and get a token through the normal flow
echo "1. Creating user and getting token..."

# Register user
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pinsimple@test.com",
    "password": "password123",
    "firstName": "PIN",
    "lastName": "Simple"
  }')

echo "Registration: $(echo $REGISTER_RESPONSE | jq -r '.message // .error')"

# Get verification code
CODE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/test-email \
  -H "Content-Type: application/json" \
  -d '{"email": "pinsimple@test.com"}')

VERIFICATION_CODE=$(echo $CODE_RESPONSE | jq -r '.code // empty')
echo "Verification code: $VERIFICATION_CODE"

if [ "$VERIFICATION_CODE" != "" ] && [ "$VERIFICATION_CODE" != "null" ]; then
  # Verify email
  VERIFY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/verify-email \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"pinsimple@test.com\", \"code\": \"$VERIFICATION_CODE\"}")

  TOKEN=$(echo $VERIFY_RESPONSE | jq -r '.token // empty')
  
  if [ "$TOKEN" != "" ] && [ "$TOKEN" != "null" ]; then
    echo "✅ Got valid token: ${TOKEN:0:20}..."
    
    # Test PIN setup
    echo ""
    echo "2. Testing PIN setup..."
    PIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/setup-pin \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d '{"pin": "1234"}')

    echo "PIN setup response:"
    echo $PIN_RESPONSE | jq '.'
    
    # Test login to check hasPIN
    echo ""
    echo "3. Testing login to verify PIN persistence..."
    LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
      -H "Content-Type: application/json" \
      -d '{
        "email": "pinsimple@test.com",
        "password": "password123"
      }')

    echo "Login response:"
    echo $LOGIN_RESPONSE | jq '.'
    
    HAS_PIN=$(echo $LOGIN_RESPONSE | jq -r '.hasPIN // false')
    if [ "$HAS_PIN" = "true" ]; then
      echo ""
      echo "✅ PIN PERSISTENCE TEST PASSED!"
      echo "✅ PIN was successfully saved to database"
    else
      echo ""
      echo "❌ PIN PERSISTENCE TEST FAILED!"
      echo "❌ hasPIN is still false after PIN setup"
    fi
  else
    echo "❌ Could not get valid token from email verification"
  fi
else
  echo "❌ Could not get verification code"
fi

echo ""
echo "=== Simple PIN Persistence Test Complete ==="
