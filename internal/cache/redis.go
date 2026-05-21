package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Client *redis.Client
}

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

func (c *Cache) Close() error {
	return c.Client.Close()
}
