package floor

import (
	"errors"
	"net/http"
	"strconv"
	"unitly/internal/building"
	bldgdto "unitly/internal/building/dto"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/middleware"
	"unitly/internal/http/reqctx"
	"unitly/internal/http/router"

	"github.com/go-chi/chi/v5"
)

type floorController struct {
	buildings  building.BuildingService
	unitRouter router.RouteRegistrar
}

func NewFloorController(b building.BuildingService, uRouter router.RouteRegistrar) router.RouteRegistrar {
	return &floorController{buildings: b, unitRouter: uRouter}
}

func (c *floorController) Routes(r chi.Router) {
	adminOnly := middleware.RequireMinRoleLevel(domain.RoleAdmin.Level)

	r.Route("/floors", func(r chi.Router) {
		r.With(adminOnly).Post("/", c.Add)

		r.Route("/{floorNumber}", func(r chi.Router) {
			r.Use(c.floorContext)
			r.With(adminOnly).Get("/", c.detail)
			r.With(adminOnly).Delete("/", c.delete)
			c.unitRouter.Routes(r)
		})
	})
}

func (c *floorController) floorContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := reqctx.GetMembership(r.Context())

		num, err := parseFloorNumber(r)
		if err != nil {
			apphttp.BadRequestError(w, domain.NewError("", "invalid floor number"))
			return
		}

		f, err := c.buildings.GetFloor(r.Context(), m.BuildingID, num)
		if err != nil {
			writeErr(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(withFloor(r.Context(), f)))
	})
}

func parseFloorNumber(r *http.Request) (int16, error) {
	n, err := strconv.ParseInt(chi.URLParam(r, "floorNumber"), 10, 16)
	return int16(n), err
}

type addFloorsRequest struct {
	Floors []struct {
		FloorNumber int16   `json:"floor_number" validate:"required"`
		Caption     *string `json:"caption" validate:"omitzero,trimmed_name=1:100"`
	} `json:"floors" validate:"required,min=1,max=20"`
}

func (c *floorController) Add(w http.ResponseWriter, r *http.Request) {
	req, err := apphttp.BindAndValidate[addFloorsRequest](r)
	if err != nil {
		apphttp.BadRequestError(w, err)
		return
	}

	inputs := make([]building.NewFloor, len(req.Floors))
	for i, f := range req.Floors {
		inputs[i] = building.NewFloor{FloorNumber: f.FloorNumber, Caption: f.Caption}
	}

	m := reqctx.GetMembership(r.Context())
	floors, err := c.buildings.AddFloors(r.Context(), m.BuildingID, inputs)
	if err != nil {
		writeErr(w, err)
		return
	}

	floorDTOs := bldgdto.MapFloorsToDTOs(floors)
	apphttp.WriteJSON(w, http.StatusCreated, floorDTOs)
}

func (c *floorController) detail(w http.ResponseWriter, r *http.Request) {
	m := reqctx.GetMembership(r.Context())
	f, _ := FromContext(r.Context())

	detail, err := c.buildings.GetFloorDetail(r.Context(), m.BuildingID, f.ID)
	if err != nil {
		writeErr(w, err)
		return
	}

	detailDTO := bldgdto.MapFloorDetailToDTO(detail)
	apphttp.WriteJSON(w, http.StatusOK, detailDTO)
}

func (c *floorController) delete(w http.ResponseWriter, r *http.Request) {
	c.doRemove(w, r, false)
}

func (c *floorController) doRemove(w http.ResponseWriter, r *http.Request, force bool) {
	f, _ := FromContext(r.Context())

	if err := c.buildings.RemoveFloor(r.Context(), f.ID, f.Version, force); err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, building.ErrFloorNotFound):
		apphttp.NotFoundError(w, err)
	case errors.Is(err, building.ErrNoFloorsProvided),
		errors.Is(err, building.ErrFloorNumberDuplicate),
		errors.Is(err, building.ErrFloorCaptionDuplicate):
		apphttp.BadRequestError(w, err)
	case errors.Is(err, building.ErrFloorNumberTaken),
		errors.Is(err, building.ErrFloorCaptionTaken),
		errors.Is(err, building.ErrFloorHasUnits),
		errors.Is(err, domain.ErrVersionConflict):
		apphttp.ConflictError(w, err)
	default:
		apphttp.InternalServerError(w, domain.ErrInternal)
	}
}
