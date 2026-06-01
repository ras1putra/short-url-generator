package middleware

import (
	"encoding/json"

	"urlshortener/pkg/logger"
	"urlshortener/pkg/turnstile"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Turnstile(secret string) fiber.Handler {
	if secret == "" {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		token := c.FormValue("cf-turnstile-response")

		if token == "" {
			body := c.Body()
			if len(body) > 0 {
				var data map[string]interface{}
				if err := json.Unmarshal(body, &data); err == nil {
					if t, ok := data["cf-turnstile-response"].(string); ok {
						token = t
					}
				}
			}
		}

		if token == "" {
			logger.Ctx(c.UserContext()).Warn("Turnstile verification failed: missing token")
			return c.Status(400).JSON(fiber.Map{
				"message": "Missing CAPTCHA verification",
			})
		}

		verified, err := turnstile.VerifyToken(secret, token)
		if err != nil {
			logger.Ctx(c.UserContext()).Error("Turnstile verification error", zap.Error(err))
			return c.Status(502).JSON(fiber.Map{
				"message": "CAPTCHA verification failed. Please try again.",
			})
		}

		if !verified {
			logger.Ctx(c.UserContext()).Warn("Turnstile verification failed: invalid token")
			return c.Status(400).JSON(fiber.Map{
				"message": "CAPTCHA verification failed. Please try again.",
			})
		}

		return c.Next()
	}
}
