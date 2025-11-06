package services

import (
	"context"
	"log"
	"time"

	"healthy_pay_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionQueue struct {
	db         *mongo.Database
	pspService *PSPService
	ticker     *time.Ticker
	stopChan   chan bool
}

func NewTransactionQueue(db *mongo.Database) *TransactionQueue {
	return &TransactionQueue{
		db:         db,
		pspService: NewPSPService(db),
		stopChan:   make(chan bool),
	}
}

func (tq *TransactionQueue) Start() {
	log.Println("Starting transaction queue processor (every 30 seconds)...")
	tq.ticker = time.NewTicker(30 * time.Second)
	
	go func() {
		for {
			select {
			case <-tq.ticker.C:
				tq.processPendingTransactions()
			case <-tq.stopChan:
				tq.ticker.Stop()
				return
			}
		}
	}()
}

func (tq *TransactionQueue) Stop() {
	log.Println("Stopping transaction queue processor...")
	tq.stopChan <- true
}

func (tq *TransactionQueue) processPendingTransactions() {
	collection := tq.db.Collection("transactions")
	
	// Find all pending transactions with PSP transaction IDs
	cursor, err := collection.Find(context.Background(), bson.M{
		"status": bson.M{"$in": []string{"collection_pending", "pending"}},
		"psp_transaction_id": bson.M{"$ne": ""},
	})
	if err != nil {
		log.Printf("Error fetching pending transactions: %v", err)
		return
	}
	defer cursor.Close(context.Background())

	var pendingTransactions []models.Transaction
	if err := cursor.All(context.Background(), &pendingTransactions); err != nil {
		log.Printf("Error decoding transactions: %v", err)
		return
	}

	if len(pendingTransactions) == 0 {
		return
	}

	log.Printf("Processing %d pending transactions...", len(pendingTransactions))
	processed := 0

	for _, transaction := range pendingTransactions {
		// Check PSP status
		status, err := tq.pspService.CheckCollectionStatus(transaction.PSPTransactionID)
		if err != nil {
			log.Printf("Error checking PSP status for transaction %s: %v", transaction.ID.Hex(), err)
			continue
		}

		updateFields := bson.M{"updated_at": time.Now()}
		
		switch status {
		case "collected", "completed", "success":
			updateFields["status"] = "completed"
			updateFields["collection_status"] = "collected"
			processed++
			log.Printf("✅ Transaction %s marked as completed", transaction.ID.Hex())
			
		case "failed", "cancelled", "error":
			updateFields["status"] = "failed"
			updateFields["collection_status"] = "failed"
			processed++
			log.Printf("❌ Transaction %s marked as failed", transaction.ID.Hex())
			
		default:
			// Still pending, check if it's been too long (24 hours)
			if time.Since(transaction.CreatedAt) > 24*time.Hour {
				updateFields["status"] = "failed"
				updateFields["collection_status"] = "timeout"
				processed++
				log.Printf("⏰ Transaction %s timed out after 24 hours", transaction.ID.Hex())
			}
		}

		// Update transaction if status changed
		if len(updateFields) > 1 { // More than just updated_at
			_, err = collection.UpdateOne(
				context.Background(),
				bson.M{"_id": transaction.ID},
				bson.M{"$set": updateFields},
			)
			if err != nil {
				log.Printf("Error updating transaction %s: %v", transaction.ID.Hex(), err)
			}
		}
	}

	if processed > 0 {
		log.Printf("🔄 Successfully processed %d transactions", processed)
	}
}
