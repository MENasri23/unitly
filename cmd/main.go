package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"unitly/internal/bootstrap"
	"unitly/internal/container"
	"unitly/internal/http/router"
	"unitly/internal/log"
	"unitly/pkg/sloglogger"

	"github.com/go-chi/chi/v5"
)

func main() {
	logLevel := &slog.LevelVar{}
	logger := sloglogger.NewSlogLogger(&slog.HandlerOptions{Level: logLevel})

	app, err := bootstrap.CreateApplication()
	if err != nil {
		logger.Error(fmt.Sprintf("initialize application failed: %v", err))
		return
	}
	logLevel.Set(getLogLevel(app))

	c, cleanup, err := container.New(app, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to build container: %v", err))
		return
	}
	defer cleanup()

	r := router.New(router.Params{
		AllowedOrigins: app.AllowedOrigins(),
		AuthMiddleware: c.AuthMW,
		Auth:           c.Controllers.Auth,
		Profile:        c.Controllers.Profile,
		Session:        c.Controllers.Session,
		Building:       c.Controllers.Building,
	})

	if app.DebugMode() {
		logRoutes(r, logger)
	}

	server := app.NewServer(r)
	logger.Info(fmt.Sprintf("trying to listen on %s", app.GetServerAddress()))
	if err := server.ListenAndServe(); err != nil {
		logger.Error(fmt.Sprintf("server failed to start: %v", err))
	}
}

func getLogLevel(app *bootstrap.Application) slog.Level {
	if app.DebugMode() {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

func logRoutes(r chi.Router, l log.Logger) {
	_ = chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		l.Info(fmt.Sprintf("%-7s %s", method, route))
		return nil
	})
}
