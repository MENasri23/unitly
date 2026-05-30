package cache

import (
	"context"
	"fmt"
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

func Default(addr, username, password string, db int) (*Cache, error) {
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

func NewCache(cfg *Config) (*Cache, error) {
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

	fmt.Println("Cache client connected successfully.")

	return &Cache{rdb}, nil
}

type Cache struct {
	Client *redis.Client
}

func (c *Cache) Close() error {
	return c.Client.Close()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", domain.ErrNotFound
	}

	return val, err
}

func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

func (c *Cache) ExistsKey(ctx context.Context, key string) (bool, error) {
	val, err := c.Client.Exists(ctx, key).Result()
	return val == 1, err
}

func (c *Cache) SetIfNotExist(ctx context.Context, key string, val any, exp time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, val, exp).Result()
}

func (c *Cache) Delete(ctx context.Context, key string) (bool, error) {
	delCount, err := c.Client.Del(ctx, key).Result()
	return delCount == 1, err
}

func (c *Cache) DeleteAll(ctx context.Context, keys ...string) (int64, error) {
	delCount, err := c.Client.Del(ctx, keys...).Result()
	return delCount, err
}
