package httpx_test

import (
	"errors"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

func TestLimit(t *testing.T) {
	p := func(n int) *int { return &n }
	tests := []struct {
		name    string
		in      *int
		want    int
		invalid bool
	}{
		{"absent defaults to 20", nil, 20, false},
		{"min", p(1), 1, false},
		{"mid", p(37), 37, false},
		{"max", p(50), 50, false},
		{"zero rejected", p(0), 0, true},
		{"negative rejected", p(-3), 0, true},
		{"above max rejected, never clamped", p(51), 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := httpx.Limit(tc.in)
			if !tc.invalid {
				if err != nil || got != tc.want {
					t.Fatalf("got %d, %v; want %d", got, err, tc.want)
				}
				return
			}
			var ae *apperr.Error
			if !errors.As(err, &ae) || ae.Code != apperr.CodeValidation {
				t.Fatalf("err = %v, want VALIDATION_ERROR", err)
			}
			if len(ae.Violations) != 1 || ae.Violations[0].Field != "limit" || ae.Violations[0].Code != "out_of_range" {
				t.Fatalf("violations = %+v", ae.Violations)
			}
			if got != 0 {
				t.Fatalf("value on error = %d", got)
			}
		})
	}
}
