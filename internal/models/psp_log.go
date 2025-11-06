package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PSPLog struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TransactionReference  string             `bson:"transaction_reference" json:"transactionReference"`
	Operation            string             `bson:"operation" json:"operation"`
	PSPName              string             `bson:"psp_name" json:"pspName"`
	RequestPayload       string             `bson:"request_payload,omitempty" json:"requestPayload,omitempty"`
	ResponsePayload      string             `bson:"response_payload,omitempty" json:"responsePayload,omitempty"`
	Status               string             `bson:"status,omitempty" json:"status,omitempty"`
	Type                 string             `bson:"type" json:"type"` // "request" or "response"
	Timestamp            time.Time          `bson:"timestamp" json:"timestamp"`
}

type TransactionPSPLog struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TransactionID     primitive.ObjectID `bson:"transaction_id" json:"transactionId"`
	PSPTransactionID  string             `bson:"psp_transaction_id" json:"pspTransactionId"`
	PSPName           string             `bson:"psp_name" json:"pspName"`
	RequestPayload    interface{}        `bson:"request_payload" json:"requestPayload"`
	ResponsePayload   interface{}        `bson:"response_payload" json:"responsePayload"`
	CreatedAt         time.Time          `bson:"created_at" json:"createdAt"`
}
