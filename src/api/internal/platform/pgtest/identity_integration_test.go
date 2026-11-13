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
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

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
	conn   *pgx.Conn
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
