package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
)

func main() {
	fmt.Println("🌟 Setting up Stellar Testnet Distributor Account")
	fmt.Println("=================================================")

	// Generate new keypair for distributor
	kp, err := keypair.Random()
	if err != nil {
		log.Fatal("Failed to generate keypair:", err)
	}

	fmt.Printf("🔑 Generated Distributor Keys:\n")
	fmt.Printf("Public Key:  %s\n", kp.Address())
	fmt.Printf("Secret Key:  %s\n", kp.Seed())

	// Fund account using Stellar testnet friendbot
	fmt.Printf("\n💰 Funding account via Stellar Friendbot...\n")
	
	friendbotURL := fmt.Sprintf("https://friendbot.stellar.org?addr=%s", kp.Address())
	resp, err := http.Get(friendbotURL)
	if err != nil {
		log.Fatal("Failed to fund account:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ Account funded successfully!\n")
	} else {
		fmt.Printf("❌ Failed to fund account. Status: %d\n", resp.StatusCode)
		return
	}

	// Verify account exists on testnet
	fmt.Printf("\n🔍 Verifying account on Stellar testnet...\n")
	
	client := horizonclient.DefaultTestNetClient
	account, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: kp.Address(),
	})
	if err != nil {
		log.Fatal("Failed to get account details:", err)
	}

	fmt.Printf("✅ Account verified on testnet!\n")
	fmt.Printf("Account ID: %s\n", account.AccountID)
	fmt.Printf("Sequence: %d\n", account.Sequence)
	
	// Display balances
	fmt.Printf("\n💰 Account Balances:\n")
	for _, balance := range account.Balances {
		if balance.Asset.Type == "native" {
			fmt.Printf("XLM: %s\n", balance.Balance)
		}
	}

	fmt.Printf("\n📝 Update your .env file with these keys:\n")
	fmt.Printf("STELLAR_DISTRIBUTOR_SECRET_KEY=%s\n", kp.Seed())
	fmt.Printf("STELLAR_DISTRIBUTOR_PUBLIC_KEY=%s\n", kp.Address())

	fmt.Printf("\n✅ Distributor account setup complete!\n")
	fmt.Printf("This account can now sponsor wallet creation on Stellar testnet.\n")
}
