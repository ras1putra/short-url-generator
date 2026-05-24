package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaUploadResponse_JSON(t *testing.T) {
	resp := MediaUploadResponse{
		URL:         "http://localhost:3900/ads/uuid123.jpg",
		ContentType: "image/jpeg",
		Size:        1024,
	}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"url":"http://localhost:3900/ads/uuid123.jpg"`)
	assert.Contains(t, string(data), `"content_type":"image/jpeg"`)
	assert.Contains(t, string(data), `"size":1024`)
}

func TestMediaUploadResponse_EmptyURL(t *testing.T) {
	resp := MediaUploadResponse{ContentType: "video/mp4", Size: 0}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"url":""`)
}
