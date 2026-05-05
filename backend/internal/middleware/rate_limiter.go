package middleware

import (
	"fmt"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func RateLimiter(redis cache.Cacher, limit int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now().Unix()
		windowKey := fmt.Sprintf("rl:%s:%d", ip, now/60)

		count, err := redis.RateLimitIncrement(c.Context(), windowKey, 2*time.Minute)
		if err != nil {
			zap.L().Error("Rate limiter Redis failure", zap.Error(err))
			return response.NewAppError(fiber.StatusInternalServerError, "Internal Server Error")
		}

		if count > limit {
			c.Set("X-RateLimit-Remaining", "0")
			return response.NewAppError(fiber.StatusTooManyRequests, "Too Many Requests")
		}

		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))
		return c.Next()
	}
}
