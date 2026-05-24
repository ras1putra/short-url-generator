package dto

import (
	"database/sql"
	"testing"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapURLToResponse_NonExpiredURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	id := uuid.New()
	userID := uuid.New()

	url := repository.Url{
		ID:        id,
		UserID:    userID,
		Slug:      "abc123",
		Original:  "https://example.com",
		Custom:    false,
		ExpiresAt: sql.NullTime{Valid: false},
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := MapURLToResponse(url, cfg)

	assert.Equal(t, "abc123", resp.Slug)
	assert.Equal(t, "http://localhost:8080/abc123", resp.ShortURL)
	assert.Equal(t, "https://example.com", resp.Original)
	assert.Equal(t, "http://localhost:8080/api/links/abc123/qr", resp.QRURL)
	assert.Nil(t, resp.ExpiresAt, "ExpiresAt should be nil for non-expiring URL")
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
}

func TestMapURLToResponse_ExpiredURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	expiresAt := now.Add(24 * time.Hour)
	id := uuid.New()
	userID := uuid.New()

	url := repository.Url{
		ID:        id,
		UserID:    userID,
		Slug:      "xyz789",
		Original:  "https://example.org/long-url",
		Custom:    true,
		ExpiresAt: sql.NullTime{Valid: true, Time: expiresAt},
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := MapURLToResponse(url, cfg)

	assert.Equal(t, "xyz789", resp.Slug)
	assert.Equal(t, "http://localhost:8080/xyz789", resp.ShortURL)
	assert.Equal(t, "https://example.org/long-url", resp.Original)
	assert.Equal(t, "http://localhost:8080/api/links/xyz789/qr", resp.QRURL)
	require.NotNil(t, resp.ExpiresAt, "ExpiresAt should not be nil for expiring URL")
	assert.Equal(t, expiresAt, *resp.ExpiresAt)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
}

func TestMapURLToResponse_DifferentBaseURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://short.ly"}
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	id := uuid.New()
	userID := uuid.New()

	url := repository.Url{
		ID:        id,
		UserID:    userID,
		Slug:      "myslug",
		Original:  "https://very-long-url.com/path",
		Custom:    true,
		ExpiresAt: sql.NullTime{Valid: false},
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := MapURLToResponse(url, cfg)

	assert.Equal(t, "https://short.ly/myslug", resp.ShortURL)
	assert.Equal(t, "https://short.ly/api/links/myslug/qr", resp.QRURL)
}

func TestMapURLToResponse_ZeroTimeExpiresAt(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	now := time.Date(2024, 3, 10, 5, 30, 0, 0, time.UTC)
	id := uuid.New()
	userID := uuid.New()

	url := repository.Url{
		ID:        id,
		UserID:    userID,
		Slug:      "zeroexp",
		Original:  "https://test.com",
		Custom:    false,
		ExpiresAt: sql.NullTime{Valid: true, Time: time.Time{}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := MapURLToResponse(url, cfg)

	assert.NotNil(t, resp.ExpiresAt, "ExpiresAt should be set when Valid is true even with zero time")
	assert.Equal(t, time.Time{}, *resp.ExpiresAt)
}