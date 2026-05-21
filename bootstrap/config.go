package bootstrap

import (
	"fmt"
	"strings"
	"time"
)

type config struct {
	Address           string
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	db                *dbConfig
	cache             *cacheConfig
}

func newConfig(env *Env) *config {
	return &config{
		Address:           env.ServerAddress,
		WriteTimeout:      env.WriteTimeout,
		ReadTimeout:       env.ReadTimeout,
		ReadHeaderTimeout: env.ReadHeaderTimeout,
		IdleTimeout:       env.IdleTimeout,
		db:                newDbConfig(env),
		cache:             NewCacheConfig(env),
	}
}

type Args map[string]string

type dbConfig struct {
	Name     string
	Host     string
	Port     string
	Username string
	Password string
	Driver   string
	Args     Args
}

func newDbConfig(env *Env) *dbConfig {
	args := Args{}
	if env.AppEnv == Development {
		args["sslmode"] = "disable"
	}

	return &dbConfig{
		Driver:   "postgres",
		Name:     env.DBName,
		Host:     env.DBHost,
		Port:     env.DBPort,
		Username: env.DBUser,
		Password: env.DBPass,
		Args:     args,
	}
}

func (c *dbConfig) GetURL() string {
	var sb strings.Builder

	sb.WriteString(c.Driver)
	sb.WriteString("://")

	if len(c.Username) > 0 || len(c.Password) > 0 {
		sb.WriteString(c.Username)
		if len(c.Password) > 0 {
			sb.WriteRune(':')
			sb.WriteString(c.Password)
		}
		sb.WriteRune('@')
	}

	sb.WriteString(c.Host)

	if len(c.Port) > 0 {
		sb.WriteRune(':')
		sb.WriteString(c.Port)
	}

	sb.WriteRune('/')
	sb.WriteString(c.Name)

	if len(c.Args) > 0 {
		sb.WriteRune('?')
		first := true
		for key, val := range c.Args {
			if !first {
				sb.WriteRune('&')
			}

			sb.WriteString(key)
			sb.WriteRune('=')
			sb.WriteString(val)

			if first {
				first = false
			}
		}
	}

	return sb.String()
}

type cacheConfig struct {
	Host     string
	Username string
	Password string
	Port     string
}

func (c *cacheConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func NewCacheConfig(env *Env) *cacheConfig {
	return &cacheConfig{
		Host:     env.RedisHost,
		Port:     env.RedisPort,
		Username: env.RedisUser,
		Password: env.RedisPass,
	}
}
