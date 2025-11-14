package main

import (
	"fmt"
	"log"
	"healthy_pay_backend/internal/services"
	"healthy_pay_backend/internal/utils"
)

func main() {
	service := services.NewStellarService()
	wallet, err := service.CreateActiveAccount()
	if err != nil {
		log.Fatal(err)
	}
	
	privateKey, _ := utils.DecryptPrivateKey(wallet.PrivateKey)
	mnemonic, _ := utils.DecryptPrivateKey(wallet.MnemonicPhrase)
	
	fmt.Printf("Public Key: %s\n", wallet.PublicKey)
	fmt.Printf("Private Key: %s\n", privateKey)
	fmt.Printf("Mnemonic: %s\n", mnemonic)
}
