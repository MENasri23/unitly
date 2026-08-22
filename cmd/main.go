package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"unitly/internal/bootstrap"
	"unitly/internal/cache"
	"unitly/internal/controller"
	"unitly/internal/db"
	"unitly/internal/domain"
	"unitly/internal/http/header"
	"unitly/internal/log"
	"unitly/internal/ratelimit"
	"unitly/internal/repository"
	otprepo "unitly/internal/repository/otp"
	"unitly/internal/service/auth"
	identitysvc "unitly/internal/service/identity"
	otpsvc "unitly/internal/service/otp"
	sessionsvc "unitly/internal/service/session"
	"unitly/internal/service/sms"
	"unitly/pkg/sloglogger"

	appmiddleware "unitly/internal/controller/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
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

func main() {
	logLevel := &slog.LevelVar{}
	logOptions := &slog.HandlerOptions{Level: logLevel}
	logger := sloglogger.NewSlogLogger(logOptions)

	app, err := bootstrap.CreateApplication()
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
	otpCfg := app.GetOtpConfig()
	otpRateLimiter, err := ratelimit.NewTokenBucketLimiter(cache, ratelimit.Config{
		Capacity:   float64(otpCfg.RequestLimit),
		RefillRate: float64(otpCfg.RequestLimit) / otpCfg.Window.Seconds(),
		TTL:        otpCfg.Window,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create OTP rate limiter: %v", err))
		return
	}

	txManager := db.NewSqlTxManager(pgxDb)

	smsSvc := newSmsService(app, logger)
	tokenService := auth.NewJWTService(app.GetAccessTokenSecret(), app.GetAccessTokenExpiry())
	sessionService := sessionsvc.NewSessionService(
		sessionRepository,
		userRepository,
		tokenService,
		txManager,
		app.GetRefreshTokenExpiry(),
		app.GetResetPasswordTokenExpiry(),
	)
	otpService := otpsvc.NewOTPService(
		otpRepository,
		smsSvc,
		otpRateLimiter,
	)
	identityService := identitysvc.NewIdentityService(userRepository, sessionService, otpService)
	authService := auth.NewAuthService(
		otpService,
		identityService,
		sessionService,
		smsSvc,
		tokenService,
		txManager,
	)

	authMiddleware := appmiddleware.NewAuthMiddleware(
		tokenService,
		userRepository,
	)

	authController := controller.NewAuthController(
		authService,
		identityService,
		authMiddleware,
		logger,
		controller.AuthConfig{
			RefreshTokenExpiry: app.GetRefreshTokenExpiry(),
			AccessTokenExpiry:  app.GetAccessTokenExpiry(),
			DebugMode:          app.DebugMode(),
		},
	)
	profileController := controller.NewProfileController(authService, identityService, authMiddleware, logger)
	sessionController := controller.NewSessionController(sessionService, authMiddleware, logger)

	sessionsvc.RunSessionCleanupWorker(context.Background(), sessionRepository, 10*time.Minute)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: app.AllowedOrigins(),
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

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			authController.Routes(r)
			r.Route("/me", func(r chi.Router) {
				profileController.Routes(r)
				sessionController.Routes(r)
			})

			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Get("/healthcheck", controller.HealthCheck)
			})
		})
	})

	if app.DebugMode() {
		logRoutes(r, logger)
	}

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

func newCacheClient(app *bootstrap.Application) (domain.Cache, error) {
	return cache.Default(
		app.GetCacheAddress(),
		app.GetCacheUsername(),
		app.GetCachePassword(),
		0,
	)
}

func newOtpRepository(app *bootstrap.Application, cache domain.Cache) domain.OTPRepository {
	var otpRepo domain.OTPRepository
	if app.DebugMode() {
		otpRepo = otprepo.NewOTPMockRepo(cache)
	} else {
		otpRepo = otprepo.NewOtpRepository(cache)
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
		app.GetOtpConfig().TempID,
		logger,
	)
}

func logRoutes(r chi.Router, l log.Logger) {
	_ = chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		l.Info(fmt.Sprintf("%-7s %s", method, route))
		return nil
	})
}
