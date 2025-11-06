package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email            string             `bson:"email" json:"email"`
	Password         string             `bson:"password" json:"-"`
	FirstName        string             `bson:"first_name" json:"firstName"`
	LastName         string             `bson:"last_name" json:"lastName"`
	PhoneNumber      string             `bson:"phone_number,omitempty" json:"phoneNumber,omitempty"`
	PIN              string             `bson:"pin" json:"-"`
	IsVerified       bool               `bson:"is_verified" json:"isVerified"`
	VerificationCode string             `bson:"verification_code,omitempty" json:"-"`
	KYCStatus        string             `bson:"kyc_status" json:"kycStatus"`
	CreatedAt        time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updatedAt"`
}

type Wallet struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"userId"`
	Balance   float64            `bson:"balance" json:"balance"`
	Currency  string             `bson:"currency" json:"currency"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}

type Transaction struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FromUserID           primitive.ObjectID `bson:"from_user_id" json:"fromUserId"`
	ToUserID             primitive.ObjectID `bson:"to_user_id,omitempty" json:"toUserId,omitempty"`
	RecipientName        string             `bson:"recipient_name,omitempty" json:"recipientName,omitempty"`
	RecipientAccount     string             `bson:"recipient_account,omitempty" json:"recipientAccount,omitempty"`
	RecipientType        string             `bson:"recipient_type,omitempty" json:"recipientType,omitempty"` // 'mobile_money', 'crypto_wallet', 'siha_wallet'
	RecipientNetwork     string             `bson:"recipient_network,omitempty" json:"recipientNetwork,omitempty"` // 'MTN', 'TELECEL', 'AIRTELTIGO', 'MPESA', etc.
	RecipientCurrency    string             `bson:"recipient_currency,omitempty" json:"recipientCurrency,omitempty"` // 'GHS', 'USD', 'KES', 'ZMW', etc.
	Amount               float64            `bson:"amount" json:"amount"`
	InvestmentAmount     float64            `bson:"investment_amount,omitempty" json:"investmentAmount,omitempty"`
	InvestmentPercentage float64            `bson:"investment_percentage,omitempty" json:"investmentPercentage,omitempty"`
	DonationChoice       string             `bson:"donation_choice,omitempty" json:"donationChoice,omitempty"`
	PaymentMethod        string             `bson:"payment_method" json:"paymentMethod"`
	PSPTransactionID     string             `bson:"psp_transaction_id,omitempty" json:"pspTransactionId,omitempty"`
	PSPRequest           interface{}        `bson:"psp_request,omitempty" json:"pspRequest,omitempty"`
	PSPResponse          interface{}        `bson:"psp_response,omitempty" json:"pspResponse,omitempty"`
	CollectionStatus     string             `bson:"collection_status,omitempty" json:"collectionStatus,omitempty"` // 'pending', 'collected', 'failed'
	InvestmentStatus     string             `bson:"investment_status,omitempty" json:"investmentStatus,omitempty"` // 'pending', 'allocated', 'failed'
	DeliveryStatus       string             `bson:"delivery_status,omitempty" json:"deliveryStatus,omitempty"`     // 'pending', 'delivered', 'failed'
	Type                 string             `bson:"type" json:"type"`
	Status               string             `bson:"status" json:"status"`
	Description          string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt            time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updatedAt"`
}

type UserPaymentMethod struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"userId"`
	Type         string             `bson:"type" json:"type"` // 'mobile_money', 'bank_account'
	Provider     string             `bson:"provider,omitempty" json:"provider,omitempty"`
	Network      string             `bson:"network,omitempty" json:"network,omitempty"`
	PhoneNumber  string             `bson:"phone_number,omitempty" json:"phoneNumber,omitempty"`
	AccountName  string             `bson:"account_name,omitempty" json:"accountName,omitempty"`
	Currency     string             `bson:"currency,omitempty" json:"currency,omitempty"`
	IsDefault    bool               `bson:"is_default" json:"isDefault"`
	IsActive     bool               `bson:"is_active" json:"isActive"`
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updatedAt"`
}

type Investment struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID             primitive.ObjectID `bson:"user_id" json:"userId"`
	Amount             float64            `bson:"amount" json:"amount"`
	InvestmentCurrency string             `bson:"investment_currency" json:"investmentCurrency"`
	Type               string             `bson:"type" json:"type"`
	Status             string             `bson:"status" json:"status"`
	Returns            float64            `bson:"returns" json:"returns"`
	Rate               *ExchangeRate      `bson:"rate,omitempty" json:"rate,omitempty"`
	CreatedAt          time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updatedAt"`
}

type ExchangeRate struct {
	FromCurrency string    `bson:"from_currency" json:"fromCurrency"`
	ToCurrency   string    `bson:"to_currency" json:"toCurrency"`
	Rate         float64   `bson:"rate" json:"rate"`
	Timestamp    time.Time `bson:"timestamp" json:"timestamp"`
}

type Recipient struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"userId"`
	Name          string             `bson:"name" json:"name"`
	Account       string             `bson:"account" json:"account"`
	Type          string             `bson:"type" json:"type"` // 'mobile' or 'bank'
	IsFrequent    bool               `bson:"is_frequent" json:"isFrequent"`
	CreatedAt     time.Time          `bson:"created_at" json:"createdAt"`
}

type MobileMoneyWallet struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userId"`
	Provider    string             `bson:"provider" json:"provider"` // MTN, Vodafone, AirtelTigo
	PhoneNumber string             `bson:"phone_number" json:"phoneNumber"`
	Balance     float64            `bson:"balance" json:"balance"`
	IsActive    bool               `bson:"is_active" json:"isActive"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}
