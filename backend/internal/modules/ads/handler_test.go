package ads

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"urlshortener/internal/modules/ads/dto"
	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupAdsApp(t *testing.T) (*fiber.App, *repository.Queries) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	db, queries := testutil.SetupTestDB(t)
	svc := NewAdService(db, queries)
	handler := NewAdHandler(svc)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		userIDStr := c.Get("X-Test-User-ID")
		if userIDStr != "" {
			c.Locals("user_id", userIDStr)
		}
		c.Locals("request_id", "test-req-id")
		return c.Next()
	})
	app.Post("/ads", handler.Create)
	app.Get("/ads", handler.List)
	app.Get("/ads/:id", handler.GetByID)
	app.Patch("/ads/:id", handler.Update)
	app.Delete("/ads/:id", handler.Delete)
	app.Get("/ads/:id/stats", handler.GetStats)
	app.Post("/ads/:id/topup", handler.TopUp)

	return app, queries
}

func TestAdsHandler_Create_Success(t *testing.T) {
	app, queries := setupAdsApp(t)
	ctx := context.Background()
	user := createAdUser(t, queries, ctx)

	body, _ := json.Marshal(dto.CreateAdRequest{
		Title: "New Ad", ImageURL: "https://ex.com/img.jpg", TargetURL: "https://ex.com",
		Category: "regular", TotalBudget: 100, AdType: "BANNER",
	})
	req := httptest.NewRequest("POST", "/ads", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", user.ID.String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
}

func TestAdsHandler_Create_InvalidJSON(t *testing.T) {
	app, queries := setupAdsApp(t)
	ctx := context.Background()
	user := createAdUser(t, queries, ctx)

	req := httptest.NewRequest("POST", "/ads", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", user.ID.String())
	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, r.StatusCode)
}

func TestAdsHandler_Create_ValidationError(t *testing.T) {
	app, queries := setupAdsApp(t)
	ctx := context.Background()
	user := createAdUser(t, queries, ctx)

	body, _ := json.Marshal(dto.CreateAdRequest{Title: "ab"})
	req := httptest.NewRequest("POST", "/ads", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", user.ID.String())
	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, r.StatusCode)
}

func TestAdsHandler_Create_Unauthorized(t *testing.T) {
	app, _ := setupAdsApp(t)

	body, _ := json.Marshal(dto.CreateAdRequest{
		Title: "New Ad", ImageURL: "https://ex.com/img.jpg", TargetURL: "https://ex.com",
		Category: "regular", TotalBudget: 100, AdType: "BANNER",
	})
	req := httptest.NewRequest("POST", "/ads", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, r.StatusCode)
}

func TestAdsHandler_List_Success(t *testing.T) {
	app, queries := setupAdsApp(t)
	ctx := context.Background()
	user := createAdUser(t, queries, ctx)

	req := httptest.NewRequest("GET", "/ads", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}

func TestAdsHandler_GetByID_Success(t *testing.T) {
	app, queries := setupAdsApp(t)
	ctx := context.Background()
	user := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, user.ID, "active")

	req := httptest.NewRequest("GET", "/ads/"+ad.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}

func TestAdsHandler_GetByID_InvalidID(t *testing.T) {
	app, _ := setupAdsApp(t)

	req := httptest.NewRequest("GET", "/ads/invalid-uuid", nil)
	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, r.StatusCode)
}

func TestAdsHandler_GetByID_NotFound(t *testing.T) {
	app, _ := setupAdsApp(t)

	req := httptest.NewRequest("GET", "/ads/"+uuid.New().String(), nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, r.StatusCode)
}

func TestAdsHandler_Update_Success(t *testing.T) {
	app, queries := setupAdsApp(t)
	ctx := context.Background()
	user := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, user.ID, "active")
	status := "paused"

	body, _ := json.Marshal(dto.UpdateAdRequest{Status: &status})
	req := httptest.NewRequest("PATCH", "/ads/"+ad.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", user.ID.String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}

func TestAdsHandler_Delete_Success(t *testing.T) {
	app, queries := setupAdsApp(t)
	ctx := context.Background()
	user := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, user.ID, "active")

	req := httptest.NewRequest("DELETE", "/ads/"+ad.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}

func TestAdsHandler_GetStats_Success(t *testing.T) {
	app, queries := setupAdsApp(t)
	ctx := context.Background()
	user := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, user.ID, "active")

	req := httptest.NewRequest("GET", "/ads/"+ad.ID.String()+"/stats", nil)
	req.Header.Set("X-Test-User-ID", user.ID.String())

	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}

func TestAdsHandler_GetStats_InvalidID(t *testing.T) {
	app, _ := setupAdsApp(t)

	req := httptest.NewRequest("GET", "/ads/invalid-uuid/stats", nil)
	r, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, r.StatusCode)
}
