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
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	appmiddleware "unitly/internal/controller/middleware"
)

const (
	// MigrationsDir is the directory where migration files are stored.
	migrationsSrc = "file://migrations"

	// defalut timeout which signals to to the request
	requestTimeout = 60 * time.Second

	// maxReadBytes defines the maximum allowed size of a JSON request body.
	// This prevents clients from sending excessively large payloads that could
	// exhaust memory or degrade service performance.
	//
	// Default: 1MB
	maxReadBytes int64 = 1024 * 1024
)

func main() {
	logLevel := &slog.LevelVar{}
	logOptions := &slog.HandlerOptions{Level: logLevel}
	logger := sloglogger.NewSlogLogger(logOptions)

	app, err := bootstrap.CreateAppliaction()
	if err != nil {
		logger.Error(fmt.Sprintf("initialize application failed: %v", err))
		return
	}
	logLevel.Set(getLogLevel(app))

	pgxDb, err := newPostgresDB(app)
	if err != nil {
		logger.Error(fmt.Sprintf("postgres connection failed: %v", err))
		return
	}
	defer pgxDb.Close()
	logger.Info("postgres database connected successfully")

	logger.Info("initializing database migrations...")
	if err := runDatabaseMigrations(migrationsSrc, app.GetDatabaseURL(), logger); err != nil {
		logger.Error(fmt.Sprintf("fatal error during migration: %v", err))
		return
	}
	logger.Info("migrations complete")

	logger.Info("starting cache client initialization...")
	cache, err := newCacheClient(app)
	if err != nil {
		logger.Error(fmt.Sprintf("cache client ping failed: %v", err))
		return
	}
	defer cache.Close()
	logger.Info("cache client connected successfully")

	// repositories
	userRepository := repository.NewUserRepository(pgxDb)
	otpRepository := newOtpRepository(app, cache)
	sessionRepository := repository.NewSessionRepository(pgxDb)

	// rate limiters
	otpRateLimiter := otprepo.NewOtpRateLimiter(cache, app.GetOtpRateLimit())

	txManager := db.NewSqlTxManager(pgxDb)

	smsSvc := newSmsService(app, logger)
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

	authMiddleware := appmiddleware.NewAuthMiddleware(
		tokenService,
		userRepository,
	)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(middleware.RequestSize(maxReadBytes))

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			controller.SetupAccountRoutes(r, authService)

			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.AuthorizeUser)
				r.Get("/healthcheck", controller.HealthCheck)
			})
		})
	})

	server := app.NewServer(r)

	logger.Info(fmt.Sprintf("trying to listen on %s", app.GetServerAddress()))
	if err = server.ListenAndServe(); err != nil {
		logger.Error("server failed to start: %v", err)
		return
	}

}

func getLogLevel(app *bootstrap.Application) slog.Level {
	if app.DebugMode() {
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
	if app.DebugMode() {
		otpRepo = otprepo.NewOTPMockRepo(cache, app.GetOtpSecret())
	} else {
		otpRepo = otprepo.NewOtpRepository(cache, app.GetOtpSecret())
	}

	return otpRepo
}

func newSmsService(app *bootstrap.Application, logger log.Logger) domain.SmsService {
	if app.DebugMode() {
		return &smsServiceMock{}
	}

	return sms.NewMeliPayamakService(
		app.GetSmsUsername(),
		app.GetSmsApiKey(),
		app.GetOtpTemplateID(),
		logger,
	)
}

// runDatabaseMigrations handles the initialization and execution of DB schema updates.
func runDatabaseMigrations(sourceURL, dbURL string, logger log.Logger) error {
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
