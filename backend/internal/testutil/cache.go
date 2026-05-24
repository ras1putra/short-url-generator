package testutil

import (
	"context"
	"errors"
	"sync"
	"time"
)

type FakeCacher struct {
	mu    sync.RWMutex
	store map[string]string
}

func NewFakeCacher() *FakeCacher {
	return &FakeCacher{
		store: make(map[string]string),
	}
}

func (f *FakeCacher) Get(ctx context.Context, key string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	val, ok := f.store[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
}

func (f *FakeCacher) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if strVal, ok := value.(string); ok {
		f.store[key] = strVal
	} else {
		f.store[key] = "1"
	}
	return nil
}

func (f *FakeCacher) Del(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.store, key)
	return nil
}

func (f *FakeCacher) Exists(ctx context.Context, key string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.store[key]
	return ok, nil
}

func (f *FakeCacher) RateLimitIncrement(ctx context.Context, key string, ttl time.Duration) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return 1, nil
}
