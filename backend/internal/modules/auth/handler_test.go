package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth/dto"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockAuthServicer struct {
	mock.Mock
}

func (m *MockAuthServicer) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthServicer) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthServicer) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthServicer) Logout(ctx context.Context, accessToken, refreshToken string) error {
	args := m.Called(ctx, accessToken, refreshToken)
	return args.Error(0)
}

func setupAuthApp() (*fiber.App, *MockAuthServicer) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockAuthServicer)
	cfg := &config.Config{JWTAccessSecret: "test-secret", JWTRefreshSecret: "test-refresh-secret", Env: "development"}
	handler := NewAuthHandler(mockSvc, cfg)
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Post("/api/auth/register", handler.Register)
	app.Post("/api/auth/login", handler.Login)
	app.Post("/api/auth/refresh", handler.Refresh)
	app.Post("/api/auth/logout", handler.Logout)
	return app, mockSvc
}

func TestHandlerRegisterSuccess(t *testing.T) {
	app, mockSvc := setupAuthApp()

	userID := uuid.New()
	respData := &dto.AuthResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		User:         dto.UserResponse{ID: userID.String(), Email: "test@example.com", Name: "Test User", CreatedAt: time.Now()},
	}

	mockSvc.On("Register", mock.Anything, mock.MatchedBy(func(req dto.RegisterRequest) bool {
		return req.Email == "test@example.com"
	})).Return(respData, nil)

	body, _ := json.Marshal(dto.RegisterRequest{Name: "Test User", Email: "test@example.com", Password: "password123"})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerRegisterValidationError(t *testing.T) {
	app, _ := setupAuthApp()

	body, _ := json.Marshal(map[string]string{"name": "", "email": "bad-email", "password": "12"})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerLoginSuccess(t *testing.T) {
	app, mockSvc := setupAuthApp()

	userID := uuid.New()
	respData := &dto.AuthResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		User:         dto.UserResponse{ID: userID.String(), Email: "test@example.com", Name: "Test User", CreatedAt: time.Now()},
	}

	mockSvc.On("Login", mock.Anything, mock.MatchedBy(func(req dto.LoginRequest) bool {
		return req.Email == "test@example.com"
	})).Return(respData, nil)

	body, _ := json.Marshal(dto.LoginRequest{Email: "test@example.com", Password: "password123"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerRefreshMissingCookie(t *testing.T) {
	app, _ := setupAuthApp()

	req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerRefreshSuccess(t *testing.T) {
	app, mockSvc := setupAuthApp()

	userID := uuid.New()
	respData := &dto.AuthResponse{
		AccessToken: "new-access-token",
		User:        dto.UserResponse{ID: userID.String(), Email: "test@example.com", Name: "Test User", CreatedAt: time.Now()},
	}

	mockSvc.On("RefreshToken", mock.Anything, "refresh-token-value").Return(respData, nil)

	req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieRefreshToken, Value: "refresh-token-value"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerRefreshInvalidToken(t *testing.T) {
	app, mockSvc := setupAuthApp()

	mockSvc.On("RefreshToken", mock.Anything, "invalid-token").Return(nil, response.NewAppError(401, "Invalid refresh token"))

	req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieRefreshToken, Value: "invalid-token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerLogoutSuccess(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockAuthServicer)
	cfg := &config.Config{JWTAccessSecret: "test-secret", JWTRefreshSecret: "test-refresh-secret", Env: "development"}
	handler := NewAuthHandler(mockSvc, cfg)
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/api/auth/logout", handler.Logout)

	mockSvc.On("Logout", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: "some-access-token"})
	req.AddCookie(&http.Cookie{Name: constants.CookieRefreshToken, Value: "some-refresh-token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerRegister_InvalidJSON(t *testing.T) {
	app, _ := setupAuthApp()

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerRegister_ServiceError(t *testing.T) {
	app, mockSvc := setupAuthApp()

	mockSvc.On("Register", mock.Anything, mock.Anything).Return(nil, response.NewAppError(500, "Internal server error"))

	body, _ := json.Marshal(dto.RegisterRequest{Name: "Test User", Email: "test@example.com", Password: "password123"})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerLogin_InvalidJSON(t *testing.T) {
	app, _ := setupAuthApp()

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerLogin_ServiceError(t *testing.T) {
	app, mockSvc := setupAuthApp()

	mockSvc.On("Login", mock.Anything, mock.Anything).Return(nil, response.NewAppError(401, "Invalid credentials"))

	body, _ := json.Marshal(dto.LoginRequest{Email: "test@example.com", Password: "password123"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerLogout_Error(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockAuthServicer)
	cfg := &config.Config{JWTAccessSecret: "test-secret", JWTRefreshSecret: "test-refresh-secret", Env: "development"}
	handler := NewAuthHandler(mockSvc, cfg)
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return c.Next()
	})
	app.Post("/api/auth/logout", handler.Logout)

	mockSvc.On("Logout", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("revoke failed"))

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: constants.CookieAccessToken, Value: "some-access-token"})
	req.AddCookie(&http.Cookie{Name: constants.CookieRefreshToken, Value: "some-refresh-token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestSameSite_Production(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockAuthServicer)
	cfg := &config.Config{JWTAccessSecret: "test-secret", JWTRefreshSecret: "test-refresh-secret", Env: "production"}
	handler := NewAuthHandler(mockSvc, cfg)
	assert.Equal(t, "Strict", handler.sameSite())
}

func TestSameSite_Development(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockAuthServicer)
	cfg := &config.Config{JWTAccessSecret: "test-secret", JWTRefreshSecret: "test-refresh-secret", Env: "development"}
	handler := NewAuthHandler(mockSvc, cfg)
	assert.Equal(t, "Lax", handler.sameSite())
}