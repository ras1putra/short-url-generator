package dto

import (
	"fmt"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/repository"
)

type CreateURLRequest struct {
	URL               string   `json:"url" validate:"required,url"`
	CustomSlug        string   `json:"custom_slug,omitempty" validate:"omitempty,slug,min=3,max=20"`
	ExpiresValue      int      `json:"expires_value,omitempty" validate:"omitempty,min=1"`
	ExpiresUnit       string   `json:"expires_unit,omitempty" validate:"omitempty,oneof=minutes hours days"`
	IsMonetized       bool     `json:"is_monetized"`
	AllowedCategories []string `json:"allowed_categories,omitempty"`
}

type UpdateURLRequest struct {
	CustomSlug        string   `json:"custom_slug,omitempty" validate:"omitempty,slug,min=3,max=20"`
	ExpiresValue      int      `json:"expires_value,omitempty" validate:"omitempty,min=1"`
	ExpiresUnit       string   `json:"expires_unit,omitempty" validate:"omitempty,oneof=minutes hours days"`
	IsMonetized       *bool    `json:"is_monetized,omitempty"`
	AllowedCategories []string `json:"allowed_categories,omitempty"`
}

type URLResponse struct {
	Slug              string     `json:"slug"`
	ShortURL          string     `json:"short_url"`
	Original          string     `json:"original"`
	QRURL             string     `json:"qr_url"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	CreatedAt         string     `json:"created_at"`
	IsMonetized       bool       `json:"is_monetized"`
	AllowedCategories []string   `json:"allowed_categories,omitempty"`
}

func MapURLToResponse(url repository.Url, cfg *config.Config) URLResponse {
	resp := URLResponse{
		Slug:              url.Slug,
		ShortURL:          fmt.Sprintf("%s/%s", cfg.BaseURL, url.Slug),
		Original:          url.Original,
		QRURL:             fmt.Sprintf("%s/api/links/%s/qr", cfg.BaseURL, url.Slug),
		CreatedAt:         url.CreatedAt.Format(time.RFC3339),
		IsMonetized:       url.IsMonetized,
		AllowedCategories: url.AllowedCategories,
	}
	if url.ExpiresAt.Valid {
		resp.ExpiresAt = &url.ExpiresAt.Time
	}
	return resp
}

type CountryCount struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

type DateCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ListResponse struct {
	Links      []URLResponse `json:"links"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

type StatsResponse struct {
	TotalClicks     int64               `json:"total_clicks"`
	UniqueClicks    int64               `json:"unique_clicks"`
	ClicksPerDay    []DateCount         `json:"clicks_per_day"`
	TopCountries    []CountryCount      `json:"top_countries"`
	Browsers        map[string]int64    `json:"browsers"`
	Devices         map[string]int64    `json:"devices"`
	Events          []AdEventItem       `json:"events"`
	EventPagination *EventPaginationInfo `json:"event_pagination,omitempty"`
}

type EventPaginationInfo struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

type AdEventItem struct {
	Time            string `json:"time"`
	EventType       string `json:"event_type"`
	AdTitle         string `json:"ad_title"`
	AdType          string `json:"ad_type"`
	IsValid         bool   `json:"is_valid"`
	QualityScore    string `json:"quality_score"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type AdEventListResponse struct {
	Events     []AdEventItem `json:"events"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}
