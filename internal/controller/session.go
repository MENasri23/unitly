package controller

import (
	"errors"
	"net/http"
	"time"
	"unitly/internal/domain"
	"unitly/internal/http/middleware"
	"unitly/internal/http/reqctx"
	"unitly/internal/log"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	apphttp "unitly/internal/http"
)

type SessionController struct {
	sessionService domain.SessionService
	authMiddleware *middleware.AuthMiddleware
	logger         log.Logger
}

func NewSessionController(
	sessionService domain.SessionService,
	authMiddleware *middleware.AuthMiddleware,
	logger log.Logger,
) *SessionController {
	return &SessionController{sessionService: sessionService, authMiddleware: authMiddleware, logger: logger}
}

// Routes mounts the session endpoints. Wire this under whatever router
// group already applies your access-token auth middleware.
func (c *SessionController) Routes(r chi.Router) {
	r.Route("/me/sessions", func(r chi.Router) {
		r.Use(c.authMiddleware.RequireAuth)
		r.Get("/", c.List)
		r.Delete("/{sessionID}", c.Revoke)
	})
}

// sessionDTO is the public shape of a session. Deliberately excludes
// RefreshTokenHash and Version — those never leave the server.
type sessionDTO struct {
	ID          uuid.UUID         `json:"id"`
	IsCurrent   bool              `json:"is_current"`
	IPAddress   string            `json:"ip_address"`
	DeviceType  domain.DeviceType `json:"device_type"`
	OSName      string            `json:"os_name,omitempty"`
	OSVersion   string            `json:"os_version,omitempty"`
	ClientLabel string            `json:"client_label,omitempty"` // device model ("iPhone 15 Pro"), "PWA", or browser name ("Chrome") — see domain.ResolveDevice
	CreatedAt   time.Time         `json:"created_at"`
	LastUsedAt  time.Time         `json:"last_used_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
}

func toSessionDTO(s *domain.Session, currentSessionID uuid.UUID) sessionDTO {
	return sessionDTO{
		ID:          s.ID,
		IsCurrent:   s.ID == currentSessionID,
		IPAddress:   s.IPAddress,
		DeviceType:  s.DeviceType,
		OSName:      s.OSName,
		OSVersion:   s.OSVersion,
		ClientLabel: s.Label,
		CreatedAt:   s.CreatedAt,
		LastUsedAt:  s.LastUsedAt,
		ExpiresAt:   s.ExpiresAt,
	}
}

// listSessionsDTO is a page of sessions plus pagination metadata for
// building UI controls ("page 2 of 3", "47 devices total", etc.).
type listSessionsDTO struct {
	Sessions   []sessionDTO `json:"sessions"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalCount int          `json:"total_count"`
	TotalPages int          `json:"total_pages"`
}

// List handles GET /me/sessions?page=1&page_size=20 — a page of the
// caller's active sessions/devices, most recently used first, with the
// session behind the current request flagged via is_current.
//
// page defaults to 1; page_size defaults to domain.DefaultSessionPageSize
// and is capped at domain.MaxSessionPageSize. Invalid/out-of-range values
// fall back to the defaults rather than erroring — the service clamps
// again regardless, this is just so a bad query param doesn't 400 a page
// load.
func (c *SessionController) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := reqctx.GetUserID(r.Context())
	currentSessionID, _ := reqctx.GetSessionID(r.Context())

	page := apphttp.QueryInt(r, "page", 1)
	pageSize := apphttp.QueryInt(r, "page_size", domain.DefaultSessionPageSize)

	result, err := c.sessionService.ListSessions(r.Context(), userID, page, pageSize)
	if err != nil {
		c.logError("failed to list sessions", userID, currentSessionID, err)
		apphttp.InternalServerError(w, errors.New("failed to list sessions"))
		return
	}

	resp := listSessionsDTO{
		Sessions:   make([]sessionDTO, 0, len(result.Data)),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages(),
	}
	for _, s := range result.Data {
		resp.Sessions = append(resp.Sessions, toSessionDTO(s, currentSessionID))
	}

	writeJSON(w, http.StatusOK, resp, c.logger)
}

// Revoke handles DELETE /me/sessions/{sessionID}.
func (c *SessionController) Revoke(w http.ResponseWriter, r *http.Request) {
	userID, _ := reqctx.GetUserID(r.Context())
	currentSessionID, _ := reqctx.GetSessionID(r.Context())

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	if currentSessionID == sessionID {
		apphttp.WriteErrJSON(w, mapErrToStatusCode(domain.ErrSessionInUse), domain.ErrSessionInUse)
		return
	}

	if err = c.sessionService.RevokeSessionByID(r.Context(), userID, sessionID); err != nil {
		c.logError("failed to revoke session by id", userID, sessionID, err)
		apphttp.WriteErrJSON(w, mapErrToStatusCode(err), err)
	}
}

func (c *SessionController) logError(msg string, userID, sessionID uuid.UUID, err error) {
	c.logger.Error(msg, "userID", userID, "sessionID", sessionID, "error", err)
}
