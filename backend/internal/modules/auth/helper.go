package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"urlshortener/internal/config"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/token"
)

type CookieHelper struct {
	cfg *config.Config
}

func NewCookieHelper(cfg *config.Config) *CookieHelper {
	return &CookieHelper{cfg: cfg}
}

func (h *CookieHelper) SetAuthCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	h.SetAccessTokenCookie(c, accessToken)
	h.SetRefreshTokenCookie(c, refreshToken)
}

func (h *CookieHelper) SetAccessTokenCookie(c *fiber.Ctx, value string) {
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

func (h *CookieHelper) SetRefreshTokenCookie(c *fiber.Ctx, value string) {
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

func (h *CookieHelper) ClearAuthCookies(c *fiber.Ctx) {
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

func (h *CookieHelper) sameSite() string {
	return constants.SameSiteLax
}

func issueTokens(user repository.User, cfg *config.Config) (accessToken, refreshToken string, err error) {
	accessToken, err = token.IssueToken(user.ID.String(), user.Role, cfg.JWTAccessSecret, constants.TokenTypeAccess, constants.AccessTokenTTL)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = token.IssueToken(user.ID.String(), user.Role, cfg.JWTRefreshSecret, constants.TokenTypeRefresh, constants.RefreshTokenTTL)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func timeUntilExpiry(claims *token.Claims) time.Duration {
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return 0
	}
	return ttl
}
