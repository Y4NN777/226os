package apperr_test

import (
	"errors"
	"os"
	"sort"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

const specPath = "../../../../../api/openapi.yaml"

func enumFromSpec(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	var s struct {
		Components struct {
			Schemas struct {
				ErrorCode struct {
					Enum []string `yaml:"enum"`
				} `yaml:"ErrorCode"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(raw, &s); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	return s.Components.Schemas.ErrorCode.Enum
}

func TestCodesMatchSpecEnumExactly(t *testing.T) {
	want := enumFromSpec(t)
	if len(want) == 0 {
		t.Fatal("ErrorCode enum not found in spec")
	}
	var got []string
	for _, c := range apperr.Codes() {
		got = append(got, string(c))
	}
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("codes = %v, spec enum = %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("codes = %v, spec enum = %v", got, want)
		}
	}
}

func TestConstantValues(t *testing.T) {
	if apperr.CodeNotFound != "NOT_FOUND" || apperr.CodeRateLimited != "RATE_LIMITED" {
		t.Fatal("constant values differ from the wire codes")
	}
}

func TestNewAndViolations(t *testing.T) {
	e := apperr.New(apperr.CodeValidation, "bad input").
		WithViolations(apperr.Violation{Field: "title", Code: "required"}).
		WithRetryAfter(90 * time.Second)
	if e.Code != apperr.CodeValidation || e.Detail != "bad input" {
		t.Fatalf("unexpected %+v", e)
	}
	if len(e.Violations) != 1 || e.Violations[0].Field != "title" {
		t.Fatalf("violations = %+v", e.Violations)
	}
	if e.RetryAfter != 90*time.Second {
		t.Fatalf("retry after = %v", e.RetryAfter)
	}
	if e.Error() != "VALIDATION_ERROR: bad input" {
		t.Fatalf("Error() = %q", e.Error())
	}
	if e.Unwrap() != nil {
		t.Fatal("New must not carry a cause")
	}
}

func TestWrapKeepsCause(t *testing.T) {
	cause := errors.New("pq: relation missing")
	e := apperr.Wrap(apperr.CodeInternal, cause)
	if !errors.Is(e, cause) {
		t.Fatal("errors.Is must reach the cause")
	}
	if e.Detail != "" {
		t.Fatalf("Wrap must not copy the cause into Detail, got %q", e.Detail)
	}
	if e.Error() != "INTERNAL_ERROR: pq: relation missing" {
		t.Fatalf("Error() = %q", e.Error())
	}
	var target *apperr.Error
	if !errors.As(e, &target) || target.Code != apperr.CodeInternal {
		t.Fatal("errors.As must find *apperr.Error")
	}
}

func TestErrorWithoutDetailOrCause(t *testing.T) {
	if got := apperr.New(apperr.CodeNotFound, "").Error(); got != "NOT_FOUND" {
		t.Fatalf("Error() = %q", got)
	}
}
