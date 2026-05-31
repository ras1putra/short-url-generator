package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"database/sql"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/mailer"
	"urlshortener/internal/modules/auth/dto"
	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/response"
	"urlshortener/pkg/token"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testCfg = &config.Config{
	BaseURL:          "http://localhost:8080",
	JWTAccessSecret:  "test-secret",
	JWTRefreshSecret: "test-refresh-secret",
	Env:              "development",
}

func setupAuthApp(t *testing.T) (*fiber.App, *repository.Queries) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	db, queries := testutil.SetupTestDB(t)
	fakeCache := testutil.NewFakeCacher()
	svc := NewAuthService(db, queries, fakeCache, testCfg, mailer.New("", "", ""))
	handler := NewAuthHandler(svc, testCfg)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Post("/api/auth/register", handler.Register)
	app.Post("/api/auth/login", handler.Login)
	app.Post("/api/auth/refresh", handler.Refresh)
	app.Post("/api/auth/logout", handler.Logout)
	app.Post("/api/auth/send-verification", handler.SendVerification)
	app.Post("/api/auth/verify-email", handler.VerifyEmail)
	app.Post("/api/auth/forgot-password", handler.ForgotPassword)
	app.Post("/api/auth/reset-password", handler.ResetPassword)
	app.Get("/api/auth/me", handler.Me)

	app.Post("/upgrade", func(c *fiber.Ctx) error {
		// Mock auth middleware locals
		userIDStr := c.Get("X-Test-User-ID")
		if userIDStr == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		c.Locals("user_id", userIDStr)
		c.Locals("role", c.Get("X-Test-Role", "user"))
		c.Locals("request_id", "test-req-id")
		return handler.UpgradeToAdvertiser(c)
	})

	return app, queries
}

func TestHandlerRegisterSuccess(t *testing.T) {
	app, queries := setupAuthApp(t)

	body, _ := json.Marshal(dto.RegisterRequest{
		Name:     "Test User",
		Email:    "handler-register@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	// Verify in DB
	dbUser, err := queries.GetUserByEmail(context.Background(), "handler-register@example.com")
	require.NoError(t, err)
	assert.Equal(t, "Test User", dbUser.Name)
}

func TestHandlerRegisterValidationError(t *testing.T) {
	app, _ := setupAuthApp(t)

	body, _ := json.Marshal(map[string]string{"name": "", "email": "bad-email", "password": "12"})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerLoginSuccess(t *testing.T) {
	app, queries := setupAuthApp(t)
	ctx := context.Background()

	// Pre-create user with a hashed password
	hashedPassword := hashPassword("password123")

	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Login User",
		Email:    "handler-login@example.com",
		Password: sql.NullString{String: hashedPassword, Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	_, err = queries.UpdateUserEmailVerified(ctx, user.ID)
	require.NoError(t, err)

	body, _ := json.Marshal(dto.LoginRequest{Email: "handler-login@example.com", Password: "password123"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerRefreshMissingCookie(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerRefreshSuccess(t *testing.T) {
	app, queries := setupAuthApp(t)
	ctx := context.Background()

	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Refresh User",
		Email:    "handler-refresh@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	refreshToken, err := token.IssueToken(user.ID.String(), user.Role, testCfg.JWTRefreshSecret, "refresh", 1*time.Hour)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieRefreshToken, Value: refreshToken})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerLogoutSuccess(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: "some-access-token"})
	req.AddCookie(&http.Cookie{Name: constants.CookieRefreshToken, Value: "some-refresh-token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerRegister_InvalidJSON(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerLogin_InvalidJSON(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestSameSite_Production(t *testing.T) {
	cfg := &config.Config{Env: "production"}
	h := NewCookieHelper(cfg)
	assert.Equal(t, "Strict", h.sameSite())
}

func TestSameSite_Development(t *testing.T) {
	cfg := &config.Config{Env: "development"}
	h := NewCookieHelper(cfg)
	assert.Equal(t, "Lax", h.sameSite())
}

func TestHandlerUpgradeToAdvertiser_Success(t *testing.T) {
	app, queries := setupAuthApp(t)
	ctx := context.Background()

	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Upgrade User",
		Email:    "handler-upgrade@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/upgrade", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	req.Header.Set("X-Test-Role", "user")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify in DB
	dbUser, err := queries.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "advertiser", dbUser.Role)
}

func TestHandlerUpgradeToAdvertiser_Unauthorized(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("POST", "/upgrade", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerSendVerification_Success(t *testing.T) {
	app, queries := setupAuthApp(t)
	ctx := context.Background()

	hashedPassword := hashPassword("password123")
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Verify User",
		Email:    "verify-handler@example.com",
		Password: sql.NullString{String: hashedPassword, Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	_, err = queries.UpdateUserEmailVerified(ctx, user.ID)
	require.NoError(t, err)

	body, _ := json.Marshal(dto.SendVerificationRequest{Email: "verify-handler@example.com"})
	req := httptest.NewRequest("POST", "/api/auth/send-verification", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerSendVerification_InvalidEmail(t *testing.T) {
	app, _ := setupAuthApp(t)

	body, _ := json.Marshal(dto.SendVerificationRequest{Email: "not-an-email"})
	req := httptest.NewRequest("POST", "/api/auth/send-verification", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerVerifyEmail_MissingToken(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("POST", "/api/auth/verify-email", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerVerifyEmail_InvalidToken(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("POST", "/api/auth/verify-email?token=invalid-token", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerForgotPassword_Success(t *testing.T) {
	app, _ := setupAuthApp(t)

	body, _ := json.Marshal(dto.ForgotPasswordRequest{Email: "nonexistent-forgot@example.com"})
	req := httptest.NewRequest("POST", "/api/auth/forgot-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerForgotPassword_NonExistentEmail(t *testing.T) {
	app, _ := setupAuthApp(t)

	body, _ := json.Marshal(dto.ForgotPasswordRequest{Email: "nonexistent@example.com"})
	req := httptest.NewRequest("POST", "/api/auth/forgot-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerResetPassword_InvalidToken(t *testing.T) {
	app, _ := setupAuthApp(t)

	body, _ := json.Marshal(dto.ResetPasswordRequest{Token: "bad-token", Password: "newpassword123"})
	req := httptest.NewRequest("POST", "/api/auth/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerResetPassword_MissingToken(t *testing.T) {
	app, _ := setupAuthApp(t)

	body, _ := json.Marshal(dto.ResetPasswordRequest{Token: "", Password: "newpassword123"})
	req := httptest.NewRequest("POST", "/api/auth/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerResetPassword_ShortPassword(t *testing.T) {
	app, _ := setupAuthApp(t)

	body, _ := json.Marshal(dto.ResetPasswordRequest{Token: "sometoken", Password: "123"})
	req := httptest.NewRequest("POST", "/api/auth/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerMe_Unauthorized(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerDowngradeToUser_Unauthorized(t *testing.T) {
	app, _ := setupAuthApp(t)

	req := httptest.NewRequest("POST", "/upgrade", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerRegister_DuplicateEmail(t *testing.T) {
	app, queries := setupAuthApp(t)
	ctx := context.Background()

	hashedPassword := hashPassword("password123")
	_, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Existing",
		Email:    "duplicate@example.com",
		Password: sql.NullString{String: hashedPassword, Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)

	body, _ := json.Marshal(dto.RegisterRequest{
		Name:     "New User",
		Email:    "duplicate@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestHandlerLogin_EmailNotVerified(t *testing.T) {
	app, _ := setupAuthApp(t)

	body, _ := json.Marshal(dto.LoginRequest{Email: "nonexistent@example.com", Password: "password123"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}
