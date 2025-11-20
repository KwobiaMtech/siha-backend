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
	// Process pending transactions
	tq.processPendingRegularTransactions()
	
	// Process pending deposits
	tq.processPendingDeposits()
}

func (tq *TransactionQueue) processPendingRegularTransactions() {
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

	if len(pendingTransactions) > 0 {
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
}

func (tq *TransactionQueue) processPendingDeposits() {
	collection := tq.db.Collection("deposits")
	
	// Find all pending deposits with transaction IDs
	cursor, err := collection.Find(context.Background(), bson.M{
		"status": bson.M{"$in": []string{"initiated", "pending"}},
		"queueStatus": bson.M{"$in": []string{"queued", "processing"}},
		"transactionId": bson.M{"$ne": ""},
	})
	if err != nil {
		log.Printf("Error fetching pending deposits: %v", err)
		return
	}
	defer cursor.Close(context.Background())

	var pendingDeposits []models.UnifiedTransaction
	if err := cursor.All(context.Background(), &pendingDeposits); err != nil {
		log.Printf("Error decoding deposits: %v", err)
		return
	}

	if len(pendingDeposits) == 0 {
		return
	}

	log.Printf("Processing %d pending deposits...", len(pendingDeposits))
	processed := 0

	for _, deposit := range pendingDeposits {
		// Check PSP status
		status, err := tq.pspService.CheckCollectionStatus(deposit.TransactionID)
		if err != nil {
			log.Printf("Error checking PSP status for deposit %s: %v", deposit.ID.Hex(), err)
			continue
		}

		updateFields := bson.M{"updatedAt": time.Now()}
		
		switch status {
		case "collected", "completed", "success":
			updateFields["status"] = "collected"
			updateFields["queueStatus"] = "completed"
			now := time.Now()
			updateFields["processedAt"] = now
			processed++
			log.Printf("✅ Deposit %s marked as collected", deposit.ID.Hex())
			
			// Process successful deposit (update wallet, investments, etc.)
			tq.processSuccessfulDeposit(deposit)
			
		case "failed", "cancelled", "error":
			updateFields["status"] = "failed"
			updateFields["queueStatus"] = "failed"
			processed++
			log.Printf("❌ Deposit %s marked as failed", deposit.ID.Hex())
			
		default:
			// Still pending, check if it's been too long (24 hours)
			if time.Since(deposit.CreatedAt) > 24*time.Hour {
				updateFields["status"] = "failed"
				updateFields["queueStatus"] = "timeout"
				processed++
				log.Printf("⏰ Deposit %s timed out after 24 hours", deposit.ID.Hex())
			}
		}

		// Update deposit if status changed
		if len(updateFields) > 1 { // More than just updatedAt
			_, err = collection.UpdateOne(
				context.Background(),
				bson.M{"_id": deposit.ID},
				bson.M{"$set": updateFields},
			)
			if err != nil {
				log.Printf("Error updating deposit %s: %v", deposit.ID.Hex(), err)
			}
		}
	}

	if processed > 0 {
		log.Printf("💰 Successfully processed %d deposits", processed)
	}
}

func (tq *TransactionQueue) processSuccessfulDeposit(deposit models.UnifiedTransaction) {
	// Update user wallet balance
	walletCollection := tq.db.Collection("wallets")
	
	// Calculate amounts
	savingsAmount := deposit.Amount * (100 - deposit.InvestmentPercentage) / 100
	investmentAmount := deposit.Amount * deposit.InvestmentPercentage / 100

	// Update wallet balance
	_, err := walletCollection.UpdateOne(
		context.Background(),
		bson.M{"userId": deposit.UserID},
		bson.M{
			"$inc": bson.M{
				"balance": savingsAmount,
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
	)
	if err != nil {
		log.Printf("Error updating wallet for deposit %s: %v", deposit.ID.Hex(), err)
		return
	}

	log.Printf("💰 Updated wallet balance for user %s: +%.2f", deposit.UserID.Hex(), savingsAmount)

	// Create investment record if applicable
	if investmentAmount > 0 {
		investmentCollection := tq.db.Collection("investments")
		investment := bson.M{
			"userId":    deposit.UserID,
			"amount":    investmentAmount,
			"type":      "deposit_investment",
			"status":    "active",
			"depositId": deposit.ID,
			"createdAt": time.Now(),
			"updatedAt": time.Now(),
		}
		_, err := investmentCollection.InsertOne(context.Background(), investment)
		if err != nil {
			log.Printf("Error creating investment for deposit %s: %v", deposit.ID.Hex(), err)
		} else {
			log.Printf("📈 Created investment record: %.2f for user %s", investmentAmount, deposit.UserID.Hex())
		}
	}

	// Handle donation if applicable
	if deposit.DonationChoice != "none" && deposit.DonationChoice != "" {
		donationCollection := tq.db.Collection("donations")
		donationAmount := 0.0
		
		if deposit.DonationChoice == "both" {
			donationAmount = deposit.Amount
		} else if deposit.DonationChoice == "profit" {
			donationAmount = investmentAmount
		}

		if donationAmount > 0 {
			donation := bson.M{
				"userId":     deposit.UserID,
				"amount":     donationAmount,
				"type":       deposit.DonationChoice,
				"depositId":  deposit.ID,
				"status":     "pending",
				"createdAt":  time.Now(),
			}
			_, err := donationCollection.InsertOne(context.Background(), donation)
			if err != nil {
				log.Printf("Error creating donation for deposit %s: %v", deposit.ID.Hex(), err)
			} else {
				log.Printf("🎁 Created donation record: %.2f (%s) for user %s", donationAmount, deposit.DonationChoice, deposit.UserID.Hex())
			}
		}
	}
}
