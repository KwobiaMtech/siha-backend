package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"healthy_pay_backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaymentMethodsHandler struct {
	db *mongo.Database
}

func NewPaymentMethodsHandler(db *mongo.Database) *PaymentMethodsHandler {
	return &PaymentMethodsHandler{db: db}
}

type AddPaymentMethodRequest struct {
	Type         string `json:"type" binding:"required"` // 'mobile_money', 'bank_card', 'wallet'
	Provider     string `json:"provider,omitempty"`
	Network      string `json:"network,omitempty"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	AccountName  string `json:"accountName,omitempty"`
	Currency     string `json:"currency,omitempty"`
	CardNumber   string `json:"cardNumber,omitempty"`
	CardHolder   string `json:"cardHolder,omitempty"`
	ExpiryDate   string `json:"expiryDate,omitempty"`
	IsDefault    bool   `json:"isDefault"`
}

// Get all payment methods for a user
func (h *PaymentMethodsHandler) GetPaymentMethods(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var paymentMethods []map[string]interface{}

	// Get all payment methods from unified collection
	cursor, err := h.db.Collection("payment_methods").Find(
		context.Background(),
		bson.M{"user_id": userID, "is_active": true},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment methods"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var method models.PaymentMethod
		if err := cursor.Decode(&method); err == nil {
			methodData := map[string]interface{}{
				"id":        method.ID.Hex(),
				"type":      method.Type,
				"isDefault": method.IsDefault,
				"isActive":  method.IsActive,
				"createdAt": method.CreatedAt,
				"updatedAt": method.UpdatedAt,
			}

			switch method.Type {
			case "mobile_money":
				// Mask phone number for security
				maskedPhone := method.PhoneNumber
				if len(method.PhoneNumber) > 4 {
					maskedPhone = method.PhoneNumber[:4] + "****" + method.PhoneNumber[len(method.PhoneNumber)-3:]
				}
				
				methodData["title"] = fmt.Sprintf("📱 %s Mobile Money", method.Network)
				methodData["subtitle"] = fmt.Sprintf("%s - %s", method.Network, maskedPhone)
				methodData["provider"] = method.Network
				methodData["network"] = method.Network
				methodData["phoneNumber"] = method.PhoneNumber
				methodData["accountName"] = method.AccountName
				methodData["currency"] = method.Currency
				methodData["hasBalance"] = false

			case "bank_card":
				// Mask card number for security
				maskedCard := method.CardNumber
				if len(method.CardNumber) > 4 {
					maskedCard = "****" + method.CardNumber[len(method.CardNumber)-4:]
				}
				
				methodData["title"] = fmt.Sprintf("💳 %s Card", method.Provider)
				methodData["subtitle"] = fmt.Sprintf("%s - %s", method.Provider, maskedCard)
				methodData["provider"] = method.Provider
				methodData["cardNumber"] = maskedCard
				methodData["cardHolder"] = method.CardHolder
				methodData["expiryDate"] = method.ExpiryDate
				methodData["hasBalance"] = false

			case "wallet":
				// Get wallet balance
				balance, _ := h.getWalletBalance(userID)
				methodData["title"] = "💰 Wallet Balance"
				methodData["subtitle"] = "Send from your platform wallet"
				methodData["balance"] = balance
				methodData["hasBalance"] = true
			}

			paymentMethods = append(paymentMethods, methodData)
		}
	}

	// Always add wallet balance if not already present
	hasWallet := false
	for _, method := range paymentMethods {
		if method["type"] == "wallet" {
			hasWallet = true
			break
		}
	}

	if !hasWallet {
		balance, _ := h.getWalletBalance(userID)
		paymentMethods = append(paymentMethods, map[string]interface{}{
			"id":         "wallet_balance",
			"title":      "💰 Wallet Balance",
			"subtitle":   "Send from your platform wallet",
			"balance":    balance,
			"type":       "wallet",
			"hasBalance": true,
			"isDefault":  false,
			"isActive":   true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"paymentMethods": paymentMethods,
		"totalMethods":   len(paymentMethods),
	})
}

// Add a new payment method
func (h *PaymentMethodsHandler) AddPaymentMethod(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If this is set as default, unset other defaults
	if req.IsDefault {
		h.db.Collection("payment_methods").UpdateMany(
			context.Background(),
			bson.M{"user_id": userID, "type": req.Type},
			bson.M{"$set": bson.M{"is_default": false, "updated_at": time.Now()}},
		)
	}

	// Create payment method
	paymentMethod := models.PaymentMethod{
		UserID:      userID,
		Type:        req.Type,
		Provider:    req.Provider,
		Network:     req.Network,
		PhoneNumber: req.PhoneNumber,
		AccountName: req.AccountName,
		Currency:    req.Currency,
		CardNumber:  req.CardNumber,
		CardHolder:  req.CardHolder,
		ExpiryDate:  req.ExpiryDate,
		IsDefault:   req.IsDefault,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := h.db.Collection("payment_methods").InsertOne(context.Background(), paymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add payment method"})
		return
	}

	paymentMethod.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusOK, gin.H{
		"message":       "Payment method added successfully",
		"paymentMethod": paymentMethod,
	})
}

// Update payment method
func (h *PaymentMethodsHandler) UpdatePaymentMethod(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	methodIDStr := c.Param("id")
	methodID, err := primitive.ObjectIDFromHex(methodIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}

	var req AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If this is set as default, unset other defaults
	if req.IsDefault {
		h.db.Collection("payment_methods").UpdateMany(
			context.Background(),
			bson.M{"user_id": userID, "type": req.Type, "_id": bson.M{"$ne": methodID}},
			bson.M{"$set": bson.M{"is_default": false, "updated_at": time.Now()}},
		)
	}

	// Update payment method
	update := bson.M{
		"$set": bson.M{
			"provider":     req.Provider,
			"network":      req.Network,
			"phone_number": req.PhoneNumber,
			"account_name": req.AccountName,
			"currency":     req.Currency,
			"card_number":  req.CardNumber,
			"card_holder":  req.CardHolder,
			"expiry_date":  req.ExpiryDate,
			"is_default":   req.IsDefault,
			"updated_at":   time.Now(),
		},
	}

	result, err := h.db.Collection("payment_methods").UpdateOne(
		context.Background(),
		bson.M{"_id": methodID, "user_id": userID},
		update,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment method"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method updated successfully"})
}

// Delete payment method
func (h *PaymentMethodsHandler) DeletePaymentMethod(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	methodIDStr := c.Param("id")
	methodID, err := primitive.ObjectIDFromHex(methodIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}

	// Soft delete by setting is_active to false
	result, err := h.db.Collection("payment_methods").UpdateOne(
		context.Background(),
		bson.M{"_id": methodID, "user_id": userID},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payment method"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method deleted successfully"})
}

func (h *PaymentMethodsHandler) getWalletBalance(userID primitive.ObjectID) (float64, error) {
	// Try blockchain wallet first (default: Stellar)
	// Fallback to traditional wallet
	collection := h.db.Collection("wallets")
	var wallet models.Wallet
	err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&wallet)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}
