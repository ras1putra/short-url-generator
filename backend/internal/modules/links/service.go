package links

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	"urlshortener/internal/modules/links/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/response"
	"urlshortener/pkg/slug"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

type URLService struct {
	repo    repository.Querier
	cache   cache.Cacher
	cfg     *config.Config
	sfGroup *singleflight.Group
}

func NewURLService(repo repository.Querier, cache cache.Cacher, cfg *config.Config) *URLService {
	return &URLService{repo: repo, cache: cache, cfg: cfg, sfGroup: &singleflight.Group{}}
}

func (s *URLService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateURLRequest) (*dto.URLResponse, error) {
	var finalSlug string
	var isCustom bool

	if req.CustomSlug != "" {
		if slug.IsReserved(req.CustomSlug) {
			return nil, response.NewAppError(409, "This slug is reserved and cannot be used")
		}
		_, err := s.repo.GetURLBySlug(ctx, req.CustomSlug)
		if err == nil {
			return nil, response.NewAppError(409, "Custom slug already taken")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			zap.L().Error("Failed to check custom slug availability", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		finalSlug = req.CustomSlug
		isCustom = true
	} else {
		for i := 0; i < 5; i++ {
			randSlug := slug.Generate(6)
			if slug.IsReserved(randSlug) {
				continue
			}
			_, err := s.repo.GetURLBySlug(ctx, randSlug)
			if errors.Is(err, sql.ErrNoRows) {
				finalSlug = randSlug
				break
			}
			if err != nil {
				zap.L().Error("Failed to check generated slug", zap.Error(err))
				return nil, response.NewAppError(500, "Internal server error")
			}
		}
		if finalSlug == "" {
			return nil, response.NewAppError(500, "Failed to generate unique slug")
		}
	}

	var expiresAt sql.NullTime
	if req.ExpiresValue > 0 {
		expiresAt.Valid = true
		switch req.ExpiresUnit {
		case "minutes":
			expiresAt.Time = time.Now().Add(time.Duration(req.ExpiresValue) * time.Minute)
		case "hours":
			expiresAt.Time = time.Now().Add(time.Duration(req.ExpiresValue) * time.Hour)
		default:
			expiresAt.Time = time.Now().AddDate(0, 0, req.ExpiresValue)
		}
	}

	url, err := s.repo.CreateURL(ctx, repository.CreateURLParams{
		UserID:    userID,
		Slug:      finalSlug,
		Original:  req.URL,
		Custom:    isCustom,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		zap.L().Error("Failed to save URL to DB", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create URL")
	}

	resp := dto.MapURLToResponse(url, s.cfg)
	return &resp, nil
}

func (s *URLService) GetBySlug(ctx context.Context, slug string) (*repository.Url, error) {
	result, err, _ := s.sfGroup.Do(slug, func() (interface{}, error) {
		cachedData, err := s.cache.Get(ctx, slug)
		if err == nil && cachedData != "" {
			var url repository.Url
			if json.Unmarshal([]byte(cachedData), &url) == nil {
				return &url, nil
			}
		}

		url, err := s.repo.GetURLBySlug(ctx, slug)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				zap.L().Error("Failed to get URL by slug", zap.String("slug", slug), zap.Error(err))
				return nil, response.NewAppError(500, "Internal server error")
			}
			return nil, response.NewAppError(404, "URL not found")
		}

		if url.ExpiresAt.Valid && url.ExpiresAt.Time.Before(time.Now()) {
			return nil, response.NewAppError(410, "URL expired")
		}

		if bytes, err := json.Marshal(url); err == nil {
			s.cache.Set(ctx, slug, bytes, 24*time.Hour)
		}

		return &url, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(*repository.Url), nil
}

func (s *URLService) ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int) (*dto.ListResponse, error) {
	total, err := s.repo.CountURLsByUser(ctx, userID)
	if err != nil {
		zap.L().Error("Failed to count URLs", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to fetch URLs")
	}

	offset := int32((page - 1) * perPage)
	urls, err := s.repo.ListURLsByUserPaginated(ctx, repository.ListURLsByUserPaginatedParams{
		UserID: userID,
		Limit:  int32(perPage),
		Offset: offset,
	})
	if err != nil {
		zap.L().Error("Failed to list URLs", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to fetch URLs")
	}

	result := make([]dto.URLResponse, len(urls))
	for i, url := range urls {
		result[i] = dto.MapURLToResponse(url, s.cfg)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &dto.ListResponse{
		Links:      result,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *URLService) Delete(ctx context.Context, userID uuid.UUID, slug string) error {
	url, err := s.repo.GetURLBySlug(ctx, slug)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			zap.L().Error("Failed to get URL for deletion", zap.String("slug", slug), zap.Error(err))
			return response.NewAppError(500, "Internal server error")
		}
		return response.NewAppError(404, "URL not found")
	}

	if url.UserID != userID {
		return response.NewAppError(403, "Forbidden")
	}

	if err := s.repo.DeleteURL(ctx, repository.DeleteURLParams{ID: url.ID, UserID: userID}); err != nil {
		zap.L().Error("Failed to delete URL", zap.Error(err))
		return response.NewAppError(500, "Failed to delete URL")
	}

	s.cache.Del(ctx, slug)
	return nil
}

func (s *URLService) GetByID(ctx context.Context, userID uuid.UUID, slug string) (*dto.URLResponse, error) {
	url, err := s.repo.GetURLBySlug(ctx, slug)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			zap.L().Error("Failed to get URL by slug", zap.String("slug", slug), zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		return nil, response.NewAppError(404, "URL not found")
	}

	if url.UserID != userID {
		return nil, response.NewAppError(403, "Forbidden")
	}

	resp := dto.MapURLToResponse(url, s.cfg)
	return &resp, nil
}

func (s *URLService) Update(ctx context.Context, userID uuid.UUID, currentSlug string, req dto.UpdateURLRequest) (*dto.URLResponse, error) {
	url, err := s.repo.GetURLBySlug(ctx, currentSlug)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			zap.L().Error("Failed to get URL for update", zap.String("slug", currentSlug), zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		return nil, response.NewAppError(404, "URL not found")
	}

	if url.UserID != userID {
		return nil, response.NewAppError(403, "Forbidden")
	}

	newSlug := url.Slug
	if req.CustomSlug != "" && req.CustomSlug != url.Slug {
		if slug.IsReserved(req.CustomSlug) {
			return nil, response.NewAppError(409, "This slug is reserved and cannot be used")
		}
		_, err := s.repo.GetURLBySlug(ctx, req.CustomSlug)
		if err == nil {
			return nil, response.NewAppError(409, "Custom slug already taken")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			zap.L().Error("Failed to check slug availability", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		newSlug = req.CustomSlug
	}

	var expiresAt sql.NullTime
	if req.ExpiresValue > 0 {
		expiresAt.Valid = true
		switch req.ExpiresUnit {
		case "minutes":
			expiresAt.Time = time.Now().Add(time.Duration(req.ExpiresValue) * time.Minute)
		case "hours":
			expiresAt.Time = time.Now().Add(time.Duration(req.ExpiresValue) * time.Hour)
		default:
			expiresAt.Time = time.Now().AddDate(0, 0, req.ExpiresValue)
		}
	} else {
		expiresAt = url.ExpiresAt
	}

	updated, err := s.repo.UpdateURL(ctx, repository.UpdateURLParams{
		ID:        url.ID,
		Slug:      newSlug,
		ExpiresAt: expiresAt,
		UserID:    userID,
	})
	if err != nil {
		zap.L().Error("Failed to update URL", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to update URL")
	}

	if newSlug != url.Slug {
		s.cache.Del(ctx, url.Slug)
	}

	resp := dto.MapURLToResponse(updated, s.cfg)
	return &resp, nil
}

func (s *URLService) GetStats(ctx context.Context, userID uuid.UUID, slug string) (*dto.StatsResponse, error) {
	url, err := s.repo.GetURLBySlug(ctx, slug)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			zap.L().Error("Failed to get URL for stats", zap.String("slug", slug), zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		return nil, response.NewAppError(404, "URL not found")
	}

	if url.UserID != userID {
		return nil, response.NewAppError(403, "Forbidden")
	}

	totalClicks, err := s.repo.GetTotalClicksBySlug(ctx, slug)
	if err != nil {
		zap.L().Error("Failed to get total clicks", zap.String("slug", slug), zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	rows, err := s.repo.GetStatsBySlug(ctx, slug)
	if err != nil {
		zap.L().Error("Failed to get stats", zap.String("slug", slug), zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	browsers := make(map[string]int64)
	devices := make(map[string]int64)
	countryMap := make(map[string]int64)
	dateMap := make(map[string]int64)

	for _, row := range rows {
		if row.Browser.Valid && row.Browser.String != "" {
			browsers[row.Browser.String] += row.ClickCount
		}
		if row.Device.Valid && row.Device.String != "" {
			devices[row.Device.String] += row.ClickCount
		}
		if row.Country.Valid && row.Country.String != "" {
			countryMap[row.Country.String] += row.ClickCount
		} else {
			countryMap["Unknown"] += row.ClickCount
		}
		dateStr := row.ClickDate.Format("2006-01-02")
		dateMap[dateStr] += row.ClickCount
	}

	topCountries := make([]dto.CountryCount, 0, len(countryMap))
	for c, cnt := range countryMap {
		topCountries = append(topCountries, dto.CountryCount{Country: c, Count: cnt})
	}
	sort.Slice(topCountries, func(i, j int) bool {
		return topCountries[i].Count > topCountries[j].Count
	})
	if len(topCountries) > 10 {
		topCountries = topCountries[:10]
	}

	clicksPerDay := make([]dto.DateCount, 0, len(dateMap))
	for d, cnt := range dateMap {
		clicksPerDay = append(clicksPerDay, dto.DateCount{Date: d, Count: cnt})
	}
	sort.Slice(clicksPerDay, func(i, j int) bool {
		return clicksPerDay[i].Date > clicksPerDay[j].Date
	})

	uniqueClicks, err := s.repo.GetUniqueClicksBySlug(ctx, slug)
	if err != nil {
		zap.L().Error("Failed to get unique clicks", zap.String("slug", slug), zap.Error(err))
		uniqueClicks = totalClicks
	}

	return &dto.StatsResponse{
		TotalClicks:  totalClicks,
		UniqueClicks: uniqueClicks,
		ClicksPerDay: clicksPerDay,
		TopCountries: topCountries,
		Browsers:     browsers,
		Devices:      devices,
	}, nil
}

func (s *URLService) GetAggregateStats(ctx context.Context, userID uuid.UUID) (*dto.StatsResponse, error) {
	totalClicks, err := s.repo.GetTotalClicksByUser(ctx, userID)
	if err != nil {
		zap.L().Error("Failed to get aggregate total clicks", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	rows, err := s.repo.GetAggregateStatsByUser(ctx, userID)
	if err != nil {
		zap.L().Error("Failed to get aggregate stats", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	browsers := make(map[string]int64)
	devices := make(map[string]int64)
	countryMap := make(map[string]int64)
	dateMap := make(map[string]int64)

	for _, row := range rows {
		if row.Browser.Valid && row.Browser.String != "" {
			browsers[row.Browser.String] += row.ClickCount
		}
		if row.Device.Valid && row.Device.String != "" {
			devices[row.Device.String] += row.ClickCount
		}
		if row.Country.Valid && row.Country.String != "" {
			countryMap[row.Country.String] += row.ClickCount
		} else {
			countryMap["Unknown"] += row.ClickCount
		}
		dateStr := row.ClickDate.Format("2006-01-02")
		dateMap[dateStr] += row.ClickCount
	}

	topCountries := make([]dto.CountryCount, 0, len(countryMap))
	for c, cnt := range countryMap {
		topCountries = append(topCountries, dto.CountryCount{Country: c, Count: cnt})
	}
	sort.Slice(topCountries, func(i, j int) bool {
		return topCountries[i].Count > topCountries[j].Count
	})
	if len(topCountries) > 10 {
		topCountries = topCountries[:10]
	}

	clicksPerDay := make([]dto.DateCount, 0, len(dateMap))
	for d, cnt := range dateMap {
		clicksPerDay = append(clicksPerDay, dto.DateCount{Date: d, Count: cnt})
	}
	sort.Slice(clicksPerDay, func(i, j int) bool {
		return clicksPerDay[i].Date > clicksPerDay[j].Date
	})

	uniqueClicks, err := s.repo.GetUniqueClicksByUser(ctx, userID)
	if err != nil {
		zap.L().Error("Failed to get aggregate unique clicks", zap.Error(err))
		uniqueClicks = totalClicks
	}

	return &dto.StatsResponse{
		TotalClicks:  totalClicks,
		UniqueClicks: uniqueClicks,
		ClicksPerDay: clicksPerDay,
		TopCountries: topCountries,
		Browsers:     browsers,
		Devices:      devices,
	}, nil
}

func (s *URLService) StartExpiryCleaner(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cleanerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err := s.repo.DeleteExpiredURLs(cleanerCtx)
			cancel()
			if err != nil {
				zap.L().Error("Expiry cleaner error", zap.Error(err))
				continue
			}
			zap.L().Info("Expiry cleaner ran successfully")
		case <-ctx.Done():
			return
		}
	}
}
