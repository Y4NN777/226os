package transport

import (
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
	view, err := res.svc.ResolveSession(r.Context(), c.Value)
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
