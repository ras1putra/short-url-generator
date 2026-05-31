package oauth

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/auth"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
)

type OAuthHandler struct {
	svc         *OAuthService
	cookies     *auth.CookieHelper
	frontendURL string
}

func NewOAuthHandler(svc *OAuthService, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		svc:         svc,
		cookies:     auth.NewCookieHelper(cfg),
		frontendURL: cfg.FrontendURL,
	}
}

func (h *OAuthHandler) Login(c *fiber.Ctx) error {
	intent := c.Query("intent", "")
	redirectURL, err := h.svc.GetLoginURL(intent)
	if err != nil {
		logger.Ctx(c.Context()).Error("Failed to generate Google login URL", zap.Error(err))
		return response.HandleError(c, err, "GoogleLogin")
	}

	return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}

func (h *OAuthHandler) Callback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return response.HandleError(c, response.NewAppError(400, "Missing code or state parameter"), "GoogleCallback")
	}

	if errParam := c.Query("error"); errParam != "" {
		logger.Ctx(c.Context()).Warn("Google OAuth error", zap.String("error", errParam))
		return c.Redirect(h.frontendURL+"/login?oauth_error="+url.QueryEscape(errParam), fiber.StatusTemporaryRedirect)
	}

	authResp, intent, err := h.svc.HandleCallback(c.Context(), code, state)
	if err != nil {
		logger.Ctx(c.Context()).Error("Google OAuth callback failed", zap.Error(err))
		return c.Redirect(h.frontendURL+"/login?oauth_error=authentication_failed", fiber.StatusTemporaryRedirect)
	}

	h.cookies.SetAuthCookies(c, authResp.AccessToken, authResp.RefreshToken)

	logger.Ctx(c.UserContext()).Info("Google OAuth login successful",
		zap.String("user_id", authResp.User.ID),
		zap.String("email", authResp.User.Email),
		zap.String("intent", intent),
		zap.String("ip", c.IP()),
	)

	redirectPath := "/dashboard"
	if intent == "advertiser" {
		redirectPath = "/dashboard/settings"
	}
	return c.Redirect(h.frontendURL+redirectPath, fiber.StatusTemporaryRedirect)
}
