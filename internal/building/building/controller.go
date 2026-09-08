package bldgimpl

import (
	"errors"
	"net/http"
	"strings"
	"unitly/internal/building"
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
	buildings building.BuildingService
	authMW    func(http.Handler) http.Handler
}

func NewBuildingController(
	b building.BuildingService,
	authMW func(http.Handler) http.Handler,
) *BuildingController {
	return &BuildingController{buildings: b, authMW: authMW}
}

// BuildingContext loads the building by :buildingGUID and the caller's
// highest-privilege active membership, both stored on context.
func BuildingContext(buildings building.BuildingService) func(http.Handler) http.Handler {
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
				case errors.Is(err, building.ErrBuildingNotFound):
					apphttp.NotFoundError(w, building.ErrBuildingNotFound)
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
	floors := building.NewFloorController(c.buildings)

	r.Route("/buildings", func(r chi.Router) {
		r.Use(c.authMW)
		r.Post("/", c.create)
		r.Get("/", c.listMine)

		r.Route("/{buildingGUID}", func(r chi.Router) {
			r.Use(BuildingContext(c.buildings))
			tenant := r.With(middleware.RequireMinRoleLevel(domain.RoleTenant.Level))
			tenant.Get("/", c.detail)

			admin := r.With(middleware.RequireMinRoleLevel(domain.RoleAdmin.Level))
			admin.Delete("/", c.delete)

			floors.Routes(r)
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
	in := building.CreateBuildingInput{
		Name:    strings.TrimSpace(req.Name),
		Address: req.Address,
	}

	b, err := c.buildings.Create(r.Context(), userID, in)
	if err != nil {
		writeBuildingErr(w, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, b.GUID)
}

func (c *BuildingController) listMine(w http.ResponseWriter, r *http.Request) {
	userID, _ := reqctx.GetUserInternalID(r.Context())
	buildings, err := c.buildings.ListMine(r.Context(), userID)
	if err != nil {
		writeBuildingErr(w, err)
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

type buildingDetailDTO struct {
	ID            uuid.UUID        `json:"id"`
	Name          string           `json:"name"`
	TotalUnits    int              `json:"total_units"`
	TotalOccupied int              `json:"total_occupied"`
	Floors        []floorDetailDTO `json:"floors"`
}

type floorDetailDTO struct {
	FloorNumber int16   `json:"floor_number"`
	Caption     *string `json:"caption"`
}

func toBuildingDetailDTO(b *building.BuildingDetail) buildingDetailDTO {
	return buildingDetailDTO{
		ID:            b.GUID,
		Name:          b.Name,
		TotalUnits:    b.TotalUnits,
		TotalOccupied: b.OccupiedUnits,
		Floors:        toFloorDetailDTOs(b.Floors),
	}
}

func toFloorDetailDTOs(floors []building.FloorDetail) []floorDetailDTO {
	if len(floors) == 0 {
		return nil
	}

	result := make([]floorDetailDTO, len(floors))
	for i, f := range floors {
		result[i] = floorDetailDTO{
			FloorNumber: f.FloorNumber,
			Caption:     f.Caption,
		}
	}
	return result
}

func (c *BuildingController) detail(w http.ResponseWriter, r *http.Request) {
	m := reqctx.GetMembership(r.Context())
	detail, err := c.buildings.GetDetail(r.Context(), m.BuildingID)
	if err != nil {
		writeBuildingErr(w, err)
		return
	}

	detailDTO := toBuildingDetailDTO(detail)
	apphttp.WriteJSON(w, http.StatusOK, detailDTO)
}

func (c *BuildingController) delete(w http.ResponseWriter, r *http.Request) {
	m := reqctx.GetMembership(r.Context())
	if err := c.buildings.Delete(r.Context(), m.BuildingID); err != nil {
		writeBuildingErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeBuildingErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, building.ErrBuildingNotFound):
		apphttp.NotFoundError(w, err)
	case errors.Is(err, building.ErrBuildingNameTaken):
		apphttp.ConflictError(w, err)
	default:
		apphttp.InternalServerError(w, domain.ErrInternal)
	}
}
