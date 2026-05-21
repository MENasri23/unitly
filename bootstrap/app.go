package bootstrap

import (
	"net/http"
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
