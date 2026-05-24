package dto

import (
	"database/sql"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"urlshortener/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMapAdToResponse(t *testing.T) {
	now := time.Now()
	adID := uuid.New()
	advID := uuid.New()

	ad := repository.Ad{
		ID:              adID,
		AdvertiserID:    advID,
		Title:           "Test Ad",
		Description:     sql.NullString{String: "Description", Valid: true},
		ImageUrl:        "https://example.com/image.jpg",
		TargetUrl:       "https://example.com",
		Category:        "regular",
		TotalBudget:     "100.00",
		RemainingBudget: "80.00",
		Cpm:             "2.50",
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	resp := MapAdToResponse(ad)

	assert.Equal(t, adID.String(), resp.ID)
	assert.Equal(t, advID.String(), resp.AdvertiserID)
	assert.Equal(t, "Test Ad", resp.Title)
	assert.Equal(t, "Description", *resp.Description)
	assert.Equal(t, "https://example.com/image.jpg", resp.ImageURL)
	assert.Equal(t, "https://example.com", resp.TargetURL)
	assert.Equal(t, "regular", resp.Category)
	assert.True(t, decimal.NewFromFloat(100.00).Equal(resp.TotalBudget))
	assert.True(t, decimal.NewFromFloat(80.00).Equal(resp.RemainingBudget))
	assert.True(t, decimal.NewFromFloat(2.50).Equal(resp.CPM))
	assert.Equal(t, "active", resp.Status)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), resp.UpdatedAt)
}

func TestMapAdToResponse_NilDescription(t *testing.T) {
	ad := repository.Ad{
		ID:              uuid.New(),
		AdvertiserID:    uuid.New(),
		Title:           "No Desc",
		ImageUrl:        "https://example.com/img.jpg",
		TargetUrl:       "https://example.com",
		Category:        "crypto",
		TotalBudget:     "50.00",
		RemainingBudget: "50.00",
		Cpm:             "1.00",
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	resp := MapAdToResponse(ad)
	assert.Nil(t, resp.Description)
}

func TestMapAdsToResponse(t *testing.T) {
	ads := []repository.Ad{
		{ID: uuid.New(), AdvertiserID: uuid.New(), Title: "Ad 1", ImageUrl: "https://ex.com/1.jpg", TargetUrl: "https://ex.com/1", Category: "regular", TotalBudget: "10.00", RemainingBudget: "10.00", Cpm: "1.00", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), AdvertiserID: uuid.New(), Title: "Ad 2", ImageUrl: "https://ex.com/2.jpg", TargetUrl: "https://ex.com/2", Category: "gambling", TotalBudget: "20.00", RemainingBudget: "20.00", Cpm: "2.00", Status: "paused", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	resp := MapAdsToResponse(ads)
	assert.Len(t, resp, 2)
	assert.Equal(t, "Ad 1", resp[0].Title)
	assert.Equal(t, "Ad 2", resp[1].Title)
}

func TestMapAdsToResponse_Empty(t *testing.T) {
	resp := MapAdsToResponse([]repository.Ad{})
	assert.Len(t, resp, 0)
}
