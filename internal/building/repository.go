package building

import (
	"context"

	"github.com/google/uuid"
)

type BuildingRepository interface {
	Create(ctx context.Context, b *Building) error
	GetByID(ctx context.Context, id int64) (*Building, error)
	GetByGUID(ctx context.Context, guid uuid.UUID) (*Building, error)
	ListByUser(ctx context.Context, userID int64) ([]*Building, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
