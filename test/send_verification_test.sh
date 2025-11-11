#!/bin/bash

BASE_URL="http://localhost:8080/api/v1/auth"

echo "Testing send-verification endpoint..."

# Test 1: Valid email (user exists)
echo -e "\n1. Testing with valid email:"
curl -X POST "$BASE_URL/send-verification" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -w "\nStatus: %{http_code}\n"

# Test 2: Invalid email format
echo -e "\n2. Testing with invalid email format:"
curl -X POST "$BASE_URL/send-verification" \
  -H "Content-Type: application/json" \
  -d '{"email": "invalid-email"}' \
  -w "\nStatus: %{http_code}\n"

# Test 3: Empty email
echo -e "\n3. Testing with empty email:"
curl -X POST "$BASE_URL/send-verification" \
  -H "Content-Type: application/json" \
  -d '{"email": ""}' \
  -w "\nStatus: %{http_code}\n"

# Test 4: Non-existent user
echo -e "\n4. Testing with non-existent user:"
curl -X POST "$BASE_URL/send-verification" \
  -H "Content-Type: application/json" \
  -d '{"email": "nonexistent@example.com"}' \
  -w "\nStatus: %{http_code}\n"
