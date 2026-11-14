//go:build integration

package pgtest

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Y4NN777/226os/src/api/internal/app"
	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
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
	conn   *pgx.Conn
}

func newIdentityStack(ctx context.Context, t *testing.T) *identityStack {
	t.Helper()
	conn := migratedDB(ctx, t)
	t.Cleanup(func() { _ = conn.Close(ctx) })
	mail := mailer.NewFake()
	// One connection serves the whole stack, so background work runs inline.
	svc := service.New(repository.New(conn), mail, []byte("0123456789abcdef0123456789abcdef"),
		service.WithBackground(func(f func()) { f() }))
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
		conn:   conn,
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

// mailTokens returns the single-use tokens of every mail captured for to, in
// send order. The token is the only bare base64url line in a body.
func (s *identityStack) mailTokens(to string) []string {
	var tokens []string
	for _, m := range s.mail.Sent() {
		if len(m.To) != 1 || m.To[0] != to {
			continue
		}
		for _, line := range strings.Split(m.Text, "\n") {
			line = strings.TrimSpace(line)
			if len(line) >= 40 && !strings.ContainsAny(line, " :") {
				tokens = append(tokens, line)
			}
		}
	}
	return tokens
}

// mailToken waits for the mail that follows the first `after` mails captured
// for to, and returns its token. Mails are sent off the request path (a
// goroutine), so the caller passes len(s.mailTokens(to)) taken BEFORE the
// action: returning the latest token instead races the send and can hand back
// an older mail's token (#91).
func (s *identityStack) mailToken(t *testing.T, to string, after int) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if tokens := s.mailTokens(to); len(tokens) > after {
			return tokens[after]
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no new mail token for %s after %d", to, after)
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
	vtok := s.mailToken(t, "e2e@example.com", 0)
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
	_ = s.mailToken(t, "rot@example.com", 0) // wait for the verification mail so the reset mail is the 2nd

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
	rtok := s.mailToken(t, "rot@example.com", 1)
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
	vtok := s.mailToken(t, "sib@example.com", 0)
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
	// Wait for both reset mails (after the verification mail); the two sends
	// race each other, so which is "older" does not matter: both are live.
	oldest := s.mailToken(t, "sib@example.com", 1)
	newest := s.mailToken(t, "sib@example.com", 2)
	if oldest == newest {
		t.Fatal("the two reset mails carry the same token")
	}

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

// signInMember registers nothing new: it signs in with the stack password and
// returns the session cookie plus the session-bound CSRF token.
func (s *identityStack) signInMember(t *testing.T, email string) (*http.Cookie, string) {
	t.Helper()
	cookie, csrf := s.csrfToken(t)
	rec := s.do(t, http.MethodPost, "/api/v1/auth/sign-in",
		`{"email":"`+email+`","password":"correct-horse-battery"}`,
		map[string]string{"X-CSRF-Token": csrf}, cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("sign-in %d %s", rec.Code, rec.Body.String())
	}
	var state struct {
		CsrfToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil || state.CsrfToken == "" {
		t.Fatalf("sign-in body %s", rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_session" {
			return c, state.CsrfToken
		}
	}
	t.Fatal("no session cookie")
	return nil, ""
}

func TestIdentityProfilesEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newIdentityStack(ctx, t)
	s.signUp(t, "prof@example.com", "profuser")
	sess, csrf := s.signInMember(t, "prof@example.com")

	// Anonymous callers get 401 on /me.
	rec := s.do(t, http.MethodGet, "/api/v1/me", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anon /me %d", rec.Code)
	}

	// Member getMe: account + own profile, checklist derived.
	rec = s.do(t, http.MethodGet, "/api/v1/me", "", nil, sess)
	if rec.Code != http.StatusOK {
		t.Fatalf("getMe %d %s", rec.Code, rec.Body.String())
	}
	var me struct {
		Account struct {
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"account"`
		Profile struct {
			Username          string   `json:"username"`
			DisplayName       string   `json:"display_name"`
			ChecklistMissing  []string `json:"checklist_missing"`
			ProfileVisibility string   `json:"profile_visibility"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.Account.Email != "prof@example.com" || me.Profile.Username != "profuser" ||
		len(me.Profile.ChecklistMissing) == 0 || me.Profile.ProfileVisibility != "public" {
		t.Fatalf("me %s", rec.Body.String())
	}

	// Seed an active tag so the skills patch resolves (same database the
	// stack serves: migratedDB clones a fresh schema per call).
	conn := s.conn
	if _, err := conn.Exec(ctx,
		`INSERT INTO tags (slug, label, status, provenance) VALUES ('go', 'Go', 'active', 'live'), ('rust', 'Rust', 'active', 'live')`); err != nil {
		t.Fatalf("seed tag: %v", err)
	}

	// Merge patch: set fields, replace skills atomically.
	rec = s.do(t, http.MethodPatch, "/api/v1/me/profile",
		`{"display_name":"Pro","intent":"help","country":"BJ","languages":["fr"],`+
			`"github":"octocat","bio":"hello world","open_to_collab":true,"skills":["go"]}`,
		map[string]string{"X-CSRF-Token": csrf}, sess)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	var patched struct {
		Bio          *string  `json:"bio"`
		Country      *string  `json:"country"`
		Languages    []string `json:"languages"`
		Skills       []string `json:"skills"`
		OpenToCollab bool     `json:"open_to_collab"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &patched); err != nil {
		t.Fatal(err)
	}
	if patched.Bio == nil || *patched.Bio != "hello world" || patched.Country == nil ||
		*patched.Country != "BJ" || len(patched.Skills) != 1 || patched.Skills[0] != "go" ||
		!patched.OpenToCollab {
		t.Fatalf("patched %s", rec.Body.String())
	}

	// Overlapping skill replacements: [go] -> [go, rust] -> [rust]. The
	// delete/insert CTEs are disjoint, so kept tags survive both directions.
	for _, step := range []struct {
		patch string
		want  []string
	}{
		{`{"skills":["go","rust"]}`, []string{"go", "rust"}},
		{`{"skills":["rust"]}`, []string{"rust"}},
	} {
		rec = s.do(t, http.MethodPatch, "/api/v1/me/profile", step.patch,
			map[string]string{"X-CSRF-Token": csrf}, sess)
		if rec.Code != http.StatusOK {
			t.Fatalf("skills %s %d %s", step.patch, rec.Code, rec.Body.String())
		}
		var got struct {
			Skills []string `json:"skills"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		if !slices.Equal(got.Skills, step.want) {
			t.Fatalf("skills %s -> %v, want %v", step.patch, got.Skills, step.want)
		}
	}

	// bio_language derived to 'en' by the service (contract exposes no field).
	var bioLang *string
	if err := conn.QueryRow(ctx,
		`SELECT bio_language FROM profiles WHERE username = 'profuser'`).Scan(&bioLang); err != nil {
		t.Fatalf("bio_language: %v", err)
	}
	if bioLang == nil || *bioLang != "en" {
		t.Fatalf("bio_language = %v", bioLang)
	}

	// Explicit null clears bio and its derived language together.
	rec = s.do(t, http.MethodPatch, "/api/v1/me/profile", `{"bio":null}`,
		map[string]string{"X-CSRF-Token": csrf}, sess)
	if rec.Code != http.StatusOK {
		t.Fatalf("clear bio %d %s", rec.Code, rec.Body.String())
	}
	var cleared struct {
		Bio *string `json:"bio"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &cleared)
	if cleared.Bio != nil {
		t.Fatalf("bio not cleared: %s", rec.Body.String())
	}
	if err := conn.QueryRow(ctx,
		`SELECT bio_language FROM profiles WHERE username = 'profuser'`).Scan(&bioLang); err != nil {
		t.Fatal(err)
	}
	if bioLang != nil {
		t.Fatalf("bio_language survived bio clear: %v", *bioLang)
	}

	// Unknown skill slug rejects the whole patch (422), nothing applied.
	rec = s.do(t, http.MethodPatch, "/api/v1/me/profile", `{"skills":["nope"],"display_name":"ShouldNotApply"}`,
		map[string]string{"X-CSRF-Token": csrf}, sess)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unknown skill %d %s", rec.Code, rec.Body.String())
	}
	assertProblemViolations(t, rec, "/skills", "invalid_value")
	rec = s.do(t, http.MethodGet, "/api/v1/me", "", nil, sess)
	var after struct {
		Profile struct {
			DisplayName string `json:"display_name"`
		} `json:"profile"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &after)
	if after.Profile.DisplayName == "ShouldNotApply" {
		t.Fatal("partial patch applied despite unknown skill")
	}

	// Unknown country code hits the FK -> 422, not 500.
	rec = s.do(t, http.MethodPatch, "/api/v1/me/profile", `{"country":"ZZ"}`,
		map[string]string{"X-CSRF-Token": csrf}, sess)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unknown country %d %s", rec.Code, rec.Body.String())
	}
	assertProblemViolations(t, rec, "/country", "invalid_value")

	// NUL in a free-text field is a 422, not a Postgres 22021 -> 500 (issue #116).
	rec = s.do(t, http.MethodPatch, "/api/v1/me/profile", `{"bio":"a\u0000b"}`,
		map[string]string{"X-CSRF-Token": csrf}, sess)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("NUL bio %d %s", rec.Code, rec.Body.String())
	}
	assertProblemViolations(t, rec, "/bio", "invalid_value")

	// Empty patch -> 422.
	rec = s.do(t, http.MethodPatch, "/api/v1/me/profile", `{}`,
		map[string]string{"X-CSRF-Token": csrf}, sess)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("empty patch %d", rec.Code)
	}

	// Public profile readable anonymously; visibility flip hides it.
	rec = s.do(t, http.MethodGet, "/api/v1/profiles/profuser", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("public profile %d %s", rec.Code, rec.Body.String())
	}
	rec = s.do(t, http.MethodPatch, "/api/v1/me/profile", `{"profile_visibility":"members"}`,
		map[string]string{"X-CSRF-Token": csrf}, sess)
	if rec.Code != http.StatusOK {
		t.Fatalf("vis patch %d", rec.Code)
	}
	rec = s.do(t, http.MethodGet, "/api/v1/profiles/profuser", "", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("anon on members profile: %d", rec.Code)
	}
	rec = s.do(t, http.MethodGet, "/api/v1/profiles/profuser", "", nil, sess)
	if rec.Code != http.StatusOK {
		t.Fatalf("member on members profile: %d", rec.Code)
	}
	rec = s.do(t, http.MethodGet, "/api/v1/profiles/ghostuser", "", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown profile: %d", rec.Code)
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

// ADR-0003: 5 failures per account per window. Successful sign-ins are not
// counted (a demo account used from several devices is never throttled); after
// five failures even the right password gets 429 with Retry-After, and an
// unknown email behaves exactly the same.
func TestIdentitySignInCountsOnlyFailures(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	s := newIdentityStack(ctx, t)
	s.signUp(t, "cnt@example.com", "cntuser")
	cookie, csrf := s.csrfToken(t)
	signIn := func(email, pw string) *httptest.ResponseRecorder {
		return s.do(t, http.MethodPost, "/api/v1/auth/sign-in",
			`{"email":"`+email+`","password":"`+pw+`"}`, map[string]string{"X-CSRF-Token": csrf}, cookie)
	}
	for i := 0; i < 7; i++ {
		if rec := signIn("cnt@example.com", "correct-horse-battery"); rec.Code != http.StatusOK {
			t.Fatalf("success %d: %d %s", i+1, rec.Code, rec.Body.String())
		}
	}
	for _, email := range []string{"cnt@example.com", "ghost@example.com"} {
		for i := 0; i < 5; i++ {
			if rec := signIn(email, "wrong-password-123"); rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s failure %d: %d", email, i+1, rec.Code)
			}
		}
		rec := signIn(email, "correct-horse-battery")
		if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" {
			t.Fatalf("%s after 5 failures: %d retry-after=%q", email, rec.Code, rec.Header().Get("Retry-After"))
		}
	}
}

// TestIdentityProfileHiddenTagFiltered proves read queries only surface
// active tags: a hidden tag linked to the user must not appear on /me or on
// the public profile, and patching the visible set back must not fail.
func TestIdentityProfileHiddenTagFiltered(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newIdentityStack(ctx, t)
	s.signUp(t, "hid@example.com", "hiduser")
	sess, csrf := s.signInMember(t, "hid@example.com")
	conn := s.conn

	if _, err := conn.Exec(ctx,
		`INSERT INTO tags (slug, label, status, provenance) VALUES
		 ('go', 'Go', 'active', 'live'), ('secret', 'Secret', 'hidden', 'live')`); err != nil {
		t.Fatalf("seed tags: %v", err)
	}
	// Link both tags directly, bypassing the service resolution.
	if _, err := conn.Exec(ctx,
		`INSERT INTO user_skills (user_id, tag_id)
		 SELECT u.id, t.id FROM users u, tags t
		 WHERE u.email = 'hid@example.com' AND t.slug IN ('go', 'secret')`); err != nil {
		t.Fatalf("link skills: %v", err)
	}

	wantSkills := func(path string, cookies ...*http.Cookie) {
		t.Helper()
		rec := s.do(t, http.MethodGet, path, "", nil, cookies...)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s %d %s", path, rec.Code, rec.Body.String())
		}
		var body struct {
			Profile struct {
				Skills []string `json:"skills"`
			} `json:"profile"`
			Skills []string `json:"skills"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		got := body.Profile.Skills
		if got == nil {
			got = body.Skills
		}
		if !slices.Equal(got, []string{"go"}) {
			t.Fatalf("%s skills = %v, want [go] (hidden tag must be filtered)", path, got)
		}
	}
	wantSkills("/api/v1/me", sess)
	wantSkills("/api/v1/profiles/hiduser")

	// Saving the visible set back must succeed: the hidden tag is neither
	// resolved nor blocking.
	rec := s.do(t, http.MethodPatch, "/api/v1/me/profile", `{"skills":["go"]}`,
		map[string]string{"X-CSRF-Token": csrf}, sess)
	if rec.Code != http.StatusOK {
		t.Fatalf("resave visible skills %d %s", rec.Code, rec.Body.String())
	}
}

// TestIdentityProfileConcurrentSkillReplace proves two concurrent skill
// replacements serialize on the profile row lock: the final set is exactly
// one caller's set, never a union and never a 500.
func TestIdentityProfileConcurrentSkillReplace(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	s := newIdentityStack(ctx, t)
	s.signUp(t, "race@example.com", "raceuser")
	conn := s.conn

	if _, err := conn.Exec(ctx,
		`INSERT INTO tags (slug, label, status, provenance) VALUES
		 ('go', 'Go', 'active', 'live'), ('rust', 'Rust', 'active', 'live')`); err != nil {
		t.Fatalf("seed tags: %v", err)
	}
	var uid string
	if err := conn.QueryRow(ctx,
		`SELECT id::text FROM users WHERE email = 'race@example.com'`).Scan(&uid); err != nil {
		t.Fatalf("user id: %v", err)
	}
	tagIDs := map[string]string{}
	rows, err := conn.Query(ctx, `SELECT slug, id::text FROM tags WHERE slug IN ('go','rust')`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var slug, id string
		if err := rows.Scan(&slug, &id); err != nil {
			t.Fatal(err)
		}
		tagIDs[slug] = id
	}
	rows.Close()

	// A second connection gives the two repositories independent sessions so
	// the updates genuinely overlap on the server.
	dsn := conn.Config().ConnString()
	conn2, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("second conn: %v", err)
	}
	defer func() { _ = conn2.Close(context.Background()) }()
	repo1 := repository.New(conn)
	repo2 := repository.New(conn2)

	write := func(ids ...string) domain.ProfileWrite {
		return domain.ProfileWrite{SkillTagIDs: domain.PatchField[[]string]{Present: true, Value: &ids}}
	}
	setA := write(tagIDs["go"])
	setB := write(tagIDs["rust"])

	finalSkills := func() []string {
		t.Helper()
		rows, err := conn.Query(ctx,
			`SELECT t.slug FROM user_skills us JOIN tags t ON t.id = us.tag_id
			 WHERE us.user_id = $1::uuid ORDER BY t.slug`, uid)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				t.Fatal(err)
			}
			out = append(out, s)
		}
		return out
	}

	// Conflicting sets: repeat enough rounds that the row-lock wait overlaps;
	// the final state must be exactly one side, never [go rust].
	for i := 0; i < 10; i++ {
		var wg sync.WaitGroup
		errs := make([]error, 2)
		wg.Add(2)
		go func() { defer wg.Done(); errs[0] = repo1.UpdateProfile(ctx, uid, setA) }()
		go func() { defer wg.Done(); errs[1] = repo2.UpdateProfile(ctx, uid, setB) }()
		wg.Wait()
		for j, err := range errs {
			if err != nil {
				t.Fatalf("round %d update %d: %v", i, j, err)
			}
		}
		got := finalSkills()
		exact := slices.Equal(got, []string{"go"}) || slices.Equal(got, []string{"rust"})
		if !exact {
			t.Fatalf("round %d final skills %v: union or partial merge", i, got)
		}
	}

	// Identical concurrent sets (the double-click case): both succeed and the
	// set stays exact.
	for i := 0; i < 5; i++ {
		var wg sync.WaitGroup
		errs := make([]error, 2)
		wg.Add(2)
		go func() { defer wg.Done(); errs[0] = repo1.UpdateProfile(ctx, uid, setA) }()
		go func() { defer wg.Done(); errs[1] = repo2.UpdateProfile(ctx, uid, setA) }()
		wg.Wait()
		for j, err := range errs {
			if err != nil {
				t.Fatalf("identical round %d update %d: %v", i, j, err)
			}
		}
		if got := finalSkills(); !slices.Equal(got, []string{"go"}) {
			t.Fatalf("identical round %d skills %v", i, got)
		}
	}
}

// assertProblemViolations checks a 422 problem body against the contract:
// every violation code is in the Violation.code enum and its field is the
// JSON pointer of the offending request member (issue #116, S7 form mapping).
func assertProblemViolations(t *testing.T, rec *httptest.ResponseRecorder, wantField, wantCode string) {
	t.Helper()
	var problem struct {
		Code       string `json:"code"`
		Violations []struct {
			Field string `json:"field"`
			Code  string `json:"code"`
		} `json:"violations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("problem body is not JSON: %v (%s)", err, rec.Body.String())
	}
	if problem.Code != "VALIDATION_ERROR" {
		t.Fatalf("problem code %q, want VALIDATION_ERROR (%s)", problem.Code, rec.Body.String())
	}
	enum := map[string]bool{
		"required": true, "too_short": true, "too_long": true, "invalid_format": true,
		"invalid_value": true, "out_of_range": true, "too_many_items": true,
		"duplicate": true, "unknown_field": true,
	}
	found := false
	for _, v := range problem.Violations {
		if !enum[v.Code] {
			t.Fatalf("violation code %q outside the contract enum (%s)", v.Code, rec.Body.String())
		}
		if !strings.HasPrefix(v.Field, "/") {
			t.Fatalf("violation field %q is not a JSON pointer (%s)", v.Field, rec.Body.String())
		}
		if v.Field == wantField && v.Code == wantCode {
			found = true
		}
	}
	if !found {
		t.Fatalf("want violation %s/%s in %+v (%s)", wantField, wantCode, problem.Violations, rec.Body.String())
	}
}
