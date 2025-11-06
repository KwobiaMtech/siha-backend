package handlers

import (
	"context"
	"net/http"
	"time"

	"healthy_pay_backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type KYCHandler struct {
	db *mongo.Database
}

func NewKYCHandler(db *mongo.Database) *KYCHandler {
	return &KYCHandler{db: db}
}

func (h *KYCHandler) UploadDocument(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		DocumentType string `json:"documentType" binding:"required"`
		DocumentURL  string `json:"documentUrl" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	document := models.KYCDocument{
		UserID:       userID,
		DocumentType: req.DocumentType,
		DocumentURL:  req.DocumentURL,
		Status:       "pending",
		UploadedAt:   time.Now(),
	}

	collection := h.db.Collection("kyc_documents")
	_, err = collection.InsertOne(context.Background(), document)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload document"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Document uploaded successfully"})
}

func (h *KYCHandler) SetupProfile(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		DateOfBirth string `json:"dateOfBirth" binding:"required"`
		Address     string `json:"address" binding:"required"`
		City        string `json:"city" binding:"required"`
		State       string `json:"state" binding:"required"`
		Country     string `json:"country" binding:"required"`
		PostalCode  string `json:"postalCode" binding:"required"`
		Occupation  string `json:"occupation" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	profile := models.Profile{
		UserID:      userID,
		DateOfBirth: dob,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		PostalCode:  req.PostalCode,
		Occupation:  req.Occupation,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	collection := h.db.Collection("profiles")
	_, err = collection.InsertOne(context.Background(), profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
		return
	}

	// Update user KYC status
	userCollection := h.db.Collection("users")
	_, err = userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"kyc_status": "completed"}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update KYC status"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profile setup completed"})
}
