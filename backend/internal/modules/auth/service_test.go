package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/response"
	"urlshortener/pkg/token"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
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

func newTestAuthService(repo *MockQuerier, cache *MockCacher) *AuthService {
	cfg := &config.Config{
		BaseURL:          "http://localhost:8080",
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
	}
	return NewAuthService(repo, cache, cfg)
}

func makeTestUser() repository.User {
	return repository.User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "",
		CreatedAt: time.Now(),
	}
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func appErrCode(err error) int {
	var appErr *response.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return -1
}

func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	existingErr := errors.New("user not found")
	mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(repository.User{}, existingErr)

	createdUser := makeTestUser()
	mockRepo.On("CreateUser", ctx, mock.MatchedBy(func(arg repository.CreateUserParams) bool {
		return arg.Email == "test@example.com" && arg.Name == "Test User"
	})).Return(createdUser, nil)

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := svc.Register(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, createdUser.ID.String(), resp.User.ID)
	assert.Equal(t, createdUser.Email, resp.User.Email)
	assert.Equal(t, createdUser.Name, resp.User.Name)
	mockRepo.AssertExpectations(t)
}

func TestRegister_EmailTaken(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	existingUser := makeTestUser()
	mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(existingUser, nil)

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := svc.Register(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestRegister_CreateUserFails(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(repository.User{}, errors.New("not found"))
	mockRepo.On("CreateUser", ctx, mock.Anything).Return(repository.User{}, errors.New("db error"))

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := svc.Register(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	hashedPassword := hashPassword("password123")
	user := makeTestUser()
	user.Password = hashedPassword
	mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := svc.Login(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, user.ID.String(), resp.User.ID)
	assert.Equal(t, user.Email, resp.User.Email)
	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	hashedPassword := hashPassword("correct-password")
	user := makeTestUser()
	user.Password = hashedPassword
	mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong-password",
	}

	resp, err := svc.Login(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	mockRepo.On("GetUserByEmail", ctx, "notfound@example.com").Return(repository.User{}, errors.New("not found"))

	req := dto.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	resp, err := svc.Login(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_Success(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	user := makeTestUser()
	refreshToken, err := token.IssueToken(user.ID.String(), "test-refresh-secret", "refresh", constants.RefreshTokenTTL)
	require.NoError(t, err)

	mockCache.On("Exists", ctx, mock.MatchedBy(func(key string) bool {
		return len(key) > 3 && key[:3] == "bl:"
	})).Return(false, nil)

	mockRepo.On("GetUserByID", ctx, user.ID).Return(user, nil)

	resp, err := svc.RefreshToken(ctx, refreshToken)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, user.ID.String(), resp.User.ID)
	assert.Equal(t, user.Email, resp.User.Email)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	resp, err := svc.RefreshToken(ctx, "invalid-token-string")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
}

func TestRefreshToken_RevokedToken(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	user := makeTestUser()
	refreshToken, err := token.IssueToken(user.ID.String(), "test-refresh-secret", "refresh", constants.RefreshTokenTTL)
	require.NoError(t, err)

	mockCache.On("Exists", ctx, mock.MatchedBy(func(key string) bool {
		return len(key) > 3 && key[:3] == "bl:"
	})).Return(true, nil)

	resp, err := svc.RefreshToken(ctx, refreshToken)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	user := makeTestUser()
	refreshToken, err := token.IssueToken(user.ID.String(), "test-refresh-secret", "refresh", constants.RefreshTokenTTL)
	require.NoError(t, err)

	mockCache.On("Exists", ctx, mock.Anything).Return(false, nil)
	mockRepo.On("GetUserByID", ctx, user.ID).Return(repository.User{}, errors.New("not found"))

	resp, err := svc.RefreshToken(ctx, refreshToken)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_WrongTokenType(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	user := makeTestUser()
	accessToken, err := token.IssueToken(user.ID.String(), "test-access-secret", "access", constants.AccessTokenTTL)
	require.NoError(t, err)

	resp, err := svc.RefreshToken(ctx, accessToken)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
}

func TestLogout_SuccessWithBothTokens(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	user := makeTestUser()
	accessToken, err := token.IssueToken(user.ID.String(), "test-access-secret", "access", constants.AccessTokenTTL)
	require.NoError(t, err)
	refreshToken, err := token.IssueToken(user.ID.String(), "test-refresh-secret", "refresh", constants.RefreshTokenTTL)
	require.NoError(t, err)

	mockCache.On("Set", ctx, mock.MatchedBy(func(key string) bool {
		return len(key) > 3 && key[:3] == "bl:"
	}), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	err = svc.Logout(ctx, accessToken, refreshToken)

	require.NoError(t, err)
	mockCache.AssertCalled(t, "Set", ctx, mock.Anything, mock.Anything, mock.AnythingOfType("time.Duration"))
	mockCache.AssertNumberOfCalls(t, "Set", 2)
}

func TestLogout_EmptyTokens(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	err := svc.Logout(ctx, "", "")

	require.NoError(t, err)
	mockCache.AssertNotCalled(t, "Set")
}

func TestLogout_ExpiredTokens(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	claims := &token.Claims{
		UserID:    uuid.New().String(),
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredTokenStr, err := expiredToken.SignedString([]byte("test-access-secret"))
	require.NoError(t, err)

	err = svc.Logout(ctx, expiredTokenStr, "")

	require.NoError(t, err)
	mockCache.AssertNotCalled(t, "Set")
}

func TestLogout_OnlyAccessToken(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	user := makeTestUser()
	accessToken, err := token.IssueToken(user.ID.String(), "test-access-secret", "access", constants.AccessTokenTTL)
	require.NoError(t, err)

	mockCache.On("Set", ctx, mock.MatchedBy(func(key string) bool {
		return len(key) > 3 && key[:3] == "bl:"
	}), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	err = svc.Logout(ctx, accessToken, "")

	require.NoError(t, err)
	mockCache.AssertNumberOfCalls(t, "Set", 1)
}

func TestTimeUntilExpiry_NilExpiresAt(t *testing.T) {
	claims := &token.Claims{
		UserID:    "user-123",
		TokenType: "access",
	}

	result := timeUntilExpiry(claims)
	assert.Equal(t, time.Duration(0), result)
}

func TestTimeUntilExpiry_PastExpiresAt(t *testing.T) {
	claims := &token.Claims{
		UserID:    "user-123",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}

	result := timeUntilExpiry(claims)
	assert.Equal(t, time.Duration(0), result)
}

func TestTimeUntilExpiry_FutureExpiresAt(t *testing.T) {
	claims := &token.Claims{
		UserID:    "user-123",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}

	result := timeUntilExpiry(claims)
	assert.True(t, result > 0, "expected positive duration for future expiry")
	assert.True(t, result <= 30*time.Minute, "expected duration <= 30 minutes")
}

func TestRefreshToken_InvalidUserID(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockCache := new(MockCacher)
	svc := newTestAuthService(mockRepo, mockCache)
	ctx := context.Background()

	refreshToken, err := token.IssueToken("not-a-uuid", "test-refresh-secret", "refresh", constants.RefreshTokenTTL)
	require.NoError(t, err)

	mockCache.On("Exists", ctx, mock.MatchedBy(func(key string) bool {
		return len(key) > 3 && key[:3] == "bl:"
	})).Return(false, nil)

	resp, err := svc.RefreshToken(ctx, refreshToken)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
	mockCache.AssertExpectations(t)
}