package helper

import (
	"net/url"
	"strings"
)

// ExtractS3ObjectKeyFromURL returns the object key path from a public S3 URL.
// It supports both:
//   - /<bucket>/<key>
//   - /<key>
func ExtractS3ObjectKeyFromURL(rawURL, bucket string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return ""
	}

	bucketPrefix := bucket + "/"
	if strings.HasPrefix(path, bucketPrefix) {
		path = strings.TrimPrefix(path, bucketPrefix)
	}

	return path
}
