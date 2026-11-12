package httpx_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

const testDeadline = 20 * time.Millisecond

func serveTimeout(t *testing.T, h http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	rec := httptest.NewRecorder()
	chain := middleware.RequestID(httpx.Timeout(logger, testDeadline)(h))
	chain.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/slow", nil))
	return rec
}

func TestTimeoutAnswersProblemWhenNothingWasWritten(t *testing.T) {
	rec := serveTimeout(t, func(_ http.ResponseWriter, r *http.Request) { <-r.Context().Done() })
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 (body %q)", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q", ct)
	}
	if ra := rec.Header().Get("Retry-After"); ra != "1" {
		t.Fatalf("Retry-After = %q, want 1", ra)
	}
	var body problemBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q: %v", rec.Body.String(), err)
	}
	if body.Code != "SERVICE_UNAVAILABLE" || body.Status != http.StatusServiceUnavailable || body.RequestID == "" {
		t.Fatalf("body = %+v", body)
	}
}

func TestTimeoutDoesNotOverwriteAPartialResponse(t *testing.T) {
	rec := serveTimeout(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("partial"))
		<-r.Context().Done()
	})
	if rec.Code != http.StatusAccepted || rec.Body.String() != "partial" {
		t.Fatalf("got %d %q, want the handler's own response", rec.Code, rec.Body.String())
	}
}

func TestTimeoutKeepsAWriteOnlyResponse(t *testing.T) {
	rec := serveTimeout(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("x"))
		<-r.Context().Done()
	})
	if rec.Code != http.StatusOK || rec.Body.String() != "x" {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestTimeoutKeepsAFlushedResponse(t *testing.T) {
	rec := serveTimeout(t, func(w http.ResponseWriter, r *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("Flush: %v", err)
		}
		<-r.Context().Done()
	})
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") == "application/problem+json" {
		t.Fatalf("a flushed response was overwritten: %d %v", rec.Code, rec.Header())
	}
}

func TestTimeoutLeavesAFastHandlerUntouched(t *testing.T) {
	rec := serveTimeout(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	if rec.Code != http.StatusCreated || rec.Body.Len() != 0 {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestTimeoutDeadlineReachesTheHandlerContext(t *testing.T) {
	var dl time.Time
	var ok bool
	before := time.Now()
	serveTimeout(t, func(_ http.ResponseWriter, r *http.Request) {
		dl, ok = r.Context().Deadline()
	})
	if !ok || dl.Before(before.Add(testDeadline)) || dl.After(time.Now().Add(testDeadline)) {
		t.Fatalf("deadline = %v, ok = %v", dl, ok)
	}
}

func TestTimeoutIgnoresAHandlerFailingForAnotherReason(t *testing.T) {
	// The parent context is cancelled (client gone), not timed out: no problem is written.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(t.Context())
	rec := httptest.NewRecorder()
	h := httpx.Timeout(logger, time.Hour)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { cancel() }))
	h.ServeHTTP(rec, httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
}
