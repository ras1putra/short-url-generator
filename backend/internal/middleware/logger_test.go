package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRequestID_WithHeader(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("request_id").(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "custom-id")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "custom-id", resp.Header.Get("X-Request-ID"))
}

func TestRequestID_WithoutHeader(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		id, ok := c.Locals("request_id").(string)
		assert.True(t, ok)
		assert.NotEmpty(t, id)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"))
}

func TestRequestLogger_Success(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestRequestLogger_WithError(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger())
	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestRequestLogger_WithUserID(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New()
	app.Use(RequestID())
	app.Use(RequestLogger())
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}