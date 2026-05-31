package analytics

import (
	"context"
	"testing"
	"database/sql"
	"time"

	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupWorkerDB(t *testing.T) *repository.Queries {
	_ = zap.ReplaceGlobals(zap.NewNop())
	_, queries := testutil.SetupTestDB(t)
	return queries
}

func createTestUser(t *testing.T, queries *repository.Queries, ctx context.Context) repository.User {
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Worker User",
		Email:    "worker@example.com",
		Password: sql.NullString{String: "password", Valid: true},
		Role:     "user",
	})
	require.NoError(t, err)
	return user
}

func createTestURL(t *testing.T, queries *repository.Queries, ctx context.Context, userID uuid.UUID) repository.Url {
	url, err := queries.CreateURL(ctx, repository.CreateURLParams{
		UserID:   userID,
		Slug:     "test",
		Original: "https://example.com",
		Custom:   false,
	})
	require.NoError(t, err)
	return url
}

func TestNewAnalyticsWorker_CreatesWorker(t *testing.T) {
	queries := setupWorkerDB(t)
	worker := NewAnalyticsWorker(queries, 100)
	assert.NotNil(t, worker)
	assert.NotNil(t, worker.jobs)
}

func TestAnalyticsWorker_Enqueue_ProcessesEvent(t *testing.T) {
	queries := setupWorkerDB(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)
	url := createTestURL(t, queries, ctx, user.ID)

	worker := NewAnalyticsWorker(queries, 100)

	event := ClickEvent{
		UrlID:    url.ID,
		IPHash:   "abc123",
		Country:  "US",
		City:     "New York",
		Device:   "mobile",
		Browser:  "Chrome",
		Referrer: "https://google.com",
	}

	worker.Enqueue(event)

	// Poll database to wait for the background worker to insert the click row
	var count int64
	var err error
	for i := 0; i < 20; i++ {
		count, err = queries.GetTotalClicksBySlug(ctx, "test")
		if err == nil && count > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestAnalyticsWorker_Enqueue_DropsWhenFull(t *testing.T) {
	queries := setupWorkerDB(t)
	worker := NewAnalyticsWorker(queries, 0)

	event := ClickEvent{
		UrlID:   uuid.New(),
		IPHash:  "abc123",
		Device:  "desktop",
		Browser: "Firefox",
	}

	worker.Enqueue(event)
}

func TestAnalyticsWorker_Enqueue_MultipleEvents(t *testing.T) {
	queries := setupWorkerDB(t)
	ctx := context.Background()
	user := createTestUser(t, queries, ctx)
	url := createTestURL(t, queries, ctx, user.ID)

	worker := NewAnalyticsWorker(queries, 100)

	for i := 0; i < 5; i++ {
		event := ClickEvent{
			UrlID:   url.ID,
			IPHash:  "hash",
			Country: "US",
			Device:  "mobile",
			Browser: "Chrome",
		}
		worker.Enqueue(event)
	}

	// Poll database to wait for background worker threads to write all 5 rows
	var count int64
	var err error
	for i := 0; i < 20; i++ {
		count, err = queries.GetTotalClicksBySlug(ctx, "test")
		if err == nil && count >= 5 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestClickEvent_Fields(t *testing.T) {
	urlID := uuid.New()
	event := ClickEvent{
		UrlID:    urlID,
		IPHash:   "hash123",
		Country:  "US",
		City:     "NYC",
		Device:   "desktop",
		Browser:  "Safari",
		Referrer: "https://example.com",
	}

	assert.Equal(t, urlID, event.UrlID)
	assert.Equal(t, "hash123", event.IPHash)
	assert.Equal(t, "US", event.Country)
	assert.Equal(t, "NYC", event.City)
	assert.Equal(t, "desktop", event.Device)
	assert.Equal(t, "Safari", event.Browser)
	assert.Equal(t, "https://example.com", event.Referrer)
}