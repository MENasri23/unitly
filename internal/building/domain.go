package building

import (
	"time"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

var ErrBuildingNotFound = domain.NewNotFoundError("building")

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
