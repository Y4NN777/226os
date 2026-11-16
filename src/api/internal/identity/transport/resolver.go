package transport

import (
	"context"
	"net/http"

	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// Resolver is the session-backed subject.Resolver of ADR-0003: the Subject
// comes only from the opaque session cookie looked up by hash in PostgreSQL,
// never from a header, query parameter or body. An absent, unknown, expired
// or revoked session resolves to the anonymous Subject (never an error), so
// anonymous routes keep working; a store failure propagates as a 500.
type Resolver struct {
	svc        *service.Service
	cookieName string
}

// NewResolver returns the production Resolver. secure selects the deployed
// cookie profile, like transport.New.
func NewResolver(svc *service.Service, secure bool) subject.Resolver {
	return &Resolver{svc: svc, cookieName: sessionCookieName(secure)}
}

// Resolve implements subject.Resolver.
func (res *Resolver) Resolve(r *http.Request) (subject.Subject, error) {
	c, err := r.Cookie(res.cookieName)
	if err != nil || c.Value == "" {
		return subject.Subject{}, nil
	}
	view, err := resolveSession(r, res.svc, c.Value)
	if err != nil {
		if isUnauthenticated(err) {
			return subject.Subject{}, nil
		}
		return subject.Subject{}, err
	}
	return subject.Subject{
		UserID:        view.Account.ID,
		EmailVerified: view.Account.EmailVerified,
		Roles:         []string{view.Account.Role},
	}, nil
}

// sessionKey keys the per-request memo by the cookie plaintext.
type sessionKey string

// sessionResolver is what resolveSession needs from the service.
type sessionResolver interface {
	ResolveSession(ctx context.Context, plain string) (service.SessionView, error)
}

// resolveSession resolves the session cookie once per request: the subject
// middleware, the CSRF middleware and the handler share the result through the
// request's memo (without one, it simply resolves each time).
func resolveSession(r *http.Request, svc sessionResolver, plain string) (service.SessionView, error) {
	v, err := subject.MemoFrom(r.Context()).Do(sessionKey(plain), func() (any, error) {
		return svc.ResolveSession(r.Context(), plain)
	})
	view, _ := v.(service.SessionView)
	return view, err
}
