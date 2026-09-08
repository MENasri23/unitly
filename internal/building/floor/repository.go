package floor

import "context"

type FloorRepository interface {
	CreateBatch(ctx context.Context, floors []*Floor) error
}