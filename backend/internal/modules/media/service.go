package media

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"urlshortener/internal/modules/media/dto"
	"urlshortener/pkg/logger"
	"urlshortener/pkg/response"
)

type MediaService struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

const mediaObjectPrefix = "ads/"

func NewMediaService(client *s3.Client, bucket, publicURL string) *MediaService {
	return &MediaService{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}
}

func (s *MediaService) Upload(ctx context.Context, data []byte, filename, contentType string, targetRatio float64) (*dto.MediaUploadResponse, error) {
	isImage, isVideo := classifyMedia(contentType)

	if !isImage && !isVideo {
		return nil, response.NewAppError(400, "Unsupported media type. Supported formats are static/animated images (PNG, JPEG, WEBP, GIF) and videos (MP4, WEBM, OGG)")
	}

	maxSize, sizeText := sizeLimit(int64(len(data)), isVideo)
	if int64(len(data)) > maxSize {
		return nil, response.NewAppError(400, "File size exceeds limits. Maximum size allowed is "+sizeText)
	}

	if err := validateMedia(ctx, data, contentType, isImage, isVideo, targetRatio); err != nil {
		return nil, err
	}

	return s.store(ctx, data, filename, contentType)
}

func (s *MediaService) CropVideo(ctx context.Context, data []byte, filename, contentType string, x, y, w, h int) (*dto.MediaUploadResponse, error) {
	tmpDir, err := os.MkdirTemp("", "media-crop-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".mp4"
	}
	inputPath := filepath.Join(tmpDir, "input"+ext)
	outputPath := filepath.Join(tmpDir, "output"+ext)

	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write temp input: %w", err)
	}

	cropFilter := fmt.Sprintf("crop=%d:%d:%d:%d", w, h, x, y)
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vf", cropFilter,
		"-c:a", "copy",
		"-y",
		outputPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Ctx(ctx).Error("FFmpeg crop failed",
			zap.String("filter", cropFilter),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return nil, response.NewAppError(500, "Failed to process video. The crop region may be invalid.")
	}

	croppedData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read cropped output: %w", err)
	}

	logger.Ctx(ctx).Info("Video cropped successfully via FFmpeg",
		zap.String("filter", cropFilter),
		zap.Int("old_size", len(data)),
		zap.Int("new_size", len(croppedData)),
	)

	return s.store(ctx, croppedData, filename, contentType)
}

func (s *MediaService) store(ctx context.Context, data []byte, filename, contentType string) (*dto.MediaUploadResponse, error) {
	ext := filepath.Ext(filename)
	objectKey := fmt.Sprintf("%s%s%s", mediaObjectPrefix, uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(data))),
		ACL:           types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upload file to S3", zap.Error(err))
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	logger.Ctx(ctx).Info("File stored successfully in S3",
		zap.String("bucket", s.bucket),
		zap.String("key", objectKey),
		zap.Int64("size", int64(len(data))),
	)

	base := strings.TrimSuffix(s.publicURL, "/")

	var fileURL string
	if strings.Contains(base, s.bucket) {
		fileURL = fmt.Sprintf("%s/%s", base, objectKey)
	} else {
		fileURL = fmt.Sprintf("%s/%s/%s", base, s.bucket, objectKey)
	}

	return &dto.MediaUploadResponse{
		URL:         fileURL,
		ContentType: contentType,
		Size:        int64(len(data)),
	}, nil
}

func validateMedia(ctx context.Context, data []byte, contentType string, isImage, isVideo bool, targetRatio float64) error {
	if isImage {
		width, height, err := getImageDimensions(data, contentType)
		if err != nil {
			logger.Ctx(ctx).Warn("Could not decode image dimensions", zap.Error(err))
			return nil
		}
		return validateRatio(width, height, targetRatio)
	}

	if isVideo {
		width, height, err := probeVideoDimensions(data, contentType)
		if err != nil {
			logger.Ctx(ctx).Warn("Could not probe video dimensions", zap.Error(err))
			return nil
		}
		return validateRatio(width, height, targetRatio)
	}

	return nil
}
