package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

const (
	maxBodyBytes  = 1 << 20 // 1 MiB
	maxFieldRunes = 128     // Violation.field maxLength in the OpenAPI schema
	unknownField  = "json: unknown field "
)

// DecodeJSON strictly decodes a JSON request body into dst. It requires an
// application/json Content-Type, caps the body at 1 MiB, rejects unknown
// fields and trailing data, and returns only *apperr.Error. Input text is
// never echoed except the (truncated) name of an unknown field.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if !isJSONContentType(r.Header.Get("Content-Type")) {
		// C5: no dedicated code in the enum yet
		return apperr.New(apperr.CodeValidation, "Content-Type must be application/json.").
			WithViolations(apperr.Violation{Field: "Content-Type", Code: "invalid_value"})
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return decodeError(err)
	}
	var extra json.RawMessage
	switch err := dec.Decode(&extra); {
	case errors.Is(err, io.EOF):
		return nil
	case err == nil:
		return apperr.New(apperr.CodeMalformedRequest, "The body must contain a single JSON value.")
	default:
		return decodeError(err)
	}
}

func isJSONContentType(ct string) bool {
	mt, params, err := mime.ParseMediaType(ct)
	if err != nil || mt != "application/json" {
		return false
	}
	charset, ok := params["charset"]
	return !ok || strings.EqualFold(charset, "utf-8")
}

func decodeError(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return apperr.New(apperr.CodePayloadTooLarge, "The request body is too large.")
	}
	// encoding/json reports an unknown field only through its message.
	if name, ok := strings.CutPrefix(err.Error(), unknownField); ok {
		name = strings.Trim(name, `"`)
		if rs := []rune(name); len(rs) > maxFieldRunes {
			name = string(rs[:maxFieldRunes])
		}
		return apperr.New(apperr.CodeValidation, "The body contains an unknown field.").
			WithViolations(apperr.Violation{Field: name, Code: "unknown_field"})
	}
	return apperr.New(apperr.CodeMalformedRequest, "The request body is not valid JSON.")
}
