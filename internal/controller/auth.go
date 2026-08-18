package controller

import (
	"net/http"
	"time"
	"unitly/internal/controller/middleware"
	"unitly/internal/controller/reqctx"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/header"
	"unitly/internal/log"
	"unitly/internal/service/auth"
	"unitly/internal/service/identity"
	"unitly/pkg/jsonutil"

	"github.com/go-chi/chi/v5"
)

const authCookiePath = "/api/v1/auth/refresh"

type AuthController struct {
	authSvc        *auth.AuthService
	identitySvc    identity.IdentityService
	authMiddleware *middleware.AuthMiddleware
	logger         log.Logger
	cfg            AuthConfig
}

type AuthConfig struct {
	RefreshTokenExpiry time.Duration
	AccessTokenExpiry  time.Duration
	DebugMode          bool
}

func NewAuthController(
	authSvc *auth.AuthService,
	identitySvc identity.IdentityService,
	authMW *middleware.AuthMiddleware,
	logger log.Logger,
	cfg AuthConfig,
) *AuthController {
	return &AuthController{authSvc: authSvc, identitySvc: identitySvc, authMiddleware: authMW, logger: logger, cfg: cfg}
}

func (c *AuthController) Routes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Route("/by-otp", func(r chi.Router) {
			r.Post("/request", c.RequestAuthByOTP)
			r.Post("/verify", c.VerifyAuthByOTP)
		})
		r.Route("/reset-password/by-otp", func(r chi.Router) {
			r.Post("/request", c.RequestResetPasswordByOTP)
			r.Post("/verify", c.VerifyResetPasswordByOTP)
			r.Post("/finalize", c.FinalizeResetPasswordByOTP)
		})
		r.Post("/by-password", c.AuthenticateByPassword)
		r.Post("/refresh", c.RefreshByToken)
		r.With(c.authMiddleware.RequireAuth).Post("/logout", c.Logout)
	})
}

type requestOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
}

type verifyOTPRequest struct {
	Sign string `json:"sign" validate:"required,max=256"`
	Otp  string `json:"otp" validate:"required,number,len=6"`
}

type authByPasswordRequest struct {
	Phone    string `json:"phone" validate:"required,e164"`
	Password string `json:"password" validate:"required"`
}

type finalizeResetPasswordRequest struct {
	Sign        string `json:"sign" validate:"required,jwt,max=512"`
	NewPassword string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,max=256"`
}

type TokensDTO struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (c *AuthController) writeJSON(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, v, c.logger)
}

func (c *AuthController) buildCookie(name, value, path string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Path:     path,
		Value:    value,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   !c.cfg.DebugMode,
		SameSite: http.SameSiteStrictMode,
	}
}

func (c *AuthController) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, c.buildCookie("access_token", accessToken, "/", int(c.cfg.AccessTokenExpiry.Seconds())))
	http.SetCookie(w, c.buildCookie("refresh_token", refreshToken, authCookiePath, int(c.cfg.RefreshTokenExpiry.Seconds())))
}

func (c *AuthController) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, c.buildCookie("access_token", "", "/", -1))
	http.SetCookie(w, c.buildCookie("refresh_token", "", authCookiePath, -1))
}

// respondTokens consolidates the "web client gets cookies, native client
// gets a JSON body" branch that was repeated in four handlers.
func (c *AuthController) respondTokens(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	if apphttp.IsWebPlatform(r) {
		c.setAuthCookies(w, accessToken, refreshToken)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	c.writeJSON(w, http.StatusOK, TokensDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// refreshTokenFromWeb reads the refresh token from the cookie for web
// clients, mirroring how RefreshByToken already sources it. Used by both
// RefreshByToken and Logout since they need the same input.
func refreshTokenFromRequest(r *http.Request) (refreshRequest, error) {
	var req refreshRequest

	if apphttp.IsWebPlatform(r) {
		rtCookie, err := r.Cookie("refresh_token")
		if err != nil {
			return req, domain.ErrInvalidToken
		}
		req.RefreshToken = rtCookie.Value
		return req, nil
	}

	if err := jsonutil.ReadJSON(r, &req); err != nil {
		return req, err
	}
	return req, nil
}

func getClientMeta(r *http.Request) domain.ClientMetadata {
	return domain.ClientMetadata{
		IPAddress: reqctx.GetClientIP(r.Context()),
		UserAgent: header.GetUserAgent(r),
	}
}

func (c *AuthController) RequestAuthByOTP(w http.ResponseWriter, r *http.Request) {
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

func (c *AuthController) VerifyAuthByOTP(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[verifyOTPRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	meta := getClientMeta(r)
	result, err := c.authSvc.VerifyAuthByOTP(r.Context(), req.Sign, req.Otp, meta)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.respondTokens(w, r, result.AccessToken, result.RefreshToken)
}

func (c *AuthController) AuthenticateByPassword(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[authByPasswordRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	meta := getClientMeta(r)
	result, err := c.authSvc.AuthenticateByPassword(r.Context(), req.Phone, req.Password, meta)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.respondTokens(w, r, result.AccessToken, result.RefreshToken)
}

func (c *AuthController) RequestResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
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

func (c *AuthController) VerifyResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
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

func (c *AuthController) FinalizeResetPasswordByOTP(w http.ResponseWriter, r *http.Request) {
	req, err := bindAndValidate[finalizeResetPasswordRequest](r)
	if err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := c.authSvc.FinalizeResetPasswordByOTP(r.Context(), req.Sign, req.NewPassword); err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.clearAuthCookies(w)
	c.writeJSON(w, http.StatusOK, true)
}

func (c *AuthController) RefreshByToken(w http.ResponseWriter, r *http.Request) {
	req, err := refreshTokenFromRequest(r)
	if err != nil {
		status := http.StatusBadRequest
		if err == domain.ErrInvalidToken {
			status = http.StatusUnauthorized
		}
		jsonutil.WriteErrJSON(w, status, err)
		return
	}

	if err := validateStruct(req); err != nil {
		jsonutil.WriteErrJSON(w, http.StatusBadRequest, err)
		return
	}

	meta := getClientMeta(r)
	result, err := c.authSvc.RefreshAuthByToken(r.Context(), req.RefreshToken, meta)
	if err != nil {
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
		return
	}

	c.respondTokens(w, r, result.AccessToken, result.RefreshToken)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	user := reqctx.GetUser(r.Context())
	userID := user.GUID
	currentSessionID, _ := reqctx.GetSessionID(r.Context())

	if err := c.authSvc.Logout(r.Context(), userID, currentSessionID); err != nil {
		c.logger.Error("failed to logout user", "userID", userID, "sessionID", currentSessionID, "error", err)
		jsonutil.WriteErrJSON(w, mapErrToStatusCode(err), err)
	}

	if apphttp.IsWebPlatform(r) {
		c.clearAuthCookies(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
