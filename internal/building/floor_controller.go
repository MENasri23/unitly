package building

import (
	"errors"
	"net/http"
	"unitly/internal/building/floor"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/reqctx"
	"unitly/internal/http/router"

	"github.com/go-chi/chi/v5"
)

type floorController struct {
	buildings BuildingService
}

func NewFloorController(b BuildingService) router.RouteRegistrar {
	return &floorController{buildings: b}
}

func (c *floorController) Routes(r chi.Router) {
	r.Route("/floors", func(r chi.Router) {
		r.Post("/", c.Add)
	})
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

func toFloorSummary(f *floor.Floor) *floorSummary {
	return &floorSummary{FloorNumber: f.FloorNumber, Caption: f.Caption}
}

func (c *floorController) Add(w http.ResponseWriter, r *http.Request) {
	req, err := apphttp.BindAndValidate[addFloorsRequest](r)
	if err != nil {
		apphttp.BadRequestError(w, err)
		return
	}

	inputs := make([]NewFloor, len(req.Floors))
	for i, f := range req.Floors {
		inputs[i] = NewFloor{FloorNumber: f.FloorNumber, Caption: f.Caption}
	}

	m := reqctx.GetMembership(r.Context())
	floors, err := c.buildings.AddFloors(r.Context(), m.BuildingID, inputs)
	if err != nil {
		writeErr(w, err)
		return
	}

	floorSummaries := make([]*floorSummary, len(floors))
	for i, f := range floors {
		floorSummaries[i] = toFloorSummary(f)
	}

	apphttp.WriteJSON(w, http.StatusCreated, floorSummaries)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, floor.ErrFloorNotFound):
		apphttp.NotFoundError(w, err)
	case errors.Is(err, floor.ErrNoFloorsProvided):
		apphttp.BadRequestError(w, err)
	case errors.Is(err, floor.ErrFloorNumberTaken),
		errors.Is(err, floor.ErrFloorCaptionTaken):
		apphttp.ConflictError(w, err)
	default:
		apphttp.InternalServerError(w, domain.ErrInternal)
	}
}
