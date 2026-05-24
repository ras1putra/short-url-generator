package auth

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/google/uuid"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth/dto"
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
	UpgradeToAdvertiser(ctx context.Context, userID uuid.UUID, currentRole string) (*dto.AuthResponse, error)
}

type AuthHandler struct {
	svc      AuthServicer
	validate *validator.Validate
	cookies  *cookieHelper
}

func NewAuthHandler(svc AuthServicer, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, validate: tjvalidator.New(), cookies: newCookieHelper(cfg)}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
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

	h.cookies.setAuthCookies(c, resp.AccessToken, resp.RefreshToken)

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
		h.cookies.clearAuthCookies(c)
		return response.Unauthorized(c, "Refresh token missing")
	}

	resp, err := h.svc.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		h.cookies.clearAuthCookies(c)
		return response.HandleError(c, err, "Token refresh")
	}

	h.cookies.setAccessTokenCookie(c, resp.AccessToken)

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

	h.cookies.clearAuthCookies(c)

	logger.Ctx(c.UserContext()).Info("User logged out successfully",
		zap.String("ip", c.IP()),
	)

	return response.OK(c, nil, "Logged out successfully")
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

	h.cookies.setAccessTokenCookie(c, resp.AccessToken)

	logger.Ctx(c.UserContext()).Info("User upgraded to advertiser",
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Account upgraded to advertiser successfully")
}
