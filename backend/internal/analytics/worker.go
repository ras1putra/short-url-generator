package analytics

import (
	"context"
	"database/sql"
	"time"

	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ClickEvent struct {
	UrlID    uuid.UUID
	IPHash   string
	Country  string
	City     string
	Device   string
	Browser  string
	Referrer string
}

type ClickSaver interface {
	SaveClick(ctx context.Context, arg repository.SaveClickParams) (repository.Click, error)
}

type AnalyticsWorker struct {
	jobs chan ClickEvent
	repo ClickSaver
}

func NewAnalyticsWorker(repo ClickSaver, bufferSize int) *AnalyticsWorker {
	w := &AnalyticsWorker{
		jobs: make(chan ClickEvent, bufferSize),
		repo: repo,
	}
	for i := 0; i < constants.AnalyticsWorkerCount; i++ {
		go w.process()
	}
	return w
}

func (w *AnalyticsWorker) process() {
	for event := range w.jobs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		_, err := w.repo.SaveClick(ctx, repository.SaveClickParams{
			UrlID:    event.UrlID,
			IpHash:   sql.NullString{String: event.IPHash, Valid: event.IPHash != ""},
			Country:  sql.NullString{String: event.Country, Valid: event.Country != ""},
			City:     sql.NullString{String: event.City, Valid: event.City != ""},
			Device:   sql.NullString{String: event.Device, Valid: event.Device != ""},
			Browser:  sql.NullString{String: event.Browser, Valid: event.Browser != ""},
			Referrer: sql.NullString{String: event.Referrer, Valid: event.Referrer != ""},
		})

		cancel()

		if err != nil {
			zap.L().Error("Failed to save analytics click", zap.Error(err))
		}
	}
}

func (w *AnalyticsWorker) Enqueue(event ClickEvent) {
	select {
	case w.jobs <- event:
	default:
		zap.L().Warn("Analytics channel full, dropping event")
	}
}
