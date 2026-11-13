package transport

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestSignUpIsGenericAccepted(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	for _, email := range []string{"new@example.com", "taken@example.com"} {
		if email == "taken@example.com" {
			if _, err := st.CreateAccount(t.Context(), newAccount(email)); err != nil {
				t.Fatal(err)
			}
		}
		rec := doJSON(t, h.SignUp, "/api/v1/auth/sign-up",
			`{"email":"`+email+`","password":"correct-horse-battery","username":"u`+email[:3]+`"}`)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("%s: %d %s", email, rec.Code, rec.Body.String())
		}
		var body struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Status != "accepted" {
			t.Fatalf("%s: body %s", email, rec.Body.String())
		}
	}
}

func TestSignUpRejectsBadJSONAndUnknownField(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	if rec := doJSON(t, h.SignUp, "/api/v1/auth/sign-up", `{"email":`); rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed: %d", rec.Code)
	}
	if rec := doJSON(t, h.SignUp, "/api/v1/auth/sign-up",
		`{"email":"a@b.co","password":"x","username":"abc","extra":1}`); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unknown field: %d %s", rec.Code, rec.Body.String())
	}
}

func TestSignInSetsSessionCookieAndReturnsState(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	plain, csrf := signupAndSignIn(t, h, "in@example.com")
	if plain == "" || csrf == "" {
		t.Fatal("missing cookie or csrf token")
	}
}

func TestSignInGenericInvalidCredentials(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	signupAndSignIn(t, h, "real@example.com")
	for _, body := range []string{
		`{"email":"ghost@example.com","password":"whatever-pass"}`,
		`{"email":"real@example.com","password":"wrong-password"}`,
	} {
		rec := doJSON(t, h.SignIn, "/api/v1/auth/sign-in", body)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%d %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "INVALID_CREDENTIALS") {
			t.Fatalf("body %s", rec.Body.String())
		}
	}
}

func TestSessionCookieFlagsPerProfile(t *testing.T) {
	// local-dev: no Secure, plain name. deployed: Secure + __Host- name.
	for _, tc := range []struct {
		secure     bool
		wantName   string
		wantSecure bool
	}{
		{false, "226os_session", false},
		{true, "__Host-226os_session", true},
	} {
		h, _, _, _ := testStack(t, tc.secure)
		rec := doJSON(t, h.SignUp, "/api/v1/auth/sign-up",
			`{"email":"c@example.com","password":"correct-horse-battery","username":"cook"}`)
		_ = rec
		rec = doJSON(t, h.SignIn, "/api/v1/auth/sign-in",
			`{"email":"c@example.com","password":"correct-horse-battery"}`)
		var sess *http.Cookie
		for _, c := range rec.Result().Cookies() {
			if strings.Contains(c.Name, "session") {
				sess = c
			}
		}
		if sess == nil {
			t.Fatalf("secure=%v: no session cookie", tc.secure)
		}
		if sess.Name != tc.wantName || sess.Secure != tc.wantSecure || !sess.HttpOnly || sess.SameSite != http.SameSiteLaxMode {
			t.Fatalf("secure=%v cookie %+v", tc.secure, sess)
		}
		if sess.Path != "/" || sess.Domain != "" {
			t.Fatalf("secure=%v path/domain %+v", tc.secure, sess)
		}
	}
}

func TestGetSessionNever401(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	rec := call(t, h.GetSession, doReq(t, http.MethodGet, "/api/v1/auth/session", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("anonymous: %d", rec.Code)
	}
	var state struct {
		Authenticated bool   `json:"authenticated"`
		CsrfToken     string `json:"csrf_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Authenticated || state.CsrfToken == "" {
		t.Fatalf("anonymous state %s", rec.Body.String())
	}
	// The anonymous answer sets the pre-session cookie.
	var pre *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_csrf" {
			pre = c
		}
	}
	if pre == nil || pre.Value == "" {
		t.Fatal("pre-session cookie not set")
	}
}

func TestGetSessionAuthenticated(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	plain, _ := signupAndSignIn(t, h, "sess@example.com")
	req := doReq(t, http.MethodGet, "/api/v1/auth/session", nil,
		&http.Cookie{Name: "226os_session", Value: plain})
	rec := call(t, h.GetSession, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	var state struct {
		Authenticated bool `json:"authenticated"`
		User          *struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if !state.Authenticated || state.User == nil || state.User.Email != "sess@example.com" {
		t.Fatalf("state %s", rec.Body.String())
	}
}

func TestSignOutRequiresMember(t *testing.T) {
	// Route: POST /api/v1/auth/sign-out — anonymous callers get 401.
	h, _, _, _ := testStack(t, false)
	rec := call(t, h.SignOut, doReq(t, http.MethodPost, "/api/v1/auth/sign-out", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous sign-out: %d", rec.Code)
	}
}

func TestAuthzSignOutAnonymous(t *testing.T) {
	// Negative authz for POST /api/v1/auth/sign-out (AUTHORIZATION_MATRIX,
	// allow: member): an anonymous request must be denied with 401.
	h, _, _, _ := testStack(t, false)
	rec := call(t, h.SignOut, doReq(t, http.MethodPost, "/api/v1/auth/sign-out", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "401") && !strings.Contains(rec.Body.String(), "UNAUTHENTICATED") {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestSignOutRevokesOwnSession(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	plain, _ := signupAndSignIn(t, h, "out@example.com")
	req := doReq(t, http.MethodPost, "/api/v1/auth/sign-out", nil,
		&http.Cookie{Name: "226os_session", Value: plain})
	// put a member subject in the context like subject.Middleware would
	req = req.WithContext(withSubject(req, "member"))
	rec := call(t, h.SignOut, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("sign-out %d %s", rec.Code, rec.Body.String())
	}
	// The revoked session no longer resolves.
	req = doReq(t, http.MethodGet, "/api/v1/auth/session", nil,
		&http.Cookie{Name: "226os_session", Value: plain})
	rec = call(t, h.GetSession, req)
	var state struct {
		Authenticated bool `json:"authenticated"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &state)
	if state.Authenticated {
		t.Fatal("session still valid after sign-out")
	}
}

func TestVerifyEmailRejectsBadToken(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	rec := doJSON(t, h.VerifyEmail, "/api/v1/auth/verify-email", `{"token":"!!!not-a-token"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_TOKEN") {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestRequestPasswordResetAlways202(t *testing.T) {
	h, _, st, mail := testStack(t, false)
	signupAndSignIn(t, h, "reset@example.com")
	for _, email := range []string{"reset@example.com", "ghost@example.com"} {
		rec := doJSON(t, h.RequestPasswordReset, "/api/v1/auth/password-reset/request",
			`{"email":"`+email+`"}`)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("%s: %d", email, rec.Code)
		}
	}
	if len(mail.Sent()) == 0 {
		t.Fatal("no reset mail sent for known account")
	}
	_ = st
}

func TestConfirmPasswordResetRejectsBadToken(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	rec := doJSON(t, h.ConfirmPasswordReset, "/api/v1/auth/password-reset/confirm",
		`{"token":"bad","new_password":"new-correct-horse"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}
