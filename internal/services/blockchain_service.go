package services

import (
	"healthy_pay_backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BlockchainService - Interface for blockchain operations
type BlockchainService interface {
	CreateWallet(userID primitive.ObjectID, network string) (*models.BlockchainWallet, error)
	GetWallet(userID primitive.ObjectID) (*models.BlockchainWallet, error)
	GetBalance(walletID primitive.ObjectID, assetCode string) (float64, error)
	SendAsset(fromWalletID primitive.ObjectID, toAddress string, amount float64, assetCode string, memo string) (*models.BlockchainTransaction, error)
	GetTransactions(walletID primitive.ObjectID) ([]models.BlockchainTransaction, error)
	GetSupportedAssets() ([]models.AssetBalance, error)
}

// BlockchainServiceFactory - Factory to create blockchain services
type BlockchainServiceFactory struct {
	db interface{}
}

func NewBlockchainServiceFactory(db interface{}) *BlockchainServiceFactory {
	return &BlockchainServiceFactory{db: db}
}

func (f *BlockchainServiceFactory) GetService(blockchain string) BlockchainService {
	switch blockchain {
	case "stellar":
		return NewStellarBlockchainService(f.db)
	case "ethereum":
		// TODO: Implement Ethereum service
		return nil
	case "bitcoin":
		// TODO: Implement Bitcoin service
		return nil
	default:
		// Default to Stellar
		return NewStellarBlockchainService(f.db)
	}
}

// GetDefaultBlockchainService - Returns the default blockchain service (Stellar)
func (f *BlockchainServiceFactory) GetDefaultBlockchainService() BlockchainService {
	defaultBC := models.GetDefaultBlockchain()
	return f.GetService(defaultBC.Code)
}
