package httpx_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

type pingerFunc func(context.Context) error

func (f pingerFunc) Ping(ctx context.Context) error { return f(ctx) }

func readyz(t *testing.T, logs *bytes.Buffer, deps httpx.Deps) *httptest.ResponseRecorder {
	t.Helper()
	deps.Server = apigen.Unimplemented{}
	h := httpx.NewRouter(slog.New(slog.NewJSONHandler(logs, nil)), deps)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil))
	return rec
}

func TestReadyzHealthyPingIs200(t *testing.T) {
	var logs bytes.Buffer
	rec := readyz(t, &logs, httpx.Deps{Ready: okPinger{}})
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" ||
		strings.TrimSpace(rec.Body.String()) != `{"status":"ready"}` {
		t.Fatalf("status = %d ct = %q body = %q", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
}

func TestReadyzFailingPingIs503WithoutErrorText(t *testing.T) {
	var logs bytes.Buffer
	const secret = "dial tcp 10.0.0.7:5432: password authentication failed for hunter2"
	rec := readyz(t, &logs, httpx.Deps{Ready: pingerFunc(func(context.Context) error { return errors.New(secret) })})
	if rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("status = %d ct = %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), "SERVICE_UNAVAILABLE") {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if containsAny(rec.Body.String(), "hunter2", "10.0.0.7", "dial tcp") {
		t.Fatalf("ping error leaked: %s", rec.Body.String())
	}
	if !strings.Contains(logs.String(), "hunter2") || !strings.Contains(logs.String(), `"request_id"`) {
		t.Fatalf("real error not logged with request id: %s", logs.String())
	}
}

func TestReadyzBlockedPingTimesOut(t *testing.T) {
	var logs bytes.Buffer
	start := time.Now()
	rec := readyz(t, &logs, httpx.Deps{
		ReadyTimeout: 20 * time.Millisecond,
		Ready:        pingerFunc(func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }),
	})
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("took %s, deadline not honoured", elapsed)
	}
}

func TestReadyzDefaultTimeoutIsTwoSeconds(t *testing.T) {
	var logs bytes.Buffer
	var got time.Duration
	readyz(t, &logs, httpx.Deps{Ready: pingerFunc(func(ctx context.Context) error {
		dl, ok := ctx.Deadline()
		if !ok {
			t.Error("no deadline on ping context")
		}
		got = time.Until(dl)
		return nil
	})})
	if got <= time.Second || got > 2*time.Second {
		t.Fatalf("deadline in %s, want about 2s", got)
	}
}
