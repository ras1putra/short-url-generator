package media

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/image/webp"

	"urlshortener/pkg/constants"
	"urlshortener/pkg/response"
)

func getImageDimensions(data []byte, contentType string) (int, int, error) {
	var width, height int
	var err error

	switch contentType {
	case constants.ContentTypePNG:
		var cfg image.Config
		cfg, err = png.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			width, height = cfg.Width, cfg.Height
		}
	case constants.ContentTypeJPEG:
		var cfg image.Config
		cfg, err = jpeg.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			width, height = cfg.Width, cfg.Height
		}
	case constants.ContentTypeGIF:
		var cfg image.Config
		cfg, err = gif.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			width, height = cfg.Width, cfg.Height
		}
	case constants.ContentTypeWEBP:
		var cfg image.Config
		cfg, err = webp.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			width, height = cfg.Width, cfg.Height
		}
	default:
		return 0, 0, fmt.Errorf("unsupported image content type: %s", contentType)
	}
	if err != nil {
		return 0, 0, fmt.Errorf("decode dimensions: %w", err)
	}
	return width, height, nil
}

func getVideoDimensions(filePath string) (int, int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("ffprobe: unexpected output %q", string(out))
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe: parse width %q: %w", parts[0], err)
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe: parse height %q: %w", parts[1], err)
	}

	return width, height, nil
}

func validateRatio(width, height int, targetRatio float64) error {
	if targetRatio <= 0 {
		return nil
	}

	actualRatio := float64(width) / float64(height)
	tolerance := math.Abs(actualRatio-targetRatio) / targetRatio

	if tolerance > constants.RatioTolerance {
		return response.NewAppError(400, fmt.Sprintf(
			"Media aspect ratio %.2f:1 does not match the required ratio %.2f:1 (tolerance %.0f%%). Please upload or crop to the correct ratio.",
			actualRatio, targetRatio, constants.RatioTolerance*100,
		))
	}

	return nil
}

func detectContentType(filename, headerType string) string {
	if headerType != "" {
		return headerType
	}

	ext := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(ext, ".png"):
		return constants.ContentTypePNG
	case strings.HasSuffix(ext, ".jpg"), strings.HasSuffix(ext, ".jpeg"):
		return constants.ContentTypeJPEG
	case strings.HasSuffix(ext, ".webp"):
		return constants.ContentTypeWEBP
	case strings.HasSuffix(ext, ".gif"):
		return constants.ContentTypeGIF
	case strings.HasSuffix(ext, ".mp4"):
		return constants.ContentTypeMP4
	case strings.HasSuffix(ext, ".webm"):
		return constants.ContentTypeWEBM
	case strings.HasSuffix(ext, ".ogg"):
		return constants.ContentTypeOGG
	}
	return ""
}

func classifyMedia(contentType string) (isImage, isVideo bool) {
	isImage = strings.HasPrefix(contentType, constants.MediaTypeImage)
	isVideo = strings.HasPrefix(contentType, constants.MediaTypeVideo)
	return
}

func sizeLimit(size int64, isVideo bool) (max int64, label string) {
	if isVideo {
		return int64(constants.MaxVideoSize), constants.MaxVideoSizeText
	}
	return int64(constants.MaxImageSize), constants.MaxImageSizeText
}

func parseTargetRatio(raw string) (float64, error) {
	if raw == "" {
		return 0, nil
	}
	ratio, err := strconv.ParseFloat(raw, 64)
	if err != nil || ratio <= 0 {
		return 0, response.NewAppError(400, "Invalid target_ratio value")
	}
	return ratio, nil
}

func parseCropParam(c *fiber.Ctx, name string) (int, error) {
	raw := c.FormValue(name, "")
	if raw == "" {
		return 0, response.NewAppError(400, "Missing crop parameter: "+name)
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val < 0 {
		return 0, response.NewAppError(400, "Invalid crop parameter: "+name)
	}
	return val, nil
}

func probeVideoDimensions(data []byte, contentType string) (int, int, error) {
	tmpDir, err := os.MkdirTemp("", "media-validate-*")
	if err != nil {
		return 0, 0, err
	}
	defer os.RemoveAll(tmpDir)

	ext := ".mp4"
	switch contentType {
	case constants.ContentTypeWEBM:
		ext = ".webm"
	case constants.ContentTypeOGG:
		ext = ".ogg"
	}

	tmpPath := filepath.Join(tmpDir, "video"+ext)
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return 0, 0, err
	}

	return getVideoDimensions(tmpPath)
}
