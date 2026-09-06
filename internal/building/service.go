package building

import (
	"context"
	"unitly/internal/domain"

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
	GetHighestLevelMembership(ctx context.Context, userID int64, buildingID uuid.UUID) (*domain.Membership, error)
	ListMine(ctx context.Context, userID int64) ([]*BuildingSummary, error)
	Delete(ctx context.Context, buildingID int64) error
}
