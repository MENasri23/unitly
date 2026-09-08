package detail

import (
	"github.com/google/uuid"
)

type UserSummary struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
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
