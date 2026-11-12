// Package subject carries the authenticated caller (or anonymity) through a
// request. The Subject always comes from a Resolver, never from a request
// header, query parameter, cookie value or body supplied by the client
// (docs/spec-kit/AUTHORIZATION_MATRIX.md, ADR-0003).
//
// S1 (identity) provides the session-backed Resolver: sessions are opaque
// server-side tokens looked up by hash in PostgreSQL (ADR-0003). Until then
// AnonymousResolver is the default. This package holds no session logic and
// no role semantics beyond the Roles field.
package subject

import (
	"context"
	"net/http"
)

// Subject is the caller of a request. An empty UserID means anonymous.
type Subject struct {
	UserID        string
	EmailVerified bool
	Roles         []string
}

// Anonymous reports whether the caller is not authenticated.
func (s Subject) Anonymous() bool { return s.UserID == "" }

// Resolver derives the Subject from the request (in S1: from the session
// cookie). It must never trust client-supplied identity values.
type Resolver interface {
	Resolve(r *http.Request) (Subject, error)
}

// AnonymousResolver is the default Resolver: every caller is anonymous.
type AnonymousResolver struct{}

// Resolve always returns an anonymous Subject.
func (AnonymousResolver) Resolve(*http.Request) (Subject, error) { return Subject{}, nil }

type ctxKey struct{}

// With returns a context carrying s.
func With(ctx context.Context, s Subject) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

// From returns the Subject stored in ctx, or the anonymous Subject if none.
func From(ctx context.Context) Subject {
	s, _ := ctx.Value(ctxKey{}).(Subject)
	return s
}

// fallbackProblem is the static RFC 9457 body of the nil-onError fallback: it carries no
// request-supplied data and no error text. The app always passes httpx.WriteError instead.
const fallbackProblem = `{"type":"urn:226os:problem:INTERNAL_ERROR","title":"Internal server error","status":500,"code":"INTERNAL_ERROR"}` + "\n"

// Middleware resolves the Subject with res (nil means AnonymousResolver) and
// stores it in the request context. On a resolver error it calls onError (a
// static problem+json 500 when onError is nil) and does not call the next handler.
func Middleware(res Resolver, onError func(w http.ResponseWriter, r *http.Request, err error)) func(http.Handler) http.Handler {
	if res == nil {
		res = AnonymousResolver{}
	}
	if onError == nil {
		onError = func(w http.ResponseWriter, _ *http.Request, _ error) {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fallbackProblem))
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, err := res.Resolve(r)
			if err != nil {
				onError(w, r, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(With(r.Context(), s)))
		})
	}
}
