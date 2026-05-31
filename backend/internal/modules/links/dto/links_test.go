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

func TestAdEventItem_Fields(t *testing.T) {
	item := AdEventItem{
		Time:            "2024-01-01T00:00:00Z",
		EventType:       "CLICK",
		AdTitle:         "Test Ad",
		AdType:          "BANNER",
		IsValid:         true,
		QualityScore:    "1.00",
		RejectionReason: "",
	}
	assert.Equal(t, "CLICK", item.EventType)
	assert.Equal(t, "Test Ad", item.AdTitle)
	assert.True(t, item.IsValid)
	assert.Empty(t, item.RejectionReason)
}

func TestAdEventItem_WithRejectionReason(t *testing.T) {
	item := AdEventItem{
		Time:            "2024-01-01T00:00:00Z",
		EventType:       "COMPLETION",
		AdTitle:         "Bad Ad",
		AdType:          "POPUP",
		IsValid:         false,
		QualityScore:    "0.00",
		RejectionReason: "HONEYPOT_HIT",
	}
	assert.False(t, item.IsValid)
	assert.Equal(t, "HONEYPOT_HIT", item.RejectionReason)
}

func TestEventPaginationInfo(t *testing.T) {
	info := EventPaginationInfo{
		Total:      100,
		Page:       2,
		PerPage:    10,
		TotalPages: 10,
	}
	assert.Equal(t, int64(100), info.Total)
	assert.Equal(t, 2, info.Page)
	assert.Equal(t, 10, info.PerPage)
	assert.Equal(t, 10, info.TotalPages)
}

func TestAdEventListResponse_Empty(t *testing.T) {
	resp := AdEventListResponse{
		Events:     []AdEventItem{},
		Total:      0,
		Page:       1,
		PerPage:    10,
		TotalPages: 0,
	}
	assert.Empty(t, resp.Events)
	assert.Equal(t, int64(0), resp.Total)
}

func TestAdEventListResponse_WithEvents(t *testing.T) {
	resp := AdEventListResponse{
		Events: []AdEventItem{
			{Time: "2024-01-01T00:00:00Z", EventType: "IMPRESSION", AdTitle: "Ad1", AdType: "BANNER"},
			{Time: "2024-01-01T00:00:01Z", EventType: "CLICK", AdTitle: "Ad1", AdType: "BANNER"},
		},
		Total:      2,
		Page:       1,
		PerPage:    10,
		TotalPages: 1,
	}
	assert.Len(t, resp.Events, 2)
}

func TestStatsResponse_WithEvents(t *testing.T) {
	stats := StatsResponse{
		TotalClicks:  10,
		UniqueClicks: 5,
		Events: []AdEventItem{
			{Time: "2024-01-01T00:00:00Z", EventType: "IMPRESSION", AdTitle: "Ad1", AdType: "BANNER"},
		},
		EventPagination: &EventPaginationInfo{
			Total: 1, Page: 1, PerPage: 10, TotalPages: 1,
		},
	}
	assert.Len(t, stats.Events, 1)
	assert.NotNil(t, stats.EventPagination)
	assert.Equal(t, int64(1), stats.EventPagination.Total)
}

func TestStatsResponse_WithoutEvents(t *testing.T) {
	stats := StatsResponse{
		TotalClicks:  0,
		UniqueClicks: 0,
		Events:       nil,
	}
	assert.Nil(t, stats.Events)
	assert.Nil(t, stats.EventPagination)
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