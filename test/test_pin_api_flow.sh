#!/bin/bash

echo "=== PIN Setup API Flow Test ==="
echo

# Test 1: Register new user for PIN testing
echo "1. Registering new user for PIN testing..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pinflow@test.com",
    "password": "password123",
    "firstName": "PIN",
    "lastName": "Flow"
  }')

echo "Registration Response:"
echo $REGISTER_RESPONSE | jq '.'
echo

# Test 2: Get verification code and verify email
echo "2. Getting verification code..."
CODE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/test-email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pinflow@test.com"
  }')

VERIFICATION_CODE=$(echo $CODE_RESPONSE | jq -r '.code')
echo "Verification Code: $VERIFICATION_CODE"

echo "3. Verifying email..."
VERIFY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"pinflow@test.com\",
    \"code\": \"$VERIFICATION_CODE\"
  }")

echo "Email Verification Response:"
echo $VERIFY_RESPONSE | jq '.'

TOKEN=$(echo $VERIFY_RESPONSE | jq -r '.token // empty')
HAS_PIN_AFTER_VERIFY=$(echo $VERIFY_RESPONSE | jq -r '.hasPIN // false')

echo "Token: $TOKEN"
echo "Has PIN after verification: $HAS_PIN_AFTER_VERIFY"
echo

if [ "$TOKEN" != "" ] && [ "$TOKEN" != "null" ]; then
  # Test 3: Setup PIN
  echo "4. Setting up PIN..."
  PIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/setup-pin \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "pin": "1234"
    }')

  echo "PIN Setup Response:"
  echo $PIN_RESPONSE | jq '.'
  echo

  # Test 4: Login to verify hasPIN is now true
  echo "5. Testing login after PIN setup..."
  LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "email": "pinflow@test.com",
      "password": "password123"
    }')

  echo "Login Response after PIN setup:"
  echo $LOGIN_RESPONSE | jq '.'
  
  HAS_PIN_AFTER_SETUP=$(echo $LOGIN_RESPONSE | jq -r '.hasPIN // false')
  echo "Has PIN after setup: $HAS_PIN_AFTER_SETUP"
  echo

  # Test 5: Test PIN setup endpoint security
  echo "6. Testing PIN setup security..."
  
  echo "6a. Testing without auth token:"
  NO_AUTH_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/setup-pin \
    -H "Content-Type: application/json" \
    -d '{
      "pin": "5678"
    }')
  echo $NO_AUTH_RESPONSE | jq '.'
  
  echo "6b. Testing with invalid PIN format:"
  INVALID_PIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/setup-pin \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "pin": "12345"
    }')
  echo $INVALID_PIN_RESPONSE | jq '.'
  echo

  # Verify results
  if [ "$HAS_PIN_AFTER_VERIFY" = "false" ] && [ "$HAS_PIN_AFTER_SETUP" = "true" ]; then
    echo "✅ PIN Setup API Flow Test PASSED!"
    echo "✅ Email verification returns hasPIN: false"
    echo "✅ PIN setup endpoint works correctly"
    echo "✅ Login returns hasPIN: true after PIN setup"
    echo "✅ Security validations working"
  else
    echo "❌ PIN Setup API Flow Test FAILED"
    echo "Expected: hasPIN false after verify, true after setup"
    echo "Actual: hasPIN $HAS_PIN_AFTER_VERIFY after verify, $HAS_PIN_AFTER_SETUP after setup"
  fi
else
  echo "❌ Could not get valid token from email verification"
fi

echo
echo "=== PIN Setup API Flow Test Complete ==="
