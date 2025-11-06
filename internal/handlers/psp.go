package handlers

import (
	"net/http"

	"healthy_pay_backend/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type PSPHandler struct {
	db         *mongo.Database
	pspService *services.PSPService
}

func NewPSPHandler(db *mongo.Database) *PSPHandler {
	return &PSPHandler{
		db:         db,
		pspService: services.NewPSPService(db),
	}
}

// GetAvailablePSPs returns list of available PSP providers
func (h *PSPHandler) GetAvailablePSPs(c *gin.Context) {
	providers := h.pspService.GetAvailableProviders()
	
	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"message":   "Available PSP providers",
	})
}

// GetPSPForProvider returns the recommended PSP for a given mobile network provider
func (h *PSPHandler) GetPSPForProvider(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider parameter is required"})
		return
	}

	// This would use the internal selection logic
	selectedPSP := h.selectPSPForProvider(provider)
	
	c.JSON(http.StatusOK, gin.H{
		"provider":     provider,
		"selectedPSP":  selectedPSP,
		"message":      "Recommended PSP for provider",
	})
}

// TestPSPConnection tests connection to a specific PSP
func (h *PSPHandler) TestPSPConnection(c *gin.Context) {
	pspName := c.Param("psp")
	
	provider := h.pspService.GetProvider(pspName)
	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PSP not found"})
		return
	}

	// Test with a small dummy collection request
	testReq := services.CollectionRequest{
		Amount:      1.0, // 1 GHS test
		PhoneNumber: "0244000000", // Test number
		Provider:    "MTN",
		Reference:   "TEST_" + pspName,
	}

	response, err := provider.InitiateCollection(testReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "PSP connection test failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"psp":      pspName,
		"status":   "connected",
		"response": response,
		"message":  "PSP connection test successful",
	})
}

// Helper method to select PSP (mirrors the service logic)
func (h *PSPHandler) selectPSPForProvider(provider string) string {
	providers := h.pspService.GetAvailableProviders()
	
	switch provider {
	case "MTN":
		for _, psp := range providers {
			if psp == "mtn" {
				return "mtn"
			}
		}
		for _, psp := range providers {
			if psp == "ogate" {
				return "ogate"
			}
		}
	case "VODAFONE":
		for _, psp := range providers {
			if psp == "vodafone" {
				return "vodafone"
			}
		}
		for _, psp := range providers {
			if psp == "ogate" {
				return "ogate"
			}
		}
	case "AIRTELTIGO":
		for _, psp := range providers {
			if psp == "ogate" {
				return "ogate"
			}
		}
	}
	
	// Return first available provider
	if len(providers) > 0 {
		return providers[0]
	}
	
	return "demo"
}
