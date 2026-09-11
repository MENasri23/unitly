package container

import (
	"unitly/internal/bootstrap"
	"unitly/internal/building"
	bldgimpl "unitly/internal/building/bldg"
	"unitly/internal/building/floor"
	"unitly/internal/building/membership"
	"unitly/internal/building/unit"
	"unitly/internal/domain"
	"unitly/internal/repository"
	otprepo "unitly/internal/repository/otp"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	User       domain.UserRepository
	OTP        domain.OTPRepository
	Session    domain.SessionRepository
	Building   building.BuildingRepository
	Membership building.MembershipRepository
	Floor      building.FloorRepository
	Unit       building.UnitRepository
}

func newRepositories(app *bootstrap.Application, pgxDB *pgxpool.Pool, cache domain.Cache) Repositories {
	return Repositories{
		User:       repository.NewUserRepository(pgxDB, cache),
		OTP:        newOtpRepository(app, cache),
		Session:    repository.NewSessionRepository(pgxDB),
		Building:   bldgimpl.NewBuildingRepository(pgxDB, cache),
		Membership: membership.NewMembershipRepository(pgxDB, cache),
		Floor:      floor.NewFloorRepository(pgxDB),
		Unit:       unit.NewUnitRepository(pgxDB),
	}
}

func newOtpRepository(app *bootstrap.Application, cache domain.Cache) domain.OTPRepository {
	if app.DebugMode() {
		return otprepo.NewOTPMockRepo(cache)
	}
	return otprepo.NewOtpRepository(cache)
}
