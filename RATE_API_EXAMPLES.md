# Rate Conversion API - Sample Requests & Responses

## Authentication
All rate endpoints require authentication. Include the JWT token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

## 1. Get Onramp Rate (Local Currency → USD)

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/rates/onramp?currency=GHS" \
  -H "Authorization: Bearer <your_jwt_token>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "from_currency": "GHS",
    "to_currency": "USD",
    "rate": 0.09195715164561921,
    "timestamp": "2025-10-30T12:38:57Z"
  }
}
```

## 2. Get Offramp Rate (USD → Local Currency)

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/rates/offramp?currency=GHS" \
  -H "Authorization: Bearer <your_jwt_token>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "from_currency": "USD",
    "to_currency": "GHS",
    "rate": 10.87463,
    "timestamp": "2025-10-30T12:38:57Z"
  }
}
```

## 3. Convert Amount Between Currencies

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/rates/convert?amount=100&from=GHS&to=USD" \
  -H "Authorization: Bearer <your_jwt_token>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "original_amount": 100,
    "from_currency": "GHS",
    "to_currency": "USD",
    "converted_amount": 9.195715164561921
  }
}
```

## 4. Get All Available Exchange Rates

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/rates/all" \
  -H "Authorization: Bearer <your_jwt_token>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "base": "",
    "date": "",
    "rates": {
      "AED": 3.6725,
      "AFN": 65.340263,
      "GHS": 10.87463,
      "KES": 129.2,
      "ZMW": 21.974416,
      "USD": 1,
      "EUR": 0.860843,
      "GBP": 0.758197,
      "JPY": 153.781,
      "...": "... (200+ currencies)"
    }
  }
}
```

## Real-World Usage Examples

### Onramp Flow (User deposits GHS, gets USD)
1. User wants to deposit 1000 GHS
2. Get onramp rate: `GET /api/v1/rates/onramp?currency=GHS`
3. Calculate USD equivalent: 1000 GHS × 0.0920 = 92.0 USD
4. Show user: "1000 GHS = 92.0 USD"

### Offramp Flow (User withdraws USD as GHS)
1. User wants to withdraw 50 USD as GHS
2. Get offramp rate: `GET /api/v1/rates/offramp?currency=GHS`
3. Calculate GHS equivalent: 50 USD × 10.87463 = 543.73 GHS
4. Show user: "50 USD = 543.73 GHS"

### Cross-Currency Conversion
1. Convert between any two currencies via USD
2. Example: 100 GHS to KES = 1188.09 KES
3. Process: GHS → USD → KES

## Error Responses

**Invalid Currency:**
```json
{
  "error": "currency GHX not found"
}
```

**Missing Parameters:**
```json
{
  "error": "currency parameter is required"
}
```

**API Failure:**
```json
{
  "error": "failed to fetch exchange rates: connection timeout"
}
```

## Supported Currencies
The API supports 200+ currencies including:
- **African:** GHS, KES, ZMW, NGN, ZAR, EGP, MAD
- **Major:** USD, EUR, GBP, JPY, CAD, AUD
- **Asian:** CNY, INR, SGD, THB, MYR, PHP
- **Crypto:** BTC, ETH, LTC, XRP, DOGE

## Rate Update Frequency
- Rates are fetched in real-time from MoneyConvert API
- No caching implemented (rates are always current)
- API updates rates multiple times per day
