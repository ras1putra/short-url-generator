package redirect

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"urlshortener/internal/analytics"
	"urlshortener/internal/repository"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockURLGetter struct {
	mock.Mock
}

func (m *MockURLGetter) GetBySlug(ctx context.Context, slug string) (*repository.Url, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Url), args.Error(1)
}

type MockClickSaverRedirect struct {
	mock.Mock
	saved chan struct{}
}

func (m *MockClickSaverRedirect) SaveClick(ctx context.Context, arg repository.SaveClickParams) (repository.Click, error) {
	args := m.Called(ctx, arg)
	if m.saved != nil {
		m.saved <- struct{}{}
	}
	return args.Get(0).(repository.Click), args.Error(1)
}

func TestRedirectHandler_Success(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	testURL := &repository.Url{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Slug:      "abc123",
		Original:  "https://example.com/long-url",
		Custom:    false,
		ExpiresAt: sql.NullTime{Valid: false},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockGetter := new(MockURLGetter)
	mockGetter.On("GetBySlug", mock.Anything, "abc123").Return(testURL, nil)

	mockSaver := new(MockClickSaverRedirect)
	mockSaver.saved = make(chan struct{}, 10)
	mockSaver.On("SaveClick", mock.Anything, mock.Anything).Return(repository.Click{}, nil)

	worker := analytics.NewAnalyticsWorker(mockSaver, 100)

	handler := NewRedirectHandler(mockGetter, worker, nil)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/:slug", handler.Redirect)

	req := httptest.NewRequest("GET", "/abc123", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)
	location := resp.Header.Get("Location")
	assert.Equal(t, "https://example.com/long-url", location)

	mockGetter.AssertExpectations(t)
}

func TestRedirectHandler_NotFound(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	mockGetter := new(MockURLGetter)
	mockGetter.On("GetBySlug", mock.Anything, "notfound").Return(nil, response.NewAppError(404, "URL not found"))

	mockSaver := new(MockClickSaverRedirect)
	worker := analytics.NewAnalyticsWorker(mockSaver, 100)

	handler := NewRedirectHandler(mockGetter, worker, nil)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/:slug", handler.Redirect)

	req := httptest.NewRequest("GET", "/notfound", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	mockGetter.AssertExpectations(t)
}

func TestRedirectHandler_MobileUserAgent(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	testURL := &repository.Url{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Slug:      "mob123",
		Original:  "https://example.com/mobile",
		Custom:    false,
		ExpiresAt: sql.NullTime{Valid: false},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockGetter := new(MockURLGetter)
	mockGetter.On("GetBySlug", mock.Anything, "mob123").Return(testURL, nil)

	mockSaver := new(MockClickSaverRedirect)
	mockSaver.saved = make(chan struct{}, 10)
	mockSaver.On("SaveClick", mock.Anything, mock.MatchedBy(func(arg repository.SaveClickParams) bool {
		return arg.Device.String == "mobile"
	})).Return(repository.Click{}, nil)

	worker := analytics.NewAnalyticsWorker(mockSaver, 100)
	handler := NewRedirectHandler(mockGetter, worker, nil)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/:slug", handler.Redirect)

	req := httptest.NewRequest("GET", "/mob123", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)

	select {
	case <-mockSaver.saved:
		mockSaver.AssertCalled(t, "SaveClick", mock.Anything, mock.MatchedBy(func(arg repository.SaveClickParams) bool {
			return arg.Device.String == "mobile"
		}))
	case <-time.After(3 * time.Second):
		t.Fatal("Click not processed within timeout")
	}
}

func TestRedirectHandler_BotUserAgent(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())

	testURL := &repository.Url{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Slug:      "bot123",
		Original:  "https://example.com/bot",
		Custom:    false,
		ExpiresAt: sql.NullTime{Valid: false},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockGetter := new(MockURLGetter)
	mockGetter.On("GetBySlug", mock.Anything, "bot123").Return(testURL, nil)

	mockSaver := new(MockClickSaverRedirect)
	mockSaver.saved = make(chan struct{}, 10)
	mockSaver.On("SaveClick", mock.Anything, mock.MatchedBy(func(arg repository.SaveClickParams) bool {
		return arg.Device.String == "bot"
	})).Return(repository.Click{}, nil)

	worker := analytics.NewAnalyticsWorker(mockSaver, 100)
	handler := NewRedirectHandler(mockGetter, worker, nil)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/:slug", handler.Redirect)

	req := httptest.NewRequest("GET", "/bot123", nil)
	req.Header.Set("User-Agent", "Googlebot/2.1 (+http://www.google.com/bot.html)")
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)

	select {
	case <-mockSaver.saved:
		mockSaver.AssertCalled(t, "SaveClick", mock.Anything, mock.MatchedBy(func(arg repository.SaveClickParams) bool {
			return arg.Device.String == "bot"
		}))
	case <-time.After(3 * time.Second):
		t.Fatal("Click not processed within timeout")
	}
}

func TestResolveGeo_EmptyIP(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	handler := &RedirectHandler{urlSvc: nil, worker: nil, geoDB: nil}
	country, city := handler.resolveGeo("")
	assert.Empty(t, country)
	assert.Empty(t, city)
}

func TestResolveGeo_NilGeoDB(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	handler := &RedirectHandler{urlSvc: nil, worker: nil, geoDB: nil}
	country, city := handler.resolveGeo("8.8.8.8")
	assert.Empty(t, country)
	assert.Empty(t, city)
}

func TestResolveGeo_InvalidIPWithGeoDB(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	db, err := geoip2.Open("../../../GeoLite2-City.mmdb")
	if err != nil {
		t.Skip("GeoIP database not available")
	}
	defer db.Close()

	handler := &RedirectHandler{urlSvc: nil, worker: nil, geoDB: db}
	country, city := handler.resolveGeo("invalid-ip")
	assert.Empty(t, country)
	assert.Empty(t, city)
}

func TestResolveGeo_ValidIPWithGeoDB(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	db, err := geoip2.Open("../../../GeoLite2-City.mmdb")
	if err != nil {
		t.Skip("GeoIP database not available")
	}
	defer db.Close()

	handler := &RedirectHandler{urlSvc: nil, worker: nil, geoDB: db}
	country, city := handler.resolveGeo("8.8.8.8")
	_ = country
	_ = city
}