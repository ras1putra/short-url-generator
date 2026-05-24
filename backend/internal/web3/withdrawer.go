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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"

	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
)

type WithdrawerService struct {
	signerKey  *ecdsa.PrivateKey
	signerAddr common.Address
	domainSep  [32]byte
	domain     EIP712Domain
	client     *ethclient.Client
	rpcURL     string
}

type WithdrawalPermit struct {
	Wallet    string `json:"wallet"`
	Amount    string `json:"amount"`
	Nonce     string `json:"nonce"`
	Deadline  string `json:"deadline"`
	Signature string `json:"signature"`
	ChainID   int64  `json:"chain_id"`
	Contract  string `json:"contract"`
}

func NewWithdrawerService(signerKeyHex string, chainID *big.Int, contractAddr common.Address, rpcURL string) (*WithdrawerService, error) {
	keyHex := strings.TrimPrefix(signerKeyHex, "0x")
	priv, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		return nil, err
	}

	signerAddr := crypto.PubkeyToAddress(priv.PublicKey)

	domain := EIP712Domain{
		Name:              "ShortURL Withdrawer",
		Version:           "1",
		ChainID:           chainID,
		VerifyingContract: contractAddr,
	}

	domainSep := computeDomainSeparator(domain)

	logger.Ctx(context.Background()).Info("Withdrawer signer initialized",
		zap.String("address", signerAddr.Hex()),
		zap.String("contract", contractAddr.Hex()),
		zap.String("chain_id", chainID.String()),
	)

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &WithdrawerService{
		signerKey:  priv,
		signerAddr: signerAddr,
		domainSep:  domainSep,
		domain:     domain,
		client:     ethClient,
		rpcURL:     rpcURL,
	}, nil
}

func (s *WithdrawerService) SignerAddress() common.Address {
	return s.signerAddr
}

func (s *WithdrawerService) signTypedData(wallet common.Address, amount, nonce, deadline *big.Int) ([]byte, error) {
	structHash := crypto.Keccak256Hash(encodeWithdrawal(wallet, amount, nonce, deadline))

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

func (s *WithdrawerService) CreatePermit(ctx context.Context, walletAddr string, amount *big.Int) (*WithdrawalPermit, error) {
	wallet := common.HexToAddress(walletAddr)

	nonce, err := s.GetNonce(ctx, walletAddr)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get withdrawal nonce", zap.String("wallet", walletAddr), zap.Error(err))
		return nil, err
	}

	deadline := big.NewInt(time.Now().Add(constants.FaucetDeadline).Unix())

	sig, err := s.signTypedData(wallet, amount, nonce, deadline)
	if err != nil {
		logger.Ctx(ctx).Error("Withdrawal signing failed", zap.String("wallet", walletAddr), zap.Error(err))
		return nil, err
	}

	logger.Ctx(ctx).Info("Withdrawal permit signed",
		zap.String("wallet", walletAddr),
		zap.String("amount", amount.String()),
	)

	return &WithdrawalPermit{
		Wallet:    walletAddr,
		Amount:    amount.String(),
		Nonce:     nonce.String(),
		Deadline:  deadline.String(),
		Signature: hexutil.Encode(sig),
		Contract:  s.domain.VerifyingContract.Hex(),
		ChainID:   s.domain.ChainID.Int64(),
	}, nil
}

func (s *WithdrawerService) VerifyWithdrawal(ctx context.Context, txHash, walletAddr string, expectedAmount *big.Int) error {
	var receipt struct {
		Status          string `json:"status"`
		To              string `json:"to"`
		TransactionHash string `json:"transactionHash"`
		Logs            []struct {
			Address string   `json:"address"`
			Topics  []string `json:"topics"`
			Data    string   `json:"data"`
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

	contractAddr := s.domain.VerifyingContract.Hex()
	if !strings.EqualFold(receipt.To, contractAddr) {
		return fmt.Errorf("tx recipient is not withdrawer contract: got %s, want %s", receipt.To, contractAddr)
	}

	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("WithdrawalClaimed(address,uint256,uint256)"))
	claimEventTopic := "0x" + fmt.Sprintf("%x", h.Sum(nil))

	for _, log := range receipt.Logs {
		if !strings.EqualFold(log.Address, contractAddr) {
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
			return fmt.Errorf("withdrawal event wallet mismatch: got %s, want %s", eventWallet, walletAddr)
		}

		if len(log.Data) < 128 {
			return fmt.Errorf("withdrawal event data too short: %s", log.Data)
		}
		eventAmount := decodeUint256(log.Data)

		if eventAmount.Cmp(expectedAmount) != 0 {
			return fmt.Errorf("withdrawal event amount mismatch: got %s, want %s", eventAmount.String(), expectedAmount.String())
		}

		return nil
	}

	return fmt.Errorf("no valid WithdrawalClaimed event found for wallet %s in tx %s", walletAddr, txHash)
}

func (s *WithdrawerService) GetNonce(ctx context.Context, walletAddr string) (*big.Int, error) {
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

func (s *WithdrawerService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
