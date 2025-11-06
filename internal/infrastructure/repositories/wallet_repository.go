package repositories

import (
	"context"
	"healthy_pay_backend/internal/domain/entities"
	"healthy_pay_backend/internal/domain/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type walletRepository struct {
	collection *mongo.Collection
}

func NewWalletRepository(db *mongo.Database) repositories.WalletRepository {
	return &walletRepository{
		collection: db.Collection("wallets"),
	}
}

func (r *walletRepository) Create(ctx context.Context, wallet *entities.Wallet) error {
	result, err := r.collection.InsertOne(ctx, wallet)
	if err != nil {
		return err
	}
	wallet.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) (*entities.Wallet, error) {
	var wallet entities.Wallet
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, userID primitive.ObjectID, amount float64) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$inc": bson.M{"balance": amount}},
	)
	return err
}
