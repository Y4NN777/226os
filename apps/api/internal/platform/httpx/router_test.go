package httpx_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/apps/api/internal/platform/httpx"
)

func newRouter(buf *bytes.Buffer) http.Handler {
	return httpx.NewRouter(slog.New(slog.NewJSONHandler(buf, nil)))
}

func TestHealthz(t *testing.T) {
	var logs bytes.Buffer
	rec := httptest.NewRecorder()
	newRouter(&logs).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["status"] != "ok" {
		t.Fatalf("body = %q err = %v", rec.Body.String(), err)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("missing X-Request-Id")
	}
	if !strings.Contains(logs.String(), `"path":"/healthz"`) || !strings.Contains(logs.String(), `"request_id"`) {
		t.Fatalf("request not logged with request_id: %s", logs.String())
	}
}

func TestHealthzRejectsUnsafeMethods(t *testing.T) {
	var logs bytes.Buffer
	rec := httptest.NewRecorder()
	newRouter(&logs).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/healthz", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUnknownRouteIsProblemJSON(t *testing.T) {
	var logs bytes.Buffer
	rec := httptest.NewRecorder()
	newRouter(&logs).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q", ct)
	}
	var p map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p["type"] != "urn:226os:problem:NOT_FOUND" || p["code"] != "NOT_FOUND" || p["status"] != float64(404) {
		t.Fatalf("problem = %v", p)
	}
	if p["request_id"] == "" || p["request_id"] == nil {
		t.Fatalf("missing request_id: %v", p)
	}
}

func TestPanicIsRedactedProblemJSON(t *testing.T) {
	var logs bytes.Buffer
	h := httpx.Recover(slog.New(slog.NewJSONHandler(&logs, nil)))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("secret-token-123")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret-token-123") {
		t.Fatalf("panic value leaked: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INTERNAL_ERROR") {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if !strings.Contains(logs.String(), "panic") {
		t.Fatalf("panic not logged: %s", logs.String())
	}
}
