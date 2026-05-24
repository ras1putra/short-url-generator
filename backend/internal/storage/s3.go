package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"

	"urlshortener/internal/config"
	"urlshortener/pkg/logger"
)

func NewS3Client(cfg *config.Config) (*s3.Client, error) {
	region := cfg.S3Region
	if region == "" {
		region = "us-east-1"
	}

	awsCfg := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Region:      region,
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		o.UsePathStyle = true
	})

	return client, nil
}

func EnsureBucket(ctx context.Context, client *s3.Client, cfg *config.Config) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.S3Bucket),
	})
	if err == nil {
		logger.Ctx(ctx).Info("S3 bucket already exists", zap.String("bucket", cfg.S3Bucket))
		return nil
	}

	var notFound *types.NotFound
	if !errors.As(err, &notFound) {
		logger.Ctx(ctx).Error("Failed to check bucket existence", zap.String("bucket", cfg.S3Bucket), zap.Error(err))
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	logger.Ctx(ctx).Info("S3 bucket does not exist, creating it now...", zap.String("bucket", cfg.S3Bucket))

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(cfg.S3Bucket),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create S3 bucket", zap.String("bucket", cfg.S3Bucket), zap.Error(err))
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "PublicRead",
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, cfg.S3Bucket)

	_, err = client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(cfg.S3Bucket),
		Policy: aws.String(policy),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to set public S3 bucket policy", zap.String("bucket", cfg.S3Bucket), zap.Error(err))
		return fmt.Errorf("failed to set public bucket policy: %w", err)
	}

	logger.Ctx(ctx).Info("S3 bucket created and policy configured successfully", zap.String("bucket", cfg.S3Bucket))
	return nil
}
