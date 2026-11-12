package httpx_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.yaml.in/yaml/v3"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

const specPath = "../../../../../api/openapi.yaml"

type problemBody struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Status     int    `json:"status"`
	Detail     string `json:"detail"`
	Instance   string `json:"instance"`
	Code       string `json:"code"`
	RequestID  string `json:"request_id"`
	Violations []struct {
		Field string `json:"field"`
		Code  string `json:"code"`
	} `json:"violations"`
}

func TestStatusTableMatchesSpecCatalogueExactly(t *testing.T) {
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Components struct {
			Schemas struct {
				ErrorCode struct {
					Enum      []string       `yaml:"enum"`
					Catalogue map[string]int `yaml:"x-error-catalogue"`
				} `yaml:"ErrorCode"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	ec := s.Components.Schemas.ErrorCode
	table, titles := httpx.StatusTable(), httpx.TitleTable()
	if len(ec.Enum) == 0 || len(table) != len(ec.Enum) || len(titles) != len(ec.Enum) || len(ec.Catalogue) != len(ec.Enum) {
		t.Fatalf("sizes: enum=%d catalogue=%d status=%d titles=%d", len(ec.Enum), len(ec.Catalogue), len(table), len(titles))
	}
	for _, name := range ec.Enum {
		code := apperr.Code(name)
		if table[code] != ec.Catalogue[name] || table[code] == 0 {
			t.Errorf("%s: status %d, spec %d", name, table[code], ec.Catalogue[name])
		}
		if titles[code] == "" {
			t.Errorf("%s: no title", name)
		}
	}
}

func serveError(t *testing.T, err error, mw ...func(http.Handler) http.Handler) (*httptest.ResponseRecorder, problemBody, string) {
	t.Helper()
	var logs bytes.Buffer
	h := httpx.WriteError(slog.New(slog.NewJSONHandler(&logs, nil)))
	var next http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { h(w, r, err) })
	next = middleware.RequestID(next)
	rec := httptest.NewRecorder()
	next.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/questions/42", nil))
	var body problemBody
	if e := json.Unmarshal(rec.Body.Bytes(), &body); e != nil {
		t.Fatalf("body %q: %v", rec.Body.String(), e)
	}
	return rec, body, logs.String()
}

func TestWriteErrorEveryCode(t *testing.T) {
	for _, code := range apperr.Codes() {
		t.Run(string(code), func(t *testing.T) {
			rec, body, _ := serveError(t, apperr.New(code, "safe detail"))
			want := httpx.StatusTable()[code]
			if rec.Code != want || body.Status != want {
				t.Fatalf("status = %d / %d, want %d", rec.Code, body.Status, want)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
				t.Fatalf("content-type = %q", ct)
			}
			if body.Type != "urn:226os:problem:"+string(code) || body.Code != string(code) {
				t.Fatalf("type/code = %q / %q", body.Type, body.Code)
			}
			if body.Title == "" || body.Instance != "/api/v1/questions/42" || body.RequestID == "" {
				t.Fatalf("incomplete problem: %+v", body)
			}
			if want < 500 && body.Detail != "safe detail" {
				t.Fatalf("detail = %q", body.Detail)
			}
		})
	}
}

func TestWriteErrorViolationsAndRetryAfter(t *testing.T) {
	e := apperr.New(apperr.CodeRateLimited, "slow down").
		WithViolations(apperr.Violation{Field: "limit", Code: "out_of_range"}).
		WithRetryAfter(1500 * time.Millisecond)
	rec, body, _ := serveError(t, fmt.Errorf("service: %w", e))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d (wrapped *apperr.Error must be found with errors.As)", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "2" {
		t.Fatalf("Retry-After = %q, want 2 (rounded up)", got)
	}
	if len(body.Violations) != 1 || body.Violations[0].Field != "limit" || body.Violations[0].Code != "out_of_range" {
		t.Fatalf("violations = %+v", body.Violations)
	}
}

func TestWriteErrorOmitsEmptyViolationsAndRetryAfter(t *testing.T) {
	rec, _, _ := serveError(t, apperr.New(apperr.CodeNotFound, ""))
	if rec.Header().Get("Retry-After") != "" {
		t.Fatal("unexpected Retry-After")
	}
	if strings.Contains(rec.Body.String(), "violations") || strings.Contains(rec.Body.String(), "detail") {
		t.Fatalf("empty members must be omitted: %s", rec.Body.String())
	}
}

func TestWriteErrorUnknownErrorIsGeneric500(t *testing.T) {
	secret := errors.New(`pq: connection to host db.internal:5432 failed: password "hunter2" for user app`)
	rec, body, logs := serveError(t, secret)
	if rec.Code != 500 || body.Code != "INTERNAL_ERROR" {
		t.Fatalf("status/code = %d / %s", rec.Code, body.Code)
	}
	for _, leak := range []string{"pq:", "db.internal", "hunter2", "password", "goroutine"} {
		if strings.Contains(rec.Body.String(), leak) {
			t.Fatalf("response leaks %q: %s", leak, rec.Body.String())
		}
	}
	if !strings.Contains(logs, "db.internal") || !strings.Contains(logs, `"request_id":"`+body.RequestID+`"`) {
		t.Fatalf("real error not logged with request_id: %s", logs)
	}
}

func TestWriteErrorInternalAppErrorHidesDetailAndCause(t *testing.T) {
	cause := errors.New("sql: duplicate key at 10.0.0.7")
	rec, body, logs := serveError(t, apperr.Wrap(apperr.CodeInternal, cause))
	if rec.Code != 500 || strings.Contains(rec.Body.String(), "10.0.0.7") {
		t.Fatalf("leak or wrong status: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(logs, "10.0.0.7") || body.RequestID == "" {
		t.Fatalf("cause not logged: %s", logs)
	}
	// A careless 5xx detail from domain code must not reach the client either.
	rec, _, _ = serveError(t, apperr.New(apperr.CodeServiceUnavailable, "pg at 10.0.0.7 down"))
	if rec.Code != 503 || strings.Contains(rec.Body.String(), "10.0.0.7") {
		t.Fatalf("5xx detail leaked: %d %s", rec.Code, rec.Body.String())
	}
}

func TestWriteErrorUndefinedCodeIsInternal(t *testing.T) {
	rec, body, logs := serveError(t, apperr.New(apperr.Code("NOT_IN_SPEC"), "x"))
	if rec.Code != 500 || body.Code != "INTERNAL_ERROR" || !strings.Contains(logs, "NOT_IN_SPEC") {
		t.Fatalf("got %d %s logs=%s", rec.Code, body.Code, logs)
	}
}
