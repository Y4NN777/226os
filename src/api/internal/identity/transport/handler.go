// Package transport is the HTTP edge of the S1 identity operations: it decodes
// contract bodies, sets the profile-correct cookies and maps service results
// to responses (ADR-0003, ADR-0004). Policy lives in the service; this package
// only encodes.
package transport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// preSessionMaxAge bounds the anonymous CSRF cookie (ADR-0003: short-lived).
const preSessionMaxAge = 24 * 60 * 60

// Service is the identity use-case surface the transport needs.
// *service.Service satisfies it; tests may stub single methods.
type Service interface {
	ResolveSession(ctx context.Context, plain string) (service.SessionView, error)
	PreSessionCSRF(secret string) (string, bool)
	AnonymousCSRF() (secret, token string, err error)
	SignUp(ctx context.Context, in service.SignUpInput) error
	SignIn(ctx context.Context, in service.SignInInput) (service.SignInResult, error)
	SignOut(ctx context.Context, plain string) error
	VerifyEmail(ctx context.Context, plain string) error
	RequestPasswordReset(ctx context.Context, email, ip string) error
	ConfirmPasswordReset(ctx context.Context, plain, newPassword, ip string) error
	GetMe(ctx context.Context, userID string) (service.MeView, error)
	UpdateMyProfile(ctx context.Context, userID string, w domain.ProfileWrite) (domain.Profile, error)
	GetProfile(ctx context.Context, username, callerUserID string) (domain.PublicProfile, error)
}

// Handler serves the eight auth operations of S1.
type Handler struct {
	svc    Service
	secure bool // deployed profiles: __Host- cookie names + Secure flag
	logger *slog.Logger
}

// New returns a handler. secure selects the deployed cookie profile
// (__Host-226os_session/__Host-226os_csrf, Secure) versus local-dev
// (226os_session/226os_csrf, plain HTTP). logger may be nil.
func New(svc Service, secure bool, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Handler{svc: svc, secure: secure, logger: logger}
}

// sessionCookieName and preSessionCookieName select the ADR-0003 cookie names
// for the deployment profile: __Host- prefixed + Secure when deployed.
func sessionCookieName(secure bool) string {
	if secure {
		return "__Host-226os_session"
	}
	return "226os_session"
}

func preSessionCookieName(secure bool) string {
	if secure {
		return "__Host-226os_csrf"
	}
	return "226os_csrf"
}

func (h *Handler) sessionCookie() string    { return sessionCookieName(h.secure) }
func (h *Handler) preSessionCookie() string { return preSessionCookieName(h.secure) }

// ---------- helpers ----------

func (h *Handler) sessionPlain(r *http.Request) string {
	c, err := r.Cookie(h.sessionCookie())
	if err != nil {
		return ""
	}
	return c.Value
}

func (h *Handler) preSessionSecret(r *http.Request) string {
	c, err := r.Cookie(h.preSessionCookie())
	if err != nil {
		return ""
	}
	return c.Value
}

func (h *Handler) setCookie(w http.ResponseWriter, name, value string, maxAge int, expires time.Time) {
	http.SetCookie(w, &http.Cookie{ //nolint:gosec // G124: Secure follows the deployment profile; only local-dev http omits it, and __Host- names require it
		Name: name, Value: value, Path: "/", MaxAge: maxAge, Expires: expires,
		HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{ //nolint:gosec // G124: Secure follows the deployment profile; only local-dev http omits it, and __Host- names require it
		Name: h.sessionCookie(), Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode,
	})
}

// clientIP returns the rate-limit identity: the right-most X-Forwarded-For
// entry. Traefik is the only ingress and appends the peer address it observed
// at the end of the header, so earlier entries are client-supplied and
// spoofable; the last one is the address Traefik actually saw. Falls back to
// the socket peer when the header is absent.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.LastIndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[i+1:])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func accountJSON(a domain.Account) (apigen.Account, error) {
	id, err := parseUUID(a.ID)
	if err != nil {
		return apigen.Account{}, err
	}
	return apigen.Account{
		Id: id, Email: openapi_types.Email(a.Email), EmailVerified: a.EmailVerified,
		Username: apigen.Username(a.Username), DisplayName: a.DisplayName,
		Roles: []apigen.AccountRoles{apigen.AccountRoles(a.Role)}, CreatedAt: a.CreatedAt,
	}, nil
}

func parseUUID(s string) (apigen.Uuid, error) {
	return uuid.Parse(s)
}

func sessionStateJSON(view service.SessionView) (apigen.SessionState, error) {
	acct, err := accountJSON(view.Account)
	if err != nil {
		return apigen.SessionState{}, err
	}
	expires := view.Expires
	return apigen.SessionState{
		Authenticated: true, CsrfToken: view.CSRF, ExpiresAt: &expires, User: &acct,
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeErr renders err as an RFC 9457 problem; httpx.WriteError already maps
// non-apperr causes to a redacted 500.
func writeErr(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	httpx.WriteError(logger)(w, r, err)
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// ---------- operations ----------

// GetCsrfToken serves GET /auth/csrf: the session-bound token for a live
// session, else the pre-session token (issuing the cookie when absent).
func (h *Handler) GetCsrfToken(w http.ResponseWriter, r *http.Request) {
	if view, err := resolveSession(r, h.svc, h.sessionPlain(r)); err == nil {
		writeJSON(w, http.StatusOK, apigen.CsrfToken{CsrfToken: view.CSRF})
		return
	} else if !isUnauthenticated(err) {
		writeErr(h.logger, w, r, err)
		return
	}
	if token, ok := h.svc.PreSessionCSRF(h.preSessionSecret(r)); ok {
		writeJSON(w, http.StatusOK, apigen.CsrfToken{CsrfToken: token})
		return
	}
	secret, token, err := h.svc.AnonymousCSRF()
	if err != nil {
		writeErr(h.logger, w, r, apperr.Wrap(apperr.CodeInternal, err))
		return
	}
	h.setCookie(w, h.preSessionCookie(), secret, preSessionMaxAge, time.Now().Add(preSessionMaxAge*time.Second))
	writeJSON(w, http.StatusOK, apigen.CsrfToken{CsrfToken: token})
}

// SignUp serves POST /auth/sign-up: always 202, never reveals existence.
func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var body apigen.SignUpRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	if body.Password == nil {
		writeErr(h.logger, w, r, missingField("password"))
		return
	}
	err := h.svc.SignUp(r.Context(), service.SignUpInput{
		Email: string(body.Email), Password: *body.Password, Username: string(body.Username),
		DisplayName: derefStr(body.DisplayName), UIlang: langOf(body.UiLang), IP: clientIP(r),
	})
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, apigen.Accepted{Status: apigen.AcceptedStatusAccepted})
}

// SignIn serves POST /auth/sign-in: generic INVALID_CREDENTIALS, rotates the
// session carried by the request cookie, sets the new session cookie.
func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var body apigen.SignInRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	if body.Password == nil {
		writeErr(h.logger, w, r, missingField("password"))
		return
	}
	res, err := h.svc.SignIn(r.Context(), service.SignInInput{
		Email: string(body.Email), Password: *body.Password,
		IP: clientIP(r), CurrentSessionID: h.currentSessionID(r),
	})
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	acct, err := accountJSON(res.Account)
	if err != nil {
		writeErr(h.logger, w, r, apperr.Wrap(apperr.CodeInternal, err))
		return
	}
	h.setCookie(w, h.sessionCookie(), res.Cookie.Value, res.Cookie.MaxAge, res.Cookie.Expires)
	writeJSON(w, http.StatusOK, apigen.SessionState{
		Authenticated: true, CsrfToken: res.Cookie.CSRF,
		ExpiresAt: &res.Cookie.Expires, User: &acct,
	})
}

// currentSessionID resolves the session cookie to its id for rotation. An
// invalid or absent cookie yields "" and rotation is skipped.
func (h *Handler) currentSessionID(r *http.Request) string {
	view, err := resolveSession(r, h.svc, h.sessionPlain(r))
	if err != nil {
		return ""
	}
	return view.SessionID
}

// SignOut serves POST /auth/sign-out: member only, revokes the caller's own
// session and clears the cookie.
func (h *Handler) SignOut(w http.ResponseWriter, r *http.Request) {
	if subject.From(r.Context()).Anonymous() {
		writeErr(h.logger, w, r, apperr.New(apperr.CodeUnauthenticated, "Authentication required."))
		return
	}
	if err := h.svc.SignOut(r.Context(), h.sessionPlain(r)); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// GetSession serves GET /auth/session: never 401. Authenticated callers get
// their account and a fresh CSRF token; anonymous callers get the pre-session
// token (and the cookie when it is missing or malformed).
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	view, err := resolveSession(r, h.svc, h.sessionPlain(r))
	if err == nil {
		state, err := sessionStateJSON(view)
		if err != nil {
			writeErr(h.logger, w, r, apperr.Wrap(apperr.CodeInternal, err))
			return
		}
		writeJSON(w, http.StatusOK, state)
		return
	}
	if !isUnauthenticated(err) {
		writeErr(h.logger, w, r, err)
		return
	}
	token, ok := h.svc.PreSessionCSRF(h.preSessionSecret(r))
	if !ok {
		var secret string
		secret, token, err = h.svc.AnonymousCSRF()
		if err != nil {
			writeErr(h.logger, w, r, apperr.Wrap(apperr.CodeInternal, err))
			return
		}
		h.setCookie(w, h.preSessionCookie(), secret, preSessionMaxAge, time.Now().Add(preSessionMaxAge*time.Second))
	}
	writeJSON(w, http.StatusOK, apigen.SessionState{Authenticated: false, CsrfToken: token})
}

// VerifyEmail serves POST /auth/verify-email: single-use 24 h token.
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var body apigen.TokenRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	if err := h.svc.VerifyEmail(r.Context(), body.Token); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RequestPasswordReset serves POST /auth/password-reset/request: always 202.
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var body apigen.PasswordResetRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	if err := h.svc.RequestPasswordReset(r.Context(), string(body.Email), clientIP(r)); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, apigen.Accepted{Status: apigen.AcceptedStatusAccepted})
}

// ConfirmPasswordReset serves POST /auth/password-reset/confirm: single-use
// 30 min token; every session of the account is revoked, so the session
// cookie is cleared here as well.
func (h *Handler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var body apigen.PasswordResetConfirm
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	if body.NewPassword == nil {
		writeErr(h.logger, w, r, missingField("new_password"))
		return
	}
	if err := h.svc.ConfirmPasswordReset(r.Context(), body.Token, *body.NewPassword, clientIP(r)); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func isUnauthenticated(err error) bool {
	var ae *apperr.Error
	return errors.As(err, &ae) && ae.Code == apperr.CodeUnauthenticated
}

func missingField(field string) error {
	return apperr.New(apperr.CodeValidation, "A required field is missing.").
		WithViolations(apperr.Violation{Field: field, Code: "required"})
}

func langOf(v *apigen.Lang) string {
	if v == nil {
		return ""
	}
	return string(*v)
}
