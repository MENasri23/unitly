package bootstrap

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type AppEnv string

const (
	Development AppEnv = "development"
	Production  AppEnv = "production"
)

type Env struct {
	AppEnv AppEnv `mapstructure:"APP_ENV"`

	ServerHost        string        `mapstructure:"SERVER_HOST"`
	ServerPort        string        `mapstructure:"SERVER_PORT"`
	WriteTimeout      time.Duration `mapstructure:"WRITE_TIMEOUT"`
	ReadTimeout       time.Duration `mapstructure:"READ_TIMEOUT"`
	ReadHeaderTimeout time.Duration `mapstructure:"READ_HEADER_TIMEOUT"`
	IdleTimeout       time.Duration `mapstructure:"IDLE_TIMEOUT"`
	ContextTimeout    time.Duration `mapstructure:"CONTEXT_TIMEOUT"`

	DBHost string `mapstructure:"DB_HOST"`
	DBPort string `mapstructure:"DB_PORT"`
	DBUser string `mapstructure:"DB_USER"`
	DBPass string `mapstructure:"DB_PASS"`
	DBName string `mapstructure:"DB_NAME"`

	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisUser string `mapstructure:"REDIS_USER"`
	RedisPass string `mapstructure:"REDIS_PASS"`

	AccessTokenSecret string        `mapstructure:"ACCESS_TOKEN_SECRET"`
	AccessTokenExpiry time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRY"`

	OtpSecret     string        `mapstructure:"OTP_SECRET"`
	OtpTemplateID string        `mapstructure:"OTP_TEMPLATE_ID"`
	OtpRateLimit  time.Duration `mapstructure:"OTP_RATE_LIMIT"`

	SmsUsername string `mapstructure:"SMS_USER"`
	SmsApiKey   string `mapstructure:"SMS_API_KEY"`
}

func NewEnv() *Env {
	env := Env{}
	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Can't find the file .env : ", err)
	}

	err = viper.Unmarshal(&env)
	if err != nil {
		log.Fatal("Environment can't be loaded: ", err)
	}

	if env.AppEnv == Development {
		log.Println("The App is running in development env")
	}

	return &env
}
