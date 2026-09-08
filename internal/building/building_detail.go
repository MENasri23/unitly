package building

import (
	"context"

	"github.com/google/uuid"
)

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

type BuildingDetailRepository interface {
	GetBuildingDetail(ctx context.Context, buildingID int64) (*BuildingDetail, error)
}
