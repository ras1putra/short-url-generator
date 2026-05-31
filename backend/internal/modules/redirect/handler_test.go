package redirect

import (
	"context"
	"net/http/httptest"
	"testing"
	"database/sql"
	"time"

	"urlshortener/internal/analytics"
	"urlshortener/internal/config"
	"urlshortener/internal/modules/links"
	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testCfg = &config.Config{
	BaseURL:          "http://localhost:8080",
	JWTAccessSecret:  "test-secret",
	JWTRefreshSecret: "test-refresh-secret",
}

func setupRedirectApp(t *testing.T) (*fiber.App, *repository.Queries, *analytics.AnalyticsWorker) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	db, queries := testutil.SetupTestDB(t)
	fakeCache := testutil.NewFakeCacher()
	urlSvc := links.NewURLService(queries, fakeCache, testCfg)
	worker := analytics.NewAnalyticsWorker(queries, 100)

	redirectSvc := NewRedirectService(urlSvc, queries, worker, nil, testCfg, db, nil)
	handler := NewRedirectHandler(redirectSvc)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/:slug", handler.Redirect)
	app.Get("/:slug/click/:adID", handler.AdClick)
	app.Get("/:slug/complete/:adID", handler.AdComplete)
	app.Post("/:slug/complete", handler.AdCompleteFlow)
	app.Get("/:slug/skip/:adID", handler.AdSkip)

	return app, queries, worker
}

func createTestUser(t *testing.T, queries *repository.Queries, ctx context.Context) repository.User {
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Redirect User",
		Email:    "redirect@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)
	return user
}

func TestRedirectHandler_Success(t *testing.T) {
	app, queries, _ := setupRedirectApp(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)

	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:   user.ID,
		Slug:     "abc123",
		Original: "https://example.com/long-url",
		Custom:   false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/abc123", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)
	location := resp.Header.Get("Location")
	assert.Equal(t, "https://example.com/long-url", location)

	// Wait briefly for background worker to persist click
	time.Sleep(100 * time.Millisecond)

	count, err := queries.GetTotalClicksBySlug(ctx, "abc123")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedirectHandler_NotFound(t *testing.T) {
	app, _, _ := setupRedirectApp(t)

	req := httptest.NewRequest("GET", "/notfound", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestRedirectHandler_MobileUserAgent(t *testing.T) {
	app, queries, _ := setupRedirectApp(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)

	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:   user.ID,
		Slug:     "mob123",
		Original: "https://example.com/mobile",
		Custom:   false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/mob123", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)

	// Wait for worker
	time.Sleep(100 * time.Millisecond)

	clicks, err := queries.GetStatsBySlug(ctx, "mob123")
	require.NoError(t, err)
	require.Len(t, clicks, 1)
	assert.Equal(t, "mobile", clicks[0].Device.String)
}

func TestRedirectHandler_BotUserAgent(t *testing.T) {
	app, queries, _ := setupRedirectApp(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)

	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:   user.ID,
		Slug:     "bot123",
		Original: "https://example.com/bot",
		Custom:   false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/bot123", nil)
	req.Header.Set("User-Agent", "Googlebot/2.1 (+http://www.google.com/bot.html)")
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)

	// Wait for worker
	time.Sleep(100 * time.Millisecond)

	clicks, err := queries.GetStatsBySlug(ctx, "bot123")
	require.NoError(t, err)
	require.Len(t, clicks, 1)
	assert.Equal(t, "bot", clicks[0].Device.String)
}

func TestResolveGeo_EmptyIP(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	service := &RedirectService{urlSvc: nil, worker: nil, geoDB: nil}
	country, city := service.resolveGeo("")
	assert.Empty(t, country)
	assert.Empty(t, city)
}

func TestResolveGeo_NilGeoDB(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	service := &RedirectService{urlSvc: nil, worker: nil, geoDB: nil}
	country, city := service.resolveGeo("8.8.8.8")
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

	service := &RedirectService{urlSvc: nil, worker: nil, geoDB: db}
	country, city := service.resolveGeo("invalid-ip")
	assert.Empty(t, country)
	assert.Empty(t, city)
}

func TestRedirectHandler_AdClick_InvalidAdID(t *testing.T) {
	app, queries, _ := setupRedirectApp(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)

	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:   user.ID,
		Slug:     "click-test",
		Original: "https://example.com/dest",
		Custom:   false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/click-test/click/invalid-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)
}

func TestRedirectHandler_AdComplete_InvalidAdID(t *testing.T) {
	app, queries, _ := setupRedirectApp(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)

	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:   user.ID,
		Slug:     "complete-test",
		Original: "https://example.com/dest",
		Custom:   false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/complete-test/complete/invalid-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestRedirectHandler_AdSkip_InvalidAdID(t *testing.T) {
	app, queries, _ := setupRedirectApp(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)

	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:   user.ID,
		Slug:     "skip-test",
		Original: "https://example.com/dest",
		Custom:   false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/skip-test/skip/invalid-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestRedirectHandler_AdCompleteFlow_InvalidToken(t *testing.T) {
	app, _, _ := setupRedirectApp(t)

	req := httptest.NewRequest("POST", "/test-slug/complete?token=invalid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestRedirectHandler_AdCompleteFlow_MissingToken(t *testing.T) {
	app, _, _ := setupRedirectApp(t)

	req := httptest.NewRequest("POST", "/test-slug/complete", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestRedirectHandler_MonetizedURLWithAds(t *testing.T) {
	app, queries, _ := setupRedirectApp(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)

	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:            user.ID,
		Slug:              "monetized",
		Original:          "https://example.com/monetized",
		Custom:            false,
		IsMonetized:       true,
		AllowedCategories: []string{"regular"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/monetized", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestRedirectHandler_ExpiredURL(t *testing.T) {
	app, queries, _ := setupRedirectApp(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)

	expired := time.Now().Add(-1 * time.Hour)
	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:    user.ID,
		Slug:      "expired-url",
		Original:  "https://example.com/expired",
		Custom:    false,
		ExpiresAt: sql.NullTime{Time: expired, Valid: true},
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/expired-url", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 410, resp.StatusCode)
}

func TestResolveGeo_ValidIPWithGeoDB(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	db, err := geoip2.Open("../../../GeoLite2-City.mmdb")
	if err != nil {
		t.Skip("GeoIP database not available")
	}
	defer db.Close()

	service := &RedirectService{urlSvc: nil, worker: nil, geoDB: db}
	country, city := service.resolveGeo("8.8.8.8")
	_ = country
	_ = city
}