package bootstrap

import (
	"net/http"
	"time"
)

type Application struct {
	config *config
}

func CreateApplication() (*Application, error) {
	env, err := NewEnv()
	if err != nil {
		return nil, err
	}

	return &Application{
		config: newConfig(env),
	}, nil
}

func (app *Application) DebugMode() bool {
	return app.config.Environment == Development
}

func (app *Application) GetServerAddress() string {
	return app.config.Address
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

func (app *Application) GetAccessTokenSecret() string {
	return app.config.token.AccessTokenSecret
}

func (app *Application) GetAccessTokenExpiry() time.Duration {
	return app.config.token.AccessTokenExpiry
}

func (app *Application) GetRefreshTokenExpiry() time.Duration {
	return app.config.token.RefreshTokenExpiry
}

func (app *Application) GetResetPasswordTokenExpiry() time.Duration {
	expiry := app.config.token.ResetPasswordTokenExpiry
	if expiry == 0 {
		return 5 * time.Minute
	}
	return expiry
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

func (app *Application) AllowedOrigins() []string {
	return app.config.AllowedOrigins
}

func (app *Application) NewServer(mux http.Handler) *http.Server {
	cfg := app.config
	return &http.Server{
		Handler:           mux,
		Addr:              cfg.Address,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}
