package helper

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestParseDecimal(t *testing.T) {
	assert.True(t, decimal.NewFromFloat(1.5).Equal(ParseDecimal("1.50")))
	assert.True(t, decimal.NewFromFloat(0.0).Equal(ParseDecimal("0.00")))
	assert.True(t, decimal.Zero.Equal(ParseDecimal("invalid")))
	assert.True(t, decimal.NewFromFloat(100.99).Equal(ParseDecimal("100.99")))
	assert.True(t, decimal.Zero.Equal(ParseDecimal("")))
}

func TestFormatDecimal(t *testing.T) {
	assert.Equal(t, "1.50000000", FormatDecimal(decimal.NewFromFloat(1.5)))
	assert.Equal(t, "0.00000000", FormatDecimal(decimal.Zero))
	assert.Equal(t, "100.99000000", FormatDecimal(decimal.NewFromFloat(100.99)))
	assert.Equal(t, "2.50000000", FormatDecimal(decimal.NewFromFloat(2.5)))
}

func TestUserIDFromCtx_Success(t *testing.T) {
	id := uuid.New()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", id.String())
		parsed, err := UserIDFromCtx(c)
		assert.NoError(t, err)
		assert.Equal(t, id, parsed)
		return nil
	})
	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUserIDFromCtx_Missing(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		_, err := UserIDFromCtx(c)
		assert.Error(t, err)
		return nil
	})
	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUserIDFromCtx_InvalidUUID(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", "not-a-uuid")
		_, err := UserIDFromCtx(c)
		assert.Error(t, err)
		return nil
	})
	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestParseAndValidate_Valid(t *testing.T) {
	v := validator.New()
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		type testReq struct {
			Name string `json:"name" validate:"required"`
		}
		var req testReq
		err := ParseAndValidate(c, v, &req)
		assert.NoError(t, err)
		assert.Equal(t, "test", req.Name)
		return nil
	})

	body := bytes.NewReader([]byte(`{"name":"test"}`))
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestParseAndValidate_InvalidJSON(t *testing.T) {
	v := validator.New()
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		type testReq struct {
			Name string `json:"name"`
		}
		var req testReq
		err := ParseAndValidate(c, v, &req)
		assert.Error(t, err)
		return nil
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestParseAndValidate_ValidationError(t *testing.T) {
	v := validator.New()
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		type testReq struct {
			Name string `json:"name" validate:"required"`
		}
		var req testReq
		err := ParseAndValidate(c, v, &req)
		assert.Error(t, err)
		return nil
	})

	body := bytes.NewReader([]byte(`{"name":""}`))
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}
