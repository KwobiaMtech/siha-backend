#!/bin/bash

echo "🧪 Testing Deposit Queue Integration"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080/api/v1"

# Test data
DEPOSIT_DATA='{
  "amount": 75.0,
  "paymentMethodId": "test_mobile_money_id",
  "investmentPercentage": 25.0,
  "donationChoice": "profit"
}'

echo -e "\n${BLUE}1. Testing deposit initiation...${NC}"
RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_token" \
  -d "$DEPOSIT_DATA" \
  "$BASE_URL/deposits/initiate")

echo "Response: $RESPONSE"

# Extract deposit ID (assuming JSON response)
DEPOSIT_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$DEPOSIT_ID" ]; then
    echo -e "${RED}❌ Failed to get deposit ID${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Deposit initiated with ID: $DEPOSIT_ID${NC}"

echo -e "\n${BLUE}2. Checking initial status...${NC}"
curl -s -H "Authorization: Bearer test_token" \
  "$BASE_URL/deposits/$DEPOSIT_ID/status" | jq '.'

echo -e "\n${BLUE}3. Monitoring queue processing (checking every 10 seconds)...${NC}"
for i in {1..6}; do
    echo -e "\n${YELLOW}Status check $i:${NC}"
    STATUS_RESPONSE=$(curl -s -H "Authorization: Bearer test_token" \
      "$BASE_URL/deposits/$DEPOSIT_ID/status")
    
    echo $STATUS_RESPONSE | jq '.'
    
    # Check if completed
    STATUS=$(echo $STATUS_RESPONSE | jq -r '.status // empty')
    QUEUE_STATUS=$(echo $STATUS_RESPONSE | jq -r '.queueStatus // empty')
    
    if [ "$STATUS" = "collected" ] && [ "$QUEUE_STATUS" = "completed" ]; then
        echo -e "${GREEN}✅ Deposit successfully processed!${NC}"
        break
    elif [ "$STATUS" = "failed" ]; then
        echo -e "${RED}❌ Deposit failed${NC}"
        break
    fi
    
    if [ $i -lt 6 ]; then
        echo "Waiting 10 seconds..."
        sleep 10
    fi
done

echo -e "\n${BLUE}4. Final status check...${NC}"
curl -s -H "Authorization: Bearer test_token" \
  "$BASE_URL/deposits/$DEPOSIT_ID/status" | jq '.'

echo -e "\n${GREEN}🎯 Test completed!${NC}"
