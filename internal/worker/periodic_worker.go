package worker

import (
	"context"
	"log"
	"time"
)

func RunPeriodicWorker(name string, ctx context.Context, interval time.Duration, task func() error) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		log.Printf("[INFO] %s worker initialized. Run Interval: %v", name, interval)

		for {
			select {
			case <-ctx.Done():
				log.Printf("[INFO] Shutting down %s worker gracefully...", name)
				return

			case <-ticker.C:
				if err := task(); err != nil {
					log.Printf("[ERROR] %s worker failed: %v", name, err)
				}
			}
		}
	}()
}
