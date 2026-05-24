package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"urlshortener/internal/config"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/token"
)

type cookieHelper struct {
	cfg *config.Config
}

func newCookieHelper(cfg *config.Config) *cookieHelper {
	return &cookieHelper{cfg: cfg}
}

func (h *cookieHelper) setAuthCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	h.setAccessTokenCookie(c, accessToken)
	h.setRefreshTokenCookie(c, refreshToken)
}

func (h *cookieHelper) setAccessTokenCookie(c *fiber.Ctx, value string) {
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

func (h *cookieHelper) setRefreshTokenCookie(c *fiber.Ctx, value string) {
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

func (h *cookieHelper) clearAuthCookies(c *fiber.Ctx) {
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

func (h *cookieHelper) sameSite() string {
	if h.cfg.IsDev() {
		return constants.SameSiteLax
	}
	return constants.SameSiteStrict
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
