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

func (h *LinksHandler) parseAndValidate(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return response.NewAppError(400, "Invalid JSON")
	}
	if err := h.validate.Struct(req); err != nil {
		return response.NewAppError(400, tjvalidator.FormatErrors(err))
	}
	return nil
}

func userIDFromCtx(c *fiber.Ctx) (uuid.UUID, error) {
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

func (h *LinksHandler) Create(c *fiber.Ctx) error {
	requestID, _ := c.Locals("request_id").(string)
	userID, err := userIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	var req dto.CreateURLRequest
	if err := h.parseAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.svc.Create(c.Context(), userID, req)
	if err != nil {
		return response.HandleError(c, err, "URL creation")
	}

	logger.WithUser(userID.String()).Info("URL shortened successfully",
		zap.String("slug", resp.Slug),
		zap.String("original", resp.Original),
		zap.String("ip", c.IP()),
		zap.String("request_id", requestID),
	)

	return response.Created(c, resp, "URL shortened successfully")
}

func (h *LinksHandler) List(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 5)
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 5
	}

	resp, err := h.svc.ListByUser(c.Context(), userID, page, perPage)
	if err != nil {
		return response.HandleError(c, err, "URL list")
	}

	return response.OK(c, resp, "URLs fetched successfully")
}

func (h *LinksHandler) Delete(c *fiber.Ctx) error {
	requestID, _ := c.Locals("request_id").(string)
	userID, err := userIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	slug := c.Params("slug")

	if err := h.svc.Delete(c.Context(), userID, slug); err != nil {
		return response.HandleError(c, err, "URL deletion")
	}

	logger.WithUser(userID.String()).Info("URL deleted successfully",
		zap.String("slug", slug),
		zap.String("ip", c.IP()),
		zap.String("request_id", requestID),
	)

	return response.OK(c, nil, "Link deleted successfully")
}

func (h *LinksHandler) Get(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	resp, err := h.svc.GetByID(c.Context(), userID, c.Params("slug"))
	if err != nil {
		return response.HandleError(c, err, "URL detail")
	}

	return response.OK(c, resp, "URL fetched successfully")
}

func (h *LinksHandler) Update(c *fiber.Ctx) error {
	requestID, _ := c.Locals("request_id").(string)
	userID, err := userIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	slug := c.Params("slug")

	var req dto.UpdateURLRequest
	if err := h.parseAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.svc.Update(c.Context(), userID, slug, req)
	if err != nil {
		return response.HandleError(c, err, "URL update")
	}

	logger.WithUser(userID.String()).Info("URL updated successfully",
		zap.String("slug", resp.Slug),
		zap.String("ip", c.IP()),
		zap.String("request_id", requestID),
	)

	return response.OK(c, resp, "URL updated successfully")
}

func (h *LinksHandler) Stats(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	resp, err := h.svc.GetStats(c.Context(), userID, c.Params("slug"))
	if err != nil {
		return response.HandleError(c, err, "URL stats")
	}

	return response.OK(c, resp, "Stats fetched successfully")
}

func (h *LinksHandler) AggregateStats(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	resp, err := h.svc.GetAggregateStats(c.Context(), userID)
	if err != nil {
		return response.HandleError(c, err, "Aggregate stats")
	}

	return response.OK(c, resp, "Aggregate stats fetched successfully")
}

func (h *LinksHandler) QRCode(c *fiber.Ctx) error {
	slug := c.Params("slug")

	url, err := h.svc.GetBySlug(c.Context(), slug)
	if err != nil {
		return err
	}

	size := 256
	if s := c.Query("size"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed >= 64 && parsed <= 1024 {
			size = parsed
		}
	}

	shortURL := dto.MapURLToResponse(*url, h.cfg).ShortURL

	png, err := qrcode.Encode(shortURL, qrcode.Medium, size)
	if err != nil {
		return response.InternalError(c, "Failed to generate QR code")
	}

	c.Set("Content-Type", "image/png")
	c.Set("Cache-Control", "public, max-age=86400")
	return c.Send(png)
}
