package redirect

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mssola/user_agent"
	"github.com/oschwald/geoip2-golang"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sqlc-dev/pqtype"
	"go.uber.org/zap"

	"urlshortener/internal/analytics"
	"urlshortener/internal/config"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
)

type qualityResult struct {
	Score        float64
	RejectReason string
	IsValid      bool
}

type CompletionRequest struct {
	Fingerprint string `json:"fingerprint"`
	HoneypotHit bool   `json:"honeypot_hit"`
	MouseMoves  int    `json:"mouse_moves"`
}

type RedirectService struct {
	urlSvc     URLGetter
	repo       repository.Querier
	worker     *analytics.AnalyticsWorker
	geoDB      *geoip2.Reader
	hmacSecret []byte
	db         *sql.DB
	redis      *redis.Client
	qualityCfg qualityConfig
}

type qualityConfig struct {
	minScore     float64
	minSessionMs int64
}

func NewRedirectService(urlSvc URLGetter, repo repository.Querier, worker *analytics.AnalyticsWorker, geoDB *geoip2.Reader, cfg *config.Config, db *sql.DB, redisClient *redis.Client) *RedirectService {
	return &RedirectService{
		urlSvc:     urlSvc,
		repo:       repo,
		worker:     worker,
		geoDB:      geoDB,
		hmacSecret: []byte(cfg.BridgeHMACSecret),
		db:         db,
		redis:      redisClient,
		qualityCfg: qualityConfig{
			minScore:     cfg.QualityMinScore,
			minSessionMs: cfg.MinSessionMs,
		},
	}
}

func (s *RedirectService) GetURL(ctx context.Context, slug string) (*repository.Url, error) {
	return s.urlSvc.GetBySlug(ctx, slug)
}

func (s *RedirectService) GetActiveAds(ctx context.Context) ([]repository.Ad, error) {
	ads, err := s.repo.GetActiveAds(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Database error getting active ads", zap.Error(err))
	}
	return ads, err
}

func (s *RedirectService) GetActiveAdsByCategory(ctx context.Context, categories []string) ([]repository.Ad, error) {
	ads, err := s.repo.GetActiveAdsByCategory(ctx, categories)
	if err != nil {
		logger.Ctx(ctx).Error("Database error getting active ads by category", zap.Error(err))
	}
	return ads, err
}

func (s *RedirectService) GetAdByID(ctx context.Context, adID uuid.UUID) (*repository.Ad, error) {
	ad, err := s.repo.GetAdByID(ctx, adID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Ctx(ctx).Error("Database error getting ad by ID", zap.String("ad_id", adID.String()), zap.Error(err))
		}
		return nil, err
	}
	return &ad, nil
}

func (s *RedirectService) TrackAdEvent(adID, linkID uuid.UUID, eventType, ip, userAgent string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	parsedIP := net.ParseIP(ip)
	var inet pqtype.Inet
	if parsedIP != nil {
		inet.IPNet = net.IPNet{IP: parsedIP, Mask: net.CIDRMask(32, 32)}
		inet.Valid = true
	}

	qualityScore := constants.QualityScoreDefault
	if eventType == constants.AdEventSkip {
		qualityScore = constants.QualityScoreSkip
	}

	_, err := s.repo.CreateAdEvent(ctx, repository.CreateAdEventParams{
		AdID:            adID,
		LinkID:          linkID,
		EventType:       eventType,
		IsValid:         true,
		QualityScore:    qualityScore,
		RejectionReason: sql.NullString{},
		IpAddress:       inet,
		UserAgent:       sql.NullString{String: userAgent, Valid: userAgent != ""},
		Metadata:        pqtype.NullRawMessage{},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to track ad event",
			zap.String("event_type", eventType),
			zap.String("ad_id", adID.String()),
			zap.String("link_id", linkID.String()),
			zap.Error(err),
		)
	}
}

func (s *RedirectService) EnqueueClick(urlID uuid.UUID, ip, userAgent, referer string) {
	ua := user_agent.New(userAgent)
	browser, _ := ua.Browser()

	device := constants.DeviceDesktop
	if ua.Mobile() {
		device = constants.DeviceMobile
	}
	if ua.Bot() {
		device = constants.DeviceBot
	}

	country, city := s.resolveGeo(ip)

	hash := sha256.Sum256([]byte(ip))
	ipHash := hex.EncodeToString(hash[:])

	s.worker.Enqueue(analytics.ClickEvent{
		UrlID:    urlID,
		IPHash:   ipHash,
		Country:  country,
		City:     city,
		Device:   device,
		Browser:  browser,
		Referrer: referer,
	})
}

func (s *RedirectService) GenerateBridgeToken(slug string, adIDs []uuid.UUID) string {
	now := time.Now().UnixMilli()
	var idsStr []string
	for _, id := range adIDs {
		idsStr = append(idsStr, id.String())
	}
	joinedIDs := strings.Join(idsStr, ",")
	payload := fmt.Sprintf("%s:%s:%d", slug, joinedIDs, now)
	mac := hmac.New(sha256.New, s.hmacSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s:%s:%d:%s", slug, joinedIDs, now, sig)
}

func (s *RedirectService) VerifyBridgeToken(token string) (string, []uuid.UUID, bool, int64) {
	parts := strings.SplitN(token, ":", 4)
	if len(parts) != 4 {
		return "", nil, false, 0
	}
	slug := parts[0]
	idsStr := parts[1]
	ts := parts[2]
	sig := parts[3]

	mac := hmac.New(sha256.New, s.hmacSecret)
	mac.Write([]byte(slug + ":" + idsStr + ":" + ts))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", nil, false, 0
	}

	var adIDs []uuid.UUID
	if idsStr != "" && idsStr != "00000000-0000-0000-0000-000000000000" {
		for _, idStr := range strings.Split(idsStr, ",") {
			adID, err := uuid.Parse(idStr)
			if err == nil {
				adIDs = append(adIDs, adID)
			}
		}
	}

	var ms int64
	fmt.Sscanf(ts, "%d", &ms)
	return slug, adIDs, true, ms
}

func (s *RedirectService) runQualityChecks(ctx context.Context, slug, adID, ip, fingerprint string, honeypotHit bool, mouseMoves int, issuedMs int64) qualityResult {
	elapsed := time.Now().UnixMilli() - issuedMs

	if honeypotHit {
		logger.Ctx(ctx).Warn("Quality check failed: honeypot hit", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip))
		return qualityResult{Score: 0, RejectReason: constants.RejectReasonHoneypotHit, IsValid: false}
	}

	if elapsed < s.qualityCfg.minSessionMs {
		logger.Ctx(ctx).Warn("Quality check failed: session too fast", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip), zap.Int64("elapsed_ms", elapsed), zap.Int64("min_session_ms", s.qualityCfg.minSessionMs))
		return qualityResult{Score: 0, RejectReason: constants.RejectReasonTooFast, IsValid: false}
	}

	if mouseMoves < 2 {
		logger.Ctx(ctx).Warn("Quality check failed: insufficient mouse movements", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip), zap.Int("mouse_moves", mouseMoves))
		return qualityResult{Score: 0, RejectReason: constants.RejectReasonNoMouseMovement, IsValid: false}
	}

	if s.redis != nil {
		ipKey := constants.RedisPrefixQualityIP + slug + ":" + ip
		exists, _ := s.redis.Exists(ctx, ipKey).Result()
		if exists > 0 {
			logger.Ctx(ctx).Warn("Quality check failed: duplicate IP session within 24h", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip))
			return qualityResult{Score: 0, RejectReason: constants.RejectReasonDuplicateIP, IsValid: false}
		}

		if fingerprint != "" {
			fpKey := constants.RedisPrefixQualityFP + slug + ":" + fingerprint
			exists, _ := s.redis.Exists(ctx, fpKey).Result()
			if exists > 0 {
				logger.Ctx(ctx).Warn("Quality check failed: duplicate browser fingerprint within 24h", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("fingerprint", fingerprint))
				return qualityResult{Score: 0, RejectReason: constants.RejectReasonDuplicateFingerprint, IsValid: false}
			}
		}

		s.redis.Set(ctx, ipKey, "1", 24*time.Hour)
		if fingerprint != "" {
			s.redis.Set(ctx, constants.RedisPrefixQualityFP+slug+":"+fingerprint, "1", 24*time.Hour)
		}
	}

	logger.Ctx(ctx).Info("Quality check passed successfully", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip))
	return qualityResult{Score: 1.00, IsValid: true}
}

func (s *RedirectService) ChargeForCompletion(ctx context.Context, slug string, adID, linkID uuid.UUID, linkOwnerID uuid.UUID, ip, userAgent string, issuedMs int64, req CompletionRequest, preQuality *qualityResult) error {
	ad, err := s.repo.GetAdByID(ctx, adID)
	if err != nil {
		return fmt.Errorf("ad not found: %w", err)
	}

	if ad.Status != constants.AdStatusActive {
		logger.Ctx(ctx).Warn("Charge skipped: ad not active", zap.String("slug", slug), zap.String("ad_id", adID.String()), zap.String("status", ad.Status))
		return nil
	}

	cpm, err := decimal.NewFromString(ad.Cpm)
	if err != nil {
		return fmt.Errorf("invalid cpm: %w", err)
	}

	costPerView := cpm.Div(decimal.NewFromInt(1000))
	remaining, err := decimal.NewFromString(ad.RemainingBudget)
	if err != nil {
		return fmt.Errorf("invalid remaining budget: %w", err)
	}

	if remaining.LessThan(costPerView) {
		logger.Ctx(ctx).Warn("Charge skipped: ad remaining budget is too low", zap.String("slug", slug), zap.String("ad_id", adID.String()), zap.String("remaining", ad.RemainingBudget), zap.String("cost_per_view", costPerView.String()))
		return nil
	}

	quality := preQuality
	if quality == nil {
		q := s.runQualityChecks(ctx, slug, adID.String(), ip, req.Fingerprint, req.HoneypotHit, req.MouseMoves, issuedMs)
		quality = &q
	}

	effectiveCost := costPerView
	effectiveShare := costPerView.Mul(decimal.NewFromInt(7)).Div(decimal.NewFromInt(10))

	if !quality.IsValid {
		effectiveCost = decimal.Zero
		effectiveShare = decimal.Zero
	} else if quality.Score < 1.00 {
		factor := decimal.NewFromFloat(quality.Score)
		effectiveCost = costPerView.Mul(factor)
		effectiveShare = effectiveCost.Mul(decimal.NewFromInt(7)).Div(decimal.NewFromInt(10))
	}

	meta := pqtype.NullRawMessage{}
	metaBytes, _ := json.Marshal(map[string]interface{}{
		"fingerprint":  req.Fingerprint,
		"honeypot_hit": req.HoneypotHit,
		"mouse_moves":  req.MouseMoves,
	})
	meta.RawMessage = metaBytes
	meta.Valid = len(metaBytes) > 0

	rejectionReason := sql.NullString{}
	if quality.RejectReason != "" {
		rejectionReason.String = quality.RejectReason
		rejectionReason.Valid = true
	}

	parsedIP := net.ParseIP(ip)
	var inet pqtype.Inet
	if parsedIP != nil {
		inet.IPNet = net.IPNet{IP: parsedIP, Mask: net.CIDRMask(32, 32)}
		inet.Valid = true
	}

	_, err = s.repo.CreateAdEvent(ctx, repository.CreateAdEventParams{
		AdID:            adID,
		LinkID:          linkID,
		EventType:       constants.AdEventCompletion,
		IsValid:         quality.IsValid,
		QualityScore:    fmt.Sprintf("%.2f", quality.Score),
		RejectionReason: rejectionReason,
		IpAddress:       inet,
		UserAgent:       sql.NullString{String: userAgent, Valid: userAgent != ""},
		Metadata:        meta,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to track completion ad event", zap.Error(err))
	}

	if !quality.IsValid || effectiveCost.IsZero() {
		logger.Ctx(ctx).Warn("Charge skipped: conversion failed quality checks or cost is zero", zap.String("slug", slug), zap.String("ad_id", adID.String()), zap.String("rejection_reason", quality.RejectReason))
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var q repository.Querier = s.repo
	if queriesInstance, ok := s.repo.(*repository.Queries); ok {
		q = queriesInstance.WithTx(tx)
	} else {
		_ = tx.Rollback()
		return fmt.Errorf("repository not transaction-compatible")
	}

	err = q.DeductAdBudget(ctx, repository.DeductAdBudgetParams{
		ID:              adID,
		RemainingBudget: helper.FormatDecimal(effectiveCost),
	})
	if err != nil {
		return fmt.Errorf("failed to deduct ad budget: %w", err)
	}

	// Auto-pause campaigns that can no longer fund one completion.
	projectedRemaining := remaining.Sub(effectiveCost)
	if projectedRemaining.LessThan(costPerView) {
		if err := q.UpdateAdStatus(ctx, repository.UpdateAdStatusParams{
			ID:     adID,
			Status: constants.AdStatusPaused,
		}); err != nil {
			return fmt.Errorf("failed to auto-pause underfunded campaign: %w", err)
		}
	}

	_, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
		UserID:  linkOwnerID,
		Balance: helper.FormatDecimal(effectiveShare),
	})
	if err != nil {
		return fmt.Errorf("failed to credit link owner: %w", err)
	}

	_, err = q.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID:   linkOwnerID,
		Amount:   helper.FormatDecimal(effectiveShare),
		Type:     constants.TxTypeEarning,
		TxHash:   sql.NullString{},
		Metadata: meta,
	})
	if err != nil {
		return fmt.Errorf("failed to create earning tx: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit charge tx: %w", err)
	}

	logger.Ctx(ctx).Info("Charged advertiser and credited link owner successfully",
		zap.String("slug", slug),
		zap.String("ad_id", adID.String()),
		zap.String("advertiser_id", ad.AdvertiserID.String()),
		zap.String("link_owner_id", linkOwnerID.String()),
		zap.String("charge_amount", effectiveCost.String()),
		zap.String("earned_amount", effectiveShare.String()),
	)

	return nil
}

func (s *RedirectService) CreditPlatformReward(ctx context.Context, linkOwnerID uuid.UUID, slug string, quality *qualityResult) error {
	if quality == nil || !quality.IsValid {
		logger.Ctx(ctx).Warn("Platform reward skipped: quality invalid", zap.String("slug", slug))
		return nil
	}

	reward, err := decimal.NewFromString(constants.PlatformReward)
	if err != nil {
		return fmt.Errorf("invalid platform reward: %w", err)
	}

	if quality.Score < 1.00 {
		reward = reward.Mul(decimal.NewFromFloat(quality.Score))
	}

	if reward.IsZero() {
		logger.Ctx(ctx).Warn("Platform reward skipped: zero after score adjustment", zap.String("slug", slug))
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var q repository.Querier = s.repo
	if queriesInstance, ok := s.repo.(*repository.Queries); ok {
		q = queriesInstance.WithTx(tx)
	} else {
		_ = tx.Rollback()
		return fmt.Errorf("repository not transaction-compatible")
	}

	_, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
		UserID:  linkOwnerID,
		Balance: helper.FormatDecimal(reward),
	})
	if err != nil {
		return fmt.Errorf("failed to credit platform reward: %w", err)
	}

	_, err = q.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID:   linkOwnerID,
		Amount:   helper.FormatDecimal(reward),
		Type:     constants.TxTypeEarning,
		TxHash:   sql.NullString{},
		Metadata: pqtype.NullRawMessage{},
	})
	if err != nil {
		return fmt.Errorf("failed to create platform reward tx: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit platform reward tx: %w", err)
	}

	logger.Ctx(ctx).Info("Platform reward credited",
		zap.String("slug", slug),
		zap.String("link_owner_id", linkOwnerID.String()),
		zap.String("amount", reward.String()),
	)

	return nil
}

func (s *RedirectService) resolveGeo(ip string) (country, city string) {
	if s.geoDB == nil || ip == "" {
		return
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		logger.Ctx(context.Background()).Warn("GeoIP lookup failed: invalid IP", zap.String("ip", ip))
		return
	}

	record, err := s.geoDB.City(parsedIP)
	if err != nil {
		logger.Ctx(context.Background()).Warn("GeoIP lookup failed", zap.String("ip", ip), zap.Error(err))
		return
	}

	country = record.Country.IsoCode
	city = record.City.Names[constants.LocaleCity]
	return
}

func (s *RedirectService) GetMinSessionMs() int64 {
	return s.qualityCfg.minSessionMs
}

func (s *RedirectService) SelectAndTrackAds(ads []repository.Ad, urlID uuid.UUID, ip, userAgent string) ([]repository.Ad, []uuid.UUID, uuid.UUID) {
	g := GroupAds(ads)
	rendered := []repository.Ad{}
	if len(g.Popup) > 0 {
		rendered = append(rendered, g.Popup[rand.Intn(len(g.Popup))])
	}
	if len(g.Banner) > 0 {
		rendered = append(rendered, g.Banner[rand.Intn(len(g.Banner))])
	}
	if len(g.Native) > 0 {
		rendered = append(rendered, g.Native[rand.Intn(len(g.Native))])
	}
	if len(g.Video) > 0 {
		rendered = append(rendered, g.Video[rand.Intn(len(g.Video))])
	}
	if len(g.Interstitial) > 0 {
		rendered = append(rendered, g.Interstitial[rand.Intn(len(g.Interstitial))])
	}

	var renderedIDs []uuid.UUID
	for _, ad := range rendered {
		renderedIDs = append(renderedIDs, ad.ID)
	}

	primaryAdID := rendered[0].ID

	for _, ad := range rendered {
		go s.TrackAdEvent(ad.ID, urlID, constants.AdEventImpression, ip, userAgent)
	}

	return rendered, renderedIDs, primaryAdID
}
