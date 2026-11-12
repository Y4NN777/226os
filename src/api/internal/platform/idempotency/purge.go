package idempotency

import (
	"context"
	"log/slog"
	"time"
)

const purgeTimeout = 30 * time.Second

// RunPurge deletes expired rows immediately, then every `every`, until ctx is
// cancelled. It returns only when ctx is done.
func RunPurge(ctx context.Context, store Store, every time.Duration, now func() time.Time, logger *slog.Logger) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for ctx.Err() == nil {
		purgeOnce(ctx, store, now, logger)
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
}

func purgeOnce(ctx context.Context, store Store, now func() time.Time, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(ctx, purgeTimeout)
	defer cancel()
	n, err := store.Purge(ctx, now())
	if err != nil {
		logger.WarnContext(ctx, "idempotency purge failed", slog.Any("error", err))
		return
	}
	logger.DebugContext(ctx, "idempotency purge", slog.Int64("deleted", n))
}
