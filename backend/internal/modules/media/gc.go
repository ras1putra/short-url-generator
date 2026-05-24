package media

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"urlshortener/internal/repository"
	"urlshortener/pkg/helper"
	"urlshortener/pkg/logger"
)

const (
	mediaGCInterval = 30 * time.Minute
	mediaGCGrace    = 6 * time.Hour
)

func StartOrphanCleaner(ctx context.Context, db *sql.DB, client *s3.Client, bucket string) {
	go func() {
		// Run one sweep shortly after startup, then periodically.
		timer := time.NewTimer(90 * time.Second)
		defer timer.Stop()

		ticker := time.NewTicker(mediaGCInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				runOrphanCleanup(ctx, db, client, bucket)
			case <-ticker.C:
				runOrphanCleanup(ctx, db, client, bucket)
			}
		}
	}()
}

func runOrphanCleanup(parent context.Context, db *sql.DB, client *s3.Client, bucket string) {
	ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
	defer cancel()

	referenced, err := fetchReferencedMediaKeys(ctx, db, bucket)
	if err != nil {
		logger.Ctx(ctx).Error("Media GC: failed to fetch referenced media keys", zap.Error(err))
		return
	}

	var scanned, deleted int
	var continuation *string

	for {
		out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(mediaObjectPrefix),
			ContinuationToken: continuation,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Media GC: failed to list S3 objects", zap.Error(err))
			return
		}

		now := time.Now()
		for _, obj := range out.Contents {
			if obj.Key == nil {
				continue
			}
			scanned++

			key := *obj.Key
			if referenced[key] {
				continue
			}
			if obj.LastModified == nil || now.Sub(*obj.LastModified) < mediaGCGrace {
				continue
			}

			if _, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			}); err != nil {
				logger.Ctx(ctx).Warn("Media GC: failed to delete orphan object", zap.String("key", key), zap.Error(err))
				continue
			}
			deleted++
		}

		if !aws.ToBool(out.IsTruncated) || out.NextContinuationToken == nil {
			break
		}
		continuation = out.NextContinuationToken
	}

	if scanned > 0 || deleted > 0 {
		logger.Ctx(ctx).Info("Media GC sweep complete",
			zap.Int("scanned", scanned),
			zap.Int("referenced", len(referenced)),
			zap.Int("deleted", deleted),
			zap.String("bucket", bucket),
		)
	}
}

func fetchReferencedMediaKeys(ctx context.Context, db *sql.DB, bucket string) (map[string]bool, error) {
	repo := repository.New(db)
	urls, err := repo.GetReferencedMediaURLs(ctx)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]bool)
	for _, imageURL := range urls {
		if key := helper.ExtractS3ObjectKeyFromURL(imageURL, bucket); strings.HasPrefix(key, mediaObjectPrefix) {
			keys[key] = true
		}
	}
	return keys, nil
}
