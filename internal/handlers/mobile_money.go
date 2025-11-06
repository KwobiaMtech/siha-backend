package handlers

import (
	"context"
	"net/http"
	"time"

	"healthy_pay_backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MobileMoneyHandler struct {
	db *mongo.Database
}

func NewMobileMoneyHandler(db *mongo.Database) *MobileMoneyHandler {
	return &MobileMoneyHandler{db: db}
}

type AddMobileWalletRequest struct {
	Provider    string `json:"provider" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

func (h *MobileMoneyHandler) AddMobileWallet(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req AddMobileWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Deactivate existing wallets
	collection := h.db.Collection("mobile_money_wallets")
	collection.UpdateMany(
		context.Background(),
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"is_active": false}},
	)

	// Create new wallet
	wallet := models.MobileMoneyWallet{
		UserID:      userID,
		Provider:    req.Provider,
		PhoneNumber: req.PhoneNumber,
		Balance:     850.00, // Default balance for demo
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := collection.InsertOne(context.Background(), wallet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add mobile wallet"})
		return
	}

	wallet.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Mobile wallet added successfully",
		"wallet":  wallet,
	})
}

func (h *MobileMoneyHandler) GetMobileWallets(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := h.db.Collection("mobile_money_wallets")
	cursor, err := collection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wallets"})
		return
	}
	defer cursor.Close(context.Background())

	var wallets []models.MobileMoneyWallet
	if err := cursor.All(context.Background(), &wallets); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode wallets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallets": wallets})
}
