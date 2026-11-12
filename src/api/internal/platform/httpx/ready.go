package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

const defaultReadyTimeout = 2 * time.Second

// Pinger is what readiness needs from a dependency (a pgx pool satisfies it).
type Pinger interface {
	Ping(ctx context.Context) error
}

// readyz answers 200 when the dependency answers within timeout, otherwise a
// 503 problem. The ping error is logged by WriteError (with the request id)
// and never serialised.
func readyz(logger *slog.Logger, ready Pinger, timeout time.Duration) http.HandlerFunc {
	if timeout <= 0 {
		timeout = defaultReadyTimeout
	}
	fail := WriteError(logger)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		if err := ready.Ping(ctx); err != nil {
			fail(w, r, apperr.Wrap(apperr.CodeServiceUnavailable, err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}
