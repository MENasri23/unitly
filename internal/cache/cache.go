package cache

import (
	"context"
	"time"

	"unitly/internal/domain"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Default(addr, username, password string, db int) (domain.Cache, error) {
	return NewCache(&Config{
		Addr:         addr,
		Username:     username,
		Password:     password,
		DB:           db,
		PoolSize:     25,
		MinIdleConns: 10,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

func NewCache(cfg *Config) (domain.Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisCache{Client: rdb}, nil
}

type redisCache struct {
	Client *redis.Client
}

var _ domain.Cache = (*redisCache)(nil)

func (c *redisCache) Close() error {
	return c.Client.Close()
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", domain.ErrNotFound
	}
	return val, err
}

func (c *redisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

func (c *redisCache) ExistsKey(ctx context.Context, key string) (bool, error) {
	val, err := c.Client.Exists(ctx, key).Result()
	return val == 1, err
}

func (c *redisCache) SetIfNotExist(ctx context.Context, key string, val any, exp time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, val, exp).Result()
}

func (c *redisCache) Delete(ctx context.Context, key string) (bool, error) {
	delCount, err := c.Client.Del(ctx, key).Result()
	return delCount == 1, err
}

func (c *redisCache) DeleteAll(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Del(ctx, keys...).Result()
}

func (c *redisCache) SetMap(ctx context.Context, key string, values map[string]any) error {
	return c.Client.HSet(ctx, key, values).Err()
}

func (c *redisCache) GetMap(ctx context.Context, key string) (map[string]string, error) {
	return c.Client.HGetAll(ctx, key).Result()
}

func (c *redisCache) HIncrBy(ctx context.Context, key, field string, increment int64) (int64, error) {
	return c.Client.HIncrBy(ctx, key, field, increment).Result()
}

func (c *redisCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.Client.Expire(ctx, key, expiration).Result()
}

// --- Transactional Pipeline Implementation ---

func (c *redisCache) TxPipeline() domain.TxPipeline {
	return &redisTxPipeline{pipe: c.Client.TxPipeline()}
}

type redisTxPipeline struct {
	pipe redis.Pipeliner
}

var _ domain.TxPipeline = (*redisTxPipeline)(nil)

func (p *redisTxPipeline) Get(ctx context.Context, key string) {
	p.pipe.Get(ctx, key)
}

func (p *redisTxPipeline) SetIfNotExist(ctx context.Context, key string, val any, exp time.Duration) {
	p.pipe.SetNX(ctx, key, val, exp)
}

func (p *redisTxPipeline) Incr(ctx context.Context, key string) {
	p.pipe.Incr(ctx, key)
}

func (p *redisTxPipeline) Delete(ctx context.Context, key string) {
	p.pipe.Del(ctx, key)
}

func (p *redisTxPipeline) Expire(ctx context.Context, key string, expiration time.Duration) {
	p.pipe.Expire(ctx, key, expiration)
}

func (p *redisTxPipeline) HSet(ctx context.Context, key string, value any) (int64, error) {
	return p.pipe.HSet(ctx, key, value).Result()
}

func (p *redisTxPipeline) Exec(ctx context.Context) error {
	_, err := p.pipe.Exec(ctx)
	return err
}

func (p *redisTxPipeline) Discard() {
	p.pipe.Discard()
}

func (c *redisCache) RunScript(ctx context.Context, script *domain.Script, keys []string, args ...any) (any, error) {
	redisScript := redis.NewScript(script.Src)
	return redisScript.Run(ctx, c.Client, keys, args...).Result()
}