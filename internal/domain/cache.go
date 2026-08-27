package domain

import (
	"context"
	"io"
	"time"
)

// TxPipeline abstracts a transactional Redis pipeline (MULTI/EXEC).
type TxPipeline interface {
	Get(ctx context.Context, key string)
	SetIfNotExist(ctx context.Context, key string, val any, exp time.Duration)
	Incr(ctx context.Context, key string)
	Delete(ctx context.Context, key string)
	Expire(ctx context.Context, key string, expiration time.Duration)
	HSet(ctx context.Context, key string, value any) (int64, error)
	Exec(ctx context.Context) error
	Discard()
}

// Script represents a Lua script to be executed against the cache.
type Script struct {
	Src  string
	Hash string
}

// NewScript constructs a Script instance given its raw Lua source code.
func NewScript(src string) *Script {
	return &Script{Src: src}
}

type Cache interface {
	io.Closer

	Get(ctx context.Context, key string) (string, error)
	Incr(ctx context.Context, key string) (int64, error)
	ExistsKey(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, val any, exp time.Duration) error
	SetIfNotExist(ctx context.Context, key string, val any, exp time.Duration) (bool, error)
	Delete(ctx context.Context, key string) (bool, error)
	DeleteAll(ctx context.Context, keys ...string) (int64, error)
	SetMap(ctx context.Context, key string, values map[string]any) error
	GetMap(ctx context.Context, key string) (map[string]string, error)
	HIncrBy(ctx context.Context, key, field string, increment int64) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)

	// Transactional Pipeline
	TxPipeline() TxPipeline

	// Lua Scripting
	RunScript(ctx context.Context, script *Script, keys []string, args ...any) (any, error)
}