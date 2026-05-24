package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"urlshortener/internal/testutil"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRateLimiter_UnderLimit(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	redisCache := testutil.SetupTestRedis(t)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(RateLimiter(redisCache, 60))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "59", resp.Header.Get("X-RateLimit-Remaining"))
}

func TestRateLimiter_OverLimit(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	redisCache := testutil.SetupTestRedis(t)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(RateLimiter(redisCache, 5))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
	assert.Equal(t, "0", resp.Header.Get("X-RateLimit-Remaining"))
}

type errorCacher struct{}

func (e *errorCacher) Get(ctx context.Context, key string) (string, error) {
	return "", assert.AnError
}

func (e *errorCacher) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return assert.AnError
}

func (e *errorCacher) Del(ctx context.Context, key string) error {
	return assert.AnError
}

func (e *errorCacher) Exists(ctx context.Context, key string) (bool, error) {
	return false, assert.AnError
}

func (e *errorCacher) RateLimitIncrement(ctx context.Context, key string, ttl time.Duration) (int, error) {
	return 0, assert.AnError
}

func TestRateLimiter_RedisError(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(RateLimiter(&errorCacher{}, 60))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}
