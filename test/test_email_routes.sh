#!/bin/bash

BASE_URL="http://localhost:8080/api/v1/auth"

echo "Testing email validation routes..."

# Test 1: Valid email
echo -e "\n1. Testing valid email:"
curl -X POST "$BASE_URL/validate-email" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -w "\nStatus: %{http_code}\n"

# Test 2: Invalid email format
echo -e "\n2. Testing invalid email format:"
curl -X POST "$BASE_URL/validate-email" \
  -H "Content-Type: application/json" \
  -d '{"email": "invalid-email"}' \
  -w "\nStatus: %{http_code}\n"

# Test 3: Empty email
echo -e "\n3. Testing empty email:"
curl -X POST "$BASE_URL/validate-email" \
  -H "Content-Type: application/json" \
  -d '{"email": ""}' \
  -w "\nStatus: %{http_code}\n"

# Test 4: Verify email endpoint
echo -e "\n4. Testing email verification:"
curl -X POST "$BASE_URL/verify-email" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "code": "123456"}' \
  -w "\nStatus: %{http_code}\n"
