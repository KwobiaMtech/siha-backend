package services

import (
	"context"
	"errors"
	"time"

	"healthy_pay_backend/internal/application/dtos"
	"healthy_pay_backend/internal/domain/entities"
	"healthy_pay_backend/internal/domain/repositories"
	"healthy_pay_backend/internal/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	userRepo   repositories.UserRepository
	walletRepo repositories.WalletRepository
}

func NewAuthService(userRepo repositories.UserRepository, walletRepo repositories.WalletRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		walletRepo: walletRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dtos.RegisterRequest) (*dtos.AuthResponse, error) {
	// Check if user exists
	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, errors.New("user already exists")
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &entities.User{
		Email:       req.Email,
		Password:    hashedPassword,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		IsVerified:  false,
		KYCStatus:   "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create wallet
	wallet := &entities.Wallet{
		UserID:    user.ID,
		Balance:   0.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.walletRepo.Create(ctx, wallet)

	// Generate token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return &dtos.AuthResponse{User: user, Token: token}, nil
}

func (s *AuthService) Login(ctx context.Context, req *dtos.LoginRequest) (*dtos.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return &dtos.AuthResponse{User: user, Token: token}, nil
}
