package services

import (
	"fmt"
	"os"
	"time"
)

// DemoPSP implements PSPProvider interface for testing/demo purposes
type DemoPSP struct {
	name string
}

func NewDemoPSP(name string) *DemoPSP {
	return &DemoPSP{name: name}
}

// NewDemoPSPFromEnv creates DemoPSP from environment variables
func NewDemoPSPFromEnv() *DemoPSP {
	name := os.Getenv("DEMO_PSP_NAME")
	if name == "" {
		name = "demo" // Default name
	}
	return NewDemoPSP(name)
}

func (d *DemoPSP) GetName() string {
	return d.name
}

func (d *DemoPSP) InitiateCollection(req CollectionRequest) (*CollectionResponse, error) {
	// Simulate collection initiation
	transactionID := fmt.Sprintf("%s_%d", d.name, time.Now().Unix())
	
	return &CollectionResponse{
		TransactionID: transactionID,
		Status:        "pending",
		Message:       fmt.Sprintf("%s demo collection initiated", d.name),
	}, nil
}

func (d *DemoPSP) CheckCollectionStatus(transactionID string) (string, error) {
	// Simulate successful collection after delay
	return "collected", nil
}

func (d *DemoPSP) InitiateDelivery(req DeliveryRequest) error {
	// Simulate delivery
	fmt.Printf("%s PSP: Delivering ₵%.2f to %s\n", d.name, req.Amount, req.RecipientAccount)
	return nil
}
