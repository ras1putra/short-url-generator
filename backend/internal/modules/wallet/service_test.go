package wallet

import (
	"context"
	"errors"
	"testing"
	"database/sql"

	"github.com/shopspring/decimal"

	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/response"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func appErrCode(err error) int {
	var appErr *response.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return -1
}

func TestService_GetWallet_Success(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	_, queries := testutil.SetupTestDB(t)
	svc := NewWalletService(queries, nil, nil, 0.005, "TST")
	ctx := context.Background()

	// Create user
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Wallet User",
		Email:    "wallet@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	// Create wallet with non-zero balance
	err = queries.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: "100.00",
	})
	require.NoError(t, err)

	// Create a transaction
	tx, err := queries.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID: user.ID,
		Amount: "10.00",
		Type:   "credit",
	})
	require.NoError(t, err)

	resp, err := svc.GetWallet(ctx, user.ID, 1, 10, "", "created_at", "desc")
	require.NoError(t, err)
	assert.True(t, decimal.NewFromFloat(100.00).Equal(resp.Balance))
	assert.Len(t, resp.Transactions, 1)
	assert.Equal(t, tx.ID.String(), resp.Transactions[0].ID)
	assert.True(t, decimal.NewFromFloat(10.00).Equal(resp.Transactions[0].Amount))
	assert.Equal(t, "credit", resp.Transactions[0].Type)
}

func TestService_GetWallet_NotFound(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	_, queries := testutil.SetupTestDB(t)
	svc := NewWalletService(queries, nil, nil, 0.005, "TST")
	ctx := context.Background()

	userID := uuid.New()
	resp, err := svc.GetWallet(ctx, userID, 1, 10, "", "created_at", "desc")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
}
