package transport

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

// unsafeMethods are the HTTP methods a mutation can arrive on (ADR-0003:
// the synchronizer token is required on every unsafe method).
var unsafeMethods = map[string]bool{
	http.MethodPost: true, http.MethodPut: true,
	http.MethodPatch: true, http.MethodDelete: true,
}

// CSRFMiddleware enforces the ADR-0003 synchronizer-token check on every
// unsafe method, plus the Origin / Sec-Fetch-Site same-origin check. It is an
// apigen.MiddlewareFunc: it runs after subject resolution and before
// idempotency, so a CSRF failure never claims an idempotency key.
//
// A live session cookie binds the check to the session token hash; anything
// else falls back to the pre-session secret cookie used by the anonymous
// auth endpoints. Failure is 403 CSRF_FAILED; an internal lookup error is 500.
func CSRFMiddleware(svc *service.Service, secure bool, logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	sessionName := sessionCookieName(secure)
	preSessionName := preSessionCookieName(secure)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !unsafeMethods[r.Method] {
				next.ServeHTTP(w, r)
				return
			}
			if !sameOriginOK(r) {
				writeCSRFFailed(logger, w, r)
				return
			}
			header := r.Header.Get("X-CSRF-Token")
			plain := cookieValue(r, sessionName)
			view, err := svc.ResolveSession(r.Context(), plain)
			switch {
			case err == nil:
				if svc.CheckCSRF(&view.TokenHash, "", header) {
					next.ServeHTTP(w, r)
					return
				}
			case isUnauthenticated(err):
				if svc.CheckCSRF(nil, cookieValue(r, preSessionName), header) {
					next.ServeHTTP(w, r)
					return
				}
			default:
				writeErr(logger, w, r, err)
				return
			}
			writeCSRFFailed(logger, w, r)
		})
	}
}

// sameOriginOK applies the Origin / Sec-Fetch-Site check of ADR-0003. Absent
// headers pass (non-browser clients still need the synchronizer token);
// a present Origin must match the request Host, and a present
// Sec-Fetch-Site must be same-origin or none.
func sameOriginOK(r *http.Request) bool {
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || !strings.EqualFold(u.Host, r.Host) {
			return false
		}
	}
	switch sfs := r.Header.Get("Sec-Fetch-Site"); sfs {
	case "", "same-origin", "none":
		return true
	default:
		return false
	}
}

func cookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

func writeCSRFFailed(logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	httpx.WriteError(logger)(w, r, apperr.New(apperr.CodeCSRFFailed, "CSRF token validation failed."))
}
