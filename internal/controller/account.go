package controller

import (
	"net/http"
	"time"
	"unitly/internal/controller/middleware"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/log"
	"unitly/internal/service/auth"
	"unitly/internal/service/identity"
	"unitly/pkg/jsonutil"

	"github.com/go-chi/chi/v5"
)

const refreshTokenCookiePath = "/api/v1/accounts/auth/refresh"

type AccountController struct {
	authSvc            *auth.AuthService
	identitySvc        identity.IdentityService
	logger             log.Logger
	refreshTokenExpiry time.Duration
	accessTokenExpiry  time.Duration
	debugMode          bool
}

type requestOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
}

type verifyOTPRequest struct {
	Sign string `json:"sign" validate:"required,base64rawurl,max=256"`
	Otp  string `json:"otp" validate:"required,number,len=6"`
}

type authByPasswordRequest struct {
	Phone    string `json:"phone" validate:"required,e164"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type finalizeResetPasswordRequest struct {
	Sign        string `json:"sign" validate:"required,jwt,max=512"`
	NewPassword string `json:"password" validate:"required,min=8,max=64"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,base64rawurl,max=256"`
}

type TokensDTO struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (c *AccountController) writeJSON(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, v, c.logger)
}

// respondTokens consolidates the "web client gets cookies, native client
// gets a JSON body" branch that was repeated in four handlers.
func (c *AccountController) respondTokens(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	if apphttp.IsWebPlatform(r) {
		c.setAccessTokenCookie(w, accessToken)
		c.setRefreshTokenCookie(w, refreshToken)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	c.writeJSON(w, http.StatusOK, TokensDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (c *AccountController) RequestAuthByOTP(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[requestOTPRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	token, err := c.authSvc.RequestAuthByOTP(r.Context(), req.Phone)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.writeJSON(w, http.StatusOK, token)
}

func (c *AccountController) VerifyAuthByOTP(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[verifyOTPRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	result, err := c.authSvc.VerifyAuthByOTP(r.Context(), req.Sign, req.Otp)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.respondTokens(w, r, result.AccessToken, result.RefreshToken)
}

func (c *AccountController) AuthenticateByPassword(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[authByPasswordRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	result, err := c.authSvc.AuthenticateByPassword(r.Context(), req.Phone, req.Password)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.respondTokens(w, r, result.AccessToken, result.RefreshToken)
}

func (c *AccountController) RequestResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[requestOTPRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	token, err := c.authSvc.RequestResetPasswordByOTP(r.Context(), req.Phone)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.writeJSON(w, http.StatusOK, token)
}

func (c *AccountController) VerifyResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[verifyOTPRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	token, err := c.authSvc.VerifyResetPasswordByOTP(r.Context(), req.Sign, req.Otp)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.writeJSON(w, http.StatusOK, token)
}

func (c *AccountController) FinalizeResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[finalizeResetPasswordRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := c.authSvc.FinalizeResetPasswordByOTP(r.Context(), req.Sign, req.NewPassword); err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.writeJSON(w, http.StatusOK, true)
}

func (c *AccountController) RefreshByToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest

	if apphttp.IsWebPlatform(r) {
		rtCookie, err := r.Cookie("refresh_token")
		if err != nil {
			jsonutil.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
			return
		}
		req.RefreshToken = rtCookie.Value
	} else if err := jsonutil.ReadJSON(r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	result, err := c.authSvc.RefreshAuthByToken(r.Context(), req.RefreshToken)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.respondTokens(w, r, result.AccessToken, result.RefreshToken)
}

func (c *AccountController) setAccessTokenCookie(w http.ResponseWriter, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Path:     "/",
		Value:    accessToken,
		MaxAge:   int(c.accessTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   !c.debugMode,
		SameSite: http.SameSiteStrictMode,
	})
}

func (c *AccountController) setRefreshTokenCookie(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Path:     refreshTokenCookiePath,
		Value:    refreshToken,
		MaxAge:   int(c.refreshTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   !c.debugMode,
		SameSite: http.SameSiteStrictMode,
	})
}

func SetupAccountRoutes(
	r chi.Router,
	authService *auth.AuthService,
	identityService identity.IdentityService,
	authMiddleware *middleware.AuthMiddleware,
	logger log.Logger,
	refreshTokenExpiry time.Duration,
	accessTokenExpiry time.Duration,
	debugMode bool,
) {
	accountController := &AccountController{
		authSvc:            authService,
		identitySvc:        identityService,
		logger:             logger,
		refreshTokenExpiry: refreshTokenExpiry,
		accessTokenExpiry:  accessTokenExpiry,
		debugMode:          debugMode,
	}

	r.Route("/accounts", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Route("/by-otp", func(r chi.Router) {
				r.Post("/request", accountController.RequestAuthByOTP)
				r.Post("/verify", accountController.VerifyAuthByOTP)
			})
			r.Post("/by-password", accountController.AuthenticateByPassword)
			r.Post("/refresh", accountController.RefreshByToken)
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
