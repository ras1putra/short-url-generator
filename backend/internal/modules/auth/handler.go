package auth

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/google/uuid"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	tjvalidator "urlshortener/pkg/validator"
)

type AuthServicer interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	GetUserByEmail(ctx context.Context, email string) (repository.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (repository.User, error)
	UpgradeToAdvertiser(ctx context.Context, userID uuid.UUID, currentRole string) (*dto.AuthResponse, error)
	DowngradeToUser(ctx context.Context, userID uuid.UUID, currentRole string) (*dto.AuthResponse, error)
	SendVerification(ctx context.Context, req dto.SendVerificationRequest) error
	VerifyEmail(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
}

type AuthHandler struct {
	svc      AuthServicer
	validate *validator.Validate
	cookies  *CookieHelper
}

func NewAuthHandler(svc AuthServicer, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, validate: tjvalidator.New(), cookies: NewCookieHelper(cfg)}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	existing, err := h.svc.GetUserByEmail(c.Context(), req.Email)
	if err == nil {
		logger.Ctx(c.UserContext()).Warn("Registration failed: email already registered", zap.String("email", req.Email), zap.String("ip", c.IP()))
		if existing.Role == constants.RoleUser {
			return c.Status(409).JSON(fiber.Map{
				"message":     "Email already registered as a regular user",
				"can_upgrade": true,
				"data":        nil,
			})
		}
		if existing.Role == constants.RoleAdvertiser {
			return c.Status(409).JSON(fiber.Map{
				"message":        "Email already registered as an advertiser",
				"can_downgrade":  true,
				"data":           nil,
			})
		}
		return c.Status(409).JSON(fiber.Map{
			"message": "Email already in use",
			"data":    nil,
		})
	}

	resp, err := h.svc.Register(c.Context(), req)
	if err != nil {
		return response.HandleError(c, err, "Registration")
	}

	logger.Ctx(c.UserContext()).Info("User registered successfully",
		zap.String("email", resp.User.Email),
		zap.String("ip", c.IP()),
	)

	return response.Created(c, resp, "User registered successfully")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	resp, err := h.svc.Login(c.Context(), req)
	if err != nil {
		return response.HandleError(c, err, "Login")
	}

	h.cookies.SetAuthCookies(c, resp.AccessToken, resp.RefreshToken)

	logger.Ctx(c.UserContext()).Info("User logged in successfully",
		zap.String("email", resp.User.Email),
		zap.String("ip", c.IP()),
		zap.String("user_agent", string(c.Request().Header.UserAgent())),
	)

	return response.OK(c, resp, "Login successful")
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies(constants.CookieRefreshToken)
	if refreshToken == "" {
		h.cookies.ClearAuthCookies(c)
		return response.Unauthorized(c, "Refresh token missing")
	}

	resp, err := h.svc.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		h.cookies.ClearAuthCookies(c)
		return response.HandleError(c, err, "Token refresh")
	}

	h.cookies.SetAccessTokenCookie(c, resp.AccessToken)

	logger.Ctx(c.UserContext()).Info("Token refreshed successfully",
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	accessToken := c.Cookies(constants.CookieAccessToken)
	refreshToken := c.Cookies(constants.CookieRefreshToken)

	if err := h.svc.Logout(c.Context(), accessToken, refreshToken); err != nil {
		logger.Ctx(c.UserContext()).Error("Logout failed to revoke tokens",
			zap.Error(err),
		)
	}

	h.cookies.ClearAuthCookies(c)

	userID, _ := helper.UserIDFromCtx(c)
	logger.Ctx(c.UserContext()).Info("User logged out successfully",
		zap.String("user_id", userID.String()),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, nil, "Logged out successfully")
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	user, err := h.svc.GetUserByID(c.Context(), userID)
	if err != nil {
		return response.HandleError(c, err, "GetCurrentUser")
	}

	return response.OK(c, dto.MapUserToResponse(user), "User fetched successfully")
}

func (h *AuthHandler) UpgradeToAdvertiser(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	role, _ := c.Locals("role").(string)

	resp, err := h.svc.UpgradeToAdvertiser(c.Context(), userID, role)
	if err != nil {
		return response.HandleError(c, err, "UpgradeToAdvertiser")
	}

	h.cookies.SetAccessTokenCookie(c, resp.AccessToken)

	logger.Ctx(c.UserContext()).Info("User upgraded to advertiser",
		zap.String("user_id", userID.String()),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Account upgraded to advertiser successfully")
}

func (h *AuthHandler) DowngradeToUser(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	role, _ := c.Locals("role").(string)

	resp, err := h.svc.DowngradeToUser(c.Context(), userID, role)
	if err != nil {
		return response.HandleError(c, err, "DowngradeToUser")
	}

	h.cookies.SetAccessTokenCookie(c, resp.AccessToken)

	logger.Ctx(c.UserContext()).Info("User downgraded to regular user",
		zap.String("user_id", userID.String()),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Account downgraded to regular user successfully")
}

func (h *AuthHandler) SendVerification(c *fiber.Ctx) error {
	var req dto.SendVerificationRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	if err := h.svc.SendVerification(c.Context(), req); err != nil {
		return response.HandleError(c, err, "SendVerification")
	}

	logger.Ctx(c.UserContext()).Info("Verification email requested", zap.String("email", req.Email), zap.String("ip", c.IP()))
	return response.OK(c, nil, "Verification email sent")
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Missing verification token", "data": nil})
	}

	if err := h.svc.VerifyEmail(c.Context(), token); err != nil {
		return response.HandleError(c, err, "VerifyEmail")
	}
 
	logger.Ctx(c.UserContext()).Info("Email verified successfully", zap.String("ip", c.IP()))
	return response.OK(c, nil, "Email verified successfully")
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	if err := h.svc.ForgotPassword(c.Context(), req); err != nil {
		return response.HandleError(c, err, "ForgotPassword")
	}

	logger.Ctx(c.UserContext()).Info("Password reset requested", zap.String("email", req.Email), zap.String("ip", c.IP()))
	return response.OK(c, nil, "If the email exists, a password reset link has been sent")
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	if err := h.svc.ResetPassword(c.Context(), req); err != nil {
		return response.HandleError(c, err, "ResetPassword")
	}

	logger.Ctx(c.UserContext()).Info("Password reset successfully", zap.String("ip", c.IP()))
	return response.OK(c, nil, "Password reset successfully. Please sign in.")
}
