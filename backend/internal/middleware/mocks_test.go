package middleware

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockCacher struct {
	mock.Mock
}

func (m *MockCacher) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockCacher) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacher) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacher) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockCacher) RateLimitIncrement(ctx context.Context, key string, ttl time.Duration) (int, error) {
	args := m.Called(ctx, key, ttl)
	return args.Get(0).(int), args.Error(1)
}