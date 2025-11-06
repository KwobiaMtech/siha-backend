package handlers

import (
	"context"
	"net/http"
	"time"

	"healthy_pay_backend/internal/models"
	"healthy_pay_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SocialHandler struct {
	db *mongo.Database
}

func NewSocialHandler(db *mongo.Database) *SocialHandler {
	return &SocialHandler{db: db}
}

func (h *SocialHandler) GoogleLogin(c *gin.Context) {
	var req struct {
		IDToken   string `json:"idToken" binding:"required"`
		Email     string `json:"email" binding:"required"`
		Name      string `json:"name" binding:"required"`
		GoogleID  string `json:"googleId" binding:"required"`
		PhotoURL  string `json:"photoUrl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.handleSocialLogin("google", req.GoogleID, req.Email, req.Name, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *SocialHandler) FacebookLogin(c *gin.Context) {
	var req struct {
		AccessToken string `json:"accessToken" binding:"required"`
		Email       string `json:"email" binding:"required"`
		Name        string `json:"name" binding:"required"`
		FacebookID  string `json:"facebookId" binding:"required"`
		PhotoURL    string `json:"photoUrl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.handleSocialLogin("facebook", req.FacebookID, req.Email, req.Name, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *SocialHandler) AppleLogin(c *gin.Context) {
	var req struct {
		IdentityToken string `json:"identityToken" binding:"required"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		AppleID       string `json:"appleId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.handleSocialLogin("apple", req.AppleID, req.Email, req.Name, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *SocialHandler) handleSocialLogin(provider, providerID, email, name, phone string) (*models.User, string, error) {
	socialCollection := h.db.Collection("social_accounts")
	userCollection := h.db.Collection("users")

	// Check if social account exists
	var socialAccount models.SocialAccount
	err := socialCollection.FindOne(context.Background(), bson.M{
		"provider":    provider,
		"provider_id": providerID,
	}).Decode(&socialAccount)

	var user models.User
	if err == mongo.ErrNoDocuments {
		// Check if user exists by email
		err = userCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
		if err == mongo.ErrNoDocuments {
			// Create new user
			names := utils.SplitName(name)
			user = models.User{
				Email:       email,
				FirstName:   names[0],
				LastName:    names[1],
				PhoneNumber: phone,
				IsVerified:  true,
				KYCStatus:   "pending",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			result, err := userCollection.InsertOne(context.Background(), user)
			if err != nil {
				return nil, "", err
			}
			user.ID = result.InsertedID.(primitive.ObjectID)

			// Create wallet
			wallet := models.Wallet{
				UserID:    user.ID,
				Balance:   0.0,
				Currency:  "USD",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			h.db.Collection("wallets").InsertOne(context.Background(), wallet)
		} else if err != nil {
			return nil, "", err
		}

		// Link social account
		socialAccount = models.SocialAccount{
			UserID:     user.ID,
			Provider:   provider,
			ProviderID: providerID,
			Email:      email,
			CreatedAt:  time.Now(),
		}
		socialCollection.InsertOne(context.Background(), socialAccount)
	} else if err != nil {
		return nil, "", err
	} else {
		// Get existing user
		err = userCollection.FindOne(context.Background(), bson.M{"_id": socialAccount.UserID}).Decode(&user)
		if err != nil {
			return nil, "", err
		}
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return nil, "", err
	}

	user.Password = ""
	return &user, token, nil
}
