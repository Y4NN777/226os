package httpx

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

// APIErrorHandler adapts the errors raised by the generated parameter binding
// (apigen.ChiServerOptions.ErrorHandlerFunc) to problem+json. A missing
// If-Match is 428 PRECONDITION_REQUIRED; Idempotency-Key is optional in the
// spec, so it has no required-case. Every other bad parameter is 400
// MALFORMED_REQUEST with a violation naming the parameter; the parser's own
// error text is never sent.
func APIErrorHandler(logger *slog.Logger) func(w http.ResponseWriter, r *http.Request, err error) {
	write := WriteError(logger)
	return func(w http.ResponseWriter, r *http.Request, err error) {
		write(w, r, paramError(err))
	}
}

func paramError(err error) *apperr.Error {
	var (
		reqParam  *apigen.RequiredParamError
		reqHeader *apigen.RequiredHeaderError
		badFormat *apigen.InvalidParamFormatError
		unmarshal *apigen.UnmarshalingParamError
		badCookie *apigen.UnescapedCookieParamError
		tooMany   *apigen.TooManyValuesForParamError
		malformed = func() *apperr.Error { return apperr.New(apperr.CodeMalformedRequest, "The request is malformed.") }
		withIssue = func(field, code string) *apperr.Error {
			return malformed().WithViolations(apperr.Violation{Field: field, Code: code})
		}
		requiredAs = func(name string) *apperr.Error {
			if strings.EqualFold(name, "If-Match") {
				return apperr.New(apperr.CodePreconditionRequired, "The If-Match header is required.")
			}
			return withIssue(name, "required")
		}
	)
	switch {
	case errors.As(err, &reqParam):
		return requiredAs(reqParam.ParamName)
	case errors.As(err, &reqHeader):
		return requiredAs(reqHeader.ParamName)
	case errors.As(err, &badFormat):
		return withIssue(badFormat.ParamName, "invalid_format")
	case errors.As(err, &unmarshal):
		return withIssue(unmarshal.ParamName, "invalid_format")
	case errors.As(err, &badCookie):
		return withIssue(badCookie.ParamName, "invalid_format")
	case errors.As(err, &tooMany):
		return withIssue(tooMany.ParamName, "too_many_items")
	default:
		return malformed()
	}
}
