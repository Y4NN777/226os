package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

// statusTable is the single code -> HTTP status mapping (ADR-0004 point 4). It
// must equal the x-error-catalogue of api/openapi.yaml; problem_test.go asserts it.
var statusTable = map[apperr.Code]int{
	apperr.CodeMalformedRequest:     http.StatusBadRequest,
	apperr.CodeInvalidCursor:        http.StatusBadRequest,
	apperr.CodeUnauthenticated:      http.StatusUnauthorized,
	apperr.CodeInvalidCredentials:   http.StatusUnauthorized,
	apperr.CodeCSRFFailed:           http.StatusForbidden,
	apperr.CodeForbidden:            http.StatusForbidden,
	apperr.CodeEmailNotVerified:     http.StatusForbidden,
	apperr.CodeSelfVoteForbidden:    http.StatusForbidden,
	apperr.CodeReauthRequired:       http.StatusForbidden,
	apperr.CodeNotFound:             http.StatusNotFound,
	apperr.CodeDuplicateApplication: http.StatusConflict,
	apperr.CodeDuplicateReview:      http.StatusConflict,
	apperr.CodeVoteConflict:         http.StatusConflict,
	apperr.CodeIdempotencyInProg:    http.StatusConflict,
	apperr.CodeStateConflict:        http.StatusConflict,
	apperr.CodePreconditionFailed:   http.StatusPreconditionFailed,
	apperr.CodePayloadTooLarge:      http.StatusRequestEntityTooLarge,
	apperr.CodeValidation:           http.StatusUnprocessableEntity,
	apperr.CodeInvalidToken:         http.StatusUnprocessableEntity,
	apperr.CodeIdempotencyReused:    http.StatusUnprocessableEntity,
	apperr.CodeInvalidGitHubURL:     http.StatusUnprocessableEntity,
	apperr.CodePreconditionRequired: http.StatusPreconditionRequired,
	apperr.CodeRateLimited:          http.StatusTooManyRequests,
	apperr.CodeInternal:             http.StatusInternalServerError,
	apperr.CodeServiceUnavailable:   http.StatusServiceUnavailable,
	apperr.CodeUpstreamUnavailable:  http.StatusServiceUnavailable,
}

// titleTable holds the stable, human-safe title of each problem type.
var titleTable = map[apperr.Code]string{
	apperr.CodeMalformedRequest:     "Malformed request",
	apperr.CodeInvalidCursor:        "Invalid cursor",
	apperr.CodeUnauthenticated:      "Authentication required",
	apperr.CodeInvalidCredentials:   "Invalid credentials",
	apperr.CodeCSRFFailed:           "CSRF check failed",
	apperr.CodeForbidden:            "Forbidden",
	apperr.CodeEmailNotVerified:     "Email address not verified",
	apperr.CodeSelfVoteForbidden:    "Voting on your own content is not allowed",
	apperr.CodeReauthRequired:       "Re-authentication required",
	apperr.CodeNotFound:             "Not found",
	apperr.CodeDuplicateApplication: "Duplicate application",
	apperr.CodeDuplicateReview:      "Duplicate review",
	apperr.CodeVoteConflict:         "Vote conflict",
	apperr.CodeIdempotencyInProg:    "A request with this idempotency key is in progress",
	apperr.CodeStateConflict:        "State conflict",
	apperr.CodePreconditionFailed:   "If-Match does not match the current ETag",
	apperr.CodePayloadTooLarge:      "Payload too large",
	apperr.CodeValidation:           "Validation failed",
	apperr.CodeInvalidToken:         "Invalid or expired token",
	apperr.CodeIdempotencyReused:    "Idempotency key reused with a different request",
	apperr.CodeInvalidGitHubURL:     "Invalid GitHub URL",
	apperr.CodePreconditionRequired: "If-Match header is required",
	apperr.CodeRateLimited:          "Too many requests",
	apperr.CodeInternal:             "Internal server error",
	apperr.CodeServiceUnavailable:   "Service unavailable",
	apperr.CodeUpstreamUnavailable:  "Upstream service unavailable",
}

type violationBody struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

type problemBody struct {
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	Status     int             `json:"status"`
	Detail     string          `json:"detail,omitempty"`
	Instance   string          `json:"instance"`
	Code       string          `json:"code"`
	RequestID  string          `json:"request_id"`
	Violations []violationBody `json:"violations,omitempty"`
}

// WriteError returns the single error writer: it renders err as RFC 9457
// application/problem+json. An *apperr.Error (found with errors.As) keeps its
// code; anything else, or a code missing from the table, becomes a generic 500
// INTERNAL_ERROR whose real cause is logged with the request_id and never sent.
// Responses with status >= 500 carry no detail.
func WriteError(logger *slog.Logger) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		var ae *apperr.Error
		if !errors.As(err, &ae) {
			ae = apperr.Wrap(apperr.CodeInternal, err)
		}
		status, known := statusTable[ae.Code]
		if !known {
			ae = apperr.Wrap(apperr.CodeInternal, err)
			status = http.StatusInternalServerError
		}
		reqID := requestID(r.Context())
		if status >= http.StatusInternalServerError {
			logger.ErrorContext(r.Context(), "request failed",
				slog.String("request_id", reqID),
				slog.String("code", string(ae.Code)),
				slog.Any("error", err),
			)
		}
		body := problemBody{
			Type:      "urn:226os:problem:" + string(ae.Code),
			Title:     titleTable[ae.Code],
			Status:    status,
			Instance:  capRunes(r.URL.Path, maxInstanceLen),
			Code:      string(ae.Code),
			RequestID: reqID,
		}
		if status < http.StatusInternalServerError {
			body.Detail = ae.Detail
		}
		for _, v := range ae.Violations {
			body.Violations = append(body.Violations, violationBody(v))
		}
		w.Header().Set("Content-Type", "application/problem+json")
		if ae.RetryAfter > 0 {
			w.Header().Set("Retry-After", strconv.FormatInt(int64((ae.RetryAfter+time.Second-1)/time.Second), 10))
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}
}

// Problem.instance and Problem.request_id are bounded by the contract (maxLength 512 and 64); the
// path and the client-supplied X-Request-Id are not, so every echo goes through these caps.
const (
	maxInstanceLen  = 512
	maxRequestIDLen = 64
)

// requestID is the request id, capped for responses and logs.
func requestID(ctx context.Context) string {
	return capRunes(middleware.GetReqID(ctx), maxRequestIDLen)
}

// RequestID returns the request id capped to the contract and audit-log limit.
func RequestID(ctx context.Context) string {
	return requestID(ctx)
}

// capRunes truncates s to at most n runes.
func capRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
