package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"healthy_pay_backend/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidateEmailInputValidation(t *testing.T) {
	// Use a nil database for input validation tests only
	authHandler := handlers.NewAuthHandler(nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/validate-email", authHandler.ValidateEmail)

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

	t.Run("Valid Email Format", func(t *testing.T) {
		payload := map[string]string{
			"email": "test@example.com",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// Should not be 400 (bad request) for valid email format
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})
}
