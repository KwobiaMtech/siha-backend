package handlers

import (
	"net/http"

	"healthy_pay_backend/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)




type StellarWalletHandler struct {
	stellarService *services.StellarBlockchainService
}

func NewStellarWalletHandler(db *mongo.Database) *StellarWalletHandler {
	return &StellarWalletHandler{
		stellarService: services.NewStellarBlockchainService(db),
	}
}

// CreateWallet creates a new Stellar wallet for the user
func (h *StellarWalletHandler) CreateWallet(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req services.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to testnet if no network specified
		req.Network = "testnet"
	}

	// Validate network
	if req.Network != "testnet" && req.Network != "mainnet" {
		req.Network = "testnet"
	}

	wallet, err := h.stellarService.CreateWallet(userID, req.Network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wallet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stellar wallet created successfully",
		"wallet":  wallet,
	})
}

// GetWallet retrieves user's Stellar wallet
func (h *StellarWalletHandler) GetWallet(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	wallet, err := h.stellarService.GetWallet(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallet": wallet})
}

// SendUSDC sends USDC to another Stellar address
func (h *StellarWalletHandler) SendUSDC(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req services.SendUSDCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction, err := h.stellarService.SendUSDC(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "USDC sent successfully",
		"transaction": transaction,
	})
}

// GetTransactions retrieves user's Stellar transactions
func (h *StellarWalletHandler) GetTransactions(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	transactions, err := h.stellarService.GetTransactions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// GetAssetInfo returns USDC asset information
func (h *StellarWalletHandler) GetAssetInfo(c *gin.Context) {
	network := c.Query("network")
	if network == "" {
		network = "testnet"
	}

	asset := h.stellarService.GetUSDCAsset(network)
	
	c.JSON(http.StatusOK, gin.H{
		"network": network,
		"asset":   asset,
	})
}



// GetWalletInfo returns wallet capabilities and network info (existing method)
func (h *StellarWalletHandler) GetWalletInfo(c *gin.Context) {
	info := gin.H{
		"network": "STELLAR",
		"defaultAsset": gin.H{
			"code":   "USDC",
			"name":   "USD Coin",
			"symbol": "USDC",
		},
		"supportedAssets": []gin.H{
			{
				"code":   "USDC",
				"name":   "USD Coin",
				"symbol": "USDC",
			},
			{
				"code":   "XLM",
				"name":   "Stellar Lumens",
				"symbol": "XLM",
			},
		},
		"networks": []string{"testnet", "mainnet"},
		"features": []string{
			"send",
			"receive", 
			"balance_check",
			"transaction_history",
		},
	}

	c.JSON(http.StatusOK, info)
}
