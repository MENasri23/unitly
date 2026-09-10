package floor

import (
	"context"
	"unitly/internal/building"
)

type ctxKey int

const floorKey ctxKey = 0

func withFloor(ctx context.Context, f *building.Floor) context.Context {
	return context.WithValue(ctx, floorKey, f)
}

func FromContext(ctx context.Context) (*building.Floor, bool) {
	f, ok := ctx.Value(floorKey).(*building.Floor)
	return f, ok
}
