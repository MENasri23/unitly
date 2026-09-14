package building

import (
	"context"
	"time"
	"unitly/internal/domain"
)

var ErrUnitNotFound = domain.NewNotFoundError("unit")
var ErrUnitNumberTaken = domain.NewError("UNIT_NUMBER_EXISTS", "a unit with this number already exists on this floor")
var ErrUnitNumbersDuplicate = domain.NewError("UNIT_NUMBERS_DUPLICATE", "unit numbers must be unique")
var ErrNoUnitsProvided = domain.NewError("UNITS_REQUIRED", "at least one unit is required")

type UnitStatus string

const (
	UnitStatusVacant   UnitStatus = "vacant"
	UnitStatusOccupied UnitStatus = "occupied"
)

type Unit struct {
	ID         int64
	BuildingID int64
	FloorID    int64
	UnitNumber string
	Caption    *string
	Status     UnitStatus
	HasParking bool
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

func (u *Unit) IsSoftDeleted() bool {
	return u.DeletedAt != nil
}

type UnitRepository interface {
	CreateBatch(ctx context.Context, units []*Unit) error

	GetByNumber(ctx context.Context, floorID int64, unitNumber string) (*Unit, error)

	SoftDelete(ctx context.Context, id int64, expectedVersion int) error
}
