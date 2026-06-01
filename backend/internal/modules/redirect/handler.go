package redirect

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"urlshortener/internal/modules/redirect/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
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

	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")

	ads, err := h.svc.GetActiveAdsByCategory(c.Context(), url.AllowedCategories)
	if err != nil {
		ads = nil
	}

	if len(ads) > 0 {
		rendered, renderedIDs, primaryAdID := h.svc.SelectAndTrackAds(ads, url.ID, ip, userAgent)

		bridgeToken := h.svc.GenerateBridgeToken(slug, renderedIDs)

		logger.Ctx(c.UserContext()).Info("Monetized interstitial bridge served",
			zap.String("slug", slug),
			zap.Int("active_ads_count", len(ads)),
		)

		return c.SendString(RenderInterstitial(rendered, *url, bridgeToken, primaryAdID, h.svc.TurnstileSiteKey()))
	}

	h.svc.EnqueueClick(url.ID, ip, userAgent, referer)
	logger.Ctx(c.UserContext()).Info("No matching ads for categories, showing interstitial with placeholders",
		zap.String("slug", slug),
		zap.Strings("allowed_categories", url.AllowedCategories),
	)
	placeholderToken := h.svc.GenerateBridgeToken(slug, []uuid.UUID{uuid.Nil})
	return c.SendString(RenderInterstitial(nil, *url, placeholderToken, uuid.Nil, h.svc.TurnstileSiteKey()))
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

	go h.svc.TrackAdEvent(adID, url.ID, constants.AdEventClick, c.IP(), c.Get("User-Agent"))

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

	go h.svc.TrackAdEvent(adID, url.ID, constants.AdEventCompletion, c.IP(), c.Get("User-Agent"))

	logger.Ctx(c.UserContext()).Info("Ad completed successfully",
		zap.String("slug", slug),
		zap.String("ad_id", adIDStr),
	)

	return c.SendString("ok")
}

func (h *RedirectHandler) AdCompleteFlow(c *fiber.Ctx) error {
	slug := c.Params("slug")
	token := c.Query("token")

	verifiedSlug, adIDs, valid, issuedMs := h.svc.VerifyBridgeToken(token)
	if !valid || verifiedSlug != slug {
		logger.Ctx(c.UserContext()).Warn("AdCompleteFlow failed: invalid or tampered bridge token", zap.String("slug", slug))
		return c.Status(400).JSON(dto.ErrorResponse{Error: "invalid or tampered token"})
	}

	var req CompletionRequest
	if err := c.BodyParser(&req); err != nil {
		req = CompletionRequest{}
	}

	if !h.svc.VerifyTurnstileToken(req.TurnstileToken) {
		logger.Ctx(c.UserContext()).Warn("AdCompleteFlow failed: turnstile verification failed", zap.String("slug", slug))
		return c.Status(400).JSON(dto.ErrorResponse{Error: "CAPTCHA verification failed"})
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

	var representativeAdID uuid.UUID
	if len(adIDs) > 0 {
		representativeAdID = adIDs[0]
	}

	quality := h.svc.runQualityChecks(c.Context(), slug, representativeAdID.String(), ip, req.Fingerprint, req.HoneypotHit, req.MouseMoves, issuedMs)

	for _, adID := range adIDs {
		if adID == uuid.Nil {
			continue
		}
		if err := h.svc.ChargeForCompletion(c.Context(), slug, adID, url.ID, url.UserID, ip, userAgent, issuedMs, req, &quality); err != nil {
			logger.Ctx(c.UserContext()).Error("ChargeForCompletion failed",
				zap.String("slug", slug),
				zap.String("ad_id", adID.String()),
				zap.Error(err),
			)
		}
	}

	if url.IsMonetized {
		if err := h.svc.CreditPlatformReward(c.Context(), url.UserID, slug, &quality); err != nil {
			logger.Ctx(c.UserContext()).Error("CreditPlatformReward failed",
				zap.String("slug", slug),
				zap.Error(err),
			)
		}
	}

	h.svc.EnqueueClick(url.ID, ip, userAgent, c.Get("Referer"))

	logger.Ctx(c.UserContext()).Info("AdCompleteFlow charges completed and redirected successfully",
		zap.String("slug", slug),
		zap.Int("charged_ads_count", len(adIDs)),
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

	go h.svc.TrackAdEvent(adID, url.ID, constants.AdEventSkip, c.IP(), c.Get("User-Agent"))

	logger.Ctx(c.UserContext()).Info("Ad skipped successfully",
		zap.String("slug", slug),
		zap.String("ad_id", adIDStr),
		zap.String("destination_url", url.Original),
	)

	return c.Redirect(url.Original, fiber.StatusFound)
}
