package redirect

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverSigner(t *testing.T) {
	// Generate a random private key for signing
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	pubKey := privKey.Public().(*ecdsa.PublicKey)
	expectedAddress := crypto.PubkeyToAddress(*pubKey).Hex()

	message := "Prove ownership of NFT Pass for /test-slug at 1718320000"

	// Prepare EIP-191 hash (exactly as Ethereum wallets do client-side)
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	prefixedMessage := append([]byte(prefix), []byte(message)...)
	hash := crypto.Keccak256(prefixedMessage)

	// Sign the hash
	sig, err := crypto.Sign(hash, privKey)
	require.NoError(t, err)

	// Adjust V value to be 27/28 to simulate MetaMask/ethers browser signature
	sig[64] += 27

	sigHex := hexutil.Encode(sig)

	// Recover the signer address using our Go helper
	recoveredAddress, err := RecoverSigner(message, sigHex)
	require.NoError(t, err)
	assert.Equal(t, expectedAddress, recoveredAddress)
}

func TestVerifyNFTBypassChallenge_Expired(t *testing.T) {
	// Test expired timestamp (skew > 300 seconds)
	expiredTimestamp := time.Now().Unix() - 600
	message := "Prove ownership of NFT Pass for /test-slug at " + strconv.FormatInt(expiredTimestamp, 10)

	_, err := VerifyNFTBypassChallenge(context.Background(), "http://localhost:8545", "0x0165878A594ca255338adfa4d48449f69242Eb8F", "test-slug", message, "0xsignature")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "challenge expired")
}

func TestVerifyNFTBypassChallenge_InvalidPrefix(t *testing.T) {
	message := "Invalid message prefix for /test-slug at " + strconv.FormatInt(time.Now().Unix(), 10)

	_, err := VerifyNFTBypassChallenge(context.Background(), "http://localhost:8545", "0x0165878A594ca255338adfa4d48449f69242Eb8F", "test-slug", message, "0xsignature")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid message format: prefix mismatch")
}
