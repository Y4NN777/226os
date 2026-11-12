package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

// timeout bounds each request with a context deadline, like chi's middleware.Timeout, but
// answers a handler that overran the deadline without writing anything through the single
// error writer (503 SERVICE_UNAVAILABLE problem+json, Retry-After: 1) instead of a bare 504.
// A handler that already started its response is left alone: a second write would corrupt it.
func timeout(logger *slog.Logger, d time.Duration) func(http.Handler) http.Handler {
	write := WriteError(logger)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			tw := &writeTracker{ResponseWriter: w}
			next.ServeHTTP(tw, r.WithContext(ctx))
			if errors.Is(ctx.Err(), context.DeadlineExceeded) && !tw.wrote {
				write(w, r, apperr.New(apperr.CodeServiceUnavailable, "The request took too long.").
					WithRetryAfter(time.Second))
			}
		})
	}
}

// writeTracker records whether the handler started a response.
type writeTracker struct {
	http.ResponseWriter
	wrote bool
}

// Unwrap lets http.NewResponseController reach SetWriteDeadline and friends.
func (t *writeTracker) Unwrap() http.ResponseWriter { return t.ResponseWriter }

func (t *writeTracker) WriteHeader(code int) {
	t.wrote = true
	t.ResponseWriter.WriteHeader(code)
}

func (t *writeTracker) Write(p []byte) (int, error) {
	t.wrote = true
	return t.ResponseWriter.Write(p)
}

// FlushError commits the headers, so it counts as writing; http.ResponseController prefers it to Unwrap.
func (t *writeTracker) FlushError() error {
	t.wrote = true
	return http.NewResponseController(t.ResponseWriter).Flush()
}
