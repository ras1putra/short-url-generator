package dto

import (
	"testing"
	"database/sql"
	"time"

	"urlshortener/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserToResponse(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	user := repository.User{
		ID:        id,
		Email:     "test@example.com",
		Name:      "Test User",
		Password: sql.NullString{String: "hashed-password", Valid: true},
		CreatedAt: now,
	}

	resp := MapUserToResponse(user)

	assert.Equal(t, id.String(), resp.ID)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "Test User", resp.Name)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
}

func TestMapUserToResponse_FieldsMappedCorrectly(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	user := repository.User{
		ID:        id,
		Email:     "user@domain.org",
		Name:      "Jane Doe",
		Password: sql.NullString{String: "secret", Valid: true},
		CreatedAt: now,
	}

	resp := MapUserToResponse(user)

	assert.Equal(t, user.ID.String(), resp.ID, "ID should be converted to string")
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.Name, resp.Name)
	assert.Equal(t, user.CreatedAt.Format(time.RFC3339), resp.CreatedAt)
}

func TestNewAuthResponse(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	user := repository.User{
		ID:        id,
		Email:     "test@example.com",
		Name:      "Test User",
		Password: sql.NullString{String: "hashed-password", Valid: true},
		CreatedAt: now,
	}

	resp := NewAuthResponse(user, "access-token-123", "refresh-token-456")

	require.NotNil(t, resp)
	assert.Equal(t, "access-token-123", resp.AccessToken)
	assert.Equal(t, "refresh-token-456", resp.RefreshToken)
	assert.Equal(t, id.String(), resp.User.ID)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "Test User", resp.User.Name)
	assert.Equal(t, now.Format(time.RFC3339), resp.User.CreatedAt)
}

func TestNewAuthResponse_NilFields(t *testing.T) {
	id := uuid.New()
	user := repository.User{
		ID:        id,
		Email:     "another@test.com",
		Name:      "Another",
		Password: sql.NullString{String: "hash", Valid: true},
		CreatedAt: time.Now(),
	}

	resp := NewAuthResponse(user, "at", "rt")

	require.NotNil(t, resp)
	assert.Equal(t, "at", resp.AccessToken)
	assert.Equal(t, "rt", resp.RefreshToken)
	assert.Equal(t, id.String(), resp.User.ID)
	assert.Equal(t, "another@test.com", resp.User.Email)
	assert.Equal(t, "Another", resp.User.Name)
}

func TestNewAccessTokenResponse(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	user := repository.User{
		ID:        id,
		Email:     "test@example.com",
		Name:      "Test User",
		Password: sql.NullString{String: "hash", Valid: true},
		CreatedAt: now,
	}

	resp := NewAccessTokenResponse(user, "access-token-123")

	require.NotNil(t, resp)
	assert.Equal(t, "access-token-123", resp.AccessToken)
	assert.Empty(t, resp.RefreshToken)
	assert.Equal(t, id.String(), resp.User.ID)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "Test User", resp.User.Name)
	assert.Equal(t, now.Format(time.RFC3339), resp.User.CreatedAt)
}