package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"urlshortener/pkg/constants"
	"urlshortener/pkg/token"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestJWTAuth_NoToken(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New()
	mockCache := new(MockCacher)
	app.Use(JWTAuth("secret", mockCache))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestJWTAuth_ValidToken(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	secret := "test-secret"

	userID := "user-123"
	tokenStr, err := token.IssueToken(userID, secret, "access", 15*time.Minute)
	require.NoError(t, err)

	app := fiber.New()
	mockCache := new(MockCacher)
	mockCache.On("Exists", mock.Anything, mock.MatchedBy(func(key string) bool {
		return len(key) > 3 && key[:3] == "bl:"
	})).Return(false, nil)

	app.Use(JWTAuth(secret, mockCache))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("user_id").(string))
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: tokenStr})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockCache.AssertExpectations(t)
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	app := fiber.New()
	mockCache := new(MockCacher)
	app.Use(JWTAuth("secret", mockCache))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: "invalid-token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestJWTAuth_RevokedToken(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	secret := "test-secret"

	tokenStr, err := token.IssueToken("user-123", secret, "access", 15*time.Minute)
	require.NoError(t, err)

	app := fiber.New()
	mockCache := new(MockCacher)
	mockCache.On("Exists", mock.Anything, mock.MatchedBy(func(key string) bool {
		return len(key) > 3 && key[:3] == "bl:"
	})).Return(true, nil)

	app.Use(JWTAuth(secret, mockCache))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: tokenStr})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	mockCache.AssertExpectations(t)
}

func TestJWTAuth_WrongTokenType(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	secret := "test-secret"

	refreshToken, err := token.IssueToken("user-123", secret, "refresh", 7*24*time.Hour)
	require.NoError(t, err)

	app := fiber.New()
	mockCache := new(MockCacher)
	app.Use(JWTAuth(secret, mockCache))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: refreshToken})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}