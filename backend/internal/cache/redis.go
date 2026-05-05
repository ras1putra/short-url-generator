package cache

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"urlshortener/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

// Create a new redis client and ping
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0, // default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		zap.L().Error("Failed to ping Redis", zap.Error(err))
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	zap.L().Info("Successfully connected to Redis")
	return &RedisClient{client: client}, nil
}

// Get a value from cache
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set a value in cache with a TTL
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete a key from cache
func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RedisClient) RateLimitIncrement(ctx context.Context, key string, ttl time.Duration) (int, error) {
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return int(incr.Val()), nil
}

func (r *RedisClient) Raw() *redis.Client {
	return r.client
}
