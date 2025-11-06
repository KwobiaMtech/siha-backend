package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OTPHandler struct {
	db *mongo.Database
}

type OTP struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PhoneNumber string             `bson:"phone_number" json:"phoneNumber"`
	Code        string             `bson:"code" json:"code"`
	ExpiresAt   time.Time          `bson:"expires_at" json:"expiresAt"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
}

func NewOTPHandler(db *mongo.Database) *OTPHandler {
	return &OTPHandler{db: db}
}

func (h *OTPHandler) SendOTP(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phoneNumber" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate 6-digit OTP
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	otp := OTP{
		PhoneNumber: req.PhoneNumber,
		Code:        code,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		CreatedAt:   time.Now(),
	}

	collection := h.db.Collection("otps")
	
	// Delete existing OTPs for this phone number
	collection.DeleteMany(context.Background(), bson.M{"phone_number": req.PhoneNumber})
	
	// Insert new OTP
	_, err := collection.InsertOne(context.Background(), otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	// In production, send SMS here
	// For development, return the code
	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
		"code":    code, // Remove this in production
	})
}

func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phoneNumber" binding:"required"`
		Code        string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.db.Collection("otps")
	var otp OTP
	
	err := collection.FindOne(context.Background(), bson.M{
		"phone_number": req.PhoneNumber,
		"code":         req.Code,
		"expires_at":   bson.M{"$gt": time.Now()},
	}).Decode(&otp)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	// Delete the used OTP
	collection.DeleteOne(context.Background(), bson.M{"_id": otp.ID})

	// Update user verification status
	userCollection := h.db.Collection("users")
	_, err = userCollection.UpdateOne(
		context.Background(),
		bson.M{"phone_number": req.PhoneNumber},
		bson.M{"$set": bson.M{"is_verified": true}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update verification status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}
