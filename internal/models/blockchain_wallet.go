package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BlockchainWallet - Generic wallet supporting multiple blockchains
type BlockchainWallet struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"userId"`
	Blockchain      string             `bson:"blockchain" json:"blockchain"` // "stellar", "ethereum", "bitcoin", etc.
	Network         string             `bson:"network" json:"network"`       // "mainnet", "testnet"
	PublicKey       string             `bson:"public_key" json:"publicKey"`
	PrivateKey      string             `bson:"private_key" json:"-"`         // Encrypted, never returned
	MnemonicPhrase  string             `bson:"mnemonic_phrase" json:"mnemonicPhrase"`
	IsDefault       bool               `bson:"is_default" json:"isDefault"`  // Default blockchain wallet
	IsActive        bool               `bson:"is_active" json:"isActive"`
	Balances        []AssetBalance     `bson:"balances" json:"balances"`
	CreatedAt       time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updatedAt"`
}

// AssetBalance - Balance for specific assets on the blockchain
type AssetBalance struct {
	AssetCode   string  `bson:"asset_code" json:"assetCode"`     // "USDC", "XLM", "ETH", "BTC"
	AssetIssuer string  `bson:"asset_issuer,omitempty" json:"assetIssuer,omitempty"` // For tokens
	Balance     float64 `bson:"balance" json:"balance"`
	Symbol      string  `bson:"symbol" json:"symbol"`
	Name        string  `bson:"name" json:"name"`
}

// BlockchainTransaction - Generic transaction supporting multiple blockchains
type BlockchainTransaction struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"userId"`
	WalletID        primitive.ObjectID `bson:"wallet_id" json:"walletId"`
	Blockchain      string             `bson:"blockchain" json:"blockchain"`
	Network         string             `bson:"network" json:"network"`
	FromAddress     string             `bson:"from_address" json:"fromAddress"`
	ToAddress       string             `bson:"to_address" json:"toAddress"`
	Amount          float64            `bson:"amount" json:"amount"`
	AssetCode       string             `bson:"asset_code" json:"assetCode"`
	AssetIssuer     string             `bson:"asset_issuer,omitempty" json:"assetIssuer,omitempty"`
	TxHash          string             `bson:"tx_hash" json:"txHash"`
	Status          string             `bson:"status" json:"status"` // "pending", "confirmed", "failed"
	Type            string             `bson:"type" json:"type"`     // "send", "receive", "deposit", "withdraw"
	Memo            string             `bson:"memo,omitempty" json:"memo,omitempty"`
	Fee             float64            `bson:"fee" json:"fee"`
	CreatedAt       time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updatedAt"`
}

// SupportedBlockchain - Configuration for supported blockchains
type SupportedBlockchain struct {
	Name         string   `json:"name"`
	Code         string   `json:"code"`
	Networks     []string `json:"networks"`
	DefaultAsset string   `json:"defaultAsset"`
	IsDefault    bool     `json:"isDefault"`
}

// GetSupportedBlockchains - Returns list of supported blockchains
func GetSupportedBlockchains() []SupportedBlockchain {
	return []SupportedBlockchain{
		{
			Name:         "Stellar",
			Code:         "stellar",
			Networks:     []string{"mainnet", "testnet"},
			DefaultAsset: "USDC",
			IsDefault:    true, // Stellar is default
		},
		{
			Name:         "Ethereum",
			Code:         "ethereum",
			Networks:     []string{"mainnet", "goerli", "sepolia"},
			DefaultAsset: "USDC",
			IsDefault:    false,
		},
		{
			Name:         "Bitcoin",
			Code:         "bitcoin",
			Networks:     []string{"mainnet", "testnet"},
			DefaultAsset: "BTC",
			IsDefault:    false,
		},
	}
}

// GetDefaultBlockchain - Returns the default blockchain (Stellar)
func GetDefaultBlockchain() SupportedBlockchain {
	blockchains := GetSupportedBlockchains()
	for _, bc := range blockchains {
		if bc.IsDefault {
			return bc
		}
	}
	return blockchains[0] // Fallback to first if no default found
}
