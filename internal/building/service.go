package building

import (
	"context"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

type CreateBuildingInput struct {
	Name    string
	Address *string
	FloorsGroundAndTop   *int16
	FloorsUnderGround    *int16
	DefaultUnitsPerFloor *int16
}

type NewFloor struct {
	FloorNumber int16
	Caption     *string
}

type BuildingService interface {
	Create(ctx context.Context, actorUserID int64, in CreateBuildingInput) (*Building, error)
	Get(ctx context.Context, buildingID int64) (*Building, error)
	GetDetail(ctx context.Context, buildingID int64) (*BuildingDetail, error)
	ListMine(ctx context.Context, userID int64) ([]*BuildingSummary, error)
	Delete(ctx context.Context, buildingID int64) error

	GetHighestLevelMembership(ctx context.Context, userID int64, buildingID uuid.UUID) (*domain.Membership, error)

	AddFloors(ctx context.Context, buildingID int64, floors []NewFloor) ([]Floor, error)
	GetFloor(ctx context.Context, buildingID int64, floorNumber int16) (*Floor, error)
	GetFloorDetail(ctx context.Context, buildingID, floorID int64) (*FloorDetail, error)
	RemoveFloor(ctx context.Context, floorID int64, expectedVersion int, force bool) error

	AddUnits(ctx context.Context, floorID int64, unitNumbers []string) ([]Unit, error)
	RemoveUnit(ctx context.Context, floorID int64, unitNumber string, force bool) error
}
