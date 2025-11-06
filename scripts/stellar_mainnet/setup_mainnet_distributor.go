package main

import (
	"fmt"
	"log"

	"github.com/stellar/go/keypair"
)

func main() {
	fmt.Println("🌟 Setting up Stellar Mainnet Distributor Account")
	fmt.Println("=================================================")

	// Generate new keypair for mainnet distributor
	kp, err := keypair.Random()
	if err != nil {
		log.Fatal("Failed to generate keypair:", err)
	}

	fmt.Printf("🔑 Generated Mainnet Distributor Keys:\n")
	fmt.Printf("Public Key:  %s\n", kp.Address())
	fmt.Printf("Secret Key:  %s\n", kp.Seed())

	fmt.Printf("\n⚠️  IMPORTANT: MAINNET ACCOUNT FUNDING REQUIRED\n")
	fmt.Printf("===============================================\n")
	fmt.Printf("1. Send XLM to this address: %s\n", kp.Address())
	fmt.Printf("2. Minimum recommended: 100 XLM for sponsorship operations\n")
	fmt.Printf("3. Each sponsored account creation costs ~1 XLM\n")
	fmt.Printf("4. Update .env file with these keys after funding\n")

	fmt.Printf("\n📝 Environment Configuration:\n")
	fmt.Printf("STELLAR_NETWORK=mainnet\n")
	fmt.Printf("STELLAR_DISTRIBUTOR_SECRET_KEY=%s\n", kp.Seed())
	fmt.Printf("STELLAR_DISTRIBUTOR_PUBLIC_KEY=%s\n", kp.Address())

	fmt.Printf("\n🔍 To verify funding, check:\n")
	fmt.Printf("https://stellar.expert/explorer/public/account/%s\n", kp.Address())

	fmt.Printf("\n⚠️  WARNING: This is MAINNET - real money involved!\n")
	fmt.Printf("Only proceed if you understand the costs and risks.\n")
}
