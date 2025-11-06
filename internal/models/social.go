package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SocialAccount struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"userId"`
	Provider   string             `bson:"provider" json:"provider"`
	ProviderID string             `bson:"provider_id" json:"providerId"`
	Email      string             `bson:"email" json:"email"`
	CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
}
