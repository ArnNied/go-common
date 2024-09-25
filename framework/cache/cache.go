package cache

import (
	"context"
	"errors"
	"time"
)

var ErrCacheMiss = errors.New("cache miss")

type Initializer[T any] func() (T, time.Duration, error)

type Cache[T any] interface {
	Get(ctx context.Context, key string, initializer Initializer[T]) (T, error)
	Set(ctx context.Context, key string, value T, duration time.Duration)
	Invalidate(ctx context.Context, key string) error
	InvalidateAll(ctx context.Context) error
}
