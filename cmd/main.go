package main

import (
	"fmt"
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

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	app := bootstrap.CreateAppliaction()
	db, err := newPostgresDB(app)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	cache, err := newCacheClient(app)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer cache.Close()

	// repositories
	userRepository := repository.NewUserRepository(db)
	otpRepository := otprepo.NewOtpRepository(cache, app.GetOtpSecret())

	// rate limiters
	otpRateLimiter := otprepo.NewOtpRateLimiter(cache, app.GetOtpRateLimit())

	// services
	// smsService := sms.NewMeliPayamakService(
	// 	app.GetSmsUsername(),
	// 	app.GetSmsApiKey(),
	// 	app.GetOtpTemplateID(),
	// )

	var smsSvc domain.SmsService = &smsServiceMock{}
	tokenService := auth.NewJWTService(app.GetTokenSecret(), app.GetTokenExpiry())
	authService := auth.NewAuthService(
		otpRepository,
		userRepository,
		smsSvc,
		tokenService,
		otpRateLimiter,
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
