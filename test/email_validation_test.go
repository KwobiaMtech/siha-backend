package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"healthy_pay_backend/internal/config"
	"healthy_pay_backend/internal/database"
	"healthy_pay_backend/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := config.Load()
	db, _ := database.Connect(cfg.MongoURI)
	
	r := gin.New()
	routes.SetupRoutes(r, db)
	return r
}

func TestValidateEmail(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name     string
		email    string
		expected int
	}{
		{"Valid email", "test@example.com", http.StatusOK},
		{"Invalid email", "invalid-email", http.StatusBadRequest},
		{"Empty email", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]string{"email": tt.email}
			jsonPayload, _ := json.Marshal(payload)
			
			req := httptest.NewRequest("POST", "/api/auth/validate-email", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestVerifyEmail(t *testing.T) {
	router := setupTestRouter()

	payload := map[string]string{
		"email": "test@example.com",
		"code":  "123456",
	}
	jsonPayload, _ := json.Marshal(payload)
	
	req := httptest.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 400 or 404 since user doesn't exist
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
}
