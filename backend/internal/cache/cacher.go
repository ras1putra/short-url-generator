package cache

import (
	"context"
	"time"
)

type Cacher interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	RateLimitIncrement(ctx context.Context, key string, ttl time.Duration) (int, error)
}