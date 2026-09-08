package container

import (
	"fmt"
	"unitly/internal/bootstrap"
	"unitly/internal/building"
	"unitly/internal/db"
	"unitly/internal/domain"
	"unitly/internal/log"
	"unitly/internal/ratelimit"
	"unitly/internal/service/auth"
	"unitly/internal/service/identity"
	otpsvc "unitly/internal/service/otp"
	sessionsvc "unitly/internal/service/session"
	"unitly/internal/service/sms"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Services struct {
	Token    domain.TokenService
	SMS      domain.SmsService
	Session  domain.SessionService
	OTP      domain.OTPService
	Identity identity.IdentityService
	Auth     *auth.AuthService
	Building building.BuildingService
}

func newServices(app *bootstrap.Application, logger log.Logger, pgxDB *pgxpool.Pool, cache domain.Cache, repos Repositories) (Services, error) {
	otpCfg := app.GetOtpConfig()
	otpRateLimiter, err := ratelimit.NewTokenBucketLimiter(cache, ratelimit.Config{
		Capacity:   float64(otpCfg.RequestLimit),
		RefillRate: float64(otpCfg.RequestLimit) / otpCfg.Window.Seconds(),
		TTL:        otpCfg.Window,
	})
	if err != nil {
		return Services{}, fmt.Errorf("failed to create OTP rate limiter: %w", err)
	}

	txManager := db.NewSqlTxManager(pgxDB)
	smsSvc := newSmsService(app, logger)
	tokenSvc := auth.NewJWTService(app.GetAccessTokenSecret(), app.GetAccessTokenExpiry())

	sessionSvc := sessionsvc.NewSessionService(
		repos.Session, repos.User, tokenSvc, txManager,
		app.GetRefreshTokenExpiry(), app.GetResetPasswordTokenExpiry(),
	)
	otpSvc := otpsvc.NewOTPService(repos.OTP, smsSvc, otpRateLimiter)
	identitySvc := identity.NewIdentityService(repos.User, sessionSvc, otpSvc)
	authSvc := auth.NewAuthService(otpSvc, identitySvc, sessionSvc, smsSvc, tokenSvc, txManager)
	buildingSvc := building.NewBuildingService(repos.Building, repos.Membership, repos.Floor, txManager)

	return Services{
		Token:    tokenSvc,
		SMS:      smsSvc,
		Session:  sessionSvc,
		OTP:      otpSvc,
		Identity: identitySvc,
		Auth:     authSvc,
		Building: buildingSvc,
	}, nil
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
