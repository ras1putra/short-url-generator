package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AppError struct {
	Code      int    `json:"-"`
	Message   string `json:"-"`
	ErrorCode string `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func NewAppErrorWithCode(code int, message, errCode string) *AppError {
	return &AppError{Code: code, Message: message, ErrorCode: errCode}
}

func HandleError(c *fiber.Ctx, err error, op string) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		payload := fiber.Map{"message": appErr.Message, "data": nil}
		if appErr.ErrorCode != "" {
			payload["code"] = appErr.ErrorCode
		}
		return c.Status(appErr.Code).JSON(payload)
	}

	requestID, _ := c.Locals("request_id").(string)
	fields := []zap.Field{
		zap.String("operation", op),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("request_id", requestID),
		zap.Error(err),
	}

	if userID, ok := c.Locals("user_id").(string); ok {
		fields = append(fields, zap.String("user_id", userID))
	}

	zap.L().Error(op+" failed", fields...)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error", "data": nil})
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	var appErr *AppError
	var fiberErr *fiber.Error

	if errors.As(err, &appErr) {
		zap.L().Warn("App error", zap.String("path", c.Path()), zap.Int("code", appErr.Code), zap.String("err", appErr.Message))
		payload := fiber.Map{"message": appErr.Message, "data": nil}
		if appErr.ErrorCode != "" {
			payload["code"] = appErr.ErrorCode
		}
		return c.Status(appErr.Code).JSON(payload)
	}

	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{"message": fiberErr.Message, "data": nil})
	}

	zap.L().Error("Internal server error", zap.String("path", c.Path()), zap.Error(err))
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error", "data": nil})
}

func OK(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": message, "data": data})
}

func Created(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": message, "data": data})
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": message, "data": nil})
}

func UnauthorizedWithCode(c *fiber.Ctx, message, errCode string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": message, "data": nil, "code": errCode})
}

func Forbidden(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": message, "data": nil})
}

func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": message, "data": nil})
}

func InternalError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": message, "data": nil})
}
