//go:build integration

package pgtest

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/app"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

// identityStack wires the real S1 stack over a migrated database: sqlc
// repository, service, transport handler, session resolver and CSRF
// middleware inside the production router (validator included).
type identityStack struct {
	router http.Handler
	mail   *mailer.Fake
}

func newIdentityStack(ctx context.Context, t *testing.T) *identityStack {
	t.Helper()
	conn := migratedDB(ctx, t)
	t.Cleanup(func() { _ = conn.Close(ctx) })
	mail := mailer.NewFake()
	svc := service.New(repository.New(conn), mail, []byte("0123456789abcdef0123456789abcdef"))
	handler := identitytransport.New(svc, false, nil)
	deps := httpx.Deps{
		Server: app.NewServer(app.WithIdentity(handler)),
		Ready:  conn,
	}
	deps.Subjects = identitytransport.NewResolver(svc, false)
	deps.CSRF = identitytransport.CSRFMiddleware(svc, false, nil)
	return &identityStack{
		router: httpx.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), deps),
		mail:   mail,
	}
}

func (s *identityStack) do(t *testing.T, method, path, body string, headers map[string]string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequestWithContext(t.Context(), method, path, rdr)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

// csrfToken runs GET /auth/csrf and returns the cookie + token pair a browser
// would hold.
func (s *identityStack) csrfToken(t *testing.T) (*http.Cookie, string) {
	t.Helper()
	rec := s.do(t, http.MethodGet, "/api/v1/auth/csrf", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("csrf %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		CsrfToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.CsrfToken == "" {
		t.Fatalf("csrf body %s", rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_csrf" {
			return c, body.CsrfToken
		}
	}
	t.Fatal("no pre-session cookie")
	return nil, ""
}

// signUp registers an account through the full stack and returns nothing; the
// verification token is read back from the captured mail.
func (s *identityStack) signUp(t *testing.T, email, username string) {
	t.Helper()
	cookie, token := s.csrfToken(t)
	rec := s.do(t, http.MethodPost, "/api/v1/auth/sign-up",
		`{"email":"`+email+`","password":"correct-horse-battery","username":"`+username+`","display_name":"Test User"}`,
		map[string]string{"X-CSRF-Token": token}, cookie)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("sign-up %d %s", rec.Code, rec.Body.String())
	}
}

// mailToken extracts the single-use token from the last captured mail body.
func (s *identityStack) mailToken(t *testing.T, to string) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		// Iterate all messages and keep the LAST match: a second flow (reset)
		// sends a newer token for the same address than the verification mail.
		var token string
		for _, m := range s.mail.Sent() {
			if len(m.To) == 1 && m.To[0] == to {
				// The token is the only bare base64url line in the body.
				for _, line := range strings.Split(m.Text, "\n") {
					line = strings.TrimSpace(line)
					if len(line) >= 40 && !strings.ContainsAny(line, " :") {
						token = line
					}
				}
			}
		}
		if token != "" {
			return token
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no mail token for %s", to)
	return ""
}

func TestIdentityEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	s := newIdentityStack(ctx, t)

	// Anonymous bootstrap: GET /auth/session is 200, never 401.
	rec := s.do(t, http.MethodGet, "/api/v1/auth/session", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("session bootstrap %d", rec.Code)
	}
	var anon struct {
		Authenticated bool `json:"authenticated"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &anon); err != nil || anon.Authenticated {
		t.Fatalf("anonymous state %s", rec.Body.String())
	}

	// CSRF failure: POST sign-up without token -> 403 CSRF_FAILED.
	rec = s.do(t, http.MethodPost, "/api/v1/auth/sign-up",
		`{"email":"e2e@example.com","password":"correct-horse-battery","username":"e2euser"}`, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("no csrf: %d", rec.Code)
	}

	// Sign-up -> verification mail -> verify -> sign-in -> session -> sign-out.
	s.signUp(t, "e2e@example.com", "e2euser")
	vtok := s.mailToken(t, "e2e@example.com")
	cookie, csrf := s.csrfToken(t)
	rec = s.do(t, http.MethodPost, "/api/v1/auth/verify-email",
		`{"token":"`+vtok+`"}`, map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("verify %d %s", rec.Code, rec.Body.String())
	}
	// Replay must fail: single-use.
	rec = s.do(t, http.MethodPost, "/api/v1/auth/verify-email",
		`{"token":"`+vtok+`"}`, map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("verify replay %d", rec.Code)
	}

	// Sign-in with the pre-session CSRF.
	rec = s.do(t, http.MethodPost, "/api/v1/auth/sign-in",
		`{"email":"e2e@example.com","password":"correct-horse-battery"}`,
		map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("sign-in %d %s", rec.Code, rec.Body.String())
	}
	var sess *http.Cookie
	var state struct {
		Authenticated bool   `json:"authenticated"`
		CsrfToken     string `json:"csrf_token"`
		User          *struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_session" {
			sess = c
		}
	}
	if sess == nil || !sess.HttpOnly || sess.Secure {
		t.Fatalf("session cookie %+v", sess)
	}
	if !state.Authenticated || state.User == nil || !state.User.EmailVerified {
		t.Fatalf("state %s", rec.Body.String())
	}

	// Authenticated session read.
	rec = s.do(t, http.MethodGet, "/api/v1/auth/session", "", nil, sess)
	if rec.Code != http.StatusOK {
		t.Fatalf("session %d", rec.Code)
	}
	var st2 struct {
		Authenticated bool `json:"authenticated"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &st2)
	if !st2.Authenticated {
		t.Fatalf("session state %s", rec.Body.String())
	}

	// Sign-out with the session-bound CSRF -> 204, session dead afterwards.
	rec = s.do(t, http.MethodPost, "/api/v1/auth/sign-out", "",
		map[string]string{"X-CSRF-Token": state.CsrfToken}, sess)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("sign-out %d %s", rec.Code, rec.Body.String())
	}
	rec = s.do(t, http.MethodGet, "/api/v1/auth/session", "", nil, sess)
	_ = json.Unmarshal(rec.Body.Bytes(), &st2)
	if st2.Authenticated {
		t.Fatal("session alive after sign-out")
	}
}

func TestIdentitySignInRotationAndReset(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newIdentityStack(ctx, t)
	s.signUp(t, "rot@example.com", "rotuser")

	// Sign in twice: second sign-in must not kill the first session.
	cookie, csrf := s.csrfToken(t)
	signIn := func() *http.Cookie {
		rec := s.do(t, http.MethodPost, "/api/v1/auth/sign-in",
			`{"email":"rot@example.com","password":"correct-horse-battery"}`,
			map[string]string{"X-CSRF-Token": csrf}, cookie)
		if rec.Code != http.StatusOK {
			t.Fatalf("sign-in %d %s", rec.Code, rec.Body.String())
		}
		for _, c := range rec.Result().Cookies() {
			if c.Name == "226os_session" {
				return c
			}
		}
		t.Fatal("no session cookie")
		return nil
	}
	s1 := signIn()
	s2 := signIn()
	if s1.Value == s2.Value {
		t.Fatal("session token not rotated")
	}
	for _, sess := range []*http.Cookie{s1, s2} {
		rec := s.do(t, http.MethodGet, "/api/v1/auth/session", "", nil, sess)
		var st struct {
			Authenticated bool `json:"authenticated"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &st)
		if !st.Authenticated {
			t.Fatalf("session %s… dead", sess.Value[:8])
		}
	}

	// Wrong password -> generic 401, same body as unknown email.
	cookie, csrf = s.csrfToken(t)
	rec := s.do(t, http.MethodPost, "/api/v1/auth/sign-in",
		`{"email":"rot@example.com","password":"wrong-password-1"}`,
		map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "INVALID_CREDENTIALS") {
		t.Fatalf("wrong password %d %s", rec.Code, rec.Body.String())
	}

	// Password reset request -> token from mail -> confirm -> all sessions dead.
	rec = s.do(t, http.MethodPost, "/api/v1/auth/password-reset/request",
		`{"email":"rot@example.com"}`, map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("reset request %d", rec.Code)
	}
	rtok := s.mailToken(t, "rot@example.com")
	cookie, csrf = s.csrfToken(t)
	rec = s.do(t, http.MethodPost, "/api/v1/auth/password-reset/confirm",
		`{"token":"`+rtok+`","new_password":"brand-new-password-1"}`,
		map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("reset confirm %d %s", rec.Code, rec.Body.String())
	}
	for _, sess := range []*http.Cookie{s1, s2} {
		rec := s.do(t, http.MethodGet, "/api/v1/auth/session", "", nil, sess)
		var st struct {
			Authenticated bool `json:"authenticated"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &st)
		if st.Authenticated {
			t.Fatal("session alive after password reset")
		}
	}
	// New password signs in.
	cookie, csrf = s.csrfToken(t)
	rec = s.do(t, http.MethodPost, "/api/v1/auth/sign-in",
		`{"email":"rot@example.com","password":"brand-new-password-1"}`,
		map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("sign-in new password %d", rec.Code)
	}
}

// Two reset mails may be live at once (3/account/hour), but confirming with
// one must burn every other open token for that account.
func TestIdentityResetBurnsSiblingTokens(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	s := newIdentityStack(ctx, t)

	s.signUp(t, "sib@example.com", "sibuser")
	vtok := s.mailToken(t, "sib@example.com")
	cookie, csrf := s.csrfToken(t)
	rec := s.do(t, http.MethodPost, "/api/v1/auth/verify-email",
		`{"token":"`+vtok+`"}`, map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("verify %d", rec.Code)
	}

	// Two reset requests -> two live tokens; mailToken returns the latest.
	cookie, csrf = s.csrfToken(t)
	for i := 0; i < 2; i++ {
		rec = s.do(t, http.MethodPost, "/api/v1/auth/password-reset/request",
			`{"email":"sib@example.com"}`, map[string]string{"X-CSRF-Token": csrf}, cookie)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("reset request %d", rec.Code)
		}
	}
	newest := s.mailToken(t, "sib@example.com")
	// The older token is the previous mail's: extract it by counting mails.
	var tokens []string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && len(tokens) < 3 {
		tokens = nil
		for _, m := range s.mail.Sent() {
			if len(m.To) == 1 && m.To[0] == "sib@example.com" {
				for _, line := range strings.Split(m.Text, "\n") {
					line = strings.TrimSpace(line)
					if len(line) >= 40 && !strings.ContainsAny(line, " :") {
						tokens = append(tokens, line)
					}
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if len(tokens) < 3 {
		t.Fatalf("expected verification + 2 reset tokens, got %d", len(tokens))
	}
	oldest := tokens[len(tokens)-2] // first reset token (before newest)

	// Confirm with the OLDER token: it must still work once...
	cookie, csrf = s.csrfToken(t)
	rec = s.do(t, http.MethodPost, "/api/v1/auth/password-reset/confirm",
		`{"token":"`+oldest+`","new_password":"sib-new-password-1"}`,
		map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("confirm older token %d %s", rec.Code, rec.Body.String())
	}
	// ...and the newer sibling must now be dead.
	cookie, csrf = s.csrfToken(t)
	rec = s.do(t, http.MethodPost, "/api/v1/auth/password-reset/confirm",
		`{"token":"`+newest+`","new_password":"sib-other-password-1"}`,
		map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sibling token survived: %d %s", rec.Code, rec.Body.String())
	}
}

func TestIdentityRateLimitFailsClosed(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	s := newIdentityStack(ctx, t)
	// sign-up is limited to 5/IP/hour; the 6th call must answer 429.
	var last *httptest.ResponseRecorder
	for i := 0; i < 6; i++ {
		cookie, token := s.csrfToken(t)
		last = s.do(t, http.MethodPost, "/api/v1/auth/sign-up",
			`{"email":"rl`+string(rune('a'+i))+`@example.com","password":"correct-horse-battery","username":"rluser`+string(rune('a'+i))+`"}`,
			map[string]string{"X-CSRF-Token": token}, cookie)
	}
	if last.Code != http.StatusTooManyRequests {
		t.Fatalf("6th sign-up: %d %s", last.Code, last.Body.String())
	}
	if ra := last.Header().Get("Retry-After"); ra == "" {
		t.Fatal("429 without Retry-After")
	}
}

func TestIdentityRateLimitPurgesOldWindows(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	conn := migratedDB(ctx, t)
	t.Cleanup(func() { _ = conn.Close(ctx) })
	now := time.Now().UTC().Truncate(time.Minute)
	for i, age := range []time.Duration{3 * time.Hour, 5 * time.Hour, 30 * time.Minute} {
		if _, err := conn.Exec(ctx,
			`INSERT INTO rate_limit_buckets (scope, key_hash, window_start, hits) VALUES ('old', $1, $2, 1)`,
			[]byte{byte(i)}, now.Add(-age)); err != nil {
			t.Fatal(err)
		}
	}
	// Opening a new window purges the finished ones older than 2 h and keeps the recent one.
	ok, _, err := repository.New(conn).Hit(ctx, "signin:ip", [32]byte{9}, 5, 15*time.Minute, now)
	if err != nil || !ok {
		t.Fatalf("hit: %v %v", ok, err)
	}
	var old, recent int
	if err := conn.QueryRow(ctx, `SELECT count(*) FILTER (WHERE window_start < $1), count(*) FILTER (WHERE window_start >= $1) FROM rate_limit_buckets`,
		now.Add(-2*time.Hour)).Scan(&old, &recent); err != nil {
		t.Fatal(err)
	}
	if old != 0 || recent != 2 {
		t.Fatalf("old=%d recent=%d, want 0 and 2 (the kept 30 min window plus the new bucket)", old, recent)
	}
}
