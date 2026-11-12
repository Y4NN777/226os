package apigen_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.yaml.in/yaml/v3"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// These tests bind the generated server code to api/openapi.yaml (the source of
// truth, ADR-0004). The spec is parsed independently of the generator, so a
// generator regression, a missing operation or a stale committed output fails here.

const specPath = "../../../../../api/openapi.yaml"

var httpMethods = map[string]bool{"get": true, "put": true, "post": true, "delete": true, "patch": true}

type operation struct {
	method, path, id string
	params           []map[string]any
}

type spec struct {
	Components struct {
		Parameters map[string]map[string]any `yaml:"parameters"`
		Schemas    map[string]map[string]any `yaml:"schemas"`
	} `yaml:"components"`
	Paths map[string]map[string]any `yaml:"paths"`
}

func loadSpec(t *testing.T) (spec, []operation) {
	t.Helper()
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	var s spec
	if err := yaml.Unmarshal(raw, &s); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	var ops []operation
	for path, item := range s.Paths {
		shared := s.resolveParams(item["parameters"])
		for method, v := range item {
			if !httpMethods[method] {
				continue
			}
			body := v.(map[string]any)
			ops = append(ops, operation{
				method: strings.ToUpper(method),
				path:   path,
				id:     body["operationId"].(string),
				params: append(append([]map[string]any{}, shared...), s.resolveParams(body["parameters"])...),
			})
		}
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].path+ops[i].method < ops[j].path+ops[j].method })
	return s, ops
}

func (s spec) resolveParams(v any) []map[string]any {
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

// sample returns a valid-format value for a parameter schema.
func (s spec) sample(schema map[string]any) string {
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
	switch schema["format"] {
	case "uuid":
		return "0190c3f2-7a3b-7c1e-8a5d-3b1f2c4d5e6f"
	}
	return "alice"
}

func TestRouteTableMatchesContract(t *testing.T) {
	_, ops := loadSpec(t)
	const wantOperations = 47
	if len(ops) != wantOperations {
		t.Fatalf("spec has %d operations, want %d", len(ops), wantOperations)
	}

	want := map[string]bool{}
	for _, o := range ops {
		want[o.method+" "+o.path] = true
	}

	got := map[string]bool{}
	router := apigen.Handler(apigen.Unimplemented{}).(chi.Routes)
	err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		got[method+" "+route] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	for k := range want {
		if !got[k] {
			t.Errorf("operation %q in api/openapi.yaml is not routed by the generated server", k)
		}
	}
	for k := range got {
		if !want[k] {
			t.Errorf("generated server routes %q, which is not in api/openapi.yaml", k)
		}
	}
}

func TestServerInterfaceMatchesOperationIDs(t *testing.T) {
	_, ops := loadSpec(t)
	want := map[string]bool{}
	for _, o := range ops {
		want[strings.ToUpper(o.id[:1])+o.id[1:]] = true
	}
	typ := reflect.TypeFor[apigen.ServerInterface]()
	got := map[string]bool{}
	for i := range typ.NumMethod() {
		got[typ.Method(i).Name] = true
	}
	for id := range want {
		if !got[id] {
			t.Errorf("operationId %q missing from ServerInterface", id)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("ServerInterface method %q has no operationId in the spec", name)
		}
	}
}

func TestEveryOperationReachesItsHandler(t *testing.T) {
	s, ops := loadSpec(t)
	srv := apigen.Handler(apigen.Unimplemented{})
	for _, o := range ops {
		t.Run(o.id, func(t *testing.T) {
			path := o.path
			header := http.Header{}
			query := url.Values{}
			for _, p := range o.params {
				schema, _ := p["schema"].(map[string]any)
				switch p["in"] {
				case "path":
					path = strings.ReplaceAll(path, "{"+p["name"].(string)+"}", s.sample(schema))
				case "query":
					if req, _ := p["required"].(bool); req {
						query.Set(p["name"].(string), s.sample(schema))
					}
				case "header":
					if req, _ := p["required"].(bool); req {
						header.Set(p["name"].(string), s.sample(schema))
					}
				}
			}
			req := httptest.NewRequestWithContext(t.Context(), o.method, path+"?"+query.Encode(), nil)
			req.Header = header
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("%s %s = %d, want 501 from Unimplemented.%s", o.method, path, rec.Code, o.id)
			}
		})
	}
}

func TestInvalidParametersAreRejectedBeforeTheHandler(t *testing.T) {
	_, ops := loadSpec(t)
	srv := apigen.Handler(apigen.Unimplemented{})
	var rejected int
	for _, o := range ops {
		for _, p := range o.params {
			if p["in"] != "path" || p["name"] != "questionId" {
				continue
			}
			path := strings.ReplaceAll(o.path, "{questionId}", "not-a-uuid")
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), o.method, path, nil))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s %s = %d, want 400 for a malformed UUID", o.method, path, rec.Code)
			}
			rejected++
		}
	}
	if rejected == 0 {
		t.Fatal("no operation with a questionId path parameter was exercised")
	}
}
