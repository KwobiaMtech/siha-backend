#!/bin/bash

echo "🧪 End-to-End Backend Deposit Flow Test"
echo "========================================"

BASE_URL="http://localhost:8080/api/v1"
TEST_EMAIL="e2e_test@healthypay.com"
TEST_PASSWORD="password123"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

test_step() {
    echo -e "\n${YELLOW}$1${NC}"
}

test_pass() {
    echo -e "${GREEN}✅ $1${NC}"
    ((TESTS_PASSED++))
}

test_fail() {
    echo -e "${RED}❌ $1${NC}"
    ((TESTS_FAILED++))
}

# Step 1: Health Check
test_step "1. Testing Backend Health"
HEALTH=$(curl -s "$BASE_URL/health")
if [[ $HEALTH == *"healthy"* ]]; then
    test_pass "Backend is running"
else
    test_fail "Backend is not running"
    exit 1
fi

# Step 2: Clean up existing test user
test_step "2. Cleaning up existing test data"
mongosh --eval "db.users.deleteOne({email: '$TEST_EMAIL'})" healthy_pay > /dev/null 2>&1
mongosh --eval "db.payment_methods.deleteMany({user_id: ObjectId('000000000000000000000000')})" healthy_pay > /dev/null 2>&1
mongosh --eval "db.deposits.deleteMany({userId: ObjectId('000000000000000000000000')})" healthy_pay > /dev/null 2>&1

# Step 3: Register new test user or use existing
test_step "3. Setting up test user"
# Try to find existing user first
EXISTING_USER=$(mongosh --quiet --eval "
    const user = db.users.findOne({email: '$TEST_EMAIL'});
    if (user) {
        print(user._id.toString());
    } else {
        print('NOT_FOUND');
    }
" healthy_pay)

if [[ $EXISTING_USER == "NOT_FOUND" ]]; then
    # Create user directly in database to bypass email verification
    USER_ID=$(mongosh --quiet --eval "
        const result = db.users.insertOne({
            email: '$TEST_EMAIL',
            password: '\$2a\$14\$EDLYXvJfSMtl5j.RY9zYceUgUPdEEurTFX0nmnHwVqdCq585gZcA2',
            first_name: 'E2E',
            last_name: 'Test',
            pin: '',
            is_verified: true,
            verification_code: '',
            kyc_status: 'pending',
            created_at: new Date(),
            updated_at: new Date()
        });
        print(result.insertedId.toString());
    " healthy_pay)
    test_pass "User created directly in database (ID: ${USER_ID:0:8}...)"
else
    USER_ID=$EXISTING_USER
    test_pass "Using existing user (ID: ${USER_ID:0:8}...)"
fi

# Step 4: Create payment method for user
test_step "4. Creating payment method"
PAYMENT_METHOD_ID=$(mongosh --quiet --eval "
    const result = db.payment_methods.insertOne({
        user_id: ObjectId('$USER_ID'),
        type: 'mobile_money',
        provider: 'MTN',
        network: 'MTN',
        phone_number: '+233241234567',
        account_name: 'E2E Test User',
        currency: 'GHS',
        is_default: true,
        is_active: true,
        created_at: new Date(),
        updated_at: new Date()
    });
    print(result.insertedId.toString());
" healthy_pay)

if [[ $PAYMENT_METHOD_ID ]]; then
    test_pass "Payment method created (ID: ${PAYMENT_METHOD_ID:0:8}...)"
else
    test_fail "Failed to create payment method"
    exit 1
fi

# Step 5: Login to get JWT token
test_step "5. Logging in to get JWT token"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [[ $TOKEN ]]; then
    test_pass "Login successful (Token: ${TOKEN:0:20}...)"
else
    echo "Login response: $LOGIN_RESPONSE"
    test_fail "Login failed"
    exit 1
fi

# Step 6: Test deposit initiation
test_step "6. Testing deposit initiation"
DEPOSIT_RESPONSE=$(curl -s -X POST "$BASE_URL/deposits/initiate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"amount\": 50.0,
    \"paymentMethodId\": \"$PAYMENT_METHOD_ID\",
    \"investmentPercentage\": 20.0,
    \"donationChoice\": \"profit\"
  }")

DEPOSIT_ID=$(echo $DEPOSIT_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
DEPOSIT_STATUS=$(echo $DEPOSIT_RESPONSE | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

if [[ $DEPOSIT_ID && $DEPOSIT_STATUS == "initiated" ]]; then
    test_pass "Deposit initiated successfully (ID: ${DEPOSIT_ID:0:8}...)"
    echo "   Amount: 50.0 GHS"
    echo "   Investment: 20%"
    echo "   Donation: profit"
else
    echo "Deposit response: $DEPOSIT_RESPONSE"
    test_fail "Deposit initiation failed"
fi

# Step 7: Test deposit status check
if [[ $DEPOSIT_ID ]]; then
    test_step "7. Testing deposit status check"
    STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/deposits/$DEPOSIT_ID/status" \
      -H "Authorization: Bearer $TOKEN")
    
    STATUS_CHECK=$(echo $STATUS_RESPONSE | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    
    if [[ $STATUS_CHECK ]]; then
        test_pass "Deposit status retrieved: $STATUS_CHECK"
    else
        echo "Status response: $STATUS_RESPONSE"
        test_fail "Deposit status check failed"
    fi
fi

# Step 8: Test get all deposits
test_step "8. Testing get all deposits"
DEPOSITS_RESPONSE=$(curl -s -X GET "$BASE_URL/deposits/" \
  -H "Authorization: Bearer $TOKEN")

if [[ $DEPOSITS_RESPONSE == *"deposits"* ]]; then
    DEPOSIT_COUNT=$(echo $DEPOSITS_RESPONSE | grep -o '"deposits":\[' | wc -l)
    test_pass "Deposits list retrieved successfully"
else
    echo "Deposits response: $DEPOSITS_RESPONSE"
    test_fail "Get deposits failed"
fi

# Step 9: Test invalid scenarios
test_step "9. Testing error scenarios"

# Test without authentication
UNAUTH_RESPONSE=$(curl -s -X POST "$BASE_URL/deposits/initiate" \
  -H "Content-Type: application/json" \
  -d "{\"amount\": 10.0, \"paymentMethodId\": \"test\"}")

if [[ $UNAUTH_RESPONSE == *"not authenticated"* || $UNAUTH_RESPONSE == *"Authorization header required"* ]]; then
    test_pass "Unauthenticated request properly rejected"
else
    test_fail "Unauthenticated request not properly handled: $UNAUTH_RESPONSE"
fi

# Test invalid amount
INVALID_AMOUNT_RESPONSE=$(curl -s -X POST "$BASE_URL/deposits/initiate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"amount\": -10.0, \"paymentMethodId\": \"$PAYMENT_METHOD_ID\"}")

if [[ $INVALID_AMOUNT_RESPONSE == *"error"* ]]; then
    test_pass "Invalid amount properly rejected"
else
    test_fail "Invalid amount not properly handled"
fi

# Step 10: Database verification
test_step "10. Verifying database records"
DB_DEPOSIT_COUNT=$(mongosh --quiet --eval "db.deposits.countDocuments({userId: ObjectId('$USER_ID')})" healthy_pay)

if [[ $DB_DEPOSIT_COUNT -gt 0 ]]; then
    test_pass "Deposit record found in database (Count: $DB_DEPOSIT_COUNT)"
else
    test_fail "No deposit records found in database"
fi

# Final Results
echo -e "\n🏁 Test Results Summary"
echo "======================="
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 All tests passed! Backend deposit flow is working correctly.${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please check the issues above.${NC}"
    exit 1
fi
