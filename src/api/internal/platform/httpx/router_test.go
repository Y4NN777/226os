package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

type okPinger struct{}

func (okPinger) Ping(context.Context) error { return nil }

func newRouter(buf *bytes.Buffer) http.Handler {
	return httpx.NewRouter(slog.New(slog.NewJSONHandler(buf, nil)),
		httpx.Deps{Server: apigen.Unimplemented{}, Ready: okPinger{}})
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

// TestAuthzHealthzNeedsNoSession is the authz evidence for the one public route: liveness answers
// the same with no credentials and with garbage credentials, exposes nothing session-specific, and
// no other path is served (404).
func TestAuthzHealthzNeedsNoSession(t *testing.T) {
	var logs bytes.Buffer
	h := newRouter(&logs)
	for name, setup := range map[string]func(*http.Request){
		"anonymous":      func(*http.Request) {},
		"garbage cookie": func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "session", Value: "not-a-session"}) },
		"garbage bearer": func(r *http.Request) { r.Header.Set("Authorization", "Bearer not-a-token") },
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
			setup(req)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
				t.Fatalf("status = %d body = %q", rec.Code, rec.Body.String())
			}
			if rec.Header().Get("Set-Cookie") != "" {
				t.Fatalf("liveness must not set a cookie: %q", rec.Header().Get("Set-Cookie"))
			}
		})
	}
	t.Run("nothing else is served", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz/admin", nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

// --- mounted generated API (ADR-0004) ---

type specDoc struct {
	Components struct {
		Parameters map[string]map[string]any `yaml:"parameters"`
		Schemas    map[string]map[string]any `yaml:"schemas"`
	} `yaml:"components"`
	Paths map[string]map[string]any `yaml:"paths"`
}

type specOp struct {
	method, path, id string
	params           []map[string]any
	hasBody          bool
}

func (s specDoc) resolve(v any) []map[string]any {
	list, _ := v.([]any)
	out := make([]map[string]any, 0, len(list))
	for _, e := range list {
		p := e.(map[string]any)
		if ref, ok := p["$ref"].(string); ok {
			p = s.Components.Parameters[strings.TrimPrefix(ref, "#/components/parameters/")]
		}
		out = append(out, p)
	}
	return out
}

func (s specDoc) sample(schema map[string]any) string {
	for {
		ref, ok := schema["$ref"].(string)
		if !ok {
			break
		}
		schema = s.Components.Schemas[strings.TrimPrefix(ref, "#/components/schemas/")]
	}
	if enum, ok := schema["enum"].([]any); ok {
		return fmt.Sprint(enum[0])
	}
	if schema["format"] == "uuid" {
		return "0190c3f2-7a3b-7c1e-8a5d-3b1f2c4d5e6f"
	}
	return "alice"
}

func loadOps(t *testing.T) (specDoc, []specOp) {
	t.Helper()
	raw, err := os.ReadFile("../../../../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var s specDoc
	if err := yaml.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	methods := map[string]bool{"get": true, "put": true, "post": true, "delete": true, "patch": true}
	var ops []specOp
	for path, item := range s.Paths {
		shared := s.resolve(item["parameters"])
		for m, v := range item {
			if !methods[m] {
				continue
			}
			body := v.(map[string]any)
			ops = append(ops, specOp{strings.ToUpper(m), path, body["operationId"].(string),
				append(append([]map[string]any{}, shared...), s.resolve(body["parameters"])...),
				body["requestBody"] != nil})
		}
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].path+ops[i].method < ops[j].path+ops[j].method })
	return s, ops
}

func TestEveryOperationIsMountedUnderAPIv1(t *testing.T) {
	s, ops := loadOps(t)
	if len(ops) != 47 {
		t.Fatalf("spec has %d operations, want 47", len(ops))
	}
	var logs bytes.Buffer
	h := newRouter(&logs)
	for _, o := range ops {
		t.Run(o.id, func(t *testing.T) {
			path := "/api/v1" + o.path
			header := http.Header{}
			query := url.Values{}
			for _, p := range o.params {
				schema, _ := p["schema"].(map[string]any)
				req, _ := p["required"].(bool)
				switch p["in"] {
				case "path":
					path = strings.ReplaceAll(path, "{"+p["name"].(string)+"}", s.sample(schema))
				case "query":
					if req {
						query.Set(p["name"].(string), s.sample(schema))
					}
				case "header":
					if req {
						header.Set(p["name"].(string), s.sample(schema))
					}
				}
			}
			r := httptest.NewRequestWithContext(t.Context(), o.method, path+"?"+query.Encode(), nil)
			r.Header = header
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)
			// The contract validator runs before the handler (ADR-0004 point 8): an
			// operation whose body is required answers 400 to this bodyless probe,
			// which proves it is mounted behind validation; the others reach the stub.
			want := http.StatusNotImplemented
			if o.hasBody {
				want = http.StatusBadRequest
			}
			if rec.Code != want {
				t.Fatalf("%s %s = %d, want %d", o.method, path, rec.Code, want)
			}
		})
	}
}

func problemOf(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content-type = %q body = %s", ct, rec.Body.String())
	}
	var p map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMountedAPIParameterErrorsAreProblemJSON(t *testing.T) {
	var logs bytes.Buffer
	h := newRouter(&logs)
	t.Run("missing If-Match is 428", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPatch,
			"/api/v1/questions/0190c3f2-7a3b-7c1e-8a5d-3b1f2c4d5e6f", nil))
		if p := problemOf(t, rec); rec.Code != http.StatusPreconditionRequired || p["code"] != "PRECONDITION_REQUIRED" {
			t.Fatalf("status = %d problem = %v", rec.Code, p)
		}
	})
	t.Run("malformed uuid is 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/questions/not-a-uuid", nil))
		if p := problemOf(t, rec); rec.Code != http.StatusBadRequest || p["code"] != "MALFORMED_REQUEST" {
			t.Fatalf("status = %d problem = %v", rec.Code, p)
		}
	})
}
