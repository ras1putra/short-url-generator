package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"

	"urlshortener/internal/repository"
	"urlshortener/pkg/helper"
)

type WalletResponse struct {
	Balance   decimal.Decimal `json:"balance"`
	UpdatedAt string          `json:"updated_at"`
}

type TransactionResponse struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Amount    decimal.Decimal `json:"amount"`
	Type      string          `json:"type"`
	Status    string          `json:"status"`
	TxHash    *string         `json:"tx_hash,omitempty"`
	Metadata  interface{}     `json:"metadata,omitempty"`
	CreatedAt string          `json:"created_at"`
}

type WithdrawRequest struct {
	Amount     decimal.Decimal `json:"amount" validate:"required,decimal_gt=0"`
	WalletAddr string          `json:"wallet_addr" validate:"required,eth_addr"`
}

type CreatePendingTransactionRequest struct {
	Amount     decimal.Decimal `json:"amount" validate:"required,decimal_gt=0"`
	Type       string          `json:"type" validate:"required,oneof=DEPOSIT WITHDRAWAL APPROVE"`
	TxHash     string          `json:"tx_hash" validate:"required,eth_tx_hash"`
	WalletAddr string          `json:"wallet_addr,omitempty" validate:"omitempty,eth_addr"`
	RequestID  string          `json:"request_id,omitempty"`
}

type WithdrawalPermitResponse struct {
	RequestID string `json:"request_id"`
	Wallet    string `json:"wallet"`
	Amount    string `json:"amount"`
	Nonce     string `json:"nonce"`
	Deadline  string `json:"deadline"`
	Signature string `json:"signature"`
	Contract  string `json:"contract"`
	ChainID   int64  `json:"chain_id"`
}

type WalletWithTransactionsResponse struct {
	Balance      decimal.Decimal       `json:"balance"`
	Available    decimal.Decimal       `json:"available"`
	Frozen       decimal.Decimal       `json:"frozen"`
	UpdatedAt    string                `json:"updated_at"`
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	PerPage      int                   `json:"per_page"`
	TotalPages   int                   `json:"total_pages"`
}

func MapWalletToResponse(wallet repository.Wallet) WalletResponse {
	return WalletResponse{
		Balance:   helper.ParseDecimal(wallet.Balance),
		UpdatedAt: wallet.UpdatedAt.Format(time.RFC3339),
	}
}

func MapTransactionToResponse(tx repository.Transaction) TransactionResponse {
	resp := TransactionResponse{
		ID:        tx.ID.String(),
		UserID:    tx.UserID.String(),
		Amount:    helper.ParseDecimal(tx.Amount),
		Type:      tx.Type,
		Status:    tx.Status,
		CreatedAt: tx.CreatedAt.Format(time.RFC3339),
	}
	if tx.TxHash.Valid {
		resp.TxHash = &tx.TxHash.String
	}
	if tx.Metadata.Valid {
		var meta interface{}
		if err := json.Unmarshal(tx.Metadata.RawMessage, &meta); err == nil {
			resp.Metadata = meta
		}
	}
	return resp
}

func MapTransactionsToResponse(txs []repository.Transaction) []TransactionResponse {
	resp := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		resp[i] = MapTransactionToResponse(tx)
	}
	return resp
}
