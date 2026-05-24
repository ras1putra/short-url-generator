package media

import (
	"context"
	"io"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"urlshortener/internal/modules/media/dto"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
)

type MediaServicer interface {
	Upload(ctx context.Context, data []byte, filename, contentType string, targetRatio float64) (*dto.MediaUploadResponse, error)
	CropVideo(ctx context.Context, data []byte, filename, contentType string, x, y, w, h int) (*dto.MediaUploadResponse, error)
}

type MediaHandler struct {
	svc MediaServicer
}

func NewMediaHandler(svc MediaServicer) *MediaHandler {
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	_, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Unauthorized media upload request", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Media upload failed: file parameter missing", zap.Error(err))
		return response.HandleError(c, response.NewAppError(400, "No file provided under parameter 'file'"), "MediaUpload")
	}

	fileStream, err := fileHeader.Open()
	if err != nil {
		logger.Ctx(c.UserContext()).Error("Media upload failed: cannot open file stream", zap.Error(err))
		return response.HandleError(c, response.NewAppError(500, "Failed to parse upload request"), "MediaUpload")
	}
	defer fileStream.Close()

	fileBytes, err := io.ReadAll(fileStream)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("Media upload failed: cannot read file bytes", zap.Error(err))
		return response.HandleError(c, response.NewAppError(500, "Failed to read file"), "MediaUpload")
	}

	contentType := detectContentType(fileHeader.Filename, fileHeader.Header.Get("Content-Type"))
	targetRatio, err := parseTargetRatio(c.FormValue("target_ratio", ""))
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Media upload failed: invalid target ratio format", zap.Error(err))
		return response.HandleError(c, err, "MediaUpload")
	}

	resp, err := h.svc.Upload(c.Context(), fileBytes, fileHeader.Filename, contentType, targetRatio)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("Media upload service failed", zap.Error(err))
		return response.HandleError(c, err, "MediaUpload")
	}

	logger.Ctx(c.UserContext()).Info("Media uploaded",
		zap.String("url", resp.URL),
		zap.String("content_type", resp.ContentType),
		zap.Int64("size", resp.Size),
		zap.String("ip", c.IP()),
	)

	return response.Created(c, resp, "Media uploaded successfully")
}

func (h *MediaHandler) CropVideo(c *fiber.Ctx) error {
	_, err := helper.UserIDFromCtx(c)
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Unauthorized video crop request", zap.Error(err))
		return response.Unauthorized(c, err.Error())
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Video crop failed: file parameter missing", zap.Error(err))
		return response.HandleError(c, response.NewAppError(400, "No file provided under parameter 'file'"), "CropVideo")
	}

	cropX, err := parseCropParam(c, "x")
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Video crop failed: invalid x parameter", zap.Error(err))
		return response.HandleError(c, err, "CropVideo")
	}
	cropY, err := parseCropParam(c, "y")
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Video crop failed: invalid y parameter", zap.Error(err))
		return response.HandleError(c, err, "CropVideo")
	}
	cropW, err := parseCropParam(c, "w")
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Video crop failed: invalid w parameter", zap.Error(err))
		return response.HandleError(c, err, "CropVideo")
	}
	cropH, err := parseCropParam(c, "h")
	if err != nil {
		logger.Ctx(c.UserContext()).Warn("Video crop failed: invalid h parameter", zap.Error(err))
		return response.HandleError(c, err, "CropVideo")
	}

	fileStream, err := fileHeader.Open()
	if err != nil {
		logger.Ctx(c.UserContext()).Error("Video crop failed: cannot open file stream", zap.Error(err))
		return response.HandleError(c, response.NewAppError(500, "Failed to parse upload request"), "CropVideo")
	}
	defer fileStream.Close()

	fileBytes, err := io.ReadAll(fileStream)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("Video crop failed: cannot read file bytes", zap.Error(err))
		return response.HandleError(c, response.NewAppError(500, "Failed to read file"), "CropVideo")
	}

	contentType := detectContentType(fileHeader.Filename, fileHeader.Header.Get("Content-Type"))

	resp, err := h.svc.CropVideo(c.Context(), fileBytes, fileHeader.Filename, contentType, cropX, cropY, cropW, cropH)
	if err != nil {
		logger.Ctx(c.UserContext()).Error("Video crop service failed", zap.Error(err))
		return response.HandleError(c, err, "CropVideo")
	}

	logger.Ctx(c.UserContext()).Info("Video cropped",
		zap.String("url", resp.URL),
		zap.String("content_type", resp.ContentType),
		zap.Int64("size", resp.Size),
		zap.Int("crop_x", cropX),
		zap.Int("crop_y", cropY),
		zap.Int("crop_w", cropW),
		zap.Int("crop_h", cropH),
		zap.String("ip", c.IP()),
	)

	return response.Created(c, resp, "Video cropped and uploaded successfully")
}
