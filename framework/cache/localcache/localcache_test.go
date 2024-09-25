package localcache_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kittipat1413/go-common/framework/cache"
	"github.com/kittipat1413/go-common/framework/cache/localcache"
)

// Ensure that ErrCacheMiss is defined in your cache package
// var ErrCacheMiss = errors.New("cache miss")

func TestLocalCache_SetAndGet(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "testKey"
	value := "testValue"
	duration := 1 * time.Minute

	// Set a value
	c.Set(ctx, key, value, duration)

	// Get the value
	gotValue, err := c.Get(ctx, key, nil)
	require.NoError(t, err)
	assert.Equal(t, value, gotValue, "The value retrieved from cache should match the expected value")
}

func TestLocalCache_GetWithInitializer(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[int]()

	key := "number"
	expectedValue := 42
	duration := 1 * time.Minute

	initializerCalled := false
	initializer := func() (int, time.Duration, error) {
		initializerCalled = true
		return expectedValue, duration, nil
	}

	// Get value with initializer
	value, err := c.Get(ctx, key, initializer)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value, "The value retrieved should match the expected value")
	assert.True(t, initializerCalled, "Initializer should have been called")

	// Get value again, initializer should not be called
	initializerCalled = false
	value, err = c.Get(ctx, key, initializer)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value, "The value retrieved should match the expected value")
	assert.False(t, initializerCalled, "Initializer should not have been called again")
}

func TestLocalCache_Expiration(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "tempKey"
	value := "tempValue"
	duration := 100 * time.Millisecond

	c.Set(ctx, key, value, duration)

	// Get the value before expiration
	gotValue, err := c.Get(ctx, key, nil)
	require.NoError(t, err)
	assert.Equal(t, value, gotValue, "Value should be retrievable before expiration")

	// Wait for the item to expire
	time.Sleep(duration + 100*time.Millisecond)

	// Try to get the value after expiration
	_, err = c.Get(ctx, key, nil)
	assert.ErrorIs(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss after expiration")
}

func TestLocalCache_Invalidate(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "testKey"
	value := "testValue"
	duration := 1 * time.Minute

	c.Set(ctx, key, value, duration)

	// Invalidate the key
	err := c.Invalidate(ctx, key)
	require.NoError(t, err)

	// Try to get the invalidated key
	_, err = c.Get(ctx, key, nil)
	assert.ErrorIs(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss after invalidation")
}

func TestLocalCache_InvalidateAll(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	keys := []string{"key1", "key2", "key3"}
	value := "testValue"
	duration := 1 * time.Minute

	for _, key := range keys {
		c.Set(ctx, key, value, duration)
	}

	// Invalidate all keys
	err := c.InvalidateAll(ctx)
	require.NoError(t, err)

	// Try to get the keys
	for _, key := range keys {
		_, err := c.Get(ctx, key, nil)
		assert.ErrorIs(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss for key %q after InvalidateAll", key)
	}
}

func TestLocalCache_Get_CacheMiss(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "missingKey"

	// Try to get a key that doesn't exist without an initializer
	_, err := c.Get(ctx, key, nil)
	assert.ErrorIs(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss when getting a missing key without initializer")
}

func TestLocalCache_Get_InitializerError(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "initErrorKey"
	expectedErr := errors.New("initializer error")

	initializer := func() (string, time.Duration, error) {
		return "", 0, expectedErr
	}

	_, err := c.Get(ctx, key, initializer)
	assert.ErrorIs(t, err, expectedErr, "Expected error from initializer")
}

func TestLocalCache_Concurrency(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[int]()

	key := "concurrentKey"
	duration := 1 * time.Minute
	initializerCallCount := 0
	var mu sync.Mutex

	initializer := func() (int, time.Duration, error) {
		mu.Lock()
		initializerCallCount++
		mu.Unlock()
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		return 42, duration, nil
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := c.Get(ctx, key, initializer)
			require.NoError(t, err)
			assert.Equal(t, 42, value, "Value retrieved should be 42")
		}()
	}

	wg.Wait()

	assert.Equal(t, 1, initializerCallCount, "Initializer should have been called exactly once")
}
