package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DepositRequest struct {
	Amount               float64 `json:"amount"`
	PaymentMethodID      string  `json:"paymentMethodId"`
	InvestmentPercentage float64 `json:"investmentPercentage"`
	DonationChoice       string  `json:"donationChoice"`
}

type DepositResponse struct {
	ID            string      `json:"id"`
	Status        string      `json:"status"`
	Message       string      `json:"message"`
	TransactionID string      `json:"transactionId"`
	PSPResponse   interface{} `json:"pspResponse,omitempty"`
}

type StatusResponse struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	QueueStatus     string    `json:"queueStatus"`
	Amount          float64   `json:"amount"`
	PaymentMethodID string    `json:"paymentMethodId"`
	ProcessedAt     *string   `json:"processedAt"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func main() {
	// Test deposit initiation with queue tracking
	fmt.Println("🧪 Testing Deposit Queue Integration")
	fmt.Println("=====================================")

	depositReq := DepositRequest{
		Amount:               50.0,
		PaymentMethodID:      "test_mobile_money_id",
		InvestmentPercentage: 30.0,
		DonationChoice:       "profit",
	}

	// Step 1: Initiate deposit
	fmt.Println("\n1. Initiating deposit...")
	jsonData, _ := json.Marshal(depositReq)
	
	resp, err := http.Post(
		"http://localhost:8080/api/v1/deposits/initiate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		fmt.Printf("❌ Error initiating deposit: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var depositResp DepositResponse
	json.NewDecoder(resp.Body).Decode(&depositResp)
	
	fmt.Printf("✅ Deposit initiated: %+v\n", depositResp)

	if depositResp.ID == "" {
		fmt.Println("❌ No deposit ID returned")
		return
	}

	// Step 2: Check initial status
	fmt.Println("\n2. Checking initial status...")
	statusResp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/deposits/%s/status", depositResp.ID))
	if err != nil {
		fmt.Printf("❌ Error checking status: %v\n", err)
		return
	}
	defer statusResp.Body.Close()

	var status StatusResponse
	json.NewDecoder(statusResp.Body).Decode(&status)
	fmt.Printf("📊 Initial status: %+v\n", status)

	// Step 3: Monitor queue processing
	fmt.Println("\n3. Monitoring queue processing...")
	for i := 0; i < 5; i++ {
		time.Sleep(10 * time.Second)
		
		statusResp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/deposits/%s/status", depositResp.ID))
		if err != nil {
			fmt.Printf("❌ Error checking status: %v\n", err)
			continue
		}
		
		json.NewDecoder(statusResp.Body).Decode(&status)
		statusResp.Body.Close()
		
		fmt.Printf("🔄 Status check %d: Status=%s, QueueStatus=%s\n", i+1, status.Status, status.QueueStatus)
		
		if status.Status == "collected" && status.QueueStatus == "completed" {
			fmt.Printf("✅ Deposit successfully processed! ProcessedAt: %v\n", status.ProcessedAt)
			break
		}
		
		if status.Status == "failed" {
			fmt.Printf("❌ Deposit failed with queue status: %s\n", status.QueueStatus)
			break
		}
	}

	fmt.Println("\n🎯 Test completed!")
}
