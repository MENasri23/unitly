package building

import (
	"time"
	"unitly/internal/domain"
)

var ErrFloorNotFound = domain.NewNotFoundError("floor")
var ErrFloorNumberTaken = domain.NewError("FLOOR_NUMBER_EXISTS", "a floor with this number already exists in this building")
var ErrFloorCaptionTaken = domain.NewError("FLOOR_CAPTION_EXISTS", "a floor with this caption already exists in this building")
var ErrNoFloorsProvided = domain.NewError("FLOORS_REQUIRED", "at least one floor is required")

type Floor struct {
	ID         int64
	BuildingID int64
	// ground = 0, above ground positive, below ground negative;
	// unique per building
	FloorNumber int16
	Caption     *string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func (f *Floor) IsSoftDeleted() bool { return f.DeletedAt != nil }
