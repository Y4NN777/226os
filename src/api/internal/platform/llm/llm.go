// Package llm is the port to a large language model provider. It exists so that no domain code
// imports a provider SDK: the assistant depends on Client, and the adapter (Google AI Studio, or
// the Fake used in tests) is chosen by configuration at the composition root.
//
// Prompts and answers are never logged, and never appear in an error message: errors carry a
// sentinel, an HTTP status and a short fixed phrase only. Classify them with errors.Is.
package llm

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Limits enforced on every request before any I/O.
const (
	MaxUserBytes           = 32768
	MaxSystemBytes         = 8192
	MaxOutputTokensLimit   = 8192
	DefaultMaxOutputTokens = 1024
)

// Sentinel errors. They stay usable with errors.Is through wrapping.
var (
	// ErrNotConfigured means no provider is configured (AI disabled, or no key).
	ErrNotConfigured = errors.New("llm: not configured")
	// ErrInvalidRequest means the request was refused locally, before any I/O.
	ErrInvalidRequest = errors.New("llm: invalid request")
	// ErrQuotaExhausted means the provider answered HTTP 429.
	ErrQuotaExhausted = errors.New("llm: quota exhausted")
	// ErrUnavailable covers network errors, timeouts, 5xx, and empty or blocked output.
	ErrUnavailable = errors.New("llm: unavailable")
	// ErrRejected covers HTTP 400, 401, 403 and 404: a bad key or an unknown model, that is a
	// configuration problem rather than a transient one.
	ErrRejected = errors.New("llm: rejected")
)

// Request is one single-turn generation request.
type Request struct {
	System          string // optional instruction, 0..8192 bytes
	User            string // the question, 1..32768 bytes
	MaxOutputTokens int    // 1..8192; zero means DefaultMaxOutputTokens
}

// Response is the provider answer.
type Response struct {
	Text         string
	Model        string
	PromptTokens int
	OutputTokens int
}

// Client generates text. Implementations are safe for concurrent use.
type Client interface {
	Generate(ctx context.Context, r Request) (Response, error)
}

// Disabled is the Client used when AI is not configured.
type Disabled struct{}

// Generate always returns ErrNotConfigured.
func (Disabled) Generate(context.Context, Request) (Response, error) {
	return Response{}, ErrNotConfigured
}

// StatusError is a provider error that carries the HTTP status. Kind is one of the sentinels, so
// errors.Is keeps working; the message is "<sentinel> (http <status>)".
type StatusError struct {
	Kind   error
	Status int
}

func (e *StatusError) Error() string { return fmt.Sprintf("%v (http %d)", e.Kind, e.Status) }

// Unwrap returns the sentinel.
func (e *StatusError) Unwrap() error { return e.Kind }

// StatusOf returns the HTTP status carried by err, or zero when there is none.
func StatusOf(err error) int {
	var se *StatusError
	if errors.As(err, &se) {
		return se.Status
	}
	return 0
}

var modelName = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)

// ValidModelName reports whether s is a safe model identifier (no path or query injection).
func ValidModelName(s string) bool { return modelName.MatchString(s) }

// normalize validates r and applies the default token budget. The error wraps ErrInvalidRequest
// and never contains request content.
func normalize(r Request) (Request, error) {
	switch {
	case r.User == "":
		return r, invalid("user text is empty")
	case len(r.User) > MaxUserBytes:
		return r, invalid("user text is too long")
	case len(r.System) > MaxSystemBytes:
		return r, invalid("system text is too long")
	case strings.IndexByte(r.User, 0) >= 0 || strings.IndexByte(r.System, 0) >= 0:
		return r, invalid("text contains a NUL byte")
	case r.MaxOutputTokens < 0 || r.MaxOutputTokens > MaxOutputTokensLimit:
		return r, invalid("max output tokens is out of range")
	case r.MaxOutputTokens == 0:
		r.MaxOutputTokens = DefaultMaxOutputTokens
	}
	return r, nil
}

func invalid(phrase string) error { return fmt.Errorf("%w: %s", ErrInvalidRequest, phrase) }
