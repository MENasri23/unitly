package floor

import (
	"errors"
	"net/http"
	"strconv"
	"unitly/internal/building"
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
		FloorNumber int16   `json:"floor_number"`
		Caption     *string `json:"caption"`
	} `json:"floors" validate:"required,min=1,dive"`
}

type floorSummary struct {
	FloorNumber int16   `json:"floor_number"`
	Caption     *string `json:"caption"`
}

func toFloorSummary(f *building.Floor) floorSummary {
	return floorSummary{FloorNumber: f.FloorNumber, Caption: f.Caption}
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

	floorSummaries := make([]floorSummary, len(floors))
	for i, f := range floors {
		floorSummaries[i] = toFloorSummary(f)
	}

	apphttp.WriteJSON(w, http.StatusCreated, floorSummaries)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, building.ErrFloorNotFound):
		apphttp.NotFoundError(w, err)
	case errors.Is(err, building.ErrNoFloorsProvided):
		apphttp.BadRequestError(w, err)
	case errors.Is(err, building.ErrFloorNumberTaken),
		errors.Is(err, building.ErrFloorCaptionTaken):
		apphttp.ConflictError(w, err)
	default:
		apphttp.InternalServerError(w, domain.ErrInternal)
	}
}
