package bootstrap

import (
	"fmt"
	"time"
	"unitly/internal/cache"
	"unitly/internal/db"
	otpsvc "unitly/internal/service/otp"
	"unitly/internal/service/sms"
	"unitly/internal/token"
)

type config struct {
	Environment       AppEnv
	Address           string
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	AllowedOrigins    []string
	db                *db.DBConfig
	cache             *cache.CacheConfig
	token             *token.TokenConfig
	otp               *otpsvc.OtpConfig
	sms               *sms.SmsConfig
}

func newConfig(env *Env) *config {
	return &config{
		Environment:       env.AppEnv,
		Address:           serverAddress(env),
		WriteTimeout:      env.WriteTimeout,
		ReadTimeout:       env.ReadTimeout,
		ReadHeaderTimeout: env.ReadHeaderTimeout,
		IdleTimeout:       env.IdleTimeout,
		AllowedOrigins:    env.AllowedOrigins,
		db:                newDbConfig(env),
		cache:             newCacheConfig(env),
		token:             newTokenConfig(env),
		otp:               newOtpConfig(env),
		sms:               newSmsConfig(env),
	}
}

func serverAddress(env *Env) string {
	return fmt.Sprintf("%s:%s", env.ServerHost, env.ServerPort)
}

func newDbConfig(env *Env) *db.DBConfig {
	args := db.Args{}
	if env.AppEnv == Development {
		args["sslmode"] = "disable"
	}

	return &db.DBConfig{
		Driver:   "postgres",
		Name:     env.DBName,
		Host:     env.DBHost,
		Port:     env.DBPort,
		Username: env.DBUser,
		Password: env.DBPass,
		Args:     args,
	}
}

func newCacheConfig(env *Env) *cache.CacheConfig {
	return &cache.CacheConfig{
		Host:     env.RedisHost,
		Port:     env.RedisPort,
		Username: env.RedisUser,
		Password: env.RedisPass,
	}
}

func newTokenConfig(env *Env) *token.TokenConfig {
	return &token.TokenConfig{
		AccessTokenSecret:        env.AccessTokenSecret,
		AccessTokenExpiry:        env.AccessTokenExpiry,
		RefreshTokenExpiry:       env.RefreshTokenExpiry,
		ResetPasswordTokenExpiry: env.ResetPasswordTokenExpiry,
	}
}

func newSmsConfig(env *Env) *sms.SmsConfig {
	return &sms.SmsConfig{
		Username: env.SmsUsername,
		ApiKey:   env.SmsApiKey,
	}
}

func newOtpConfig(env *Env) *otpsvc.OtpConfig {
	return &otpsvc.OtpConfig{
		TempID:    env.OtpTemplateID,
		RequestLimit: 1,
		Window: env.OtpRateLimit,
	}
}
