package oauth

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	authdto "urlshortener/internal/modules/auth/dto"
	oauthdto "urlshortener/internal/modules/oauth/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	"urlshortener/pkg/token"

	"go.uber.org/zap"
)

const (
	googleAuthURL    = "https://accounts.google.com/o/oauth2/v2/auth"
	oauthStatePrefix = "oauth_state:"
	oauthStateTTL    = 10 * time.Minute
)

type OAuthService struct {
	db    *sql.DB
	repo  *repository.Queries
	cache cache.Cacher
	cfg   *config.Config
	http  *http.Client
}

func NewOAuthService(db *sql.DB, repo *repository.Queries, cache cache.Cacher, cfg *config.Config) *OAuthService {
	return &OAuthService{
		db:    db,
		repo:  repo,
		cache: cache,
		cfg:   cfg,
		http:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *OAuthService) GetLoginURL(intent string) (string, error) {
	state, err := GenerateState()
	if err != nil {
		return "", err
	}

	storedIntent := intent
	if storedIntent == "" {
		storedIntent = "default"
	}

	if err := s.cache.Set(context.Background(), oauthStatePrefix+state, storedIntent, oauthStateTTL); err != nil {
		return "", err
	}

	v := url.Values{}
	v.Set("client_id", s.cfg.GoogleClientID)
	v.Set("redirect_uri", s.cfg.GoogleRedirectURL)
	v.Set("response_type", "code")
	v.Set("scope", "openid email profile")
	v.Set("state", state)
	v.Set("access_type", "online")

	return googleAuthURL + "?" + v.Encode(), nil
}

func (s *OAuthService) consumeState(ctx context.Context, state string) (string, error) {
	key := oauthStatePrefix + state
	intent, err := s.cache.Get(ctx, key)
	if err != nil || intent == "" {
		return "", response.NewAppError(400, "Invalid or expired state parameter")
	}
	s.cache.Del(ctx, key)
	return intent, nil
}

func (s *OAuthService) findOrCreateUser(ctx context.Context, info *oauthdto.GoogleUserInfo) (*repository.User, error) {
	existing, err := s.repo.GetOAuthAccountByProvider(ctx, repository.GetOAuthAccountByProviderParams{
		Provider:   "google",
		ProviderID: info.ID,
	})
	if err == nil {
		user, err := s.repo.GetUserByID(ctx, existing.UserID)
		if err != nil {
			if err != sql.ErrNoRows {
				logger.Ctx(ctx).Error("Database error during OAuth GetUserByID lookup", zap.Error(err))
				return nil, response.NewAppError(500, "Internal server error")
			}
			return nil, response.NewAppError(500, "Failed to find linked user")
		}
		return &user, nil
	}
	if err != sql.ErrNoRows {
		logger.Ctx(ctx).Error("Database error during GetOAuthAccountByProvider lookup", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	match, err := s.repo.GetUserByEmail(ctx, info.Email)
	if err == nil {
		if err := s.repo.CreateOAuthAccount(ctx, repository.CreateOAuthAccountParams{
			UserID:     match.ID,
			Provider:   "google",
			ProviderID: info.ID,
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to link OAuth account to existing user", zap.Error(err))
			return nil, response.NewAppError(500, "Failed to link OAuth account")
		}
		return &match, nil
	}
	if err != sql.ErrNoRows {
		logger.Ctx(ctx).Error("Database error during GetUserByEmail lookup in OAuth findOrCreateUser", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	name := info.Name
	if name == "" {
		name = info.Email
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}
	defer func() { _ = tx.Rollback() }()

	q := s.repo.WithTx(tx)

	user, err := q.CreateUser(ctx, repository.CreateUserParams{
		Name:     name,
		Email:    info.Email,
		Password: sql.NullString{Valid: false},
		Role:     constants.RoleUser,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create OAuth user", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create user")
	}

	if err := q.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: constants.DefaultBalance,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to create wallet for OAuth user", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create wallet")
	}

	if err := q.CreateOAuthAccount(ctx, repository.CreateOAuthAccountParams{
		UserID:     user.ID,
		Provider:   "google",
		ProviderID: info.ID,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to create OAuth account", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create OAuth account")
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	return &user, nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, code, state string) (*authdto.AuthResponse, string, error) {
	intent, err := s.consumeState(ctx, state)
	if err != nil {
		return nil, "", err
	}

	info, err := ExchangeGoogleCode(ctx, s.http, s.cfg.GoogleClientID, s.cfg.GoogleClientSecret, s.cfg.GoogleRedirectURL, code)
	if err != nil {
		logger.Ctx(ctx).Error("Google OAuth exchange failed", zap.Error(err))
		return nil, "", response.NewAppError(401, "Google authentication failed")
	}

	user, err := s.findOrCreateUser(ctx, info)
	if err != nil {
		return nil, "", err
	}

	accessToken, err := token.IssueToken(user.ID.String(), user.Role, s.cfg.JWTAccessSecret, constants.TokenTypeAccess, constants.AccessTokenTTL)
	if err != nil {
		return nil, "", response.NewAppError(500, "Failed to generate token")
	}

	refreshToken, err := token.IssueToken(user.ID.String(), user.Role, s.cfg.JWTRefreshSecret, constants.TokenTypeRefresh, constants.RefreshTokenTTL)
	if err != nil {
		return nil, "", response.NewAppError(500, "Failed to generate token")
	}

	return authdto.NewAuthResponse(*user, accessToken, refreshToken), intent, nil
}
