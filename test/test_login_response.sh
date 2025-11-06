#!/bin/bash

echo "Testing login response format..."

# First, let's try common verification codes
echo "Trying to verify with common codes..."
for code in "123456" "000000" "111111" "999999"; do
    echo "Trying code: $code"
    response=$(curl -s -X POST http://localhost:8080/api/v1/auth/verify-email \
      -H "Content-Type: application/json" \
      -d "{\"email\": \"test@example.com\", \"code\": \"$code\"}")
    
    if echo "$response" | grep -q "success"; then
        echo "✅ Verification successful with code: $code"
        break
    fi
done

# Now test login
echo -e "\n🔐 Testing login..."
login_response=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}')

echo "Login response:"
echo "$login_response" | jq .

# Check if hasPIN and hasPaymentMethod are in response
if echo "$login_response" | jq -e '.hasPIN' > /dev/null; then
    echo "✅ hasPIN field found: $(echo "$login_response" | jq -r '.hasPIN')"
else
    echo "❌ hasPIN field missing"
fi

if echo "$login_response" | jq -e '.hasPaymentMethod' > /dev/null; then
    echo "✅ hasPaymentMethod field found: $(echo "$login_response" | jq -r '.hasPaymentMethod')"
else
    echo "❌ hasPaymentMethod field missing"
fi
