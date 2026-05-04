package bootstrap

import "time"

type config struct {
	Address           string
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
}

func newConfig(env *Env) *config {
	return &config{
		Address:           env.ServerAddress,
		WriteTimeout:      env.WriteTimeout,
		ReadTimeout:       env.ReadTimeout,
		ReadHeaderTimeout: env.ReadHeaderTimeout,
		IdleTimeout:       env.IdleTimeout,
	}
}
