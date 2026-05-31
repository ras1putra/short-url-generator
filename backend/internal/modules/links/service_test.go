package links

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/links/dto"
	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/response"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testCfg = &config.Config{
	BaseURL:          "http://localhost:8080",
	JWTAccessSecret:  "test-secret",
	JWTRefreshSecret: "test-refresh-secret",
}

func newTestURLService(t *testing.T) (*URLService, *repository.Queries) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	_, queries := testutil.SetupTestDB(t)
	fakeCache := testutil.NewFakeCacher()
	return NewURLService(queries, fakeCache, testCfg), queries
}

func createTestUser(t *testing.T, queries *repository.Queries, ctx context.Context, name, email string) repository.User {
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     name,
		Email:    email,
		Password: sql.NullString{String: "hashed", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)
	return user
}

func createTestURL(t *testing.T, queries *repository.Queries, ctx context.Context, userID uuid.UUID, slug, original string) repository.Url {
	url, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:   userID,
		Slug:     slug,
		Original: original,
		Custom:   false,
	})
	require.NoError(t, err)
	return url
}

func appErrCode(err error) int {
	var appErr *response.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return -1
}

func TestURLService_Create_SuccessAutoSlug(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")

	req := dto.CreateURLRequest{
		URL: "https://example.com",
	}

	resp, err := svc.Create(ctx, user.ID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://example.com", resp.Original)
	assert.Contains(t, resp.ShortURL, "http://localhost:8080/")
	assert.NotEmpty(t, resp.Slug)
	assert.Len(t, resp.Slug, 6)
}

func TestURLService_Create_SuccessCustomSlug(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")

	req := dto.CreateURLRequest{
		URL:        "https://example.com",
		CustomSlug: "my-custom",
	}

	resp, err := svc.Create(ctx, user.ID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "my-custom", resp.Slug)
	assert.Equal(t, "https://example.com", resp.Original)
	assert.Equal(t, "http://localhost:8080/my-custom", resp.ShortURL)
}

func TestURLService_Create_ReservedSlug(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")

	req := dto.CreateURLRequest{
		URL:        "https://example.com",
		CustomSlug: "api",
	}

	resp, err := svc.Create(ctx, user.ID, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
}

func TestURLService_Create_TakenCustomSlug(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "taken-slug", "https://example.com")

	req := dto.CreateURLRequest{
		URL:        "https://example.org",
		CustomSlug: "taken-slug",
	}

	resp, err := svc.Create(ctx, user.ID, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
}

func TestURLService_Create_WithExpiry(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")

	req := dto.CreateURLRequest{
		URL:          "https://example.com",
		ExpiresValue: 2,
		ExpiresUnit:  "hours",
	}

	resp, err := svc.Create(ctx, user.ID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Slug)
	assert.NotNil(t, resp.ExpiresAt)
}

func TestURLService_GetBySlug_CacheMiss(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "dburl", "https://example.com")

	result, err := svc.GetBySlug(ctx, "dburl")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "dburl", result.Slug)
	assert.Equal(t, "https://example.com", result.Original)
}

func TestURLService_GetBySlug_NotFound(t *testing.T) {
	svc, _ := newTestURLService(t)
	ctx := context.Background()

	result, err := svc.GetBySlug(ctx, "notfound")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 404, appErrCode(err))
}

func TestURLService_GetBySlug_Expired(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")

	_, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:    user.ID,
		Slug:      "expired",
		Original:  "https://example.com",
		Custom:    false,
		ExpiresAt: sql.NullTime{Valid: true, Time: time.Now().Add(-time.Hour)},
	})
	require.NoError(t, err)

	result, err := svc.GetBySlug(ctx, "expired")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 410, appErrCode(err))
}

func TestURLService_ListByUser_Success(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "abc1", "https://example.com/1")
	createTestURL(t, queries, ctx, user.ID, "abc2", "https://example.com/2")

	result, err := svc.ListByUser(ctx, user.ID, 1, 10, "", nil, "created_at", "desc")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PerPage)
	assert.Equal(t, 1, result.TotalPages)
	assert.Len(t, result.Links, 2)
}

func TestURLService_ListByUser_Pagination(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	for i := 0; i < 15; i++ {
		createTestURL(t, queries, ctx, user.ID, uuid.New().String()[:6], "https://example.com/"+uuid.New().String())
	}

	result, err := svc.ListByUser(ctx, user.ID, 3, 5, "", nil, "created_at", "desc")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(15), result.Total)
	assert.Equal(t, 3, result.Page)
	assert.Equal(t, 5, result.PerPage)
	assert.Equal(t, 3, result.TotalPages)
	assert.Len(t, result.Links, 5)
}

func TestURLService_Delete_Success(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "del-me", "https://example.com")

	err := svc.Delete(ctx, user.ID, "del-me")
	require.NoError(t, err)

	_, err = queries.GetURLBySlug(ctx, "del-me")
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestURLService_Delete_NotFound(t *testing.T) {
	svc, _ := newTestURLService(t)
	ctx := context.Background()

	err := svc.Delete(ctx, uuid.New(), "notfound")
	require.Error(t, err)
	assert.Equal(t, 404, appErrCode(err))
}

func TestURLService_Delete_Forbidden(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	owner := createTestUser(t, queries, ctx, "Owner", "owner@example.com")
	other := createTestUser(t, queries, ctx, "Other", "other@example.com")
	createTestURL(t, queries, ctx, owner.ID, "owned-url", "https://example.com")

	err := svc.Delete(ctx, other.ID, "owned-url")
	require.Error(t, err)
	assert.Equal(t, 403, appErrCode(err))
}

func TestURLService_GetByID_Success(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "get-me", "https://example.com")

	resp, err := svc.GetByID(ctx, user.ID, "get-me")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "get-me", resp.Slug)
	assert.Equal(t, "https://example.com", resp.Original)
	assert.Equal(t, "http://localhost:8080/get-me", resp.ShortURL)
}

func TestURLService_GetByID_NotFound(t *testing.T) {
	svc, _ := newTestURLService(t)
	ctx := context.Background()

	resp, err := svc.GetByID(ctx, uuid.New(), "notfound")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
}

func TestURLService_GetByID_Forbidden(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	owner := createTestUser(t, queries, ctx, "Owner", "owner@example.com")
	other := createTestUser(t, queries, ctx, "Other", "other@example.com")
	createTestURL(t, queries, ctx, owner.ID, "owned", "https://example.com")

	resp, err := svc.GetByID(ctx, other.ID, "owned")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 403, appErrCode(err))
}

func TestURLService_Update_SuccessChangeSlug(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "old-slug", "https://example.com")

	req := dto.UpdateURLRequest{
		CustomSlug: "new-slug",
	}

	resp, err := svc.Update(ctx, user.ID, "old-slug", req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "new-slug", resp.Slug)

	_, err = queries.GetURLBySlug(ctx, "old-slug")
	require.ErrorIs(t, err, sql.ErrNoRows)

	_, err = queries.GetURLBySlug(ctx, "new-slug")
	require.NoError(t, err)
}

func TestURLService_Update_SuccessNoSlugChange(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "same-slug", "https://example.com")

	req := dto.UpdateURLRequest{}

	resp, err := svc.Update(ctx, user.ID, "same-slug", req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "same-slug", resp.Slug)
}

func TestURLService_Update_NotFound(t *testing.T) {
	svc, _ := newTestURLService(t)
	ctx := context.Background()

	req := dto.UpdateURLRequest{CustomSlug: "new-slug"}

	resp, err := svc.Update(ctx, uuid.New(), "notfound", req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
}

func TestURLService_Update_Forbidden(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	owner := createTestUser(t, queries, ctx, "Owner", "owner@example.com")
	other := createTestUser(t, queries, ctx, "Other", "other@example.com")
	createTestURL(t, queries, ctx, owner.ID, "owned", "https://example.com")

	req := dto.UpdateURLRequest{CustomSlug: "new-slug"}

	resp, err := svc.Update(ctx, other.ID, "owned", req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 403, appErrCode(err))
}

func TestURLService_Update_ReservedSlug(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "current-slug", "https://example.com")

	req := dto.UpdateURLRequest{CustomSlug: "api"}

	resp, err := svc.Update(ctx, user.ID, "current-slug", req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
}

func TestURLService_Update_TakenSlug(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	createTestURL(t, queries, ctx, user.ID, "current-slug", "https://example.com")
	createTestURL(t, queries, ctx, user.ID, "taken-slug", "https://example.com/taken")

	req := dto.UpdateURLRequest{CustomSlug: "taken-slug"}

	resp, err := svc.Update(ctx, user.ID, "current-slug", req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
}

func TestURLService_GetStats_Success(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	url := createTestURL(t, queries, ctx, user.ID, "stats-slug", "https://example.com")

	_, err := queries.SaveClick(ctx, repository.SaveClickParams{
		UrlID:      url.ID,
		IpHash:     sql.NullString{String: "hash1", Valid: true},
		Country:    sql.NullString{String: "US", Valid: true},
		Device:     sql.NullString{String: "Mobile", Valid: true},
		Browser:    sql.NullString{String: "Chrome", Valid: true},
		Referrer:   sql.NullString{String: "https://google.com", Valid: true},
	})
	require.NoError(t, err)

	_, err = queries.SaveClick(ctx, repository.SaveClickParams{
		UrlID:      url.ID,
		IpHash:     sql.NullString{String: "hash2", Valid: true},
		Country:    sql.NullString{String: "ID", Valid: true},
		Device:     sql.NullString{String: "Desktop", Valid: true},
		Browser:    sql.NullString{String: "Firefox", Valid: true},
		Referrer:   sql.NullString{String: "", Valid: false},
	})
	require.NoError(t, err)

	resp, err := svc.GetStats(ctx, user.ID, "stats-slug")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(2), resp.TotalClicks)
	assert.Equal(t, int64(2), resp.UniqueClicks)
	assert.Len(t, resp.TopCountries, 2)
	assert.Len(t, resp.ClicksPerDay, 1)
	assert.Contains(t, resp.Browsers, "Chrome")
	assert.Contains(t, resp.Browsers, "Firefox")
	assert.Contains(t, resp.Devices, "Mobile")
	assert.Contains(t, resp.Devices, "Desktop")
}

func TestURLService_GetStats_NotFound(t *testing.T) {
	svc, _ := newTestURLService(t)
	ctx := context.Background()

	resp, err := svc.GetStats(ctx, uuid.New(), "notfound")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
}

func TestURLService_GetStats_Forbidden(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	owner := createTestUser(t, queries, ctx, "Owner", "owner@example.com")
	other := createTestUser(t, queries, ctx, "Other", "other@example.com")
	createTestURL(t, queries, ctx, owner.ID, "owned-stats", "https://example.com")

	resp, err := svc.GetStats(ctx, other.ID, "owned-stats")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 403, appErrCode(err))
}

func TestURLService_StartExpiryCleaner_StopsOnContextCancellation(t *testing.T) {
	svc, _ := newTestURLService(t)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.StartExpiryCleaner(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("StartExpiryCleaner did not exit after context cancellation")
	}
}

func TestURLService_GetAggregateStats_Success(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	url1 := createTestURL(t, queries, ctx, user.ID, "agg1", "https://example.com/1")
	url2 := createTestURL(t, queries, ctx, user.ID, "agg2", "https://example.com/2")

	_, err := queries.SaveClick(ctx, repository.SaveClickParams{
		UrlID:      url1.ID,
		IpHash:     sql.NullString{String: "hash1", Valid: true},
		Country:    sql.NullString{String: "US", Valid: true},
		Device:     sql.NullString{String: "Mobile", Valid: true},
		Browser:    sql.NullString{String: "Chrome", Valid: true},
	})
	require.NoError(t, err)

	_, err = queries.SaveClick(ctx, repository.SaveClickParams{
		UrlID:      url2.ID,
		IpHash:     sql.NullString{String: "hash2", Valid: true},
		Country:    sql.NullString{String: "ID", Valid: true},
		Device:     sql.NullString{String: "Desktop", Valid: true},
		Browser:    sql.NullString{String: "Firefox", Valid: true},
	})
	require.NoError(t, err)

	resp, err := svc.GetAggregateStats(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(2), resp.TotalClicks)
	assert.Equal(t, int64(2), resp.UniqueClicks)
	assert.Len(t, resp.TopCountries, 2)
	assert.Contains(t, resp.Browsers, "Chrome")
	assert.Contains(t, resp.Browsers, "Firefox")
	assert.Contains(t, resp.Devices, "Mobile")
	assert.Contains(t, resp.Devices, "Desktop")
}

func TestURLService_GetAggregateStats_NoClicks(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")

	resp, err := svc.GetAggregateStats(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(0), resp.TotalClicks)
	assert.Equal(t, int64(0), resp.UniqueClicks)
	assert.Empty(t, resp.TopCountries)
	assert.Empty(t, resp.Browsers)
	assert.Empty(t, resp.Devices)
}

func TestURLService_GetAggregateStats_UnknownCountry(t *testing.T) {
	svc, queries := newTestURLService(t)
	ctx := context.Background()

	user := createTestUser(t, queries, ctx, "Test User", "test@example.com")
	url := createTestURL(t, queries, ctx, user.ID, "unknown", "https://example.com")

	_, err := queries.SaveClick(ctx, repository.SaveClickParams{
		UrlID:      url.ID,
		IpHash:     sql.NullString{String: "hash1", Valid: true},
		Device:     sql.NullString{String: "Bot", Valid: true},
	})
	require.NoError(t, err)

	resp, err := svc.GetAggregateStats(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.TotalClicks)
	found := false
	for _, c := range resp.TopCountries {
		if c.Country == "Unknown" {
			found = true
			assert.Equal(t, int64(1), c.Count)
		}
	}
	assert.True(t, found, "Expected Unknown country entry")
}
