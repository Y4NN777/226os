package idempotency_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
)

// The request hash covers method, route, subject and body, deliberately NOT the query string,
// because every Idempotency-Key operation takes no query parameter. These tests fail the day
// that stops being true, so the assumption cannot rot silently.

const queryGuardAdvice = "an Idempotency-Key operation now declares a query parameter: add the query to the " +
	"idempotency request hash (envelope.go) and extend the replay tests (same key with a different query must " +
	"be refused) before changing the contract"

// boundOperationsWithQueryParams returns the "METHOD path" of every bound operation whose own
// parameters, or whose path item's, include one with in: query (components/parameters $refs followed).
func boundOperationsWithQueryParams(spec []byte, bound []idempotency.Operation) ([]string, error) {
	var doc struct {
		Servers []struct {
			URL string `yaml:"url"`
		} `yaml:"servers"`
		Paths      map[string]map[string]yaml.Node `yaml:"paths"`
		Components struct {
			Parameters map[string]struct {
				In string `yaml:"in"`
			} `yaml:"parameters"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	prefix := ""
	if len(doc.Servers) > 0 {
		prefix = doc.Servers[0].URL
	}
	isQuery := func(n *yaml.Node) (bool, error) {
		var list []struct {
			Ref string `yaml:"$ref"`
			In  string `yaml:"in"`
		}
		if err := n.Decode(&list); err != nil {
			return false, err
		}
		for _, p := range list {
			in := p.In
			if p.Ref != "" {
				in = doc.Components.Parameters[p.Ref[strings.LastIndex(p.Ref, "/")+1:]].In
			}
			if in == "query" {
				return true, nil
			}
		}
		return false, nil
	}
	var hits []string
	for _, op := range bound {
		item := doc.Paths[strings.TrimPrefix(op.Pattern, prefix)]
		for _, key := range []string{"parameters", strings.ToLower(op.Method)} {
			node, ok := item[key]
			if !ok {
				continue
			}
			params := node
			if key != "parameters" {
				var o map[string]yaml.Node
				if err := node.Decode(&o); err != nil {
					return nil, err
				}
				if params, ok = o["parameters"]; !ok {
					continue
				}
			}
			q, err := isQuery(&params)
			if err != nil {
				return nil, err
			}
			if q {
				hits = append(hits, op.Method+" "+op.Pattern)
				break
			}
		}
	}
	slices.Sort(hits)
	return hits, nil
}

func TestQueryGuardDetectsQueryParameters(t *testing.T) {
	const spec = `
servers: [{url: /api/v1}]
paths:
  /own:
    post:
      parameters:
        - {name: q, in: query, schema: {type: string}}
  /byref:
    post:
      parameters:
        - $ref: '#/components/parameters/Filter'
  /pathitem:
    parameters:
      - {name: q, in: query, schema: {type: string}}
    post: {responses: {}}
  /clean:
    post:
      parameters:
        - $ref: '#/components/parameters/Key'
        - {name: id, in: path}
  /bare:
    post: {responses: {}}
components:
  parameters:
    Filter: {name: f, in: query}
    Key: {name: Idempotency-Key, in: header}
`
	post := func(p string) idempotency.Operation {
		return idempotency.Operation{Method: http.MethodPost, Pattern: "/api/v1" + p}
	}
	got, err := boundOperationsWithQueryParams([]byte(spec), []idempotency.Operation{
		post("/own"), post("/byref"), post("/pathitem"), post("/clean"), post("/bare"), post("/absent"),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"POST /api/v1/byref", "POST /api/v1/own", "POST /api/v1/pathitem"}
	if !slices.Equal(got, want) {
		t.Fatalf("hits = %v, want %v", got, want)
	}
}

func TestQueryGuardRejectsAnUnparsableSpec(t *testing.T) {
	for name, spec := range map[string]string{
		"not yaml":                "paths: [",
		"parameters not a list":   "paths: {/x: {post: {parameters: 3}}}",
		"operation not a mapping": "paths: {/x: {post: [1]}}",
	} {
		op := []idempotency.Operation{{Method: http.MethodPost, Pattern: "/x"}}
		if _, err := boundOperationsWithQueryParams([]byte(spec), op); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// TestBoundOperationsTakeNoQueryParameters runs the guard on the real contract.
func TestBoundOperationsTakeNoQueryParameters(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
	raw, err := os.ReadFile(filepath.Join(dir, "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	hits, err := boundOperationsWithQueryParams(raw, idempotency.Operations())
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("%s: %v", queryGuardAdvice, hits)
	}
}
