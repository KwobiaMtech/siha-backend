package test

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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestValidateEmail(t *testing.T) {
	db := setupTestDB()
	authHandler := handlers.NewAuthHandler(db)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/validate-email", authHandler.ValidateEmail)

	// Create existing user for testing
	existingUser := models.User{
		ID:        primitive.NewObjectID(),
		Email:     "existing@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "hashedpassword",
	}
	db.Collection("users").InsertOne(context.Background(), existingUser)

	t.Run("Email Already Exists", func(t *testing.T) {
		payload := map[string]string{
			"email": "existing@example.com",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		assert.Equal(t, true, response["exists"])
		assert.Equal(t, "Email is already registered", response["message"])
	})

	t.Run("Email Available", func(t *testing.T) {
		payload := map[string]string{
			"email": "available@example.com",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		assert.Equal(t, false, response["exists"])
		assert.Equal(t, "Email is available", response["message"])
	})

	t.Run("Invalid Email Format", func(t *testing.T) {
		payload := map[string]string{
			"email": "invalid-email",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing Email Field", func(t *testing.T) {
		payload := map[string]string{}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Empty Email", func(t *testing.T) {
		payload := map[string]string{
			"email": "",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
