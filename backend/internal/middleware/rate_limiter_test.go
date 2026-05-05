package middleware

import (
	"net/http/httptest"
	"testing"

	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRateLimiter_UnderLimit(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	mockCache := new(MockCacher)
	mockCache.On("RateLimitIncrement", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(1, nil)

	app.Use(RateLimiter(mockCache, 60))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "59", resp.Header.Get("X-RateLimit-Remaining"))

	mockCache.AssertExpectations(t)
}

func TestRateLimiter_OverLimit(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	mockCache := new(MockCacher)
	mockCache.On("RateLimitIncrement", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(61, nil)

	app.Use(RateLimiter(mockCache, 60))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
	assert.Equal(t, "0", resp.Header.Get("X-RateLimit-Remaining"))

	mockCache.AssertExpectations(t)
}

func TestRateLimiter_RedisError(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	mockCache := new(MockCacher)
	mockCache.On("RateLimitIncrement", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(0, assert.AnError)

	app.Use(RateLimiter(mockCache, 60))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	mockCache.AssertExpectations(t)
}