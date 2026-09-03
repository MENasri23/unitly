package building

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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

type BuildingRepository interface {
	Create(ctx context.Context, b *Building) error
	GetByID(ctx context.Context, id uuid.UUID) (*Building, error)
	GetByUserID(ctx context.Context, userID int64) ([]Building, error)
	Update(ctx context.Context, b *Building) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
