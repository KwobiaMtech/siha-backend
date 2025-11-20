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

type TransactionHandler struct {
	db         *mongo.Database
	pspService *services.PSPService
}

func NewTransactionHandler(db *mongo.Database) *TransactionHandler {
	return &TransactionHandler{
		db:         db,
		pspService: services.NewPSPService(db),
	}
}

type SendMoneyRequest struct {
	PaymentMethodID      string  `json:"paymentMethodId" binding:"required"`
	RecipientName        string  `json:"recipientName" binding:"required"`
	RecipientAccount     string  `json:"recipientAccount" binding:"required"`
	RecipientType        string  `json:"recipientType" binding:"required"`
	RecipientNetwork     string  `json:"recipientNetwork,omitempty"`
	RecipientCurrency    string  `json:"recipientCurrency"`
	Amount               float64 `json:"amount" binding:"required,gt=0"`
	InvestmentPercentage float64 `json:"investmentPercentage"`
	DonationChoice       string  `json:"donationChoice"`
	Description          string  `json:"description"`
}

func (h *TransactionHandler) getPaymentMethodByID(userID primitive.ObjectID, paymentMethodID string) (*models.UserPaymentMethod, error) {
	if paymentMethodID == "wallet_balance" {
		return &models.UserPaymentMethod{
			Type:   "wallet",
			UserID: userID,
		}, nil
	}

	objectID, err := primitive.ObjectIDFromHex(paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method ID")
	}

	// First try user_payment_methods collection (new format)
	var paymentMethod models.UserPaymentMethod
	err = h.db.Collection("user_payment_methods").FindOne(
		context.Background(),
		bson.M{"_id": objectID, "user_id": userID},
	).Decode(&paymentMethod)

	if err == nil {
		return &paymentMethod, nil
	}

	// If not found, try payment_methods collection (auth format)
	var authPaymentMethod struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		UserID      primitive.ObjectID `bson:"user_id"`
		Type        string             `bson:"type"`
		Network     string             `bson:"network,omitempty"`
		PhoneNumber string             `bson:"phone_number,omitempty"`
		AccountName string             `bson:"account_name,omitempty"`
		Currency    string             `bson:"currency,omitempty"`
		IsDefault   bool               `bson:"is_default"`
		IsActive    bool               `bson:"is_active"`
	}

	err = h.db.Collection("payment_methods").FindOne(
		context.Background(),
		bson.M{"_id": objectID, "user_id": userID, "is_active": true},
	).Decode(&authPaymentMethod)

	if err != nil {
		return nil, fmt.Errorf("payment method not found")
	}

	// Convert auth format to UserPaymentMethod format
	return &models.UserPaymentMethod{
		ID:          authPaymentMethod.ID,
		UserID:      authPaymentMethod.UserID,
		Type:        authPaymentMethod.Type,
		Network:     authPaymentMethod.Network,
		PhoneNumber: authPaymentMethod.PhoneNumber,
		AccountName: authPaymentMethod.AccountName,
		Currency:    authPaymentMethod.Currency,
		IsDefault:   authPaymentMethod.IsDefault,
		IsActive:    authPaymentMethod.IsActive,
	}, nil
}

func (h *TransactionHandler) SendMoney(c *gin.Context) {
	userIDStr := c.GetString("userID")
	fromUserID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req SendMoneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default currency to GHS if not provided
	if req.RecipientCurrency == "" {
		req.RecipientCurrency = "GHS"
	}

	// Validate supported currencies
	supportedCurrencies := []string{"GHS", "USD", "KES", "ZMW"}
	isSupported := false
	for _, currency := range supportedCurrencies {
		if req.RecipientCurrency == currency {
			isSupported = true
			break
		}
	}
	if !isSupported {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported currency"})
		return
	}

	// Query payment method details using ID
	paymentMethod, err := h.getPaymentMethodByID(fromUserID, req.PaymentMethodID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
		return
	}

	// Validate network for mobile money recipients
	if req.RecipientType == "mobile_money" && req.RecipientNetwork == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Network is required for mobile money recipients"})
		return
	}

	investmentAmount := 0.0
	totalAmount := req.Amount
	
	if req.InvestmentPercentage > 0 {
		investmentAmount = req.Amount * (req.InvestmentPercentage / 100)
		totalAmount = req.Amount + investmentAmount
	}

	if paymentMethod.Type == "wallet" {
		h.processWalletPayment(c, fromUserID, req, totalAmount, investmentAmount)
	} else if paymentMethod.Type == "mobile_money" {
		h.processTwoStageMobileMoneyPayment(c, fromUserID, paymentMethod, req, totalAmount, investmentAmount)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported payment method"})
	}
}

func (h *TransactionHandler) processWalletPayment(c *gin.Context, fromUserID primitive.ObjectID, req SendMoneyRequest, totalAmount, investmentAmount float64) {
	balance, err := h.getWalletBalance(fromUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wallet balance"})
		return
	}
	if balance < totalAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	transaction := h.createTransaction(fromUserID, req, totalAmount, investmentAmount, "completed")
	result, err := h.saveTransaction(transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	if err := h.updateBalance(fromUserID, req.PaymentMethodID, -totalAmount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	h.deliverToRecipient(req, req.Amount)
	h.handlePostTransaction(fromUserID, investmentAmount, req)

	transaction.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusOK, gin.H{
		"message":     "Money sent successfully",
		"transaction": transaction,
	})
}

func (h *TransactionHandler) processTwoStageMobileMoneyPayment(c *gin.Context, fromUserID primitive.ObjectID, paymentMethod *models.UserPaymentMethod, req SendMoneyRequest, totalAmount, investmentAmount float64) {
	// Create transaction with two-stage status tracking
	transaction := h.createTwoStageTransaction(fromUserID, req, totalAmount, investmentAmount, "collection_pending")
	transaction.CollectionStatus = "pending"
	transaction.InvestmentStatus = "pending"
	transaction.DeliveryStatus = "pending"

	result, err := h.saveTransaction(transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Stage 1: Initiate collection from sender to Siha platform
	collectionReq := services.CollectionRequest{
		Amount:      totalAmount, // Collect total amount including investment
		PhoneNumber: paymentMethod.PhoneNumber,
		Provider:    paymentMethod.Provider,
		Reference:   result.InsertedID.(primitive.ObjectID).Hex(),
	}

	collectionResp, err := h.pspService.InitiateCollection(collectionReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate collection"})
		return
	}

	// Update transaction with PSP details
	h.updateTransactionWithPSPData(result.InsertedID.(primitive.ObjectID), collectionReq, collectionResp)
	
	// Start async processing for collection monitoring and distribution
	go h.processTwoStageCollectionAndDistribution(result.InsertedID.(primitive.ObjectID), req, investmentAmount, fromUserID)

	transaction.ID = result.InsertedID.(primitive.ObjectID)
	transaction.PSPTransactionID = collectionResp.TransactionID

	c.JSON(http.StatusOK, gin.H{
		"message":     "Collection initiated, processing payment",
		"transaction": transaction,
		"status":      "collection_pending",
	})
}

func (h *TransactionHandler) processTwoStageCollectionAndDistribution(transactionID primitive.ObjectID, req SendMoneyRequest, investmentAmount float64, fromUserID primitive.ObjectID) {
	// Check status multiple times with exponential backoff
	maxRetries := 10
	baseDelay := 5 * time.Second
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Wait before checking (exponential backoff)
		delay := time.Duration(attempt+1) * baseDelay
		time.Sleep(delay)

		collection := h.db.Collection("transactions")
		var transaction models.Transaction
		err := collection.FindOne(context.Background(), bson.M{"_id": transactionID}).Decode(&transaction)
		if err != nil {
			continue
		}

		// Skip if already processed
		if transaction.Status == "completed" || transaction.Status == "failed" {
			return
		}

		// Check collection status
		status, err := h.pspService.CheckCollectionStatus(transaction.PSPTransactionID)
		if err != nil {
			continue
		}

		if status == "collected" {
			// Update collection status
			collection.UpdateOne(
				context.Background(),
				bson.M{"_id": transactionID},
				bson.M{"$set": bson.M{
					"collection_status": "collected",
					"status":           "processing_distribution",
					"updated_at":       time.Now(),
				}},
			)

			// Stage 2a: Process investment allocation
			if investmentAmount > 0 {
				err := h.processInvestmentAllocation(fromUserID, investmentAmount, req.DonationChoice, req.RecipientCurrency)
				if err == nil {
					collection.UpdateOne(
						context.Background(),
						bson.M{"_id": transactionID},
						bson.M{"$set": bson.M{
							"investment_status": "allocated",
							"updated_at":       time.Now(),
						}},
					)
				}
			}

			// Stage 2b: Deliver to recipient
			deliveryReq := services.DeliveryRequest{
				Amount:           req.Amount,
				RecipientType:    req.RecipientType,
				RecipientAccount: req.RecipientAccount,
				RecipientNetwork: req.RecipientNetwork,
				Reference:        transactionID.Hex(),
			}

			err = h.pspService.InitiateDelivery(deliveryReq)
			if err == nil {
				collection.UpdateOne(
					context.Background(),
					bson.M{"_id": transactionID},
					bson.M{"$set": bson.M{
						"delivery_status": "delivered",
						"status":         "completed",
						"updated_at":     time.Now(),
					}},
				)
			}

			// Save recipient for future use
			h.saveRecipient(fromUserID, req.RecipientName, req.RecipientAccount, req.RecipientType)
			return
		} else if status == "failed" {
			// Mark transaction as failed
			collection.UpdateOne(
				context.Background(),
				bson.M{"_id": transactionID},
				bson.M{"$set": bson.M{
					"collection_status": "failed",
					"status":           "failed",
					"updated_at":       time.Now(),
				}},
			)
			return
		}
		// If still pending, continue loop
	}
	
	// If we've exhausted retries, mark as failed
	collection := h.db.Collection("transactions")
	collection.UpdateOne(
		context.Background(),
		bson.M{"_id": transactionID},
		bson.M{"$set": bson.M{
			"status":     "failed",
			"updated_at": time.Now(),
		}},
	)
}

func (h *TransactionHandler) createTwoStageTransaction(fromUserID primitive.ObjectID, req SendMoneyRequest, totalAmount, investmentAmount float64, status string) models.Transaction {
	return models.Transaction{
		FromUserID:           fromUserID,
		RecipientName:        req.RecipientName,
		RecipientAccount:     req.RecipientAccount,
		RecipientType:        req.RecipientType,
		RecipientNetwork:     req.RecipientNetwork,
		RecipientCurrency:    req.RecipientCurrency,
		Amount:               req.Amount,
		InvestmentAmount:     investmentAmount,
		InvestmentPercentage: req.InvestmentPercentage,
		DonationChoice:       req.DonationChoice,
		PaymentMethod:        req.PaymentMethodID,
		Type:                 "send",
		Status:               status,
		Description:          req.Description,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func (h *TransactionHandler) processInvestmentAllocation(userID primitive.ObjectID, amount float64, donationChoice, currency string) error {
	if amount <= 0 {
		return nil
	}

	// Get current USD exchange rate (rate is always against USD)
	rate := h.getCurrentUSDRate(currency)

	investment := models.Investment{
		UserID:             userID,
		Amount:             amount,
		InvestmentCurrency: currency,
		Type:               "send_flow_investment",
		Status:             "pending",
		Returns:            0.0,
		Rate:               rate,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	_, err := h.db.Collection("investments").InsertOne(context.Background(), investment)
	return err
}

func (h *TransactionHandler) saveRecipient(userID primitive.ObjectID, name, account, recipientType string) error {
	// Check if recipient already exists
	var existingRecipient models.Recipient
	err := h.db.Collection("recipients").FindOne(
		context.Background(),
		bson.M{"user_id": userID, "account": account},
	).Decode(&existingRecipient)

	if err == mongo.ErrNoDocuments {
		// Create new recipient
		recipient := models.Recipient{
			UserID:     userID,
			Name:       name,
			Account:    account,
			Type:       recipientType,
			IsFrequent: false,
			CreatedAt:  time.Now(),
		}
		_, err = h.db.Collection("recipients").InsertOne(context.Background(), recipient)
		return err
	}

	// Update existing recipient usage
	_, err = h.db.Collection("recipients").UpdateOne(
		context.Background(),
		bson.M{"_id": existingRecipient.ID},
		bson.M{"$set": bson.M{"updated_at": time.Now()}},
	)
	return err
}

func (h *TransactionHandler) GetPaymentMethods(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var paymentMethods []map[string]interface{}

	// Get user's mobile money payment methods from payment_methods collection
	cursor, err := h.db.Collection("payment_methods").Find(
		context.Background(),
		bson.M{"user_id": userID, "type": "mobile_money"},
	)
	if err == nil {
		defer cursor.Close(context.Background())
		
		for cursor.Next(context.Background()) {
			var method models.UserPaymentMethod
			if err := cursor.Decode(&method); err == nil {
				subtitle := fmt.Sprintf("%s - %s", method.Network, method.PhoneNumber)
				if len(method.PhoneNumber) > 4 {
					// Mask phone number for security
					masked := method.PhoneNumber[:4] + "****" + method.PhoneNumber[len(method.PhoneNumber)-3:]
					subtitle = fmt.Sprintf("%s - %s", method.Network, masked)
				}

				paymentMethods = append(paymentMethods, map[string]interface{}{
					"id":         method.ID.Hex(),
					"title":      fmt.Sprintf("📱 %s Mobile Money", method.Network),
					"subtitle":   subtitle,
					"type":       "mobile_money",
					"provider":   method.Network,
					"isDefault":  method.IsDefault,
					"hasBalance": false,
				})
			}
		}
	}

	// Add wallet balance option
	balance, err := h.getWalletBalance(userID)
	if err == nil {
		paymentMethods = append(paymentMethods, map[string]interface{}{
			"id":         "wallet_balance",
			"title":      "💰 Wallet Balance",
			"subtitle":   "Send from your platform wallet",
			"balance":    balance,
			"type":       "wallet",
			"hasBalance": true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"paymentMethods": paymentMethods,
	})
}

func (h *TransactionHandler) GetMobileNetworks(c *gin.Context) {
	currency := c.Query("currency")
	if currency == "" {
		currency = "GHS" // Default to Ghana Cedis
	}

	var networks []map[string]interface{}

	switch currency {
	case "GHS":
		networks = []map[string]interface{}{
			{"id": "MTN", "name": "MTN Ghana", "code": "MTN"},
			{"id": "TELECEL", "name": "Telecel Ghana", "code": "TELECEL"},
			{"id": "AIRTELTIGO", "name": "AirtelTigo", "code": "AIRTELTIGO"},
		}
	case "KES":
		networks = []map[string]interface{}{
			{"id": "MPESA", "name": "M-Pesa", "code": "MPESA"},
			{"id": "AIRTEL", "name": "Airtel Money", "code": "AIRTEL"},
		}
	case "NGN":
		networks = []map[string]interface{}{
			{"id": "MTN", "name": "MTN Nigeria", "code": "MTN"},
			{"id": "AIRTEL", "name": "Airtel Nigeria", "code": "AIRTEL"},
			{"id": "GLO", "name": "Glo Mobile", "code": "GLO"},
			{"id": "9MOBILE", "name": "9mobile", "code": "9MOBILE"},
		}
	case "UGX":
		networks = []map[string]interface{}{
			{"id": "MTN", "name": "MTN Uganda", "code": "MTN"},
			{"id": "AIRTEL", "name": "Airtel Uganda", "code": "AIRTEL"},
		}
	default:
		networks = []map[string]interface{}{
			{"id": "MTN", "name": "MTN", "code": "MTN"},
			{"id": "AIRTEL", "name": "Airtel", "code": "AIRTEL"},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"currency": currency,
		"networks": networks,
	})
}

func (h *TransactionHandler) GetRecipientDeliveryOptions(c *gin.Context) {
	options := []map[string]interface{}{
		{
			"id":             "mobile_money",
			"title":          "📱 Mobile Money",
			"subtitle":       "Send to mobile money account",
			"placeholder":    "0244123456",
			"requiresNetwork": true,
		},
		{
			"id":             "crypto_wallet",
			"title":          "₿ Crypto Wallet",
			"subtitle":       "Send to crypto wallet address",
			"placeholder":    "0x1234...abcd",
			"requiresNetwork": false,
		},
		{
			"id":             "stellar_wallet",
			"title":          "⭐ Stellar Wallet",
			"subtitle":       "Send USDC to Stellar address",
			"placeholder":    "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			"requiresNetwork": false,
		},
		{
			"id":             "siha_wallet",
			"title":          "💰 Siha Wallet",
			"subtitle":       "Send to Siha wallet ID",
			"placeholder":    "SW123456789",
			"requiresNetwork": false,
		},
	}

	c.JSON(http.StatusOK, gin.H{"deliveryOptions": options})
}

func (h *TransactionHandler) GetRecipients(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := h.db.Collection("recipients")
	cursor, err := collection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipients"})
		return
	}
	defer cursor.Close(context.Background())

	var recipients []models.Recipient
	if err := cursor.All(context.Background(), &recipients); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode recipients"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recipients": recipients})
}

func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := h.db.Collection("transactions")
	// Sort by createdAt descending to get latest first
	opts := options.Find().SetSort(bson.D{{"createdAt", -1}})
	
	// Get all transactions for this user (deposits, sends, receives)
	cursor, err := collection.Find(context.Background(), bson.M{"userId": userID}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	defer cursor.Close(context.Background())

	var transactions []models.UnifiedTransaction
	if err := cursor.All(context.Background(), &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// Helper methods
func (h *TransactionHandler) createTransaction(fromUserID primitive.ObjectID, req SendMoneyRequest, totalAmount, investmentAmount float64, status string) models.Transaction {
	return models.Transaction{
		FromUserID:           fromUserID,
		RecipientName:        req.RecipientName,
		RecipientAccount:     req.RecipientAccount,
		RecipientType:        req.RecipientType,
		RecipientNetwork:     req.RecipientNetwork,
		RecipientCurrency:    req.RecipientCurrency,
		Amount:               req.Amount,
		InvestmentAmount:     investmentAmount,
		InvestmentPercentage: req.InvestmentPercentage,
		DonationChoice:       req.DonationChoice,
		PaymentMethod:        req.PaymentMethodID,
		Type:                 "send_money",
		Status:               status,
		Description:          req.Description,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func (h *TransactionHandler) saveTransaction(transaction models.Transaction) (*mongo.InsertOneResult, error) {
	collection := h.db.Collection("transactions")
	return collection.InsertOne(context.Background(), transaction)
}

func (h *TransactionHandler) updateTransactionWithPSPData(transactionID primitive.ObjectID, request services.CollectionRequest, response *services.CollectionResponse) {
	collection := h.db.Collection("transactions")
	collection.UpdateOne(
		context.Background(),
		bson.M{"_id": transactionID},
		bson.M{"$set": bson.M{
			"psp_transaction_id": response.TransactionID,
			"psp_request":       request,
			"psp_response":      response.RawResponse, // Store raw API response
			"updated_at":        time.Now(),
		}},
	)
}

func (h *TransactionHandler) CheckTransactionStatus(c *gin.Context) {
	transactionID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	var transaction models.Transaction
	err = h.db.Collection("transactions").FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&transaction)
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": transaction})
}

func (h *TransactionHandler) getUserPaymentMethod(userID primitive.ObjectID, methodType string) (*models.UserPaymentMethod, error) {
	collection := h.db.Collection("user_payment_methods")
	var method models.UserPaymentMethod
	err := collection.FindOne(context.Background(), bson.M{
		"user_id":    userID,
		"type":       methodType,
		"is_default": true,
	}).Decode(&method)
	return &method, err
}

func (h *TransactionHandler) deliverToRecipient(req SendMoneyRequest, amount float64) {
	deliveryReq := services.DeliveryRequest{
		Amount:           amount,
		RecipientType:    req.RecipientType,
		RecipientAccount: req.RecipientAccount,
		Reference:        fmt.Sprintf("wallet_%d", time.Now().Unix()),
	}
	go func() {
		pspService := services.NewPSPService(h.db)
		pspService.InitiateDelivery(deliveryReq)
	}()
}

func (h *TransactionHandler) handlePostTransaction(fromUserID primitive.ObjectID, investmentAmount float64, req SendMoneyRequest) {
	if investmentAmount > 0 {
		h.createInvestment(fromUserID, investmentAmount, req.DonationChoice, req.RecipientCurrency)
	}
	h.saveRecipient(fromUserID, req.RecipientName, req.RecipientAccount, req.RecipientType)
}

func (h *TransactionHandler) getWalletBalance(userID primitive.ObjectID) (float64, error) {
	// Try blockchain wallet first (default: Stellar)
	blockchainServiceFactory := services.NewBlockchainServiceFactory(h.db)
	defaultService := blockchainServiceFactory.GetDefaultBlockchainService()
	
	if defaultService != nil {
		wallet, err := defaultService.GetWallet(userID)
		if err == nil {
			// Return USDC balance as primary balance
			for _, balance := range wallet.Balances {
				if balance.AssetCode == "USDC" {
					return balance.Balance, nil
				}
			}
		}
	}

	// Fallback to traditional wallet
	collection := h.db.Collection("wallets")
	var wallet models.Wallet
	err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&wallet)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

func (h *TransactionHandler) updateBalance(userID primitive.ObjectID, paymentMethod string, amount float64) error {
	if paymentMethod == "wallet_balance" {
		collection := h.db.Collection("wallets")
		_, err := collection.UpdateOne(
			context.Background(),
			bson.M{"user_id": userID},
			bson.M{"$inc": bson.M{"balance": amount}},
		)
		return err
	}
	return nil
}

func (h *TransactionHandler) getCurrentUSDRate(currency string) *models.ExchangeRate {
	// Mock exchange rates - in production, this would fetch from a real exchange rate API
	rates := map[string]float64{
		"GHS": 12.50, // 1 USD = 12.50 GHS
		"KES": 150.0, // 1 USD = 150 KES
		"ZMW": 25.0,  // 1 USD = 25 ZMW
		"USD": 1.0,   // 1 USD = 1 USD
	}

	rate, exists := rates[currency]
	if !exists {
		rate = 1.0 // Default to 1:1 if currency not found
	}

	return &models.ExchangeRate{
		FromCurrency: currency,
		ToCurrency:   "USD",
		Rate:         rate,
		Timestamp:    time.Now(),
	}
}

func (h *TransactionHandler) createInvestment(userID primitive.ObjectID, amount float64, donationChoice, currency string) error {
	// Get current USD exchange rate (rate is always against USD)
	rate := h.getCurrentUSDRate(currency)

	investment := models.Investment{
		UserID:             userID,
		Amount:             amount,
		InvestmentCurrency: currency,
		Type:               "send_investment",
		Status:             "active",
		Returns:            0,
		Rate:               rate,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	collection := h.db.Collection("investments")
	_, err := collection.InsertOne(context.Background(), investment)
	return err
}

func (h *TransactionHandler) ProcessPendingTransactions(c *gin.Context) {
	collection := h.db.Collection("transactions")
	
	// Find all pending transactions
	cursor, err := collection.Find(context.Background(), bson.M{
		"status": bson.M{"$in": []string{"collection_pending", "pending"}},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending transactions"})
		return
	}
	defer cursor.Close(context.Background())

	var pendingTransactions []models.Transaction
	if err := cursor.All(context.Background(), &pendingTransactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode transactions"})
		return
	}

	processed := 0
	for _, transaction := range pendingTransactions {
		if transaction.PSPTransactionID != "" {
			// Check PSP status
			status, err := h.pspService.CheckCollectionStatus(transaction.PSPTransactionID)
			if err == nil {
				if status == "collected" {
					// Update to completed
					collection.UpdateOne(
						context.Background(),
						bson.M{"_id": transaction.ID},
						bson.M{"$set": bson.M{
							"status":     "completed",
							"updated_at": time.Now(),
						}},
					)
					processed++
				} else if status == "failed" {
					// Update to failed
					collection.UpdateOne(
						context.Background(),
						bson.M{"_id": transaction.ID},
						bson.M{"$set": bson.M{
							"status":     "failed",
							"updated_at": time.Now(),
						}},
					)
					processed++
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Processed pending transactions",
		"total_pending": len(pendingTransactions),
		"processed": processed,
	})
}


