package building

import (
	"time"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

var ErrBuildingNotFound = domain.NewNotFoundError("building")
var ErrBuildingNameTaken = domain.NewError("BUILDING_EXIST", "a building with this name already exists")

type BuildingSummary struct {
	ID   uuid.UUID
	Name string
}

type Building struct {
	ID        int64
	GUID      uuid.UUID
	Name      string
	Address   *string
	CreatedBy int64
	Version   int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (b *Building) IsSoftDeleted() bool {
	return b.DeletedAt != nil
}

type FloorDetail struct {
	FloorNumber int16
	Caption     *string
}

type BuildingDetail struct {
	GUID          uuid.UUID
	Name          string
	TotalUnits    int
	OccupiedUnits int
	Floors        []FloorDetail
}

type ResidentSummary struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
}