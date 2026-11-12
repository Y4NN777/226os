package httpx_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

func TestAPIErrorHandler(t *testing.T) {
	inner := errors.New("strconv: parsing \"abc\" secret-internal")
	tests := []struct {
		name      string
		err       error
		status    int
		code      string
		wantField string
		wantVCode string
	}{
		{"missing If-Match (query-style error)", &apigen.RequiredParamError{ParamName: "If-Match"}, 428, "PRECONDITION_REQUIRED", "", ""},
		{"missing If-Match (header error)", &apigen.RequiredHeaderError{ParamName: "If-Match"}, 428, "PRECONDITION_REQUIRED", "", ""},
		{"missing other required param", &apigen.RequiredParamError{ParamName: "q"}, 400, "MALFORMED_REQUEST", "q", "required"},
		{"missing other required header", &apigen.RequiredHeaderError{ParamName: "X-Thing"}, 400, "MALFORMED_REQUEST", "X-Thing", "required"},
		{"invalid format", &apigen.InvalidParamFormatError{ParamName: "id", Err: inner}, 400, "MALFORMED_REQUEST", "id", "invalid_format"},
		{"unmarshaling", &apigen.UnmarshalingParamError{ParamName: "lang", Err: inner}, 400, "MALFORMED_REQUEST", "lang", "invalid_format"},
		{"too many values", &apigen.TooManyValuesForParamError{ParamName: "tag", Count: 2}, 400, "MALFORMED_REQUEST", "tag", "too_many_items"},
		{"bad cookie", &apigen.UnescapedCookieParamError{ParamName: "sid", Err: inner}, 400, "MALFORMED_REQUEST", "sid", "invalid_format"},
		{"unknown generator error", inner, 400, "MALFORMED_REQUEST", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytesBuf
			h := apiHandler(&logs)
			rec, body := run(t, func(w http.ResponseWriter, r *http.Request) { h(w, r, tc.err) })
			if rec.Code != tc.status || body.Code != tc.code {
				t.Fatalf("got %d %s, want %d %s", rec.Code, body.Code, tc.status, tc.code)
			}
			if tc.wantField == "" {
				if len(body.Violations) != 0 {
					t.Fatalf("unexpected violations %+v", body.Violations)
				}
			} else if len(body.Violations) != 1 || body.Violations[0].Field != tc.wantField || body.Violations[0].Code != tc.wantVCode {
				t.Fatalf("violations = %+v", body.Violations)
			}
			if containsAny(rec.Body.String(), "secret-internal", "strconv") {
				t.Fatalf("leaks parser error text: %s", rec.Body.String())
			}
		})
	}
}
