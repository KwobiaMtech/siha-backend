package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentMethod struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userId"`
	Type        string             `bson:"type" json:"type"` // 'mobile_money', 'bank_card', 'wallet'
	
	// Mobile Money fields
	Provider    string `bson:"provider,omitempty" json:"provider,omitempty"`
	Network     string `bson:"network,omitempty" json:"network,omitempty"`
	PhoneNumber string `bson:"phone_number,omitempty" json:"phoneNumber,omitempty"`
	AccountName string `bson:"account_name,omitempty" json:"accountName,omitempty"`
	Currency    string `bson:"currency,omitempty" json:"currency,omitempty"`
	
	// Bank Card fields
	CardNumber string `bson:"card_number,omitempty" json:"cardNumber,omitempty"`
	CardHolder string `bson:"card_holder,omitempty" json:"cardHolder,omitempty"`
	ExpiryDate string `bson:"expiry_date,omitempty" json:"expiryDate,omitempty"`
	
	// Common fields
	IsDefault bool      `bson:"is_default" json:"isDefault"`
	IsActive  bool      `bson:"is_active" json:"isActive"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}
