package bootstrap

import (
	"net/http"
	"time"
)

type Application struct {
	config *config
}

func CreateAppliaction() *Application {
	env := NewEnv()
	cfg := newConfig(env)

	return &Application{
		config: cfg,
	}
}

func (app *Application) GetDatabaseURL() string {
	return app.config.db.GetURL()
}

func (app *Application) GetCacheAddress() string {
	return app.config.cache.GetAddress()
}

func (app *Application) GetCacheUsername() string {
	return app.config.cache.Username
}

func (app *Application) GetCachePassword() string {
	return app.config.cache.Password
}

func (app *Application) GetTokenSecret() string {
	return app.config.token.accessTokenSecret
}

func (app *Application) GetTokenExpiry() time.Duration {
	return app.config.token.accessTokenExpiry
}

func (app *Application) GetSmsUsername() string {
	return app.config.otp.SmsConfig.Username
}

func (app *Application) GetSmsApiKey() string {
	return app.config.otp.SmsConfig.ApiKey
}

func (app *Application) GetOtpSecret() string {
	return app.config.otp.Secret
}

func (app *Application) GetOtpTemplateID() string {
	return app.config.otp.TempID
}

func (app *Application) GetOtpRateLimit() time.Duration {
	return app.config.otp.RateLimit
}

func (app *Application) Run(mux http.Handler) error {
	cfg := app.config
	server := http.Server{
		Handler:           mux,
		Addr:              cfg.Address,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	return server.ListenAndServe()
}
