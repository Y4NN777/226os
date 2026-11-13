package transport

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// anonFail stubs only AnonymousCSRF so its entropy-failure branch is
// reachable; every other method delegates to the embedded Service.
type anonFail struct {
	Service
	err error
}

func (s anonFail) AnonymousCSRF() (string, string, error) { return "", "", s.err }

func TestGetCsrfTokenFlows(t *testing.T) {
	h, _, _, _ := testStack(t, false)

	// Anonymous: issues the pre-session cookie and a derived token.
	rec := call(t, h.GetCsrfToken, doReq(t, http.MethodGet, "/api/v1/auth/csrf", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("anonymous: %d %s", rec.Code, rec.Body.String())
	}
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
		t.Fatalf("body %s", rec.Body.String())
	}

	// Existing valid pre-session secret: same token, no new cookie.
	rec = call(t, h.GetCsrfToken, doReq(t, http.MethodGet, "/api/v1/auth/csrf", nil, secret))
	if rec.Code != http.StatusOK {
		t.Fatalf("pre-session: %d", rec.Code)
	}
	var body2 struct {
		CsrfToken string `json:"csrf_token"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body2)
	if body2.CsrfToken != body.CsrfToken {
		t.Fatal("pre-session token not stable")
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("cookie re-issued for a valid secret")
	}

	// Live session: session-bound token wins.
	plain, sessCSRF := signupAndSignIn(t, h, "csrfflow@example.com")
	rec = call(t, h.GetCsrfToken, doReq(t, http.MethodGet, "/api/v1/auth/csrf", nil,
		&http.Cookie{Name: "226os_session", Value: plain}, secret))
	if rec.Code != http.StatusOK {
		t.Fatalf("session: %d", rec.Code)
	}
	var body3 struct {
		CsrfToken string `json:"csrf_token"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body3)
	if body3.CsrfToken != sessCSRF {
		t.Fatal("session CSRF token mismatch")
	}
}

func TestGetCsrfTokenInternalError(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	st.failSessionLookup = true
	rec := call(t, h.GetCsrfToken, doReq(t, http.MethodGet, "/api/v1/auth/csrf", nil,
		&http.Cookie{Name: "226os_session", Value: validTokenPlain()}))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("store error: %d %s", rec.Code, rec.Body.String())
	}
}

func TestGetCsrfTokenAnonymousCSRFFailure(t *testing.T) {
	_, svc, _, _ := testStack(t, false)
	h := New(anonFail{Service: svc, err: errors.New("no entropy")}, false, nil)
	rec := call(t, h.GetCsrfToken, doReq(t, http.MethodGet, "/api/v1/auth/csrf", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("entropy: %d %s", rec.Code, rec.Body.String())
	}
}

func TestSignUpMissingPasswordAndServiceError(t *testing.T) {
	h, _, st, _ := testStack(t, false)

	// Missing password field -> 422 validation.
	rec := doJSON(t, h.SignUp, "/api/v1/auth/sign-up",
		`{"email":"np@example.com","username":"npuser"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing password: %d %s", rec.Code, rec.Body.String())
	}

	// Service failure (rate-limit store down) -> problem response, not 202.
	st.failHits = true
	rec = doJSON(t, h.SignUp, "/api/v1/auth/sign-up",
		`{"email":"np2@example.com","password":"correct-horse-battery","username":"npuser2","ui_lang":"en"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("store error: %d %s", rec.Code, rec.Body.String())
	}
}

func TestSignInDecodeErrorAndRotation(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	rec := doJSON(t, h.SignIn, "/api/v1/auth/sign-in", `{"email":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed: %d", rec.Code)
	}
	rec = doJSON(t, h.SignIn, "/api/v1/auth/sign-in", `{"email":"a@b.co"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing password: %d", rec.Code)
	}

	// Second sign-in carrying the session cookie rotates it away.
	plain, _ := signupAndSignIn(t, h, "rot@example.com")
	rec = doJSONWithCookie(t, h.SignIn, "/api/v1/auth/sign-in",
		`{"email":"rot@example.com","password":"correct-horse-battery"}`,
		&http.Cookie{Name: "226os_session", Value: plain})
	if rec.Code != http.StatusOK {
		t.Fatalf("rotating sign-in: %d %s", rec.Code, rec.Body.String())
	}
	rec = call(t, h.GetSession, doReq(t, http.MethodGet, "/api/v1/auth/session", nil,
		&http.Cookie{Name: "226os_session", Value: plain}))
	var state struct {
		Authenticated bool `json:"authenticated"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &state)
	if state.Authenticated {
		t.Fatal("rotated session still valid")
	}
}

func doJSONWithCookie(t *testing.T, fn op, path, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	fn(rec, req)
	return rec
}

func TestSignInAccountJSONError(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	// Corrupt the stored account id: sign-in succeeds but cannot be encoded.
	signupAndSignIn(t, h, "badid@example.com")
	for id, a := range st.accounts {
		if a.Email == "badid@example.com" {
			a.ID = "not-a-uuid"
			st.accounts[id] = a
		}
	}
	rec := doJSON(t, h.SignIn, "/api/v1/auth/sign-in",
		`{"email":"badid@example.com","password":"correct-horse-battery"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("bad account id: %d %s", rec.Code, rec.Body.String())
	}
}

func TestSignOutServiceError(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	plain, _ := signupAndSignIn(t, h, "outerr@example.com")
	st.failRevoke = true
	req := doReq(t, http.MethodPost, "/api/v1/auth/sign-out", nil,
		&http.Cookie{Name: "226os_session", Value: plain})
	req = req.WithContext(withSubject(req, "member"))
	rec := call(t, h.SignOut, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("revoke error: %d %s", rec.Code, rec.Body.String())
	}
}

func TestGetSessionErrorPaths(t *testing.T) {
	h, _, st, _ := testStack(t, false)

	// Store failure on session lookup -> 500, never anonymous fallback.
	st.failSessionLookup = true
	rec := call(t, h.GetSession, doReq(t, http.MethodGet, "/api/v1/auth/session", nil,
		&http.Cookie{Name: "226os_session", Value: validTokenPlain()}))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("lookup error: %d", rec.Code)
	}
	st.failSessionLookup = false

	// Authenticated account with an unencodable id -> 500.
	plain, _ := signupAndSignIn(t, h, "sessbad@example.com")
	for id, a := range st.accounts {
		if a.Email == "sessbad@example.com" {
			a.ID = "not-a-uuid"
			st.accounts[id] = a
		}
	}
	rec = call(t, h.GetSession, doReq(t, http.MethodGet, "/api/v1/auth/session", nil,
		&http.Cookie{Name: "226os_session", Value: plain}))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("encode error: %d %s", rec.Code, rec.Body.String())
	}
}

func TestGetSessionPreSessionAndSecureProfile(t *testing.T) {
	// Anonymous with a valid pre-session cookie: derived token, no new cookie.
	h, _, _, _ := testStack(t, false)
	rec := call(t, h.GetSession, doReq(t, http.MethodGet, "/api/v1/auth/session", nil))
	var secret *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_csrf" {
			secret = c
		}
	}
	if secret == nil {
		t.Fatal("no pre-session cookie issued")
	}
	var first struct {
		CsrfToken string `json:"csrf_token"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &first)
	rec = call(t, h.GetSession, doReq(t, http.MethodGet, "/api/v1/auth/session", nil, secret))
	var second struct {
		CsrfToken string `json:"csrf_token"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &second)
	if rec.Code != http.StatusOK || second.CsrfToken != first.CsrfToken {
		t.Fatalf("pre-session reuse: %d %s", rec.Code, rec.Body.String())
	}

	// Deployed profile: anonymous session answer sets __Host-226os_csrf.
	hSecure, _, _, _ := testStack(t, true)
	rec = call(t, hSecure.GetSession, doReq(t, http.MethodGet, "/api/v1/auth/session", nil))
	var pre *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "__Host-226os_csrf" {
			pre = c
		}
	}
	if pre == nil || !pre.Secure || !pre.HttpOnly {
		t.Fatalf("secure pre-session cookie %+v", pre)
	}
}

func TestGetSessionAnonymousCSRFFailure(t *testing.T) {
	_, svc, _, _ := testStack(t, false)
	h := New(anonFail{Service: svc, err: errors.New("no entropy")}, false, nil)
	rec := call(t, h.GetSession, doReq(t, http.MethodGet, "/api/v1/auth/session", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("entropy: %d %s", rec.Code, rec.Body.String())
	}
}

func TestVerifyEmailSuccess(t *testing.T) {
	h, _, _, mail := testStack(t, false)
	doJSON(t, h.SignUp, "/api/v1/auth/sign-up",
		`{"email":"verok@example.com","password":"correct-horse-battery","username":"verok"}`)
	sent := mail.Sent()
	if len(sent) == 0 {
		t.Fatal("no verification mail")
	}
	token := tokenFromMail(t, sent[0].Text, "Confirm your e-mail address:")

	rec := doJSON(t, h.VerifyEmail, "/api/v1/auth/verify-email", `{"token":"`+token+`"}`)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("verify: %d %s", rec.Code, rec.Body.String())
	}
	// Replay -> 422 INVALID_TOKEN.
	rec = doJSON(t, h.VerifyEmail, "/api/v1/auth/verify-email", `{"token":"`+token+`"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("replay: %d", rec.Code)
	}
}

func TestVerifyEmailDecodeError(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	rec := doJSON(t, h.VerifyEmail, "/api/v1/auth/verify-email", `{"token":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed: %d", rec.Code)
	}
}

func TestRequestPasswordResetDecodeAndStoreError(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	rec := doJSON(t, h.RequestPasswordReset, "/api/v1/auth/password-reset/request", `{"email":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed: %d", rec.Code)
	}
	st.failHits = true
	rec = doJSON(t, h.RequestPasswordReset, "/api/v1/auth/password-reset/request",
		`{"email":"rl@example.com"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("store error: %d %s", rec.Code, rec.Body.String())
	}
}

func TestConfirmPasswordResetFlows(t *testing.T) {
	h, _, _, mail := testStack(t, false)
	rec := doJSON(t, h.ConfirmPasswordReset, "/api/v1/auth/password-reset/confirm", `{"token":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed: %d", rec.Code)
	}
	rec = doJSON(t, h.ConfirmPasswordReset, "/api/v1/auth/password-reset/confirm",
		`{"token":"sometoken"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing new_password: %d", rec.Code)
	}

	// Full flow: sign up, request reset, confirm with the mailed token.
	plain, _ := signupAndSignIn(t, h, "rstok@example.com")
	rec = doJSON(t, h.RequestPasswordReset, "/api/v1/auth/password-reset/request",
		`{"email":"rstok@example.com"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("reset request: %d", rec.Code)
	}
	sent := mail.Sent()
	token := tokenFromMail(t, sent[len(sent)-1].Text, "Reset your password:")
	rec = doJSONWithCookie(t, h.ConfirmPasswordReset, "/api/v1/auth/password-reset/confirm",
		`{"token":"`+token+`","new_password":"another-good-password"}`,
		&http.Cookie{Name: "226os_session", Value: plain})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("confirm: %d %s", rec.Code, rec.Body.String())
	}
	// Session cookie cleared.
	cleared := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_session" && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("session cookie not cleared")
	}
}

func tokenFromMail(t *testing.T, text, marker string) string {
	t.Helper()
	parts := strings.Split(text, "\n\n")
	for i, p := range parts {
		if strings.Contains(p, marker) && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	t.Fatalf("marker %q not found in %q", marker, text)
	return ""
}

func validTokenPlain() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		xff        string
		remoteAddr string
		want       string
	}{
		{"xff list", "203.0.113.9, 10.0.0.1", "192.0.2.1:1234", "203.0.113.9"},
		{"xff single", "203.0.113.10", "192.0.2.1:1234", "203.0.113.10"},
		{"xff padded", "  203.0.113.11 , 10.0.0.2", "192.0.2.1:1234", "203.0.113.11"},
		{"remote hostport", "", "192.0.2.2:5555", "192.0.2.2"},
		{"remote bare", "", "192.0.2.3", "192.0.2.3"},
	}
	for _, tc := range cases {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		req.RemoteAddr = tc.remoteAddr
		if tc.xff != "" {
			req.Header.Set("X-Forwarded-For", tc.xff)
		}
		if got := clientIP(req); got != tc.want {
			t.Fatalf("%s: clientIP=%q want %q", tc.name, got, tc.want)
		}
	}
}

func TestCSRFMiddlewareInternalError(t *testing.T) {
	_, svc, st, _ := testStack(t, false)
	st.failSessionLookup = true
	mw := CSRFMiddleware(svc, false, nil)
	req := csrfReq(http.MethodPost, "/api/v1/auth/sign-out",
		map[string]string{"X-Csrf-Token": "x"},
		&http.Cookie{Name: "226os_session", Value: validTokenPlain()})
	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("store error: %d %s", rec.Code, rec.Body.String())
	}
}

func TestResolverInternalError(t *testing.T) {
	_, svc, st, _ := testStack(t, false)
	st.failSessionLookup = true
	res := NewResolver(svc, false)
	req := doReq(t, http.MethodGet, "/api/v1/x", nil,
		&http.Cookie{Name: "226os_session", Value: validTokenPlain()})
	if _, err := res.Resolve(req); err == nil {
		t.Fatal("store error swallowed")
	}
}
