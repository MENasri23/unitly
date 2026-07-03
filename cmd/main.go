package main

import (
	"log"
	"time"
	"unitly/bootstrap"
	"unitly/internal/cache"
	"unitly/internal/controller"
	"unitly/internal/db"
	"unitly/internal/domain"
	"unitly/internal/repository"
	otprepo "unitly/internal/repository/otp"
	"unitly/internal/service/auth"
	"unitly/internal/service/sms"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	app := bootstrap.CreateAppliaction()
	pgxDb, err := newPostgresDB(app)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
		return
	}
	defer pgxDb.Close()

	cache, err := newCacheClient(app)
	if err != nil {
		log.Fatalf("redis ping failed: %v", err)
		return
	}
	defer cache.Close()

	// repositories
	userRepository := repository.NewUserRepository(pgxDb)
	otpRepository := newOtpRepository(app, cache)
	sessionRepository := repository.NewSessionRepository(pgxDb)

	// rate limiters
	otpRateLimiter := otprepo.NewOtpRateLimiter(cache, app.GetOtpRateLimit())

	txManager := db.NewSqlTxManager(pgxDb)

	smsSvc := newSmsService(app)
	tokenService := auth.NewJWTService(app.GetTokenSecret(), app.GetTokenExpiry())
	authService := auth.NewAuthService(
		otpRepository,
		userRepository,
		sessionRepository,
		smsSvc,
		tokenService,
		otpRateLimiter,
		txManager,
	)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			controller.SetupAccountRoutes(r, authService)
		})
	})

	app.Run(r)
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
