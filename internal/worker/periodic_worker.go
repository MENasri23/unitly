package worker

import (
	"context"
	"log"
	"time"
)

type periodicWorker struct {
	name   string
	ticker *time.Ticker
	task   func(context.Context) error
}

func (w *periodicWorker) Do(ctx context.Context) {
	defer w.ticker.Stop()
	w.task(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Shutting down %s worker gracefully...", w.name)
			return

		case <-w.ticker.C:
			if err := w.task(ctx); err != nil {
				log.Printf("[ERROR] %s worker failed: %v", w.name, err)
			}
		}
	}
}

func RunPeriodicWorker(name string, ctx context.Context, interval time.Duration, task func(context.Context) error) Worker {
	w := &periodicWorker{
		name:   name,
		ticker: time.NewTicker(interval),
		task:   task,
	}
	go w.Do(ctx)
	return w
}
