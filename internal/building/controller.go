package building

import (
	"errors"
	"net/http"
	"strings"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/middleware"
	"unitly/internal/http/reqctx"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateBuildingRequest struct {
	Name    string  `json:"name" validate:"required,trimmed_name=2:100"`
	Address *string `json:"address" validate:"omitzero,trimmed_name=5:1000"`
}

type BuildingSummaryDTO struct {
	GUID uuid.UUID `json:"guid"`
	Name string    `json:"name"`
}

type BuildingController struct {
	buildings BuildingService
	authMW    func(http.Handler) http.Handler
}

func NewBuildingController(
	b BuildingService,
	authMW func(http.Handler) http.Handler,
) *BuildingController {
	return &BuildingController{buildings: b, authMW: authMW}
}

// BuildingContext loads the building by :buildingGUID and the caller's
// highest-privilege active membership, both stored on context.
func BuildingContext(buildings BuildingService) func(http.Handler) http.Handler {
	f := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			userID, ok := reqctx.GetUserInternalID(r.Context())
			if !ok {
				apphttp.UnauthorizedError(w, domain.ErrInvalidToken)
				return
			}

			buildingID, err := uuid.Parse(chi.URLParam(r, "buildingGUID"))
			if err != nil {
				apphttp.BadRequestError(w, domain.NewError("", "invalid building id"))
				return
			}

			m, err := buildings.GetHighestLevelMembership(r.Context(), userID, buildingID)
			if err != nil {
				switch {
				case errors.Is(err, domain.ErrMembershipNotFound):
					apphttp.ForbiddenError(w, domain.ErrForbidden)
				case errors.Is(err, ErrBuildingNotFound):
					apphttp.NotFoundError(w, ErrBuildingNotFound)
				default:
					apphttp.InternalServerError(w, domain.ErrInternal)

				}
				return
			}

			ctx := reqctx.WithMembership(r.Context(), m)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
	return f
}

func (c *BuildingController) Routes(r chi.Router) {
	r.Route("/buildings", func(r chi.Router) {
		r.Use(c.authMW)
		r.Post("/", c.create)
		r.Get("/", c.listMine)

		r.Route("/{buildingGUID}", func(r chi.Router) {
			r.Use(BuildingContext(c.buildings))
			r.With(middleware.RequireMinRoleLevel(domain.RoleAdmin.Level)).Delete("/", c.delete)
		})
	})
}

func (c *BuildingController) create(w http.ResponseWriter, r *http.Request) {
	userID, _ := reqctx.GetUserInternalID(r.Context())

	req, err := apphttp.BindAndValidate[CreateBuildingRequest](r)
	if err != nil {
		apphttp.BadRequestError(w, err)
		return
	}
	address := req.Address
	if address != nil {
		*address = strings.TrimSpace(*address)
	}
	in := CreateBuildingInput{
		Name:    strings.TrimSpace(req.Name),
		Address: req.Address,
	}

	b, err := c.buildings.Create(r.Context(), userID, in)
	if err != nil {
		if errors.Is(err, ErrBuildingNameTaken) {
			apphttp.ConflictError(w, err)
		} else {
			apphttp.InternalServerError(w, err)
		}
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, b.GUID)
}

func (c *BuildingController) listMine(w http.ResponseWriter, r *http.Request) {
	userID, _ := reqctx.GetUserInternalID(r.Context())
	buildings, err := c.buildings.ListMine(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrBuildingNotFound) {
			apphttp.NotFoundError(w, domain.ErrNotFound)
		} else {
			apphttp.InternalServerError(w, err)
		}
		return
	}

	BuildingDTOs := make([]BuildingSummaryDTO, len(buildings))
	for i, b := range buildings {
		BuildingDTOs[i] = BuildingSummaryDTO{
			GUID: b.ID,
			Name: b.Name,
		}
	}
	apphttp.WriteJSON(w, http.StatusOK, BuildingDTOs)
}

func (c *BuildingController) delete(w http.ResponseWriter, r *http.Request) {
	m := reqctx.GetMembership(r.Context())
	if err := c.buildings.Delete(r.Context(), m.BuildingID); err != nil {
		if errors.Is(err, ErrBuildingNotFound) {
			apphttp.NotFoundError(w, domain.ErrNotFound)
		} else {
			apphttp.InternalServerError(w, domain.ErrInternal)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
