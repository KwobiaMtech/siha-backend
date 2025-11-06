package services

import (
	"context"
	"testing"

	"healthy_pay_backend/internal/application/dtos"
	"healthy_pay_backend/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mockUserRepository struct {
	users map[string]*entities.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *entities.User) error {
	user.ID = primitive.NewObjectID()
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, mongo.ErrNoDocuments
}

func (m *mockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

func (m *mockUserRepository) Update(ctx context.Context, user *entities.User) error {
	m.users[user.Email] = user
	return nil
}

type mockWalletRepository struct{}

func (m *mockWalletRepository) Create(ctx context.Context, wallet *entities.Wallet) error {
	wallet.ID = primitive.NewObjectID()
	return nil
}

func (m *mockWalletRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) (*entities.Wallet, error) {
	return &entities.Wallet{ID: primitive.NewObjectID(), UserID: userID}, nil
}

func (m *mockWalletRepository) UpdateBalance(ctx context.Context, userID primitive.ObjectID, amount float64) error {
	return nil
}

func TestAuthService_Register(t *testing.T) {
	userRepo := &mockUserRepository{users: make(map[string]*entities.User)}
	walletRepo := &mockWalletRepository{}
	service := NewAuthService(userRepo, walletRepo)

	req := &dtos.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		PhoneNumber: "+1234567890",
	}

	response, err := service.Register(context.Background(), req)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if response == nil {
		t.Error("Expected response, got nil")
	}
	
	if response.Token == "" {
		t.Error("Expected token, got empty string")
	}
}

func TestAuthService_RegisterDuplicate(t *testing.T) {
	userRepo := &mockUserRepository{users: make(map[string]*entities.User)}
	walletRepo := &mockWalletRepository{}
	service := NewAuthService(userRepo, walletRepo)

	req := &dtos.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		PhoneNumber: "+1234567890",
	}

	// Register first user
	service.Register(context.Background(), req)
	
	// Try to register duplicate
	_, err := service.Register(context.Background(), req)
	
	if err == nil {
		t.Error("Expected error for duplicate user, got nil")
	}
	
	if err.Error() != "user already exists" {
		t.Errorf("Expected 'user already exists', got %v", err)
	}
}
