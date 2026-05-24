package response

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupApp() *fiber.App {
	_ = zap.ReplaceGlobals(zap.NewNop())
	return fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
}

func TestNewAppError(t *testing.T) {
	err := NewAppError(404, "not found")
	assert.Equal(t, 404, err.Code)
	assert.Equal(t, "not found", err.Message)
	assert.Equal(t, "not found", err.Error())
}

func TestHandleError_AppError(t *testing.T) {
	app := setupApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return HandleError(c, NewAppError(409, "conflict"), "TestOp")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "conflict", body["message"])
	assert.Nil(t, body["data"])
}

func TestHandleError_GenericError(t *testing.T) {
	app := setupApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return HandleError(c, errors.New("something broke"), "TestOp")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "Internal server error", body["message"])
}

func TestErrorHandler_AppError(t *testing.T) {
	app := setupApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return NewAppError(400, "bad request")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestErrorHandler_FiberError(t *testing.T) {
	app := setupApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestErrorHandler_GenericError(t *testing.T) {
	app := setupApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		return errors.New("unknown error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestOK(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return OK(c, fiber.Map{"key": "value"}, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "success", body["message"])
}

func TestCreated(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return Created(c, fiber.Map{"id": 1}, "created")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestUnauthorized(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return Unauthorized(c, "no auth")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestForbidden(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return Forbidden(c, "forbidden")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestNotFound(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return NotFound(c, "not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestInternalError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return InternalError(c, "internal")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestHandleError_WithUserID(t *testing.T) {
	app := setupApp()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		return HandleError(c, errors.New("something broke"), "TestOp")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "Internal server error", body["message"])
}