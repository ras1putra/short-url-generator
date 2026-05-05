package analytics

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"urlshortener/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClickSaver struct {
	mock.Mock
	saved chan struct{}
}

func (m *MockClickSaver) SaveClick(ctx context.Context, arg repository.SaveClickParams) (repository.Click, error) {
	args := m.Called(ctx, arg)
	if m.saved != nil {
		m.saved <- struct{}{}
	}
	return args.Get(0).(repository.Click), args.Error(1)
}

func TestNewAnalyticsWorker_CreatesWorker(t *testing.T) {
	mockSaver := new(MockClickSaver)
	worker := NewAnalyticsWorker(mockSaver, 100)
	assert.NotNil(t, worker)
	assert.NotNil(t, worker.jobs)
}

func TestAnalyticsWorker_Enqueue_ProcessesEvent(t *testing.T) {
	mockSaver := new(MockClickSaver)
	mockSaver.saved = make(chan struct{}, 100)

	worker := NewAnalyticsWorker(mockSaver, 100)

	urlID := uuid.New()
	event := ClickEvent{
		UrlID:    urlID,
		IPHash:   "abc123",
		Country:  "US",
		City:     "New York",
		Device:   "mobile",
		Browser:  "Chrome",
		Referrer: "https://google.com",
	}

	mockSaver.On("SaveClick", mock.Anything, mock.MatchedBy(func(arg repository.SaveClickParams) bool {
		return arg.UrlID == urlID
	})).Return(repository.Click{}, nil)

	worker.Enqueue(event)

	select {
	case <-mockSaver.saved:
		mockSaver.AssertCalled(t, "SaveClick", mock.Anything, mock.Anything)
	case <-time.After(3 * time.Second):
		t.Fatal("SaveClick was not called within timeout")
	}
}

func TestAnalyticsWorker_Enqueue_DropsWhenFull(t *testing.T) {
	mockSaver := new(MockClickSaver)
	worker := NewAnalyticsWorker(mockSaver, 0)

	event := ClickEvent{
		UrlID:   uuid.New(),
		IPHash:  "abc123",
		Device:  "desktop",
		Browser: "Firefox",
	}

	worker.Enqueue(event)
}

func TestAnalyticsWorker_Enqueue_MultipleEvents(t *testing.T) {
	mockSaver := new(MockClickSaver)
	mockSaver.saved = make(chan struct{}, 100)

	worker := NewAnalyticsWorker(mockSaver, 100)

	mockSaver.On("SaveClick", mock.Anything, mock.Anything).Return(repository.Click{}, nil)

	for i := 0; i < 5; i++ {
		event := ClickEvent{
			UrlID:   uuid.New(),
			IPHash:  "hash",
			Country: "US",
			Device:  "mobile",
			Browser: "Chrome",
		}
		worker.Enqueue(event)
	}

	received := 0
	timeout := time.After(3 * time.Second)
	for received < 5 {
		select {
		case <-mockSaver.saved:
			received++
		case <-timeout:
			t.Fatalf("Only received %d out of 5 events", received)
		}
	}

	mockSaver.AssertNumberOfCalls(t, "SaveClick", 5)
}

func TestAnalyticsWorker_Process_SaveClickError(t *testing.T) {
	mockSaver := new(MockClickSaver)
	mockSaver.saved = make(chan struct{}, 100)

	worker := NewAnalyticsWorker(mockSaver, 100)

	urlID := uuid.New()
	event := ClickEvent{
		UrlID:   urlID,
		IPHash:  "abc",
		Country: "US",
	}

	mockSaver.On("SaveClick", mock.Anything, mock.Anything).Return(repository.Click{}, sql.ErrConnDone)

	worker.Enqueue(event)

	select {
	case <-mockSaver.saved:
		mockSaver.AssertCalled(t, "SaveClick", mock.Anything, mock.Anything)
	case <-time.After(2 * time.Second):
		t.Fatal("SaveClick was not called within timeout")
	}
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