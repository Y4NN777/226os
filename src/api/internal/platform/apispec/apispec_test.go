package apispec_test

import (
	"bytes"
	"os"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/Y4NN777/226os/src/api/internal/platform/apispec"
)

// Same repo-root-relative path as apigen/contract_test.go.
const specPath = "../../../../../api/openapi.yaml"

func TestEmbeddedSpecMatchesContract(t *testing.T) {
	want, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	if len(apispec.YAML) == 0 {
		t.Fatal("embedded spec is empty")
	}
	if !bytes.Equal(apispec.YAML, want) {
		t.Fatal("apispec/openapi.gen.yaml differs from api/openapi.yaml; refresh the copy")
	}
}

// The request validator answers 422 to a malformed path username (the `Username` parameter pattern), so the contract
// must declare that response on every operation that takes it (issue #116, owner decision 2026-11-14).
func TestGetProfileDeclaresTheValidatorsUnprocessable(t *testing.T) {
	var doc struct {
		Paths map[string]map[string]struct {
			Responses map[string]any `yaml:"responses"`
		} `yaml:"paths"`
	}
	if err := yaml.Unmarshal(apispec.YAML, &doc); err != nil {
		t.Fatalf("parse embedded spec: %v", err)
	}
	if _, ok := doc.Paths["/profiles/{username}"]["get"].Responses["422"]; !ok {
		t.Fatal("getProfile does not declare 422 for a malformed username")
	}
}
