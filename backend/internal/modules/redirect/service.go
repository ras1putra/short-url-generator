package redirect

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	urlSvc      URLGetter
	repo        repository.Querier
	worker      *analytics.AnalyticsWorker
	geoDB       *geoip2.Reader
	hmacSecret  []byte
	db          *sql.DB
	redis       *redis.Client
	qualityCfg  qualityConfig
}

type qualityConfig struct {
	minScore float64
	minSessionMs int64
}

func NewRedirectService(urlSvc URLGetter, repo repository.Querier, worker *analytics.AnalyticsWorker, geoDB *geoip2.Reader, cfg *config.Config, db *sql.DB, redisClient *redis.Client) *RedirectService {
	return &RedirectService{
		urlSvc:      urlSvc,
		repo:        repo,
		worker:      worker,
		geoDB:       geoDB,
		hmacSecret:  []byte(cfg.BridgeHMACSecret),
		db:          db,
		redis:       redisClient,
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
	return ads, err
}



func (s *RedirectService) GetAdByID(ctx context.Context, adID uuid.UUID) (*repository.Ad, error) {
	ad, err := s.repo.GetAdByID(ctx, adID)
	if err != nil {
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

	qualityScore := "1.00"
	if eventType == "SKIP" {
		qualityScore = "0.50"
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

func (s *RedirectService) GenerateBridgeToken(slug string, adID uuid.UUID) string {
	now := time.Now().UnixMilli()
	payload := fmt.Sprintf("%s:%s:%d", slug, adID.String(), now)
	mac := hmac.New(sha256.New, s.hmacSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s:%s:%d:%s", slug, adID.String(), now, sig)
}

func (s *RedirectService) VerifyBridgeToken(token string) (string, uuid.UUID, bool, int64) {
	parts := strings.SplitN(token, ":", 4)
	if len(parts) != 4 {
		return "", uuid.Nil, false, 0
	}
	slug := parts[0]
	adIDStr := parts[1]
	ts := parts[2]
	sig := parts[3]

	adID, err := uuid.Parse(adIDStr)
	if err != nil {
		return "", uuid.Nil, false, 0
	}

	mac := hmac.New(sha256.New, s.hmacSecret)
	mac.Write([]byte(slug + ":" + adIDStr + ":" + ts))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", uuid.Nil, false, 0
	}

	var ms int64
	fmt.Sscanf(ts, "%d", &ms)
	return slug, adID, true, ms
}

func (s *RedirectService) runQualityChecks(ctx context.Context, slug, adID, ip, fingerprint string, honeypotHit bool, mouseMoves int, issuedMs int64) qualityResult {
	elapsed := time.Now().UnixMilli() - issuedMs

	if honeypotHit {
		logger.Ctx(ctx).Warn("Quality check failed: honeypot hit", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip))
		return qualityResult{Score: 0, RejectReason: "HONEYPOT_HIT", IsValid: false}
	}

	if elapsed < s.qualityCfg.minSessionMs {
		logger.Ctx(ctx).Warn("Quality check failed: session too fast", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip), zap.Int64("elapsed_ms", elapsed), zap.Int64("min_session_ms", s.qualityCfg.minSessionMs))
		return qualityResult{Score: 0, RejectReason: "TOO_FAST", IsValid: false}
	}

	if mouseMoves < 2 {
		logger.Ctx(ctx).Warn("Quality check failed: insufficient mouse movements", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip), zap.Int("mouse_moves", mouseMoves))
		return qualityResult{Score: 0, RejectReason: "NO_MOUSE_MOVEMENT", IsValid: false}
	}

	if s.redis != nil {
		ipKey := constants.RedisPrefixQualityIP + slug + ":" + ip
		exists, _ := s.redis.Exists(ctx, ipKey).Result()
		if exists > 0 {
			logger.Ctx(ctx).Warn("Quality check failed: duplicate IP session within 24h", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("ip", ip))
			return qualityResult{Score: 0, RejectReason: "DUPLICATE_IP", IsValid: false}
		}

		if fingerprint != "" {
			fpKey := constants.RedisPrefixQualityFP + slug + ":" + fingerprint
			exists, _ := s.redis.Exists(ctx, fpKey).Result()
			if exists > 0 {
				logger.Ctx(ctx).Warn("Quality check failed: duplicate browser fingerprint within 24h", zap.String("slug", slug), zap.String("ad_id", adID), zap.String("fingerprint", fingerprint))
				return qualityResult{Score: 0, RejectReason: "DUPLICATE_FINGERPRINT", IsValid: false}
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

func (s *RedirectService) ChargeForCompletion(ctx context.Context, slug string, adID, linkID uuid.UUID, linkOwnerID uuid.UUID, ip, userAgent string, issuedMs int64, req CompletionRequest) error {
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

	quality := s.runQualityChecks(ctx, slug, adID.String(), ip, req.Fingerprint, req.HoneypotHit, req.MouseMoves, issuedMs)

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
		EventType:       "COMPLETION",
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
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

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
