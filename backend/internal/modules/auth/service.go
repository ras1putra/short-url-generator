package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	"urlshortener/internal/mailer"
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
	db     *sql.DB
	repo   *repository.Queries
	cache  cache.Cacher
	cfg    *config.Config
	mailer *mailer.Mailer
}

func NewAuthService(db *sql.DB, repo *repository.Queries, cache cache.Cacher, cfg *config.Config, mailer *mailer.Mailer) *AuthService {
	return &AuthService{db: db, repo: repo, cache: cache, cfg: cfg, mailer: mailer}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (repository.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, response.NewAppError(409, "Email already in use")
	}
	if err != sql.ErrNoRows {
		logger.Ctx(ctx).Error("Database error checking user existence during registration", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to hash password", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	role := constants.RoleUser

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
		Password: sql.NullString{String: string(hash), Valid: true},
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

	tokenStr, err := generateToken()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to generate verification token", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create verification token")
	}

	_, err = s.repo.UpdateUserEmailVerificationToken(ctx, repository.UpdateUserEmailVerificationTokenParams{
		ID:                     user.ID,
		EmailVerificationToken: sql.NullString{String: tokenStr, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store verification token", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	if err := s.mailer.SendVerificationEmail(user.Email, user.Name, tokenStr); err != nil {
		logger.Ctx(ctx).Error("Failed to send verification email", zap.Error(err))
	} else {
		logger.Ctx(ctx).Info("Verification email sent",
			zap.String("email", user.Email),
		)
	}

	return dto.NewAuthResponse(user, "", ""), nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, response.NewAppError(401, "Invalid credentials")
		}
		logger.Ctx(ctx).Error("Database error during login user lookup", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	if !user.Password.Valid {
		return nil, response.NewAppError(401, "This account uses Google sign-in. Please sign in with Google.")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(req.Password)); err != nil {
		return nil, response.NewAppError(401, "Invalid credentials")
	}

	if !user.EmailVerified {
		return nil, response.NewAppErrorWithCode(403, "Please verify your email before signing in", "EMAIL_NOT_VERIFIED")
	}

	accessToken, refreshToken, err := issueTokens(user, s.cfg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to issue tokens", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to generate token")
	}

	return dto.NewAuthResponse(user, accessToken, refreshToken), nil
}

func (s *AuthService) SendVerification(ctx context.Context, req dto.SendVerificationRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	if user.EmailVerified {
		return nil
	}

	if user.EmailVerificationSentAt.Valid {
		elapsed := time.Since(user.EmailVerificationSentAt.Time)
		if elapsed < constants.EmailVerificationCooldown {
			remaining := constants.EmailVerificationCooldown - elapsed
			return response.NewAppError(429, fmt.Sprintf("Please wait %d seconds before requesting a new verification email", int(remaining.Seconds())+1))
		}
	}

	tokenStr, err := generateToken()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to generate verification token", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	_, err = s.repo.UpdateUserEmailVerificationToken(ctx, repository.UpdateUserEmailVerificationTokenParams{
		ID:                     user.ID,
		EmailVerificationToken: sql.NullString{String: tokenStr, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store verification token", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	if err := s.mailer.SendVerificationEmail(user.Email, user.Name, tokenStr); err != nil {
		logger.Ctx(ctx).Error("Failed to send verification email", zap.Error(err))
		return response.NewAppError(500, "Failed to send verification email")
	}

	logger.Ctx(ctx).Info("Verification email sent",
		zap.String("email", user.Email),
	)

	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, tokenStr string) error {
	user, err := s.repo.GetUserByEmailVerificationToken(ctx, sql.NullString{String: tokenStr, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NewAppError(400, "Invalid verification code, or your code has been used")
		}
		logger.Ctx(ctx).Error("Database error during verification token lookup", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	if user.EmailVerified {
		return nil
	}

	if user.EmailVerificationSentAt.Valid {
		elapsed := time.Since(user.EmailVerificationSentAt.Time)
		if elapsed > constants.EmailVerificationTokenTTL {
			return response.NewAppError(400, "Verification code has expired. Request a new one.")
		}
	}

	_, err = s.repo.UpdateUserEmailVerified(ctx, user.ID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update email verified", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		logger.Ctx(ctx).Error("Database error during forgot password user lookup", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	if user.PasswordResetSentAt.Valid {
		elapsed := time.Since(user.PasswordResetSentAt.Time)
		if elapsed < constants.PasswordResetCooldown {
			remaining := constants.PasswordResetCooldown - elapsed
			return response.NewAppError(429, fmt.Sprintf("Please wait %d seconds before requesting a new reset email", int(remaining.Seconds())+1))
		}
	}

	tokenStr, err := generateToken()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to generate reset token", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	_, err = s.repo.UpdateUserPasswordResetToken(ctx, repository.UpdateUserPasswordResetTokenParams{
		ID:                 user.ID,
		PasswordResetToken: sql.NullString{String: tokenStr, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store password reset token", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	if err := s.mailer.SendPasswordResetEmail(user.Email, user.Name, tokenStr); err != nil {
		logger.Ctx(ctx).Error("Failed to send password reset email", zap.Error(err))
		return response.NewAppError(500, "Failed to send password reset email")
	}

	logger.Ctx(ctx).Info("Password reset email sent",
		zap.String("email", user.Email),
	)

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	user, err := s.repo.GetUserByPasswordResetToken(ctx, sql.NullString{String: req.Token, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NewAppError(400, "Invalid reset code, or your code has been used")
		}
		logger.Ctx(ctx).Error("Database error during reset token lookup", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	if user.PasswordResetSentAt.Valid {
		elapsed := time.Since(user.PasswordResetSentAt.Time)
		if elapsed > constants.PasswordResetTokenTTL {
			return response.NewAppError(400, "Reset code has expired. Request a new one.")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to hash password", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	_, err = s.repo.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID:       user.ID,
		Password: sql.NullString{String: string(hash), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update password", zap.Error(err))
		return response.NewAppError(500, "Internal server error")
	}

	return nil
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
		if err == sql.ErrNoRows {
			return nil, response.NewAppError(401, "User not found")
		}
		logger.Ctx(ctx).Error("Database error during refresh token user lookup", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
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

func (s *AuthService) DowngradeToUser(ctx context.Context, userID uuid.UUID, currentRole string) (*dto.AuthResponse, error) {
	if currentRole != constants.RoleAdvertiser {
		return nil, response.NewAppError(400, "Only users with role 'advertiser' can be downgraded")
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
		Role: constants.RoleUser,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to downgrade user role", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to downgrade role")
	}

	// Pause all active ads
	_, err = tx.ExecContext(ctx, `UPDATE ads SET status = 'paused', updated_at = NOW() WHERE advertiser_id = $1 AND status = 'active'`, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to pause ads on downgrade", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to pause ads")
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

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (repository.User, error) {
	return s.repo.GetUserByID(ctx, userID)
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
