package helper

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"urlshortener/pkg/response"
	tjvalidator "urlshortener/pkg/validator"
)

func ParseDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

func FormatDecimal(d decimal.Decimal) string {
	return d.StringFixed(8)
}

func UserIDFromCtx(c *fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("user_id").(string)
	if !ok {
		return uuid.Nil, response.NewAppError(401, "Missing user ID")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, response.NewAppError(401, "Invalid user ID")
	}
	return id, nil
}

func ParseAndValidate(c *fiber.Ctx, v *validator.Validate, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return response.NewAppError(400, "Invalid JSON")
	}
	if err := v.Struct(req); err != nil {
		return response.NewAppError(400, tjvalidator.FormatErrors(err))
	}
	return nil
}
