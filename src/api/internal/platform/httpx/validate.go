package httpx

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

// maxViolations bounds the violations[] array of one problem response.
const maxViolations = 20

// violationCodes maps the keyword of a failed JSON Schema constraint to the
// closed Violation.code enum of api/openapi.yaml. Anything absent (type, oneOf,
// not, const, ...) is invalid_value.
var violationCodes = map[string]string{
	"required":         "required",
	"minLength":        "too_short",
	"minItems":         "too_short",
	"maxLength":        "too_long",
	"maxItems":         "too_many_items",
	"properties":       "unknown_field", // additionalProperties: false
	"enum":             "invalid_value",
	"minimum":          "out_of_range",
	"maximum":          "out_of_range",
	"exclusiveMinimum": "out_of_range",
	"exclusiveMaximum": "out_of_range",
	"format":           "invalid_format",
	"pattern":          "invalid_format",
	"uniqueItems":      "duplicate",
}

// formatValidators enforce the request formats the engine does not know by
// default; scoped to this validator, so no package-level registry is touched.
var formatValidators = map[string]openapi3.StringFormatValidator{
	"uuid":  openapi3.NewRegexpFormatValidator(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
	"email": openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForEmail),
	"uri": openapi3.NewCallbackValidator(func(s string) error {
		u, err := url.Parse(s)
		if err == nil && u.Scheme == "" {
			err = errors.New("missing scheme")
		}
		return err
	}),
}

var pointerEscaper = strings.NewReplacer("~", "~0", "/", "~1")

// NewValidator builds, once, a request validator for the OpenAPI document spec
// (ADR-0004 point 8) in the shape apigen.ChiServerOptions.HandlerMiddlewares
// expects. It runs after the generated code has bound the path, query and
// header parameters and before the operation handler: the body is read once,
// capped at 1 MiB, checked against the contract (parameters, then the request
// body with its additionalProperties, maxLength, maxItems, enum, format and
// required rules) and handed back untouched to the handler. Security
// requirements are not evaluated here; authentication and CSRF are separate
// middleware. Every refusal is a problem+json through WriteError; submitted
// values are never echoed.
func NewValidator(logger *slog.Logger, spec []byte) (apigen.MiddlewareFunc, error) {
	c, err := loadContract(spec)
	if err != nil {
		return nil, err
	}
	return c.middleware(logger), nil
}

// contract is a parsed OpenAPI document ready to validate requests. It is read-only once built, so
// one instance serves any number of routers and requests.
type contract struct {
	router  routers.Router
	options *openapi3filter.Options
}

// contractCache keeps the last parsed spec: parsing and compiling the whole contract costs tens of
// milliseconds (several times that under -race), which every NewRouter call in a process would pay
// again for the same embedded bytes. A different spec rebuilds.
var contractCache struct {
	sync.Mutex
	spec string
	c    *contract
	err  error
}

func loadContract(spec []byte) (*contract, error) {
	contractCache.Lock()
	defer contractCache.Unlock()
	if contractCache.c != nil || contractCache.err != nil {
		if contractCache.spec == string(spec) {
			return contractCache.c, contractCache.err
		}
	}
	c, err := buildContract(spec)
	contractCache.spec, contractCache.c, contractCache.err = string(spec), c, err
	return c, err
}

func buildContract(spec []byte) (*contract, error) {
	doc, err := openapi3.NewLoader().LoadFromData(spec)
	if err != nil {
		return nil, fmt.Errorf("load openapi document: %w", err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("build openapi router: %w", err)
	}
	// kin-openapi switches to a JSON Schema 2020-12 engine for any 3.1 document.
	// That engine recompiles the schema on every request and flattens failures
	// into text (no keyword, no path), so it cannot drive the violations[] of
	// ADR-0004. The built-in engine reports each failure with its keyword and
	// path and handles the 3.1 constructs this contract uses (anyOf with
	// type null, type arrays); validator_test.go pins that. The router above
	// has already validated the document as 3.1.
	doc.OpenAPI = "3.0.3"
	return &contract{router: router, options: &openapi3filter.Options{
		AuthenticationFunc:      openapi3filter.NoopAuthenticationFunc,
		MultiError:              true,
		SkipSettingDefaults:     true, // never rewrite the body the handler will read
		SchemaValidationOptions: []openapi3.SchemaValidationOption{openapi3.WithStringFormatValidators(formatValidators)},
	}}, nil
}

func (c *contract) middleware(logger *slog.Logger) apigen.MiddlewareFunc {
	write := WriteError(logger)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := validateRequest(w, r, c.router, c.options); err != nil {
				write(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func validateRequest(w http.ResponseWriter, r *http.Request, router routers.Router, options *openapi3filter.Options) error {
	// An absolute-form request target (RFC 9112) carries a host that the relative server URL
	// /api/v1 of the contract cannot match: route on the path alone.
	routed := r
	if r.URL.Host != "" || r.URL.Scheme != "" {
		routed = r.Clone(r.Context())
		routed.URL.Host, routed.URL.Scheme = "", ""
	}
	route, pathParams, err := router.FindRoute(routed)
	if err != nil {
		// chi matched this route, so the contract must know it: refuse, log, never pass.
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		return decodeError(err)
	}
	restore := func() { r.Body = io.NopCloser(bytes.NewReader(data)) }
	restore()
	if len(data) > 0 && route.Operation.RequestBody != nil && !isJSONContentType(r.Header.Get("Content-Type")) {
		// C5: no dedicated code in the enum yet (same answer as DecodeJSON)
		return apperr.New(apperr.CodeValidation, "Content-Type must be application/json.").
			WithViolations(apperr.Violation{Field: "Content-Type", Code: "invalid_value"})
	}
	if trailingJSON(data) {
		// The library validates the first value only; a handler must never read a different document.
		return apperr.New(apperr.CodeMalformedRequest, "The request body must be a single JSON value.")
	}
	err = openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
		Options:    options,
	})
	restore()
	if err == nil {
		return nil
	}
	return contractError(err)
}

// findings accumulates what one failed validation says about the request.
type findings struct {
	malformed  *apperr.Error
	violations []apperr.Violation
}

// contractError turns a kin-openapi error tree into 400 MALFORMED_REQUEST
// (unparseable body or parameter) or 422 VALIDATION_ERROR (constraint
// violations). Library error text is never forwarded: it quotes the value.
func contractError(err error) error {
	var f findings
	f.visit(err)
	if f.malformed != nil {
		return f.malformed
	}
	n := min(len(f.violations), maxViolations)
	return apperr.New(apperr.CodeValidation, "The request does not match the API contract.").
		WithViolations(f.violations[:n]...)
}

func (f *findings) visit(err error) {
	// kin-openapi returns its own error types unwrapped, and errors.As would
	// look through a MultiError or a RequestError and lose the sibling errors or
	// the parameter, so these are plain type assertions.
	if multi, ok := err.(openapi3.MultiError); ok { //nolint:errorlint // see above
		for _, inner := range multi {
			f.visit(inner)
		}
		return
	}
	// Anything that is not a RequestError is refused as malformed, never passed.
	requestErr, _ := err.(*openapi3filter.RequestError) //nolint:errorlint // see above
	f.request(cmp.Or(requestErr, &openapi3filter.RequestError{Err: err}))
}

func (f *findings) request(e *openapi3filter.RequestError) {
	if e.Parameter != nil {
		f.parameter(e)
		return
	}
	if errors.Is(e.Err, openapi3filter.ErrInvalidRequired) {
		f.malformed = apperr.New(apperr.CodeMalformedRequest, "The request body is required.")
		return
	}
	if !f.schema(e.Err, "") {
		f.malformed = apperr.New(apperr.CodeMalformedRequest, "The request body is not valid JSON.")
	}
}

func (f *findings) parameter(e *openapi3filter.RequestError) {
	name := e.Parameter.Name
	switch {
	case errors.Is(e.Err, openapi3filter.ErrInvalidRequired):
		f.malformed = f.malformedParam(name, "required")
	case name == "cursor":
		// ADR-0004 point 5: any unusable cursor is 400 INVALID_CURSOR, not a generic 422.
		f.malformed = apperr.New(apperr.CodeInvalidCursor, "The cursor is not valid.")
	case !f.schema(e.Err, name):
		f.malformed = f.malformedParam(name, "invalid_format")
	}
}

// trailingJSON reports whether data is a well-formed JSON value followed by more data.
func trailingJSON(data []byte) bool {
	dec := json.NewDecoder(bytes.NewReader(data))
	var first json.RawMessage
	if dec.Decode(&first) != nil {
		return false // empty or malformed: the contract validation reports it
	}
	var next json.RawMessage
	return !errors.Is(dec.Decode(&next), io.EOF)
}

func (*findings) malformedParam(name, code string) *apperr.Error {
	return apperr.New(apperr.CodeMalformedRequest, "The request is malformed.").
		WithViolations(apperr.Violation{Field: name, Code: code})
}

// schema records the constraint violations in err and reports whether err was a
// schema-validation failure at all. param, when set, names the violating
// parameter; otherwise the violation field is the JSON pointer into the body.
func (f *findings) schema(err error, param string) bool {
	switch e := err.(type) { //nolint:errorlint // see visit
	case openapi3.MultiError:
		ok := true
		for _, inner := range e {
			ok = f.schema(inner, param) && ok
		}
		return ok
	case *openapi3.SchemaError:
		f.violations = append(f.violations, apperr.Violation{
			Field: truncate(cmp.Or(param, pointerOf(e))),
			Code:  cmp.Or(violationCodes[e.SchemaField], "invalid_value"),
		})
		return true
	default:
		return false
	}
}

// pointerOf renders the location of a body error as a JSON pointer ("/" is the
// document root, because Violation.field cannot be empty). An unknown property
// is reported by its (truncated) name, the only submitted text ever returned,
// as DecodeJSON does.
func pointerOf(e *openapi3.SchemaError) string {
	parts := e.JSONPointer()
	if e.SchemaField == "properties" {
		parts = append(parts, unknownKey(e))
	}
	for i, p := range parts {
		parts[i] = pointerEscaper.Replace(p)
	}
	return "/" + strings.Join(parts, "/")
}

// unknownKey names, without parsing the library message, the first property of
// the offending object that the schema does not declare.
func unknownKey(e *openapi3.SchemaError) string {
	object, _ := e.Value.(map[string]any)
	undeclared := slices.DeleteFunc(slices.Sorted(maps.Keys(object)), func(key string) bool {
		_, declared := e.Schema.Properties[key]
		return declared
	})
	return strings.Join(undeclared[:min(1, len(undeclared))], "")
}

func truncate(field string) string {
	if rs := []rune(field); len(rs) > maxFieldRunes {
		return string(rs[:maxFieldRunes])
	}
	return field
}
