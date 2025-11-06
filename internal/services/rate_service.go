package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RateService struct {
	client *http.Client
}

type ExchangeRateResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

type ConversionRate struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
	Timestamp    string  `json:"timestamp"`
}

func NewRateService() *RateService {
	return &RateService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (rs *RateService) GetExchangeRates() (*ExchangeRateResponse, error) {
	resp, err := rs.client.Get("https://cdn.moneyconvert.net/api/latest.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var rateResp ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &rateResp, nil
}

func (rs *RateService) GetOnrampRate(fromCurrency string) (*ConversionRate, error) {
	rates, err := rs.GetExchangeRates()
	if err != nil {
		return nil, err
	}

	rate, exists := rates.Rates[fromCurrency]
	if !exists {
		return nil, fmt.Errorf("currency %s not found", fromCurrency)
	}

	return &ConversionRate{
		FromCurrency: fromCurrency,
		ToCurrency:   "USD",
		Rate:         1 / rate, // Convert to USD
		Timestamp:    time.Now().Format(time.RFC3339),
	}, nil
}

func (rs *RateService) GetOfframpRate(toCurrency string) (*ConversionRate, error) {
	rates, err := rs.GetExchangeRates()
	if err != nil {
		return nil, err
	}

	rate, exists := rates.Rates[toCurrency]
	if !exists {
		return nil, fmt.Errorf("currency %s not found", toCurrency)
	}

	return &ConversionRate{
		FromCurrency: "USD",
		ToCurrency:   toCurrency,
		Rate:         rate, // Convert from USD
		Timestamp:    time.Now().Format(time.RFC3339),
	}, nil
}

func (rs *RateService) ConvertAmount(amount float64, fromCurrency, toCurrency string) (float64, error) {
	rates, err := rs.GetExchangeRates()
	if err != nil {
		return 0, err
	}

	if fromCurrency == "USD" {
		rate, exists := rates.Rates[toCurrency]
		if !exists {
			return 0, fmt.Errorf("currency %s not found", toCurrency)
		}
		return amount * rate, nil
	}

	if toCurrency == "USD" {
		rate, exists := rates.Rates[fromCurrency]
		if !exists {
			return 0, fmt.Errorf("currency %s not found", fromCurrency)
		}
		return amount / rate, nil
	}

	// Convert via USD
	fromRate, fromExists := rates.Rates[fromCurrency]
	toRate, toExists := rates.Rates[toCurrency]
	
	if !fromExists {
		return 0, fmt.Errorf("currency %s not found", fromCurrency)
	}
	if !toExists {
		return 0, fmt.Errorf("currency %s not found", toCurrency)
	}

	usdAmount := amount / fromRate
	return usdAmount * toRate, nil
}
