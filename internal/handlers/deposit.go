package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"healthy_pay_backend/internal/models"
	"healthy_pay_backend/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DepositHandler struct {
	db         *mongo.Database
	pspService *services.PSPService
}

func NewDepositHandler(db *mongo.Database) *DepositHandler {
	return &DepositHandler{
		db:         db,
		pspService: services.NewPSPService(db),
	}
}

func (h *DepositHandler) InitiateDeposit(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paymentMethod, err := h.getPaymentMethodByID(userID, req.PaymentMethodID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
		return
	}

	reference := fmt.Sprintf("DEP_%d_%s", time.Now().Unix(), userID.Hex()[:8])

	transaction := models.UnifiedTransaction{
		ID:                   primitive.NewObjectID(),
		UserID:               userID,
		Type:                 "deposit",
		Amount:               req.Amount,
		Status:               "pending",
		TransactionID:        "",
		PSPReference:         reference,
		PaymentMethodID:      req.PaymentMethodID,
		InvestmentPercentage: req.InvestmentPercentage,
		DonationChoice:       req.DonationChoice,
		QueueStatus:          "",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	collection := h.db.Collection("transactions")
	_, err = collection.InsertOne(context.Background(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction record"})
		return
	}

	var pspResponse *services.CollectionResponse
	if paymentMethod.Type == "mobile_money" {
		collectionReq := services.CollectionRequest{
			Amount:      req.Amount,
			PhoneNumber: paymentMethod.PhoneNumber,
			Provider:    paymentMethod.Network,
			Reference:   reference,
		}

		pspResponse, err = h.pspService.InitiateCollection(collectionReq)
		if err != nil {
			h.updateTransactionStatus(transaction.ID, "failed", nil)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate payment collection"})
			return
		}

		h.updateTransactionStatus(transaction.ID, "initiated", pspResponse)
		transaction.TransactionID = pspResponse.TransactionID
		transaction.Status = "initiated"
	} else {
		h.updateTransactionStatus(transaction.ID, "initiated", nil)
		transaction.Status = "initiated"
		transaction.TransactionID = reference
	}

	response := models.DepositResponse{
		ID:            transaction.ID.Hex(),
		Status:        transaction.Status,
		Message:       "Deposit initiated successfully. Please complete payment on your mobile device.",
		TransactionID: transaction.TransactionID,
		PSPResponse:   pspResponse,
	}

	c.JSON(http.StatusOK, response)
}

func (h *DepositHandler) CheckDepositStatus(c *gin.Context) {
	depositID := c.Param("id")
	if depositID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deposit ID required"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(depositID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deposit ID"})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := h.db.Collection("transactions")
	var transaction models.UnifiedTransaction
	err = collection.FindOne(context.Background(), bson.M{
		"_id":    objectID,
		"userId": userID,
		"type":   "deposit",
	}).Decode(&transaction)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposit"})
		return
	}

	if (transaction.Status == "initiated" || transaction.Status == "pending") && transaction.TransactionID != "" {
		pspStatus, err := h.pspService.CheckCollectionStatus(transaction.TransactionID)
		if err == nil && pspStatus != transaction.Status {
			h.updateTransactionStatus(transaction.ID, pspStatus, nil)
			transaction.Status = pspStatus

			if pspStatus == "collected" || pspStatus == "success" {
				h.processSuccessfulDeposit(transaction)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              transaction.ID.Hex(),
		"status":          transaction.Status,
		"amount":          transaction.Amount,
		"paymentMethodId": transaction.PaymentMethodID,
		"queueStatus":     transaction.QueueStatus,
		"createdAt":       transaction.CreatedAt,
		"updatedAt":       transaction.UpdatedAt,
	})
}

func (h *DepositHandler) GetDeposits(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := h.db.Collection("transactions")
	opts := options.Find().SetSort(bson.D{{"createdAt", -1}})
	
	cursor, err := collection.Find(context.Background(), bson.M{
		"userId": userID,
		"type":   "deposit",
	}, opts)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposits"})
		return
	}
	defer cursor.Close(context.Background())

	var transactions []models.UnifiedTransaction
	if err = cursor.All(context.Background(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode deposits"})
		return
	}

	// Convert to deposit format for backward compatibility
	deposits := make([]map[string]interface{}, len(transactions))
	for i, tx := range transactions {
		deposits[i] = map[string]interface{}{
			"id":                   tx.ID.Hex(),
			"userId":               tx.UserID.Hex(),
			"amount":               tx.Amount,
			"paymentMethodId":      tx.PaymentMethodID,
			"investmentPercentage": tx.InvestmentPercentage,
			"donationChoice":       tx.DonationChoice,
			"status":               tx.Status,
			"transactionId":        tx.TransactionID,
			"pspReference":         tx.PSPReference,
			"pspResponse":          tx.PSPResponse,
			"queueStatus":          tx.QueueStatus,
			"processedAt":          tx.ProcessedAt,
			"createdAt":            tx.CreatedAt,
			"updatedAt":            tx.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"deposits": deposits})
}

func (h *DepositHandler) updateTransactionStatus(transactionID primitive.ObjectID, status string, pspResponse interface{}) {
	collection := h.db.Collection("transactions")
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	if pspResponse != nil {
		update["$set"].(bson.M)["pspResponse"] = pspResponse
		
		if collResp, ok := pspResponse.(*services.CollectionResponse); ok {
			update["$set"].(bson.M)["transactionId"] = collResp.TransactionID
			update["$set"].(bson.M)["pspReference"] = collResp.TransactionID
		}
	}

	collection.UpdateOne(context.Background(), bson.M{"_id": transactionID}, update)
}

func (h *DepositHandler) processSuccessfulDeposit(transaction models.UnifiedTransaction) {
	walletCollection := h.db.Collection("wallets")
	
	savingsAmount := transaction.Amount * (100 - transaction.InvestmentPercentage) / 100
	investmentAmount := transaction.Amount * transaction.InvestmentPercentage / 100

	walletCollection.UpdateOne(
		context.Background(),
		bson.M{"userId": transaction.UserID},
		bson.M{
			"$inc": bson.M{"balance": savingsAmount},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)

	if investmentAmount > 0 {
		investmentCollection := h.db.Collection("investments")
		investment := bson.M{
			"userId":    transaction.UserID,
			"amount":    investmentAmount,
			"type":      "deposit_investment",
			"status":    "active",
			"createdAt": time.Now(),
			"updatedAt": time.Now(),
		}
		investmentCollection.InsertOne(context.Background(), investment)
	}
}

func (h *DepositHandler) getPaymentMethodByID(userID primitive.ObjectID, paymentMethodID string) (*models.PaymentMethod, error) {
	if paymentMethodID == "wallet_balance" {
		return &models.PaymentMethod{
			Type:   "wallet",
			UserID: userID,
		}, nil
	}

	objectID, err := primitive.ObjectIDFromHex(paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method ID")
	}

	collection := h.db.Collection("payment_methods")
	var paymentMethod models.PaymentMethod
	err = collection.FindOne(context.Background(), bson.M{
		"_id":       objectID,
		"user_id":   userID,
		"is_active": true,
	}).Decode(&paymentMethod)

	if err != nil {
		return nil, fmt.Errorf("payment method not found")
	}

	return &paymentMethod, nil
}
