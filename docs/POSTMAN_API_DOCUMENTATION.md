# HealthyPay Backend API - Postman Collection Documentation

## Overview

This comprehensive Postman collection provides complete API documentation for the HealthyPay backend services, including sample requests, responses, and automated testing scripts.

## Collection File

**File**: `HealthyPay_Complete_API.postman_collection.json`

## Setup Instructions

### 1. Import Collection
1. Open Postman
2. Click "Import" button
3. Select the `HealthyPay_Complete_API.postman_collection.json` file
4. Collection will be imported with all endpoints and examples

### 2. Environment Variables
The collection uses the following variables:

```json
{
  "base_url": "http://localhost:8080",
  "auth_token": ""
}
```

**Setup**:
1. Create a new environment in Postman
2. Add the variables above
3. Set `base_url` to your backend server URL
4. `auth_token` will be automatically set after login

### 3. Authentication Flow
The collection includes automatic token management:
1. Run "Login User" request
2. Token is automatically extracted and stored in `auth_token` variable
3. All subsequent requests use this token automatically

## API Endpoints Overview

### 🔐 Authentication
- **POST** `/api/v1/auth/register` - Register new user
- **POST** `/api/v1/auth/login` - Login user (auto-saves token)
- **POST** `/api/v1/auth/verify-email` - Verify email with code
- **POST** `/api/v1/auth/setup-pin` - Setup user PIN

### 💰 Wallet Management
- **GET** `/api/v1/wallet/balance` - Get wallet balance
- **POST** `/api/v1/wallet/create-blockchain` - Create blockchain wallet
- **GET** `/api/v1/wallet/supported-blockchains` - List supported blockchains
- **POST** `/api/v1/wallet/add-funds` - Add funds to wallet

### 💸 Send Money Flow
- **GET** `/api/v1/send/payment-methods` - Get available payment methods
- **GET** `/api/v1/send/recipients` - Get saved recipients
- **POST** `/api/v1/send/money` - Send money transaction

### 🏦 PSP Management
- **GET** `/api/v1/psp/providers` - List available PSP providers
- **GET** `/api/v1/psp/recommend?provider=MTN` - Get PSP recommendation
- **POST** `/api/v1/psp/test/:psp` - Test PSP connection

### ⭐ Stellar Blockchain
- **GET** `/api/v1/stellar/info` - Get Stellar wallet info
- **POST** `/api/v1/stellar/wallet` - Create Stellar wallet
- **POST** `/api/v1/stellar/send-usdc` - Send USDC on Stellar

### 📊 Transactions
- **GET** `/api/v1/transactions` - Get transaction history

### 📱 Mobile Money
- **POST** `/api/v1/mobile-money/add-wallet` - Add mobile money wallet
- **GET** `/api/v1/mobile-money/wallets` - Get mobile money wallets

### 🏥 Health Check
- **GET** `/api/v1/health` - Backend health check

## Sample Request/Response Examples

### Authentication

#### Register User
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "firstName": "John",
  "lastName": "Doe"
}
```

**Response (201)**:
```json
{
  "message": "User registered successfully. Please verify your email.",
  "userId": "64f7b8c9e1234567890abcde",
  "email": "user@example.com"
}
```

#### Login User
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200)**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "64f7b8c9e1234567890abcde",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "isVerified": true,
    "kycStatus": "pending"
  }
}
```

### Wallet Management

#### Get Wallet Balance
```bash
GET /api/v1/wallet/balance
Authorization: Bearer {token}
```

**Response (200)**:
```json
{
  "type": "blockchain",
  "blockchain": "stellar",
  "wallet": {
    "id": "64f7b8c9e1234567890abcde",
    "userId": "64f7b8c9e1234567890abcde",
    "blockchain": "stellar",
    "network": "testnet",
    "publicKey": "GXXX...",
    "isDefault": true,
    "balances": [
      {
        "assetCode": "USDC",
        "balance": 1000.50,
        "symbol": "USDC",
        "name": "USD Coin"
      },
      {
        "assetCode": "XLM",
        "balance": 25.0,
        "symbol": "XLM",
        "name": "Stellar Lumens"
      }
    ]
  }
}
```

### Send Money Flow

#### Get Payment Methods
```bash
GET /api/v1/send/payment-methods
Authorization: Bearer {token}
```

**Response (200)**:
```json
{
  "paymentMethods": [
    {
      "id": "wallet_balance",
      "title": "💰 Wallet Balance",
      "subtitle": "Send from your platform wallet",
      "balance": 2540.50,
      "type": "wallet",
      "hasBalance": true
    },
    {
      "id": "pm_64f7b8c9e1234567890abcde",
      "title": "📱 MTN Mobile Money",
      "subtitle": "MTN - 024****456",
      "type": "mobile_money",
      "provider": "MTN",
      "isDefault": true,
      "hasBalance": false
    }
  ]
}
```

#### Send Money
```bash
POST /api/v1/send/money
Authorization: Bearer {token}
Content-Type: application/json

{
  "paymentMethodId": "wallet_balance",
  "recipientName": "John Doe",
  "recipientAccount": "0244123456",
  "recipientType": "mobile_money",
  "recipientNetwork": "MTN",
  "amount": 100.00,
  "investmentPercentage": 5.0,
  "donationChoice": "profit",
  "description": "Money transfer via HealthyPay"
}
```

**Response (200)**:
```json
{
  "message": "Money sent successfully",
  "transaction": {
    "id": "64f7b8c9e1234567890abcde",
    "amount": 100.00,
    "investmentAmount": 5.00,
    "status": "completed",
    "createdAt": "2024-10-25T12:00:00Z"
  }
}
```

### PSP Management

#### Get Available PSPs
```bash
GET /api/v1/psp/providers
Authorization: Bearer {token}
```

**Response (200)**:
```json
{
  "providers": ["ogate", "mtn", "demo"],
  "message": "Available PSP providers"
}
```

#### Test PSP Connection
```bash
POST /api/v1/psp/test/ogate
Authorization: Bearer {token}
```

**Response (200)**:
```json
{
  "psp": "ogate",
  "status": "connected",
  "response": {
    "transactionId": "OG_1234567890",
    "status": "pending",
    "message": "Collection initiated successfully"
  },
  "message": "PSP connection test successful"
}
```

## Error Responses

### Common Error Formats

#### Authentication Error (401)
```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

#### Validation Error (400)
```json
{
  "error": "Validation failed",
  "details": {
    "email": "Invalid email format",
    "password": "Password must be at least 8 characters"
  }
}
```

#### Not Found Error (404)
```json
{
  "error": "Resource not found",
  "message": "The requested resource was not found"
}
```

#### Server Error (500)
```json
{
  "error": "Internal server error",
  "message": "An unexpected error occurred"
}
```

## Testing Workflows

### 1. Complete User Registration Flow
1. **Register User** → Get userId
2. **Verify Email** → Get token
3. **Setup PIN** → Complete profile
4. **Create Blockchain Wallet** → Setup wallet

### 2. Send Money Flow
1. **Login User** → Get token
2. **Get Payment Methods** → Choose method
3. **Get Recipients** → Choose or add recipient
4. **Send Money** → Execute transaction
5. **Get Transactions** → Verify transaction

### 3. PSP Testing Flow
1. **Get Available PSPs** → See available providers
2. **Get PSP Recommendation** → Get best PSP for network
3. **Test PSP Connection** → Verify PSP connectivity

## Environment Configuration

### Development Environment
```json
{
  "base_url": "http://localhost:8080",
  "auth_token": ""
}
```

### Staging Environment
```json
{
  "base_url": "https://staging-api.healthypay.com",
  "auth_token": ""
}
```

### Production Environment
```json
{
  "base_url": "https://api.healthypay.com",
  "auth_token": ""
}
```

## Automated Testing Scripts

The collection includes automated test scripts:

### Login Request Test Script
```javascript
if (pm.response.code === 200) {
    const response = pm.response.json();
    pm.collectionVariables.set('auth_token', response.token);
    pm.test("Token saved successfully", function () {
        pm.expect(response.token).to.be.a('string');
    });
}
```

### Response Validation
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has required fields", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('message');
});
```

## Usage Tips

### 1. **Sequential Testing**
- Run requests in order for dependent operations
- Use "Run Collection" feature for complete flow testing

### 2. **Environment Switching**
- Create separate environments for dev/staging/prod
- Switch environments easily in Postman

### 3. **Token Management**
- Login request automatically saves token
- Token is used in all subsequent requests
- Re-login if token expires

### 4. **Error Debugging**
- Check response status codes
- Review error messages in response body
- Verify request headers and body format

## Support

For API issues or questions:
- Check response status codes and error messages
- Verify authentication tokens are valid
- Ensure request body format matches examples
- Review server logs for detailed error information

This comprehensive Postman collection provides everything needed to test and integrate with the HealthyPay backend API.
