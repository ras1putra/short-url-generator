package redirect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mssola/user_agent"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"

	"urlshortener/internal/analytics"
	"urlshortener/internal/repository"
	"urlshortener/pkg/logger"
)

type URLGetter interface {
	GetBySlug(ctx context.Context, slug string) (*repository.Url, error)
}

type RedirectHandler struct {
	urlSvc  URLGetter
	worker *analytics.AnalyticsWorker
	geoDB  *geoip2.Reader
}

func NewRedirectHandler(urlSvc URLGetter, worker *analytics.AnalyticsWorker, geoDB *geoip2.Reader) *RedirectHandler {
	return &RedirectHandler{urlSvc: urlSvc, worker: worker, geoDB: geoDB}
}

func (h *RedirectHandler) Redirect(c *fiber.Ctx) error {
	slug := c.Params("slug")

	url, err := h.urlSvc.GetBySlug(c.Context(), slug)
	if err != nil {
		return err
	}

	ip := c.IP()
	userAgent := c.Get("User-Agent")
	referer := c.Get("Referer")

	h.enqueueClick(url.ID, ip, userAgent, referer)

	logger.WithFields(
		zap.String("slug", slug),
		zap.String("ip", ip),
		zap.String("user_agent", userAgent),
	).Info("Redirected",
		zap.String("original", url.Original),
	)

	return c.Redirect(url.Original, fiber.StatusFound)
}

func (h *RedirectHandler) enqueueClick(urlID uuid.UUID, ip, userAgent, referer string) {
	ua := user_agent.New(userAgent)
	browser, _ := ua.Browser()

	device := "desktop"
	if ua.Mobile() {
		device = "mobile"
	}
	if ua.Bot() {
		device = "bot"
	}

	country, city := h.resolveGeo(ip)

	hash := sha256.Sum256([]byte(ip))
	ipHash := hex.EncodeToString(hash[:])

	h.worker.Enqueue(analytics.ClickEvent{
		UrlID:    urlID,
		IPHash:   ipHash,
		Country:  country,
		City:     city,
		Device:   device,
		Browser:  browser,
		Referrer: referer,
	})
}

func (h *RedirectHandler) resolveGeo(ip string) (country, city string) {
	if h.geoDB == nil || ip == "" {
		return
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		logger.WithFields(zap.String("ip", ip)).Warn("GeoIP lookup failed: invalid IP")
		return
	}

	record, err := h.geoDB.City(parsedIP)
	if err != nil {
		logger.WithFields(zap.String("ip", ip), zap.Error(err)).Warn("GeoIP lookup failed")
		return
	}

	country = record.Country.IsoCode
	city = record.City.Names["en"]
	return
}
