package auth

import (
	"testing"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/repository"
	"urlshortener/pkg/token"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeUntilExpiry_Future(t *testing.T) {
	claims := &token.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	d := timeUntilExpiry(claims)
	assert.Greater(t, d, time.Duration(0))
	assert.LessOrEqual(t, d, 1*time.Hour)
}

func TestTimeUntilExpiry_Past(t *testing.T) {
	claims := &token.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	d := timeUntilExpiry(claims)
	assert.Equal(t, time.Duration(0), d)
}

func TestTimeUntilExpiry_Zero(t *testing.T) {
	claims := &token.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now()),
		},
	}
	d := timeUntilExpiry(claims)
	assert.Equal(t, time.Duration(0), d)
}

func TestIssueTokens_Success(t *testing.T) {
	cfg := &config.Config{
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
	}
	user := repository.User{
		ID:   uuid.New(),
		Role: "user",
	}

	accessToken, refreshToken, err := issueTokens(user, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	parsedAccess, err := token.Validate(accessToken, cfg.JWTAccessSecret, "access")
	require.NoError(t, err)
	assert.Equal(t, user.ID.String(), parsedAccess.UserID)
	assert.Equal(t, "access", parsedAccess.TokenType)

	parsedRefresh, err := token.Validate(refreshToken, cfg.JWTRefreshSecret, "refresh")
	require.NoError(t, err)
	assert.Equal(t, user.ID.String(), parsedRefresh.UserID)
	assert.Equal(t, "refresh", parsedRefresh.TokenType)
}

func TestIssueTokens_AdvertiserRole(t *testing.T) {
	cfg := &config.Config{
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
	}
	user := repository.User{
		ID:   uuid.New(),
		Role: "advertiser",
	}

	accessToken, _, err := issueTokens(user, cfg)
	require.NoError(t, err)

	parsed, err := token.Validate(accessToken, cfg.JWTAccessSecret, "access")
	require.NoError(t, err)
	assert.Equal(t, "advertiser", parsed.Role)
}

func TestNewCookieHelper_SameSite_Dev(t *testing.T) {
	cfg := &config.Config{Env: "development"}
	h := newCookieHelper(cfg)
	assert.Equal(t, "Lax", h.sameSite())
}

func TestNewCookieHelper_SameSite_Prod(t *testing.T) {
	cfg := &config.Config{Env: "production"}
	h := newCookieHelper(cfg)
	assert.Equal(t, "Strict", h.sameSite())
}
