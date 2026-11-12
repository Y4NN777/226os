package httpx_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

type countingResolver struct{ calls int }

func (c *countingResolver) Resolve(*http.Request) (subject.Subject, error) {
	c.calls++
	return subject.Subject{}, errors.New("session store down")
}

func discardRouter(deps httpx.Deps) http.Handler {
	return httpx.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), deps)
}

// ADR-0004 point 8: requests are schema-validated BEFORE authorization-sensitive work. Subject
// resolution (a database lookup once S1 lands) must not run for a request that is invalid anyway.
func TestValidationRunsBeforeSubjectResolution(t *testing.T) {
	res := &countingResolver{}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/questions", strings.NewReader(`{"nope":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	discardRouter(httpx.Deps{Server: apigen.Unimplemented{}, Ready: okPinger{}, Subjects: res}).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity || res.calls != 0 {
		t.Fatalf("status = %d, resolver calls = %d; want 422 and 0", rec.Code, res.calls)
	}
}

// A proxy-form request target (RFC 9112 absolute-form) must be served like the origin form, not 500.
func TestAbsoluteFormTargetIsServed(t *testing.T) {
	probe := &subjectProbe{}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://anything.example/api/v1/auth/session", nil)
	rec := httptest.NewRecorder()
	discardRouter(httpx.Deps{Server: probe, Ready: okPinger{}}).ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || !probe.called {
		t.Fatalf("status = %d, handler called = %v; want 204 and true", rec.Code, probe.called)
	}
}

// The body the handler reads is the body that was validated: a second JSON value after the first
// would otherwise reach a handler that decodes only the first.
func TestValidatorRejectsTrailingJSONValue(t *testing.T) {
	rec := serve(t, bodyEcho{}, http.MethodPost, "/api/v1/questions", "application/json", validQuestion+validQuestion)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body %s", rec.Code, rec.Body.String())
	}
	if p := decodeProblem(t, rec); p.Code != "MALFORMED_REQUEST" {
		t.Fatalf("code = %q", p.Code)
	}
}

// Problem.instance (maxLength 512) and Problem.request_id (maxLength 64) must hold whatever the client sends.
func TestProblemEchoesAreBounded(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/"+strings.Repeat("a", 3000), nil)
	req.Header.Set("X-Request-Id", strings.Repeat("r", 5000))
	rec := httptest.NewRecorder()
	discardRouter(httpx.Deps{Server: apigen.Unimplemented{}, Ready: okPinger{}}).ServeHTTP(rec, req)
	var p struct {
		Instance  string `json:"instance"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Instance) > 512 || len(p.RequestID) > 64 || len(rec.Header().Get("X-Request-Id")) > 64 {
		t.Fatalf("instance %d, request_id %d, header %d bytes; want <=512, <=64, <=64",
			len(p.Instance), len(p.RequestID), len(rec.Header().Get("X-Request-Id")))
	}
}

// ADR-0004 point 5: a malformed cursor is 400 INVALID_CURSOR, not a generic 422.
func TestMalformedCursorParameterIsInvalidCursor(t *testing.T) {
	rec := serve(t, apigen.Unimplemented{}, http.MethodGet, "/api/v1/questions?cursor=%00", "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body %s", rec.Code, rec.Body.String())
	}
	if p := decodeProblem(t, rec); p.Code != "INVALID_CURSOR" {
		t.Fatalf("code = %q, want INVALID_CURSOR", p.Code)
	}
}
