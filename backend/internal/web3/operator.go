package web3

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"

	"urlshortener/pkg/logger"
)

type OperatorService struct {
	operatorKey  *ecdsa.PrivateKey
	operatorAddr common.Address
	client       *ethclient.Client
	chainID      *big.Int
	tokenAddr    common.Address
	mu           sync.Mutex
}

func NewOperatorService(keyHex, tokenAddrStr, rpcURL string, chainID *big.Int) (*OperatorService, error) {
	keyHex = strings.TrimPrefix(keyHex, "0x")
	priv, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse operator private key: %w", err)
	}

	operatorAddr := crypto.PubkeyToAddress(priv.PublicKey)
	tokenAddr := common.HexToAddress(tokenAddrStr)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC for operator service: %w", err)
	}

	logger.Ctx(context.Background()).Info("Operator service initialized",
		zap.String("address", operatorAddr.Hex()),
		zap.String("token_addr", tokenAddr.Hex()),
		zap.String("chain_id", chainID.String()),
	)

	return &OperatorService{
		operatorKey:  priv,
		operatorAddr: operatorAddr,
		client:       client,
		chainID:      chainID,
		tokenAddr:    tokenAddr,
	}, nil
}

func (s *OperatorService) OperatorAddress() common.Address {
	return s.operatorAddr
}

func (s *OperatorService) TokenBalance(ctx context.Context) (*big.Int, error) {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("balanceOf(address)"))
	selector := h.Sum(nil)[:4]

	data := make([]byte, 36)
	copy(data[:4], selector)
	copy(data[4:36], common.LeftPadBytes(s.operatorAddr.Bytes(), 32))

	msg := ethereum.CallMsg{
		To:   &s.tokenAddr,
		Data: data,
	}

	result, err := s.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(result), nil
}

func (s *OperatorService) EthBalance(ctx context.Context) (*big.Int, error) {
	return s.client.BalanceAt(ctx, s.operatorAddr, nil)
}

func (s *OperatorService) SendERC20(ctx context.Context, to common.Address, amount *big.Int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("transfer(address,uint256)"))
	methodID := h.Sum(nil)[:4]

	data := make([]byte, 68)
	copy(data[:4], methodID)
	copy(data[4:36], common.LeftPadBytes(to.Bytes(), 32))
	copy(data[36:68], common.LeftPadBytes(amount.Bytes(), 32))

	nonce, err := s.client.PendingNonceAt(ctx, s.operatorAddr)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	gasLimit, err := s.client.EstimateGas(ctx, ethereum.CallMsg{
		From: s.operatorAddr,
		To:   &s.tokenAddr,
		Data: data,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas: %w", err)
	}

	tip, err := s.client.SuggestGasTipCap(ctx)
	if err != nil {
		tip = big.NewInt(0)
	}

	var tx *types.Transaction
	var maxFee *big.Int

	feeHistory, err := s.client.FeeHistory(ctx, 1, nil, nil)
	if err == nil && len(feeHistory.BaseFee) > 0 {
		baseFee := feeHistory.BaseFee[0]
		maxFee = new(big.Int).Mul(baseFee, big.NewInt(2))
		maxFee = new(big.Int).Add(maxFee, tip)

		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   s.chainID,
			Nonce:     nonce,
			GasTipCap: tip,
			GasFeeCap: maxFee,
			Gas:       gasLimit,
			To:        &s.tokenAddr,
			Value:     big.NewInt(0),
			Data:      data,
		})
	} else {
		gasPrice, err := s.client.SuggestGasPrice(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to suggest gas price: %w", err)
		}
		gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(120))
		gasPrice = new(big.Int).Div(gasPrice, big.NewInt(100))
		maxFee = gasPrice

		tx = types.NewTransaction(nonce, s.tokenAddr, big.NewInt(0), gasLimit, gasPrice, data)
	}

	balance, err := s.client.BalanceAt(ctx, s.operatorAddr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to check operator ETH balance: %w", err)
	}
	needed := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), maxFee)
	if balance.Cmp(needed) < 0 {
		return "", fmt.Errorf("operator ETH balance too low: have %s wei, need %s wei", balance.String(), needed.String())
	}

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(s.chainID), s.operatorKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	if err := s.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()

	for i := 0; i < 120; i++ {
		receipt, err := s.client.TransactionReceipt(ctx, signedTx.Hash())
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if receipt.Status == 0 {
			return txHash, fmt.Errorf("transaction reverted on-chain: %s", txHash)
		}
		return txHash, nil
	}

	return txHash, fmt.Errorf("timed out waiting for receipt: %s", txHash)
}

func (s *OperatorService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
