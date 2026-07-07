package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unitly/bootstrap"
	"unitly/internal/cache"
	"unitly/internal/controller"
	"unitly/internal/db"
	"unitly/internal/domain"
	"unitly/internal/log"
	"unitly/internal/repository"
	otprepo "unitly/internal/repository/otp"
	"unitly/internal/service/auth"
	"unitly/internal/service/sms"
	"unitly/pkg/sloglogger"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// MigrationsDir is the directory where migration files are stored.
	migrationsSrc = "file://migrations"
)

func main() {
	app := bootstrap.CreateAppliaction()

	logOptions := &slog.HandlerOptions{Level: getLogLevel(app), AddSource: true}
	logger := sloglogger.NewSlogLogger(logOptions)

	pgxDb, err := newPostgresDB(app)
	if err != nil {
		logger.Error(fmt.Sprintf("Postgres connection failed: %v", err))
		return
	}
	defer pgxDb.Close()
	logger.Info("Postgres database connected successfully")

	logger.Info("Initializing database migrations...")
	if err := runDatabaseMigrations(logger, migrationsSrc, app.GetDatabaseURL()); err != nil {
		logger.Error(fmt.Sprintf("Fatal error during migration: %v", err))
		return
	}
	logger.Info("Migrations complete")

	logger.Info("Starting cache client initialization...")
	cache, err := newCacheClient(app)
	if err != nil {
		logger.Error(fmt.Sprintf("Cache client ping failed: %v", err))
		return
	}
	defer cache.Close()
	logger.Info("Cache client connected successfully")

	// repositories
	userRepository := repository.NewUserRepository(pgxDb)
	otpRepository := newOtpRepository(app, cache)
	sessionRepository := repository.NewSessionRepository(pgxDb)

	// rate limiters
	otpRateLimiter := otprepo.NewOtpRateLimiter(cache, app.GetOtpRateLimit())

	txManager := db.NewSqlTxManager(pgxDb)

	smsSvc := newSmsService(app)
	tokenService := auth.NewJWTService(app.GetAccessTokenSecret(), app.GetAccessTokenExpiry())
	authService := auth.NewAuthService(
		otpRepository,
		userRepository,
		sessionRepository,
		smsSvc,
		tokenService,
		otpRateLimiter,
		txManager,
		app.GetRefreshTokenExpiry(),
	)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			controller.SetupAccountRoutes(r, authService)
		})
	})

	app.Run(r)
}

func getLogLevel(app *bootstrap.Application) slog.Level {
	if app.IsDevelopmentEnv() {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

func newPostgresDB(app *bootstrap.Application) (*pgxpool.Pool, error) {
	return db.NewPostgresDB(
		app.GetDatabaseURL(),
		25,
		10,
		1*time.Hour,
		30*time.Minute,
		1*time.Minute,
	)
}

func newCacheClient(app *bootstrap.Application) (*cache.Cache, error) {
	return cache.Default(
		app.GetCacheAddress(),
		app.GetCacheUsername(),
		app.GetCachePassword(),
		0,
	)
}

func newOtpRepository(app *bootstrap.Application, cache *cache.Cache) domain.OTPRepository {
	var otpRepo domain.OTPRepository
	if app.IsDevelopmentEnv() {
		otpRepo = otprepo.NewOTPMockRepo(cache, app.GetOtpSecret())
	} else {
		otpRepo = otprepo.NewOtpRepository(cache, app.GetOtpSecret())
	}

	return otpRepo
}

func newSmsService(app *bootstrap.Application) domain.SmsService {
	if app.IsDevelopmentEnv() {
		return &smsServiceMock{}
	}

	return sms.NewMeliPayamakService(
		app.GetSmsUsername(),
		app.GetSmsApiKey(),
		app.GetOtpTemplateID(),
	)
}

// runDatabaseMigrations handles the initialization and execution of DB schema updates.
func runDatabaseMigrations(logger log.Logger, sourceURL, dbURL string) error {
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return err
	}

	// Defer closing the underlying database connection and source drivers
	// to prevent memory/connection leaks.
	defer m.Close()

	// Execute the "Up" migrations
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("Schema is up-to-date. No new migrations applied")
			return nil
		}
		return err
	}

	logger.Info("Database migrations applied successfully")
	return nil
}
