package container

import (
	"fmt"
	"unitly/internal/bootstrap"
	"unitly/internal/http/middleware"
	"unitly/internal/log"
)

type Container struct {
	Repos       Repositories
	Services    Services
	Controllers Controllers
	AuthMW      *middleware.AuthMiddleware
}

// New wires everything and returns a cleanup func for the caller to defer.
func New(app *bootstrap.Application, logger log.Logger) (*Container, func(), error) {
	pgxDB, err := newPostgresDB(app)
	if err != nil {
		return nil, func() {}, fmt.Errorf("postgres connection failed: %w", err)
	}

	cacheClient, err := newCacheClient(app)
	if err != nil {
		pgxDB.Close()
		return nil, func() {}, fmt.Errorf("cache client failed: %w", err)
	}
	cleanup := func() {
		cacheClient.Close()
		pgxDB.Close()
	}

	repos := newRepositories(app, pgxDB, cacheClient)

	services, err := newServices(app, logger, pgxDB, cacheClient, repos)
	if err != nil {
		return nil, cleanup, err
	}

	authMW := middleware.NewAuthMiddleware(services.Token, repos.User)
	controllers := newControllers(app, logger, services, repos, authMW)

	return &Container{
		Repos:       repos,
		Services:    services,
		Controllers: controllers,
		AuthMW:      authMW,
	}, cleanup, nil
}
