package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StellarWallet struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"userId"`
	PublicKey       string             `bson:"public_key" json:"publicKey"`
	PrivateKey      string             `bson:"private_key" json:"-"` // Encrypted, never returned
	MnemonicPhrase  string             `bson:"mnemonic_phrase" json:"-"` // Encrypted, never returned
	Network         string             `bson:"network" json:"network"`
	USDCBalance     float64            `bson:"usdc_balance" json:"usdcBalance"`
	XLMBalance      float64            `bson:"xlm_balance" json:"xlmBalance"`
	IsActive        bool               `bson:"is_active" json:"isActive"`
	CreatedAt       time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updatedAt"`
}
