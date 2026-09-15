package building

import (
	"github.com/google/uuid"
)

type UnitSummary struct {
	UnitNumber string
	Caption    *string
	Status     UnitStatus
}

type FloorDetail struct {
	FloorNumber int16
	Caption     *string
	Units       []UnitSummary
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
