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

func NewAccountController(
	authSvc *auth.AuthService,
	identitySvc identity.IdentityService,
	logger log.Logger,
	refreshTokenExpiry time.Duration,
	accessTokenExpiry time.Duration,
	debugMode bool,
) *AccountController {
	return &AccountController{
		authSvc:            authSvc,
		identitySvc:        identitySvc,
		logger:             logger,
		refreshTokenExpiry: refreshTokenExpiry,
		accessTokenExpiry:  accessTokenExpiry,
		debugMode:          debugMode,
	}
}

func (c *AccountController) RequestAuthByOTP(w http.ResponseWriter, r *http.Request) {
	var req requestOTPRequest
	if err := jsonutil.ReadJSON(r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, domain.ErrInvalidPhone)
		return
	}

	token, err := c.authSvc.RequestAuthByOTP(r.Context(), req.Phone)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, token); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, domain.NewError(domain.ErrorCode("WRITE_RESPONSE_FAILED"), "failed to write response"))
	}
}

func (c *AccountController) VerifyAuthByOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := jsonutil.ReadJSON(r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	result, err := c.authSvc.VerifyAuthByOTP(r.Context(), req.Sign, req.Otp)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	if apphttp.IsWebPlatform(r) {
		c.setAccessTokenCookie(w, result.AccessToken)
		c.setRefreshTokenCookie(w, result.RefreshToken)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	tokens := TokensDTO{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}
	if err = jsonutil.WriteJSON(w, http.StatusOK, tokens); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, domain.NewError(domain.ErrorCode("WRITE_RESPONSE_FAILED"), "failed to write response"))
	}
}

func (c *AccountController) AuthenticateByPassword(w http.ResponseWriter, r *http.Request) {
	var req authByPasswordRequest
	if err := jsonutil.ReadJSON(r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	result, err := c.authSvc.AuthenticateByPassword(r.Context(), req.Phone, req.Password)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	if apphttp.IsWebPlatform(r) {
		c.setAccessTokenCookie(w, result.AccessToken)
		c.setRefreshTokenCookie(w, result.RefreshToken)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	tokens := TokensDTO{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, tokens); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, domain.NewError(domain.ErrorCode("WRITE_RESPONSE_FAILED"), "failed to write response"))
	}
}

func (c *AccountController) RequestResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	var req requestOTPRequest
	if err := jsonutil.ReadJSON(r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	token, err := c.authSvc.RequestResetPasswordByOTP(r.Context(), req.Phone)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, token); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, domain.NewError(domain.ErrorCode("WRITE_RESPONSE_FAILED"), "failed to write response"))
	}
}

func (c *AccountController) VerifyResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := jsonutil.ReadJSON(r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	token, err := c.authSvc.VerifyResetPasswordByOTP(r.Context(), req.Sign, req.Otp)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, token); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, domain.NewError(domain.ErrorCode("WRITE_RESPONSE_FAILED"), "failed to write response"))
	}
}

func (c *AccountController) FinalizeResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	var req finalizeResetPasswordRequest
	if err := jsonutil.ReadJSON(r, &req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := c.authSvc.FinalizeResetPasswordByOTP(r.Context(), req.Sign, req.NewPassword); err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	if err := jsonutil.WriteJSON(w, http.StatusOK, true); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, domain.NewError(domain.ErrorCode("WRITE_RESPONSE_FAILED"), "failed to write response"))
	}
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
	} else {
		if err := jsonutil.ReadJSON(r, &req); err != nil {
			jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
			return
		}
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

	if apphttp.IsWebPlatform(r) {
		c.setAccessTokenCookie(w, result.AccessToken)
		c.setRefreshTokenCookie(w, result.RefreshToken)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	tokens := TokensDTO{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}

	if err = jsonutil.WriteJSON(w, http.StatusOK, tokens); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, domain.NewError(domain.ErrorCode("WRITE_RESPONSE_FAILED"), "failed to write response"))
	}
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
		Path:     "/api/v1/accounts/auth/refresh",
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
	accountController := NewAccountController(
		authService,
		identityService,
		logger,
		refreshTokenExpiry,
		accessTokenExpiry,
		debugMode,
	)

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
