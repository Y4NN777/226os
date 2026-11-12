package idempotency_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
)

// TestOperationsMatchTheContract fails when the code and api/openapi.yaml
// disagree on which operations bind the Idempotency-Key parameter.
func TestOperationsMatchTheContract(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		if parent := filepath.Dir(dir); parent != dir {
			dir = parent
			continue
		}
		t.Fatal("go.mod not found")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Servers []struct{ URL string } `yaml:"servers"`
		Paths   map[string]map[string]struct {
			Parameters []struct {
				Ref string `yaml:"$ref"`
			} `yaml:"parameters"`
		} `yaml:"paths"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	var want []string
	for path, item := range doc.Paths {
		for method, op := range item {
			for _, p := range op.Parameters {
				if strings.HasSuffix(p.Ref, "/parameters/IdempotencyKey") {
					want = append(want, strings.ToUpper(method)+" "+doc.Servers[0].URL+path)
				}
			}
		}
	}
	var got []string
	for _, op := range idempotency.Operations() {
		got = append(got, op.Method+" "+op.Pattern)
	}
	slices.Sort(want)
	slices.Sort(got)
	if len(want) != 6 || !slices.Equal(want, got) {
		t.Fatalf("contract operations %v, code operations %v", want, got)
	}
}

func TestOperationsReturnsACopy(t *testing.T) {
	ops := idempotency.Operations()
	ops[0].Pattern = "/tampered"
	if idempotency.Operations()[0].Pattern == "/tampered" {
		t.Fatal("Operations must not expose the internal set")
	}
}
