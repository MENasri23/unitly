package router

import (
	"net/http"
	"time"
	"unitly/internal/controller"
	"unitly/internal/http/header"

	appmw "unitly/internal/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	// default timeout which signals to to the request
	requestTimeout = 60 * time.Second

	// maxReadBytes defines the maximum allowed size of a JSON request body.
	// This prevents clients from sending excessively large payloads that could
	// exhaust memory or degrade service performance.
	//
	// Default: 1MB
	maxReadBytes int64 = 1024 * 1024
)

type Params struct {
	AllowedOrigins []string
	AuthMiddleware *appmw.AuthMiddleware
	Auth           RouteRegistrar
	Profile        RouteRegistrar
	Session        RouteRegistrar
}

func New(p Params) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: p.AllowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodPatch,
		},
		AllowedHeaders: []string{
			header.Accept,
			header.Authorization,
			header.ContentType,
			header.ClientPlatform,
			header.UserAgent,
		},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromXFFTrustedProxies(1))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(middleware.RequestSize(maxReadBytes))

	r.Route("/api/v1", func(r chi.Router) {
		p.Auth.Routes(r)
		r.Route("/me", func(r chi.Router) {
			p.Profile.Routes(r)
			p.Session.Routes(r)
		})
		r.Group(func(r chi.Router) {
			r.Use(p.AuthMiddleware.RequireAuth, p.AuthMiddleware.RequireUser)
			r.Get("/healthcheck", controller.HealthCheck)
		})
	})

	return r
}
