package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unitly/internal/cache"
	"unitly/internal/domain"
)

type FixedWindowLimiter struct {
	client    *cache.Cache
	limit     int64
	window    time.Duration
	clock     domain.Clock
	keyPrefix string
}

func DefaultFixedWindowLimiter(
	client *cache.Cache,
	limit int64,
	window time.Duration,
) (*FixedWindowLimiter, error) {
	return NewFixedWindowLimiter(client, limit, window, nil, "")
}

func NewFixedWindowLimiter(
	client *cache.Cache,
	limit int64,
	window time.Duration,
	clock domain.Clock,
	keyPrefix string,
) (*FixedWindowLimiter, error) {
	if client == nil {
		return nil, errors.New("redis client cannot be nil")
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
		client:    client,
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

	windowKey := l.formatWindowKey(key)
	count, err := l.client.Incr(ctx, windowKey)
	if err != nil {
		return false, fmt.Errorf("redis incr failed: %w", err)
	}

	if count == 1 {
		// Add small buffer to prevent key expiring prematurely during micro clock drift
		ttl := l.window + (2 * time.Second)
		if _, err := l.client.Expire(ctx, windowKey, ttl); err != nil {
			return false, fmt.Errorf("redis expire failed: %w", err)
		}
	}

	return count <= l.limit, nil
}

// Reset clears the current window for the key.
// Returns true if the key was actively tracked and cleared (Option A).
func (l *FixedWindowLimiter) Reset(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, errors.New("key cannot be empty")
	}

	windowKey := l.formatWindowKey(key)

	deleted, err := l.client.Delete(ctx, windowKey)
	if err != nil {
		return false, fmt.Errorf("redis del failed: %w", err)
	}

	return deleted, nil
}

// formatWindowKey truncates current time to window boundaries to construct a deterministic key:
// e.g., "rl:otp:user_123:1700000000"
func (l *FixedWindowLimiter) formatWindowKey(key string) string {
	now := l.clock.Now().Truncate(l.window)
	return fmt.Sprintf("%s:%s:%d", l.keyPrefix, key, now.Unix())
}
