package wallet

import (
	"context"
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
	svc := NewWalletService(queries, nil, nil)
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

	return app, queries
}

func TestGetWallet_Success(t *testing.T) {
	app, queries := setupWalletApp(t)
	ctx := context.Background()

	// Create user
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Wallet User",
		Email:    "wallet@example.com",
		Password: "password",
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
