package tknbckt

import (
	"context"
	"fmt"
	"math"
	"time"
	"unitly/internal/domain"

	"github.com/redis/go-redis/v9"
)

// Config defines the rate limiter settings.
type Config struct {
	// Capacity is the maximum token bucket capacity (burst size).
	Capacity float64

	// RefillRate is the rate at which tokens replenish per second.
	RefillRate float64

	// KeyPrefix is an optional prefix for Redis keys (default: "rate_limit:").
	KeyPrefix string

	// TTL specifies the key expiration in Redis for idle keys.
	// If set to 0, it auto-calculates based on time to reach full capacity * 2.
	TTL time.Duration

	// Clock allows plugging in a custom clock for testing (default: System Clock).
	Clock domain.Clock
}

// TokenBucketLimiter implements RateLimiter using Redis and Lua.
type TokenBucketLimiter struct {
	client    redis.Cmdable
	cfg       Config
	luaScript *redis.Script
}

// Atomic Lua Script for Token Bucket Calculation
const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2]) -- Tokens per millisecond
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])         -- Current timestamp in milliseconds
local ttl = tonumber(ARGV[5])         -- Expiration TTL in seconds

-- Fetch current state
local data = redis.call("HMGET", key, "tokens", "last_updated")
local tokens = tonumber(data[1])
local last_updated = tonumber(data[2])

if tokens == nil or last_updated == nil then
    -- Initial state: bucket full
    tokens = capacity
    last_updated = now
else
    -- Calculate replenished tokens since last request
    local elapsed = math.max(0, now - last_updated)
    tokens = math.min(capacity, tokens + (elapsed * refill_rate))
    last_updated = now
end

local allowed = 0
if tokens >= requested then
    tokens = tokens - requested
    allowed = 1
end

-- Save updated state
redis.call("HSET", key, "tokens", tokens, "last_updated", last_updated)

-- Automatically handle memory cleanup via key expiration
if ttl > 0 then
    redis.call("EXPIRE", key, ttl)
end

return allowed
`

// NewTokenBucketLimiter initializes and validates the rate limiter.
func NewTokenBucketLimiter(client redis.Cmdable, cfg Config) (*TokenBucketLimiter, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client cannot be nil")
	}
	if cfg.Capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}
	if cfg.RefillRate <= 0 {
		return nil, fmt.Errorf("refill rate must be greater than 0")
	}

	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "rate_limit:"
	}

	if cfg.Clock == nil {
		cfg.Clock = domain.NewRealClock()
	}

	// Calculate default TTL if not supplied: twice the time needed to fill an empty bucket
	if cfg.TTL <= 0 {
		timeToFullSec := cfg.Capacity / cfg.RefillRate
		cfg.TTL = time.Duration(math.Ceil(timeToFullSec*2)) * time.Second
		if cfg.TTL < 10*time.Second {
			cfg.TTL = 10 * time.Second // Minimum safety floor
		}
	}

	return &TokenBucketLimiter{
		client:    client,
		cfg:       cfg,
		luaScript: redis.NewScript(tokenBucketScript),
	}, nil
}

// Allow checks if a request with the given key is permitted.
func (r *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := r.cfg.KeyPrefix + key
	nowMs := r.cfg.Clock.Now().UnixNano() / int64(time.Millisecond)
	refillPerMs := r.cfg.RefillRate / 1000.0
	ttlSeconds := int64(r.cfg.TTL.Seconds())

	res, err := r.luaScript.Run(ctx, r.client, []string{fullKey},
		r.cfg.Capacity,
		refillPerMs,
		1, // Consume 1 token per request
		nowMs,
		ttlSeconds,
	).Int()

	if err != nil {
		return false, fmt.Errorf("rate limiter script execution failed: %w", err)
	}

	return res == 1, nil
}

// Reset clears the bucket state for a specific key in Redis.
func (r *TokenBucketLimiter) Reset(ctx context.Context, key string) (bool, error) {
	fullKey := r.cfg.KeyPrefix + key
	deleted, err := r.client.Del(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to reset rate limit key %s: %w", fullKey, err)
	}

	return deleted > 0, nil
}
