package unit

import (
	"errors"
	"net/http"
	"unitly/internal/building"
	"unitly/internal/building/floor"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/middleware"
	"unitly/internal/http/router"

	"github.com/go-chi/chi/v5"
)

type UnitController struct {
	buildings building.BuildingService
}

func NewUnitController(b building.BuildingService) router.RouteRegistrar {
	return &UnitController{buildings: b}
}

func (c *UnitController) Routes(r chi.Router) {
	adminOnly := middleware.RequireMinRoleLevel(domain.RoleAdmin.Level)

	r.Route("/units", func(r chi.Router) {
		r.With(adminOnly).Post("/", c.Add)
	})
}

type addUnitsRequest struct {
	UnitNumbers []string `json:"unit_numbers" validate:"required,gt=0,lt=20,dive,required,trimmed_name=1:20"`
}

type UnitSummary struct {
	UnitNumber string  `json:"unit_number"`
	Caption    *string `json:"caption"`
}

func toUnitSummary(u building.Unit) UnitSummary {
	return UnitSummary{UnitNumber: u.UnitNumber, Caption: u.Caption}
}

func (c *UnitController) Add(w http.ResponseWriter, r *http.Request) {
	f, _ := floor.FromContext(r.Context())

	req, err := apphttp.BindAndValidate[addUnitsRequest](r)
	if err != nil {
		apphttp.BadRequestError(w, err)
		return
	}

	Units, err := c.buildings.AddUnits(r.Context(), f.ID, req.UnitNumbers)
	if err != nil {
		writeErr(w, err)
		return
	}

	UnitSummaries := make([]UnitSummary, len(Units))
	for i, u := range Units {
		UnitSummaries[i] = toUnitSummary(u)
	}

	apphttp.WriteJSON(w, http.StatusCreated, UnitSummaries)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, building.ErrUnitNotFound):
		apphttp.NotFoundError(w, err)
	case errors.Is(err, building.ErrNoUnitsProvided):
		apphttp.BadRequestError(w, err)
	case errors.Is(err, building.ErrUnitNumberTaken):
		apphttp.ConflictError(w, err)
	case errors.Is(err, building.ErrUnitNumbersDuplicate):
		apphttp.BadRequestError(w, err)
	default:
		apphttp.InternalServerError(w, domain.ErrInternal)
	}
}
