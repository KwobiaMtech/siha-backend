package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"healthy_pay_backend/internal/models"
	"healthy_pay_backend/internal/services"
	"healthy_pay_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	db *mongo.Database
}

func NewAuthHandler(db *mongo.Database) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=6"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.db.Collection("users")
	
	var existingUser models.User
	err := collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Generate email verification code
	verificationCode := utils.GenerateOTP()

	user := models.User{
		Email:            req.Email,
		Password:         hashedPassword,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		IsVerified:       false,
		VerificationCode: verificationCode,
		KYCStatus:        "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	userID := result.InsertedID.(primitive.ObjectID)
	
	// Create default blockchain wallet
	blockchainServiceFactory := services.NewBlockchainServiceFactory(h.db)
	blockchainService := blockchainServiceFactory.GetService("stellar")
	if blockchainService != nil {
		_, err = blockchainService.CreateWallet(userID, "stellar")
		if err != nil {
			// Log error but don't fail registration
			fmt.Printf("⚠️ Failed to create blockchain wallet for user %s: %v\n", userID.Hex(), err)
		}
	}
	
	// Send verification email
	err = utils.SendVerificationEmail(req.Email, verificationCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	user.ID = userID
	user.Password = ""
	user.VerificationCode = ""
	
	c.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"message": "Registration successful. Please check your email for verification code.",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.db.Collection("users")
	var user models.User
	
	err := collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if email is verified
	if !user.IsVerified {
		// Generate new verification code
		verificationCode := utils.GenerateOTP()
		
		// Update user with new verification code
		collection.UpdateOne(
			context.Background(),
			bson.M{"_id": user.ID},
			bson.M{"$set": bson.M{"verification_code": verificationCode}},
		)
		
		// Send verification email
		utils.SendVerificationEmail(user.Email, verificationCode)
		
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Email not verified",
			"message": "Please check your email for verification code",
			"userId":  user.ID.Hex(),
		})
		return
	}

	// Check if PIN is set
	hasPIN := user.PIN != ""
	
	// Check if user has any payment methods
	paymentCollection := h.db.Collection("payment_methods")
	paymentCount, _ := paymentCollection.CountDocuments(context.Background(), bson.M{"user_id": user.ID, "is_active": true})
	hasPaymentMethod := paymentCount > 0

	// Create wallet if doesn't exist
	walletCollection := h.db.Collection("wallets")
	var existingWallet models.Wallet
	err = walletCollection.FindOne(context.Background(), bson.M{"user_id": user.ID}).Decode(&existingWallet)
	if err != nil {
		wallet := models.Wallet{
			UserID:    user.ID,
			Balance:   0.0,
			Currency:  "USD",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		walletCollection.InsertOne(context.Background(), wallet)
	}

	// Create default onchain wallet if doesn't exist
	h.ensureDefaultOnchainWallet(user.ID)

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.Password = ""
	user.VerificationCode = ""
	
	c.JSON(http.StatusOK, gin.H{
		"user":             user,
		"token":            token,
		"hasPIN":           hasPIN,
		"hasPaymentMethod": hasPaymentMethod,
	})
}

func (h *AuthHandler) EnsureDefaultOnchainWallet(userID primitive.ObjectID) error {
	// Get default blockchain configuration from environment
	defaultBlockchain := os.Getenv("DEFAULT_BLOCKCHAIN")
	if defaultBlockchain == "" {
		defaultBlockchain = "stellar" // Fallback to Stellar
	}
	
	defaultStablecoin := os.Getenv("DEFAULT_STABLECOIN")
	if defaultStablecoin == "" {
		defaultStablecoin = "USDC" // Fallback to USDC
	}

	fmt.Printf("Creating onchain wallet for user %s - Blockchain: %s, Stablecoin: %s\n", userID.Hex(), defaultBlockchain, defaultStablecoin)

	// Check if user already has the default blockchain wallet
	var existingWallet models.BlockchainWallet
	err := h.db.Collection("blockchain_wallets").FindOne(
		context.Background(),
		bson.M{"user_id": userID, "blockchain": defaultBlockchain},
	).Decode(&existingWallet)
	
	if err == mongo.ErrNoDocuments {
		fmt.Printf("No existing wallet found, creating new %s wallet...\n", defaultBlockchain)
		
		// Create wallet using blockchain service factory
		blockchainServiceFactory := services.NewBlockchainServiceFactory(h.db)
		blockchainService := blockchainServiceFactory.GetService(defaultBlockchain)
		
		if blockchainService != nil {
			wallet, err := blockchainService.CreateWallet(userID, defaultBlockchain)
			if err == nil {
				fmt.Printf("Wallet created successfully: %s\n", wallet.PublicKey)
				
				// Add default stablecoin
				wallet.Balances = []models.AssetBalance{
					{
						AssetCode: defaultStablecoin,
						Symbol:    defaultStablecoin,
						Name:      getAssetName(defaultStablecoin),
						Balance:   0.0,
					},
				}
				// Update wallet with default stablecoin
				_, updateErr := h.db.Collection("blockchain_wallets").UpdateOne(
					context.Background(),
					bson.M{"_id": wallet.ID},
					bson.M{"$set": bson.M{"balances": wallet.Balances}},
				)
				if updateErr != nil {
					fmt.Printf("Error updating wallet balances: %v\n", updateErr)
					return updateErr
				} else {
					fmt.Printf("Added %s to wallet successfully\n", defaultStablecoin)
				}
			} else {
				fmt.Printf("Error creating wallet: %v\n", err)
				return err
			}
		} else {
			fmt.Printf("No blockchain service available for %s\n", defaultBlockchain)
			return fmt.Errorf("no blockchain service available for %s", defaultBlockchain)
		}
	} else if err != nil {
		fmt.Printf("Error checking existing wallet: %v\n", err)
		return err
	} else {
		fmt.Printf("User already has %s wallet: %s\n", defaultBlockchain, existingWallet.PublicKey)
	}
	return nil
}

func (h *AuthHandler) ensureDefaultOnchainWallet(userID primitive.ObjectID) {
	err := h.EnsureDefaultOnchainWallet(userID)
	if err != nil {
		fmt.Printf("Failed to ensure onchain wallet: %v\n", err)
	}
}

func getAssetName(assetCode string) string {
	assetNames := map[string]string{
		"USDC": "USD Coin",
		"USDT": "Tether USD",
		"BUSD": "Binance USD",
		"DAI":  "Dai Stablecoin",
		"XLM":  "Stellar Lumens",
		"ETH":  "Ethereum",
		"BTC":  "Bitcoin",
	}
	
	if name, exists := assetNames[assetCode]; exists {
		return name
	}
	return assetCode
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.db.Collection("users")
	var user models.User
	
	err := collection.FindOne(context.Background(), bson.M{
		"email":             req.Email,
		"verification_code": req.Code,
	}).Decode(&user)
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Update user as verified
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"is_verified":       true,
			"verification_code": "",
			"updated_at":        time.Now(),
		}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// Create wallet for verified user
	wallet := models.Wallet{
		UserID:    user.ID,
		Balance:   0.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	h.db.Collection("wallets").InsertOne(context.Background(), wallet)

	// Generate token for verified user
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.Password = ""
	user.VerificationCode = ""
	user.IsVerified = true

	// Check if user has any payment methods
	paymentCollection := h.db.Collection("payment_methods")
	paymentCount, _ := paymentCollection.CountDocuments(context.Background(), bson.M{"user_id": user.ID, "is_active": true})
	hasPaymentMethod := paymentCount > 0

	c.JSON(http.StatusOK, gin.H{
		"message":          "Email verified successfully",
		"token":            token,
		"hasPIN":           user.PIN != "",
		"hasPaymentMethod": hasPaymentMethod,
	})
}

func (h *AuthHandler) TestEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate test verification code
	testCode := utils.GenerateOTP()
	
	// Try to send real email first
	err := utils.SendVerificationEmailReal(req.Email, testCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send test email",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test email sent successfully",
		"email": req.Email,
		"code": testCode,
		"note": "Check your email for the verification code",
	})
}

func (h *AuthHandler) SetupPIN(c *gin.Context) {
	var req struct {
		PIN string `json:"pin" binding:"required,len=4"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PIN must be exactly 4 digits"})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert string userID to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	hashedPIN, err := utils.HashPassword(req.PIN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash PIN"})
		return
	}

	collection := h.db.Collection("users")
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{
			"pin":        hashedPIN,
			"updated_at": time.Now(),
		}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set PIN"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PIN set successfully"})
}

func (h *AuthHandler) SetupPaymentMethod(c *gin.Context) {
	var req struct {
		Type        string `json:"type" binding:"required"`
		PhoneNumber string `json:"phoneNumber,omitempty"`
		AccountName string `json:"accountName,omitempty"`
		Network     string `json:"network,omitempty"`
		Currency    string `json:"currency,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment method type is required"})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if this is the user's first payment method
	paymentCollection := h.db.Collection("payment_methods")
	existingCount, _ := paymentCollection.CountDocuments(context.Background(), bson.M{"user_id": userID, "is_active": true})
	isDefault := existingCount == 0

	// Create new payment method
	paymentMethod := models.PaymentMethod{
		UserID:      userID,
		Type:        req.Type,
		IsDefault:   isDefault,
		IsActive:    true,
		PhoneNumber: req.PhoneNumber,
		AccountName: req.AccountName,
		Network:     req.Network,
		Currency:    req.Currency,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := paymentCollection.InsertOne(context.Background(), paymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment method"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Payment method created successfully",
		"paymentMethodId": result.InsertedID,
		"isDefault":       isDefault,
	})
}

func (h *AuthHandler) GetPaymentMethod(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	paymentCollection := h.db.Collection("payment_methods")
	cursor, err := paymentCollection.Find(context.Background(), bson.M{"user_id": userID, "is_active": true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment methods"})
		return
	}
	defer cursor.Close(context.Background())

	var paymentMethods []models.PaymentMethod
	if err = cursor.All(context.Background(), &paymentMethods); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode payment methods"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasPaymentMethod":  len(paymentMethods) > 0,
		"paymentMethods":    paymentMethods,
		"totalMethods":      len(paymentMethods),
	})
}

// TestPaymentFlow - Test endpoint for payment method flow (remove in production)
func (h *AuthHandler) TestPaymentFlow(c *gin.Context) {
	// Create a test user and simulate the complete payment method flow
	testUserID := primitive.NewObjectID()
	
	// Step 1: Create test payment method
	paymentMethod := models.PaymentMethod{
		UserID:      testUserID,
		Type:        "mobile_money",
		IsDefault:   true,
		IsActive:    true,
		PhoneNumber: "+233123456789",
		AccountName: "Test User",
		Network:     "MTN",
		Currency:    "GHS",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	paymentCollection := h.db.Collection("payment_methods")
	result, err := paymentCollection.InsertOne(context.Background(), paymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test payment method"})
		return
	}
	
	// Step 2: Retrieve the payment method
	var retrievedMethod models.PaymentMethod
	err = paymentCollection.FindOne(context.Background(), bson.M{"_id": result.InsertedID}).Decode(&retrievedMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payment method"})
		return
	}
	
	// Step 3: Test multiple payment methods
	secondMethod := models.PaymentMethod{
		UserID:      testUserID,
		Type:        "bank_card",
		IsDefault:   false,
		IsActive:    true,
		CardNumber:  "1234567890123456",
		CardHolder:  "Test User",
		ExpiryDate:  "12/25",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	result2, err := paymentCollection.InsertOne(context.Background(), secondMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create second payment method"})
		return
	}
	
	// Step 4: Get all payment methods for user
	cursor, err := paymentCollection.Find(context.Background(), bson.M{"user_id": testUserID, "is_active": true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment methods"})
		return
	}
	defer cursor.Close(context.Background())
	
	var allMethods []models.PaymentMethod
	if err = cursor.All(context.Background(), &allMethods); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode payment methods"})
		return
	}
	
	// Cleanup test data
	paymentCollection.DeleteMany(context.Background(), bson.M{"user_id": testUserID})
	
	c.JSON(http.StatusOK, gin.H{
		"message":           "Payment method flow test completed successfully",
		"testUserId":        testUserID.Hex(),
		"firstMethodId":     result.InsertedID,
		"secondMethodId":    result2.InsertedID,
		"totalMethods":      len(allMethods),
		"retrievedMethod":   retrievedMethod,
		"allMethods":        allMethods,
		"flowStatus":        "SUCCESS",
	})
}

func (h *AuthHandler) TestWalletCreation(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.EnsureDefaultOnchainWallet(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wallet creation completed"})
}
