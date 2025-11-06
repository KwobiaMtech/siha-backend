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

type InvestmentHandler struct {
	db *mongo.Database
}

func NewInvestmentHandler(db *mongo.Database) *InvestmentHandler {
	return &InvestmentHandler{db: db}
}

func (h *InvestmentHandler) CreateInvestment(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
		Type   string  `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check wallet balance
	walletCollection := h.db.Collection("wallets")
	var wallet models.Wallet
	err = walletCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&wallet)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	if wallet.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Create investment
	investment := models.Investment{
		UserID:    userID,
		Amount:    req.Amount,
		Type:      req.Type,
		Status:    "active",
		Returns:   0.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := h.db.Collection("investments")
	_, err = collection.InsertOne(context.Background(), investment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create investment"})
		return
	}

	// Deduct from wallet
	_, err = walletCollection.UpdateOne(
		context.Background(),
		bson.M{"user_id": userID},
		bson.M{"$inc": bson.M{"balance": -req.Amount}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Investment created successfully"})
}

func (h *InvestmentHandler) GetInvestments(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := h.db.Collection("investments")
	cursor, err := collection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch investments"})
		return
	}
	defer cursor.Close(context.Background())

	var investments []models.Investment
	if err := cursor.All(context.Background(), &investments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode investments"})
		return
	}

	c.JSON(http.StatusOK, investments)
}
