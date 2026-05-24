package web3

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"

	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
)

type FaucetService struct {
	signerKey  *ecdsa.PrivateKey
	signerAddr common.Address
	domainSep  [32]byte
	domain     EIP712Domain
	redis      *redis.Client
	client     *ethclient.Client
	rpcURL     string
	done       chan struct{}
}

type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract common.Address
}

type FaucetClaimPayload struct {
	Wallet   common.Address
	Amount   *big.Int
	Nonce    *big.Int
	Deadline *big.Int
}

type FaucetClaimResponse struct {
	Wallet     string `json:"wallet"`
	Amount     string `json:"amount"`
	Nonce      string `json:"nonce"`
	Deadline   string `json:"deadline"`
	Signature  string `json:"signature"`
	ChainID    int64  `json:"chain_id"`
	FaucetAddr string `json:"faucet_addr"`
}

func NewFaucetService(signerKeyHex string, chainID *big.Int, contractAddr common.Address, rdb *redis.Client, rpcURL string) (*FaucetService, error) {
	keyHex := strings.TrimPrefix(signerKeyHex, "0x")
	priv, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		return nil, err
	}

	signerAddr := crypto.PubkeyToAddress(priv.PublicKey)

	domain := EIP712Domain{
		Name:              "ShortURL Faucet",
		Version:           "1",
		ChainID:           chainID,
		VerifyingContract: contractAddr,
	}

	domainSep := computeDomainSeparator(domain)

	logger.Ctx(context.Background()).Info("Faucet signer initialized",
		zap.String("address", signerAddr.Hex()),
		zap.String("contract", contractAddr.Hex()),
		zap.String("chain_id", chainID.String()),
	)

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &FaucetService{
		signerKey:  priv,
		signerAddr: signerAddr,
		domainSep:  domainSep,
		domain:     domain,
		redis:      rdb,
		client:     ethClient,
		rpcURL:     rpcURL,
		done:       make(chan struct{}),
	}, nil
}

func (s *FaucetService) SignerAddress() common.Address {
	return s.signerAddr
}

func (s *FaucetService) signTypedData(wallet common.Address, amount, nonce, deadline *big.Int) ([]byte, error) {
	structHash := crypto.Keccak256Hash(encodeFaucetClaim(wallet, amount, nonce, deadline))

	buf := new(bytes.Buffer)
	buf.Write([]byte{0x19, 0x01})
	buf.Write(s.domainSep[:])
	buf.Write(structHash.Bytes())

	digest := crypto.Keccak256Hash(buf.Bytes())

	sig, err := crypto.Sign(digest.Bytes(), s.signerKey)
	if err != nil {
		return nil, err
	}

	sig[64] += 27
	return sig, nil
}

func (s *FaucetService) Claim(ctx context.Context, walletAddr string, faucetAmount int64) (*FaucetClaimResponse, error) {
	wallet := common.HexToAddress(walletAddr)

	amount := new(big.Int).Mul(big.NewInt(faucetAmount), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("faucetBalance()"))
	selector := h.Sum(nil)[:4]

	msg := ethereum.CallMsg{
		To:   &s.domain.VerifyingContract,
		Data: selector,
	}

	result, err := s.client.CallContract(ctx, msg, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to query faucet contract balance", zap.Error(err))
		return nil, err
	}

	contractBalance := new(big.Int).SetBytes(result)
	if contractBalance.Cmp(amount) < 0 {
		return nil, &FaucetExhaustedError{
			FaucetAddress: s.domain.VerifyingContract.Hex(),
			Message:       "insufficient ERC20 token balance in faucet contract",
		}
	}

	nonce, err := s.GetNonce(ctx, walletAddr)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get faucet nonce", zap.String("wallet", walletAddr), zap.Error(err))
		return nil, err
	}

	deadline := big.NewInt(time.Now().Add(constants.FaucetDeadline).Unix())

	sig, err := s.signTypedData(wallet, amount, nonce, deadline)
	if err != nil {
		logger.Ctx(ctx).Error("Faucet signing failed", zap.String("wallet", walletAddr), zap.Error(err))
		return nil, err
	}

	logger.Ctx(ctx).Info("Faucet claim signed",
		zap.String("wallet", walletAddr),
		zap.String("amount", amount.String()),
	)

	return &FaucetClaimResponse{
		Wallet:     walletAddr,
		Amount:     amount.String(),
		Nonce:      nonce.String(),
		Deadline:   deadline.String(),
		Signature:  hexutil.Encode(sig),
		FaucetAddr: s.domain.VerifyingContract.Hex(),
		ChainID:    s.domain.ChainID.Int64(),
	}, nil
}

func (s *FaucetService) VerifyClaim(ctx context.Context, txHash, walletAddr string, expectedAmount *big.Int) error {
	var receipt struct {
		Status        string   `json:"status"`
		To            string   `json:"to"`
		TransactionHash string `json:"transactionHash"`
		Logs          []struct {
			Address     string   `json:"address"`
			Topics      []string `json:"topics"`
			Data        string   `json:"data"`
		} `json:"logs"`
	}

	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"params":  []interface{}{txHash},
		"id":      1,
	}

	if err := rpcCall(ctx, s.rpcURL, body, &receipt); err != nil {
		return fmt.Errorf("failed to fetch tx receipt: %w", err)
	}

	if receipt.Status == "" {
		return fmt.Errorf("transaction receipt not found: %s", txHash)
	}

	status := parseHexUint64(receipt.Status)
	if status != 1 {
		return fmt.Errorf("transaction reverted on-chain: %s", txHash)
	}

	faucetAddr := s.domain.VerifyingContract.Hex()
	if !strings.EqualFold(receipt.To, faucetAddr) {
		return fmt.Errorf("tx recipient is not faucet contract: got %s, want %s", receipt.To, faucetAddr)
	}

	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("ClaimRequested(address,uint256,uint256)"))
	claimEventTopic := "0x" + fmt.Sprintf("%x", h.Sum(nil))

	for _, log := range receipt.Logs {
		if !strings.EqualFold(log.Address, faucetAddr) {
			continue
		}
		if len(log.Topics) == 0 || log.Topics[0] != claimEventTopic {
			continue
		}
		if len(log.Topics) < 2 {
			continue
		}

		eventWallet := decodeAddress(log.Topics[1])
		if !strings.EqualFold(eventWallet, walletAddr) {
			return fmt.Errorf("claim event wallet mismatch: got %s, want %s", eventWallet, walletAddr)
		}

		if len(log.Data) < 128 {
			return fmt.Errorf("claim event data too short: %s", log.Data)
		}
		eventAmount := decodeUint256(log.Data)

		if eventAmount.Cmp(expectedAmount) != 0 {
			return fmt.Errorf("claim event amount mismatch: got %s, want %s", eventAmount.String(), expectedAmount.String())
		}

		return nil
	}

	return fmt.Errorf("no valid ClaimRequested event found for wallet %s in tx %s", walletAddr, txHash)
}

func (s *FaucetService) SendDevETH(ctx context.Context, walletAddr string) (string, error) {
	wallet := common.HexToAddress(walletAddr)

	balance, err := s.client.BalanceAt(ctx, s.signerAddr, nil)
	if err != nil {
		return "", err
	}

	value := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	gasLimit := uint64(constants.DevETHGasLimit)

	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	required := new(big.Int).Add(value, new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice))
	if balance.Cmp(required) < 0 {
		return "", &FaucetExhaustedError{
			FaucetAddress: s.signerAddr.Hex(),
			Message:       "insufficient native ETH balance for dev claim",
		}
	}

	nonce, err := s.client.PendingNonceAt(ctx, s.signerAddr)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, wallet, value, gasLimit, gasPrice, nil)

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(s.domain.ChainID), s.signerKey)
	if err != nil {
		return "", err
	}

	err = s.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

func (s *FaucetService) GetNonce(ctx context.Context, walletAddr string) (*big.Int, error) {
	wallet := common.HexToAddress(walletAddr)

	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("nonces(address)"))
	selector := h.Sum(nil)[:4]

	data := make([]byte, 36)
	copy(data[:4], selector)
	copy(data[4:36], common.LeftPadBytes(wallet.Bytes(), 32))

	msg := ethereum.CallMsg{
		To:   &s.domain.VerifyingContract,
		Data: data,
	}

	result, err := s.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(result), nil
}

func (s *FaucetService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *FaucetService) Start(ctx context.Context) {
	logger.Ctx(ctx).Info("Faucet service started")
	ticker := time.NewTicker(constants.FaucetHeartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logger.Ctx(ctx).Debug("Faucet service heartbeat")
		case <-ctx.Done():
			logger.Ctx(ctx).Info("Faucet service stopped")
			return
		case <-s.done:
			return
		}
	}
}

func (s *FaucetService) Stop() {
	close(s.done)
}
