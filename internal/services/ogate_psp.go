package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Currency string

const (
	GHS Currency = "GHS"
	USD Currency = "USD"
	KES Currency = "KES"
)

type TransactionState string

const (
	Initiated     TransactionState = "INITIATED"
	Failed        TransactionState = "FAILED"
	Completed     TransactionState = "COMPLETED"
	Pending       TransactionState = "PENDING"
	Cancelled     TransactionState = "CANCELLED"
	Expired       TransactionState = "EXPIRED"
	Success       TransactionState = "SUCCESS"
	ErrorOccurred TransactionState = "ERROR OCCURRED"
	RequireReview TransactionState = "REQUIRE REVIEW"
)

type OgateResponseStatus string

const (
	OgateInitiated OgateResponseStatus = "INITIATED"
	OgateFailed    OgateResponseStatus = "FAILED"
	OgateCompleted OgateResponseStatus = "COMPLETED"
)

type OgateApiCollectionRequestModel struct {
	Amount        int    `json:"amount"`
	Reason        string `json:"reason"`
	Currency      string `json:"currency"`
	Network       string `json:"network"`
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	Reference     string `json:"reference"`
	CallbackURL   string `json:"callbackUrl,omitempty"`
}

type OgatePayoutsApiRequestModel struct {
	Recipients []struct {
		Network       string `json:"network"`
		Currency      string `json:"currency"`
		Amount        int    `json:"amount"`
		AccountName   string `json:"accountName"`
		AccountNumber string `json:"accountNumber"`
	} `json:"recipients"`
	Reference   string `json:"reference"`
	CallbackURL string `json:"callbackURL,omitempty"`
}

type OgateApiCollectionResponseModel struct {
	ID             string `json:"id"`
	Amount         int    `json:"amount"`
	Currency       string `json:"currency"`
	Status         string `json:"status"`
	Channel        string `json:"channel"`
	Type           string `json:"type"`
	Customer       struct {
		AccountName   string `json:"accountName"`
		AccountNumber string `json:"accountNumber"`
	} `json:"customer"`
	Metadata        interface{} `json:"metadata"`
	Reason          string      `json:"reason"`
	Fee             int         `json:"fee"`
	CallbackURL     string      `json:"callback_url"`
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"updated_at"`
	Reference       string      `json:"reference"`
	Network         string      `json:"network"`
	APIKey          string      `json:"api_key"`
	OpeningBalance  int         `json:"opening_balance"`
	ClosingBalance  int         `json:"closing_balance"`
}

// OgatePSP implements the PSPProvider interface
type OgatePSP struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewOgatePSP(baseURL, apiKey string) *OgatePSP {
	return &OgatePSP{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{},
	}
}

// NewOgatePSPFromEnv creates OgatePSP from environment variables
func NewOgatePSPFromEnv() *OgatePSP {
	baseURL := os.Getenv("OGATE_BASE_URL")
	apiKey := os.Getenv("OGATE_API_KEY")
	
	if baseURL == "" {
		baseURL = "https://api.ogate.com" // Default URL
	}
	
	if apiKey == "" {
		return nil // Return nil if no API key configured
	}
	
	return NewOgatePSP(baseURL, apiKey)
}

func (o *OgatePSP) GetName() string {
	return "ogate"
}

func (o *OgatePSP) InitiateCollection(req CollectionRequest) (*CollectionResponse, error) {
	payload := OgateApiCollectionRequestModel{
		Amount:        int(req.Amount), // Convert to pesewas
		Reason:        fmt.Sprintf("Collection with reference %s", req.Reference),
		Currency:      string(GHS),
		Network:       o.getNetwork(req.Provider),
		AccountName:   "Customer",
		AccountNumber: req.PhoneNumber,
		Reference:     req.Reference,
	}

	rawResponse, err := o.makeRequest("POST", "/collections/mobilemoney", payload)
	if err != nil {
		return nil, err
	}

	responseModel, ok := rawResponse.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format from PSP")
	}
	
	transactionID, ok := responseModel["id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing transaction ID in PSP response")
	}
	
	return &CollectionResponse{
		TransactionID: transactionID,
		Status:        "pending",
		Message:       "Collection initiated successfully",
		RawResponse:   rawResponse, // Store raw API response
	}, nil
}

func (o *OgatePSP) CheckCollectionStatus(transactionID string) (string, error) {
	response, err := o.makeRequest("GET", fmt.Sprintf("/payments/%s", transactionID), nil)
	if err != nil {
		return "failed", err
	}

	responseModel, ok := response.(map[string]interface{})
	if !ok {
		return "failed", fmt.Errorf("invalid response format from PSP")
	}
	
	statusStr, ok := responseModel["status"].(string)
	if !ok {
		return "failed", fmt.Errorf("missing status in PSP response")
	}
	
	status := OgateResponseStatus(statusStr)
	
	switch status {
	case OgateCompleted:
		return "collected", nil
	case OgateFailed:
		return "failed", nil
	default:
		return "pending", nil
	}
}

func (o *OgatePSP) InitiateDelivery(req DeliveryRequest) error {
	payload := OgatePayoutsApiRequestModel{
		Recipients: []struct {
			Network       string `json:"network"`
			Currency      string `json:"currency"`
			Amount        int    `json:"amount"`
			AccountName   string `json:"accountName"`
			AccountNumber string `json:"accountNumber"`
		}{
			{
				Network:       o.getNetwork(req.RecipientNetwork),
				Currency:      string(GHS),
				Amount:        int(req.Amount * 100), // Convert to pesewas
				AccountName:   "Recipient",
				AccountNumber: req.RecipientAccount,
			},
		},
		Reference: req.Reference,
	}

	_, err := o.makeRequest("POST", "/disbursements/mobilemoney", payload)
	return err
}

func (o *OgatePSP) getNetwork(network string) string {
	switch strings.ToUpper(network) {
	case "MTN":
		return "MTN"
	case "AIRTEL", "AIRTELTIGO", "AIRTEL_TIGO", "TIGO":
		return "ATM"
	case "VODAFONE":
		return "VOD"
	default:
		return "MTN" // Default fallback
	}
}

func (o *OgatePSP) makeRequest(method, endpoint string, payload interface{}) (interface{}, error) {
	var body []byte
	var err error
	
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		// Log request payload to endpoint
		log.Printf("PSP Request to %s%s: %s", o.baseURL, endpoint, string(body))
	}

	req, err := http.NewRequest(method, o.baseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body for logging
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// Log response from endpoint
	log.Printf("PSP Response from %s%s: %s", o.baseURL, endpoint, string(responseBody))

	var result interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}
