package services

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"healthy_pay_backend/internal/utils"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Config
type StellarConfig struct {
	Network              string
	DistributorSecretKey string
	USDCContractAddress  string
	PublicKey            string
	SecretKey            string
}

type GlobalConfig struct {
	AppEnv  string
	Stellar StellarConfig
}

func LoadConfig() *GlobalConfig {
	// Load .env file from current directory
	loadEnvFile(".env")
	
	// Check both APP_ENV and environment variables
	appEnv := getEnv("APP_ENV", "dev")
	

	
	var network, usdcContract string
	
	if appEnv == "dev" {
		network = "https://horizon-testnet.stellar.org"
		usdcContract = getEnv("STELLAR_USDC_CONTRACT_ADDRESS", "GBBD47IF6LWK7P7MDEVSCWR7DPUWV3NY3DTQEVFL4NAT4AQH3ZLLFLA5")
	} else {
		network = "https://horizon.stellar.org"
		usdcContract = getEnv("STELLAR_USDC_CONTRACT_ADDRESS", "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN")
	}

	return &GlobalConfig{
		AppEnv: appEnv,
		Stellar: StellarConfig{
			Network:              network,
			DistributorSecretKey: getEnv("STELLAR_DISTRIBUTOR_SECRET_KEY", ""),
			USDCContractAddress:  usdcContract,
			PublicKey:            getEnv("STELLAR_PUBLIC_KEY", ""),
			SecretKey:            getEnv("STELLAR_SECRET_KEY", ""),
		},
	}
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // Silently ignore if .env file doesn't exist
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
			os.Setenv(key, value)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Models
type Chain string
type WalletType string

const (
	ChainStellar           Chain      = "STELLAR"
	WalletTypeNonCustodial WalletType = "NON_CUSTODIAL"
	WalletTypeCustodial    WalletType = "CUSTODIAL"
)

type WalletModel struct {
	PublicKey  string     `json:"public_key"`
	Address    string     `json:"address"`
	PrivateKey string     `json:"private_key,omitempty"`
	MnemonicPhrase string `json:"mnemonic_phrase,omitempty"`
	Chain      Chain      `json:"chain"`
	Type       WalletType `json:"type"`
}

type TransactionResult struct {
	TransactionHash string `json:"transaction_hash"`
	Status          bool   `json:"status"`
}

type TransactionDetails struct {
	SenderAddress      string  `json:"sender_address"`
	DestinationAddress string  `json:"destination_address"`
	TokenValue         float64 `json:"token_value"`
	TransactionHash    string  `json:"transaction_hash"`
	Asset              string  `json:"asset"`
	Memo               string  `json:"memo,omitempty"`
}

type MuxedWallet struct {
	MuxedAddress   string `json:"muxed_address"`
	BaseAddress    string `json:"base_address"`
	MuxedAccountID string `json:"muxed_account_id"`
	PublicKey      string `json:"public_key"`
	PrivateKey     string `json:"private_key"`
	Chain          Chain  `json:"chain"`
}

type SendUSDCRequest struct {
	FromUserID  primitive.ObjectID `json:"from_user_id"`
	ToAddress   string             `json:"to_address"`
	Amount      string             `json:"amount"`
}


// Service
type StellarService struct {
	client            *horizonclient.Client
	networkPassphrase string
	config            *GlobalConfig
}

func NewStellarService() *StellarService {
	cfg := LoadConfig()
	var client *horizonclient.Client
	var networkPassphrase string

	if cfg.AppEnv == "dev" {
		client = horizonclient.DefaultTestNetClient
		networkPassphrase = network.TestNetworkPassphrase
	} else {
		client = horizonclient.DefaultPublicNetClient
		networkPassphrase = network.PublicNetworkPassphrase
	}

	return &StellarService{
		client:            client,
		networkPassphrase: networkPassphrase,
		config:            cfg,
	}
}

// Core Stellar Operations
func (s *StellarService) SendToken(senderSecret, receiverPublicKey, amount string) (*TransactionResult, error) {
	return s.sendTokenWithRetry(senderSecret, receiverPublicKey, amount, "", 3)
}

func (s *StellarService) SendTokenWithMemo(senderSecret, receiverPublicKey, amount, memo string) (*TransactionResult, error) {
	return s.sendTokenWithRetry(senderSecret, receiverPublicKey, amount, memo, 3)
}

func (s *StellarService) sendTokenWithRetry(senderSecret, receiverPublicKey, amount, memo string, maxRetries int) (*TransactionResult, error) {
	baseDelay := 2 * time.Second
	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := s.sendTokenInternal(senderSecret, receiverPublicKey, amount, memo)
		if err == nil {
			return result, nil
		}
		if !s.isRetryableError(err) || attempt == maxRetries {
			return nil, err
		}
		delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func (s *StellarService) sendTokenInternal(senderSecret, receiverPublicKey, amount, memo string) (*TransactionResult, error) {
	if !strings.HasPrefix(senderSecret, "S") || len(senderSecret) != 56 {
		return nil, fmt.Errorf("invalid Stellar secret key format")
	}

	if s.config.AppEnv == "dev" {
		amount = "0.02"
	}

	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	amount = fmt.Sprintf("%.6f", amountFloat)

	senderKP, err := keypair.ParseFull(senderSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid sender secret key: %w", err)
	}

	sourceAccount, err := s.client.AccountDetail(horizonclient.AccountRequest{AccountID: senderKP.Address()})
	if err != nil {
		return nil, fmt.Errorf("failed to load sender account: %w", err)
	}

	usdcAsset := txnbuild.CreditAsset{Code: "USDC", Issuer: s.config.Stellar.USDCContractAddress}
	payment := &txnbuild.Payment{
		Destination:   receiverPublicKey,
		Amount:        amount,
		Asset:         usdcAsset,
		SourceAccount: senderKP.Address(),
	}

	txParams := txnbuild.TransactionParams{
		SourceAccount:        &sourceAccount,
		IncrementSequenceNum: true,
		Operations:           []txnbuild.Operation{payment},
		BaseFee:              txnbuild.MinBaseFee,
		Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(180)},
	}

	if memo != "" {
		txParams.Memo = txnbuild.MemoText(memo)
	}

	tx, err := txnbuild.NewTransaction(txParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	tx, err = tx.Sign(s.networkPassphrase, senderKP)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	resp, err := s.client.SubmitTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	return &TransactionResult{TransactionHash: resp.Hash, Status: resp.Successful}, nil
}

func (s *StellarService) GetUSDCBalance(accountPublicKey string) (string, error) {
	account, err := s.client.AccountDetail(horizonclient.AccountRequest{AccountID: accountPublicKey})
	if err != nil {
		return "0", fmt.Errorf("failed to load account: %w", err)
	}

	for _, balance := range account.Balances {
		if balance.Asset.Type == "credit_alphanum4" &&
			balance.Asset.Code == "USDC" &&
			balance.Asset.Issuer == s.config.Stellar.USDCContractAddress {
			return balance.Balance, nil
		}
	}
	return "0", nil
}

func (s *StellarService) CreateAccount(destinationPublicKey string) (string, error) {
	distributorKP, err := keypair.ParseFull(s.config.Stellar.DistributorSecretKey)
	if err != nil {
		return "", fmt.Errorf("invalid distributor secret: %w", err)
	}

	sourceAccount, err := s.client.AccountDetail(horizonclient.AccountRequest{AccountID: distributorKP.Address()})
	if err != nil {
		return "", fmt.Errorf("failed to load distributor account: %w", err)
	}

	createAccount := &txnbuild.CreateAccount{Destination: destinationPublicKey, Amount: "2.0"}
	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &sourceAccount,
		IncrementSequenceNum: true,
		Operations:           []txnbuild.Operation{createAccount},
		BaseFee:              txnbuild.MinBaseFee,
		Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
	})
	if err != nil {
		return "", fmt.Errorf("failed to build transaction: %w", err)
	}

	tx, err = tx.Sign(s.networkPassphrase, distributorKP)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	resp, err := s.client.SubmitTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}
	return resp.Hash, nil
}

func (s *StellarService) EstablishTrustLine(userPublicKey, userPrivateKey string) error {
	if s.config.Stellar.DistributorSecretKey == "" {
		return fmt.Errorf("STELLAR_DISTRIBUTOR_SECRET_KEY is not configured")
	}

	// Create keypairs - Parse returns Full keypairs when given secret keys
	userKp, err := keypair.ParseFull(userPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to parse user keypair: %v", err)
	}

	sponsorKp, err := keypair.ParseFull(s.config.Stellar.DistributorSecretKey)
	if err != nil {
		return fmt.Errorf("failed to parse sponsor keypair: %v", err)
	}

	// Create Horizon client
	var networkPassphrase string
	if s.config.AppEnv == "dev" {
		networkPassphrase = network.TestNetworkPassphrase
	} else {
		networkPassphrase = network.PublicNetworkPassphrase
	}

	client := horizonclient.DefaultTestNetClient
	if s.config.AppEnv != "dev" {
		client = horizonclient.DefaultPublicNetClient
	}

	// Load sponsor account
	sponsorAccount, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: sponsorKp.Address(),
	})
	if err != nil {
		return fmt.Errorf("failed to load sponsor account: %v", err)
	}

	// Create USDC asset
	usdcAsset := txnbuild.CreditAsset{
		Code:   "USDC",
		Issuer: s.config.Stellar.USDCContractAddress,
	}

	// Build sponsored account creation and trustline transaction
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &sponsorAccount,
			IncrementSequenceNum: true,
			Operations: []txnbuild.Operation{
				// Begin sponsorship
				&txnbuild.BeginSponsoringFutureReserves{
					SponsoredID: userKp.Address(),
				},
				// Create account on-chain
				&txnbuild.CreateAccount{
					Destination: userKp.Address(),
					Amount:      "0", // No XLM needed from user
				},
				// End sponsorship for account creation
				&txnbuild.EndSponsoringFutureReserves{
					SourceAccount: userKp.Address(),
				},
				// Begin sponsorship for trustline
				&txnbuild.BeginSponsoringFutureReserves{
					SponsoredID: userKp.Address(),
				},
				// Create USDC trustline
				&txnbuild.ChangeTrust{
					Line:          txnbuild.ChangeTrustAssetWrapper{Asset: usdcAsset},
					Limit:         "1000000",
					SourceAccount: userKp.Address(),
				},
				// End sponsorship for trustline
				&txnbuild.EndSponsoringFutureReserves{
					SourceAccount: userKp.Address(),
				},
			},
			BaseFee:       txnbuild.MinBaseFee,
			Preconditions: txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to build transaction: %v", err)
	}

	// Sign transaction - use SignWithKeyString or direct signing
	tx, err = tx.Sign(networkPassphrase, sponsorKp, userKp)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Submit transaction
	_, err = client.SubmitTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to submit sponsored account creation transaction: %v", err)
	}

	return nil
}




func (s *StellarService) StreamPayments(ctx context.Context, callback func(*TransactionDetails)) error {
	cursor := "now"
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			payments, err := s.client.Payments(horizonclient.OperationRequest{
				ForAccount: s.config.Stellar.PublicKey,
				Cursor:     cursor,
				Order:      horizonclient.OrderAsc,
				Limit:      10,
			})
			if err != nil {
				log.Printf("Error fetching payments: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, payment := range payments.Embedded.Records {
				if payment.GetType() == "payment" {
					details := &TransactionDetails{
						TransactionHash: payment.GetTransactionHash(),
						Asset:           "USDC",
					}
					callback(details)
				}
				cursor = payment.PagingToken()
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// Wallet Operations
func (s *StellarService) CreateActiveAccount() (*WalletModel, error) {
	pair, err := keypair.Random()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}
	address:= pair.Address()
	privateKey:= pair.Seed()
	mnemonic := utils.GenerateMnemonic()

	err = s.EstablishTrustLine(address, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to establish trust line: %w", err)
	}

	encryptedPrivateKey, err := utils.EncryptPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %v", err)
	}

	encryptedMnemonic, err := utils.EncryptPrivateKey(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt mnemonic: %v", err)
	}

	return &WalletModel{
		PublicKey:  address,
		Address:    address,
		PrivateKey: encryptedPrivateKey,
		MnemonicPhrase: encryptedMnemonic,
		Chain:      ChainStellar,
		Type:       WalletTypeNonCustodial,
	}, nil
}



func (s *StellarService) SponsorAccount(userWallet *keypair.Full) error {
	if s.config.Stellar.DistributorSecretKey == "" {
		return fmt.Errorf("STELLAR_DISTRIBUTOR_SECRET_KEY is not configured")
	}

	maxRetries := 3
	baseDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := s.sponsorAccountInternal(userWallet)
		if err == nil {
			log.Printf("Transaction submitted successfully on attempt %d", attempt)
			return nil
		}

		log.Printf("Error in sponsorAccount (attempt %d/%d): %v", attempt, maxRetries, err)

		isRetryable := s.isRetryableError(err)
		if !isRetryable || attempt == maxRetries {
			return err
		}

		delay := time.Duration(float64(baseDelay)*math.Pow(2, float64(attempt-1))) + time.Duration(float64(time.Second)*0.001*1000)
		log.Printf("Retrying in %v...", delay)
		time.Sleep(delay)
	}
	return fmt.Errorf("max retries exceeded")
}

func (s *StellarService) sponsorAccountInternal(userWallet *keypair.Full) error {
	newUser := userWallet
	sponsor, err := keypair.ParseFull(s.config.Stellar.DistributorSecretKey)
	if err != nil {
		return fmt.Errorf("invalid distributor secret: %w", err)
	}

	sponsorAccount, err := s.client.AccountDetail(horizonclient.AccountRequest{AccountID: sponsor.Address()})
	if err != nil {
		return fmt.Errorf("failed to load sponsor account: %w", err)
	}

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &sponsorAccount,
		IncrementSequenceNum: true,
		Operations: []txnbuild.Operation{
			&txnbuild.CreateAccount{Destination: newUser.Address(), Amount: "1.0"},
		},
		BaseFee:       txnbuild.MinBaseFee,
		Preconditions: txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(180)},
	})
	if err != nil {
		return fmt.Errorf("failed to build transaction: %w", err)
	}

	// Only sponsor signs for account creation
	tx, err = tx.Sign(s.networkPassphrase, sponsor)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	_, err = s.client.SubmitTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}
	return nil
}

func (s *StellarService) GenerateMuxedAddress(userID string) (*MuxedWallet, error) {
	if s.config.Stellar.SecretKey == "" || s.config.Stellar.PublicKey == "" {
		return nil, fmt.Errorf("STELLAR_SECRET_KEY and STELLAR_PUBLIC_KEY must be configured")
	}

	muxedAccountID := s.generateMuxedAccountID(userID)
	muxedAddress := fmt.Sprintf("M%s%d", s.config.Stellar.PublicKey[1:], muxedAccountID)

	return &MuxedWallet{
		MuxedAddress:   muxedAddress,
		BaseAddress:    s.config.Stellar.PublicKey,
		MuxedAccountID: fmt.Sprintf("%d", muxedAccountID),
		PublicKey:      s.config.Stellar.PublicKey,
		PrivateKey:     s.config.Stellar.SecretKey,
		Chain:          ChainStellar,
	}, nil
}

func (s *StellarService) CreateCustodialWallet(userID string) (*WalletModel, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID must be a non-empty string")
	}

	muxedWallet, err := s.GenerateMuxedAddress(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate muxed address: %w", err)
	}

	return &WalletModel{
		PublicKey: muxedWallet.PublicKey,
		Address:   muxedWallet.MuxedAddress,
		Chain:     ChainStellar,
		Type:      WalletTypeCustodial,
	}, nil
}

func (s *StellarService) generateMuxedAccountID(userID string) uint64 {
	hash := sha256.Sum256([]byte(userID))
	muxedID := binary.BigEndian.Uint64(hash[:8]) % 1000000000
	if muxedID == 0 {
		muxedID = 1
	}
	return muxedID
}

func (s *StellarService) IsMuxedAddress(address string) bool {
	return address != "" && address[0] == 'M'
}

func (s *StellarService) isRetryableError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "504") ||
		strings.Contains(errStr, "ECONNRESET") ||
		strings.Contains(errStr, "ETIMEDOUT")
}


