package httpx

import "github.com/Y4NN777/226os/src/api/internal/platform/apperr"

// Pagination bounds (ADR-0004): default 20, maximum 50, never clamped.
const (
	defaultLimit = 20
	maxLimit     = 50
)

// Limit validates the optional `limit` query parameter. Out-of-range values
// are a 422 VALIDATION_ERROR, never silently clamped.
func Limit(v *int) (int, error) {
	if v == nil {
		return defaultLimit, nil
	}
	if *v < 1 || *v > maxLimit {
		return 0, apperr.New(apperr.CodeValidation, "limit must be between 1 and 50.").
			WithViolations(apperr.Violation{Field: "limit", Code: "out_of_range"})
	}
	return *v, nil
}
