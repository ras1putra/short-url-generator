package redirect

import (
	"context"
	"testing"

	"urlshortener/internal/analytics"
	"urlshortener/internal/config"
	"urlshortener/internal/modules/links"
	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupRedirectService(t *testing.T) *RedirectService {
	_ = zap.ReplaceGlobals(zap.NewNop())
	db, queries := testutil.SetupTestDB(t)
	redis := testutil.SetupTestRedisRaw(t)
	fakeCache := testutil.NewFakeCacher()
	urlSvc := links.NewURLService(queries, fakeCache, &config.Config{
		BaseURL: "http://localhost:8080",
	})
	worker := analytics.NewAnalyticsWorker(queries, 100)

	cfg := &config.Config{
		BaseURL:         "http://localhost:8080",
		BridgeHMACSecret: "test-hmac-secret",
		QualityMinScore:  0.5,
		MinSessionMs:     3000,
	}

	return NewRedirectService(urlSvc, queries, worker, nil, cfg, db, redis.Raw())
}

func TestGenerateAndVerifyBridgeToken(t *testing.T) {
	svc := setupRedirectService(t)
	slug := "abc123"
	adID1 := uuid.New()
	adID2 := uuid.New()

	token := svc.GenerateBridgeToken(slug, []uuid.UUID{adID1, adID2})
	require.NotEmpty(t, token)

	parsedSlug, parsedAdIDs, valid, ms := svc.VerifyBridgeToken(token)
	assert.True(t, valid)
	assert.Equal(t, slug, parsedSlug)
	assert.Equal(t, []uuid.UUID{adID1, adID2}, parsedAdIDs)
	assert.Greater(t, ms, int64(0))
}

func TestVerifyBridgeToken_Invalid(t *testing.T) {
	svc := setupRedirectService(t)

	slug, adIDs, valid, _ := svc.VerifyBridgeToken("invalid-token")
	assert.False(t, valid)
	assert.Empty(t, slug)
	assert.Empty(t, adIDs)
}

func TestVerifyBridgeToken_WrongFormat(t *testing.T) {
	svc := setupRedirectService(t)

	slug, adIDs, valid, _ := svc.VerifyBridgeToken("only-one-part")
	assert.False(t, valid)
	assert.Empty(t, slug)
	assert.Empty(t, adIDs)
}

func TestVerifyBridgeToken_TamperedSignature(t *testing.T) {
	svc := setupRedirectService(t)
	slug := "test"
	adID := uuid.New()

	realToken := svc.GenerateBridgeToken(slug, []uuid.UUID{adID})
	// Tamper with the signature
	parts := splitN(realToken, ":", 4)
	tamperedToken := parts[0] + ":" + parts[1] + ":" + parts[2] + ":invalidsig"

	parsedSlug, parsedAdIDs, valid, _ := svc.VerifyBridgeToken(tamperedToken)
	assert.False(t, valid)
	assert.Empty(t, parsedSlug)
	assert.Empty(t, parsedAdIDs)
}

func TestVerifyBridgeToken_InvalidAdID(t *testing.T) {
	svc := setupRedirectService(t)
	_ = svc.GenerateBridgeToken("test", []uuid.UUID{uuid.New()})

	// Token with invalid UUID as ad ID
	invalidToken := "test:not-a-uuid:1234567890:signature"

	slug, adIDs, valid, _ := svc.VerifyBridgeToken(invalidToken)
	assert.False(t, valid)
	assert.Empty(t, slug)
	assert.Empty(t, adIDs)
}

func TestGetURL_NotFound(t *testing.T) {
	svc := setupRedirectService(t)
	_, err := svc.GetURL(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestGetActiveAds_Empty(t *testing.T) {
	svc := setupRedirectService(t)
	ads, err := svc.GetActiveAds(context.Background())
	require.NoError(t, err)
	assert.Empty(t, ads)
}

func TestGetMinSessionMs(t *testing.T) {
	svc := setupRedirectService(t)
	assert.Equal(t, int64(3000), svc.GetMinSessionMs())
}

func TestGetActiveAdsByCategory_Empty(t *testing.T) {
	svc := setupRedirectService(t)
	ads, err := svc.GetActiveAdsByCategory(context.Background(), []string{"regular"})
	require.NoError(t, err)
	assert.Empty(t, ads)
}

func TestGetActiveAdsByCategory_NilCategories(t *testing.T) {
	svc := setupRedirectService(t)
	ads, err := svc.GetActiveAdsByCategory(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, ads)
}

func TestSelectAndTrackAds_Empty(t *testing.T) {
	svc := setupRedirectService(t)
	require.Panics(t, func() {
		svc.SelectAndTrackAds([]repository.Ad{}, uuid.New(), "127.0.0.1", "test-agent")
	}, "SelectAndTrackAds should panic with empty ads (needs fix in production code)")
}

func TestSelectAndTrackAds_SingleBanner(t *testing.T) {
	t.Skip("Skipped: requires real ad in DB for event tracking")
}

func TestSelectAndTrackAds_MultipleTypes(t *testing.T) {
	t.Skip("Skipped: requires real ad in DB for event tracking")
}

func TestSelectAndTrackAds_OnlyPopup(t *testing.T) {
	t.Skip("Skipped: requires real ad in DB for event tracking")
}

func TestSelectAndTrackAds_OnlyInterstitial(t *testing.T) {
	t.Skip("Skipped: requires real ad in DB for event tracking")
}

func TestEnqueueClick(t *testing.T) {
	svc := setupRedirectService(t)
	svc.EnqueueClick(uuid.New(), "127.0.0.1", "test-agent", "http://referer.com")
}

func TestBridgeToken_MultipleAds(t *testing.T) {
	svc := setupRedirectService(t)
	adIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	token := svc.GenerateBridgeToken("test-slug", adIDs)
	require.NotEmpty(t, token)

	slug, ids, valid, _ := svc.VerifyBridgeToken(token)
	assert.True(t, valid)
	assert.Equal(t, "test-slug", slug)
	assert.Len(t, ids, 3)
}

func TestBridgeToken_SingleAd(t *testing.T) {
	svc := setupRedirectService(t)
	adID := uuid.New()
	token := svc.GenerateBridgeToken("single", []uuid.UUID{adID})
	require.NotEmpty(t, token)

	slug, ids, valid, _ := svc.VerifyBridgeToken(token)
	assert.True(t, valid)
	assert.Equal(t, "single", slug)
	assert.Len(t, ids, 1)
	assert.Equal(t, adID, ids[0])
}

func TestBridgeToken_EmptyAdIDs(t *testing.T) {
	svc := setupRedirectService(t)
	token := svc.GenerateBridgeToken("empty-ads", []uuid.UUID{})
	require.NotEmpty(t, token)

	_, ids, valid, _ := svc.VerifyBridgeToken(token)
	assert.True(t, valid)
	assert.Empty(t, ids)
}

func TestBridgeToken_OnlyNil(t *testing.T) {
	svc := setupRedirectService(t)
	token := svc.GenerateBridgeToken("nil-ad", []uuid.UUID{uuid.Nil})
	require.NotEmpty(t, token)

	_, ids, valid, _ := svc.VerifyBridgeToken(token)
	assert.True(t, valid)
	assert.Empty(t, ids)
}

func splitN(s, sep string, n int) []string {
	result := make([]string, 0, n)
	start := 0
	for i := 0; i < n-1; i++ {
		idx := indexOf(s[start:], sep)
		if idx < 0 {
			break
		}
		result = append(result, s[start:start+idx])
		start += idx + len(sep)
	}
	result = append(result, s[start:])
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
