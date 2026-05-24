package redirect

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"urlshortener/internal/modules/redirect/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
)

type URLGetter interface {
	GetBySlug(ctx context.Context, slug string) (*repository.Url, error)
}

type RedirectHandler struct {
	svc *RedirectService
}

func NewRedirectHandler(svc *RedirectService) *RedirectHandler {
	return &RedirectHandler{svc: svc}
}

func (h *RedirectHandler) Redirect(c *fiber.Ctx) error {
	slug := c.Params("slug")

	url, err := h.svc.GetURL(c.Context(), slug)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Redirect failed: link not found", zap.String("slug", slug), zap.Error(err))
		return response.HandleError(c, err, "Redirect")
	}

	if url.ExpiresAt.Valid && url.ExpiresAt.Time.Before(time.Now()) {
		logger.Ctx(c.UserContext()).Warn("Redirect failed: link expired", zap.String("slug", slug))
		return c.Status(410).SendString("This link has expired.")
	}

	ip := c.IP()
	userAgent := c.Get("User-Agent")
	referer := c.Get("Referer")

	if !url.IsMonetized || len(url.AllowedCategories) == 0 {
		h.svc.EnqueueClick(url.ID, ip, userAgent, referer)
		logger.Ctx(c.UserContext()).Info("Redirected without monetization",
			zap.String("slug", slug),
			zap.String("ip", ip),
			zap.String("user_agent", userAgent),
			zap.String("original", url.Original),
		)
		return c.Redirect(url.Original, fiber.StatusFound)
	}

	ads, err := h.svc.GetActiveAds(c.Context())
	if err != nil || len(ads) == 0 {
		h.svc.EnqueueClick(url.ID, ip, userAgent, referer)
		logger.Ctx(c.UserContext()).Info("Redirected: no active ads available",
			zap.String("slug", slug),
			zap.String("ip", ip),
			zap.String("user_agent", userAgent),
			zap.String("original", url.Original),
		)
		return c.Redirect(url.Original, fiber.StatusFound)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")

	primaryAdID := ads[0].ID

	for _, ad := range ads {
		go h.svc.TrackAdEvent(ad.ID, url.ID, "IMPRESSION", ip, userAgent)
	}

	bridgeToken := h.svc.GenerateBridgeToken(slug, primaryAdID)

	logger.Ctx(c.UserContext()).Info("Monetized interstitial bridge served",
		zap.String("slug", slug),
		zap.Int("active_ads_count", len(ads)),
	)

	return c.SendString(RenderInterstitial(ads, *url, bridgeToken, primaryAdID))
}

func (h *RedirectHandler) AdClick(c *fiber.Ctx) error {
	slug := c.Params("slug")
	adIDStr := c.Params("adID")

	adID, err := uuid.Parse(adIDStr)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("AdClick failed: invalid ad ID format", zap.String("ad_id", adIDStr), zap.Error(err))
		return c.Redirect("/"+slug, fiber.StatusFound)
	}

	url, err := h.svc.GetURL(c.Context(), slug)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("AdClick failed: link not found", zap.String("slug", slug), zap.Error(err))
		return c.Redirect("/", fiber.StatusFound)
	}

	ad, err := h.svc.GetAdByID(c.Context(), adID)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("AdClick failed: ad not found", zap.String("ad_id", adIDStr), zap.Error(err))
		return c.Redirect(url.Original, fiber.StatusFound)
	}

	go h.svc.TrackAdEvent(adID, url.ID, "CLICK", c.IP(), c.Get("User-Agent"))

	logger.Ctx(c.UserContext()).Info("Ad clicked successfully",
		zap.String("slug", slug),
		zap.String("ad_id", adIDStr),
		zap.String("target_url", ad.TargetUrl),
	)

	return c.Redirect(ad.TargetUrl, fiber.StatusFound)
}

func (h *RedirectHandler) AdComplete(c *fiber.Ctx) error {
	slug := c.Params("slug")
	adIDStr := c.Params("adID")

	adID, err := uuid.Parse(adIDStr)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("AdComplete failed: invalid ad ID format", zap.String("ad_id", adIDStr), zap.Error(err))
		return c.Status(400).SendString("Invalid ad ID")
	}

	url, err := h.svc.GetURL(c.Context(), slug)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("AdComplete failed: link not found", zap.String("slug", slug), zap.Error(err))
		return c.Status(404).SendString("Not found")
	}

	go h.svc.TrackAdEvent(adID, url.ID, "COMPLETION", c.IP(), c.Get("User-Agent"))

	logger.Ctx(c.UserContext()).Info("Ad completed successfully",
		zap.String("slug", slug),
		zap.String("ad_id", adIDStr),
	)

	return c.SendString("ok")
}

func (h *RedirectHandler) AdCompleteFlow(c *fiber.Ctx) error {
	slug := c.Params("slug")
	token := c.Query("token")

	verifiedSlug, adID, valid, issuedMs := h.svc.VerifyBridgeToken(token)
	if !valid || verifiedSlug != slug {
		logger.Ctx(c.UserContext()).Warn("AdCompleteFlow failed: invalid or tampered bridge token", zap.String("slug", slug))
		return c.Status(400).JSON(dto.ErrorResponse{Error: "invalid or tampered token"})
	}

	var req CompletionRequest
	if err := c.BodyParser(&req); err != nil {
		req = CompletionRequest{}
	}

	minSessionMs := h.svc.GetMinSessionMs()
	elapsed := time.Now().UnixMilli() - issuedMs
	if elapsed < minSessionMs {
		logger.Ctx(c.UserContext()).Warn("AdCompleteFlow failed: too early completion attempt",
			zap.String("slug", slug),
			zap.Int64("elapsed_ms", elapsed),
			zap.Int64("min_session_ms", minSessionMs),
		)
		return c.Status(429).JSON(dto.ErrorResponse{
			Error:       "too early",
			RemainingMs: minSessionMs - elapsed,
		})
	}

	url, err := h.svc.GetURL(c.Context(), slug)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("AdCompleteFlow failed: link not found", zap.String("slug", slug), zap.Error(err))
		return c.Status(404).JSON(dto.ErrorResponse{Error: "link not found"})
	}

	ip := c.IP()
	userAgent := c.Get("User-Agent")

	if err := h.svc.ChargeForCompletion(c.Context(), slug, adID, url.ID, url.UserID, ip, userAgent, issuedMs, req); err != nil {
		logger.Ctx(c.UserContext()).Error("ChargeForCompletion failed",
			zap.String("slug", slug),
			zap.String("ad_id", adID.String()),
			zap.Error(err),
		)
	}

	h.svc.EnqueueClick(url.ID, ip, userAgent, c.Get("Referer"))

	logger.Ctx(c.UserContext()).Info("AdCompleteFlow charge and redirected successfully",
		zap.String("slug", slug),
		zap.String("ad_id", adID.String()),
		zap.String("destination_url", url.Original),
	)

	return c.JSON(dto.CompletionResponse{
		Success:        true,
		DestinationURL: url.Original,
	})
}

func (h *RedirectHandler) AdSkip(c *fiber.Ctx) error {
	slug := c.Params("slug")
	adIDStr := c.Params("adID")

	adID, err := uuid.Parse(adIDStr)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("AdSkip failed: invalid ad ID format", zap.String("ad_id", adIDStr), zap.Error(err))
		return c.Status(400).SendString("Invalid ad ID")
	}

	url, err := h.svc.GetURL(c.Context(), slug)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("AdSkip failed: link not found", zap.String("slug", slug), zap.Error(err))
		return c.Status(404).SendString("Not found")
	}

	go h.svc.TrackAdEvent(adID, url.ID, "SKIP", c.IP(), c.Get("User-Agent"))

	logger.Ctx(c.UserContext()).Info("Ad skipped successfully",
		zap.String("slug", slug),
		zap.String("ad_id", adIDStr),
		zap.String("destination_url", url.Original),
	)

	return c.Redirect(url.Original, fiber.StatusFound)
}
