package controller

import (
	"net/http"
	"unitly/internal/controller/middleware"
	"unitly/internal/controller/reqctx"
	"unitly/internal/domain"
	"unitly/internal/log"
	"unitly/internal/service/auth"
	"unitly/internal/service/identity"
	"unitly/pkg/jsonutil"

	"github.com/go-chi/chi/v5"
)

type profileController struct {
	authSvc     *auth.AuthService
	identitySvc identity.IdentityService
	logger      log.Logger
}

type updateFullNameRequest struct {
	FirstName *string `json:"first_name" validate:"omitzero,profile_name"`
	LastName  *string `json:"last_name" validate:"omitzero,profile_name"`
}

type profileInfoDTO struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	HasPassword bool    `json:"has_password"`
	Phone       string  `json:"phone"`
}

type updatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

func (c *profileController) updateFullName(w http.ResponseWriter, r *http.Request) {
	var req updateFullNameRequest
	if err := jsonutil.ReadJSON(r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(&req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}
	user := reqctx.GetUser(r.Context())
	user.FirstName = req.FirstName
	user.LastName = req.LastName

	if err := c.identitySvc.UpdateFullName(r.Context(), user); err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *profileController) getProfileInfo(w http.ResponseWriter, r *http.Request) {
	user := reqctx.GetUser(r.Context())

	dto := profileInfoDTO{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		HasPassword: user.IsPasswordSet(),
		Phone:       user.Phone,
	}

	if err := jsonutil.WriteJSON(w, http.StatusOK, dto); err != nil {
		c.logger.Error("failed to write response", "error", err)
	}
}

func (c *profileController) updatePassword(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[updatePasswordRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	user := reqctx.GetUser(r.Context())
	sessionID, ok := reqctx.GetSessionID(r.Context())
	if !ok {
		jsonutil.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
		return
	}

	if err := c.identitySvc.UpdatePassword(r.Context(), user, sessionID, req.OldPassword, req.NewPassword); err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SetupProfileRoutes(
	r chi.Router,
	authService *auth.AuthService,
	identityService identity.IdentityService,
	authMiddleware *middleware.AuthMiddleware,
	logger log.Logger,
) {
	controller := &profileController{
		authSvc:     authService,
		identitySvc: identityService,
		logger:      logger,
	}

	r.Route("/profiles", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)
		r.Get("/info", controller.getProfileInfo)
		r.Patch("/change-name", controller.updateFullName)
		r.Patch("/change-password", controller.updatePassword)
	})
}
