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
	// Get userID from context (it's a string from JWT)
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert string to ObjectID
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

	// Get payment method details
	paymentMethod, err := h.getPaymentMethodByID(userID, req.PaymentMethodID)
	if err != nil {
		fmt.Printf("Payment method lookup failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
		return
	}

	fmt.Printf("Payment method found: Type=%s, Network=%s, Phone=%s\n", 
		paymentMethod.Type, paymentMethod.Network, paymentMethod.PhoneNumber)

	// Generate unique reference
	reference := fmt.Sprintf("DEP_%d_%s", time.Now().Unix(), userID.Hex()[:8])

	// Create deposit record
	deposit := models.Deposit{
		ID:                   primitive.NewObjectID(),
		UserID:               userID,
		Amount:               req.Amount,
		PaymentMethodID:      req.PaymentMethodID,
		InvestmentPercentage: req.InvestmentPercentage,
		DonationChoice:       req.DonationChoice,
		Status:               "pending",
		PSPReference:         reference,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save deposit to database first
	collection := h.db.Collection("deposits")
	_, err = collection.InsertOne(context.Background(), deposit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deposit record"})
		return
	}

	// Initiate PSP collection for mobile money
	var pspResponse *services.CollectionResponse
	if paymentMethod.Type == "mobile_money" {
		fmt.Printf("Initiating PSP collection for mobile money: %s %s\n", 
			paymentMethod.Network, paymentMethod.PhoneNumber)
		
		collectionReq := services.CollectionRequest{
			Amount:      req.Amount,
			PhoneNumber: paymentMethod.PhoneNumber,
			Provider:    paymentMethod.Network,
			Reference:   reference,
		}

		pspResponse, err = h.pspService.InitiateCollection(collectionReq)
		if err != nil {
			fmt.Printf("PSP collection failed: %v\n", err)
			// Update deposit status to failed
			h.updateDepositStatus(deposit.ID, "failed", nil)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to initiate payment collection: %v", err)})
			return
		}

		fmt.Printf("PSP collection successful: TransactionID=%s, Status=%s\n", 
			pspResponse.TransactionID, pspResponse.Status)

		// Update deposit with PSP response
		h.updateDepositStatus(deposit.ID, "initiated", pspResponse)
		deposit.TransactionID = pspResponse.TransactionID
		deposit.PSPReference = pspResponse.TransactionID
		deposit.Status = "initiated"
		
		// Update the database record with PSP transaction details
		collection.UpdateOne(context.Background(), bson.M{"_id": deposit.ID}, bson.M{
			"$set": bson.M{
				"transactionId": pspResponse.TransactionID,
				"pspReference": pspResponse.TransactionID,
				"status": "initiated",
				"updatedAt": time.Now(),
			},
		})
	} else {
		fmt.Printf("Non-mobile money payment method, marking as initiated for manual processing\n")
		// For non-mobile money, mark as initiated for manual processing
		h.updateDepositStatus(deposit.ID, "initiated", nil)
		deposit.Status = "initiated"
		deposit.TransactionID = reference
	}

	response := models.DepositResponse{
		ID:            deposit.ID.Hex(),
		Status:        deposit.Status,
		Message:       "Deposit initiated successfully. Please complete payment on your mobile device.",
		TransactionID: deposit.TransactionID,
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

	// Find deposit
	collection := h.db.Collection("deposits")
	var deposit models.Deposit
	err = collection.FindOne(context.Background(), bson.M{
		"_id":    objectID,
		"userId": userID,
	}).Decode(&deposit)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposit"})
		return
	}

	// Check PSP status if still pending/initiated
	if (deposit.Status == "initiated" || deposit.Status == "pending") && deposit.TransactionID != "" {
		pspStatus, err := h.pspService.CheckCollectionStatus(deposit.TransactionID)
		if err == nil && pspStatus != deposit.Status {
			// Update status if changed
			h.updateDepositStatus(deposit.ID, pspStatus, nil)
			deposit.Status = pspStatus

			// Process successful deposit
			if pspStatus == "collected" || pspStatus == "success" {
				h.processSuccessfulDeposit(deposit)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              deposit.ID.Hex(),
		"status":          deposit.Status,
		"amount":          deposit.Amount,
		"paymentMethodId": deposit.PaymentMethodID,
		"queueStatus":     deposit.QueueStatus,
		"createdAt":       deposit.CreatedAt,
		"updatedAt":       deposit.UpdatedAt,
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

	collection := h.db.Collection("deposits")
	cursor, err := collection.Find(context.Background(), bson.M{
		"userId": userID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposits"})
		return
	}
	defer cursor.Close(context.Background())

	var deposits []models.Deposit
	if err = cursor.All(context.Background(), &deposits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode deposits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deposits": deposits})
}

func (h *DepositHandler) updateDepositStatus(depositID primitive.ObjectID, status string, pspResponse interface{}) {
	collection := h.db.Collection("deposits")
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	if pspResponse != nil {
		update["$set"].(bson.M)["pspResponse"] = pspResponse
		
		// If pspResponse is a CollectionResponse, extract transaction ID
		if collResp, ok := pspResponse.(*services.CollectionResponse); ok {
			update["$set"].(bson.M)["transactionId"] = collResp.TransactionID
			update["$set"].(bson.M)["pspReference"] = collResp.TransactionID
		}
	}

	collection.UpdateOne(context.Background(), bson.M{"_id": depositID}, update)
}

func (h *DepositHandler) processSuccessfulDeposit(deposit models.Deposit) {
	// Update user wallet balance
	walletCollection := h.db.Collection("wallets")
	
	// Calculate amounts
	savingsAmount := deposit.Amount * (100 - deposit.InvestmentPercentage) / 100
	investmentAmount := deposit.Amount * deposit.InvestmentPercentage / 100

	// Update wallet balance
	walletCollection.UpdateOne(
		context.Background(),
		bson.M{"userId": deposit.UserID},
		bson.M{
			"$inc": bson.M{
				"balance": savingsAmount,
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
	)

	// Create investment record if applicable
	if investmentAmount > 0 {
		investmentCollection := h.db.Collection("investments")
		investment := bson.M{
			"userId":    deposit.UserID,
			"amount":    investmentAmount,
			"type":      "deposit_investment",
			"status":    "active",
			"createdAt": time.Now(),
			"updatedAt": time.Now(),
		}
		investmentCollection.InsertOne(context.Background(), investment)
	}

	// Handle donation if applicable
	if deposit.DonationChoice != "none" && deposit.DonationChoice != "" {
		donationCollection := h.db.Collection("donations")
		donationAmount := 0.0
		
		if deposit.DonationChoice == "both" {
			donationAmount = deposit.Amount
		} else if deposit.DonationChoice == "profit" {
			donationAmount = investmentAmount
		}

		if donationAmount > 0 {
			donation := bson.M{
				"userId":     deposit.UserID,
				"amount":     donationAmount,
				"type":       deposit.DonationChoice,
				"depositId":  deposit.ID,
				"status":     "pending",
				"createdAt":  time.Now(),
			}
			donationCollection.InsertOne(context.Background(), donation)
		}
	}
}

func (h *DepositHandler) detectNetworkFromPhone(phoneNumber string) string {
	// Remove country code and normalize
	phone := phoneNumber
	if len(phone) > 10 {
		phone = phone[len(phone)-10:] // Get last 10 digits
	}
	
	if len(phone) < 10 {
		return "MTN" // Default fallback
	}
	
	prefix := phone[:3]
	
	// MTN prefixes
	mtnPrefixes := []string{"024", "054", "055", "059"}
	for _, p := range mtnPrefixes {
		if prefix == p {
			return "MTN"
		}
	}
	
	// Telecel prefixes
	telecelPrefixes := []string{"020", "050"}
	for _, p := range telecelPrefixes {
		if prefix == p {
			return "TELECEL"
		}
	}
	
	// AirtelTigo prefixes
	airtelPrefixes := []string{"027", "057", "026", "056"}
	for _, p := range airtelPrefixes {
		if prefix == p {
			return "AIRTELTIGO"
		}
	}
	
	return "MTN" // Default fallback
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
