package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DepositRequest struct {
	Amount               float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethodID      string  `json:"paymentMethodId" binding:"required"`
	InvestmentPercentage float64 `json:"investmentPercentage" binding:"min=0,max=100"`
	DonationChoice       string  `json:"donationChoice"`
}

type Deposit struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID               primitive.ObjectID `bson:"userId" json:"userId"`
	Amount               float64            `bson:"amount" json:"amount"`
	PaymentMethodID      string             `bson:"paymentMethodId" json:"paymentMethodId"`
	InvestmentPercentage float64            `bson:"investmentPercentage" json:"investmentPercentage"`
	DonationChoice       string             `bson:"donationChoice" json:"donationChoice"`
	Status               string             `bson:"status" json:"status"`
	TransactionID        string             `bson:"transactionId" json:"transactionId"`
	PSPReference         string             `bson:"pspReference" json:"pspReference"`
	PSPResponse          interface{}        `bson:"pspResponse" json:"pspResponse"`
	QueueStatus          string             `bson:"queueStatus" json:"queueStatus"`
	ProcessedAt          *time.Time         `bson:"processedAt,omitempty" json:"processedAt,omitempty"`
	CreatedAt            time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type DepositResponse struct {
	ID            string      `json:"id"`
	Status        string      `json:"status"`
	Message       string      `json:"message"`
	TransactionID string      `json:"transactionId"`
	PSPResponse   interface{} `json:"pspResponse,omitempty"`
}
