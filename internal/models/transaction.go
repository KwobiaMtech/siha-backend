package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UnifiedTransaction struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID               primitive.ObjectID `bson:"userId" json:"userId"`
	Type                 string             `bson:"type" json:"type"` // "deposit", "send", "receive"
	Amount               float64            `bson:"amount" json:"amount"`
	Status               string             `bson:"status" json:"status"`
	
	// Common fields
	TransactionID        string             `bson:"transactionId" json:"transactionId"`
	PSPReference         string             `bson:"pspReference" json:"pspReference"`
	PSPResponse          interface{}        `bson:"pspResponse" json:"pspResponse"`
	
	// Deposit specific fields
	PaymentMethodID      string             `bson:"paymentMethodId,omitempty" json:"paymentMethodId,omitempty"`
	InvestmentPercentage float64            `bson:"investmentPercentage,omitempty" json:"investmentPercentage,omitempty"`
	DonationChoice       string             `bson:"donationChoice,omitempty" json:"donationChoice,omitempty"`
	
	// Send/Receive specific fields
	RecipientName        string             `bson:"recipientName,omitempty" json:"recipientName,omitempty"`
	RecipientAccount     string             `bson:"recipientAccount,omitempty" json:"recipientAccount,omitempty"`
	RecipientType        string             `bson:"recipientType,omitempty" json:"recipientType,omitempty"`
	RecipientNetwork     string             `bson:"recipientNetwork,omitempty" json:"recipientNetwork,omitempty"`
	SenderName           string             `bson:"senderName,omitempty" json:"senderName,omitempty"`
	
	// Status tracking
	QueueStatus          string             `bson:"queueStatus" json:"queueStatus"`
	ProcessedAt          *time.Time         `bson:"processedAt,omitempty" json:"processedAt,omitempty"`
	CreatedAt            time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type TransactionRequest struct {
	Type                 string  `json:"type" binding:"required"`
	Amount               float64 `json:"amount" binding:"required,gt=0"`
	
	// Deposit fields
	PaymentMethodID      string  `json:"paymentMethodId,omitempty"`
	InvestmentPercentage float64 `json:"investmentPercentage,omitempty"`
	DonationChoice       string  `json:"donationChoice,omitempty"`
	
	// Send fields
	RecipientName        string  `json:"recipientName,omitempty"`
	RecipientAccount     string  `json:"recipientAccount,omitempty"`
	RecipientType        string  `json:"recipientType,omitempty"`
	RecipientNetwork     string  `json:"recipientNetwork,omitempty"`
}

type TransactionResponse struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	Status        string      `json:"status"`
	Message       string      `json:"message"`
	TransactionID string      `json:"transactionId"`
	PSPResponse   interface{} `json:"pspResponse,omitempty"`
}
