package redirect

import (
	"testing"

	"urlshortener/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGroupAds_Empty(t *testing.T) {
	g := GroupAds([]repository.Ad{})
	assert.Empty(t, g.Popup)
	assert.Empty(t, g.Banner)
	assert.Empty(t, g.Native)
	assert.Empty(t, g.Video)
	assert.Empty(t, g.Interstitial)
}

func TestGroupAds_AllTypes(t *testing.T) {
	ads := []repository.Ad{
		{ID: uuid.New(), AdType: "POPUP"},
		{ID: uuid.New(), AdType: "BANNER"},
		{ID: uuid.New(), AdType: "NATIVE"},
		{ID: uuid.New(), AdType: "VIDEO"},
		{ID: uuid.New(), AdType: "INTERSTITIAL"},
	}

	g := GroupAds(ads)
	assert.Len(t, g.Popup, 1)
	assert.Len(t, g.Banner, 1)
	assert.Len(t, g.Native, 1)
	assert.Len(t, g.Video, 1)
	assert.Len(t, g.Interstitial, 1)
}

func TestGroupAds_UnknownTypeIgnored(t *testing.T) {
	ads := []repository.Ad{
		{ID: uuid.New(), AdType: "POPUP"},
		{ID: uuid.New(), AdType: "UNKNOWN"},
	}

	g := GroupAds(ads)
	assert.Len(t, g.Popup, 1)
	assert.Empty(t, g.Banner)
	assert.Empty(t, g.Native)
	assert.Empty(t, g.Video)
	assert.Empty(t, g.Interstitial)
}

func TestGroupAds_MultipleOfSameType(t *testing.T) {
	ads := []repository.Ad{
		{ID: uuid.New(), AdType: "BANNER"},
		{ID: uuid.New(), AdType: "BANNER"},
		{ID: uuid.New(), AdType: "BANNER"},
	}

	g := GroupAds(ads)
	assert.Empty(t, g.Popup)
	assert.Len(t, g.Banner, 3)
	assert.Empty(t, g.Native)
}
