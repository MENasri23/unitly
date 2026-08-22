package container

import (
	"time"
	"unitly/internal/bootstrap"
	"unitly/internal/cache"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newPostgresDB(app *bootstrap.Application) (*pgxpool.Pool, error) {
	return db.NewPostgresDB(app.GetDatabaseURL(), 25, 10, 1*time.Hour, 30*time.Minute, 1*time.Minute)
}

func newCacheClient(app *bootstrap.Application) (domain.Cache, error) {
	return cache.Default(app.GetCacheAddress(), app.GetCacheUsername(), app.GetCachePassword(), 0)
}
