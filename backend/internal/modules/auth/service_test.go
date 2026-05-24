package auth

import (
	"context"
	"errors"
	"testing"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth/dto"
	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/response"
	"urlshortener/pkg/token"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newTestAuthService(t *testing.T) (*AuthService, repository.Querier) {
	db, queries := testutil.SetupTestDB(t)

	cfg := &config.Config{
		BaseURL:          "http://localhost:8080",
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
	}

	fakeCache := testutil.NewFakeCacher()
	return NewAuthService(db, queries, fakeCache, cfg), queries
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
	svc, queries := newTestAuthService(t)
	ctx := context.Background()

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
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "Test User", resp.User.Name)
	assert.Equal(t, "user", resp.User.Role)

	// Verify the user actually exists in the database!
	parsedID, err := uuid.Parse(resp.User.ID)
	require.NoError(t, err)

	dbUser, err := queries.GetUserByID(ctx, parsedID)
	require.NoError(t, err)
	assert.Equal(t, "Test User", dbUser.Name)
	assert.Equal(t, "test@example.com", dbUser.Email)

	// Verify that their default wallet was automatically created under the transaction!
	wallet, err := queries.GetWalletByUserID(ctx, parsedID)
	require.NoError(t, err)
	assert.Equal(t, constants.DefaultBalance, wallet.Balance)
}

func TestRegister_EmailTaken(t *testing.T) {
	svc, queries := newTestAuthService(t)
	ctx := context.Background()

	// Pre-create the user in the database
	_, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Existing User",
		Email:    "test@example.com",
		Password: hashPassword("password123"),
		Role:     "user",
	})
	require.NoError(t, err)

	req := dto.RegisterRequest{
		Name:     "New User",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := svc.Register(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 409, appErrCode(err))
}

func TestLogin_Success(t *testing.T) {
	svc, queries := newTestAuthService(t)
	ctx := context.Background()

	// Pre-create the user in the database
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Login User",
		Email:    "login@example.com",
		Password: hashPassword("correct-password"),
		Role:     "user",
	})
	require.NoError(t, err)

	req := dto.LoginRequest{
		Email:    "login@example.com",
		Password: "correct-password",
	}

	resp, err := svc.Login(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, user.ID.String(), resp.User.ID)
	assert.Equal(t, "login@example.com", resp.User.Email)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, queries := newTestAuthService(t)
	ctx := context.Background()

	_, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Login User",
		Email:    "login@example.com",
		Password: hashPassword("correct-password"),
		Role:     "user",
	})
	require.NoError(t, err)

	req := dto.LoginRequest{
		Email:    "login@example.com",
		Password: "wrong-password",
	}

	resp, err := svc.Login(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()

	req := dto.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	resp, err := svc.Login(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
}

func TestRefreshToken_Success(t *testing.T) {
	svc, queries := newTestAuthService(t)
	ctx := context.Background()

	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Refresh User",
		Email:    "refresh@example.com",
		Password: hashPassword("password123"),
		Role:     "user",
	})
	require.NoError(t, err)

	refreshToken, err := token.IssueToken(user.ID.String(), user.Role, "test-refresh-secret", "refresh", constants.RefreshTokenTTL)
	require.NoError(t, err)

	resp, err := svc.RefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, user.ID.String(), resp.User.ID)
	assert.Equal(t, "refresh@example.com", resp.User.Email)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()

	resp, err := svc.RefreshToken(ctx, "invalid-token-string")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()

	nonExistentUserID := uuid.New()
	refreshToken, err := token.IssueToken(nonExistentUserID.String(), "user", "test-refresh-secret", "refresh", constants.RefreshTokenTTL)
	require.NoError(t, err)

	resp, err := svc.RefreshToken(ctx, refreshToken)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
}

func TestRefreshToken_WrongTokenType(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()

	userID := uuid.New()
	accessToken, err := token.IssueToken(userID.String(), "user", "test-access-secret", "access", constants.AccessTokenTTL)
	require.NoError(t, err)

	resp, err := svc.RefreshToken(ctx, accessToken)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 401, appErrCode(err))
}

func TestLogout_SuccessWithBothTokens(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()

	userID := uuid.New()
	accessToken, err := token.IssueToken(userID.String(), "user", "test-access-secret", "access", constants.AccessTokenTTL)
	require.NoError(t, err)
	refreshToken, err := token.IssueToken(userID.String(), "user", "test-refresh-secret", "refresh", constants.RefreshTokenTTL)
	require.NoError(t, err)

	err = svc.Logout(ctx, accessToken, refreshToken)
	require.NoError(t, err)
}

func TestUpgradeToAdvertiser_Success(t *testing.T) {
	svc, queries := newTestAuthService(t)
	ctx := context.Background()

	// Pre-create standard user and standard wallet
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Advertiser Candidate",
		Email:    "candidate@example.com",
		Password: hashPassword("password123"),
		Role:     "user",
	})
	require.NoError(t, err)

	err = queries.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: "0.00",
	})
	require.NoError(t, err)

	resp, err := svc.UpgradeToAdvertiser(ctx, user.ID, "user")
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, user.ID.String(), resp.User.ID)
	assert.Equal(t, "advertiser", resp.User.Role)

	// Verify that their role is actually updated in the database!
	dbUser, err := queries.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "advertiser", dbUser.Role)
}

func TestUpgradeToAdvertiser_AlreadyAdvertiser(t *testing.T) {
	svc, _ := newTestAuthService(t)
	ctx := context.Background()

	resp, err := svc.UpgradeToAdvertiser(ctx, uuid.New(), "advertiser")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 400, appErrCode(err))
}