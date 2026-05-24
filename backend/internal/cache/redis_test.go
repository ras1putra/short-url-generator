package cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"urlshortener/internal/config"
)

func newTestRedisClient(t *testing.T) *RedisClient {
	_ = zap.ReplaceGlobals(zap.NewNop())

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	cfg := &config.Config{RedisAddr: addr}
	client, err := NewRedisClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.client.FlushDB(ctx).Err()
	require.NoError(t, err)

	t.Cleanup(func() {
		client.client.FlushDB(ctx)
		client.client.Close()
	})

	return client
}

func TestNewRedisClient_Success(t *testing.T) {
	client := newTestRedisClient(t)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestRedisClient_SetAndGet(t *testing.T) {
	client := newTestRedisClient(t)
	ctx := context.Background()

	err := client.Set(ctx, "test:key", "hello", time.Minute)
	require.NoError(t, err)

	val, err := client.Get(ctx, "test:key")
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestRedisClient_Get_NotFound(t *testing.T) {
	client := newTestRedisClient(t)
	ctx := context.Background()

	_, err := client.Get(ctx, "test:nonexistent")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestRedisClient_Del(t *testing.T) {
	client := newTestRedisClient(t)
	ctx := context.Background()

	err := client.Set(ctx, "test:del-key", "value", time.Minute)
	require.NoError(t, err)

	val, err := client.Get(ctx, "test:del-key")
	require.NoError(t, err)
	assert.Equal(t, "value", val)

	err = client.Del(ctx, "test:del-key")
	require.NoError(t, err)

	_, err = client.Get(ctx, "test:del-key")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestRedisClient_Exists(t *testing.T) {
	client := newTestRedisClient(t)
	ctx := context.Background()

	exists, err := client.Exists(ctx, "test:exists-key")
	require.NoError(t, err)
	assert.False(t, exists)

	err = client.Set(ctx, "test:exists-key", "value", time.Minute)
	require.NoError(t, err)

	exists, err = client.Exists(ctx, "test:exists-key")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisClient_RateLimitIncrement(t *testing.T) {
	client := newTestRedisClient(t)
	ctx := context.Background()

	count, err := client.RateLimitIncrement(ctx, "test:rate-limit", 2*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = client.RateLimitIncrement(ctx, "test:rate-limit", 2*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestRedisClient_Raw(t *testing.T) {
	client := newTestRedisClient(t)
	raw := client.Raw()
	assert.NotNil(t, raw)
}
