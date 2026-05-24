package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/internal/config"

	"github.com/stretchr/testify/require"
)

func SetupTestRedis(t *testing.T) cache.Cacher {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	cfg := &config.Config{
		RedisAddr: addr,
	}

	client, err := cache.NewRedisClient(cfg)
	require.NoError(t, err, "Failed to connect to test Redis")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Raw().FlushDB(ctx).Err()
	require.NoError(t, err, "Failed to flush Redis before test")

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		client.Raw().FlushDB(cleanupCtx)
	})

	return client
}

func SetupTestRedisRaw(t *testing.T) *cache.RedisClient {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	cfg := &config.Config{
		RedisAddr: addr,
	}

	client, err := cache.NewRedisClient(cfg)
	require.NoError(t, err, "Failed to connect to test Redis")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Raw().FlushDB(ctx).Err()
	require.NoError(t, err, "Failed to flush Redis before test")

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		client.Raw().FlushDB(cleanupCtx)
	})

	return client
}
