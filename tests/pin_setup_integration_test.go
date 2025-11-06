package tests

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

func setupPINTestDB() *mongo.Database {
	client, _ := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	db := client.Database("healthy_pay_pin_test")
	return db
}

func createPINTestUser(db *mongo.Database, withPIN bool) (primitive.ObjectID, string) {
	hashedPassword, _ := utils.HashPassword("password123")
	
	user := models.User{
		ID:         primitive.NewObjectID(),
		Email:      "pintest@example.com",
		Password:   hashedPassword,
		FirstName:  "PIN",
		LastName:   "Test",
		IsVerified: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	if withPIN {
		hashedPIN, _ := utils.HashPassword("1234")
		user.PIN = hashedPIN
	}
	
	collection := db.Collection("users")
	collection.InsertOne(context.Background(), user)
	
	token, _ := utils.GenerateJWT(user.ID)
	return user.ID, token
}

func TestPINSetupFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPINTestDB()
	defer db.Drop(context.Background())

	t.Run("PIN Setup Success", func(t *testing.T) {
		userID, token := createPINTestUser(db, false)
		
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
		
		// Verify PIN was saved
		collection := db.Collection("users")
		var user models.User
		collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
		assert.NotEmpty(t, user.PIN)
	})

	t.Run("PIN Setup Without Auth", func(t *testing.T) {
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
	})

	t.Run("PIN Setup Invalid Format", func(t *testing.T) {
		_, token := createPINTestUser(db, false)
		
		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		
		protected := router.Group("/")
		protected.Use(middleware.AuthMiddleware())
		protected.POST("/auth/setup-pin", authHandler.SetupPIN)

		pinData := map[string]string{"pin": "12345"}
		jsonData, _ := json.Marshal(pinData)

		req, _ := http.NewRequest("POST", "/auth/setup-pin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLoginPINStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPINTestDB()
	defer db.Drop(context.Background())

	t.Run("Login User Without PIN", func(t *testing.T) {
		hashedPassword, _ := utils.HashPassword("password123")
		user := models.User{
			ID:         primitive.NewObjectID(),
			Email:      "nopin@test.com",
			Password:   hashedPassword,
			FirstName:  "No",
			LastName:   "PIN",
			IsVerified: true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		
		collection := db.Collection("users")
		collection.InsertOne(context.Background(), user)

		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		router.POST("/auth/login", authHandler.Login)

		loginData := map[string]string{
			"email":    "nopin@test.com",
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
	})

	t.Run("Login User With PIN", func(t *testing.T) {
		hashedPassword, _ := utils.HashPassword("password123")
		hashedPIN, _ := utils.HashPassword("1234")
		user := models.User{
			ID:         primitive.NewObjectID(),
			Email:      "withpin@test.com",
			Password:   hashedPassword,
			PIN:        hashedPIN,
			FirstName:  "With",
			LastName:   "PIN",
			IsVerified: true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		
		collection := db.Collection("users")
		collection.InsertOne(context.Background(), user)

		router := gin.New()
		authHandler := handlers.NewAuthHandler(db)
		router.POST("/auth/login", authHandler.Login)

		loginData := map[string]string{
			"email":    "withpin@test.com",
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
	})
}
