package links

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	"urlshortener/internal/modules/links/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/logger"
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

func (s *URLService) getURLBySlug(ctx context.Context, slug string, userID uuid.UUID) (repository.Url, error) {
	url, err := s.repo.GetURLBySlug(ctx, slug)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get URL by slug", zap.String("slug", slug), zap.Error(err))
			return repository.Url{}, response.NewAppError(500, "Internal server error")
		}
		return repository.Url{}, response.NewAppError(404, "URL not found")
	}
	if url.UserID != userID {
		return repository.Url{}, response.NewAppError(403, "Forbidden")
	}
	return url, nil
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
			logger.Ctx(ctx).Error("Failed to check custom slug availability", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		finalSlug = req.CustomSlug
		isCustom = true
	} else {
		for i := 0; i < constants.MaxLinkSlugRetries; i++ {
			randSlug := slug.Generate(constants.DefaultSlugLength)
			if slug.IsReserved(randSlug) {
				continue
			}
			_, err := s.repo.GetURLBySlug(ctx, randSlug)
			if errors.Is(err, sql.ErrNoRows) {
				finalSlug = randSlug
				break
			}
			if err != nil {
				logger.Ctx(ctx).Error("Failed to check generated slug", zap.Error(err))
				return nil, response.NewAppError(500, "Internal server error")
			}
		}
		if finalSlug == "" {
			logger.Ctx(ctx).Error("Failed to generate unique slug after attempts", zap.Int("max_attempts", constants.MaxLinkSlugRetries))
			return nil, response.NewAppError(500, "Failed to generate unique slug")
		}
	}

	expiresAt, err := computeExpiresAt(req.ExpiresValue, req.ExpiresUnit)
	if err != nil {
		return nil, err
	}

	allowed := req.AllowedCategories
	if allowed == nil {
		allowed = []string{}
	}

	url, err := s.repo.CreateURL(ctx, repository.CreateURLParams{
		UserID:            userID,
		Slug:              finalSlug,
		Original:          req.URL,
		Custom:            isCustom,
		ExpiresAt:         expiresAt,
		IsMonetized:       req.IsMonetized,
		AllowedCategories: allowed,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to save URL to DB", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create URL")
	}

	resp := dto.MapURLToResponse(url, s.cfg)
	logger.Ctx(ctx).Info("URL created successfully",
		zap.String("slug", url.Slug),
		zap.String("original", url.Original),
		zap.Bool("custom", url.Custom),
	)
	return &resp, nil
}

func (s *URLService) GetBySlug(ctx context.Context, slug string) (*repository.Url, error) {
	result, err, _ := s.sfGroup.Do(slug, func() (interface{}, error) {
		cachedData, err := s.cache.Get(ctx, slug)
		if err == nil && cachedData != "" {
			var url repository.Url
			if json.Unmarshal([]byte(cachedData), &url) == nil {
				logger.Ctx(ctx).Debug("URL cache hit", zap.String("slug", slug))
				return &url, nil
			}
		}

		logger.Ctx(ctx).Debug("URL cache miss, fetching from DB", zap.String("slug", slug))
		url, err := s.repo.GetURLBySlug(ctx, slug)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				logger.Ctx(ctx).Error("Failed to get URL by slug", zap.String("slug", slug), zap.Error(err))
				return nil, response.NewAppError(500, "Internal server error")
			}
			return nil, response.NewAppError(404, "URL not found")
		}

		if url.ExpiresAt.Valid && url.ExpiresAt.Time.Before(time.Now()) {
			return nil, response.NewAppError(410, "URL expired")
		}

		if bytes, err := json.Marshal(url); err == nil {
			s.cache.Set(ctx, slug, bytes, constants.DefaultURLCacheTTL)
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
		logger.Ctx(ctx).Error("Failed to count URLs", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to fetch URLs")
	}

	offset := int32((page - 1) * perPage)
	urls, err := s.repo.ListURLsByUserPaginated(ctx, repository.ListURLsByUserPaginatedParams{
		UserID: userID,
		Limit:  int32(perPage),
		Offset: offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list URLs", zap.Error(err))
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
	url, err := s.getURLBySlug(ctx, slug, userID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteURL(ctx, repository.DeleteURLParams{ID: url.ID, UserID: userID}); err != nil {
		logger.Ctx(ctx).Error("Failed to delete URL", zap.Error(err))
		return response.NewAppError(500, "Failed to delete URL")
	}

	s.cache.Del(ctx, slug)
	logger.Ctx(ctx).Info("URL deleted successfully",
		zap.String("slug", slug),
		zap.String("url_id", url.ID.String()),
	)
	return nil
}

func (s *URLService) GetByID(ctx context.Context, userID uuid.UUID, slug string) (*dto.URLResponse, error) {
	url, err := s.getURLBySlug(ctx, slug, userID)
	if err != nil {
		return nil, err
	}

	resp := dto.MapURLToResponse(url, s.cfg)
	return &resp, nil
}

func (s *URLService) Update(ctx context.Context, userID uuid.UUID, currentSlug string, req dto.UpdateURLRequest) (*dto.URLResponse, error) {
	url, err := s.getURLBySlug(ctx, currentSlug, userID)
	if err != nil {
		return nil, err
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
			logger.Ctx(ctx).Error("Failed to check slug availability", zap.Error(err))
			return nil, response.NewAppError(500, "Internal server error")
		}
		newSlug = req.CustomSlug
	}

	var expiresAt sql.NullTime
	if req.ExpiresValue > 0 {
		expiresAt, err = computeExpiresAt(req.ExpiresValue, req.ExpiresUnit)
		if err != nil {
			return nil, err
		}
	} else {
		expiresAt = url.ExpiresAt
	}

	isMonetized := url.IsMonetized
	if req.IsMonetized != nil {
		isMonetized = *req.IsMonetized
	}

	allowedCats := url.AllowedCategories
	if req.AllowedCategories != nil {
		allowedCats = req.AllowedCategories
	}

	updated, err := s.repo.UpdateURL(ctx, repository.UpdateURLParams{
		ID:                url.ID,
		Slug:              newSlug,
		ExpiresAt:         expiresAt,
		UserID:            userID,
		IsMonetized:       isMonetized,
		AllowedCategories: allowedCats,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update URL", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to update URL")
	}

	if newSlug != url.Slug {
		s.cache.Del(ctx, url.Slug)
	}

	resp := dto.MapURLToResponse(updated, s.cfg)
	logger.Ctx(ctx).Info("URL updated successfully",
		zap.String("old_slug", url.Slug),
		zap.String("new_slug", updated.Slug),
		zap.String("url_id", url.ID.String()),
	)
	return &resp, nil
}

func (s *URLService) GetStats(ctx context.Context, userID uuid.UUID, slug string) (*dto.StatsResponse, error) {
	_, err := s.getURLBySlug(ctx, slug, userID)
	if err != nil {
		return nil, err
	}

	totalClicks, err := s.repo.GetTotalClicksBySlug(ctx, slug)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get total clicks", zap.String("slug", slug), zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	rows, err := s.repo.GetStatsBySlug(ctx, slug)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get stats", zap.String("slug", slug), zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	clickRows := clickRowsFromStatsRows(rows)

	uniqueClicks, err := s.repo.GetUniqueClicksBySlug(ctx, slug)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get unique clicks", zap.String("slug", slug), zap.Error(err))
		uniqueClicks = totalClicks
	}

	return buildStatsResponse(totalClicks, uniqueClicks, clickRows), nil
}

func (s *URLService) GetAggregateStats(ctx context.Context, userID uuid.UUID) (*dto.StatsResponse, error) {
	totalClicks, err := s.repo.GetTotalClicksByUser(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get aggregate total clicks", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	rows, err := s.repo.GetAggregateStatsByUser(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get aggregate stats", zap.Error(err))
		return nil, response.NewAppError(500, "Internal server error")
	}

	clickRows := clickRowsFromAggRows(rows)

	uniqueClicks, err := s.repo.GetUniqueClicksByUser(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get aggregate unique clicks", zap.Error(err))
		uniqueClicks = totalClicks
	}

	return buildStatsResponse(totalClicks, uniqueClicks, clickRows), nil
}

func (s *URLService) StartExpiryCleaner(ctx context.Context) {
	ticker := time.NewTicker(constants.ExpiryCleanerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cleanerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err := s.repo.DeleteExpiredURLs(cleanerCtx)
			cancel()
			if err != nil {
				logger.Ctx(cleanerCtx).Error("Expiry cleaner error", zap.Error(err))
				continue
			}
			logger.Ctx(cleanerCtx).Info("Expiry cleaner ran successfully")
		case <-ctx.Done():
			return
		}
	}
}
