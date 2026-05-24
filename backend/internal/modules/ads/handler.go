package ads

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"urlshortener/internal/modules/ads/dto"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	tjvalidator "urlshortener/pkg/validator"
)

type AdServicer interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateAdRequest) (*dto.AdResponse, error)
	GetByID(ctx context.Context, adID, userID uuid.UUID) (*dto.AdResponse, error)
	ListByAdvertiser(ctx context.Context, userID uuid.UUID) ([]dto.AdResponse, error)
	Update(ctx context.Context, adID, userID uuid.UUID, req dto.UpdateAdRequest) (*dto.AdResponse, error)
	Delete(ctx context.Context, adID, userID uuid.UUID) error
	GetStats(ctx context.Context, adID, userID uuid.UUID) (*dto.AdStatsResponse, error)
	ListCategories(ctx context.Context) ([]dto.CategoryResponse, error)
	ListAdTypes(ctx context.Context) ([]dto.AdTypeResponse, error)
	TopUp(ctx context.Context, adID, userID uuid.UUID, req dto.TopUpAdRequest) (*dto.AdResponse, error)
}

type AdHandler struct {
	svc      AdServicer
	validate *validator.Validate
}

func NewAdHandler(svc AdServicer) *AdHandler {
	return &AdHandler{svc: svc, validate: tjvalidator.New()}
}

func (h *AdHandler) Create(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	var req dto.CreateAdRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	resp, err := h.svc.Create(c.Context(), userID, req)
	if err != nil {
		return response.HandleError(c, err, "CreateAd")
	}

	logger.Ctx(c.UserContext()).Info("Ad campaign created",
		zap.String("ad_id", resp.ID),
		zap.String("ip", c.IP()),
	)

	return response.Created(c, resp, "Campaign created successfully")
}

func (h *AdHandler) List(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	resp, err := h.svc.ListByAdvertiser(c.Context(), userID)
	if err != nil {
		return response.HandleError(c, err, "ListAds")
	}

	logger.Ctx(c.UserContext()).Info("Ad campaigns listed",
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Campaigns fetched successfully")
}

func (h *AdHandler) GetByID(c *fiber.Ctx) error {
	adID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.NewAppError(400, "Invalid ad ID")
	}

	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	resp, err := h.svc.GetByID(c.Context(), adID, userID)
	if err != nil {
		return response.HandleError(c, err, "GetAdByID")
	}

	logger.Ctx(c.UserContext()).Info("Ad campaign fetched",
		zap.String("ad_id", adID.String()),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Campaign fetched successfully")
}

func (h *AdHandler) Update(c *fiber.Ctx) error {
	adID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.NewAppError(400, "Invalid ad ID")
	}

	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	var req dto.UpdateAdRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	resp, err := h.svc.Update(c.Context(), adID, userID, req)
	if err != nil {
		return response.HandleError(c, err, "UpdateAd")
	}

	logger.Ctx(c.UserContext()).Info("Ad campaign updated",
		zap.String("ad_id", adID.String()),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Campaign updated successfully")
}

func (h *AdHandler) Delete(c *fiber.Ctx) error {
	adID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.NewAppError(400, "Invalid ad ID")
	}

	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	if err := h.svc.Delete(c.Context(), adID, userID); err != nil {
		return response.HandleError(c, err, "DeleteAd")
	}

	logger.Ctx(c.UserContext()).Info("Ad campaign deleted",
		zap.String("ad_id", adID.String()),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, nil, "Campaign deleted successfully")
}

func (h *AdHandler) ListCategories(c *fiber.Ctx) error {
	resp, err := h.svc.ListCategories(c.Context())
	if err != nil {
		return response.HandleError(c, err, "ListCategories")
	}

	logger.Ctx(c.UserContext()).Info("Ad categories listed")

	return response.OK(c, resp, "Categories fetched successfully")
}

func (h *AdHandler) ListAdTypes(c *fiber.Ctx) error {
	resp, err := h.svc.ListAdTypes(c.Context())
	if err != nil {
		return response.HandleError(c, err, "ListAdTypes")
	}

	logger.Ctx(c.UserContext()).Info("Ad types listed")

	return response.OK(c, resp, "Ad types fetched successfully")
}

func (h *AdHandler) GetStats(c *fiber.Ctx) error {
	adID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.NewAppError(400, "Invalid ad ID")
	}

	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	resp, err := h.svc.GetStats(c.Context(), adID, userID)
	if err != nil {
		return response.HandleError(c, err, "GetAdStats")
	}

	logger.Ctx(c.UserContext()).Info("Ad campaign stats fetched",
		zap.String("ad_id", adID.String()),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Campaign stats fetched successfully")
}

func (h *AdHandler) TopUp(c *fiber.Ctx) error {
	adID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.NewAppError(400, "Invalid ad ID")
	}

	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	var req dto.TopUpAdRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	resp, err := h.svc.TopUp(c.Context(), adID, userID, req)
	if err != nil {
		return response.HandleError(c, err, "TopUpAd")
	}

	logger.Ctx(c.UserContext()).Info("Ad campaign topped up",
		zap.String("ad_id", adID.String()),
		zap.String("amount", helper.FormatDecimal(decimal.NewFromFloat(req.Amount))),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "Campaign topped up successfully")
}
