package bootstrap

import "time"

type Config struct {
	Address      string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	IdleTimeout  time.Duration
}
