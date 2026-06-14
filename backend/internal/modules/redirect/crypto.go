package redirect

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// RecoverSigner recovers the Ethereum address from the EIP-191 message and signature
func RecoverSigner(message string, sigHex string) (string, error) {
	sigHex = strings.TrimPrefix(sigHex, "0x")
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return "", fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(sig) != 65 {
		return "", fmt.Errorf("invalid signature length: expected 65, got %d", len(sig))
	}

	// Adjust V value to be 0 or 1 for go-ethereum
	if sig[64] == 27 || sig[64] == 28 {
		sig[64] -= 27
	} else if sig[64] != 0 && sig[64] != 1 {
		return "", fmt.Errorf("invalid signature V value: %d", sig[64])
	}

	// Prepare EIP-191 message prefix
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	prefixedMessage := append([]byte(prefix), []byte(message)...)
	hash := crypto.Keccak256(prefixedMessage)

	// Recover public key
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key: %w", err)
	}

	// Convert public key to address
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr.Hex(), nil
}

// QueryNFTPassBalance queries the balanceOf method on the NFTPass contract
func QueryNFTPassBalance(ctx context.Context, rpcURL string, nftContractAddress string, userAddress string) (*big.Int, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	if !common.IsHexAddress(userAddress) {
		return nil, fmt.Errorf("invalid user address: %s", userAddress)
	}
	if !common.IsHexAddress(nftContractAddress) {
		return nil, fmt.Errorf("invalid nft contract address: %s", nftContractAddress)
	}

	userAddr := common.HexToAddress(userAddress)
	toAddr := common.HexToAddress(nftContractAddress)

	// Selector for balanceOf(address) is 0x70a08231
	data := make([]byte, 36)
	copy(data[0:4], []byte{0x70, 0xa0, 0x82, 0x31})
	copy(data[4:36], common.LeftPadBytes(userAddr.Bytes(), 32))

	msg := ethereum.CallMsg{
		To:   &toAddr,
		Data: data,
	}

	res, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("contract call failed: %w", err)
	}

	if len(res) == 0 {
		return big.NewInt(0), nil
	}

	return new(big.Int).SetBytes(res), nil
}

// VerifyNFTBypassChallenge validates the challenge message skew and recovers/verifies the signer balance
func VerifyNFTBypassChallenge(ctx context.Context, rpcURL string, nftContract string, slug string, message string, signature string) (string, error) {
	expectedPrefix := "Prove ownership of NFT Pass for /" + slug + " at "
	if !strings.HasPrefix(message, expectedPrefix) {
		return "", fmt.Errorf("invalid message format: prefix mismatch")
	}

	timestampStr := strings.TrimPrefix(message, expectedPrefix)
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Prevent replay attacks by checking timestamp skew (max 5 minutes)
	timeSkew := time.Now().Unix() - timestamp
	if timeSkew < 0 {
		timeSkew = -timeSkew
	}
	if timeSkew > 300 {
		return "", fmt.Errorf("challenge expired: time skew too large (%d seconds)", timeSkew)
	}

	// Recover the signer address
	signerAddress, err := RecoverSigner(message, signature)
	if err != nil {
		return "", fmt.Errorf("failed to recover signer: %w", err)
	}

	// Query blockchain balance
	balance, err := QueryNFTPassBalance(ctx, rpcURL, nftContract, signerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to check NFT balance: %w", err)
	}

	if balance.Sign() <= 0 {
		return "", fmt.Errorf("address does not own the NFT Pass")
	}

	return signerAddress, nil
}
