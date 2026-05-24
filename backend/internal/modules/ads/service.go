package ads

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"urlshortener/internal/modules/ads/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
)

type AdService struct {
	db   *sql.DB
	repo repository.Querier
}

func NewAdService(db *sql.DB, repo repository.Querier) *AdService {
	return &AdService{db: db, repo: repo}
}

func (s *AdService) getBaseCPM(ctx context.Context, adType string) (decimal.Decimal, error) {
	cpmStr, err := s.repo.GetCPMByAdType(ctx, strings.ToUpper(adType))
	if err != nil {
		return decimal.Zero, response.NewAppError(400, "Invalid ad type: "+adType)
	}
	return helper.ParseDecimal(cpmStr), nil
}

func (s *AdService) getCategoryMultiplier(ctx context.Context, category string) (decimal.Decimal, error) {
	multStr, err := s.repo.GetCategoryMultiplier(ctx, category)
	if err != nil {
		return decimal.Zero, response.NewAppError(400, "Invalid category: "+category)
	}
	return helper.ParseDecimal(multStr), nil
}

func (s *AdService) effectiveCPM(ctx context.Context, adType, category string) (decimal.Decimal, error) {
	base, err := s.getBaseCPM(ctx, adType)
	if err != nil {
		return decimal.Zero, err
	}
	mult, err := s.getCategoryMultiplier(ctx, category)
	if err != nil {
		return decimal.Zero, err
	}
	return base.Mul(mult), nil
}

func (s *AdService) ListCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	cats, err := s.repo.ListAdCategories(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list categories", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to list categories")
	}
	resp := make([]dto.CategoryResponse, len(cats))
	for i, c := range cats {
		resp[i] = dto.CategoryResponse{
			Category:   c.Category,
			Label:      c.Label,
			Multiplier: c.Multiplier,
		}
	}
	return resp, nil
}

func (s *AdService) ListAdTypes(ctx context.Context) ([]dto.AdTypeResponse, error) {
	types, err := s.repo.ListAdTypes(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list ad types", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to list ad types")
	}
	resp := make([]dto.AdTypeResponse, len(types))
	for i, t := range types {
		cpmVal := decimal.Zero
		if t.Cpm.Valid {
			cpmVal = helper.ParseDecimal(t.Cpm.String)
		}
		resp[i] = dto.AdTypeResponse{
			AdType:                t.AdType,
			CPM:                   cpmVal,
			Label:                 t.Label,
			AspectRatio:           helper.ParseDecimal(t.AspectRatio),
			RecommendedResolution: t.RecommendedResolution,
		}
	}
	return resp, nil
}

func (s *AdService) deductWalletForAdSpend(ctx context.Context, q repository.Querier, userID uuid.UUID, amount decimal.Decimal) error {
	wallet, err := q.GetWalletByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.NewAppError(404, "Wallet not found")
		}
		logger.Ctx(ctx).Error("Failed to retrieve wallet for ad spend", zap.Error(err))
		return response.NewAppError(500, "Failed to retrieve wallet")
	}

	balanceDec := helper.ParseDecimal(wallet.Balance)
	if balanceDec.LessThan(amount) {
		return response.NewAppError(400, "Insufficient wallet balance")
	}

	negAmount := amount.Neg()
	_, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
		UserID:  userID,
		Balance: helper.FormatDecimal(negAmount),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to deduct wallet balance for ad spend", zap.Error(err))
		return response.NewAppError(500, "Failed to deduct wallet balance")
	}

	_, err = q.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID: userID,
		Amount: helper.FormatDecimal(amount),
		Type:   constants.TxTypeAdSpend,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create spend transaction record for ad spend", zap.Error(err))
		return response.NewAppError(500, "Failed to create spend transaction record")
	}

	return nil
}

func (s *AdService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateAdRequest) (*dto.AdResponse, error) {
	totalBudget := decimal.NewFromFloat(req.TotalBudget)
	cpm, err := s.effectiveCPM(ctx, req.AdType, req.Category)
	if err != nil {
		return nil, err
	}

	var q repository.Querier = s.repo

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to start transaction for campaign creation", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to start transaction")
	}
	dbTx := tx
	defer func() {
		if r := recover(); r != nil {
			_ = dbTx.Rollback()
			panic(r)
		} else if err != nil {
			_ = dbTx.Rollback()
		}
	}()

	if queriesInstance, ok := s.repo.(*repository.Queries); ok {
		q = queriesInstance.WithTx(tx)
	} else {
		_ = dbTx.Rollback()
		logger.Ctx(ctx).Error("Repository is not transaction-compatible for campaign creation")
		return nil, response.NewAppError(500, "Repository is not transaction-compatible")
	}

	if err = s.deductWalletForAdSpend(ctx, q, userID, totalBudget); err != nil {
		return nil, err
	}

	ad, err := q.CreateAd(ctx, repository.CreateAdParams{
		AdvertiserID:    userID,
		Title:           req.Title,
		Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
		ImageUrl:        req.ImageURL,
		TargetUrl:       req.TargetURL,
		Category:        req.Category,
		TotalBudget:     helper.FormatDecimal(totalBudget),
		RemainingBudget: helper.FormatDecimal(totalBudget),
		Cpm:             helper.FormatDecimal(cpm),
		AdType:          req.AdType,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create ad campaign", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to create ad campaign")
	}

	if err = dbTx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit campaign creation transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to finalize campaign creation")
	}

	resp := dto.MapAdToResponse(ad)
	return &resp, nil
}

func (s *AdService) getAd(ctx context.Context, adID, userID uuid.UUID) (repository.Ad, error) {
	ad, err := s.repo.GetAdByID(ctx, adID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "not found" {
			return repository.Ad{}, response.NewAppError(404, "Ad campaign not found")
		}
		logger.Ctx(ctx).Error("Failed to get ad campaign", zap.Error(err))
		return repository.Ad{}, response.NewAppError(500, "Failed to get ad campaign")
	}
	if ad.AdvertiserID != userID {
		return repository.Ad{}, response.NewAppError(403, "You do not own this campaign")
	}
	return ad, nil
}

func (s *AdService) GetByID(ctx context.Context, adID, userID uuid.UUID) (*dto.AdResponse, error) {
	ad, err := s.getAd(ctx, adID, userID)
	if err != nil {
		return nil, err
	}
	resp := dto.MapAdToResponse(ad)
	return &resp, nil
}

func (s *AdService) ListByAdvertiser(ctx context.Context, userID uuid.UUID) ([]dto.AdResponse, error) {
	ads, err := s.repo.ListAdsByAdvertiser(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to list campaigns")
	}

	return dto.MapAdsToResponse(ads), nil
}

func (s *AdService) Update(ctx context.Context, adID, userID uuid.UUID, req dto.UpdateAdRequest) (*dto.AdResponse, error) {
	existing, err := s.getAd(ctx, adID, userID)
	if err != nil {
		return nil, err
	}

	title := existing.Title
	description := existing.Description
	imageURL := existing.ImageUrl
	targetURL := existing.TargetUrl
	category := existing.Category
	status := existing.Status
	totalBudget := existing.TotalBudget
	remainingBudget := existing.RemainingBudget
	cpm := existing.Cpm
	adType := existing.AdType

	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		description = sql.NullString{String: *req.Description, Valid: *req.Description != ""}
	}
	if req.ImageURL != nil {
		imageURL = *req.ImageURL
	}
	if req.TargetURL != nil {
		targetURL = *req.TargetURL
	}
	if req.Status != nil {
		status = *req.Status
	}

	ad, err := s.repo.UpdateAd(ctx, repository.UpdateAdParams{
		ID:              adID,
		Title:           title,
		Description:     description,
		ImageUrl:        imageURL,
		TargetUrl:       targetURL,
		Category:        category,
		TotalBudget:     totalBudget,
		RemainingBudget: remainingBudget,
		Cpm:             cpm,
		Status:          status,
		AdType:          adType,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update campaign", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to update campaign")
	}

	resp := dto.MapAdToResponse(ad)
	return &resp, nil
}

func (s *AdService) TopUp(ctx context.Context, adID, userID uuid.UUID, req dto.TopUpAdRequest) (*dto.AdResponse, error) {
	existing, err := s.getAd(ctx, adID, userID)
	if err != nil {
		return nil, err
	}

	topUpDec := decimal.NewFromFloat(req.Amount)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to start transaction for campaign top-up", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	var q *repository.Queries
	if queriesInstance, ok := s.repo.(*repository.Queries); ok {
		q = queriesInstance.WithTx(tx)
	} else {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Repository is not transaction-compatible for campaign top-up")
		return nil, response.NewAppError(500, "Repository is not transaction-compatible")
	}

	if err = s.deductWalletForAdSpend(ctx, q, userID, topUpDec); err != nil {
		return nil, err
	}

	totalDec := helper.ParseDecimal(existing.TotalBudget)
	remainingDec := helper.ParseDecimal(existing.RemainingBudget)

	newTotal := helper.FormatDecimal(totalDec.Add(topUpDec))
	newRemaining := helper.FormatDecimal(remainingDec.Add(topUpDec))

	ad, err := q.UpdateAd(ctx, repository.UpdateAdParams{
		ID:              adID,
		Title:           existing.Title,
		Description:     existing.Description,
		ImageUrl:        existing.ImageUrl,
		TargetUrl:       existing.TargetUrl,
		Category:        existing.Category,
		TotalBudget:     newTotal,
		RemainingBudget: newRemaining,
		Cpm:             existing.Cpm,
		Status:          existing.Status,
		AdType:          existing.AdType,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update campaign budget during top-up", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to update campaign budget")
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit campaign top-up transaction", zap.Error(err))
		return nil, response.NewAppError(500, "Failed to finalize campaign top-up")
	}

	resp := dto.MapAdToResponse(ad)
	return &resp, nil
}

func (s *AdService) Delete(ctx context.Context, adID, userID uuid.UUID) error {
	_, err := s.getAd(ctx, adID, userID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateAdStatus(ctx, repository.UpdateAdStatusParams{
		ID:     adID,
		Status: constants.AdStatusDeleted,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to delete campaign", zap.Error(err))
		return response.NewAppError(500, "Failed to delete campaign")
	}

	return nil
}

func (s *AdService) GetStats(ctx context.Context, adID, userID uuid.UUID) (*dto.AdStatsResponse, error) {
	_, err := s.getAd(ctx, adID, userID)
	if err != nil {
		return nil, err
	}

	stats, err := s.repo.GetAdEventStats(ctx, adID)
	if err != nil {
		stats = repository.GetAdEventStatsRow{}
	}

	return &dto.AdStatsResponse{
		AdID:        adID.String(),
		Impressions: stats.Impressions,
		Clicks:      stats.Clicks,
		Completions: stats.Completions,
	}, nil
}
