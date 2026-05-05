package auth

import (
	"context"
	"fmt"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/response"
	"urlshortener/pkg/token"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo  repository.Querier
	cache cache.Cacher
	cfg   *config.Config
}

func NewAuthService(repo repository.Querier, cache cache.Cacher, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cache: cache, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, response.NewAppError(409, "Email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("Failed to hash password", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
	})
	if err != nil {
		zap.L().Error("Failed to create user in DB", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create user")
	}

	accessToken, err := token.IssueToken(user.ID.String(), s.cfg.JWTAccessSecret, "access", constants.AccessTokenTTL)
	if err != nil {
		zap.L().Error("Failed to issue access token", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	refreshToken, err := token.IssueToken(user.ID.String(), s.cfg.JWTRefreshSecret, "refresh", constants.RefreshTokenTTL)
	if err != nil {
		zap.L().Error("Failed to issue refresh token", zap.Error(err))
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

	accessToken, err := token.IssueToken(user.ID.String(), s.cfg.JWTAccessSecret, "access", constants.AccessTokenTTL)
	if err != nil {
		zap.L().Error("Failed to issue access token", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	refreshToken, err := token.IssueToken(user.ID.String(), s.cfg.JWTRefreshSecret, "refresh", constants.RefreshTokenTTL)
	if err != nil {
		zap.L().Error("Failed to issue refresh token", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	return dto.NewAuthResponse(user, accessToken, refreshToken), nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	claims, err := token.Validate(refreshToken, s.cfg.JWTRefreshSecret, "refresh")
	if err != nil {
		return nil, response.NewAppError(401, "Invalid refresh token")
	}

	blacklistKey := fmt.Sprintf("bl:%s", refreshToken)
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

	accessToken, err := token.IssueToken(user.ID.String(), s.cfg.JWTAccessSecret, "access", constants.AccessTokenTTL)
	if err != nil {
		zap.L().Error("Failed to issue access token during refresh", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		User:        dto.MapUserToResponse(user),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken != "" {
		claims, err := token.Validate(accessToken, s.cfg.JWTAccessSecret, "access")
		if err == nil {
			ttl := timeUntilExpiry(claims)
			if ttl > 0 {
				s.cache.Set(ctx, fmt.Sprintf("bl:%s", accessToken), "1", ttl)
			}
		}
	}

	if refreshToken != "" {
		claims, err := token.Validate(refreshToken, s.cfg.JWTRefreshSecret, "refresh")
		if err == nil {
			ttl := timeUntilExpiry(claims)
			if ttl > 0 {
				s.cache.Set(ctx, fmt.Sprintf("bl:%s", refreshToken), "1", ttl)
			}
		}
	}

	return nil
}

func timeUntilExpiry(claims *token.Claims) time.Duration {
	if claims.ExpiresAt == nil {
		return 0
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return 0
	}
	return ttl
}
