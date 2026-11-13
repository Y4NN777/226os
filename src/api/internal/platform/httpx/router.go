// Package httpx holds cross-cutting HTTP transport: router, middleware, server lifecycle.
package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apispec"
	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const requestTimeout = 30 * time.Second

// Deps are the collaborators the router needs; app wires them.
type Deps struct {
	Server       apigen.ServerInterface
	Ready        Pinger
	ReadyTimeout time.Duration // zero means 2s
	// Subjects resolves who is calling. Nil means everyone is anonymous; S1 provides the
	// session-backed resolver (ADR-0003). The subject always comes from the resolver.
	Subjects subject.Resolver
	// Idempotency stores Idempotency-Key records (ADR-0004 point 6). Nil disables the
	// middleware (pass-through); app always wires the PostgreSQL store.
	Idempotency idempotency.Store
	// CSRF enforces the ADR-0003 synchronizer-token and Origin checks on unsafe
	// methods. Nil installs a pass-through so the middleware chain stays stable.
	CSRF func(http.Handler) http.Handler
}

// NewRouter builds the root router and mounts the generated API under /api/v1.
func NewRouter(logger *slog.Logger, deps Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(echoRequestID)
	r.Use(accessLog(logger))
	r.Use(Recover(logger))
	r.Use(timeout(logger, requestTimeout))
	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz(logger, deps.Ready, deps.ReadyTimeout))
	apigen.HandlerWithOptions(deps.Server, apigen.ChiServerOptions{
		BaseURL:          "/api/v1",
		BaseRouter:       r,
		ErrorHandlerFunc: APIErrorHandler(logger),
		// apigen wraps the handler with each element in turn, so the LAST element is the OUTERMOST and
		// runs first: schema validation before subject resolution, which becomes a database lookup
		// (ADR-0004 point 8: validate before authorization-sensitive work). CSRF runs after subject
		// resolution and before idempotency, so a CSRF failure never claims an idempotency key;
		// idempotency runs last: it needs the resolved subject and a valid request.
		Middlewares: []apigen.MiddlewareFunc{
			idempotency.Middleware(deps.Idempotency, logger, WriteError(logger)),
			csrfOrPass(deps.CSRF),
			subject.Middleware(deps.Subjects, WriteError(logger)),
			embeddedValidator(logger),
		},
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeProblem(w, r, http.StatusNotFound, "NOT_FOUND", "Not Found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeProblem(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method Not Allowed")
	})
	return r
}

// embeddedValidator builds the contract validator from the spec embedded in the
// binary. A failure means the shipped api/openapi.yaml is unusable, a build-time
// invariant (apispec's tests and the router tests load it), so refusing to
// construct the router beats serving /api/v1 unvalidated.
func embeddedValidator(logger *slog.Logger) apigen.MiddlewareFunc {
	validate, err := NewValidator(logger, apispec.YAML)
	if err != nil {
		panic("httpx: embedded OpenAPI contract is unusable: " + err.Error())
	}
	return validate
}

// csrfOrPass returns the CSRF middleware or a pass-through when the app did
// not wire one, keeping the middleware chain deterministic for the B1 gate.
func csrfOrPass(mw func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	if mw == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return mw
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func echoRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(middleware.RequestIDHeader, requestID(r.Context()))
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

// Unwrap lets http.NewResponseController reach Flush, SetWriteDeadline and friends.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

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
				slog.String("request_id", requestID(r.Context())),
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
						slog.String("request_id", requestID(r.Context())),
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
		"instance":   capRunes(r.URL.Path, maxInstanceLen),
		"request_id": requestID(r.Context()),
	})
}
