package web3

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"

	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
)

const (
	depositEventSig = "Deposit(address,bytes32,uint256)"
)

var depositEventTopic = func() string {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(depositEventSig))
	return "0x" + fmt.Sprintf("%x", h.Sum(nil))
}()

type DepositEvent struct {
	User   string   `json:"user"`
	RefID  string   `json:"refId"`
	Amount *big.Int `json:"amount"`
	TxHash string   `json:"txHash"`
	Block  uint64   `json:"block"`
}

type DepositHandler interface {
	HandleDeposit(ctx context.Context, event *DepositEvent) error
}

type Listener struct {
	client   *ETHClient
	handler  DepositHandler
	contract string
	done     chan struct{}
	isDev    bool
}

func NewListener(client *ETHClient, contractAddr string, handler DepositHandler, isDev bool) *Listener {
	return &Listener{
		client:   client,
		handler:  handler,
		contract: contractAddr,
		done:     make(chan struct{}),
		isDev:    isDev,
	}
}

func (l *Listener) Start(ctx context.Context) {
	logger.Ctx(ctx).Info("Deposit listener started",
		zap.String("contract", l.contract),
		zap.Duration("interval", constants.PollInterval),
	)

	ticker := time.NewTicker(constants.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.poll(ctx)
		case <-ctx.Done():
			logger.Ctx(ctx).Info("Deposit listener stopped")
			return
		case <-l.done:
			return
		}
	}
}

func (l *Listener) Stop() {
	close(l.done)
}

func (l *Listener) poll(ctx context.Context) {
	lastBlock, err := l.client.GetLastProcessedBlock(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get last processed block", zap.Error(err))
		return
	}

	currentBlock, err := l.getCurrentBlock(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get current block", zap.Error(err))
		return
	}

	if lastBlock == 0 {
		// Dev mode
		if l.isDev {
			lastBlock = 1
		} else {
			lastBlock = currentBlock
		}
		// Write to Redis immediately to break 0-value loop
		_ = l.client.SetLastProcessedBlock(ctx, lastBlock)
	}

	confirmations := uint64(constants.BlockConfirmations)
	if l.isDev {
		confirmations = 0
	}

	if currentBlock <= lastBlock+confirmations {
		return
	}

	// Only process events with sufficient confirmations
	safeBlock := currentBlock - confirmations
	if safeBlock <= lastBlock {
		return
	}

	events, err := l.fetchDepositEvents(ctx, lastBlock+1, safeBlock)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to fetch deposit events", zap.Error(err))
		return
	}

	processedTo := lastBlock
	for _, event := range events {
		logger.Ctx(ctx).Info("Processing deposit event",
			zap.String("user", event.User),
			zap.String("refId", event.RefID),
			zap.String("amount", event.Amount.String()),
			zap.String("tx", event.TxHash),
		)

		if err := l.handler.HandleDeposit(ctx, &event); err != nil {
			logger.Ctx(ctx).Error("Failed to handle deposit event",
				zap.String("refId", event.RefID),
				zap.Error(err),
			)
			break
		}

		if event.Block > processedTo {
			processedTo = event.Block
		}
	}

	if processedTo > lastBlock {
		if err := l.client.SetLastProcessedBlock(ctx, processedTo); err != nil {
			logger.Ctx(ctx).Error("Failed to update last processed block", zap.Error(err))
		}
	}
}

func (l *Listener) getCurrentBlock(ctx context.Context) (uint64, error) {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}

	var result string
	if err := rpcCall(ctx, l.client.RPCURL(), body, &result); err != nil {
		return 0, err
	}

	if len(result) < 3 {
		return 0, fmt.Errorf("invalid block number response: %s", result)
	}

	block := new(big.Int)
	block.SetString(result[2:], 16)
	return block.Uint64(), nil
}

func (l *Listener) fetchDepositEvents(ctx context.Context, fromBlock, toBlock uint64) ([]DepositEvent, error) {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getLogs",
		"params": []interface{}{
			map[string]interface{}{
				"address":   l.contract,
				"fromBlock": fmt.Sprintf("0x%x", fromBlock),
				"toBlock":   fmt.Sprintf("0x%x", toBlock),
				"topics":    []string{depositEventTopic},
			},
		},
		"id": 1,
	}

	var raw json.RawMessage
	if err := rpcCall(ctx, l.client.RPCURL(), body, &raw); err != nil {
		return nil, err
	}

	var logs []struct {
		Address         string   `json:"address"`
		Topics          []string `json:"topics"`
		Data            string   `json:"data"`
		BlockNumber     string   `json:"blockNumber"`
		TransactionHash string   `json:"transactionHash"`
	}

	if err := json.Unmarshal(raw, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}

	var events []DepositEvent
	for _, log := range logs {
		if len(log.Topics) < 3 {
			continue
		}

		events = append(events, DepositEvent{
			User:   decodeAddress(log.Topics[1]),
			RefID:  decodeBytes32(log.Topics[2]),
			Amount: decodeUint256(log.Data),
			TxHash: log.TransactionHash,
			Block:  parseHexUint64(log.BlockNumber),
		})
	}

	return events, nil
}
