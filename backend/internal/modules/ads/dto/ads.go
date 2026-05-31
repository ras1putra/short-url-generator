package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"urlshortener/internal/repository"
	"urlshortener/pkg/helper"
)

type CreateAdRequest struct {
	Title       string  `json:"title" validate:"required,min=3,max=255"`
	Description string  `json:"description,omitempty"`
	ImageURL    string  `json:"image_url" validate:"required,url"`
	TargetURL   string  `json:"target_url" validate:"required,url"`
	Category    string  `json:"category" validate:"required"`
	TotalBudget float64 `json:"total_budget" validate:"required,min=1"`
	AdType      string  `json:"ad_type" validate:"required"`
}

type UpdateAdRequest struct {
	Title       *string  `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
	Description *string  `json:"description,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty" validate:"omitempty,url"`
	TargetURL   *string  `json:"target_url,omitempty" validate:"omitempty,url"`
	Status      *string  `json:"status,omitempty" validate:"omitempty,oneof=active paused"`
}

type TopUpAdRequest struct {
	Amount float64 `json:"amount" validate:"required,min=0.01"`
}

type AdResponse struct {
	ID              string          `json:"id"`
	AdvertiserID    string          `json:"advertiser_id"`
	Title           string          `json:"title"`
	Description     *string         `json:"description,omitempty"`
	ImageURL        string          `json:"image_url"`
	TargetURL       string          `json:"target_url"`
	Category        string          `json:"category"`
	TotalBudget     decimal.Decimal `json:"total_budget"`
	RemainingBudget decimal.Decimal `json:"remaining_budget"`
	CPM             decimal.Decimal `json:"cpm"`
	Status          string          `json:"status"`
	AdType          string          `json:"ad_type"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

type CategoryResponse struct {
	Category   string `json:"category"`
	Label      string `json:"label"`
	Multiplier string `json:"multiplier"`
}

type AdTypeResponse struct {
	AdType                string          `json:"ad_type"`
	CPM                   decimal.Decimal `json:"cpm"`
	Label                 string          `json:"label"`
	AspectRatio           decimal.Decimal `json:"aspect_ratio"`
	RecommendedResolution string          `json:"recommended_resolution"`
}

type CampaignListResponse struct {
	Campaigns  []AdResponse `json:"campaigns"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	TotalPages int          `json:"total_pages"`
}

type AdStatsResponse struct {
	AdID              string  `json:"ad_id"`
	Impressions       int64   `json:"impressions"`
	Clicks            int64   `json:"clicks"`
	Completions       int64   `json:"completions"`
	ValidCompletions  int64   `json:"valid_completions"`
	InvalidCompletions int64  `json:"invalid_completions"`
	Skips             int64   `json:"skips"`
	AvgQualityScore   float64 `json:"avg_quality_score"`
}

func MapAdToResponse(ad repository.Ad) AdResponse {
	resp := AdResponse{
		ID:              ad.ID.String(),
		AdvertiserID:    ad.AdvertiserID.String(),
		Title:           ad.Title,
		ImageURL:        ad.ImageUrl,
		TargetURL:       ad.TargetUrl,
		Category:        ad.Category,
		TotalBudget:     helper.ParseDecimal(ad.TotalBudget),
		RemainingBudget: helper.ParseDecimal(ad.RemainingBudget),
		CPM:             helper.ParseDecimal(ad.Cpm),
		Status:          ad.Status,
		AdType:          ad.AdType,
		CreatedAt:       ad.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       ad.UpdatedAt.Format(time.RFC3339),
	}
	if ad.Description.Valid {
		resp.Description = &ad.Description.String
	}
	return resp
}

func MapAdsToResponse(ads []repository.Ad) []AdResponse {
	resp := make([]AdResponse, len(ads))
	for i, ad := range ads {
		resp[i] = MapAdToResponse(ad)
	}
	return resp
}

