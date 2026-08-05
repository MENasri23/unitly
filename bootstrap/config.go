package bootstrap

import (
	"fmt"
	"time"
	"unitly/internal/db"
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
	cache             *cacheConfig
	token             *tokenConfig
	otp               *otpConfig
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
		otp:               newOtpConfig(env, newSmsConfig(env)),
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

type cacheConfig struct {
	Host     string
	Username string
	Password string
	Port     string
}

func (c *cacheConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func newCacheConfig(env *Env) *cacheConfig {
	return &cacheConfig{
		Host:     env.RedisHost,
		Port:     env.RedisPort,
		Username: env.RedisUser,
		Password: env.RedisPass,
	}
}

type tokenConfig struct {
	accessTokenSecret        string
	accessTokenExpiry        time.Duration
	refreshTokenExpiry       time.Duration
	resetPasswordTokenExpiry time.Duration
}

func newTokenConfig(env *Env) *tokenConfig {
	return &tokenConfig{
		accessTokenSecret:        env.AccessTokenSecret,
		accessTokenExpiry:        env.AccessTokenExpiry,
		refreshTokenExpiry:       env.RefreshTokenExpiry,
		resetPasswordTokenExpiry: env.ResetPasswordTokenExpiry,
	}
}

type smsConfig struct {
	Username string
	ApiKey   string
}

func newSmsConfig(env *Env) *smsConfig {
	return &smsConfig{
		Username: env.SmsUsername,
		ApiKey:   env.SmsApiKey,
	}
}

type otpConfig struct {
	Secret    string
	TempID    string
	RateLimit time.Duration
	SmsConfig *smsConfig
}

func newOtpConfig(env *Env, smsCfg *smsConfig) *otpConfig {
	return &otpConfig{
		Secret:    env.OtpSecret,
		TempID:    env.OtpTemplateID,
		RateLimit: env.OtpRateLimit,
		SmsConfig: smsCfg,
	}
}
