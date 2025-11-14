package main

import (
	"fmt"
	"log"
	"healthy_pay_backend/internal/services"
	"healthy_pay_backend/internal/utils"
)

func main() {
	fmt.Println("Testing CreateActiveAccount function...")
	
	// Initialize Stellar service
	stellarService := services.NewStellarService()
	
	// Create active account
	wallet, err := stellarService.CreateActiveAccount()
	if err != nil {
		log.Fatalf("Failed to create active account: %v", err)
	}
	
	// Decrypt private key and mnemonic for display
	decryptedPrivateKey, err := utils.DecryptPrivateKey(wallet.PrivateKey)
	if err != nil {
		log.Printf("Warning: Could not decrypt private key: %v", err)
		decryptedPrivateKey = "ENCRYPTED: " + wallet.PrivateKey
	}
	
	decryptedMnemonic, err := utils.DecryptPrivateKey(wallet.MnemonicPhrase)
	if err != nil {
		log.Printf("Warning: Could not decrypt mnemonic: %v", err)
		decryptedMnemonic = "ENCRYPTED: " + wallet.MnemonicPhrase
	}
	
	// Display account details
	fmt.Println("\n=== STELLAR ACCOUNT CREATED ===")
	fmt.Printf("Public Key: %s\n", wallet.PublicKey)
	fmt.Printf("Private Key: %s\n", decryptedPrivateKey)
	fmt.Printf("Mnemonic Phrase: %s\n", decryptedMnemonic)
	fmt.Printf("Chain: %s\n", wallet.Chain)
	fmt.Printf("Type: %s\n", wallet.Type)
	fmt.Println("================================")
}
