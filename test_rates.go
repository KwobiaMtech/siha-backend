package main

import (
	"encoding/json"
	"fmt"
	"healthy_pay_backend/internal/services"
)

func main() {
	fmt.Println("=== Testing Rate Conversion Service ===\n")
	
	rateService := services.NewRateService()
	
	// Test 1: Get all exchange rates
	fmt.Println("1. Getting all exchange rates:")
	rates, err := rateService.GetExchangeRates()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Base Currency: %s\n", rates.Base)
	fmt.Printf("Date: %s\n", rates.Date)
	fmt.Printf("Sample rates - GHS: %.6f, KES: %.6f, ZMW: %.6f\n\n", 
		rates.Rates["GHS"], rates.Rates["KES"], rates.Rates["ZMW"])
	
	// Test 2: Get onramp rate (GHS to USD)
	fmt.Println("2. Onramp Rate (GHS → USD):")
	onrampRate, err := rateService.GetOnrampRate("GHS")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	onrampJSON, _ := json.MarshalIndent(onrampRate, "", "  ")
	fmt.Printf("Response: %s\n\n", onrampJSON)
	
	// Test 3: Get offramp rate (USD to GHS)
	fmt.Println("3. Offramp Rate (USD → GHS):")
	offrampRate, err := rateService.GetOfframpRate("GHS")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	offrampJSON, _ := json.MarshalIndent(offrampRate, "", "  ")
	fmt.Printf("Response: %s\n\n", offrampJSON)
	
	// Test 4: Convert amount (100 GHS to USD)
	fmt.Println("4. Convert 100 GHS to USD:")
	convertedAmount, err := rateService.ConvertAmount(100, "GHS", "USD")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("100 GHS = %.2f USD\n\n", convertedAmount)
	
	// Test 5: Convert amount (50 USD to GHS)
	fmt.Println("5. Convert 50 USD to GHS:")
	convertedAmount2, err := rateService.ConvertAmount(50, "USD", "GHS")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("50 USD = %.2f GHS\n\n", convertedAmount2)
	
	// Test 6: Cross-currency conversion (100 GHS to KES)
	fmt.Println("6. Convert 100 GHS to KES:")
	convertedAmount3, err := rateService.ConvertAmount(100, "GHS", "KES")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("100 GHS = %.2f KES\n", convertedAmount3)
}
