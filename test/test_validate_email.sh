#!/bin/bash

echo "🧪 Testing validate-email route"

BASE_URL="http://localhost:8080/api/auth"

# Test 1: Valid email format - should return 200
echo "📧 Test 1: Valid email format"
response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/validate-email" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}')

http_code="${response: -3}"
if [ "$http_code" = "200" ]; then
    echo "✅ Valid email format test passed"
else
    echo "❌ Valid email format test failed (HTTP $http_code)"
fi

# Test 2: Invalid email format - should return 400
echo "📧 Test 2: Invalid email format"
response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/validate-email" \
  -H "Content-Type: application/json" \
  -d '{"email": "invalid-email"}')

http_code="${response: -3}"
if [ "$http_code" = "400" ]; then
    echo "✅ Invalid email format test passed"
else
    echo "❌ Invalid email format test failed (HTTP $http_code)"
fi

# Test 3: Missing email field - should return 400
echo "📧 Test 3: Missing email field"
response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/validate-email" \
  -H "Content-Type: application/json" \
  -d '{}')

http_code="${response: -3}"
if [ "$http_code" = "400" ]; then
    echo "✅ Missing email field test passed"
else
    echo "❌ Missing email field test failed (HTTP $http_code)"
fi

# Test 4: Empty email - should return 400
echo "📧 Test 4: Empty email"
response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/validate-email" \
  -H "Content-Type: application/json" \
  -d '{"email": ""}')

http_code="${response: -3}"
if [ "$http_code" = "400" ]; then
    echo "✅ Empty email test passed"
else
    echo "❌ Empty email test failed (HTTP $http_code)"
fi

# Test 5: Check response format for valid email
echo "📧 Test 5: Response format validation"
response=$(curl -s -X POST "$BASE_URL/validate-email" \
  -H "Content-Type: application/json" \
  -d '{"email": "newuser@example.com"}')

if echo "$response" | grep -q '"exists"' && echo "$response" | grep -q '"message"'; then
    echo "✅ Response format test passed"
else
    echo "❌ Response format test failed"
    echo "Response: $response"
fi

echo "🏁 validate-email route tests completed"
