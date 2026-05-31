package wallet

import (
	"bytes"
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupWalletApp(t *testing.T) (*fiber.App, *repository.Queries) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	_, queries := testutil.SetupTestDB(t)
	svc := NewWalletService(queries, nil, nil, 0.005)
	handler := NewWalletHandler(svc)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		userIDStr := c.Get("X-Test-User-ID")
		if userIDStr != "" {
			c.Locals("user_id", userIDStr)
		}
		c.Locals("request_id", "test-req-id")
		return c.Next()
	})
	app.Get("/wallet", handler.GetWallet)
	app.Post("/wallet/pending", handler.CreatePendingTransaction)

	return app, queries
}

func TestGetWallet_Success(t *testing.T) {
	app, queries := setupWalletApp(t)
	ctx := context.Background()

	// Create user
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Wallet User",
		Email:    "wallet@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	// Create wallet
	err = queries.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: "100.50",
	})
	require.NoError(t, err)

	// Create transaction
	_, err = queries.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID: user.ID,
		Amount: "100.50",
		Type:   "credit",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/wallet", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}

func TestGetWallet_NotFound(t *testing.T) {
	app, _ := setupWalletApp(t)

	req := httptest.NewRequest("GET", "/wallet", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, r.StatusCode)
}

func TestGetWallet_Unauthorized(t *testing.T) {
	app, _ := setupWalletApp(t)

	req := httptest.NewRequest("GET", "/wallet", nil)

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, r.StatusCode)
}

func TestCreatePendingTransaction_Unauthorized(t *testing.T) {
	app, _ := setupWalletApp(t)

	req := httptest.NewRequest("POST", "/wallet/pending", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestCreatePendingTransaction_InvalidJSON(t *testing.T) {
	app, _ := setupWalletApp(t)

	req := httptest.NewRequest("POST", "/wallet/pending", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("X-Test-User-ID", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetWallet_WithPagination(t *testing.T) {
	app, queries := setupWalletApp(t)
	ctx := context.Background()

	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Paginated User",
		Email:    "paginated@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	err = queries.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: "50.00",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/wallet?page=1&per_page=10", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetWallet_SearchTransactions(t *testing.T) {
	app, queries := setupWalletApp(t)
	ctx := context.Background()

	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Search User",
		Email:    "search@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	err = queries.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: "100.00",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/wallet?q=DEPOSIT", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetWallet_SortByAmount(t *testing.T) {
	app, queries := setupWalletApp(t)
	ctx := context.Background()

	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Sort User",
		Email:    "sort@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	err = queries.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: "100.00",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/wallet?sort_by=amount&sort_dir=asc", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetWallet_InvalidUserID(t *testing.T) {
	app, _ := setupWalletApp(t)

	req := httptest.NewRequest("GET", "/wallet", nil)
	req.Header.Set("X-Test-User-ID", "not-a-uuid")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}
