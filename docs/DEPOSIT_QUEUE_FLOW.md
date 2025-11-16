# Deposit Queue Flow Documentation

## Overview
The deposit process has been enhanced with queue flow tracking to ensure reliable processing of successful deposits and proper status management.

## Deposit Flow

### 1. Deposit Initiation
- User initiates deposit via `/api/v1/deposits/initiate`
- Deposit record created with:
  - `status`: "pending" → "initiated"
  - `queueStatus`: "queued" → "processing"
  - `transactionId`: PSP transaction ID
  - `pspReference`: Unique reference

### 2. Queue Processing
The transaction queue runs every 30 seconds and processes:

#### Regular Transactions
- Checks PSP status for pending transactions
- Updates transaction status based on PSP response

#### Deposits
- Checks PSP status for pending deposits
- Updates deposit status and queue status
- Processes successful deposits automatically

### 3. Successful Deposit Processing
When a deposit is marked as "collected":

1. **Wallet Update**
   - Calculates savings amount: `amount * (100 - investmentPercentage) / 100`
   - Updates user wallet balance

2. **Investment Creation** (if applicable)
   - Calculates investment amount: `amount * investmentPercentage / 100`
   - Creates investment record in `investments` collection

3. **Donation Handling** (if applicable)
   - Creates donation record based on `donationChoice`:
     - "both": Full deposit amount
     - "profit": Investment amount only

## Status Tracking

### Deposit Status
- `pending`: Initial state
- `initiated`: PSP collection initiated
- `collected`: Payment successfully collected
- `failed`: Payment failed or timed out

### Queue Status
- `queued`: Added to processing queue
- `processing`: Being processed by queue
- `completed`: Successfully processed
- `failed`: Processing failed
- `timeout`: Timed out after 24 hours

## API Endpoints

### Initiate Deposit
```http
POST /api/v1/deposits/initiate
Content-Type: application/json

{
  "amount": 100.0,
  "paymentMethodId": "payment_method_id",
  "investmentPercentage": 20.0,
  "donationChoice": "profit"
}
```

### Check Deposit Status
```http
GET /api/v1/deposits/{id}/status
```

Response includes:
- `status`: Current deposit status
- `queueStatus`: Queue processing status
- `processedAt`: Timestamp when processing completed
- Other deposit details

### Get All Deposits
```http
GET /api/v1/deposits/
```

## Database Collections

### deposits
```javascript
{
  "_id": ObjectId,
  "userId": ObjectId,
  "amount": Number,
  "paymentMethodId": String,
  "investmentPercentage": Number,
  "donationChoice": String,
  "status": String,
  "queueStatus": String,
  "transactionId": String,
  "pspReference": String,
  "pspResponse": Object,
  "processedAt": Date,
  "createdAt": Date,
  "updatedAt": Date
}
```

## Queue Processing Logic

The queue processor:
1. Finds deposits with status "initiated" or "pending"
2. Checks PSP status for each deposit
3. Updates deposit status based on PSP response
4. For successful deposits:
   - Updates wallet balance
   - Creates investment records
   - Handles donations
   - Marks as completed with timestamp

## Error Handling

- PSP check failures are logged but don't stop processing
- Database update failures are logged
- Deposits timeout after 24 hours if not resolved
- Failed deposits are marked appropriately

## Testing

Use the test file `test/test_deposit_queue.go` to verify:
1. Deposit initiation
2. Queue status tracking
3. Automatic processing
4. Status updates

## Monitoring

Queue processing logs include:
- Number of deposits processed
- Individual deposit status changes
- Wallet balance updates
- Investment and donation creation
- Error conditions

Example log output:
```
Processing 3 pending deposits...
✅ Deposit 507f1f77bcf86cd799439011 marked as collected
💰 Updated wallet balance for user 507f1f77bcf86cd799439012: +80.00
📈 Created investment record: 20.00 for user 507f1f77bcf86cd799439012
🎁 Created donation record: 20.00 (profit) for user 507f1f77bcf86cd799439012
💰 Successfully processed 3 deposits
```
