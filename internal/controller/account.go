package controller

import (
	"net/http"
	"unitly/internal/service/auth"
	"unitly/pkg/jsonutil"
)

type AccountController struct {
	authSvc *auth.AuthService
}

type requestOTPRequest struct {
	Phone string `json:"phone"`
}

type verifyOTPRequest struct {
	Sign string `json:"sign"`
	Otp  string `json:"otp"`
}

func NewAccountController(authSvc *auth.AuthService) *AccountController {
	return &AccountController{
		authSvc: authSvc,
	}
}

func (c *AccountController) RequestAuthByOTP(w http.ResponseWriter, r *http.Request) {
	var req requestOTPRequest
	if err := jsonutil.ReadJSON(w, r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, ErrInvalidReqest.Error())
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
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, ErrInvalidReqest.Error())
		return
	}

	token, err := c.authSvc.VerifyAuthByOTP(r.Context(), req.Sign, req.Otp)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err.Error())
		return
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, token); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, "failed to write response")
	}
}
