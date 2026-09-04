package building

import (
	"context"

	"github.com/google/uuid"
)

type CreateBuildingInput struct {
	Name    string
	Address *string
}

type BuildingService interface {
	Create(ctx context.Context, actorUserID int64, in CreateBuildingInput) (*Building, error)
	Get(ctx context.Context, buildingID int64) (*Building, error)
	GetInternalID(ctx context.Context, buildingID uuid.UUID) (int64, error)
	ListMine(ctx context.Context, userID int64) ([]*Building, error)
	Delete(ctx context.Context, buildingID uuid.UUID) error
}
