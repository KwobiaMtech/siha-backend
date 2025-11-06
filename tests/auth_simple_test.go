package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"healthy_pay_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPasswordHashing(t *testing.T) {
	password := "password123"
	
	// Test password hashing
	hashedPassword, err := utils.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
	
	// Test password verification
	isValid := utils.CheckPassword(password, hashedPassword)
	assert.True(t, isValid)
	
	// Test wrong password
	isInvalid := utils.CheckPassword("wrongpassword", hashedPassword)
	assert.False(t, isInvalid)
}

func TestJWTGeneration(t *testing.T) {
	// Mock user ID
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	
	// Generate JWT token
	token, err := utils.GenerateJWT(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	
	// Verify token format (should have 3 parts separated by dots)
	parts := len(bytes.Split([]byte(token), []byte(".")))
	assert.Equal(t, 3, parts)
}

func TestRegistrationValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Mock handler for validation testing
	router.POST("/register", func(c *gin.Context) {
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
		
		c.JSON(http.StatusOK, gin.H{"message": "validation passed"})
	})

	t.Run("Valid Registration Data", func(t *testing.T) {
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
		
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid Email", func(t *testing.T) {
		payload := map[string]string{
			"email":     "invalid-email",
			"password":  "password123",
			"firstName": "John",
			"lastName":  "Doe",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Short Password", func(t *testing.T) {
		payload := map[string]string{
			"email":     "test@example.com",
			"password":  "123",
			"firstName": "John",
			"lastName":  "Doe",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
			// Missing firstName and lastName
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
