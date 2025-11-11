# Validate Email Route Tests

## Overview
Test suite for the `/api/auth/validate-email` route that checks if an email address is already registered.

## Test Files

### 1. `validate_email_final_test.go`
**Unit tests for input validation**
- ✅ Valid email format acceptance
- ✅ Invalid email format rejection  
- ✅ Missing email field handling
- ✅ Empty email handling
- ✅ Various valid email formats
- ✅ Various invalid email formats

### 2. `validate_email.http`
**Manual HTTP tests**
- Test existing email check
- Test available email check
- Test validation errors

### 3. `test_validate_email.sh`
**Integration tests** (requires running server)
- HTTP status code validation
- Response format validation
- End-to-end testing

## Running Tests

### Unit Tests
```bash
cd /Users/kwabena/Documents/project_files/healthyPay/backend
go test -run TestValidateEmailRoute ./test/ -v
```

### Integration Tests
```bash
# Start the server first
go run main.go

# In another terminal
./test/test_validate_email.sh
```

### Manual Testing
Use the `validate_email.http` file with VS Code REST Client extension or similar tools.

## Test Coverage

### Input Validation ✅
- Email format validation using Gin's built-in validator
- Required field validation
- Empty value handling

### Response Format ✅
- Correct JSON structure
- Proper HTTP status codes
- Error message format

### Edge Cases ✅
- Various valid email formats
- Common invalid email patterns
- Missing/empty fields

## Expected Responses

### Valid Email (Available)
```json
{
  "exists": false,
  "message": "Email is available"
}
```

### Valid Email (Already Registered)
```json
{
  "exists": true,
  "message": "Email is already registered"
}
```

### Invalid Input
```json
{
  "error": "Key: 'Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```
