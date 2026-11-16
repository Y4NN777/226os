package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

func csrfReq(method, path string, headers map[string]string, cookies ...*http.Cookie) *http.Request {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
}

func TestCSRFSafeMethodsPass(t *testing.T) {
	_, svc, _, _ := testStack(t, false)
	mw := CSRFMiddleware(svc, false, nil)
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		rec := httptest.NewRecorder()
		mw(okHandler()).ServeHTTP(rec, csrfReq(m, "/api/v1/x", nil))
		if rec.Code != http.StatusTeapot {
			t.Fatalf("%s: %d", m, rec.Code)
		}
	}
}

func TestCSRFMissingTokenFails(t *testing.T) {
	_, svc, _, _ := testStack(t, false)
	mw := CSRFMiddleware(svc, false, nil)
	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, csrfReq(http.MethodPost, "/api/v1/auth/sign-up", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("missing token: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "CSRF_FAILED") {
		t.Fatalf("body %s", rec.Body.String())
	}
}

// preSessionFixture obtains a secret cookie + derived token through the real
// GET /auth/csrf handler.
func preSessionFixture(t *testing.T, h *Handler) (*http.Cookie, string) {
	t.Helper()
	rec := call(t, h.GetCsrfToken, doReq(t, http.MethodGet, "/api/v1/auth/csrf", nil))
	var secret *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_csrf" {
			secret = c
		}
	}
	if secret == nil {
		t.Fatal("no pre-session cookie")
	}
	var body struct {
		CsrfToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.CsrfToken == "" {
		t.Fatalf("no token: %s", rec.Body.String())
	}
	return secret, body.CsrfToken
}

func TestCSRFPreSessionFlow(t *testing.T) {
	h, svc, _, _ := testStack(t, false)
	mw := CSRFMiddleware(svc, false, nil)
	secret, token := preSessionFixture(t, h)

	// POST without token -> 403.
	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, csrfReq(http.MethodPost, "/api/v1/auth/sign-up", nil, secret))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("no header: %d", rec.Code)
	}
	// POST with wrong token -> 403.
	rec = httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, csrfReq(http.MethodPost, "/api/v1/auth/sign-up",
		map[string]string{"X-Csrf-Token": "wrong"}, secret))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("wrong token: %d", rec.Code)
	}
	// POST with the derived token -> pass.
	rec = httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, csrfReq(http.MethodPost, "/api/v1/auth/sign-up",
		map[string]string{"X-Csrf-Token": token}, secret))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("valid token: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCSRFSessionBound(t *testing.T) {
	h, svc, _, _ := testStack(t, false)
	plain, csrf := signupAndSignIn(t, h, "csrf@example.com")
	mw := CSRFMiddleware(svc, false, nil)
	sess := &http.Cookie{Name: "226os_session", Value: plain}

	// Session cookie + session-bound token -> pass.
	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, csrfReq(http.MethodPost, "/api/v1/auth/sign-out",
		map[string]string{"X-Csrf-Token": csrf}, sess))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("session token: %d %s", rec.Code, rec.Body.String())
	}
	// A bogus token fails even with a live session.
	rec = httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, csrfReq(http.MethodPost, "/api/v1/auth/sign-out",
		map[string]string{"X-Csrf-Token": "bogus"}, sess))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("bogus session token: %d", rec.Code)
	}
	// The pre-session token must not satisfy a session-bound check.
	_, preToken := preSessionFixture(t, h)
	rec = httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, csrfReq(http.MethodPost, "/api/v1/auth/sign-out",
		map[string]string{"X-Csrf-Token": preToken}, sess))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("pre-session token on session path: %d", rec.Code)
	}
}

func TestCSRFOriginAndFetchSite(t *testing.T) {
	h, svc, _, _ := testStack(t, false)
	mw := CSRFMiddleware(svc, false, nil)
	secret, token := preSessionFixture(t, h)

	// Hostile Origin -> 403 even with a valid token.
	req := &http.Request{Method: http.MethodPost, URL: mustURL("/api/v1/auth/sign-up"),
		Header: http.Header{"Origin": {"https://evil.example"}, "X-Csrf-Token": {token}},
		Host:   "226os.tech"}
	req.AddCookie(secret)
	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin: %d", rec.Code)
	}

	// Sec-Fetch-Site: cross-site -> 403.
	req = &http.Request{Method: http.MethodPost, URL: mustURL("/api/v1/auth/sign-up"),
		Header: http.Header{"Sec-Fetch-Site": {"cross-site"}, "X-Csrf-Token": {token}},
		Host:   "226os.tech"}
	req.AddCookie(secret)
	rec = httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-site fetch: %d", rec.Code)
	}

	// Same-origin Origin + same-origin fetch site + valid token -> pass.
	req = &http.Request{Method: http.MethodPost, URL: mustURL("/api/v1/auth/sign-up"),
		Header: http.Header{
			"Origin":         {"https://226os.tech"},
			"Sec-Fetch-Site": {"same-origin"},
			"X-Csrf-Token":   {token},
		},
		Host: "226os.tech"}
	req.AddCookie(secret)
	rec = httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("same-origin: %d %s", rec.Code, rec.Body.String())
	}
}

func mustURL(path string) *url.URL {
	u, err := url.Parse("http://226os.tech" + path)
	if err != nil {
		panic(err)
	}
	return u
}

func TestResolverAnonymousWithoutCookie(t *testing.T) {
	_, svc, _, _ := testStack(t, false)
	res := NewResolver(svc, false)
	s, err := res.Resolve(doReq(t, http.MethodGet, "/api/v1/x", nil))
	if err != nil || !s.Anonymous() {
		t.Fatalf("anonymous: %+v %v", s, err)
	}
}

func TestResolverBuildsSubjectFromSession(t *testing.T) {
	h, svc, _, _ := testStack(t, false)
	plain, _ := signupAndSignIn(t, h, "sub@example.com")
	res := NewResolver(svc, false)
	req := doReq(t, http.MethodGet, "/api/v1/x", nil,
		&http.Cookie{Name: "226os_session", Value: plain})
	s, err := res.Resolve(req)
	if err != nil {
		t.Fatal(err)
	}
	if s.Anonymous() || s.UserID == "" || s.EmailVerified {
		t.Fatalf("subject %+v", s)
	}
	if len(s.Roles) != 1 || s.Roles[0] != "member" {
		t.Fatalf("roles %v", s.Roles)
	}
}

func TestResolverUnknownCookieIsAnonymous(t *testing.T) {
	_, svc, _, _ := testStack(t, false)
	res := NewResolver(svc, false)
	req := doReq(t, http.MethodGet, "/api/v1/x", nil,
		&http.Cookie{Name: "226os_session", Value: "bogus-token"})
	s, err := res.Resolve(req)
	if err != nil || !s.Anonymous() {
		t.Fatalf("bogus: %+v %v", s, err)
	}
}

// A request resolves its session once: the subject middleware, the CSRF
// middleware and the handler share one lookup (hot path, one query).
func TestSessionIsResolvedOncePerRequest(t *testing.T) {
	h, svc, st, _ := testStack(t, false)
	plain, csrf := signupAndSignIn(t, h, "once@example.com")
	before := st.sessionLookups
	chain := subject.Middleware(NewResolver(svc, false), nil)(
		CSRFMiddleware(svc, false, nil)(http.HandlerFunc(h.SignOut)))
	req := doReq(t, http.MethodPost, "/api/v1/auth/sign-out",
		map[string]string{"X-CSRF-Token": csrf, "Origin": "http://example.com", "Sec-Fetch-Site": "same-origin"},
		&http.Cookie{Name: sessionCookieName(false), Value: plain})
	req.Host = "example.com"
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("sign-out %d %s", rec.Code, rec.Body.String())
	}
	// One shared resolution plus SignOut's own lookup of the row to revoke.
	if got := st.sessionLookups - before; got != 2 {
		t.Fatalf("session lookups per request = %d, want 2", got)
	}
}
