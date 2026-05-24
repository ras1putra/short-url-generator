package redirect

import (
	"context"
	"testing"

	"urlshortener/internal/analytics"
	"urlshortener/internal/config"
	"urlshortener/internal/modules/links"
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
	adID := uuid.New()

	token := svc.GenerateBridgeToken(slug, adID)
	require.NotEmpty(t, token)

	parsedSlug, parsedAdID, valid, ms := svc.VerifyBridgeToken(token)
	assert.True(t, valid)
	assert.Equal(t, slug, parsedSlug)
	assert.Equal(t, adID, parsedAdID)
	assert.Greater(t, ms, int64(0))
}

func TestVerifyBridgeToken_Invalid(t *testing.T) {
	svc := setupRedirectService(t)

	slug, adID, valid, _ := svc.VerifyBridgeToken("invalid-token")
	assert.False(t, valid)
	assert.Empty(t, slug)
	assert.Equal(t, uuid.Nil, adID)
}

func TestVerifyBridgeToken_WrongFormat(t *testing.T) {
	svc := setupRedirectService(t)

	slug, adID, valid, _ := svc.VerifyBridgeToken("only-one-part")
	assert.False(t, valid)
	assert.Empty(t, slug)
	assert.Equal(t, uuid.Nil, adID)
}

func TestVerifyBridgeToken_TamperedSignature(t *testing.T) {
	svc := setupRedirectService(t)
	slug := "test"
	adID := uuid.New()

	realToken := svc.GenerateBridgeToken(slug, adID)
	// Tamper with the signature
	parts := splitN(realToken, ":", 4)
	tamperedToken := parts[0] + ":" + parts[1] + ":" + parts[2] + ":invalidsig"

	parsedSlug, parsedAdID, valid, _ := svc.VerifyBridgeToken(tamperedToken)
	assert.False(t, valid)
	assert.Empty(t, parsedSlug)
	assert.Equal(t, uuid.Nil, parsedAdID)
}

func TestVerifyBridgeToken_InvalidAdID(t *testing.T) {
	svc := setupRedirectService(t)
	_ = svc.GenerateBridgeToken("test", uuid.New())

	// Token with invalid UUID as ad ID
	invalidToken := "test:not-a-uuid:1234567890:signature"

	slug, adID, valid, _ := svc.VerifyBridgeToken(invalidToken)
	assert.False(t, valid)
	assert.Empty(t, slug)
	assert.Equal(t, uuid.Nil, adID)
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
