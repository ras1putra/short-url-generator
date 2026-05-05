package links

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/links/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/response"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CountURLsByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) CreateURL(ctx context.Context, arg repository.CreateURLParams) (repository.Url, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.Url), args.Error(1)
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.User), args.Error(1)
}

func (m *MockQuerier) DeleteExpiredURLs(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockQuerier) DeleteURL(ctx context.Context, arg repository.DeleteURLParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) GetStatsBySlug(ctx context.Context, slug string) ([]repository.GetStatsBySlugRow, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).([]repository.GetStatsBySlugRow), args.Error(1)
}

func (m *MockQuerier) GetTotalClicksBySlug(ctx context.Context, slug string) (int64, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) GetURLBySlug(ctx context.Context, slug string) (repository.Url, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(repository.Url), args.Error(1)
}

func (m *MockQuerier) GetUniqueClicksBySlug(ctx context.Context, slug string) (int64, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) GetAggregateStatsByUser(ctx context.Context, userID uuid.UUID) ([]repository.GetAggregateStatsByUserRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]repository.GetAggregateStatsByUserRow), args.Error(1)
}

func (m *MockQuerier) GetTotalClicksByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) GetUniqueClicksByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) GetUserByEmail(ctx context.Context, email string) (repository.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(repository.User), args.Error(1)
}

func (m *MockQuerier) GetUserByID(ctx context.Context, id uuid.UUID) (repository.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.User), args.Error(1)
}

func (m *MockQuerier) ListURLsByUser(ctx context.Context, userID uuid.UUID) ([]repository.Url, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]repository.Url), args.Error(1)
}

func (m *MockQuerier) ListURLsByUserPaginated(ctx context.Context, arg repository.ListURLsByUserPaginatedParams) ([]repository.Url, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.Url), args.Error(1)
}

func (m *MockQuerier) SaveClick(ctx context.Context, arg repository.SaveClickParams) (repository.Click, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.Click), args.Error(1)
}

func (m *MockQuerier) UpdateURL(ctx context.Context, arg repository.UpdateURLParams) (repository.Url, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.Url), args.Error(1)
}

type MockCacher struct {
	mock.Mock
}

func (m *MockCacher) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockCacher) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacher) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacher) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockCacher) RateLimitIncrement(ctx context.Context, key string, ttl time.Duration) (int, error) {
	args := m.Called(ctx, key, ttl)
	return args.Get(0).(int), args.Error(1)
}

var testCfg = &config.Config{
	BaseURL:          "http://localhost:8080",
	JWTAccessSecret:  "test-secret",
	JWTRefreshSecret: "test-refresh-secret",
}

func newTestURLService(repo *MockQuerier, cache *MockCacher) *URLService {
	return NewURLService(repo, cache, testCfg)
}

func makeTestURL(userID uuid.UUID, slug string) repository.Url {
	return repository.Url{
		ID:        uuid.New(),
		UserID:    userID,
		Slug:      slug,
		Original:  "https://example.com",
		Custom:    false,
		ExpiresAt: sql.NullTime{Valid: false},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func makeTestUser() repository.User {
	return repository.User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashed",
		CreatedAt: time.Now(),
	}
}

func appErrCode(err error) int {
	var appErr *response.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return -1
}

func TestURLService_Create_SuccessAutoSlug(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "abc123")

	mockRepo.On("GetURLBySlug", ctx, mock.AnythingOfType("string")).Return(repository.Url{}, sql.ErrNoRows)
	mockRepo.On("CreateURL", ctx, mock.MatchedBy(func(arg repository.CreateURLParams) bool {
		return arg.Original == "https://example.com" && arg.UserID == userID && !arg.Custom
	})).Return(testURL, nil)

	req := dto.CreateURLRequest{
		URL: "https://example.com",
	}

	resp, err := svc.Create(ctx, userID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "abc123", resp.Slug)
	assert.Equal(t, "https://example.com", resp.Original)
	assert.Equal(t, "http://localhost:8080/abc123", resp.ShortURL)
	mockRepo.AssertExpectations(t)
}

func TestURLService_Create_SuccessCustomSlug(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "my-custom")
	testURL.Custom = true

	mockRepo.On("GetURLBySlug", ctx, "my-custom").Return(repository.Url{}, sql.ErrNoRows)
	mockRepo.On("CreateURL", ctx, mock.MatchedBy(func(arg repository.CreateURLParams) bool {
		return arg.Slug == "my-custom" && arg.Custom
	})).Return(testURL, nil)

	req := dto.CreateURLRequest{
		URL:        "https://example.com",
		CustomSlug: "my-custom",
	}

	resp, err := svc.Create(ctx, userID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "my-custom", resp.Slug)
	mockRepo.AssertExpectations(t)
}

func TestURLService_Create_ReservedSlug(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()

	req := dto.CreateURLRequest{
		URL:        "https://example.com",
		CustomSlug: "api",
	}

	resp, err := svc.Create(ctx, userID, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
}

func TestURLService_Create_TakenCustomSlug(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	existingURL := makeTestURL(uuid.New(), "taken-slug")

	mockRepo.On("GetURLBySlug", ctx, "taken-slug").Return(existingURL, nil)

	req := dto.CreateURLRequest{
		URL:        "https://example.com",
		CustomSlug: "taken-slug",
	}

	resp, err := svc.Create(ctx, userID, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Create_WithExpiry(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "exp123")
	testURL.ExpiresAt = sql.NullTime{Valid: true, Time: time.Now().Add(2 * time.Hour)}

	mockRepo.On("GetURLBySlug", ctx, mock.AnythingOfType("string")).Return(repository.Url{}, sql.ErrNoRows)
	mockRepo.On("CreateURL", ctx, mock.MatchedBy(func(arg repository.CreateURLParams) bool {
		return arg.Original == "https://example.com" && arg.ExpiresAt.Valid
	})).Return(testURL, nil)

	req := dto.CreateURLRequest{
		URL:          "https://example.com",
		ExpiresValue: 2,
		ExpiresUnit:  "hours",
	}

	resp, err := svc.Create(ctx, userID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "exp123", resp.Slug)
	assert.NotNil(t, resp.ExpiresAt)
	mockRepo.AssertExpectations(t)
}

func TestURLService_Create_SlugCheckDBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetURLBySlug", ctx, "my-custom").Return(repository.Url{}, errors.New("db error"))

	req := dto.CreateURLRequest{
		URL:        "https://example.com",
		CustomSlug: "my-custom",
	}

	resp, err := svc.Create(ctx, userID, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetBySlug_CacheHit(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	testURL := repository.Url{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Slug:      "cached123",
		Original:  "https://cached.com",
		Custom:    false,
		ExpiresAt: sql.NullTime{Valid: false},
		CreatedAt: now,
		UpdatedAt: now,
	}

	urlJSON, err := json.Marshal(testURL)
	require.NoError(t, err)
	mockCache.On("Get", ctx, "cached123").Return(string(urlJSON), nil)

	result, err := svc.GetBySlug(ctx, "cached123")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "cached123", result.Slug)
	assert.Equal(t, "https://cached.com", result.Original)
	mockCache.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetURLBySlug")
}

func TestURLService_GetBySlug_CacheMiss(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "dburl")

	mockCache.On("Get", ctx, "dburl").Return("", errors.New("cache miss"))
	mockRepo.On("GetURLBySlug", ctx, "dburl").Return(testURL, nil)
	mockCache.On("Set", ctx, "dburl", mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	result, err := svc.GetBySlug(ctx, "dburl")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "dburl", result.Slug)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestURLService_GetBySlug_NotFound(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	mockCache.On("Get", ctx, "notfound").Return("", errors.New("cache miss"))
	mockRepo.On("GetURLBySlug", ctx, "notfound").Return(repository.Url{}, sql.ErrNoRows)

	result, err := svc.GetBySlug(ctx, "notfound")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 404, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetBySlug_Expired(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	testURL := makeTestURL(uuid.New(), "expired")
	testURL.ExpiresAt = sql.NullTime{Valid: true, Time: time.Now().Add(-time.Hour)}

	mockCache.On("Get", ctx, "expired").Return("", errors.New("cache miss"))
	mockRepo.On("GetURLBySlug", ctx, "expired").Return(testURL, nil)

	result, err := svc.GetBySlug(ctx, "expired")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 410, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_ListByUser_Success(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	url1 := makeTestURL(userID, "abc1")
	url2 := makeTestURL(userID, "abc2")

	mockRepo.On("CountURLsByUser", ctx, userID).Return(int64(2), nil)
	mockRepo.On("ListURLsByUserPaginated", ctx, mock.MatchedBy(func(arg repository.ListURLsByUserPaginatedParams) bool {
		return arg.UserID == userID && arg.Limit == 10 && arg.Offset == 0
	})).Return([]repository.Url{url1, url2}, nil)

	result, err := svc.ListByUser(ctx, userID, 1, 10)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PerPage)
	assert.Equal(t, 1, result.TotalPages)
	assert.Len(t, result.Links, 2)
	assert.Equal(t, "abc1", result.Links[0].Slug)
	assert.Equal(t, "abc2", result.Links[1].Slug)
	mockRepo.AssertExpectations(t)
}

func TestURLService_ListByUser_Pagination(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	urls := make([]repository.Url, 5)
	for i := range urls {
		urls[i] = makeTestURL(userID, "slug")
	}

	mockRepo.On("CountURLsByUser", ctx, userID).Return(int64(15), nil)
	mockRepo.On("ListURLsByUserPaginated", ctx, mock.MatchedBy(func(arg repository.ListURLsByUserPaginatedParams) bool {
		return arg.UserID == userID && arg.Limit == 5 && arg.Offset == int32(10)
	})).Return(urls, nil)

	result, err := svc.ListByUser(ctx, userID, 3, 5)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(15), result.Total)
	assert.Equal(t, 3, result.Page)
	assert.Equal(t, 5, result.PerPage)
	assert.Equal(t, 3, result.TotalPages)
	assert.Len(t, result.Links, 5)
	mockRepo.AssertExpectations(t)
}

func TestURLService_ListByUser_CountError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	mockRepo.On("CountURLsByUser", ctx, userID).Return(int64(0), errors.New("db error"))

	result, err := svc.ListByUser(ctx, userID, 1, 10)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Delete_Success(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "del-me")

	mockRepo.On("GetURLBySlug", ctx, "del-me").Return(testURL, nil)
	mockRepo.On("DeleteURL", ctx, mock.MatchedBy(func(arg repository.DeleteURLParams) bool {
		return arg.ID == testURL.ID && arg.UserID == userID
	})).Return(nil)
	mockCache.On("Del", ctx, "del-me").Return(nil)

	err := svc.Delete(ctx, userID, "del-me")

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestURLService_Delete_NotFound(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	mockRepo.On("GetURLBySlug", ctx, "notfound").Return(repository.Url{}, sql.ErrNoRows)

	err := svc.Delete(ctx, userID, "notfound")

	require.Error(t, err)
	assert.Equal(t, 404, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Delete_Forbidden(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	ownerID := uuid.New()
	otherID := uuid.New()
	testURL := makeTestURL(ownerID, "owned-url")

	mockRepo.On("GetURLBySlug", ctx, "owned-url").Return(testURL, nil)

	err := svc.Delete(ctx, otherID, "owned-url")

	require.Error(t, err)
	assert.Equal(t, 403, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "get-me")

	mockRepo.On("GetURLBySlug", ctx, "get-me").Return(testURL, nil)

	resp, err := svc.GetByID(ctx, userID, "get-me")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "get-me", resp.Slug)
	assert.Equal(t, "https://example.com", resp.Original)
	assert.Equal(t, "http://localhost:8080/get-me", resp.ShortURL)
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	mockRepo.On("GetURLBySlug", ctx, "notfound").Return(repository.Url{}, sql.ErrNoRows)

	resp, err := svc.GetByID(ctx, uuid.New(), "notfound")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetByID_Forbidden(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	ownerID := uuid.New()
	otherID := uuid.New()
	testURL := makeTestURL(ownerID, "owned")

	mockRepo.On("GetURLBySlug", ctx, "owned").Return(testURL, nil)

	resp, err := svc.GetByID(ctx, otherID, "owned")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 403, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Update_SuccessChangeSlug(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "old-slug")

	mockRepo.On("GetURLBySlug", ctx, "old-slug").Return(testURL, nil)
	mockRepo.On("GetURLBySlug", ctx, "new-slug").Return(repository.Url{}, sql.ErrNoRows)
	mockRepo.On("UpdateURL", ctx, mock.MatchedBy(func(arg repository.UpdateURLParams) bool {
		return arg.Slug == "new-slug"
	})).Return(repository.Url{
		ID:        testURL.ID,
		UserID:    userID,
		Slug:      "new-slug",
		Original:  testURL.Original,
		Custom:    true,
		CreatedAt: testURL.CreatedAt,
		UpdatedAt: time.Now(),
	}, nil)
	mockCache.On("Del", ctx, "old-slug").Return(nil)

	req := dto.UpdateURLRequest{
		CustomSlug: "new-slug",
	}

	resp, err := svc.Update(ctx, userID, "old-slug", req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "new-slug", resp.Slug)
	mockRepo.AssertExpectations(t)
	mockCache.AssertCalled(t, "Del", ctx, "old-slug")
}

func TestURLService_Update_SuccessNoSlugChange(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "same-slug")

	mockRepo.On("GetURLBySlug", ctx, "same-slug").Return(testURL, nil)
	mockRepo.On("UpdateURL", ctx, mock.MatchedBy(func(arg repository.UpdateURLParams) bool {
		return arg.Slug == "same-slug"
	})).Return(testURL, nil)

	req := dto.UpdateURLRequest{}

	resp, err := svc.Update(ctx, userID, "same-slug", req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "same-slug", resp.Slug)
	mockRepo.AssertExpectations(t)
	mockCache.AssertNotCalled(t, "Del")
}

func TestURLService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	mockRepo.On("GetURLBySlug", ctx, "notfound").Return(repository.Url{}, sql.ErrNoRows)

	req := dto.UpdateURLRequest{CustomSlug: "new-slug"}

	resp, err := svc.Update(ctx, uuid.New(), "notfound", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Update_Forbidden(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	ownerID := uuid.New()
	otherID := uuid.New()
	testURL := makeTestURL(ownerID, "owned")

	mockRepo.On("GetURLBySlug", ctx, "owned").Return(testURL, nil)

	req := dto.UpdateURLRequest{CustomSlug: "new-slug"}

	resp, err := svc.Update(ctx, otherID, "owned", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 403, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Update_ReservedSlug(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "current-slug")

	mockRepo.On("GetURLBySlug", ctx, "current-slug").Return(testURL, nil)

	req := dto.UpdateURLRequest{CustomSlug: "api"}

	resp, err := svc.Update(ctx, userID, "current-slug", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
}

func TestURLService_Update_TakenSlug(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "current-slug")
	existingURL := makeTestURL(uuid.New(), "taken-slug")

	mockRepo.On("GetURLBySlug", ctx, "current-slug").Return(testURL, nil)
	mockRepo.On("GetURLBySlug", ctx, "taken-slug").Return(existingURL, nil)

	req := dto.UpdateURLRequest{CustomSlug: "taken-slug"}

	resp, err := svc.Update(ctx, userID, "current-slug", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetStats_Success(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()
	testURL := makeTestURL(userID, "stats-slug")

	mockRepo.On("GetURLBySlug", ctx, "stats-slug").Return(testURL, nil)
	mockRepo.On("GetTotalClicksBySlug", ctx, "stats-slug").Return(int64(150), nil)
	mockRepo.On("GetStatsBySlug", ctx, "stats-slug").Return([]repository.GetStatsBySlugRow{
		{
			Country:    sql.NullString{String: "US", Valid: true},
			Device:     sql.NullString{String: "Mobile", Valid: true},
			Browser:    sql.NullString{String: "Chrome", Valid: true},
			ClickDate:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			ClickCount: 100,
		},
		{
			Country:    sql.NullString{String: "", Valid: false},
			Device:     sql.NullString{String: "Desktop", Valid: true},
			Browser:    sql.NullString{String: "Firefox", Valid: true},
			ClickDate:  time.Date(2024, 6, 14, 0, 0, 0, 0, time.UTC),
			ClickCount: 50,
		},
	}, nil)
	mockRepo.On("GetUniqueClicksBySlug", ctx, "stats-slug").Return(int64(80), nil)

	resp, err := svc.GetStats(ctx, userID, "stats-slug")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(150), resp.TotalClicks)
	assert.Equal(t, int64(80), resp.UniqueClicks)
	assert.Len(t, resp.TopCountries, 2)
	assert.Len(t, resp.ClicksPerDay, 2)
	assert.Contains(t, resp.Browsers, "Chrome")
	assert.Contains(t, resp.Browsers, "Firefox")
	assert.Contains(t, resp.Devices, "Mobile")
	assert.Contains(t, resp.Devices, "Desktop")
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetStats_NotFound(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	mockRepo.On("GetURLBySlug", ctx, "notfound").Return(repository.Url{}, sql.ErrNoRows)

	resp, err := svc.GetStats(ctx, uuid.New(), "notfound")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetStats_Forbidden(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	ownerID := uuid.New()
	otherID := uuid.New()
	testURL := makeTestURL(ownerID, "owned-stats")

	mockRepo.On("GetURLBySlug", ctx, "owned-stats").Return(testURL, nil)

	resp, err := svc.GetStats(ctx, otherID, "owned-stats")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 403, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_StartExpiryCleaner_StopsOnContextCancellation(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.StartExpiryCleaner(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("StartExpiryCleaner did not exit after context cancellation")
	}
}

func TestURLService_Create_AllSlugsTaken(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()
	existingURL := makeTestURL(uuid.New(), "taken")

	mockRepo.On("GetURLBySlug", ctx, mock.AnythingOfType("string")).Return(existingURL, nil)

	req := dto.CreateURLRequest{URL: "https://example.com"}
	resp, err := svc.Create(ctx, userID, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
}

func TestURLService_Create_RandomSlugDBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetURLBySlug", ctx, mock.AnythingOfType("string")).Return(repository.Url{}, errors.New("db error"))

	req := dto.CreateURLRequest{URL: "https://example.com"}
	resp, err := svc.Create(ctx, userID, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
}

func TestURLService_GetBySlug_DBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	mockCache.On("Get", ctx, "error-slug").Return("", errors.New("cache miss"))
	mockRepo.On("GetURLBySlug", ctx, "error-slug").Return(repository.Url{}, errors.New("db error"))

	result, err := svc.GetBySlug(ctx, "error-slug")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestURLService_ListByUser_PaginationDBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("CountURLsByUser", ctx, userID).Return(int64(10), nil)
	mockRepo.On("ListURLsByUserPaginated", ctx, mock.Anything).Return([]repository.Url{}, errors.New("db error"))

	result, err := svc.ListByUser(ctx, userID, 1, 5)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Delete_GetDBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetURLBySlug", ctx, "error-slug").Return(repository.Url{}, errors.New("db error"))

	err := svc.Delete(ctx, userID, "error-slug")

	require.Error(t, err)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Delete_DeleteDBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()
	testURL := makeTestURL(userID, "del-me")

	mockRepo.On("GetURLBySlug", ctx, "del-me").Return(testURL, nil)
	mockRepo.On("DeleteURL", ctx, mock.MatchedBy(func(arg repository.DeleteURLParams) bool {
		return arg.ID == testURL.ID && arg.UserID == userID
	})).Return(errors.New("db error"))

	err := svc.Delete(ctx, userID, "del-me")

	require.Error(t, err)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetByID_DBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	mockRepo.On("GetURLBySlug", ctx, "error-slug").Return(repository.Url{}, errors.New("db error"))

	resp, err := svc.GetByID(ctx, uuid.New(), "error-slug")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Update_SlugCheckDBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()
	testURL := makeTestURL(userID, "current-slug")

	mockRepo.On("GetURLBySlug", ctx, "current-slug").Return(testURL, nil)
	mockRepo.On("GetURLBySlug", ctx, "new-slug").Return(repository.Url{}, errors.New("db error"))

	req := dto.UpdateURLRequest{CustomSlug: "new-slug"}

	resp, err := svc.Update(ctx, userID, "current-slug", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_Update_UpdateDBError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()
	testURL := makeTestURL(userID, "same-slug")

	mockRepo.On("GetURLBySlug", ctx, "same-slug").Return(testURL, nil)
	mockRepo.On("UpdateURL", ctx, mock.Anything).Return(repository.Url{}, errors.New("db error"))

	req := dto.UpdateURLRequest{}

	resp, err := svc.Update(ctx, userID, "same-slug", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetStats_DBErrorOnGetURL(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	mockRepo.On("GetURLBySlug", ctx, "error-slug").Return(repository.Url{}, errors.New("db error"))

	resp, err := svc.GetStats(ctx, uuid.New(), "error-slug")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetStats_TotalClicksError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()
	testURL := makeTestURL(userID, "stats-slug")

	mockRepo.On("GetURLBySlug", ctx, "stats-slug").Return(testURL, nil)
	mockRepo.On("GetTotalClicksBySlug", ctx, "stats-slug").Return(int64(0), errors.New("db error"))

	resp, err := svc.GetStats(ctx, userID, "stats-slug")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetStats_StatsBySlugError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()
	testURL := makeTestURL(userID, "stats-slug")

	mockRepo.On("GetURLBySlug", ctx, "stats-slug").Return(testURL, nil)
	mockRepo.On("GetTotalClicksBySlug", ctx, "stats-slug").Return(int64(100), nil)
	mockRepo.On("GetStatsBySlug", ctx, "stats-slug").Return([]repository.GetStatsBySlugRow{}, errors.New("db error"))

	resp, err := svc.GetStats(ctx, userID, "stats-slug")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetStats_UniqueClicksErrorFallsBack(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()
	userID := uuid.New()
	testURL := makeTestURL(userID, "stats-slug")

	mockRepo.On("GetURLBySlug", ctx, "stats-slug").Return(testURL, nil)
	mockRepo.On("GetTotalClicksBySlug", ctx, "stats-slug").Return(int64(100), nil)
	mockRepo.On("GetStatsBySlug", ctx, "stats-slug").Return([]repository.GetStatsBySlugRow{}, nil)
	mockRepo.On("GetUniqueClicksBySlug", ctx, "stats-slug").Return(int64(0), errors.New("db error"))

	resp, err := svc.GetStats(ctx, userID, "stats-slug")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(100), resp.UniqueClicks)
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetAggregateStats_Success(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetTotalClicksByUser", ctx, userID).Return(int64(300), nil)
	mockRepo.On("GetAggregateStatsByUser", ctx, userID).Return([]repository.GetAggregateStatsByUserRow{
		{
			Country:    sql.NullString{String: "US", Valid: true},
			Device:     sql.NullString{String: "Mobile", Valid: true},
			Browser:    sql.NullString{String: "Chrome", Valid: true},
			ClickDate:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			ClickCount: 200,
		},
		{
			Country:    sql.NullString{String: "ID", Valid: true},
			Device:     sql.NullString{String: "Desktop", Valid: true},
			Browser:    sql.NullString{String: "Firefox", Valid: true},
			ClickDate:  time.Date(2024, 6, 14, 0, 0, 0, 0, time.UTC),
			ClickCount: 100,
		},
	}, nil)
	mockRepo.On("GetUniqueClicksByUser", ctx, userID).Return(int64(180), nil)

	resp, err := svc.GetAggregateStats(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(300), resp.TotalClicks)
	assert.Equal(t, int64(180), resp.UniqueClicks)
	assert.Len(t, resp.TopCountries, 2)
	assert.Contains(t, resp.Browsers, "Chrome")
	assert.Contains(t, resp.Devices, "Mobile")
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetAggregateStats_TotalClicksError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetTotalClicksByUser", ctx, userID).Return(int64(0), errors.New("db error"))

	resp, err := svc.GetAggregateStats(ctx, userID)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetAggregateStats_StatsError(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetTotalClicksByUser", ctx, userID).Return(int64(100), nil)
	mockRepo.On("GetAggregateStatsByUser", ctx, userID).Return([]repository.GetAggregateStatsByUserRow{}, errors.New("db error"))

	resp, err := svc.GetAggregateStats(ctx, userID)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetAggregateStats_UniqueClicksErrorFallsBack(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetTotalClicksByUser", ctx, userID).Return(int64(100), nil)
	mockRepo.On("GetAggregateStatsByUser", ctx, userID).Return([]repository.GetAggregateStatsByUserRow{}, nil)
	mockRepo.On("GetUniqueClicksByUser", ctx, userID).Return(int64(0), errors.New("db error"))

	resp, err := svc.GetAggregateStats(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(100), resp.UniqueClicks)
	mockRepo.AssertExpectations(t)
}

func TestURLService_GetAggregateStats_UnknownCountry(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestURLService(mockRepo, mockCache)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetTotalClicksByUser", ctx, userID).Return(int64(50), nil)
	mockRepo.On("GetAggregateStatsByUser", ctx, userID).Return([]repository.GetAggregateStatsByUserRow{
		{
			Country:    sql.NullString{String: "", Valid: false},
			Device:     sql.NullString{String: "Bot", Valid: true},
			Browser:    sql.NullString{String: "", Valid: false},
			ClickDate:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			ClickCount: 50,
		},
	}, nil)
	mockRepo.On("GetUniqueClicksByUser", ctx, userID).Return(int64(25), nil)

	resp, err := svc.GetAggregateStats(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(50), resp.TotalClicks)
	assert.Equal(t, int64(25), resp.UniqueClicks)
	found := false
	for _, c := range resp.TopCountries {
		if c.Country == "Unknown" {
			found = true
			assert.Equal(t, int64(50), c.Count)
		}
	}
	assert.True(t, found, "Expected Unknown country entry")
	mockRepo.AssertExpectations(t)
}