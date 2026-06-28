package controller

import (
	"net/http"
	"unitly/internal/service/auth"
	"unitly/pkg/jsonutil"

	"github.com/go-chi/chi/v5"
)

type AccountController struct {
	authSvc *auth.AuthService
}

type requestOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
}

type verifyOTPRequest struct {
	Sign string `json:"sign" validate:"required,base64rawurl"`
	Otp  string `json:"otp" validate:"required,number,len=6"`
}

type finalizeResetPasswordRequest struct {
	Sign       string `json:"token" validate:"required,jwt"`
	NewPassword string `json:"password" validate:"required,min=8,max=64"`
}

func NewAccountController(authSvc *auth.AuthService) *AccountController {
	return &AccountController{
		authSvc: authSvc,
	}
}

func (c *AccountController) RequestAuthByOTP(w http.ResponseWriter, r *http.Request) {
	var req requestOTPRequest
	if err := jsonutil.ReadJSON(w, r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := c.authSvc.RequestAuthByOTP(r.Context(), req.Phone)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err.Error())
		return
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, token); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, "failed to write response")
	}
}

func (c *AccountController) VerifyAuthByOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := jsonutil.ReadJSON(w, r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.authSvc.VerifyAuthByOTP(r.Context(), req.Sign, req.Otp)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err.Error())
		return
	}

	envelop := struct {
		AccessToken  string `json:"access-token"`
		RefreshToken string `json:"refresh-token"`
	}{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}
	if err = jsonutil.WriteJSON(w, http.StatusOK, envelop); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, "failed to write response")
	}
}

func (c *AccountController) RequestResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	var req requestOTPRequest
	if err := jsonutil.ReadJSON(w, r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := c.authSvc.RequestResetPasswordByOTP(r.Context(), req.Phone)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err.Error())
		return
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, token); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, "failed to write response")
	}
}

func (c *AccountController) VerifyResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := jsonutil.ReadJSON(w, r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := c.authSvc.VerifyResetPasswordByOTP(r.Context(), req.Sign, req.Otp)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err.Error())
		return
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, token); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, "failed to write response")
	}
}

func (c *AccountController) FinalizeResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	var req finalizeResetPasswordRequest
	if err := jsonutil.ReadJSON(w, r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.authSvc.FinalizeResetPasswordByOTP(r.Context(), req.Sign, req.NewPassword); err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err.Error())
		return
	}

	if err := jsonutil.WriteJSON(w, http.StatusOK, true); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, "failed to write response")
	}
}

func SetupAccountRoutes(
	r chi.Router,
	authService *auth.AuthService,
) {
	accountController := NewAccountController(authService)

	r.Route("/accounts", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Route("/by-otp", func(r chi.Router) {
				r.Post("/request", accountController.RequestAuthByOTP)
				r.Post("/verify", accountController.VerifyAuthByOTP)
			})
		})

		r.Route("/reset-password", func(r chi.Router) {
			r.Route("/by-otp", func(r chi.Router) {
				r.Post("/request", accountController.RequestResetPasswordByOTP)
				r.Post("/verify", accountController.VerifyResetPasswordByOTP)
				r.Post("/finalize", accountController.FinalizeResetPasswordByOTP)
			})
		})

	})

}
