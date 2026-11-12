package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	// HeaderKey is the request header carrying the key.
	HeaderKey = "Idempotency-Key"
	// HeaderReplayed marks a response served from the store.
	HeaderReplayed = "Idempotency-Replayed"
	// HeaderBodyOmitted marks a replay whose stored body was lost (too large, not a JSON
	// object, or refused by PostgreSQL): status and headers only.
	HeaderBodyOmitted = "Idempotency-Body-Omitted"

	minKeyLen = 16
	maxKeyLen = 128
	// maxRequestBody mirrors the 1 MiB cap of the contract validator.
	maxRequestBody = 1 << 20
	// storeTimeout bounds every database call made by the middleware.
	storeTimeout = 5 * time.Second
	// logPrefixLen is how much of a key may appear in logs.
	logPrefixLen = 8
	maxReqIDLen  = 64
)

// OnError writes an error response; httpx passes WriteError. It receives
// *apperr.Error values.
type OnError = func(w http.ResponseWriter, r *http.Request, err error)

// Middleware applies the Idempotency-Key convention to the operations of
// Operations(), when the header is present and the caller is authenticated.
// Anything else passes through untouched. A nil store disables it (pass
// through); the composition root always wires the PostgreSQL store.
//
// Attach it as an apigen ChiServerOptions middleware, not with chi's Use:
// the route pattern is only final once routing ended.
func Middleware(store Store, logger *slog.Logger, onError OnError) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if store == nil {
			return next
		}
		m := &mw{store: store, logger: logger, onError: onError, next: next}
		return http.HandlerFunc(m.serve)
	}
}

type mw struct {
	store   Store
	logger  *slog.Logger
	onError OnError
	next    http.Handler
}

func (m *mw) serve(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get(HeaderKey)
	rc := chi.RouteContext(r.Context())
	if key == "" || rc == nil || !bound(r.Method, rc.RoutePattern()) {
		m.next.ServeHTTP(w, r)
		return
	}
	sub := subject.From(r.Context())
	if sub.Anonymous() {
		m.next.ServeHTTP(w, r) // the operation answers 401 itself
		return
	}
	if err := checkKey(key); err != nil {
		m.onError(w, r, err)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBody))
	if err != nil {
		m.onError(w, r, bodyError(err))
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	route := rc.RoutePattern()
	outcome, attempt, stored, err := m.begin(r.Context(), sub.UserID, key, r.Method, route, hashRequest(r.Method, route, r.URL.EscapedPath(), body))
	if err != nil {
		m.logger.WarnContext(r.Context(), "idempotency begin failed", append(m.attrs(r, key), slog.Any("error", err))...)
		m.onError(w, r, apperr.Wrap(apperr.CodeServiceUnavailable, err))
		return
	}
	switch outcome {
	case Replay:
		replay(w, stored)
	case InProgress:
		m.onError(w, r, apperr.New(apperr.CodeIdempotencyInProg, "A request with this Idempotency-Key is still being processed."))
	case Started:
		m.run(w, r, sub.UserID, key, attempt)
	default: // Reused
		m.onError(w, r, apperr.New(apperr.CodeIdempotencyReused, "This Idempotency-Key was already used with a different request."))
	}
}

func (m *mw) begin(ctx context.Context, userID, key, method, route string, hash [32]byte) (Outcome, Attempt, Response, error) {
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	return m.store.Begin(ctx, userID, key, method, route, hash, time.Now())
}

// run serves the request as the owner of the key and records the outcome.
func (m *mw) run(w http.ResponseWriter, r *http.Request, userID, key string, attempt Attempt) {
	rec := &recorder{ResponseWriter: w}
	finished := false
	defer func() {
		if !finished { // the handler panicked: free the key, the panic keeps going up
			m.abandon(r, userID, key, attempt)
		}
	}()
	m.next.ServeHTTP(rec, r)
	finished = true
	// A handler that wrote nothing after the request context ended gave up (timeout, client
	// cancel): that is not a result. Writing nothing with a live context is an implicit 200.
	if !cacheable(rec.status()) || (rec.code == 0 && r.Context().Err() != nil) {
		m.abandon(r, userID, key, attempt)
		return
	}
	m.complete(r, userID, key, attempt, rec)
}

// cacheable is the caching policy: 2xx and 4xx outcomes are stored, except 401, 403, 408,
// 425 and 429, which describe the caller's session or capacity rather than the outcome of
// this request. Those, redirects and every 5xx are abandoned so that a retry runs again.
func cacheable(status int) bool {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests:
		return false
	}
	return (status >= 200 && status < 300) || (status >= 400 && status < 500)
}

// storeCtx outlives the request: a client that went away must not leave the
// row in progress.
func storeCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(r.Context()), storeTimeout)
}

func (m *mw) abandon(r *http.Request, userID, key string, attempt Attempt) {
	ctx, cancel := storeCtx(r)
	defer cancel()
	if err := m.store.Abandon(ctx, userID, key, attempt); err != nil {
		m.logger.WarnContext(r.Context(), "idempotency abandon failed", append(m.attrs(r, key), slog.Any("error", err))...)
	}
}

func (m *mw) complete(r *http.Request, userID, key string, attempt Attempt, rec *recorder) {
	body := rec.storableBody()
	if body == nil && rec.written > 0 {
		m.logger.WarnContext(r.Context(), "idempotency response body not stored: not a JSON object within the size cap",
			append(m.attrs(r, key), slog.Int("bytes", rec.written))...)
	}
	ctx, cancel := storeCtx(r)
	defer cancel()
	if err := m.store.Complete(ctx, userID, key, attempt, Response{Status: rec.status(), ContentType: rec.contentType, Headers: rec.headers, Body: body}, time.Now()); err != nil {
		m.logger.WarnContext(r.Context(), "idempotency complete failed", append(m.attrs(r, key), slog.Any("error", err))...)
	}
}

// attrs never carries more than a short prefix of the key.
func (m *mw) attrs(r *http.Request, key string) []any {
	return []any{
		slog.String("request_id", reqID(r)),
		slog.String("key_prefix", key[:min(logPrefixLen, len(key))]),
	}
}

func reqID(r *http.Request) string {
	id := chimw.GetReqID(r.Context())
	return id[:min(maxReqIDLen, len(id))]
}

// replay serves the stored response. Without a stored body it sends no Content-Type (a 2xx
// claiming application/json with an empty body makes res.json() throw in a client) and, when a
// content type was recorded (so a body existed and was lost), HeaderBodyOmitted.
func replay(w http.ResponseWriter, s Response) {
	switch {
	case len(s.Body) > 0:
		if s.ContentType != "" {
			w.Header().Set("Content-Type", s.ContentType)
		}
	case s.ContentType != "":
		w.Header().Set(HeaderBodyOmitted, "true")
	}
	for _, name := range replayedHeaders {
		if v, ok := s.Headers[name]; ok {
			w.Header().Set(name, v)
		}
	}
	w.Header().Set(HeaderReplayed, "true")
	w.WriteHeader(s.Status)
	_, _ = w.Write(s.Body)
}

// hashRequest identifies a request: method, route template, the concrete escaped path
// (not the query) and the body, NUL-separated. The concrete path is what makes the same
// key and body posted under another parent id a different request.
func hashRequest(method, route, path string, body []byte) [32]byte {
	h := sha256.New()
	h.Write([]byte(method))
	h.Write([]byte{0})
	h.Write([]byte(route))
	h.Write([]byte{0})
	h.Write([]byte(path))
	h.Write([]byte{0})
	h.Write(body)
	var sum [32]byte
	h.Sum(sum[:0])
	return sum
}

// checkKey re-validates what the contract validator already enforced: 16..128
// printable ASCII characters.
func checkKey(key string) *apperr.Error {
	violation := func(code string) *apperr.Error {
		return apperr.New(apperr.CodeValidation, "The Idempotency-Key header is invalid.").
			WithViolations(apperr.Violation{Field: HeaderKey, Code: code})
	}
	switch {
	case len(key) < minKeyLen:
		return violation("too_short")
	case len(key) > maxKeyLen:
		return violation("too_long")
	}
	for i := range len(key) {
		if key[i] <= ' ' || key[i] > '~' {
			return violation("invalid_format")
		}
	}
	return nil
}

func bodyError(err error) *apperr.Error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return apperr.New(apperr.CodePayloadTooLarge, "The request body is too large.")
	}
	return apperr.New(apperr.CodeMalformedRequest, "The request body could not be read.")
}

// recorder tees the response: everything reaches the client unchanged, while
// the final status, content type, allowlisted headers and (bounded) body are kept for the store.
type recorder struct {
	http.ResponseWriter
	code        int
	contentType string
	headers     map[string]string
	buf         bytes.Buffer
	written     int
}

// Unwrap lets http.NewResponseController reach Flush, SetWriteDeadline and friends.
func (c *recorder) Unwrap() http.ResponseWriter { return c.ResponseWriter }

// FlushError is what http.NewResponseController calls. Flushing before any WriteHeader makes
// net/http send an implicit 200: that is the status the client saw, so it is the recorded one,
// and a later WriteHeader is ignored (net/http would only log it as superfluous).
func (c *recorder) FlushError() error {
	if c.code == 0 {
		c.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(c.ResponseWriter).Flush()
}

// Flush implements http.Flusher for handlers that assert it directly.
func (c *recorder) Flush() { _ = c.FlushError() }

func (c *recorder) WriteHeader(code int) {
	// 1xx (other than 101) are informational: they pass through and do not pin the final status.
	if code >= http.StatusContinue && code < http.StatusOK && code != http.StatusSwitchingProtocols {
		c.ResponseWriter.WriteHeader(code)
		return
	}
	if c.code != 0 {
		return // the response line is already out: net/http ignores a second WriteHeader too
	}
	c.code = code
	h := c.Header()
	c.contentType = h.Get("Content-Type")
	for _, name := range replayedHeaders {
		if v := h.Get(name); v != "" {
			if c.headers == nil {
				c.headers = map[string]string{}
			}
			c.headers[name] = v
		}
	}
	c.ResponseWriter.WriteHeader(code)
}

func (c *recorder) Write(p []byte) (int, error) {
	if c.code == 0 {
		c.WriteHeader(http.StatusOK)
	}
	if c.written <= MaxStoredBody { // one byte past the cap is enough to know it overflowed
		c.buf.Write(p[:min(len(p), MaxStoredBody+1-c.written)])
	}
	n, err := c.ResponseWriter.Write(p)
	c.written += n
	return n, err
}

func (c *recorder) status() int {
	if c.code == 0 {
		return http.StatusOK
	}
	return c.code
}

// storableBody is the body when it is a JSON object within the cap of the
// response_body column, otherwise nil.
func (c *recorder) storableBody() []byte {
	b := c.buf.Bytes()
	if c.written > MaxStoredBody || !json.Valid(b) || !isObject(b) {
		return nil
	}
	return bytes.Clone(b)
}
