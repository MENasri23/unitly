package container

import (
	"unitly/internal/bootstrap"
	"unitly/internal/domain"
	"unitly/internal/repository"
	otprepo "unitly/internal/repository/otp"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	User    domain.UserRepository
	OTP     domain.OTPRepository
	Session domain.SessionRepository
}

func newRepositories(app *bootstrap.Application, pgxDB *pgxpool.Pool, cache domain.Cache) Repositories {
	return Repositories{
		User:    repository.NewUserRepository(pgxDB),
		OTP:     newOtpRepository(app, cache),
		Session: repository.NewSessionRepository(pgxDB),
	}
}

func newOtpRepository(app *bootstrap.Application, cache domain.Cache) domain.OTPRepository {
	if app.DebugMode() {
		return otprepo.NewOTPMockRepo(cache)
	}
	return otprepo.NewOtpRepository(cache)
}