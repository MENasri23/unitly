package container

import (
	"unitly/internal/bootstrap"
	bldgimpl "unitly/internal/building/building"
	"unitly/internal/controller"
	"unitly/internal/http/middleware"
	"unitly/internal/log"
)

type Controllers struct {
	Auth     *controller.AuthController
	Profile  *controller.ProfileController
	Session  *controller.SessionController
	Building *bldgimpl.BuildingController
}

func newControllers(
	app *bootstrap.Application,
	logger log.Logger,
	s Services,
	authMW *middleware.AuthMiddleware,
) Controllers {
	return Controllers{
		Auth: controller.NewAuthController(s.Auth, s.Identity, authMW, logger, controller.AuthConfig{
			RefreshTokenExpiry: app.GetRefreshTokenExpiry(),
			AccessTokenExpiry:  app.GetAccessTokenExpiry(),
			DebugMode:          app.DebugMode(),
		}),
		Profile:  controller.NewProfileController(s.Auth, s.Identity, authMW, logger),
		Session:  controller.NewSessionController(s.Session, authMW, logger),
		Building: bldgimpl.NewBuildingController(s.Building, authMW.RequireAuth),
	}
}
