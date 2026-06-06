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
	return l.catch.SetIfNotExist(ctx, key, 1, l.window)
}

func (l *OtpRateLimiter) Reset(ctx context.Context, key string) (bool, error) {
	return l.catch.Delete(ctx, key)
}
