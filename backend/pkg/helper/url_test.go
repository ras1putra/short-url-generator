package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractS3ObjectKeyFromURL_WithBucketPrefix(t *testing.T) {
	key := ExtractS3ObjectKeyFromURL("http://localhost:3900/short-url-ads/ads/uuid123.jpg", "short-url-ads")
	assert.Equal(t, "ads/uuid123.jpg", key)
}

func TestExtractS3ObjectKeyFromURL_WithoutBucketPrefix(t *testing.T) {
	key := ExtractS3ObjectKeyFromURL("http://localhost:3900/ads/uuid123.jpg", "short-url-ads")
	assert.Equal(t, "ads/uuid123.jpg", key)
}

func TestExtractS3ObjectKeyFromURL_EmptyPath(t *testing.T) {
	key := ExtractS3ObjectKeyFromURL("http://localhost:3900", "short-url-ads")
	assert.Equal(t, "", key)
}

func TestExtractS3ObjectKeyFromURL_InvalidURL(t *testing.T) {
	key := ExtractS3ObjectKeyFromURL("://invalid", "short-url-ads")
	assert.Equal(t, "", key)
}

func TestExtractS3ObjectKeyFromURL_BucketOnlyPath(t *testing.T) {
	key := ExtractS3ObjectKeyFromURL("http://localhost:3900/short-url-ads", "short-url-ads")
	assert.Equal(t, "short-url-ads", key)
}

func TestExtractS3ObjectKeyFromURL_NestedPath(t *testing.T) {
	key := ExtractS3ObjectKeyFromURL("http://localhost:3900/short-url-ads/ads/subdir/file.png", "short-url-ads")
	assert.Equal(t, "ads/subdir/file.png", key)
}
