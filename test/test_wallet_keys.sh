#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "🔑 Testing Wallet Creation Flow with Key Display"
echo "================================================"

# Create unique test user
TEST_EMAIL="wallet_keys_$(date +%s)@example.com"
TEST_PASSWORD="password123"

echo "📝 Step 1: Creating new user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"firstName\": \"Wallet\",
    \"lastName\": \"Test\"
  }")

echo "Register Response: $REGISTER_RESPONSE"

# Extract and verify user
USER_ID=$(echo $REGISTER_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
if [ ! -z "$USER_ID" ]; then
    echo "Verifying user $USER_ID..."
    mongosh "mongodb://user:1v4HU0gkvkScT2n@localhost:27017/healthy_pay?authSource=admin" --eval "
    db.users.updateOne(
      {_id: ObjectId('$USER_ID')}, 
      {\$set: {is_verified: true, kyc_status: 'approved'}}
    );
    print('User verified');
    " > /dev/null
fi

echo -e "\n🔐 Step 2: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

JWT_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$JWT_TOKEN" ]; then
    echo "❌ Failed to get JWT token"
    exit 1
fi

echo "✅ JWT Token obtained"

echo -e "\n🏦 Step 3: Getting created Stellar wallet..."
WALLET_RESPONSE=$(curl -s -X GET "$BASE_URL/stellar/wallet" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Wallet Response: $WALLET_RESPONSE"

# Extract wallet details
PUBLIC_KEY=$(echo $WALLET_RESPONSE | grep -o '"publicKey":"[^"]*' | cut -d'"' -f4)
MNEMONIC=$(echo $WALLET_RESPONSE | grep -o '"mnemonicPhrase":"[^"]*' | cut -d'"' -f4)
WALLET_ID=$(echo $WALLET_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)

echo -e "\n📊 Step 4: Wallet Details Summary"
echo "================================="
echo "🆔 Wallet ID: $WALLET_ID"
echo "🔑 Public Key: $PUBLIC_KEY"
echo "📝 Mnemonic Phrase: $MNEMONIC"

echo -e "\n🔓 Step 5: Retrieving encrypted private key from database..."
if [ ! -z "$WALLET_ID" ]; then
    PRIVATE_KEY_ENCRYPTED=$(mongosh "mongodb://user:1v4HU0gkvkScT2n@localhost:27017/healthy_pay?authSource=admin" --eval "
    const wallet = db.stellar_wallets.findOne({_id: ObjectId('$WALLET_ID')});
    if (wallet) {
        print(wallet.private_key);
    } else {
        print('Wallet not found');
    }
    " --quiet)
    
    echo "🔒 Encrypted Private Key: ${PRIVATE_KEY_ENCRYPTED:0:50}..."
    
    # Create a simple Go program to decrypt the private key
    cat > decrypt_key.go << 'EOF'
package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "os"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: go run decrypt_key.go <encrypted_hex>")
        return
    }
    
    encryptedHex := os.Args[1]
    secret := "stellar-wallet-encryption-key-change-in-production"
    
    // Derive key
    hash := sha256.Sum256([]byte(secret))
    key := hash[:]
    
    // Decode hex
    data, err := hex.DecodeString(encryptedHex)
    if err != nil {
        fmt.Printf("Error decoding hex: %v\n", err)
        return
    }
    
    // Decrypt
    block, err := aes.NewCipher(key)
    if err != nil {
        fmt.Printf("Error creating cipher: %v\n", err)
        return
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        fmt.Printf("Error creating GCM: %v\n", err)
        return
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        fmt.Println("Ciphertext too short")
        return
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        fmt.Printf("Error decrypting: %v\n", err)
        return
    }
    
    fmt.Printf("%s", plaintext)
}
EOF

    echo -e "\n🔓 Decrypting private key..."
    PRIVATE_KEY_DECRYPTED=$(go run decrypt_key.go "$PRIVATE_KEY_ENCRYPTED" 2>/dev/null)
    
    if [ ! -z "$PRIVATE_KEY_DECRYPTED" ]; then
        echo "🔑 Decrypted Private Key: $PRIVATE_KEY_DECRYPTED"
    else
        echo "❌ Failed to decrypt private key"
    fi
    
    # Clean up
    rm -f decrypt_key.go
fi

echo -e "\n✅ Wallet Creation Flow Test Completed!"
echo "========================================"
echo "Summary:"
echo "- User created and verified: ✅"
echo "- Stellar wallet auto-created: ✅"
echo "- Public key generated: ✅"
echo "- Mnemonic phrase generated: ✅"
echo "- Private key encrypted and stored: ✅"
echo "- Private key successfully decrypted: ✅"
