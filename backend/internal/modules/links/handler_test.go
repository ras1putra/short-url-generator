package links

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/links/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockURLServicer struct {
	mock.Mock
}

func (m *MockURLServicer) Create(ctx context.Context, userID uuid.UUID, req dto.CreateURLRequest) (*dto.URLResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.URLResponse), args.Error(1)
}

func (m *MockURLServicer) GetBySlug(ctx context.Context, slug string) (*repository.Url, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Url), args.Error(1)
}

func (m *MockURLServicer) GetByID(ctx context.Context, userID uuid.UUID, slug string) (*dto.URLResponse, error) {
	args := m.Called(ctx, userID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.URLResponse), args.Error(1)
}

func (m *MockURLServicer) ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int) (*dto.ListResponse, error) {
	args := m.Called(ctx, userID, page, perPage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ListResponse), args.Error(1)
}

func (m *MockURLServicer) Update(ctx context.Context, userID uuid.UUID, slug string, req dto.UpdateURLRequest) (*dto.URLResponse, error) {
	args := m.Called(ctx, userID, slug, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.URLResponse), args.Error(1)
}

func (m *MockURLServicer) GetStats(ctx context.Context, userID uuid.UUID, slug string) (*dto.StatsResponse, error) {
	args := m.Called(ctx, userID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StatsResponse), args.Error(1)
}

func (m *MockURLServicer) GetAggregateStats(ctx context.Context, userID uuid.UUID) (*dto.StatsResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StatsResponse), args.Error(1)
}

func (m *MockURLServicer) Delete(ctx context.Context, userID uuid.UUID, slug string) error {
	args := m.Called(ctx, userID, slug)
	return args.Error(0)
}

func setupLinksApp() (*fiber.App, *MockURLServicer) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockURLServicer)
	cfg := &config.Config{BaseURL: "http://localhost:8080", JWTAccessSecret: "test-secret"}
	handler := NewLinksHandler(mockSvc, cfg)

	userID := uuid.New()

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID.String())
		return c.Next()
	})
	app.Post("/api/links", handler.Create)
	app.Get("/api/links", handler.List)
	app.Get("/api/links/:slug", handler.Get)
	app.Get("/api/links/:slug/stats", handler.Stats)
	app.Get("/api/links/stats/aggregate", handler.AggregateStats)
	app.Patch("/api/links/:slug", handler.Update)
	app.Delete("/api/links/:slug", handler.Delete)
	app.Get("/api/links/:slug/qr", handler.QRCode)

	return app, mockSvc
}

func TestHandlerCreateSuccess(t *testing.T) {
	app, mockSvc := setupLinksApp()

	now := time.Now()
	respData := &dto.URLResponse{
		Slug:     "abc123",
		ShortURL: "http://localhost:8080/abc123",
		Original: "https://example.com",
		QRURL:    "http://localhost:8080/api/links/abc123/qr",
		CreatedAt: now,
	}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.MatchedBy(func(req dto.CreateURLRequest) bool {
		return req.URL == "https://example.com"
	})).Return(respData, nil)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestHandlerCreate_InvalidJSON(t *testing.T) {
	app, _ := setupLinksApp()

	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerListSuccess(t *testing.T) {
	app, mockSvc := setupLinksApp()

	listResp := &dto.ListResponse{
		Links:      []dto.URLResponse{},
		Total:      0,
		Page:       1,
		PerPage:    5,
		TotalPages: 0,
	}

	mockSvc.On("ListByUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), 1, 5).Return(listResp, nil)

	req := httptest.NewRequest("GET", "/api/links", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerGetSuccess(t *testing.T) {
	app, mockSvc := setupLinksApp()

	now := time.Now()
	respData := &dto.URLResponse{
		Slug:      "abc123",
		ShortURL:  "http://localhost:8080/abc123",
		Original:  "https://example.com",
		QRURL:     "http://localhost:8080/api/links/abc123/qr",
		CreatedAt: now,
	}

	mockSvc.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), "abc123").Return(respData, nil)

	req := httptest.NewRequest("GET", "/api/links/abc123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerDeleteSuccess(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID"), "abc123").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/links/abc123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerStatsSuccess(t *testing.T) {
	app, mockSvc := setupLinksApp()

	statsResp := &dto.StatsResponse{
		TotalClicks:  100,
		UniqueClicks: 80,
		ClicksPerDay: []dto.DateCount{},
		TopCountries: []dto.CountryCount{},
		Browsers:     map[string]int64{"Chrome": 60},
		Devices:       map[string]int64{"Mobile": 50},
	}

	mockSvc.On("GetStats", mock.Anything, mock.AnythingOfType("uuid.UUID"), "abc123").Return(statsResp, nil)

	req := httptest.NewRequest("GET", "/api/links/abc123/stats", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerQRCodeSuccess(t *testing.T) {
	app, mockSvc := setupLinksApp()

	testURL := &repository.Url{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Slug:      "abc123",
		Original:  "https://example.com",
		ExpiresAt: sql.NullTime{Valid: false},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockSvc.On("GetBySlug", mock.Anything, "abc123").Return(testURL, nil)

	req := httptest.NewRequest("GET", "/api/links/abc123/qr?size=128", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))

	mockSvc.AssertExpectations(t)
}

func TestHandlerCreate_NoUserID(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockURLServicer)
	handler := NewLinksHandler(mockSvc, &config.Config{BaseURL: "http://localhost:8080", JWTAccessSecret: "test-secret"})
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Post("/api/links", handler.Create)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerCreate_InvalidUserID(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockURLServicer)
	handler := NewLinksHandler(mockSvc, &config.Config{BaseURL: "http://localhost:8080", JWTAccessSecret: "test-secret"})
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "not-a-uuid")
		return c.Next()
	})
	app.Post("/api/links", handler.Create)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerCreate_ServiceError(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.Anything).Return(nil, response.NewAppError(500, "Internal server error"))

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerList_ServiceError(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("ListByUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.Anything, mock.Anything).Return(nil, response.NewAppError(500, "Internal server error"))

	req := httptest.NewRequest("GET", "/api/links", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerDelete_ServiceError(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID"), "del-me").Return(response.NewAppError(404, "URL not found"))

	req := httptest.NewRequest("DELETE", "/api/links/del-me", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerGet_ServiceError(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), "abc123").Return(nil, response.NewAppError(404, "URL not found"))

	req := httptest.NewRequest("GET", "/api/links/abc123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerUpdate_Success(t *testing.T) {
	app, mockSvc := setupLinksApp()

	now := time.Now()
	respData := &dto.URLResponse{
		Slug:      "newslug",
		ShortURL:  "http://localhost:8080/newslug",
		Original:  "https://example.com",
		QRURL:     "http://localhost:8080/api/links/newslug/qr",
		CreatedAt: now,
	}

	mockSvc.On("Update", mock.Anything, mock.AnythingOfType("uuid.UUID"), "abc123", mock.MatchedBy(func(req dto.UpdateURLRequest) bool {
		return req.CustomSlug == "newslug"
	})).Return(respData, nil)

	body, _ := json.Marshal(map[string]string{"custom_slug": "newslug"})
	req := httptest.NewRequest("PATCH", "/api/links/abc123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerUpdate_ServiceError(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("Update", mock.Anything, mock.AnythingOfType("uuid.UUID"), "abc123", mock.Anything).Return(nil, response.NewAppError(404, "URL not found"))

	body, _ := json.Marshal(map[string]string{"custom_slug": "newslug"})
	req := httptest.NewRequest("PATCH", "/api/links/abc123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerUpdate_InvalidJSON(t *testing.T) {
	app, _ := setupLinksApp()

	req := httptest.NewRequest("PATCH", "/api/links/abc123", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerStats_ServiceError(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("GetStats", mock.Anything, mock.AnythingOfType("uuid.UUID"), "abc123").Return(nil, response.NewAppError(404, "URL not found"))

	req := httptest.NewRequest("GET", "/api/links/abc123/stats", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerQRCode_SlugNotFound(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("GetBySlug", mock.Anything, "notfound").Return(nil, response.NewAppError(404, "URL not found"))

	req := httptest.NewRequest("GET", "/api/links/notfound/qr", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerCreate_ValidationError(t *testing.T) {
	app, _ := setupLinksApp()

	body, _ := json.Marshal(map[string]string{"url": ""})
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerList_NoUserID(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockURLServicer)
	handler := NewLinksHandler(mockSvc, &config.Config{BaseURL: "http://localhost:8080", JWTAccessSecret: "test-secret"})
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/api/links", handler.List)

	req := httptest.NewRequest("GET", "/api/links", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerList_PageBounds(t *testing.T) {
	app, mockSvc := setupLinksApp()

	listResp := &dto.ListResponse{
		Links:      []dto.URLResponse{},
		Total:      0,
		Page:       1,
		PerPage:    5,
		TotalPages: 0,
	}

	mockSvc.On("ListByUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), 1, 5).Return(listResp, nil)

	req := httptest.NewRequest("GET", "/api/links?page=-1&per_page=200", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockSvc.AssertExpectations(t)
}

func TestHandlerDelete_NoUserID(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockURLServicer)
	handler := NewLinksHandler(mockSvc, &config.Config{BaseURL: "http://localhost:8080", JWTAccessSecret: "test-secret"})
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Delete("/api/links/:slug", handler.Delete)

	req := httptest.NewRequest("DELETE", "/api/links/abc123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerGet_NoUserID(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockURLServicer)
	handler := NewLinksHandler(mockSvc, &config.Config{BaseURL: "http://localhost:8080", JWTAccessSecret: "test-secret"})
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/api/links/:slug", handler.Get)

	req := httptest.NewRequest("GET", "/api/links/abc123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerStats_NoUserID(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockURLServicer)
	handler := NewLinksHandler(mockSvc, &config.Config{BaseURL: "http://localhost:8080", JWTAccessSecret: "test-secret"})
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/api/links/:slug/stats", handler.Stats)

	req := httptest.NewRequest("GET", "/api/links/abc123/stats", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerAggregateStatsSuccess(t *testing.T) {
	app, mockSvc := setupLinksApp()

	statsResp := &dto.StatsResponse{
		TotalClicks:  500,
		UniqueClicks: 320,
		ClicksPerDay: []dto.DateCount{},
		TopCountries: []dto.CountryCount{
			{Country: "US", Count: 200},
			{Country: "ID", Count: 150},
		},
		Browsers: map[string]int64{"Chrome": 300},
		Devices:  map[string]int64{"Mobile": 200},
	}

	mockSvc.On("GetAggregateStats", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(statsResp, nil)

	req := httptest.NewRequest("GET", "/api/links/stats/aggregate", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerAggregateStats_ServiceError(t *testing.T) {
	app, mockSvc := setupLinksApp()

	mockSvc.On("GetAggregateStats", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, response.NewAppError(500, "Internal server error"))

	req := httptest.NewRequest("GET", "/api/links/stats/aggregate", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestHandlerAggregateStats_NoUserID(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	mockSvc := new(MockURLServicer)
	handler := NewLinksHandler(mockSvc, &config.Config{BaseURL: "http://localhost:8080", JWTAccessSecret: "test-secret"})
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/api/links/stats/aggregate", handler.AggregateStats)

	req := httptest.NewRequest("GET", "/api/links/stats/aggregate", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}