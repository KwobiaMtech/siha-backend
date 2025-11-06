package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KYCDocument struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"userId"`
	DocumentType string             `bson:"document_type" json:"documentType"`
	DocumentURL  string             `bson:"document_url" json:"documentUrl"`
	Status       string             `bson:"status" json:"status"`
	UploadedAt   time.Time          `bson:"uploaded_at" json:"uploadedAt"`
	VerifiedAt   *time.Time         `bson:"verified_at,omitempty" json:"verifiedAt,omitempty"`
}

type Profile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userId"`
	DateOfBirth time.Time          `bson:"date_of_birth" json:"dateOfBirth"`
	Address     string             `bson:"address" json:"address"`
	City        string             `bson:"city" json:"city"`
	State       string             `bson:"state" json:"state"`
	Country     string             `bson:"country" json:"country"`
	PostalCode  string             `bson:"postal_code" json:"postalCode"`
	Occupation  string             `bson:"occupation" json:"occupation"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}
