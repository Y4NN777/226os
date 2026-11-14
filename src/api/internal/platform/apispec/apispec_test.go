package apispec_test

import (
	"bytes"
	"os"
	"strings"
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

// TestLinkTypesMatchTheStoredPairs pins the owner decision of 2026-11-14 (S3): a question links
// only to a resource, a project or an article, and no write link may target a series, because
// content_relations stores no such pair (migration 00009).
func TestLinkTypesMatchTheStoredPairs(t *testing.T) {
	type schema struct {
		Properties map[string]struct {
			Enum  []string `yaml:"enum"`
			Items struct {
				Ref string `yaml:"$ref"`
			} `yaml:"items"`
		} `yaml:"properties"`
	}
	var doc struct {
		Components struct {
			Schemas map[string]schema `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(apispec.YAML, &doc); err != nil {
		t.Fatalf("parse embedded spec: %v", err)
	}
	s := doc.Components.Schemas
	for _, name := range []string{"QuestionCreate", "QuestionUpdate"} {
		if got := s[name].Properties["links"].Items.Ref; got != "#/components/schemas/QuestionLinkRef" {
			t.Errorf("%s.links items = %q, want QuestionLinkRef", name, got)
		}
	}
	if got := strings.Join(s["QuestionLinkRef"].Properties["type"].Enum, ","); got != "resource,project,article" {
		t.Errorf("QuestionLinkRef types = %q", got)
	}
	if got := strings.Join(s["LinkRef"].Properties["type"].Enum, ","); got != "question,resource,project,article" {
		t.Errorf("LinkRef types = %q", got)
	}
}
