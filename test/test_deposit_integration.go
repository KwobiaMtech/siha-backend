package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DepositRequest struct {
	Amount               float64 `json:"amount"`
	PaymentMethod        string  `json:"paymentMethod"`
	PhoneNumber          string  `json:"phoneNumber"`
	Network              string  `json:"network"`
	InvestmentPercentage float64 `json:"investmentPercentage"`
	DonationChoice       string  `json:"donationChoice"`
}

func main() {
	// Test deposit initiation
	depositReq := DepositRequest{
		Amount:               100.0,
		PaymentMethod:        "mobile_money",
		PhoneNumber:          "0241234567",
		Network:              "MTN",
		InvestmentPercentage: 20.0,
		DonationChoice:       "profit",
	}

	jsonData, _ := json.Marshal(depositReq)
	
	resp, err := http.Post(
		"http://localhost:8080/api/v1/deposits/initiate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	fmt.Printf("Response: %+v\n", result)
}
