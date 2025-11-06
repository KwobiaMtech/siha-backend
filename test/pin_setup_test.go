package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"healthy_pay_backend/internal/handlers"
	"healthy_pay_backend/internal/middleware"
	"healthy_pay_backend/internal/models"
	"healthy_pay_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupPinTestDB() *mongo.Database {
	client, _ := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	db := client.Database("healthy_pay_test")
	return db
}

func createTestUser(db *mongo.Database) (primitive.ObjectID, string) {
	user := models.User{
		ID:        primitive.NewObjectID(),
		Email:     "pintest@example.com",
		Password:  "hashedpassword",
		FirstName: "PIN",
		LastName:  "Test",
		IsVerified: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	collection := db.Collection("users")
	collection.InsertOne(context.Background(), user)
	
	token, _ := utils.GenerateJWT(user.ID)
	return user.ID, token
}

func TestPINSetupEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPinTestDB()
	defer db.Drop(context.Background())

	t.Run("Setup PIN with valid token and PIN", func(t *testing.T) {
		userID, token := createTestUser(db)
		
		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		
		protected := router.Group("/")
		protected.Use(middleware.AuthMiddleware())
		protected.POST("/auth/setup-pin", authHandler.SetupPIN)

		pinData := map[string]string{"pin": "1234"}
		jsonData, _ := json.Marshal(pinData)

		req, _ := http.NewRequest("POST", "/auth/setup-pin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "PIN set successfully", response["message"])

		// Verify PIN was saved in database
		collection := db.Collection("users")
		var user models.User
		collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
		assert.NotEmpty(t, user.PIN)
		assert.True(t, utils.CheckPassword("1234", user.PIN))
	})

	t.Run("Setup PIN without authentication token", func(t *testing.T) {
		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		
		protected := router.Group("/")
		protected.Use(middleware.AuthMiddleware())
		protected.POST("/auth/setup-pin", authHandler.SetupPIN)

		pinData := map[string]string{"pin": "1234"}
		jsonData, _ := json.Marshal(pinData)

		req, _ := http.NewRequest("POST", "/auth/setup-pin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Authorization header required", response["error"])
	})

	t.Run("Setup PIN with invalid token", func(t *testing.T) {
		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		
		protected := router.Group("/")
		protected.Use(middleware.AuthMiddleware())
		protected.POST("/auth/setup-pin", authHandler.SetupPIN)

		pinData := map[string]string{"pin": "1234"}
		jsonData, _ := json.Marshal(pinData)

		req, _ := http.NewRequest("POST", "/auth/setup-pin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer invalid_token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid token", response["error"])
	})

	t.Run("Setup PIN with invalid PIN format", func(t *testing.T) {
		_, token := createTestUser(db)
		
		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		
		protected := router.Group("/")
		protected.Use(middleware.AuthMiddleware())
		protected.POST("/auth/setup-pin", authHandler.SetupPIN)

		// Test with 5-digit PIN
		pinData := map[string]string{"pin": "12345"}
		jsonData, _ := json.Marshal(pinData)

		req, _ := http.NewRequest("POST", "/auth/setup-pin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "PIN must be exactly 4 digits", response["error"])
	})

	t.Run("Setup PIN with 3-digit PIN", func(t *testing.T) {
		_, token := createTestUser(db)
		
		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		
		protected := router.Group("/")
		protected.Use(middleware.AuthMiddleware())
		protected.POST("/auth/setup-pin", authHandler.SetupPIN)

		pinData := map[string]string{"pin": "123"}
		jsonData, _ := json.Marshal(pinData)

		req, _ := http.NewRequest("POST", "/auth/setup-pin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Setup PIN with non-numeric PIN", func(t *testing.T) {
		_, token := createTestUser(db)
		
		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		
		protected := router.Group("/")
		protected.Use(middleware.AuthMiddleware())
		protected.POST("/auth/setup-pin", authHandler.SetupPIN)

		pinData := map[string]string{"pin": "abcd"}
		jsonData, _ := json.Marshal(pinData)

		req, _ := http.NewRequest("POST", "/auth/setup-pin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLoginWithPINStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPinTestDB()
	defer db.Drop(context.Background())

	t.Run("Login returns hasPIN false for user without PIN", func(t *testing.T) {
		// Create user without PIN
		hashedPassword, _ := utils.HashPassword("password123")
		user := models.User{
			ID:        primitive.NewObjectID(),
			Email:     "nopin@example.com",
			Password:  hashedPassword,
			FirstName: "No",
			LastName:  "PIN",
			IsVerified: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		collection := db.Collection("users")
		collection.InsertOne(context.Background(), user)

		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		router.POST("/auth/login", authHandler.Login)

		loginData := map[string]string{
			"email":    "nopin@example.com",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, false, response["hasPIN"])
		assert.NotEmpty(t, response["token"])
	})

	t.Run("Login returns hasPIN true for user with PIN", func(t *testing.T) {
		// Create user with PIN
		hashedPassword, _ := utils.HashPassword("password123")
		hashedPIN, _ := utils.HashPassword("1234")
		user := models.User{
			ID:        primitive.NewObjectID(),
			Email:     "withpin@example.com",
			Password:  hashedPassword,
			PIN:       hashedPIN,
			FirstName: "With",
			LastName:  "PIN",
			IsVerified: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		collection := db.Collection("users")
		collection.InsertOne(context.Background(), user)

		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		router.POST("/auth/login", authHandler.Login)

		loginData := map[string]string{
			"email":    "withpin@example.com",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["hasPIN"])
		assert.NotEmpty(t, response["token"])
	})
}

func TestEmailVerificationWithPINStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPinTestDB()
	defer db.Drop(context.Background())

	t.Run("Email verification returns hasPIN false for new user", func(t *testing.T) {
		// Create unverified user without PIN
		hashedPassword, _ := utils.HashPassword("password123")
		user := models.User{
			ID:               primitive.NewObjectID(),
			Email:            "verify@example.com",
			Password:         hashedPassword,
			FirstName:        "Verify",
			LastName:         "Test",
			IsVerified:       false,
			VerificationCode: "123456",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		
		collection := db.Collection("users")
		collection.InsertOne(context.Background(), user)

		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		router.POST("/auth/verify-email", authHandler.VerifyEmail)

		verifyData := map[string]string{
			"email": "verify@example.com",
			"code":  "123456",
		}
		jsonData, _ := json.Marshal(verifyData)

		req, _ := http.NewRequest("POST", "/auth/verify-email", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, false, response["hasPIN"])
		assert.NotEmpty(t, response["token"])
		assert.Equal(t, "Email verified successfully", response["message"])
	})
}
