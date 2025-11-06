package entities

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email       string             `bson:"email" json:"email"`
	Password    string             `bson:"password" json:"-"`
	FirstName   string             `bson:"first_name" json:"firstName"`
	LastName    string             `bson:"last_name" json:"lastName"`
	PhoneNumber string             `bson:"phone_number" json:"phoneNumber"`
	IsVerified  bool               `bson:"is_verified" json:"isVerified"`
	KYCStatus   string             `bson:"kyc_status" json:"kycStatus"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
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
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FromUserID  primitive.ObjectID `bson:"from_user_id" json:"fromUserId"`
	ToUserID    primitive.ObjectID `bson:"to_user_id" json:"toUserId"`
	Amount      float64            `bson:"amount" json:"amount"`
	Type        string             `bson:"type" json:"type"`
	Status      string             `bson:"status" json:"status"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
}
