package floor

import (
	"context"
	"unitly/internal/building"
)

type FloorRepository interface {
	CreateBatch(ctx context.Context, floors []*building.Floor) error
}
