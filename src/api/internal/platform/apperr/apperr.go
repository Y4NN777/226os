// Package apperr defines the typed application error shared by domain code and
// the HTTP layer. Domain code returns *Error; only httpx maps it to a status
// (ADR-0004), so this package has no HTTP knowledge.
package apperr

import "time"

// Code is a stable machine code; the set equals the ErrorCode enum of api/openapi.yaml.
type Code string

// One constant per value of the ErrorCode enum.
const (
	CodeMalformedRequest     Code = "MALFORMED_REQUEST"
	CodeInvalidCursor        Code = "INVALID_CURSOR"
	CodeUnauthenticated      Code = "UNAUTHENTICATED"
	CodeInvalidCredentials   Code = "INVALID_CREDENTIALS" //nolint:gosec // G101: an error-code identifier from the contract enum, not a credential
	CodeCSRFFailed           Code = "CSRF_FAILED"
	CodeForbidden            Code = "FORBIDDEN"
	CodeEmailNotVerified     Code = "EMAIL_NOT_VERIFIED"
	CodeSelfVoteForbidden    Code = "SELF_VOTE_FORBIDDEN"
	CodeReauthRequired       Code = "REAUTH_REQUIRED"
	CodeNotFound             Code = "NOT_FOUND"
	CodeDuplicateApplication Code = "DUPLICATE_APPLICATION"
	CodeDuplicateReview      Code = "DUPLICATE_REVIEW"
	CodeVoteConflict         Code = "VOTE_CONFLICT"
	CodeIdempotencyInProg    Code = "IDEMPOTENCY_IN_PROGRESS"
	CodeStateConflict        Code = "STATE_CONFLICT"
	CodePreconditionFailed   Code = "PRECONDITION_FAILED"
	CodePayloadTooLarge      Code = "PAYLOAD_TOO_LARGE"
	CodeValidation           Code = "VALIDATION_ERROR"
	CodeInvalidToken         Code = "INVALID_TOKEN"
	CodeIdempotencyReused    Code = "IDEMPOTENCY_KEY_REUSED"
	CodeInvalidGitHubURL     Code = "INVALID_GITHUB_URL"
	CodePreconditionRequired Code = "PRECONDITION_REQUIRED"
	CodeRateLimited          Code = "RATE_LIMITED"
	CodeInternal             Code = "INTERNAL_ERROR"
	CodeServiceUnavailable   Code = "SERVICE_UNAVAILABLE"
	CodeUpstreamUnavailable  Code = "UPSTREAM_UNAVAILABLE"
)

// Codes lists every defined code, in spec order.
func Codes() []Code {
	return []Code{
		CodeMalformedRequest, CodeInvalidCursor, CodeUnauthenticated, CodeInvalidCredentials,
		CodeCSRFFailed, CodeForbidden, CodeEmailNotVerified, CodeSelfVoteForbidden,
		CodeReauthRequired, CodeNotFound, CodeDuplicateApplication, CodeDuplicateReview,
		CodeVoteConflict, CodeIdempotencyInProg, CodeStateConflict, CodePreconditionFailed,
		CodePayloadTooLarge, CodeValidation, CodeInvalidToken, CodeIdempotencyReused,
		CodeInvalidGitHubURL, CodePreconditionRequired, CodeRateLimited, CodeInternal,
		CodeServiceUnavailable, CodeUpstreamUnavailable,
	}
}

// Violation is one field-level problem (RFC 9457 extension member).
type Violation struct {
	Field string
	Code  string
}

// Error is the application error. Detail is client-safe text; the wrapped
// cause is for logs only and is never serialised.
type Error struct {
	Code       Code
	Detail     string
	Violations []Violation
	RetryAfter time.Duration // honoured on any response that carries Retry-After (429, 503)
	cause      error
}

// New builds an error with a client-safe detail.
func New(code Code, detail string) *Error {
	return &Error{Code: code, Detail: detail}
}

// Wrap attaches an internal cause (logged, never sent to the client).
func Wrap(code Code, cause error) *Error {
	return &Error{Code: code, cause: cause}
}

// WithViolations adds field violations and returns the receiver.
func (e *Error) WithViolations(v ...Violation) *Error {
	e.Violations = append(e.Violations, v...)
	return e
}

// WithRetryAfter sets the Retry-After hint and returns the receiver.
func (e *Error) WithRetryAfter(d time.Duration) *Error {
	e.RetryAfter = d
	return e
}

// Error implements error; for logs, not for responses.
func (e *Error) Error() string {
	s := string(e.Code)
	if e.Detail != "" {
		s += ": " + e.Detail
	}
	if e.cause != nil {
		s += ": " + e.cause.Error()
	}
	return s
}

// Unwrap exposes the cause to errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.cause }
