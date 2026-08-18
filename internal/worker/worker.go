package worker

import "context"

type Worker interface {
	Do(ctx context.Context)
}