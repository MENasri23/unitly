package unit

import (
	"errors"
	"net/http"
	"unitly/internal/building"
	bldgdto "unitly/internal/building/dto"
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

		r.Route("/{unitNumber}", func(r chi.Router) {
			r.With(adminOnly).Delete("/", c.remove)
		})
	})
}

type addUnitsRequest struct {
	UnitNumbers []string `json:"unit_numbers" validate:"required,gt=0,lt=21,dive,required,trimmed_name=1:20"`
}

func parseUnitNumberParam(r *http.Request) (string, error) {
	unitNumber := chi.URLParam(r, "unitNumber")
	if len(unitNumber) > 20 {
		return "", errors.New("invalid unitNumber")
	}
	return unitNumber, nil
}

func (c *UnitController) Add(w http.ResponseWriter, r *http.Request) {
	f, _ := floor.FromContext(r.Context())

	req, err := apphttp.BindAndValidate[addUnitsRequest](r)
	if err != nil {
		apphttp.BadRequestError(w, err)
		return
	}

	units, err := c.buildings.AddUnits(r.Context(), f.ID, req.UnitNumbers)
	if err != nil {
		writeErr(w, err)
		return
	}

	unitDTos := bldgdto.MapUnitsToSummaryDTOs(units)
	apphttp.WriteJSON(w, http.StatusCreated, unitDTos)
}

func (c *UnitController) remove(w http.ResponseWriter, r *http.Request) {
	c.doRemove(w, r, false)
}

func (c *UnitController) doRemove(w http.ResponseWriter, r *http.Request, force bool) {
	f, _ := floor.FromContext(r.Context())

	unitNumber, err := parseUnitNumberParam(r)
	if err != nil {
		apphttp.BadRequestError(w, err)
		return
	}

	if err := c.buildings.RemoveUnit(r.Context(), f.ID, unitNumber, force); err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, building.ErrUnitNotFound):
		apphttp.NotFoundError(w, err)
	case errors.Is(err, building.ErrNoUnitsProvided),
		errors.Is(err, building.ErrUnitNumbersDuplicate):
		apphttp.BadRequestError(w, err)
	case errors.Is(err, building.ErrUnitNumberTaken),
		errors.Is(err, domain.ErrVersionConflict):
		apphttp.ConflictError(w, err)
	default:
		apphttp.InternalServerError(w, domain.ErrInternal)
	}
}
