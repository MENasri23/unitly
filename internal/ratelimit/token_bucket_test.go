package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// MockClock allows manually stepping time in unit tests
type MockClock struct {
	currentTime time.Time
}

func (m *MockClock) Now() time.Time {
	return m.currentTime
}

func (m *MockClock) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

func TestTokenBucketWithMockClock(t *testing.T) {
	// Setup Redis client or mini redis instance
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	mockClock := &MockClock{currentTime: time.Unix(1700000000, 0)}

	limiter, _ := NewTokenBucketLimiter(rdb, Config{
		Capacity:   2,
		RefillRate: 1.0, // 1 token / second
		Clock:      mockClock,
	})

	ctx := context.Background()
	key := "test_user"

	// Reset key state before test
	limiter.Reset(ctx, key)

	// Consume 2 tokens
	a1, _ := limiter.Allow(ctx, key) // true (1 left)
	a2, _ := limiter.Allow(ctx, key) // true (0 left)
	a3, _ := limiter.Allow(ctx, key) // false (empty)

	if !a1 || !a2 || a3 {
		t.Fatalf("Expected true, true, false but got %v, %v, %v", a1, a2, a3)
	}

	// Advance time by 1 second -> replenishes 1 token
	mockClock.Advance(1 * time.Second)

	a4, _ := limiter.Allow(ctx, key) // true
	if !a4 {
		t.Fatalf("Expected true after 1 second refill")
	}
}
