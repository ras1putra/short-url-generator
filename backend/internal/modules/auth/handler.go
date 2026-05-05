package auth

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth/dto"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	tjvalidator "urlshortener/pkg/validator"
)

type AuthServicer interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}

type AuthHandler struct {
	svc      AuthServicer
	validate *validator.Validate
	cfg      *config.Config
}

func NewAuthHandler(svc AuthServicer, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, validate: tjvalidator.New(), cfg: cfg}
}

func (h *AuthHandler) parseAndValidate(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return response.NewAppError(400, "Invalid JSON")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.NewAppError(400, tjvalidator.FormatErrors(err))
	}
	return nil
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	requestID, _ := c.Locals("request_id").(string)

	var req dto.RegisterRequest
	if err := h.parseAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.svc.Register(c.Context(), req)
	if err != nil {
		return response.HandleError(c, err, "Registration")
	}

	logger.WithUser(resp.User.ID).Info("User registered successfully",
		zap.String("email", resp.User.Email),
		zap.String("ip", c.IP()),
		zap.String("request_id", requestID),
	)

	return response.Created(c, resp, "User registered successfully")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	requestID, _ := c.Locals("request_id").(string)

	var req dto.LoginRequest
	if err := h.parseAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.svc.Login(c.Context(), req)
	if err != nil {
		return response.HandleError(c, err, "Login")
	}

	h.setAuthCookies(c, resp)

	logger.WithUser(resp.User.ID).Info("User logged in successfully",
		zap.String("email", resp.User.Email),
		zap.String("ip", c.IP()),
		zap.String("user_agent", string(c.Request().Header.UserAgent())),
		zap.String("request_id", requestID),
	)

	return response.OK(c, resp, "Login successful")
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	requestID, _ := c.Locals("request_id").(string)
	refreshToken := c.Cookies(constants.CookieRefreshToken)
	if refreshToken == "" {
		h.clearAuthCookies(c)
		return response.Unauthorized(c, "Refresh token missing")
	}

	resp, err := h.svc.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		h.clearAuthCookies(c)
		return response.HandleError(c, err, "Token refresh")
	}

	h.setAccessTokenCookie(c, resp.AccessToken)

	logger.WithUser(resp.User.ID).Info("Token refreshed successfully",
		zap.String("ip", c.IP()),
		zap.String("request_id", requestID),
	)

	return response.OK(c, resp, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	requestID, _ := c.Locals("request_id").(string)
	userID, _ := c.Locals("user_id").(string)

	accessToken := c.Cookies(constants.CookieAccessToken)
	refreshToken := c.Cookies(constants.CookieRefreshToken)

	if err := h.svc.Logout(c.Context(), accessToken, refreshToken); err != nil {
		logger.WithUser(userID).Error("Logout failed to revoke tokens",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
	}

	h.clearAuthCookies(c)

	logger.WithUser(userID).Info("User logged out successfully",
		zap.String("ip", c.IP()),
		zap.String("request_id", requestID),
	)

	return response.OK(c, nil, "Logged out successfully")
}

func (h *AuthHandler) setAuthCookies(c *fiber.Ctx, resp *dto.AuthResponse) {
	h.setAccessTokenCookie(c, resp.AccessToken)
	h.setRefreshTokenCookie(c, resp.RefreshToken)
}

func (h *AuthHandler) setAccessTokenCookie(c *fiber.Ctx, value string) {
	c.Cookie(&fiber.Cookie{
		Name:     constants.CookieAccessToken,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(constants.AccessTokenTTL),
		HTTPOnly: true,
		Secure:   !h.cfg.IsDev(),
		SameSite: h.sameSite(),
	})
}

func (h *AuthHandler) setRefreshTokenCookie(c *fiber.Ctx, value string) {
	c.Cookie(&fiber.Cookie{
		Name:     constants.CookieRefreshToken,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(constants.RefreshTokenTTL),
		HTTPOnly: true,
		Secure:   !h.cfg.IsDev(),
		SameSite: h.sameSite(),
	})
}

func (h *AuthHandler) clearAuthCookies(c *fiber.Ctx) {
	expired := time.Now().Add(-time.Hour)
	sameSite := h.sameSite()
	secure := !h.cfg.IsDev()

	for _, name := range []string{constants.CookieAccessToken, constants.CookieRefreshToken} {
		c.Cookie(&fiber.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			Expires:  expired,
			MaxAge:   -1,
			HTTPOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		})
	}
}

func (h *AuthHandler) sameSite() string {
	if h.cfg.IsDev() {
		return "Lax"
	}
	return "Strict"
}
