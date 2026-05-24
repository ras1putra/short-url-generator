package media

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"urlshortener/internal/modules/media/dto"
	"urlshortener/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockMediaService struct {
	uploadFn    func(ctx context.Context, data []byte, filename, contentType string, targetRatio float64) (*dto.MediaUploadResponse, error)
	cropVideoFn func(ctx context.Context, data []byte, filename, contentType string, x, y, w, h int) (*dto.MediaUploadResponse, error)
}

func (m *mockMediaService) Upload(ctx context.Context, data []byte, filename, contentType string, targetRatio float64) (*dto.MediaUploadResponse, error) {
	return m.uploadFn(ctx, data, filename, contentType, targetRatio)
}

func (m *mockMediaService) CropVideo(ctx context.Context, data []byte, filename, contentType string, x, y, w, h int) (*dto.MediaUploadResponse, error) {
	return m.cropVideoFn(ctx, data, filename, contentType, x, y, w, h)
}

func setupMediaHandlerTest(t *testing.T, svc *mockMediaService) *fiber.App {
	_ = zap.ReplaceGlobals(zap.NewNop())
	handler := NewMediaHandler(svc)

	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		c.Locals("request_id", "test-req-id")
		return c.Next()
	})
	app.Post("/media/upload", handler.Upload)
	app.Post("/media/crop", handler.CropVideo)
	return app
}

func TestMediaHandler_Upload_Success(t *testing.T) {
	svc := &mockMediaService{
		uploadFn: func(ctx context.Context, data []byte, filename, contentType string, targetRatio float64) (*dto.MediaUploadResponse, error) {
			return &dto.MediaUploadResponse{
				URL:         "http://localhost:3900/ads/uuid.jpg",
				ContentType: "image/jpeg",
				Size:        1024,
			}, nil
		},
	}
	app := setupMediaHandlerTest(t, svc)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	require.NoError(t, err)
	_, err = part.Write([]byte("fake-image-data"))
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NotNil(t, result["data"])
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "http://localhost:3900/ads/uuid.jpg", data["url"])
}

func TestMediaHandler_Upload_NoFile(t *testing.T) {
	svc := &mockMediaService{}
	app := setupMediaHandlerTest(t, svc)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestMediaHandler_Upload_ServiceError(t *testing.T) {
	svc := &mockMediaService{
		uploadFn: func(ctx context.Context, data []byte, filename, contentType string, targetRatio float64) (*dto.MediaUploadResponse, error) {
			return nil, response.NewAppError(400, "Unsupported media type")
		},
	}
	app := setupMediaHandlerTest(t, svc)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	require.NoError(t, err)
	_, err = part.Write([]byte("fake-image-data"))
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestMediaHandler_Upload_Unauthorized(t *testing.T) {
	svc := &mockMediaService{}
	_ = zap.ReplaceGlobals(zap.NewNop())
	handler := NewMediaHandler(svc)
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Post("/media/upload", handler.Upload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestMediaHandler_CropVideo_Success(t *testing.T) {
	svc := &mockMediaService{
		cropVideoFn: func(ctx context.Context, data []byte, filename, contentType string, x, y, w, h int) (*dto.MediaUploadResponse, error) {
			return &dto.MediaUploadResponse{
				URL:         "http://localhost:3900/ads/uuid.mp4",
				ContentType: "video/mp4",
				Size:        2048,
			}, nil
		},
	}
	app := setupMediaHandlerTest(t, svc)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("x", "10")
	_ = writer.WriteField("y", "10")
	_ = writer.WriteField("w", "100")
	_ = writer.WriteField("h", "100")
	part, err := writer.CreateFormFile("file", "test.mp4")
	require.NoError(t, err)
	_, err = part.Write([]byte("fake-video-data"))
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/media/crop", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestMediaHandler_CropVideo_MissingParam(t *testing.T) {
	app := setupMediaHandlerTest(t, &mockMediaService{})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("x", "10")
	_ = writer.WriteField("y", "10")
	_ = writer.WriteField("w", "100")
	// missing h
	writer.Close()

	req := httptest.NewRequest("POST", "/media/crop", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestMediaHandler_CropVideo_InvalidParam(t *testing.T) {
	app := setupMediaHandlerTest(t, &mockMediaService{})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("x", "abc")
	_ = writer.WriteField("y", "10")
	_ = writer.WriteField("w", "100")
	_ = writer.WriteField("h", "100")
	writer.Close()

	req := httptest.NewRequest("POST", "/media/crop", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}
