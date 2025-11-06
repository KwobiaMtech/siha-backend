package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"healthy_pay_backend/internal/handlers"
	"healthy_pay_backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB() *mongo.Database {
	client, _ := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	db := client.Database("healthypay_test")
	
	// Clean up collections
	db.Collection("users").Drop(context.Background())
	db.Collection("wallets").Drop(context.Background())
	
	return db
}

func TestAuthFlow(t *testing.T) {
	db := setupTestDB()
	authHandler := handlers.NewAuthHandler(db)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	t.Run("Register User", func(t *testing.T) {
		payload := map[string]string{
			"email":     "test@example.com",
			"password":  "password123",
			"firstName": "John",
			"lastName":  "Doe",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		assert.Contains(t, response, "user")
		assert.Contains(t, response, "token")
		
		user := response["user"].(map[string]interface{})
		assert.Equal(t, "test@example.com", user["email"])
		assert.Equal(t, "John", user["firstName"])
		assert.Equal(t, "Doe", user["lastName"])
		assert.Equal(t, true, user["isVerified"])
	})

	t.Run("User Exists in Database", func(t *testing.T) {
		var user models.User
		err := db.Collection("users").FindOne(context.Background(), bson.M{"email": "test@example.com"}).Decode(&user)
		
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.True(t, user.IsVerified)
		assert.NotEmpty(t, user.Password)
		assert.NotEqual(t, "password123", user.Password) // Should be hashed
	})

	t.Run("Wallet Created for User", func(t *testing.T) {
		var user models.User
		db.Collection("users").FindOne(context.Background(), bson.M{"email": "test@example.com"}).Decode(&user)
		
		var wallet models.Wallet
		err := db.Collection("wallets").FindOne(context.Background(), bson.M{"user_id": user.ID}).Decode(&wallet)
		
		assert.NoError(t, err)
		assert.Equal(t, user.ID, wallet.UserID)
		assert.Equal(t, 0.0, wallet.Balance)
		assert.Equal(t, "USD", wallet.Currency)
	})

	t.Run("Login with Valid Credentials", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		assert.Contains(t, response, "user")
		assert.Contains(t, response, "token")
		assert.NotEmpty(t, response["token"])
	})

	t.Run("Login with Invalid Credentials", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Duplicate Registration", func(t *testing.T) {
		payload := map[string]string{
			"email":     "test@example.com",
			"password":  "password123",
			"firstName": "Jane",
			"lastName":  "Smith",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Invalid Registration Data", func(t *testing.T) {
		payload := map[string]string{
			"email":    "invalid-email",
			"password": "123", // Too short
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
