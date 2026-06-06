package otprepo

import (
	"context"
	"time"
	"unitly/internal/cache"
)

type OtpRateLimiter struct {
	catch  *cache.Cache
	window time.Duration
}

func NewOtpRateLimiter(catch *cache.Cache, window time.Duration) *OtpRateLimiter {
	return &OtpRateLimiter{
		catch:  catch,
		window: window,
	}
}

func (l *OtpRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	set, err := l.catch.SetIfNotExist(ctx, key, 1, l.window)
	return set, err
}

func (l *OtpRateLimiter) Reset(ctx context.Context, key string) (bool, error) {
	deleted, err := l.catch.Delete(ctx, key)
	return deleted, err
}
