# Multi-PSP Integration Documentation

## Overview

HealthyPay now supports multiple Payment Service Providers (PSPs) through a flexible, extensible architecture. The system integrates Ogate as the primary PSP while supporting additional providers like MTN MoMo API and others.

## Architecture

### PSP Interface
```go
type PSPProvider interface {
    GetName() string
    InitiateCollection(req CollectionRequest) (*CollectionResponse, error)
    CheckCollectionStatus(transactionID string) (string, error)
    InitiateDelivery(req DeliveryRequest) error
}
```

### Implemented PSPs

#### 1. **Ogate PSP** (`ogate_psp.go`)
- **Primary PSP**: Handles MTN, Vodafone, AirtelTigo networks
- **Collection**: Mobile money collection via Ogate API
- **Delivery**: Mobile money disbursements
- **Configuration**: `OGATE_BASE_URL`, `OGATE_API_KEY`

#### 2. **MTN PSP** (`mtn_psp.go`)
- **Direct MTN Integration**: MTN MoMo API integration
- **Collection**: MTN MoMo collection API
- **Delivery**: MTN MoMo disbursement API
- **Configuration**: `MTN_API_KEY`, `MTN_USER_ID`, `MTN_BASE_URL`, `MTN_SUBSCRIPTION_KEY`

#### 3. **Demo PSP** (`demo_psp.go`)
- **Testing/Development**: Simulated PSP for testing
- **No External Dependencies**: Works without API keys
- **Multiple Instances**: Can create multiple demo PSPs

## PSP Selection Logic

### Automatic Provider Selection
```go
func selectPSPForProvider(provider string) string {
    switch provider {
    case "MTN":
        // Prefer MTN PSP, fallback to Ogate
        return "mtn" || "ogate"
    case "VODAFONE":
        // Prefer Vodafone PSP, fallback to Ogate
        return "vodafone" || "ogate"
    case "AIRTELTIGO":
        // Use Ogate for AirtelTigo
        return "ogate"
    }
}
```

### Provider Priority
1. **Network-specific PSP** (e.g., MTN PSP for MTN network)
2. **Ogate PSP** (supports all Ghana networks)
3. **First available PSP** (fallback)
4. **Demo PSP** (ultimate fallback)

## API Endpoints

### PSP Management Endpoints
```
GET  /api/v1/psp/providers           - List available PSPs
GET  /api/v1/psp/recommend?provider= - Get recommended PSP for provider
POST /api/v1/psp/test/:psp          - Test PSP connection
```

### Example Responses

#### Available PSPs
```json
{
  "providers": ["ogate", "mtn", "demo", "vodafone_demo"],
  "message": "Available PSP providers"
}
```

#### PSP Recommendation
```json
{
  "provider": "MTN",
  "selectedPSP": "mtn",
  "message": "Recommended PSP for provider"
}
```

#### PSP Connection Test
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

## Configuration

### Environment Variables
```bash
# Ogate PSP
OGATE_BASE_URL=https://api.ogate.com
OGATE_API_KEY=your_ogate_api_key

# MTN PSP
MTN_API_KEY=your_mtn_api_key
MTN_USER_ID=your_mtn_user_id
MTN_BASE_URL=https://sandbox.momodeveloper.mtn.com
MTN_SUBSCRIPTION_KEY=your_mtn_subscription_key

# Additional PSPs can be added here
```

### Default Configuration
- **Default PSP**: Ogate
- **Fallback**: Demo PSP (always available)
- **Auto-initialization**: PSPs initialize based on available credentials

## Transaction Flow

### Collection Flow
1. **PSP Selection**: Select appropriate PSP based on mobile network
2. **Collection Request**: Initiate collection via selected PSP
3. **Status Monitoring**: Monitor collection status
4. **Completion**: Process successful collection

### Delivery Flow
1. **Recipient Type Check**: Determine delivery method
2. **PSP Selection**: Select PSP for mobile money delivery
3. **Delivery Request**: Initiate delivery via selected PSP
4. **Confirmation**: Confirm successful delivery

## Adding New PSPs

### Step 1: Implement PSPProvider Interface
```go
type NewPSP struct {
    apiKey  string
    baseURL string
}

func (n *NewPSP) GetName() string {
    return "newpsp"
}

func (n *NewPSP) InitiateCollection(req CollectionRequest) (*CollectionResponse, error) {
    // Implement collection logic
}

func (n *NewPSP) CheckCollectionStatus(transactionID string) (string, error) {
    // Implement status check logic
}

func (n *NewPSP) InitiateDelivery(req DeliveryRequest) error {
    // Implement delivery logic
}
```

### Step 2: Register in PSPService
```go
func (p *PSPService) initializeProviders() {
    // Add new PSP initialization
    if newPSPAPIKey := os.Getenv("NEWPSP_API_KEY"); newPSPAPIKey != "" {
        p.providers["newpsp"] = NewNewPSP(newPSPAPIKey, baseURL)
    }
}
```

### Step 3: Update Selection Logic
```go
func (p *PSPService) selectPSPForProvider(provider string) string {
    switch provider {
    case "NEWNETWORK":
        if _, exists := p.providers["newpsp"]; exists {
            return "newpsp"
        }
    }
}
```

## Benefits

### 1. **Flexibility**
- Support multiple PSPs simultaneously
- Easy switching between PSPs
- Network-specific optimizations

### 2. **Reliability**
- Automatic fallback mechanisms
- Multiple provider redundancy
- Graceful error handling

### 3. **Scalability**
- Easy addition of new PSPs
- Pluggable architecture
- Independent PSP configurations

### 4. **Cost Optimization**
- Route transactions to cheapest PSP
- Network-specific rate optimization
- Load balancing across PSPs

## Testing

### PSP Connection Testing
```bash
# Test Ogate PSP
curl -X POST "http://localhost:8080/api/v1/psp/test/ogate" \
  -H "Authorization: Bearer $JWT_TOKEN"

# Test MTN PSP
curl -X POST "http://localhost:8080/api/v1/psp/test/mtn" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Integration Testing
```bash
# Get available PSPs
curl "http://localhost:8080/api/v1/psp/providers" \
  -H "Authorization: Bearer $JWT_TOKEN"

# Get PSP recommendation
curl "http://localhost:8080/api/v1/psp/recommend?provider=MTN" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Monitoring & Analytics

### Key Metrics
- **PSP Success Rates**: Track success rate per PSP
- **Response Times**: Monitor PSP response times
- **Cost Analysis**: Compare transaction costs across PSPs
- **Network Performance**: Track performance by mobile network

### Logging
- **PSP Selection**: Log which PSP was selected and why
- **Transaction Flow**: Track complete transaction lifecycle
- **Error Handling**: Log PSP-specific errors and fallbacks

## Security Considerations

### API Key Management
- Store PSP credentials in environment variables
- Use secure key rotation practices
- Monitor API key usage and limits

### Transaction Security
- Validate all PSP responses
- Implement transaction signing where supported
- Monitor for suspicious transaction patterns

### Error Handling
- Never expose PSP credentials in error messages
- Implement proper error logging
- Graceful degradation on PSP failures

## Future Enhancements

### Planned Features
1. **Dynamic PSP Selection**: AI-based PSP selection
2. **Load Balancing**: Distribute load across PSPs
3. **Cost Optimization**: Route based on transaction costs
4. **Real-time Monitoring**: PSP health monitoring dashboard
5. **Webhook Support**: PSP callback handling

### Additional PSPs
- **Vodafone Cash API**: Direct Vodafone integration
- **AirtelTigo Money**: Direct AirtelTigo integration
- **Bank APIs**: Direct bank transfer integration
- **International PSPs**: Cross-border payment support

The multi-PSP architecture provides a robust, scalable foundation for payment processing while maintaining flexibility for future enhancements and integrations.
