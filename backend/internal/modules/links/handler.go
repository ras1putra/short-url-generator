package links

import (
	"context"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"

	"urlshortener/internal/config"
	"urlshortener/internal/modules/links/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
	tjvalidator "urlshortener/pkg/validator"
)

type URLServicer interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateURLRequest) (*dto.URLResponse, error)
	GetBySlug(ctx context.Context, slug string) (*repository.Url, error)
	GetByID(ctx context.Context, userID uuid.UUID, slug string) (*dto.URLResponse, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int) (*dto.ListResponse, error)
	Update(ctx context.Context, userID uuid.UUID, slug string, req dto.UpdateURLRequest) (*dto.URLResponse, error)
	GetStats(ctx context.Context, userID uuid.UUID, slug string) (*dto.StatsResponse, error)
	GetAggregateStats(ctx context.Context, userID uuid.UUID) (*dto.StatsResponse, error)
	Delete(ctx context.Context, userID uuid.UUID, slug string) error
}

type LinksHandler struct {
	svc      URLServicer
	validate *validator.Validate
	cfg      *config.Config
}

func NewLinksHandler(svc URLServicer, cfg *config.Config) *LinksHandler {
	return &LinksHandler{svc: svc, validate: tjvalidator.New(), cfg: cfg}
}

func (h *LinksHandler) Create(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	var req dto.CreateURLRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	resp, err := h.svc.Create(c.Context(), userID, req)
	if err != nil {
		return response.HandleError(c, err, "URL creation")
	}

	logger.Ctx(c.UserContext()).Info("URL shortened successfully",
		zap.String("slug", resp.Slug),
		zap.String("original", resp.Original),
		zap.String("ip", c.IP()),
	)

	return response.Created(c, resp, "URL shortened successfully")
}

func (h *LinksHandler) List(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	page := c.QueryInt("page", constants.DefaultPage)
	perPage := c.QueryInt("per_page", constants.DefaultPerPage)
	if page < 1 {
		page = constants.DefaultPage
	}
	if perPage < 1 || perPage > constants.MaxPerPage {
		perPage = constants.DefaultPerPage
	}

	resp, err := h.svc.ListByUser(c.Context(), userID, page, perPage)
	if err != nil {
		return response.HandleError(c, err, "URL list")
	}

	logger.Ctx(c.UserContext()).Info("URLs listed",
		zap.Int("page", page),
		zap.Int("per_page", perPage),
	)

	return response.OK(c, resp, "URLs fetched successfully")
}

func (h *LinksHandler) Delete(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	slug := c.Params("slug")

	if err := h.svc.Delete(c.Context(), userID, slug); err != nil {
		return response.HandleError(c, err, "URL deletion")
	}

	logger.Ctx(c.UserContext()).Info("URL deleted successfully",
		zap.String("slug", slug),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, nil, "Link deleted successfully")
}

func (h *LinksHandler) Get(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	slug := c.Params("slug")

	resp, err := h.svc.GetByID(c.Context(), userID, slug)
	if err != nil {
		return response.HandleError(c, err, "URL detail")
	}

	logger.Ctx(c.UserContext()).Info("URL fetched",
		zap.String("slug", slug),
	)

	return response.OK(c, resp, "URL fetched successfully")
}

func (h *LinksHandler) Update(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	slug := c.Params("slug")

	var req dto.UpdateURLRequest
	if err := helper.ParseAndValidate(c, h.validate, &req); err != nil {
		return err
	}

	resp, err := h.svc.Update(c.Context(), userID, slug, req)
	if err != nil {
		return response.HandleError(c, err, "URL update")
	}

	logger.Ctx(c.UserContext()).Info("URL updated successfully",
		zap.String("slug", resp.Slug),
		zap.String("ip", c.IP()),
	)

	return response.OK(c, resp, "URL updated successfully")
}

func (h *LinksHandler) Stats(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	slug := c.Params("slug")

	resp, err := h.svc.GetStats(c.Context(), userID, slug)
	if err != nil {
		return response.HandleError(c, err, "URL stats")
	}

	logger.Ctx(c.UserContext()).Info("URL stats fetched",
		zap.String("slug", slug),
	)

	return response.OK(c, resp, "Stats fetched successfully")
}

func (h *LinksHandler) AggregateStats(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	resp, err := h.svc.GetAggregateStats(c.Context(), userID)
	if err != nil {
		return response.HandleError(c, err, "Aggregate stats")
	}

	logger.Ctx(c.UserContext()).Info("Aggregate stats fetched")

	return response.OK(c, resp, "Aggregate stats fetched successfully")
}

func (h *LinksHandler) QRCode(c *fiber.Ctx) error {
	slug := c.Params("slug")

	url, err := h.svc.GetBySlug(c.Context(), slug)
	if err != nil {
		return response.HandleError(c, err, "QR Code generation")
	}

	size := constants.QRCodeDefaultSize
	if s := c.Query("size"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed >= constants.QRCodeMinSize && parsed <= constants.QRCodeMaxSize {
			size = parsed
		}
	}

	shortURL := dto.MapURLToResponse(*url, h.cfg).ShortURL

	png, err := qrcode.Encode(shortURL, qrcode.Medium, size)
	if err != nil {
		return response.InternalError(c, "Failed to generate QR code")
	}

	logger.Ctx(c.UserContext()).Info("QR code generated",
		zap.String("slug", slug),
		zap.Int("size", size),
	)

	c.Set("Content-Type", "image/png")
	c.Set("Cache-Control", "public, max-age=86400")
	return c.Send(png)
}
