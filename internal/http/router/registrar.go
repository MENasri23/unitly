package router

import "github.com/go-chi/chi/v5"

// RouteRegistrar is anything that can attach its own routes.
type RouteRegistrar interface {
	Routes(r chi.Router)
}
