package apispec_test

import (
	"bytes"
	"os"
	"testing"

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
