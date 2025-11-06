package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"healthy_pay_backend/internal/services"
)

type RateHandler struct {
	rateService *services.RateService
}

func NewRateHandler() *RateHandler {
	return &RateHandler{
		rateService: services.NewRateService(),
	}
}

func (rh *RateHandler) GetOnrampRate(c *gin.Context) {
	currency := c.Query("currency")
	if currency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currency parameter is required"})
		return
	}

	rate, err := rh.rateService.GetOnrampRate(currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rate,
	})
}

func (rh *RateHandler) GetOfframpRate(c *gin.Context) {
	currency := c.Query("currency")
	if currency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currency parameter is required"})
		return
	}

	rate, err := rh.rateService.GetOfframpRate(currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rate,
	})
}

func (rh *RateHandler) ConvertAmount(c *gin.Context) {
	amountStr := c.Query("amount")
	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")

	if amountStr == "" || fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount, from, and to parameters are required"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount format"})
		return
	}

	convertedAmount, err := rh.rateService.ConvertAmount(amount, fromCurrency, toCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"original_amount":   amount,
			"from_currency":     fromCurrency,
			"to_currency":       toCurrency,
			"converted_amount":  convertedAmount,
		},
	})
}

func (rh *RateHandler) GetAllRates(c *gin.Context) {
	rates, err := rh.rateService.GetExchangeRates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rates,
	})
}
