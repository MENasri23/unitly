package bootstrap

import (
	"net/http"
)

type Application struct {
	config Config
}

func CreateAppliaction() *Application {
	env := NewEnv()
	config := Config{
		Address:      env.ServerAddress,
		WriteTimeout: env.WriteTimeout,
		ReadTimeout:  env.ReadTimeout,
		IdleTimeout:  env.IdleTimeout,
	}

	return &Application{
		config: config,
	}
}

func (app *Application) Run(mux http.Handler) error {
	server := http.Server{
		Handler:      mux,
		Addr:         app.config.Address,
		WriteTimeout: app.config.WriteTimeout,
		ReadTimeout:  app.config.ReadTimeout,
		IdleTimeout:  app.config.IdleTimeout,
	}
	return server.ListenAndServe()
}
