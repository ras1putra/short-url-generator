package middleware

import (
	"fmt"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	"urlshortener/pkg/token"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func clearAuthCookies(c *fiber.Ctx) {
	expired := time.Now().Add(-time.Hour)
	for _, name := range []string{constants.CookieAccessToken, constants.CookieRefreshToken} {
		c.Cookie(&fiber.Cookie{
			Name:    name,
			Value:   "",
			Path:    "/",
			Expires: expired,
			MaxAge:  -1,
		})
	}
}

func JWTAuth(secret string, redis cache.Cacher, repo repository.Querier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := c.Cookies(constants.CookieAccessToken)
		if tokenStr == "" {
			return response.UnauthorizedWithCode(c, "Missing token", "TOKEN_MISSING")
		}

		claims, err := token.Validate(tokenStr, secret, constants.TokenTypeAccess)
		if err != nil {
			clearAuthCookies(c)
			return response.UnauthorizedWithCode(c, "Invalid or expired token", "TOKEN_INVALID_OR_EXPIRED")
		}

		blacklistKey := fmt.Sprintf("%s%s", constants.RedisPrefixBlacklist, tokenStr)
		blacklisted, _ := redis.Exists(c.Context(), blacklistKey)
		if blacklisted {
			clearAuthCookies(c)
			return response.UnauthorizedWithCode(c, "Token revoked", "TOKEN_REVOKED")
		}

		parsedUserID, err := uuid.Parse(claims.UserID)
		if err != nil {
			clearAuthCookies(c)
			return response.UnauthorizedWithCode(c, "Invalid user ID in token", "TOKEN_USER_INVALID")
		}

		user, err := repo.GetUserByID(c.Context(), parsedUserID)
		if err != nil {
			clearAuthCookies(c)
			return response.UnauthorizedWithCode(c, "User not found", "USER_NOT_FOUND")
		}

		c.Locals("user_id", user.ID.String())
		c.Locals("role", user.Role)

		// Append user_id
		reqLogger := logger.Ctx(c.UserContext()).With(zap.String("user_id", user.ID.String()))
		c.SetUserContext(logger.WithCtx(c.UserContext(), reqLogger))

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return response.Forbidden(c, "Role not found in session")
		}

		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}

		return response.Forbidden(c, "You do not have permission to access this resource")
	}
}
