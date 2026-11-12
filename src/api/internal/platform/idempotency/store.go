// Package idempotency implements the Idempotency-Key convention of ADR-0004
// point 6: a key is bound to the subject, the route template and a hash of the
// request (method, route template, concrete escaped path and body), kept for 24
// hours; the same key and request replays the stored response, a different request
// answers 422 IDEMPOTENCY_KEY_REUSED and a concurrent duplicate answers 409
// IDEMPOTENCY_IN_PROGRESS. Database uniqueness (the table primary key) is the real
// safety net.
//
// Caching policy: the outcome of a request is stored (Complete) when its status is 2xx
// or 4xx, except 401, 403, 408, 425 and 429, which describe the caller's session or
// capacity rather than the outcome of this request. Those, redirects and every status
// >= 500 are abandoned so that a retry runs again; so is a handler that wrote nothing
// because the request context ended (timeout, client cancel).
//
// A replay is the stored response: status, Content-Type, the allowlisted headers
// (Location, ETag; never Set-Cookie or the first request's X-Request-Id) and the body.
// The table has no header column, so response_body holds a versioned envelope
// {"v":1,"content_type":...,"headers":{...},"body":{...}}; the body is omitted
// ("body_omitted":true) when it is absent, not a JSON object, PostgreSQL refuses it (invalid
// UTF-8, a \u0000 escape) or the envelope would exceed the 16384-byte column limit. A row of
// an older shape (any other object, or NULL) replays as its status and application/json.
//
// The size limit is measured on the jsonb text form, which PostgreSQL renders with a space
// after every ":" and ",": the effective raw body limit is therefore BELOW 16 KiB, by an
// amount that grows with the number of keys and array items. The handlers of the six bound
// operations must keep their 201 bodies small (id, Location) for a full replay. A replay
// whose body was lost is served with the status and the allowlisted headers only, without
// the stored Content-Type (an empty body labelled application/json breaks res.json() in a
// client) and with "Idempotency-Body-Omitted: true" so that the loss is explicit. If
// PostgreSQL also refuses a header value (a NUL in Location), the row is still completed,
// with a minimal envelope (version, content type, body_omitted) and no header.
//
// A replayed body is the jsonb normal form of the first body: key order and whitespace may
// differ, duplicate keys collapse to the last one, and a number such as 1e2 becomes 100. It
// is semantically equal to the first response, not byte-identical.
//
// The idempotency row is not part of the business transaction: a crash between the
// handler's commit and Complete leaves the row in progress until the 2-minute takeover
// (AbandonAfter), after which an identical retry may run the handler again. The unique
// constraints of the business tables are the backstop.
package idempotency

import (
	"context"
	"errors"
	"time"
)

// Retention is how long a key is remembered.
const Retention = 24 * time.Hour

// AbandonAfter is the age after which an in_progress row is considered
// abandoned (its owner crashed; the router timeout is 30 s) and may be taken
// over by a new identical request.
const AbandonAfter = 2 * time.Minute

// MaxStoredBody is the largest response body the response_body column accepts.
const MaxStoredBody = 16384

// Outcome is the result of Begin.
type Outcome int

// Outcomes of Begin.
const (
	// Started: the row was inserted as in_progress; the caller owns the key.
	Started Outcome = iota + 1
	// Replay: the same request already completed; the Response holds its response.
	Replay
	// InProgress: the same request is still running and not abandoned.
	InProgress
	// Reused: the key belongs to a different request (hash, method or route).
	Reused
)

// Attempt identifies one claim of a key: the created_at of the row as stored
// (microseconds since the Unix epoch). A takeover refreshes it, so the original owner of a
// taken-over key can neither complete nor abandon the new owner's row.
type Attempt int64

// ErrLostKey means the row is no longer this attempt's: it was completed, purged or taken over.
var ErrLostKey = errors.New("idempotency key is no longer owned by this attempt")

// Response is a completed response as stored and replayed. Body is nil when none was
// stored (absent, not a JSON object, or too large); Headers holds only the replayed
// allowlist (Location, ETag).
type Response struct {
	Status      int
	ContentType string
	Headers     map[string]string
	Body        []byte
}

// Store persists idempotency records. userID is the subject (never client
// supplied); key is the raw header value.
type Store interface {
	// Begin claims the key or reports why it cannot. An expired row counts as absent.
	// The Attempt is meaningful for Started only; Response only for Replay.
	Begin(ctx context.Context, userID, key, method, route string, hash [32]byte, now time.Time) (Outcome, Attempt, Response, error)
	// Complete records the response of the owner of an in_progress key; it returns
	// ErrLostKey when the attempt no longer owns the row.
	Complete(ctx context.Context, userID, key string, attempt Attempt, resp Response, now time.Time) error
	// Abandon deletes the in_progress row of this attempt so that a retry can run; it
	// returns ErrLostKey when the attempt no longer owns the row.
	Abandon(ctx context.Context, userID, key string, attempt Attempt) error
	// Purge deletes the rows with expires_at <= now and returns how many.
	Purge(ctx context.Context, now time.Time) (int64, error)
}
