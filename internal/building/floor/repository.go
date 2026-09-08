package floor

import "context"

type FloorRepository interface {
	Create(ctx context.Context, f *Floor) error
	CreateBatch(ctx context.Context, floors []*Floor) error
}