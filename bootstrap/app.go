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
		Address: env.ServerAddress,
	}

	return &Application{
		config: config,
	}
}

func (app *Application) Run(mux http.Handler) error {
	return http.ListenAndServe(app.config.Address, mux)
}
