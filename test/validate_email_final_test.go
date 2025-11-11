package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock handler for testing input validation only
func mockValidateEmailHandler(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock response for valid input
	c.JSON(http.StatusOK, gin.H{
		"exists":  false,
		"message": "Email validation test passed",
	})
}

func TestValidateEmailRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/validate-email", mockValidateEmailHandler)

	t.Run("Valid Email Format", func(t *testing.T) {
		payload := map[string]string{
			"email": "test@example.com",
		}
		
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		assert.Contains(t, response, "exists")
		assert.Contains(t, response, "message")
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

	t.Run("Various Valid Email Formats", func(t *testing.T) {
		validEmails := []string{
			"user@example.com",
			"test.email@domain.co.uk",
			"user+tag@example.org",
			"123@numbers.com",
		}

		for _, email := range validEmails {
			payload := map[string]string{
				"email": email,
			}
			
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code, "Failed for email: %s", email)
		}
	})

	t.Run("Various Invalid Email Formats", func(t *testing.T) {
		invalidEmails := []string{
			"plainaddress",
			"@missingdomain.com",
			"missing@.com",
			"spaces @domain.com",
			"double@@domain.com",
		}

		for _, email := range invalidEmails {
			payload := map[string]string{
				"email": email,
			}
			
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/validate-email", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusBadRequest, w.Code, "Should fail for email: %s", email)
		}
	})
}
