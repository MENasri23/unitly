package controller

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/middleware"
	"unitly/internal/http/reqctx"
	"unitly/internal/log"
	"unitly/internal/service/auth"
	"unitly/internal/service/identity"
)

type ProfileController struct {
	authSvc     *auth.AuthService
	identitySvc identity.IdentityService
	authMw      *middleware.AuthMiddleware
	logger      log.Logger
}

func NewProfileController(
	authSvc *auth.AuthService,
	identitySvc identity.IdentityService,
	authMw *middleware.AuthMiddleware,
	logger log.Logger,
) *ProfileController {
	return &ProfileController{authSvc: authSvc, identitySvc: identitySvc, authMw: authMw, logger: logger}
}

func (c *ProfileController) Routes(r chi.Router) {
	r.Route("/profile", func(r chi.Router) {
		r.Use(c.authMw.RequireAuth, c.authMw.RequireUser)
		r.Get("/", c.getProfileInfo)
		r.Patch("/name", c.updateFullName)
		r.Patch("/password", c.updatePassword)
		r.Route("/phone", func(r chi.Router) {
			r.Post("/request", c.updatePhoneRequest)
			r.Post("/verify", c.updatePhoneVerify)
		})
	})
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
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password" validate:"required"`
}

type updatePhoneRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
}

type updatePhoneVerify struct {
	Sign string `json:"sign" validate:"required,max=256"`
	Otp  string `json:"otp" validate:"required,number,len=6"`
}

func (c *ProfileController) updateFullName(w http.ResponseWriter, r *http.Request) {
	var req updateFullNameRequest
	if err := apphttp.ReadJSON(r, &req); err != nil {
		apphttp.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(&req); err != nil {
		apphttp.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}
	user := reqctx.GetUser(r.Context())
	user.FirstName = req.FirstName
	user.LastName = req.LastName

	if err := c.identitySvc.UpdateFullName(r.Context(), user); err != nil {
		apphttp.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ProfileController) getProfileInfo(w http.ResponseWriter, r *http.Request) {
	user := reqctx.GetUser(r.Context())

	dto := profileInfoDTO{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		HasPassword: user.IsPasswordSet(),
		Phone:       user.Phone,
	}

	if err := apphttp.WriteJSON(w, http.StatusOK, dto); err != nil {
		c.logger.Error("failed to write response", "error", err)
	}
}

func (c *ProfileController) updatePassword(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[updatePasswordRequest](r)
	if err != nil {
		apphttp.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	user := reqctx.GetUser(r.Context())
	sessionID, ok := reqctx.GetSessionID(r.Context())
	if !ok {
		apphttp.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
		return
	}

	if err := c.identitySvc.UpdatePassword(r.Context(), user, sessionID, req.OldPassword, req.NewPassword); err != nil {
		apphttp.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ProfileController) updatePhoneRequest(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[updatePhoneRequest](r)
	if err != nil {
		apphttp.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	user := reqctx.GetUser(r.Context())
	token, err := c.identitySvc.UpdatePhoneRequest(r.Context(), user, req.Phone)
	if err != nil {
		apphttp.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	writeJSON(w, http.StatusOK, token, c.logger)
}

func (c *ProfileController) updatePhoneVerify(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[updatePhoneVerify](r)
	if err != nil {
		apphttp.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	user := reqctx.GetUser(r.Context())
	if err = c.identitySvc.UpdatePhoneVerify(r.Context(), user, req.Sign, req.Otp); err != nil {
		apphttp.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
