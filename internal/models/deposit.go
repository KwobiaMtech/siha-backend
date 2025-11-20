package models

// Deprecated: Use UnifiedTransaction instead
// Keeping only request/response types for backward compatibility

type DepositRequest struct {
	Amount               float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethodID      string  `json:"paymentMethodId" binding:"required"`
	InvestmentPercentage float64 `json:"investmentPercentage" binding:"min=0,max=100"`
	DonationChoice       string  `json:"donationChoice"`
}

type DepositResponse struct {
	ID            string      `json:"id"`
	Status        string      `json:"status"`
	Message       string      `json:"message"`
	TransactionID string      `json:"transactionId"`
	PSPResponse   interface{} `json:"pspResponse,omitempty"`
}
