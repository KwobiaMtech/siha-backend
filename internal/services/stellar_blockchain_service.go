package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"healthy_pay_backend/internal/models"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/network"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// StellarBlockchainService - Stellar implementation of BlockchainService
type StellarBlockchainService struct {
	db                    *mongo.Database
	network               string
	networkPassphrase     string
	horizonClient         *horizonclient.Client
	distributorSecretKey  string
	distributorPublicKey  string
	stellarService 	   *StellarService
}

type CreateWalletRequest struct {
	UserID  primitive.ObjectID `json:"user_id"`
	Network string             `json:"network"`
}

func NewStellarBlockchainService(db interface{}) *StellarBlockchainService {
	stellarNetwork := os.Getenv("STELLAR_NETWORK")
	if stellarNetwork == "" {
		stellarNetwork = "testnet"
	}

	var horizonURL, passphrase string
	if stellarNetwork == "mainnet" {
		horizonURL = "https://horizon.stellar.org"
		passphrase = network.PublicNetworkPassphrase
	} else {
		horizonURL = "https://horizon-testnet.stellar.org"
		passphrase = network.TestNetworkPassphrase
	}

	distributorSecret := os.Getenv("STELLAR_DISTRIBUTOR_SECRET")
	distributorPublic := os.Getenv("STELLAR_DISTRIBUTOR_PUBLIC")

	return &StellarBlockchainService{
		db:                    db.(*mongo.Database),
		network:               stellarNetwork,
		networkPassphrase:     passphrase,
		horizonClient:         &horizonclient.Client{HorizonURL: horizonURL},
		distributorSecretKey:  distributorSecret,
		distributorPublicKey:  distributorPublic,
		stellarService:        NewStellarService(),
	}
}

func (s *StellarBlockchainService) CreateWallet(userID primitive.ObjectID, network string) (*models.BlockchainWallet, error) {
	// Check if blockchain wallet already exists
	var existingWallet models.BlockchainWallet
	err := s.db.Collection("blockchain_wallets").FindOne(
		context.Background(),
		bson.M{"user_id": userID, "blockchain": "stellar", "is_active": true},
	).Decode(&existingWallet)
	
	if err == nil {
		// Wallet already exists, return it
		fmt.Printf("Blockchain wallet already exists for user %s\n", userID.Hex())
		return &existingWallet, nil
	}

	stellarNetwork := os.Getenv("STELLAR_NETWORK")
	if stellarNetwork == "" {
		stellarNetwork = "testnet"
	}
	
	fmt.Printf("Creating Stellar wallet on %s network\n", stellarNetwork)
	
	// Use StellarService - must succeed with trustlines or fail completely
	//stellarService := NewStellarService()
	stellarWallet, err := s.stellarService.CreateActiveAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to create stellar wallet: %v", err)
	}

	// Convert StellarWallet to BlockchainWallet format
	wallet := &models.BlockchainWallet{
		UserID:         userID,
		Blockchain:     "stellar",
		Network:        stellarNetwork,
		PublicKey:      stellarWallet.PublicKey,
		PrivateKey:     stellarWallet.PrivateKey,
		MnemonicPhrase: stellarWallet.MnemonicPhrase,
		IsDefault:      true,
		IsActive:       true,
		Balances: []models.AssetBalance{
			{AssetCode: "XLM", Balance: 0.0, Symbol: "XLM", Name: "Stellar Lumens"},
			{AssetCode: "USDC", Balance: 0.0, Symbol: "USDC", Name: "USD Coin"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fmt.Printf("Saving wallet to blockchain_wallets collection: %s\n", wallet.PublicKey)
	
	result, err := s.db.Collection("blockchain_wallets").InsertOne(context.Background(), wallet)
	if err != nil {
		fmt.Printf("Error inserting blockchain wallet: %v\n", err)
		return nil, fmt.Errorf("failed to save blockchain wallet: %v", err)
	}

	wallet.ID = result.InsertedID.(primitive.ObjectID)
	fmt.Printf("Successfully created blockchain wallet with ID: %s\n", wallet.ID.Hex())
	return wallet, nil
}

func (s *StellarBlockchainService) GetWallet(userID primitive.ObjectID) (*models.BlockchainWallet, error) {
	var wallet models.BlockchainWallet
	err := s.db.Collection("blockchain_wallets").FindOne(
		context.Background(),
		bson.M{"user_id": userID, "blockchain": "stellar", "is_active": true},
	).Decode(&wallet)

	if err != nil {
		return nil, err
	}

	// Update balances from Stellar network
	s.updateWalletBalances(&wallet)
	return &wallet, nil
}

func (s *StellarBlockchainService) GetBalance(walletID primitive.ObjectID, assetCode string) (float64, error) {
	wallet, err := s.getWalletByID(walletID)
	if err != nil {
		return 0, err
	}

	for _, balance := range wallet.Balances {
		if balance.AssetCode == assetCode {
			return balance.Balance, nil
		}
	}

	return 0, fmt.Errorf("asset not found")
}

func (s *StellarBlockchainService) SendAsset(fromWalletID primitive.ObjectID, toAddress string, amount float64, assetCode string, memo string) (*models.BlockchainTransaction, error) {
	// Implementation for sending assets on Stellar
	// This would use the existing Stellar service logic
	return nil, fmt.Errorf("not implemented")
}

func (s *StellarBlockchainService) GetTransactions(walletID primitive.ObjectID) ([]models.BlockchainTransaction, error) {
	var transactions []models.BlockchainTransaction
	cursor, err := s.db.Collection("blockchain_transactions").Find(
		context.Background(),
		bson.M{"wallet_id": walletID, "blockchain": "stellar"},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &transactions)
	return transactions, err
}

func (s *StellarBlockchainService) GetSupportedAssets() ([]models.AssetBalance, error) {
	return []models.AssetBalance{
		{AssetCode: "XLM", Symbol: "XLM", Name: "Stellar Lumens"},
		{AssetCode: "USDC", Symbol: "USDC", Name: "USD Coin"},
	}, nil
}

// Helper methods
func (s *StellarBlockchainService) getWalletByID(walletID primitive.ObjectID) (*models.BlockchainWallet, error) {
	var wallet models.BlockchainWallet
	err := s.db.Collection("blockchain_wallets").FindOne(
		context.Background(),
		bson.M{"_id": walletID, "blockchain": "stellar"},
	).Decode(&wallet)
	return &wallet, err
}

func (s *StellarBlockchainService) updateWalletBalances(wallet *models.BlockchainWallet) {
	fmt.Printf("Updating balances for wallet: %s\n", wallet.PublicKey)
	
	// Get account details from Stellar network
	client := s.getHorizonClient()
	account, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: wallet.PublicKey,
	})
	if err != nil {
		fmt.Printf("Error getting account details: %v\n", err)
		return
	}

	fmt.Printf("Found %d balances on Stellar network\n", len(account.Balances))

	// Update balances from Stellar account
	for i, balance := range wallet.Balances {
		for _, stellarBalance := range account.Balances {
			if balance.AssetCode == "XLM" && stellarBalance.Asset.Type == "native" {
				if xlmBalance, err := strconv.ParseFloat(stellarBalance.Balance, 64); err == nil {
					fmt.Printf("Updating XLM balance from %f to %f\n", wallet.Balances[i].Balance, xlmBalance)
					wallet.Balances[i].Balance = xlmBalance
				}
			} else if balance.AssetCode == "USDC" && stellarBalance.Asset.Type == "credit_alphanum4" && stellarBalance.Asset.Code == "USDC" {
				if usdcBalance, err := strconv.ParseFloat(stellarBalance.Balance, 64); err == nil {
					fmt.Printf("Updating USDC balance from %f to %f\n", wallet.Balances[i].Balance, usdcBalance)
					wallet.Balances[i].Balance = usdcBalance
				}
			}
		}
	}
}

func (s *StellarBlockchainService) getHorizonClient() *horizonclient.Client {
	stellarNetwork := os.Getenv("STELLAR_NETWORK")
	if stellarNetwork == "testnet" {
		return horizonclient.DefaultTestNetClient
	}
	return horizonclient.DefaultPublicNetClient
}


// Placeholder methods for compatibility
func (s *StellarBlockchainService) SendUSDC(userID primitive.ObjectID, req SendUSDCRequest) (interface{}, error) {
	return nil, fmt.Errorf("SendUSDC not implemented in updated service")
}



func (s *StellarBlockchainService) GetUSDCAsset(network string) interface{} {
	return nil
}
