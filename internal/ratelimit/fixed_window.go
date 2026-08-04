package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unitly/internal/domain"
)

type FixedWindowLimiter struct {
	cache     domain.Cache
	limit     int64
	window    time.Duration
	clock     domain.Clock
	keyPrefix string
}

func DefaultFixedWindowLimiter(
	cache domain.Cache,
	limit int64,
	window time.Duration,
) (*FixedWindowLimiter, error) {
	return NewFixedWindowLimiter(cache, limit, window, nil, "")
}

func NewFixedWindowLimiter(
	cache domain.Cache,
	limit int64,
	window time.Duration,
	clock domain.Clock,
	keyPrefix string,
) (*FixedWindowLimiter, error) {
	if cache == nil {
		return nil, errors.New("cache client cannot be nil")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}
	if window <= 0 {
		return nil, errors.New("window duration must be greater than 0")
	}
	if clock == nil {
		clock = domain.NewRealClock()
	}
	if keyPrefix == "" {
		keyPrefix = "rl"
	}

	return &FixedWindowLimiter{
		cache:     cache,
		limit:     limit,
		window:    window,
		clock:     clock,
		keyPrefix: keyPrefix,
	}, nil
}

// Allow evaluates whether a request identified by key is allowed in the current window.
func (l *FixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, errors.New("key cannot be empty")
	}

	now := l.clock.Now()
	windowKey := l.formatWindowKeyAt(key, now)

	count, err := l.cache.Incr(ctx, windowKey)
	if err != nil {
		return false, fmt.Errorf("cache incr failed: %w", err)
	}

	if count == 1 {
		// Calculate remaining time in the window + clock drift buffer
		windowStart := now.Truncate(l.window)
		windowEnd := windowStart.Add(l.window)
		ttl := windowEnd.Sub(now) + (2 * time.Second)

		if _, err := l.cache.Expire(ctx, windowKey, ttl); err != nil {
			// Best-effort cleanup using background context so context cancellation
			// does not leave a key leaking permanently without a TTL.
			_, _ = l.cache.Delete(context.Background(), windowKey)
			return false, fmt.Errorf("cache expire failed: %w", err)
		}
	}

	return count <= l.limit, nil
}

// Reset clears the current window for the key.
// Returns true if the key was actively tracked and cleared.
func (l *FixedWindowLimiter) Reset(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, errors.New("key cannot be empty")
	}

	now := l.clock.Now()
	windowKey := l.formatWindowKeyAt(key, now)

	deleted, err := l.cache.Delete(ctx, windowKey)
	if err != nil {
		return false, fmt.Errorf("cache del failed: %w", err)
	}

	return deleted, nil
}

// formatWindowKey truncates current time to window boundaries to construct a deterministic key.
func (l *FixedWindowLimiter) formatWindowKey(key string) string {
	return l.formatWindowKeyAt(key, l.clock.Now())
}

// formatWindowKeyAt constructs the deterministic window key for a explicit timestamp.
func (l *FixedWindowLimiter) formatWindowKeyAt(key string, now time.Time) string {
	windowStart := now.Truncate(l.window)
	return fmt.Sprintf("%s:%s:%d", l.keyPrefix, key, windowStart.Unix())
}