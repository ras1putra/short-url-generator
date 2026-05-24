package media

import (
	"testing"

	"urlshortener/pkg/response"

	"github.com/stretchr/testify/assert"
)

func TestDetectContentType_EmptyHeader(t *testing.T) {
	assert.Equal(t, "image/png", detectContentType("image.png", ""))
	assert.Equal(t, "image/jpeg", detectContentType("photo.jpg", ""))
	assert.Equal(t, "image/jpeg", detectContentType("photo.jpeg", ""))
	assert.Equal(t, "image/webp", detectContentType("img.webp", ""))
	assert.Equal(t, "image/gif", detectContentType("anim.gif", ""))
	assert.Equal(t, "video/mp4", detectContentType("video.mp4", ""))
	assert.Equal(t, "video/webm", detectContentType("video.webm", ""))
	assert.Equal(t, "video/ogg", detectContentType("video.ogg", ""))
}

func TestDetectContentType_UnknownExtension(t *testing.T) {
	assert.Equal(t, "", detectContentType("file.bin", ""))
	assert.Equal(t, "", detectContentType("file", ""))
}

func TestDetectContentType_HeaderPreferred(t *testing.T) {
	assert.Equal(t, "image/png", detectContentType("file.jpg", "image/png"))
	assert.Equal(t, "video/mp4", detectContentType("file.png", "video/mp4"))
}

func TestDetectContentType_UpperCaseExtension(t *testing.T) {
	assert.Equal(t, "image/png", detectContentType("image.PNG", ""))
	assert.Equal(t, "image/jpeg", detectContentType("photo.JPG", ""))
}

func TestClassifyMedia_Image(t *testing.T) {
	isImage, isVideo := classifyMedia("image/png")
	assert.True(t, isImage)
	assert.False(t, isVideo)
}

func TestClassifyMedia_Video(t *testing.T) {
	isImage, isVideo := classifyMedia("video/mp4")
	assert.False(t, isImage)
	assert.True(t, isVideo)
}

func TestClassifyMedia_Unknown(t *testing.T) {
	isImage, isVideo := classifyMedia("application/octet-stream")
	assert.False(t, isImage)
	assert.False(t, isVideo)
}

func TestSizeLimit_Image(t *testing.T) {
	max, label := sizeLimit(100, false)
	assert.Equal(t, int64(5<<20), max)
	assert.Equal(t, "5MB", label)
}

func TestSizeLimit_Video(t *testing.T) {
	max, label := sizeLimit(100, true)
	assert.Equal(t, int64(20<<20), max)
	assert.Equal(t, "20MB", label)
}

func TestParseTargetRatio_Empty(t *testing.T) {
	ratio, err := parseTargetRatio("")
	assert.NoError(t, err)
	assert.Equal(t, float64(0), ratio)
}

func TestParseTargetRatio_Valid(t *testing.T) {
	ratio, err := parseTargetRatio("1.5")
	assert.NoError(t, err)
	assert.Equal(t, 1.5, ratio)
}

func TestParseTargetRatio_Invalid(t *testing.T) {
	_, err := parseTargetRatio("abc")
	assert.Error(t, err)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
}

func TestParseTargetRatio_Zero(t *testing.T) {
	_, err := parseTargetRatio("0")
	assert.Error(t, err)
}

func TestParseTargetRatio_Negative(t *testing.T) {
	_, err := parseTargetRatio("-1.5")
	assert.Error(t, err)
}

func TestValidateRatio_NoTarget(t *testing.T) {
	err := validateRatio(1920, 1080, 0)
	assert.NoError(t, err)
}

func TestValidateRatio_ExactMatch(t *testing.T) {
	err := validateRatio(16, 9, 16.0/9.0)
	assert.NoError(t, err)
}

func TestValidateRatio_WithinTolerance(t *testing.T) {
	err := validateRatio(1920, 1080, 16.0/9.0)
	assert.NoError(t, err)
}

func TestValidateRatio_OutOfTolerance(t *testing.T) {
	err := validateRatio(100, 100, 16.0/9.0)
	assert.Error(t, err)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
}

func TestGetImageDimensions_UnsupportedContentType(t *testing.T) {
	_, _, err := getImageDimensions([]byte{0, 1, 2}, "application/octet-stream")
	assert.Error(t, err)
}

func TestGetImageDimensions_SmallPNG(t *testing.T) {
	// Minimal valid PNG: 1x1 red pixel
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 pixel
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0x68, 0xC0, 0xF0, 0x1F,
		0x00, 0x00, 0x15, 0x00, 0x01, 0x9A, 0xA4, 0xD9,
		0xCE, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	w, h, err := getImageDimensions(pngData, "image/png")
	assert.NoError(t, err)
	assert.Equal(t, 1, w)
	assert.Equal(t, 1, h)
}
