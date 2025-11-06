package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PSPProvider interface for payment service providers
type PSPProvider interface {
	GetName() string
	InitiateCollection(req CollectionRequest) (*CollectionResponse, error)
	CheckCollectionStatus(transactionID string) (string, error)
	InitiateDelivery(req DeliveryRequest) error
}

type PSPService struct {
	db        *mongo.Database
	providers map[string]PSPProvider
	defaultPSP string
}

type CollectionRequest struct {
	Amount      float64 `json:"amount"`
	PhoneNumber string  `json:"phoneNumber"`
	Provider    string  `json:"provider"`
	Reference   string  `json:"reference"`
}

type CollectionResponse struct {
	TransactionID string      `json:"transactionId"`
	Status        string      `json:"status"`
	Message       string      `json:"message"`
	PSPName       string      `json:"pspName"`
	RawResponse   interface{} `json:"rawResponse,omitempty"`
}

type DeliveryRequest struct {
	Amount           float64 `json:"amount"`
	RecipientType    string  `json:"recipientType"`
	RecipientAccount string  `json:"recipientAccount"`
	RecipientNetwork string  `json:"recipientNetwork,omitempty"`
	Reference        string  `json:"reference"`
}

func NewPSPService(db *mongo.Database) *PSPService {
	service := &PSPService{
		db:        db,
		providers: make(map[string]PSPProvider),
		defaultPSP: "ogate", // Default to Ogate
	}

	// Initialize PSP providers
	service.initializeProviders()
	return service
}

func (p *PSPService) initializeProviders() {
	// Initialize Ogate PSP
	if ogatePSP := NewOgatePSPFromEnv(); ogatePSP != nil {
		p.providers["ogate"] = ogatePSP
	}

	// Initialize MTN PSP
	if mtnPSP := NewMTNPSPFromEnv(); mtnPSP != nil {
		p.providers["mtn"] = mtnPSP
	}
}

func (p *PSPService) GetProvider(name string) PSPProvider {
	if provider, exists := p.providers[name]; exists {
		return provider
	}
	// Return default provider if specific one not found
	return p.providers[p.defaultPSP]
}

func (p *PSPService) GetAvailableProviders() []string {
	var providers []string
	for name := range p.providers {
		providers = append(providers, name)
	}
	return providers
}

func (p *PSPService) InitiateCollection(req CollectionRequest) (*CollectionResponse, error) {
	// Select PSP based on provider or use default
	pspName := p.selectPSPForProvider(req.Provider)
	provider := p.GetProvider(pspName)
	
	if provider == nil {
		return nil, fmt.Errorf("no PSP available for provider: %s", req.Provider)
	}

	// Track PSP request payload
	p.logPSPRequest(req.Reference, "InitiateCollection", req, pspName)

	response, err := provider.InitiateCollection(req)
	if err != nil {
		// Track failed response
		p.logPSPResponse(req.Reference, "InitiateCollection", nil, err.Error(), pspName)
		return nil, err
	}

	// Track successful response
	p.logPSPResponse(req.Reference, "InitiateCollection", response, "success", pspName)
	
	response.PSPName = pspName
	return response, nil
}

func (p *PSPService) CheckCollectionStatus(transactionID string) (string, error) {
	// For now, use default PSP. In production, you'd track which PSP was used
	provider := p.GetProvider(p.defaultPSP)
	if provider == nil {
		return "failed", fmt.Errorf("no default PSP available")
	}

	return provider.CheckCollectionStatus(transactionID)
}

func (p *PSPService) InitiateDelivery(req DeliveryRequest) error {
	switch req.RecipientType {
	case "mobile_money":
		return p.deliverToMobileMoney(req)
	case "crypto_wallet":
		return p.deliverToCrypto(req)
	case "stellar_wallet":
		return p.deliverToStellar(req)
	case "siha_wallet":
		return p.deliverToSihaWallet(req)
	default:
		return fmt.Errorf("unsupported recipient type: %s", req.RecipientType)
	}
}

func (p *PSPService) selectPSPForProvider(provider string) string {
	// Logic to select appropriate PSP based on mobile network provider
	switch provider {
	case "MTN":
		// Prefer MTN PSP for MTN network, fallback to Ogate
		if _, exists := p.providers["mtn"]; exists {
			return "mtn"
		}
		if _, exists := p.providers["ogate"]; exists {
			return "ogate"
		}
	case "VODAFONE":
		// Use Vodafone-specific PSP if available, fallback to Ogate
		if _, exists := p.providers["vodafone"]; exists {
			return "vodafone"
		}
		if _, exists := p.providers["ogate"]; exists {
			return "ogate"
		}
	case "AIRTELTIGO":
		// Use Ogate for AirtelTigo
		if _, exists := p.providers["ogate"]; exists {
			return "ogate"
		}
	}
	
	// Return first available provider as fallback
	for name := range p.providers {
		return name
	}
	
	return "demo" // Ultimate fallback
}

func (p *PSPService) deliverToMobileMoney(req DeliveryRequest) error {
	// Select PSP for mobile money delivery
	pspName := p.selectPSPForProvider(req.RecipientNetwork)
	provider := p.GetProvider(pspName)
	
	if provider == nil {
		return fmt.Errorf("no PSP available for network: %s", req.RecipientNetwork)
	}

	return provider.InitiateDelivery(req)
}

func (p *PSPService) deliverToStellar(req DeliveryRequest) error {
	// TODO: Integrate with Stellar network to send USDC
	fmt.Printf("Delivering ₵%.2f as USDC to Stellar address: %s\n", req.Amount, req.RecipientAccount)
	
	// In production, this would:
	// 1. Convert GHS amount to USDC using exchange rate
	// 2. Use Stellar SDK to send USDC to recipient address
	// 3. Handle transaction fees and confirmations
	
	return nil
}

func (p *PSPService) deliverToCrypto(req DeliveryRequest) error {
	// TODO: Integrate with crypto wallet APIs
	fmt.Printf("Delivering ₵%.2f to crypto wallet: %s\n", req.Amount, req.RecipientAccount)
	return nil
}

func (p *PSPService) deliverToSihaWallet(req DeliveryRequest) error {
	// TODO: Credit Siha wallet directly
	fmt.Printf("Delivering ₵%.2f to Siha wallet: %s\n", req.Amount, req.RecipientAccount)
	
	// Find user by wallet ID and credit balance
	collection := p.db.Collection("wallets")
	userID, _ := primitive.ObjectIDFromHex(req.RecipientAccount)
	
	_, err := collection.UpdateOne(
		context.Background(),
		map[string]interface{}{"user_id": userID},
		map[string]interface{}{
			"$inc": map[string]interface{}{"balance": req.Amount},
			"$set": map[string]interface{}{"updated_at": time.Now()},
		},
	)
	
	return err
}

// PSP tracking methods
func (p *PSPService) logPSPRequest(transactionRef, operation string, payload interface{}, pspName string) {
	payloadJSON, _ := json.Marshal(payload)
	
	pspLog := bson.M{
		"transaction_reference": transactionRef,
		"operation":            operation,
		"psp_name":             pspName,
		"request_payload":      string(payloadJSON),
		"timestamp":            time.Now(),
		"type":                 "request",
	}
	
	p.db.Collection("psp_logs").InsertOne(context.Background(), pspLog)
	log.Printf("PSP Request [%s] %s: %s", pspName, operation, string(payloadJSON))
}

func (p *PSPService) logPSPResponse(transactionRef, operation string, response interface{}, status, pspName string) {
	responseJSON, _ := json.Marshal(response)
	
	pspLog := bson.M{
		"transaction_reference": transactionRef,
		"operation":            operation,
		"psp_name":             pspName,
		"response_payload":     string(responseJSON),
		"status":               status,
		"timestamp":            time.Now(),
		"type":                 "response",
	}
	
	p.db.Collection("psp_logs").InsertOne(context.Background(), pspLog)
	log.Printf("PSP Response [%s] %s: %s - Status: %s", pspName, operation, string(responseJSON), status)
}
