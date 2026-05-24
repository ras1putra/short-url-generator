package auth

import (
	"context"
	"database/sql"
	"fmt"

	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	"urlshortener/pkg/token"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db    *sql.DB
	repo  *repository.Queries
	cache cache.Cacher
	cfg   *config.Config
}

func NewAuthService(db *sql.DB, repo *repository.Queries, cache cache.Cacher, cfg *config.Config) *AuthService {
	return &AuthService{db: db, repo: repo, cache: cache, cfg: cfg}
}

func (s *AuthService) revokeToken(ctx context.Context, tokenStr, secret, tokenType string) {
	claims, err := token.Validate(tokenStr, secret, tokenType)
	if err != nil {
		return
	}
	ttl := timeUntilExpiry(claims)
	if ttl > 0 {
		s.cache.Set(ctx, fmt.Sprintf("%s%s", constants.RedisPrefixBlacklist, tokenStr), "1", ttl)
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, response.NewAppError(409, "Email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to hash password", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	role := req.Role
	if role == "" {
		role = constants.RoleUser
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}
	defer func() { _ = tx.Rollback() }()

	q := s.repo.WithTx(tx)

	user, err := q.CreateUser(ctx, repository.CreateUserParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
		Role:     role,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create user in DB", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create user")
	}

	if err := q.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: constants.DefaultBalance,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to create wallet for new user", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create wallet")
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	accessToken, refreshToken, err := issueTokens(user, s.cfg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to issue tokens", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	return dto.NewAuthResponse(user, accessToken, refreshToken), nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, response.NewAppError(401, "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, response.NewAppError(401, "Invalid credentials")
	}

	accessToken, refreshToken, err := issueTokens(user, s.cfg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to issue tokens", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	return dto.NewAuthResponse(user, accessToken, refreshToken), nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	claims, err := token.Validate(refreshToken, s.cfg.JWTRefreshSecret, constants.TokenTypeRefresh)
	if err != nil {
		return nil, response.NewAppError(401, "Invalid refresh token")
	}

	blacklistKey := fmt.Sprintf("%s%s", constants.RedisPrefixBlacklist, refreshToken)
	blacklisted, _ := s.cache.Exists(ctx, blacklistKey)
	if blacklisted {
		return nil, response.NewAppError(401, "Refresh token revoked")
	}

	parsedUserID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, response.NewAppError(401, "Invalid user ID in token")
	}

	user, err := s.repo.GetUserByID(ctx, parsedUserID)
	if err != nil {
		return nil, response.NewAppError(401, "User not found")
	}

	accessToken, err := token.IssueToken(user.ID.String(), user.Role, s.cfg.JWTAccessSecret, constants.TokenTypeAccess, constants.AccessTokenTTL)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to issue access token during refresh", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	return dto.NewAccessTokenResponse(user, accessToken), nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	s.revokeToken(ctx, accessToken, s.cfg.JWTAccessSecret, constants.TokenTypeAccess)
	s.revokeToken(ctx, refreshToken, s.cfg.JWTRefreshSecret, constants.TokenTypeRefresh)
	return nil
}

func (s *AuthService) UpgradeToAdvertiser(ctx context.Context, userID uuid.UUID, currentRole string) (*dto.AuthResponse, error) {
	if currentRole != constants.RoleUser {
		return nil, response.NewAppError(400, "Only users with role 'user' can be upgraded")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}
	defer func() { _ = tx.Rollback() }()

	q := s.repo.WithTx(tx)

	user, err := q.UpdateUserRole(ctx, repository.UpdateUserRoleParams{
		ID:   userID,
		Role: constants.RoleAdvertiser,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upgrade user role", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to upgrade role")
	}

	if err := q.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  userID,
		Balance: constants.DefaultBalance,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to create wallet", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create wallet")
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	accessToken, err := token.IssueToken(user.ID.String(), user.Role, s.cfg.JWTAccessSecret, constants.TokenTypeAccess, constants.AccessTokenTTL)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to issue access token", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	return dto.NewAccessTokenResponse(user, accessToken), nil
}
