package middleware

import (
	"fmt"

	"urlshortener/internal/cache"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/response"
	"urlshortener/pkg/token"

	"github.com/gofiber/fiber/v2"
)

func JWTAuth(secret string, redis cache.Cacher) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := c.Cookies(constants.CookieAccessToken)
		if tokenStr == "" {
			return response.Unauthorized(c, "Missing token")
		}

		claims, err := token.Validate(tokenStr, secret, "access")
		if err != nil {
			return response.Unauthorized(c, "Invalid or expired token")
		}

		blacklistKey := fmt.Sprintf("bl:%s", tokenStr)
		blacklisted, _ := redis.Exists(c.Context(), blacklistKey)
		if blacklisted {
			return response.Unauthorized(c, "Token revoked")
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}
