package sessionsvc

import (
	"context"
	"log"
	"time"
	"unitly/internal/domain"
	"unitly/internal/worker"
)

const (
	keepExpiredSessionsFor = 1 * 24 * time.Hour // 1 day
	keepRevokedSessionsFor = 7 * 24 * time.Hour // 7 days
	sessionCleanupTimeout  = 30 * time.Second
)

func RunSessionCleanupWorker(
	ctx context.Context,
	sessionRepo domain.SessionRepository,
	interval time.Duration,
) {
	worker.RunPeriodicWorker("session-cleanup", ctx, interval, func() error {
		execCtx, cancel := context.WithTimeout(context.Background(), sessionCleanupTimeout)
		defer cancel()

		// Pass the thresholds into the execution function
		deleted, err := sessionRepo.CleanupOldSessions(execCtx, keepExpiredSessionsFor, keepRevokedSessionsFor)

		if err == nil && deleted > 0 {
			log.Printf("[INFO] Session cleanup completed. Purged %d old rows.", deleted)
		}

		return err
	})
}
