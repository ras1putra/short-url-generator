package redirect

import (
	"database/sql"
	"testing"

	"urlshortener/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAdMediaTag_ImageJPG(t *testing.T) {
	result := adMediaTag("https://example.com/img.jpg", "BANNER")
	assert.Contains(t, result, `<img src="https://example.com/img.jpg"`)
	assert.NotContains(t, result, "<video")
}

func TestAdMediaTag_ImagePNG(t *testing.T) {
	result := adMediaTag("https://example.com/img.png", "POPUP")
	assert.Contains(t, result, `<img src="https://example.com/img.png"`)
}

func TestAdMediaTag_VideoByType(t *testing.T) {
	result := adMediaTag("https://example.com/ad.mp4", "VIDEO")
	assert.Contains(t, result, `<video src="https://example.com/ad.mp4"`)
	assert.Contains(t, result, "autoplay")
	assert.Contains(t, result, "playsinline")
}

func TestAdMediaTag_VideoByExtension(t *testing.T) {
	result := adMediaTag("https://example.com/video.webm", "BANNER")
	assert.Contains(t, result, `<video src="https://example.com/video.webm"`)
}

func TestAdMediaTag_MP4Extension(t *testing.T) {
	result := adMediaTag("https://example.com/movie.mp4", "BANNER")
	assert.Contains(t, result, `<video src="https://example.com/movie.mp4"`)
}

func TestPopupOverlay(t *testing.T) {
	ad := &repository.Ad{
		ID:       uuid.New(),
		ImageUrl: "https://example.com/img.jpg",
		AdType:   "POPUP",
	}
	slug := "abc123"
	result := popupOverlay(ad, slug)
	assert.Contains(t, result, `id="popup-overlay"`)
	assert.Contains(t, result, `Skip in 10s`)
	assert.Contains(t, result, slug+"/click/"+ad.ID.String())
	assert.Contains(t, result, `<img src="https://example.com/img.jpg"`)
}

func TestPopUpPlaceholder(t *testing.T) {
	result := popupPlaceholder()
	assert.Contains(t, result, `id="popup-overlay"`)
	assert.Contains(t, result, `Boost Your Reach Instantly`)
	assert.Contains(t, result, `Start Advertising`)
}

func TestBannerStrip(t *testing.T) {
	ad := &repository.Ad{
		ID:       uuid.New(),
		ImageUrl: "https://example.com/banner.jpg",
		AdType:   "BANNER",
	}
	slug := "slug123"
	result := bannerStrip(ad, slug)
	assert.Contains(t, result, `class="banner-strip"`)
	assert.Contains(t, result, slug+"/click/"+ad.ID.String())
}

func TestBannerPlaceholder(t *testing.T) {
	result := bannerPlaceholder()
	assert.Contains(t, result, `Advertise Here`)
}

func TestNativeCard_WithDescription(t *testing.T) {
	ad := &repository.Ad{
		ID:          uuid.New(),
		Title:       "Test Ad",
		Description: sql.NullString{String: "Great offer", Valid: true},
		ImageUrl:    "https://example.com/native.jpg",
		AdType:      "NATIVE",
	}
	slug := "nat123"
	result := nativeCard(ad, slug)
	assert.Contains(t, result, `Test Ad`)
	assert.Contains(t, result, `Great offer`)
	assert.Contains(t, result, slug+"/click/"+ad.ID.String())
}

func TestNativeCard_NoDescription(t *testing.T) {
	ad := &repository.Ad{
		ID:       uuid.New(),
		Title:    "Test Ad",
		ImageUrl: "https://example.com/native.jpg",
		AdType:   "NATIVE",
	}
	slug := "nat123"
	result := nativeCard(ad, slug)
	assert.Contains(t, result, `Sponsored content`)
}

func TestNativePlaceholder(t *testing.T) {
	result := nativePlaceholder()
	assert.Contains(t, result, `Advertise Here`)
}

func TestVideoSection(t *testing.T) {
	ad := &repository.Ad{
		ID:       uuid.New(),
		Title:    "Video Ad",
		ImageUrl: "https://example.com/vid.mp4",
		AdType:   "VIDEO",
	}
	slug := "vid123"
	result := videoSection(ad, slug)
	assert.Contains(t, result, `<video src="https://example.com/vid.mp4"`)
	assert.Contains(t, result, `Video Ad`)
	assert.Contains(t, result, slug+"/click/"+ad.ID.String())
}

func TestVideoPlaceholder(t *testing.T) {
	result := videoPlaceholder()
	assert.Contains(t, result, `Advertise Here`)
}

func TestInterstSection_WithDescription(t *testing.T) {
	ad := &repository.Ad{
		ID:          uuid.New(),
		Title:       "Interstitial",
		Description: sql.NullString{String: "Full screen ad", Valid: true},
		ImageUrl:    "https://example.com/inter.jpg",
		AdType:      "INTERSTITIAL",
	}
	slug := "int123"
	result := interstSection(ad, slug)
	assert.Contains(t, result, `Interstitial`)
	assert.Contains(t, result, `Full screen ad`)
	assert.Contains(t, result, slug+"/click/"+ad.ID.String())
}

func TestInterstSection_NoDescription(t *testing.T) {
	ad := &repository.Ad{
		ID:       uuid.New(),
		Title:    "Interstitial",
		ImageUrl: "https://example.com/inter.jpg",
		AdType:   "INTERSTITIAL",
	}
	slug := "int123"
	result := interstSection(ad, slug)
	assert.Contains(t, result, `Sponsored`)
}

func TestInterstPlaceholder(t *testing.T) {
	result := interstPlaceholder()
	assert.Contains(t, result, `Advertise Here`)
}

func TestRenderInterstitial_NoAds(t *testing.T) {
	url := repository.Url{
		Slug:     "test123",
		Original: "https://example.com/dest",
	}
	result := RenderInterstitial([]repository.Ad{}, url, "bridge-token-xyz", uuid.Nil)
	assert.Contains(t, result, `https://example.com/dest`)
	assert.Contains(t, result, `bridge-token-xyz`)
	assert.Contains(t, result, `Boost Your Reach Instantly`)
	assert.Contains(t, result, `Start Advertising`)
}

func TestRenderInterstitial_WithAds(t *testing.T) {
	ad := repository.Ad{
		ID:          uuid.New(),
		Title:       "Test Ad",
		Description: sql.NullString{String: "Desc", Valid: true},
		ImageUrl:    "https://example.com/img.jpg",
		TargetUrl:   "https://example.com/land",
		AdType:      "BANNER",
		Category:    "regular",
	}
	url := repository.Url{
		Slug:     "test123",
		Original: "https://example.com/dest",
	}
	result := RenderInterstitial([]repository.Ad{ad}, url, "bridge-token-xyz", uuid.Nil)
	assert.Contains(t, result, `https://example.com/dest`)
	assert.Contains(t, result, `bridge-token-xyz`)
	assert.Contains(t, result, `https://example.com/img.jpg`)
	assert.Contains(t, result, ad.ID.String())
}

func TestRenderInterstitial_WithPopupAsPrimary(t *testing.T) {
	popupID := uuid.New()
	bannerID := uuid.New()
	ads := []repository.Ad{
		{ID: popupID, Title: "Popup Ad", Description: sql.NullString{Valid: false}, ImageUrl: "https://ex.com/popup.jpg", AdType: "POPUP", Category: "regular"},
		{ID: bannerID, Title: "Banner Ad", Description: sql.NullString{Valid: false}, ImageUrl: "https://ex.com/banner.jpg", AdType: "BANNER", Category: "regular"},
	}
	url := repository.Url{
		Slug:     "pop123",
		Original: "https://example.com/dest",
	}
	result := RenderInterstitial(ads, url, "bridge-token-xyz", popupID)
	assert.Contains(t, result, `popup-overlay`)
	assert.Contains(t, result, `https://ex.com/popup.jpg`)
	assert.Contains(t, result, popupID.String())
	assert.Contains(t, result, bannerID.String())
	assert.Contains(t, result, `bridge-token-xyz`)
}

func TestRenderInterstitial_WithVideoAd(t *testing.T) {
	ads := []repository.Ad{
		{ID: uuid.New(), Title: "Vid", Description: sql.NullString{Valid: false}, ImageUrl: "https://ex.com/vid.mp4", AdType: "VIDEO", Category: "regular"},
	}
	url := repository.Url{
		Slug:     "vid123",
		Original: "https://example.com/dest",
	}
	result := RenderInterstitial(ads, url, "bridge-token-xyz", uuid.Nil)
	assert.Contains(t, result, `video`)
	assert.Contains(t, result, `Vid`)
}

func TestRenderInterstitial_WithNativeAd(t *testing.T) {
	ads := []repository.Ad{
		{ID: uuid.New(), Title: "Native", Description: sql.NullString{String: "Native desc", Valid: true}, ImageUrl: "https://ex.com/nat.jpg", AdType: "NATIVE", Category: "regular"},
	}
	url := repository.Url{
		Slug:     "nat123",
		Original: "https://example.com/dest",
	}
	result := RenderInterstitial(ads, url, "bridge-token-xyz", uuid.Nil)
	assert.Contains(t, result, `native`)
	assert.Contains(t, result, `Native desc`)
}

func TestRenderInterstitial_WithInterstitialAd(t *testing.T) {
	ads := []repository.Ad{
		{ID: uuid.New(), Title: "Inter", Description: sql.NullString{Valid: false}, ImageUrl: "https://ex.com/inter.jpg", AdType: "INTERSTITIAL", Category: "regular"},
	}
	url := repository.Url{
		Slug:     "int123",
		Original: "https://example.com/dest",
	}
	result := RenderInterstitial(ads, url, "bridge-token-xyz", uuid.Nil)
	assert.Contains(t, result, `Inter`)
	assert.Contains(t, result, `interst-section`)
}
