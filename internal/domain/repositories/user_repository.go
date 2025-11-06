package repositories

import (
	"context"
	"healthy_pay_backend/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
}

type WalletRepository interface {
	Create(ctx context.Context, wallet *entities.Wallet) error
	GetByUserID(ctx context.Context, userID primitive.ObjectID) (*entities.Wallet, error)
	UpdateBalance(ctx context.Context, userID primitive.ObjectID, amount float64) error
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *entities.Transaction) error
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*entities.Transaction, error)
}
