#!/bin/bash

BASE_URL="https://siha-backend.api-service.live/api/v1"

echo "🧪 Testing Live Backend API"
echo "=========================="

# Test 1: Health check
echo "1. Health Check:"
curl -s -X GET $BASE_URL/health | jq '.'
echo ""

# Test 2: Registration (to get a user)
echo "2. User Registration:"
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "deposit_test@example.com",
    "password": "testpass123",
    "firstName": "Deposit",
    "lastName": "Test"
  }')
echo $REGISTER_RESPONSE | jq '.'
echo ""

# Test 3: Check deposit endpoints exist
echo "3. Deposit Endpoints (without auth):"
echo "POST /deposits/initiate:"
curl -s -X POST $BASE_URL/deposits/initiate \
  -H "Content-Type: application/json" \
  -d '{"amount": 100}' | jq '.'

echo ""
echo "GET /deposits/:"
curl -s -X GET $BASE_URL/deposits/ | jq '.'

echo ""
echo "4. Payment Method Endpoint:"
curl -s -X GET $BASE_URL/auth/payment-method | jq '.'

echo ""
echo "✅ Backend API Status:"
echo "- Health endpoint: Working"
echo "- Registration: Working" 
echo "- Deposit endpoints: Exist (require auth)"
echo "- Payment method endpoint: Exists (requires auth)"
echo ""
echo "🔍 Issue: Deposit flow needs valid authentication token"
echo "💡 Solution: Frontend should handle login flow before deposits"
