package building

import (
	"context"
	"unitly/internal/building/floor"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

type CreateBuildingInput struct {
	Name    string
	Address *string
}

type NewFloor struct {
	FloorNumber int16
	Caption     *string
}

type BuildingService interface {
	Create(ctx context.Context, actorUserID int64, in CreateBuildingInput) (*Building, error)
	Get(ctx context.Context, buildingID int64) (*Building, error)
	GetHighestLevelMembership(ctx context.Context, userID int64, buildingID uuid.UUID) (*domain.Membership, error)
	ListMine(ctx context.Context, userID int64) ([]*BuildingSummary, error)
	Delete(ctx context.Context, buildingID int64) error

	AddFloor(ctx context.Context, buildingID int64, floorNumber int16, caption *string) (*floor.Floor, error)
	AddFloors(ctx context.Context, buildingID int64, floors []NewFloor) ([]*floor.Floor, error)
}
