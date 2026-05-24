package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"urlshortener/internal/repository"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/assert"
)

func TestMapWalletToResponse(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	wallet := repository.Wallet{
		UserID:    userID,
		Balance:   "150.50",
		UpdatedAt: now,
	}

	resp := MapWalletToResponse(wallet)

	assert.True(t, decimal.NewFromFloat(150.50).Equal(resp.Balance))
	assert.Equal(t, now.Format(time.RFC3339), resp.UpdatedAt)
}

func TestMapWalletToResponse_ZeroBalance(t *testing.T) {
	wallet := repository.Wallet{
		UserID:    uuid.New(),
		Balance:   "0.00",
		UpdatedAt: time.Now(),
	}

	resp := MapWalletToResponse(wallet)
	assert.True(t, decimal.Zero.Equal(resp.Balance))
}

func TestMapTransactionToResponse(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	txID := uuid.New()

	tx := repository.Transaction{
		ID:        txID,
		UserID:    userID,
		Amount:    "99.99",
		Type:      "DEPOSIT",
		Metadata:  pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"deposit_amount":99.99}`), Valid: true},
		CreatedAt: now,
	}

	resp := MapTransactionToResponse(tx)

	assert.Equal(t, txID.String(), resp.ID)
	assert.Equal(t, userID.String(), resp.UserID)
	assert.True(t, decimal.NewFromFloat(99.99).Equal(resp.Amount))
	assert.Equal(t, "DEPOSIT", resp.Type)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
	assert.NotNil(t, resp.Metadata)
}

func TestMapTransactionToResponse_NoMetadata(t *testing.T) {
	tx := repository.Transaction{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Amount:    "10.00",
		Type:      "EARNING",
		CreatedAt: time.Now(),
	}

	resp := MapTransactionToResponse(tx)
	assert.Nil(t, resp.Metadata)
}

func TestMapTransactionsToResponse(t *testing.T) {
	txs := []repository.Transaction{
		{ID: uuid.New(), UserID: uuid.New(), Amount: "10.00", Type: "DEPOSIT", CreatedAt: time.Now()},
		{ID: uuid.New(), UserID: uuid.New(), Amount: "5.00", Type: "EARNING", CreatedAt: time.Now()},
	}

	resp := MapTransactionsToResponse(txs)
	assert.Len(t, resp, 2)
	assert.Equal(t, "DEPOSIT", resp[0].Type)
	assert.Equal(t, "EARNING", resp[1].Type)
}

func TestMapTransactionsToResponse_Empty(t *testing.T) {
	resp := MapTransactionsToResponse([]repository.Transaction{})
	assert.Len(t, resp, 0)
}
