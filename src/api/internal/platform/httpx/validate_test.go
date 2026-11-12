package httpx_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apispec"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

const validQuestion = `{"title":"How do I parse JSON?","body_md":"Some body","lang":"en","tags":["go"]}`

// bodyEcho is a server whose createQuestion handler proves the body survived
// validation: it answers 201 with the bytes it read.
type bodyEcho struct{ apigen.Unimplemented }

func (bodyEcho) CreateQuestion(w http.ResponseWriter, r *http.Request, _ apigen.CreateQuestionParams) {
	b, _ := io.ReadAll(r.Body)
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(b)
}

func serve(t *testing.T, srv apigen.ServerInterface, method, target, contentType, body string) *httptest.ResponseRecorder {
	t.Helper()
	var logs bytes.Buffer
	h := httpx.NewRouter(slog.New(slog.NewJSONHandler(&logs, nil)), httpx.Deps{Server: srv, Ready: okPinger{}})
	req := httptest.NewRequestWithContext(t.Context(), method, target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) problemBody {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q, body %q", ct, rec.Body.String())
	}
	var p problemBody
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("body %q: %v", rec.Body.String(), err)
	}
	if p.RequestID == "" || p.Status != rec.Code {
		t.Fatalf("problem = %+v, status %d", p, rec.Code)
	}
	return p
}

// withJSON replaces one top-level field of validQuestion.
func withJSON(t *testing.T, field string, value any) string {
	t.Helper()
	m := map[string]any{}
	if err := json.Unmarshal([]byte(validQuestion), &m); err != nil {
		t.Fatal(err)
	}
	m[field] = value
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func repeat(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}

func TestValidatorBodyViolations(t *testing.T) {
	longKey := strings.Repeat("k", 300)
	tests := []struct {
		name      string
		body      string
		wantField string
		wantCode  string
	}{
		{"unknown field", withJSON(t, "nope", 1), "/nope", "unknown_field"},
		{"unknown field name truncated", withJSON(t, longKey, 1), "/" + longKey[:127], "unknown_field"},
		{"unknown field pointer escaping", withJSON(t, "a/b~c", 1), "/a~1b~0c", "unknown_field"},
		{"title too long", withJSON(t, "title", strings.Repeat("é", 201)), "/title", "too_long"},
		{"title too short", withJSON(t, "title", "abcd"), "/title", "too_short"},
		{"too many tags", withJSON(t, "tags", []string{"a", "b", "c", "d", "e", "f"}), "/tags", "too_many_items"},
		{"no tags", withJSON(t, "tags", []string{}), "/tags", "too_short"},
		{"duplicate tags", withJSON(t, "tags", []string{"go", "go"}), "/tags", "duplicate"},
		{"too many links", withJSON(t, "links", repeat("x", 11)), "/links", "too_many_items"},
		{"bad enum", withJSON(t, "lang", "de"), "/lang", "invalid_value"},
		{"bad pattern", withJSON(t, "country", "bf"), "/country", "invalid_format"},
		{"wrong type", withJSON(t, "title", 12345), "/title", "invalid_value"},
		{"bad tag item", withJSON(t, "tags", []string{"Bad Tag"}), "/tags/0", "invalid_format"},
		{"missing required", `{"title":"How do I parse JSON?","body_md":"x","lang":"en"}`, "/tags", "required"},
		{"body not an object", `"hello"`, "/", "invalid_value"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(t, apigen.Unimplemented{}, http.MethodPost, "/api/v1/questions", "application/json", tc.body)
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
			}
			p := decodeProblem(t, rec)
			if p.Code != "VALIDATION_ERROR" {
				t.Fatalf("code = %q", p.Code)
			}
			found := false
			for _, v := range p.Violations {
				if v.Field == tc.wantField && v.Code == tc.wantCode {
					found = true
				}
			}
			if !found {
				t.Fatalf("violations = %+v, want {%s %s}", p.Violations, tc.wantField, tc.wantCode)
			}
			if strings.Contains(rec.Body.String(), "12345") || strings.Contains(rec.Body.String(), "Bad Tag") {
				t.Fatalf("submitted value echoed: %s", rec.Body.String())
			}
		})
	}
}

func TestValidatorCapsViolations(t *testing.T) {
	rec := serve(t, apigen.Unimplemented{}, http.MethodPost, "/api/v1/questions", "application/json",
		withJSON(t, "tags", repeat("BAD", 40)))
	p := decodeProblem(t, rec)
	if rec.Code != http.StatusUnprocessableEntity || len(p.Violations) != 20 {
		t.Fatalf("status %d, %d violations", rec.Code, len(p.Violations))
	}
}

func TestValidatorMalformedBody(t *testing.T) {
	tests := map[string]string{"bad json": `{"title":`, "empty body": ""}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, apigen.Unimplemented{}, http.MethodPost, "/api/v1/questions", "application/json", body)
			p := decodeProblem(t, rec)
			if rec.Code != http.StatusBadRequest || p.Code != "MALFORMED_REQUEST" {
				t.Fatalf("status %d code %q", rec.Code, p.Code)
			}
		})
	}
}

func TestValidatorContentType(t *testing.T) {
	for _, ct := range []string{"text/plain", "", "application/json; charset=latin1"} {
		t.Run("ct="+ct, func(t *testing.T) {
			rec := serve(t, apigen.Unimplemented{}, http.MethodPost, "/api/v1/questions", ct, validQuestion)
			p := decodeProblem(t, rec)
			if rec.Code != http.StatusUnprocessableEntity || p.Code != "VALIDATION_ERROR" ||
				len(p.Violations) != 1 || p.Violations[0].Field != "Content-Type" || p.Violations[0].Code != "invalid_value" {
				t.Fatalf("status %d problem %+v", rec.Code, p)
			}
		})
	}
}

func TestValidatorBodyTooLarge(t *testing.T) {
	big := withJSON(t, "body_md", strings.Repeat("a", 1<<20))
	rec := serve(t, apigen.Unimplemented{}, http.MethodPost, "/api/v1/questions", "application/json", big)
	p := decodeProblem(t, rec)
	if rec.Code != http.StatusRequestEntityTooLarge || p.Code != "PAYLOAD_TOO_LARGE" {
		t.Fatalf("status %d code %q", rec.Code, p.Code)
	}
}

func TestValidatorValidRequestReachesHandlerWithBodyIntact(t *testing.T) {
	rec := serve(t, apigen.Unimplemented{}, http.MethodPost, "/api/v1/questions", "application/json", validQuestion)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("stub status = %d, body %s", rec.Code, rec.Body.String())
	}
	rec = serve(t, bodyEcho{}, http.MethodPost, "/api/v1/questions", "application/json; charset=utf-8", validQuestion)
	if rec.Code != http.StatusCreated || rec.Body.String() != validQuestion {
		t.Fatalf("status %d, handler saw %q", rec.Code, rec.Body.String())
	}
}

func TestValidatorQueryParameters(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wantCode  int
		wantProb  string
		wantField string
		wantViol  string
	}{
		{"limit above max", "/api/v1/questions?limit=51", 422, "VALIDATION_ERROR", "limit", "out_of_range"},
		{"limit below min", "/api/v1/questions?limit=0", 422, "VALIDATION_ERROR", "limit", "out_of_range"},
		{"limit not a number", "/api/v1/questions?limit=abc", 400, "MALFORMED_REQUEST", "limit", "invalid_format"},
		{"empty q", "/api/v1/questions?q=", 422, "VALIDATION_ERROR", "q", "too_short"},
		{"empty limit", "/api/v1/questions?limit=", 400, "MALFORMED_REQUEST", "limit", "invalid_format"},
		{"bad uuid path", "/api/v1/questions/not-a-uuid", 400, "MALFORMED_REQUEST", "questionId", "invalid_format"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(t, apigen.Unimplemented{}, http.MethodGet, tc.target, "", "")
			p := decodeProblem(t, rec)
			if rec.Code != tc.wantCode || p.Code != tc.wantProb {
				t.Fatalf("status %d code %q, body %s", rec.Code, p.Code, rec.Body.String())
			}
			if len(p.Violations) != 1 || p.Violations[0].Field != tc.wantField || p.Violations[0].Code != tc.wantViol {
				t.Fatalf("violations = %+v", p.Violations)
			}
			// The echo check reads the fields a client could influence, not the raw body: request_id comes from
			// a sequential counter whose digits can contain "51" or "abc" by chance (order-dependent flake).
			fields := p.Title + "|" + p.Detail + "|" + p.Instance
			for _, v := range p.Violations {
				fields += "|" + v.Field + "|" + v.Code
			}
			if strings.Contains(fields, "abc") || strings.Contains(fields, "51") && tc.wantField == "limit" {
				t.Fatalf("submitted value echoed: %s", fields)
			}
		})
	}
}

func TestValidatorKeepsMissingIfMatchAs428(t *testing.T) {
	rec := serve(t, apigen.Unimplemented{}, http.MethodPatch,
		"/api/v1/questions/0198f6c0-0000-7000-8000-000000000001", "application/json", `{}`)
	p := decodeProblem(t, rec)
	if rec.Code != http.StatusPreconditionRequired || p.Code != "PRECONDITION_REQUIRED" {
		t.Fatalf("status %d code %q", rec.Code, p.Code)
	}
}

func TestValidatorSkipsProbes(t *testing.T) {
	for _, path := range []string{"/healthz", "/readyz"} {
		rec := serve(t, apigen.Unimplemented{}, http.MethodGet, path+"?limit=abc", "", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, rec.Code)
		}
	}
}

func TestNewValidatorErrors(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	if _, err := httpx.NewValidator(logger, []byte("not: [valid")); err == nil {
		t.Fatal("unparseable spec accepted")
	}
	if _, err := httpx.NewValidator(logger, []byte("openapi: 3.1.0\ninfo: {}\npaths: {}\n")); err == nil {
		t.Fatal("invalid document accepted")
	}
}

func TestValidatorUnknownRouteFailsClosed(t *testing.T) {
	var logs bytes.Buffer
	mw, err := httpx.NewValidator(slog.New(slog.NewJSONHandler(&logs, nil)), apispec.YAML)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	h := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/nope", nil))
	if rec.Code != http.StatusInternalServerError || called {
		t.Fatalf("status %d, handler called %v", rec.Code, called)
	}
}

func TestNewRouterPanicsOnInvalidEmbeddedSpec(t *testing.T) {
	orig := apispec.YAML
	apispec.YAML = []byte("not: [valid")
	t.Cleanup(func() { apispec.YAML = orig })
	defer func() {
		if recover() == nil {
			t.Fatal("NewRouter accepted an invalid embedded spec")
		}
	}()
	newRouter(&bytes.Buffer{})
}

func TestValidatorFormats(t *testing.T) {
	tests := []struct {
		name, target, body, wantField string
	}{
		{"email", "/api/v1/auth/sign-up", `{"email":"nobody","password":"correct horse battery","username":"alice"}`, "/email"},
		{"uuid", "/api/v1/questions", withJSON(t, "links", []map[string]string{{"type": "question", "id": "not-a-uuid"}}), "/links/0/id"},
		{"uri without scheme", "/api/v1/projects", `{"github_url":"github.com/a/b","lang":"en"}`, "/github_url"},
		{"uri unparseable", "/api/v1/projects", `{"github_url":"http://[::1","lang":"en"}`, "/github_url"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(t, apigen.Unimplemented{}, http.MethodPost, tc.target, "application/json", tc.body)
			p := decodeProblem(t, rec)
			found := false
			for _, v := range p.Violations {
				found = found || v.Field == tc.wantField && v.Code == "invalid_format"
			}
			if rec.Code != http.StatusUnprocessableEntity || !found {
				t.Fatalf("status %d violations %+v", rec.Code, p.Violations)
			}
		})
	}
	ok := serve(t, apigen.Unimplemented{}, http.MethodPost, "/api/v1/projects", "application/json",
		`{"github_url":"https://github.com/a/b","lang":"en"}`)
	if ok.Code != http.StatusNotImplemented {
		t.Fatalf("valid uri refused: %d %s", ok.Code, ok.Body.String())
	}
}

// 3.1 constructs of the contract (anyOf with type null) must keep working on
// the engine the validator selects.
func TestValidatorAcceptsNullInMergePatch(t *testing.T) {
	rec := serve(t, apigen.Unimplemented{}, http.MethodPatch, "/api/v1/me/profile", "application/json",
		`{"country":null,"intent":null}`)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	rec = serve(t, apigen.Unimplemented{}, http.MethodPatch, "/api/v1/me/profile", "application/json", `{"country":"bf"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid country accepted: %d", rec.Code)
	}
}

// The validator must stand on its own when used without apigen's binding.
func TestValidatorParameterFaultsWithoutBinding(t *testing.T) {
	mw, err := httpx.NewValidator(slog.New(slog.NewJSONHandler(io.Discard, nil)), apispec.YAML)
	if err != nil {
		t.Fatal(err)
	}
	h := middleware.RequestID(mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("handler reached") })))
	tests := []struct{ method, target, wantField, wantCode string }{
		{http.MethodGet, "/api/v1/questions?limit=abc", "limit", "invalid_format"},
		{http.MethodPatch, "/api/v1/questions/0198f6c0-0000-7000-8000-000000000001", "If-Match", "required"},
	}
	for _, tc := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.target, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(rec, req)
		p := decodeProblem(t, rec)
		if rec.Code != http.StatusBadRequest || len(p.Violations) == 0 ||
			p.Violations[0].Field != tc.wantField || p.Violations[0].Code != tc.wantCode {
			t.Fatalf("%s: status %d problem %+v", tc.target, rec.Code, p)
		}
	}
}
