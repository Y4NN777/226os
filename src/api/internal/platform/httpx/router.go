// Package httpx holds cross-cutting HTTP transport: router, middleware, server lifecycle.
package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const requestTimeout = 30 * time.Second

// NewRouter builds the root router. Domain modules mount their routes here later.
func NewRouter(logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(echoRequestID)
	r.Use(accessLog(logger))
	r.Use(Recover(logger))
	r.Use(middleware.Timeout(requestTimeout))
	r.Get("/healthz", healthz)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Not Found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeProblem(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method Not Allowed")
	})
	return r
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func echoRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(middleware.RequestIDHeader, middleware.GetReqID(r.Context()))
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func accessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rec, r)
			logger.LogAttrs(r.Context(), slog.LevelInfo, "request",
				slog.String("request_id", middleware.GetReqID(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// Recover turns a panic into a redacted 500 problem response and logs it.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
						slog.String("request_id", middleware.GetReqID(r.Context())),
						slog.Any("panic", rec),
					)
					writeProblem(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal Server Error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// writeProblem emits an RFC 9457 problem document (ADR-0004).
func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, title string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":       "urn:226os:problem:" + code,
		"title":      title,
		"status":     status,
		"code":       code,
		"instance":   r.URL.Path,
		"request_id": middleware.GetReqID(r.Context()),
	})
}
