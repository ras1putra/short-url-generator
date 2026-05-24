package links

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupLinksApp(t *testing.T) (*fiber.App, *repository.Queries) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	_, queries := testutil.SetupTestDB(t)
	fakeCache := testutil.NewFakeCacher()
	svc := NewURLService(queries, fakeCache, testCfg)
	handler := NewLinksHandler(svc, testCfg)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		userIDStr := c.Get("X-Test-User-ID")
		if userIDStr != "" {
			c.Locals("user_id", userIDStr)
		}
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

	return app, queries
}

func createHandlerTestUser(t *testing.T, queries *repository.Queries, ctx context.Context) repository.User {
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Test User",
		Email:    "test-" + uuid.New().String() + "@example.com",
		Password: "password",
		Role:     "user",
	})
	require.NoError(t, err)
	return user
}

func TestHandlerCreateSuccess(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", user.ID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestHandlerCreate_InvalidJSON(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)

	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", user.ID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerListSuccess(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	req := httptest.NewRequest("GET", "/api/links", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerGetSuccess(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	req := httptest.NewRequest("GET", "/api/links/abc123", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerDeleteSuccess(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	req := httptest.NewRequest("DELETE", "/api/links/abc123", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerStatsSuccess(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	url := createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	_, err := queries.SaveClick(ctx, repository.SaveClickParams{
		UrlID:   url.ID,
		IpHash:  sql.NullString{String: "ip1", Valid: true},
		Country: sql.NullString{String: "US", Valid: true},
		Device:  sql.NullString{String: "Mobile", Valid: true},
		Browser: sql.NullString{String: "Chrome", Valid: true},
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/links/abc123/stats", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerQRCodeSuccess(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	req := httptest.NewRequest("GET", "/api/links/abc123/qr?size=128", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))
}

func TestHandlerQRCode_CustomSize(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	req := httptest.NewRequest("GET", "/api/links/abc123/qr?size=512", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))
}

func TestHandlerQRCode_InvalidSize(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	req := httptest.NewRequest("GET", "/api/links/abc123/qr?size=0", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerQRCode_InvalidSizeParam(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	req := httptest.NewRequest("GET", "/api/links/abc123/qr?size=notanumber", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerQRCode_QRGenFailed(t *testing.T) {
	app, _ := setupLinksApp(t)

	req := httptest.NewRequest("GET", "/api/links/unknown/qr", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestHandlerCreate_NoUserID(t *testing.T) {
	app, _ := setupLinksApp(t)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerCreate_InvalidUserID(t *testing.T) {
	app, _ := setupLinksApp(t)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", "not-a-uuid")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerList_NoUserID(t *testing.T) {
	app, _ := setupLinksApp(t)

	req := httptest.NewRequest("GET", "/api/links", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerList_PageBounds(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)

	req := httptest.NewRequest("GET", "/api/links?page=-1&per_page=200", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerDelete_NoUserID(t *testing.T) {
	app, _ := setupLinksApp(t)

	req := httptest.NewRequest("DELETE", "/api/links/abc123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerGet_NoUserID(t *testing.T) {
	app, _ := setupLinksApp(t)

	req := httptest.NewRequest("GET", "/api/links/abc123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerStats_NoUserID(t *testing.T) {
	app, _ := setupLinksApp(t)

	req := httptest.NewRequest("GET", "/api/links/abc123/stats", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerUpdate_Success(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)
	createTestURL(t, queries, ctx, user.ID, "abc123", "https://example.com")

	body, _ := json.Marshal(map[string]string{"custom_slug": "newslug"})
	req := httptest.NewRequest("PATCH", "/api/links/abc123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", user.ID.String())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerUpdate_NoUserID(t *testing.T) {
	app, _ := setupLinksApp(t)

	body, _ := json.Marshal(map[string]string{"custom_slug": "newslug"})
	req := httptest.NewRequest("PATCH", "/api/links/abc123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandlerUpdate_ValidationError(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)

	body, _ := json.Marshal(map[string]string{"custom_slug": "ab"})
	req := httptest.NewRequest("PATCH", "/api/links/abc123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHandlerAggregateStatsSuccess(t *testing.T) {
	app, queries := setupLinksApp(t)
	ctx := context.Background()
	user := createHandlerTestUser(t, queries, ctx)

	req := httptest.NewRequest("GET", "/api/links/stats/aggregate", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlerAggregateStats_NoUserID(t *testing.T) {
	app, _ := setupLinksApp(t)

	req := httptest.NewRequest("GET", "/api/links/stats/aggregate", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}