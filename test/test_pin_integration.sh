#!/bin/bash

echo "=== PIN Setup Integration Test ==="
echo

# Test 1: Register new user
echo "1. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pinintegration@test.com",
    "password": "password123",
    "firstName": "PIN",
    "lastName": "Integration"
  }')

echo "Registration Response:"
echo $REGISTER_RESPONSE | jq '.'
echo

# Test 2: Try login before verification (should fail)
echo "2. Testing login before email verification..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pinintegration@test.com",
    "password": "password123"
  }')

echo "Login Response (should fail):"
echo $LOGIN_RESPONSE | jq '.'
echo

# Test 3: Get verification code
echo "3. Getting verification code..."
CODE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/test-email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pinintegration@test.com"
  }')

echo "Code Response:"
echo $CODE_RESPONSE | jq '.'
VERIFICATION_CODE=$(echo $CODE_RESPONSE | jq -r '.code')
echo "Verification Code: $VERIFICATION_CODE"
echo

# Test 4: Verify email
echo "4. Verifying email..."
VERIFY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"pinintegration@test.com\",
    \"code\": \"$VERIFICATION_CODE\"
  }")

echo "Verification Response:"
echo $VERIFY_RESPONSE | jq '.'

# Extract token and hasPIN
TOKEN=$(echo $VERIFY_RESPONSE | jq -r '.token // empty')
HAS_PIN=$(echo $VERIFY_RESPONSE | jq -r '.hasPIN // false')

echo "Token: $TOKEN"
echo "Has PIN: $HAS_PIN"
echo

if [ "$TOKEN" != "" ] && [ "$TOKEN" != "null" ]; then
  # Test 5: Setup PIN
  echo "5. Setting up PIN..."
  PIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/setup-pin \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "pin": "1234"
    }')

  echo "PIN Setup Response:"
  echo $PIN_RESPONSE | jq '.'
  echo

  # Test 6: Login again to verify hasPIN is now true
  echo "6. Testing login after PIN setup..."
  FINAL_LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "pinintegration@test.com",
      "password": "password123"
    }')

  echo "Final Login Response:"
  echo $FINAL_LOGIN_RESPONSE | jq '.'
  
  FINAL_HAS_PIN=$(echo $FINAL_LOGIN_RESPONSE | jq -r '.hasPIN // false')
  echo "Final Has PIN: $FINAL_HAS_PIN"
  
  if [ "$FINAL_HAS_PIN" = "true" ]; then
    echo "✅ PIN Setup Integration Test PASSED!"
  else
    echo "❌ PIN Setup Integration Test FAILED - hasPIN should be true"
  fi
else
  echo "❌ Could not get valid token from email verification"
fi

echo
echo "=== Test Complete ==="
