package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key"

func TestIssueToken_Access(t *testing.T) {
	token, err := IssueToken("user-123", "user", testSecret, "access", 15*time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := Validate(token, testSecret, "access")
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, "access", claims.TokenType)
}

func TestIssueToken_Refresh(t *testing.T) {
	token, err := IssueToken("user-456", "advertiser", testSecret, "refresh", 7*24*time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := Validate(token, testSecret, "refresh")
	assert.NoError(t, err)
	assert.Equal(t, "user-456", claims.UserID)
	assert.Equal(t, "advertiser", claims.Role)
	assert.Equal(t, "refresh", claims.TokenType)
}

func TestValidate_WrongSecret(t *testing.T) {
	token, err := IssueToken("user-123", "user", testSecret, "access", 15*time.Minute)
	assert.NoError(t, err)

	claims, err := Validate(token, "wrong-secret", "access")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidate_WrongTokenType(t *testing.T) {
	token, err := IssueToken("user-123", "user", testSecret, "refresh", 7*24*time.Hour)
	assert.NoError(t, err)

	claims, err := Validate(token, testSecret, "access")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidate_ExpiredToken(t *testing.T) {
	token, err := IssueToken("user-123", "user", testSecret, "access", -1*time.Hour)
	assert.NoError(t, err)

	claims, err := Validate(token, testSecret, "access")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidate_InvalidToken(t *testing.T) {
	claims, err := Validate("not-a-token", testSecret, "access")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidate_EmptyToken(t *testing.T) {
	claims, err := Validate("", testSecret, "access")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestIssueToken_DifferentUsers(t *testing.T) {
	token1, _ := IssueToken("user-1", "user", testSecret, "access", 15*time.Minute)
	token2, _ := IssueToken("user-2", "user", testSecret, "access", 15*time.Minute)
	assert.NotEqual(t, token1, token2)
}

func TestValidate_ClaimsFields(t *testing.T) {
	token, err := IssueToken("user-789", "user", testSecret, "access", 15*time.Minute)
	assert.NoError(t, err)

	claims, err := Validate(token, testSecret, "access")
	assert.NoError(t, err)
	assert.Equal(t, "user-789", claims.UserID)
	assert.Equal(t, "access", claims.TokenType)
	assert.NotZero(t, claims.IssuedAt)
	assert.NotZero(t, claims.ExpiresAt)
}

func TestValidate_WrongSigningMethod(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	claims := Claims{
		UserID:    "user-123",
		Role:      "user",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	result, err := Validate(tokenString, testSecret, "access")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidToken, err)
}
