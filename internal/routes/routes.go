package routes

import (
	"healthy_pay_backend/internal/handlers"
	"healthy_pay_backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(r *gin.Engine, db *mongo.Database) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	socialHandler := handlers.NewSocialHandler(db)
	userHandler := handlers.NewUserHandler(db)
	walletHandler := handlers.NewWalletHandler(db)
	transactionHandler := handlers.NewTransactionHandler(db)
	investmentHandler := handlers.NewInvestmentHandler(db)
	kycHandler := handlers.NewKYCHandler(db)
	otpHandler := handlers.NewOTPHandler(db)
	mobileMoneyHandler := handlers.NewMobileMoneyHandler(db)
	stellarWalletHandler := handlers.NewStellarWalletHandler(db)
	pspHandler := handlers.NewPSPHandler(db)
	rateHandler := handlers.NewRateHandler()

	// Public routes
	api := r.Group("/api/v1")
	{
		// Test endpoint
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy", "message": "Backend is running"})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/validate-email", authHandler.ValidateEmail)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/test-email", authHandler.TestEmail)
			auth.POST("/test-wallet", authHandler.TestWalletCreation)
			auth.POST("/google", socialHandler.GoogleLogin)
			auth.POST("/facebook", socialHandler.FacebookLogin)
			auth.POST("/apple", socialHandler.AppleLogin)
		}

		otp := api.Group("/otp")
		{
			otp.POST("/send", otpHandler.SendOTP)
			otp.POST("/verify", otpHandler.VerifyOTP)
		}
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// Test endpoint
		protected.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Protected route works"})
		})

		// Auth routes (protected)
		protected.POST("/auth/setup-pin", authHandler.SetupPIN)
		protected.POST("/auth/setup-payment-method", authHandler.SetupPaymentMethod)
		protected.GET("/auth/payment-method", authHandler.GetPaymentMethod)

		// User routes
		user := protected.Group("/user")
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
		}

		// KYC routes
		kyc := protected.Group("/kyc")
		{
			kyc.POST("/upload-document", kycHandler.UploadDocument)
			kyc.POST("/setup-profile", kycHandler.SetupProfile)
		}

		// Wallet routes
		wallet := protected.Group("/wallet")
		{
			wallet.GET("", walletHandler.GetBalance) // Default wallet endpoint
			wallet.GET("/balance", walletHandler.GetBalance)
			wallet.POST("/add-funds", walletHandler.AddFunds)
			wallet.POST("/create-blockchain", walletHandler.CreateBlockchainWallet)
			wallet.POST("/activate-blockchain", walletHandler.ActivateBlockchain)
			wallet.GET("/supported-blockchains", walletHandler.GetSupportedBlockchains)
		}

		// Send flow routes
		send := protected.Group("/send")
		{
			send.GET("/payment-methods", transactionHandler.GetPaymentMethods)
			send.GET("/recipients", transactionHandler.GetRecipients)
			send.GET("/delivery-options", transactionHandler.GetRecipientDeliveryOptions)
			send.GET("/mobile-networks", transactionHandler.GetMobileNetworks)
			send.POST("/money", transactionHandler.SendMoney)
		}

		// Mobile money routes
		mobile := protected.Group("/mobile-money")
		{
			mobile.POST("/add-wallet", mobileMoneyHandler.AddMobileWallet)
			mobile.GET("/wallets", mobileMoneyHandler.GetMobileWallets)
		}

		// Transaction routes
		transactions := protected.Group("/transactions")
		{
			transactions.POST("/send", transactionHandler.SendMoney) // Legacy endpoint
			transactions.GET("/", transactionHandler.GetTransactions)
			transactions.GET("/:id/status", transactionHandler.CheckTransactionStatus)
			transactions.POST("/process-pending", transactionHandler.ProcessPendingTransactions)
		}

		// Stellar wallet routes
		stellar := protected.Group("/stellar")
		{
			stellar.GET("/info", stellarWalletHandler.GetWalletInfo)
			stellar.POST("/wallet", stellarWalletHandler.CreateWallet)
			stellar.GET("/wallet", stellarWalletHandler.GetWallet)
			stellar.POST("/send-usdc", stellarWalletHandler.SendUSDC)
			stellar.GET("/transactions", stellarWalletHandler.GetTransactions)
			stellar.GET("/asset-info", stellarWalletHandler.GetAssetInfo)
		}

		// Investment routes
		investments := protected.Group("/investments")
		{
			investments.POST("/", investmentHandler.CreateInvestment)
			investments.GET("/", investmentHandler.GetInvestments)
		}

		// PSP routes
		psp := protected.Group("/psp")
		{
			psp.GET("/providers", pspHandler.GetAvailablePSPs)
			psp.GET("/recommend", pspHandler.GetPSPForProvider)
			psp.POST("/test/:psp", pspHandler.TestPSPConnection)
		}

		// Rate conversion routes
		rates := protected.Group("/rates")
		{
			rates.GET("/onramp", rateHandler.GetOnrampRate)
			rates.GET("/offramp", rateHandler.GetOfframpRate)
			rates.GET("/convert", rateHandler.ConvertAmount)
			rates.GET("/all", rateHandler.GetAllRates)
		}
	}
}
