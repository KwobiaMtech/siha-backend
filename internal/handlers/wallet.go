package handlers

import (
	"context"
	"net/http"

	"healthy_pay_backend/internal/models"
	"healthy_pay_backend/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WalletHandler struct {
	db                       *mongo.Database
	blockchainServiceFactory *services.BlockchainServiceFactory
}

func NewWalletHandler(db *mongo.Database) *WalletHandler {
	return &WalletHandler{
		db:                       db,
		blockchainServiceFactory: services.NewBlockchainServiceFactory(db),
	}
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Always try to get onchain wallet first (Stellar)
	blockchainService := h.blockchainServiceFactory.GetService("stellar")
	if blockchainService != nil {
		blockchainWallet, err := blockchainService.GetWallet(userID)
		if err == nil && blockchainWallet.IsActive {
			// Return onchain balance directly from blockchain
			c.JSON(http.StatusOK, gin.H{
				"type":       "onchain",
				"blockchain": "stellar",
				"wallet":     blockchainWallet,
				"status":     "active",
			})
			return
		} else if err == nil && !blockchainWallet.IsActive {
			// Wallet exists but not activated
			c.JSON(http.StatusOK, gin.H{
				"type":       "onchain",
				"blockchain": "stellar", 
				"wallet": map[string]interface{}{
					"id":        blockchainWallet.ID,
					"userId":    blockchainWallet.UserID,
					"publicKey": blockchainWallet.PublicKey,
					"balance":   0,
					"currency":  "USDC",
					"isActive":  false,
				},
				"status":     "inactive",
				"message":    "Activate Wallet",
			})
			return
		}
	}

	// No onchain wallet available - show activation needed
	c.JSON(http.StatusOK, gin.H{
		"type":    "onchain",
		"blockchain": "stellar",
		"wallet": map[string]interface{}{
			"balance":  0,
			"currency": "USDC",
			"isActive": false,
		},
		"status":  "not_created",
		"message": "Activate Wallet",
	})
}

func (h *WalletHandler) CreateBlockchainWallet(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Blockchain string `json:"blockchain"`
		Network    string `json:"network"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to Stellar if not specified
	if req.Blockchain == "" {
		req.Blockchain = "stellar"
	}
	if req.Network == "" {
		req.Network = "testnet"
	}

	blockchainService := h.blockchainServiceFactory.GetService(req.Blockchain)
	if blockchainService == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported blockchain"})
		return
	}

	wallet, err := blockchainService.CreateWallet(userID, req.Network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blockchain wallet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blockchain wallet created successfully",
		"wallet":  wallet,
	})
}

func (h *WalletHandler) GetSupportedBlockchains(c *gin.Context) {
	blockchains := models.GetSupportedBlockchains()
	c.JSON(http.StatusOK, gin.H{
		"blockchains": blockchains,
	})
}

func (h *WalletHandler) AddFunds(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Amount     float64 `json:"amount" binding:"required,gt=0"`
		Blockchain string  `json:"blockchain"`
		AssetCode  string  `json:"assetCode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to traditional wallet if no blockchain specified
	if req.Blockchain == "" {
		collection := h.db.Collection("wallets")
		update := bson.M{
			"$inc": bson.M{"balance": req.Amount},
		}

		_, err = collection.UpdateOne(context.Background(), bson.M{"user_id": userID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add funds"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Funds added successfully"})
		return
	}

	// Handle blockchain wallet funding
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Blockchain wallet funding not implemented"})
}

func (h *WalletHandler) ActivateBlockchain(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Blockchain   string `json:"blockchain" binding:"required"`
		Stablecoin   string `json:"stablecoin,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if wallet already exists
	var existingWallet models.BlockchainWallet
	err = h.db.Collection("blockchain_wallets").FindOne(
		context.Background(),
		bson.M{"user_id": userID, "blockchain": req.Blockchain},
	).Decode(&existingWallet)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Blockchain wallet already exists"})
		return
	}

	// Create new blockchain wallet
	blockchainServiceFactory := services.NewBlockchainServiceFactory(h.db)
	blockchainService := blockchainServiceFactory.GetService(req.Blockchain)

	if blockchainService == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported blockchain"})
		return
	}

	wallet, err := blockchainService.CreateWallet(userID, req.Blockchain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wallet"})
		return
	}

	// Add stablecoin if specified
	if req.Stablecoin != "" {
		wallet.Balances = []models.AssetBalance{
			{
				AssetCode: req.Stablecoin,
				Symbol:    req.Stablecoin,
				Name:      getAssetName(req.Stablecoin),
				Balance:   0.0,
			},
		}
		h.db.Collection("blockchain_wallets").UpdateOne(
			context.Background(),
			bson.M{"_id": wallet.ID},
			bson.M{"$set": bson.M{"balances": wallet.Balances}},
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blockchain wallet activated successfully",
		"wallet":  wallet,
	})
}
