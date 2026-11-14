//go:build integration

package pgtest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Y4NN777/226os/src/api/internal/app"
	"github.com/Y4NN777/226os/src/api/internal/discover"
	discoverservice "github.com/Y4NN777/226os/src/api/internal/discover/service"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	"github.com/Y4NN777/226os/src/api/internal/learn"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
	"github.com/Y4NN777/226os/src/api/internal/platform/ratelimit"
	"github.com/Y4NN777/226os/src/api/internal/search"
	searchservice "github.com/Y4NN777/226os/src/api/internal/search/service"
)

var s2Secret = []byte("0123456789abcdef0123456789abcdef")

// limiter is the port both modules use; the stack lets a test replace it to inject failures.
type limiter interface {
	discoverservice.Limiter
	searchservice.Limiter
}

// s2Stack is the production router with identity, the catalogue, error reports, search and Learn over a
// migrated, catalogue-seeded database and a pooled connection (the tests run requests concurrently).
type s2Stack struct {
	router http.Handler
	pool   *pgxpool.Pool
	conn   *pgx.Conn
	mail   *mailer.Fake
}

func newS2Stack(ctx context.Context, t *testing.T, override limiter) *s2Stack {
	t.Helper()
	conn := migratedDB(ctx, t)
	seedCatalogue(ctx, t, conn)
	if _, err := conn.Exec(ctx, `INSERT INTO resource_sources (id, resource_id, kind, title, url, checked_at) VALUES
		('`+catSrcBeta+`', '`+catBeta+`', 'official_docs', 'Beta docs', 'https://docs.example/beta', now())`); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, conn.Config().ConnString())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	mail := mailer.NewFake()
	svc := service.New(repository.New(pool), mail, s2Secret)
	codec, err := cursor.New(s2Secret)
	if err != nil {
		t.Fatal(err)
	}
	var lim limiter = ratelimit.New(pool, s2Secret, nil)
	if override != nil {
		lim = override
	}
	catalogue := discover.NewHandler(pool, codec, lim, nil)
	deps := httpx.Deps{
		Server: app.NewServer(
			app.WithIdentity(identitytransport.New(svc, false, nil)),
			app.WithCatalogue(catalogue),
			app.WithErrorReports(catalogue),
			app.WithSearch(search.NewHandler(pool, lim, nil)),
			app.WithLearn(learn.NewHandler(pool, codec, lim, nil)),
		),
		Ready: pool,
		// As in production (app.newDeps): the six bound operations, createLearnArticle among them, honour Idempotency-Key.
		Idempotency: idempotency.NewPgStore(pool),
	}
	deps.Subjects = identitytransport.NewResolver(svc, false)
	deps.CSRF = identitytransport.CSRFMiddleware(svc, false, nil)
	return &s2Stack{
		router: httpx.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), deps),
		pool:   pool, conn: conn, mail: mail,
	}
}

const catSrcBeta = "019a0000-0000-7000-8000-0000000000c4"

func (s *s2Stack) do(t *testing.T, method, path, body string, headers map[string]string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
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

// preSession returns the anonymous CSRF cookie and token.
func (s *s2Stack) preSession(t *testing.T) (*http.Cookie, string) {
	t.Helper()
	rec := s.do(t, http.MethodGet, "/api/v1/auth/csrf", "", nil)
	var body struct {
		CsrfToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.CsrfToken == "" {
		t.Fatalf("csrf %d %s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_csrf" {
			return c, body.CsrfToken
		}
	}
	t.Fatal("no pre-session cookie")
	return nil, ""
}

// verificationToken waits for the single verification mail sent to to.
func (s *s2Stack) verificationToken(t *testing.T, to string) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, m := range s.mail.Sent() {
			if len(m.To) == 1 && m.To[0] == to {
				for _, line := range strings.Split(m.Text, "\n") {
					if line = strings.TrimSpace(line); len(line) >= 40 && !strings.ContainsAny(line, " :") {
						return line
					}
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no verification mail for %s", to)
	return ""
}

type member struct {
	session *http.Cookie
	csrf    string
	id      string
}

// signIn creates a member through the real sign-up, optional e-mail verification and sign-in. Each
// member comes from its own network (X-Forwarded-For) so the identity IP limits never interfere.
func (s *s2Stack) signIn(t *testing.T, name string, verified bool, network int) member {
	t.Helper()
	email := name + "@example.com"
	xff := map[string]string{"X-Forwarded-For": fmt.Sprintf("198.51.%d.9", network)}
	with := func(token string) map[string]string {
		return map[string]string{"X-CSRF-Token": token, "X-Forwarded-For": xff["X-Forwarded-For"]}
	}
	cookie, token := s.preSession(t)
	rec := s.do(t, http.MethodPost, "/api/v1/auth/sign-up",
		`{"email":"`+email+`","password":"correct-horse-battery","username":"`+name+`","display_name":"Test User"}`, with(token), cookie)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("sign-up %d %s", rec.Code, rec.Body.String())
	}
	if verified {
		rec = s.do(t, http.MethodPost, "/api/v1/auth/verify-email", `{"token":"`+s.verificationToken(t, email)+`"}`, with(token), cookie)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("verify %d %s", rec.Code, rec.Body.String())
		}
	}
	rec = s.do(t, http.MethodPost, "/api/v1/auth/sign-in", `{"email":"`+email+`","password":"correct-horse-battery"}`, with(token), cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("sign-in %d %s", rec.Code, rec.Body.String())
	}
	var state struct {
		CsrfToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	m := member{csrf: state.CsrfToken}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "226os_session" {
			m.session = c
		}
	}
	if m.session == nil {
		t.Fatal("no session cookie")
	}
	if err := s.conn.QueryRow(t.Context(), `SELECT id::text FROM users WHERE email = $1`, email).Scan(&m.id); err != nil {
		t.Fatal(err)
	}
	return m
}

// report posts an error report as m (with its session CSRF token), or anonymously when m is nil.
func (s *s2Stack) report(t *testing.T, m *member, resource, body string) *httptest.ResponseRecorder {
	t.Helper()
	path := "/api/v1/resources/" + resource + "/error-reports"
	if m == nil {
		cookie, token := s.preSession(t)
		return s.do(t, http.MethodPost, path, body, map[string]string{"X-CSRF-Token": token}, cookie)
	}
	return s.do(t, http.MethodPost, path, body, map[string]string{"X-CSRF-Token": m.csrf}, m.session)
}

func (s *s2Stack) reportRows(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM error_reports WHERE resource_id = $1`, catAlpha).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// problem decodes a problem+json body and drops the per-request members.
func problem(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type %q, body %s", ct, rec.Body.String())
	}
	var p map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	delete(p, "request_id")
	delete(p, "instance")
	return p
}

func assertProblem(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) map[string]any {
	t.Helper()
	p := problem(t, rec)
	if rec.Code != status || p["code"] != code {
		t.Fatalf("status %d code %v, want %d %s: %s", rec.Code, p["code"], status, code, rec.Body.String())
	}
	return p
}

const goodReport = `{"kind":"outdated","body":"The pricing page moved."}`

// awaitFreshMinute avoids a fixed one-minute window rolling over in the middle of a rate limit test.
func awaitFreshMinute(t *testing.T) {
	t.Helper()
	if rem := time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)); rem < 8*time.Second {
		time.Sleep(rem + 200*time.Millisecond)
	}
}

func TestErrorReportAuthorizationMatrix(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	verified := s.signIn(t, "verified", true, 1)
	unverified := s.signIn(t, "unverified", false, 2)

	t.Run("anonymous is 401", func(t *testing.T) {
		assertProblem(t, s.report(t, nil, catAlpha, goodReport), http.StatusUnauthorized, "UNAUTHENTICATED")
	})
	t.Run("member without a verified email is 403", func(t *testing.T) {
		assertProblem(t, s.report(t, &unverified, catAlpha, goodReport), http.StatusForbidden, "EMAIL_NOT_VERIFIED")
	})
	t.Run("missing CSRF token is 403 CSRF_FAILED", func(t *testing.T) {
		rec := s.do(t, http.MethodPost, "/api/v1/resources/"+catAlpha+"/error-reports", goodReport, nil, verified.session)
		assertProblem(t, rec, http.StatusForbidden, "CSRF_FAILED")
	})
	t.Run("wrong CSRF token is 403 CSRF_FAILED", func(t *testing.T) {
		rec := s.do(t, http.MethodPost, "/api/v1/resources/"+catAlpha+"/error-reports", goodReport,
			map[string]string{"X-CSRF-Token": "not-the-token"}, verified.session)
		assertProblem(t, rec, http.StatusForbidden, "CSRF_FAILED")
	})
	if n := s.reportRows(t); n != 0 {
		t.Fatalf("%d reports stored by denied requests", n)
	}
}

func TestErrorReportCreateAndDuplicate(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "alice", true, 1)
	bob := s.signIn(t, "bob", true, 2)

	rec := s.report(t, &alice, catAlpha, `{"kind":"outdated","body":"The pricing page moved.","source_id":"`+catSrcDocs+`"}`)
	if rec.Code != http.StatusCreated || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if len(created) != 5 || created["resource_id"] != catAlpha || created["kind"] != "outdated" || created["status"] != "received" {
		t.Fatalf("response %v: want exactly id, resource_id, kind, status, created_at", created)
	}
	if strings.Contains(rec.Body.String(), alice.id) || strings.Contains(rec.Body.String(), "reporter") {
		t.Fatalf("the response exposes the reporter: %s", rec.Body.String())
	}
	var reporter, message, source string
	if err := s.conn.QueryRow(ctx, `SELECT reporter_id::text, message, source_id::text FROM error_reports WHERE id = $1`, created["id"]).Scan(&reporter, &message, &source); err != nil {
		t.Fatal(err)
	}
	if reporter != alice.id || message != "The pricing page moved." || source != catSrcDocs {
		t.Fatalf("stored reporter %s message %q source %s", reporter, message, source)
	}

	// Duplicate: same resource, reporter and kind while the report is open -> the SAME report, no new row.
	for i := range 2 {
		dup := s.report(t, &alice, catAlpha, `{"kind":"outdated","body":"Said again."}`)
		var again map[string]any
		_ = json.Unmarshal(dup.Body.Bytes(), &again)
		if dup.Code != http.StatusCreated || again["id"] != created["id"] {
			t.Fatalf("duplicate %d: %d %s, want 201 with id %v", i, dup.Code, dup.Body.String(), created["id"])
		}
	}
	if n := s.reportRows(t); n != 1 {
		t.Fatalf("rows = %d, want 1 after duplicates", n)
	}

	// A triaged report is still open.
	if _, err := s.conn.Exec(ctx, `UPDATE error_reports SET status = 'triaged'`); err != nil {
		t.Fatal(err)
	}
	var triaged map[string]any
	rec = s.report(t, &alice, catAlpha, goodReport)
	_ = json.Unmarshal(rec.Body.Bytes(), &triaged)
	if rec.Code != http.StatusCreated || triaged["id"] != created["id"] || triaged["status"] != "triaged" || s.reportRows(t) != 1 {
		t.Fatalf("triaged duplicate: %d %s", rec.Code, rec.Body.String())
	}

	// Another kind, and another reporter, are new reports.
	if rec = s.report(t, &alice, catAlpha, `{"kind":"broken_link","body":"x"}`); rec.Code != http.StatusCreated {
		t.Fatalf("other kind %d %s", rec.Code, rec.Body.String())
	}
	var bobs map[string]any
	rec = s.report(t, &bob, catAlpha, goodReport)
	_ = json.Unmarshal(rec.Body.Bytes(), &bobs)
	if rec.Code != http.StatusCreated || bobs["id"] == created["id"] || s.reportRows(t) != 3 {
		t.Fatalf("other reporter: %d %s rows %d", rec.Code, rec.Body.String(), s.reportRows(t))
	}

	// After the moderator resolves it, the same member may report the same kind again.
	if _, err := s.conn.Exec(ctx, `UPDATE error_reports SET status = 'resolved', handled_by = $1, handled_at = now() WHERE id = $2`, bob.id, created["id"]); err != nil {
		t.Fatal(err)
	}
	var next map[string]any
	rec = s.report(t, &alice, catAlpha, goodReport)
	_ = json.Unmarshal(rec.Body.Bytes(), &next)
	if rec.Code != http.StatusCreated || next["id"] == created["id"] || next["status"] != "received" || s.reportRows(t) != 4 {
		t.Fatalf("after resolution: %d %s rows %d", rec.Code, rec.Body.String(), s.reportRows(t))
	}
}

// Two simultaneous identical requests must yield one row and one id (unique index + ON CONFLICT).
func TestErrorReportConcurrentDuplicatesCreateOneRow(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "alice", true, 1)

	const n = 12
	ids := make([]string, n)
	codes := make([]int, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range n {
		wg.Go(func() {
			<-start
			rec := s.report(t, &alice, catAlpha, goodReport)
			var body struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal(rec.Body.Bytes(), &body)
			ids[i], codes[i] = body.ID, rec.Code
		})
	}
	close(start)
	wg.Wait()
	for i := range n {
		if codes[i] != http.StatusCreated || ids[i] == "" || ids[i] != ids[0] {
			t.Fatalf("request %d: %d id %q, want 201 and id %q", i, codes[i], ids[i], ids[0])
		}
	}
	if rows := s.reportRows(t); rows != 1 {
		t.Fatalf("rows = %d, want 1", rows)
	}
}

func TestErrorReportInvisibleResourcesAreIdentical404(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "alice", true, 1)

	var want map[string]any
	for _, tc := range []struct{ name, id string }{
		{"unknown id", "019a0000-0000-7000-8000-00000000ffff"},
		{"removed resource", catRemoved},
		{"resource without a provider", catOrphan},
		{"resource of a removed provider", catSecret},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := assertProblem(t, s.report(t, &alice, tc.id, goodReport), http.StatusNotFound, "NOT_FOUND")
			if want == nil {
				want = p
			}
			if fmt.Sprint(p) != fmt.Sprint(want) {
				t.Fatalf("body %v differs from the unknown-id body %v", p, want)
			}
		})
	}
	var n int
	if err := s.conn.QueryRow(ctx, `SELECT count(*) FROM error_reports`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("reports = %d, %v; want none", n, err)
	}
}

func TestErrorReportValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "alice", true, 1)

	cases := []struct {
		name, body string
		status     int
		code       string
	}{
		{"invalid kind", `{"kind":"spam","body":"x"}`, 422, "VALIDATION_ERROR"},
		{"missing kind", `{"body":"x"}`, 422, "VALIDATION_ERROR"},
		{"empty body", `{"kind":"other","body":""}`, 422, "VALIDATION_ERROR"},
		{"blank body", `{"kind":"other","body":"   "}`, 422, "VALIDATION_ERROR"},
		{"body of 2001 characters", `{"kind":"other","body":"` + strings.Repeat("a", 2001) + `"}`, 422, "VALIDATION_ERROR"},
		{"unknown field", `{"kind":"other","body":"x","reporter_id":"` + alice.id + `"}`, 422, "VALIDATION_ERROR"},
		{"malformed source id", `{"kind":"other","body":"x","source_id":"nope"}`, 422, "VALIDATION_ERROR"},
		{"malformed JSON", `{"kind":`, 400, "MALFORMED_REQUEST"},
		{"source of another resource", `{"kind":"other","body":"x","source_id":"` + catSrcBeta + `"}`, 422, "VALIDATION_ERROR"},
		{"source that does not exist", `{"kind":"other","body":"x","source_id":"019a0000-0000-7000-8000-00000000fffe"}`, 422, "VALIDATION_ERROR"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := s.report(t, &alice, catAlpha, tc.body)
			assertProblem(t, rec, tc.status, tc.code)
		})
	}

	t.Run("a foreign source is a violation on source_id that names no other resource", func(t *testing.T) {
		rec := s.report(t, &alice, catAlpha, `{"kind":"other","body":"x","source_id":"`+catSrcBeta+`"}`)
		p := assertProblem(t, rec, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		v, _ := json.Marshal(p["violations"])
		if !strings.Contains(string(v), `"/source_id"`) || !strings.Contains(string(v), `"invalid_value"`) {
			t.Fatalf("violations %s", v)
		}
		for _, leak := range []string{catSrcBeta, catBeta, "Beta"} {
			if strings.Contains(rec.Body.String(), leak) {
				t.Fatalf("response leaks %q: %s", leak, rec.Body.String())
			}
		}
	})
	if n := s.reportRows(t); n != 0 {
		t.Fatalf("%d reports stored by invalid requests", n)
	}
}

func TestErrorReportRateLimitIs60PerUserPerMinute(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "alice", true, 1)
	bob := s.signIn(t, "bob", true, 2)
	awaitFreshMinute(t)

	for i := 1; i <= 60; i++ {
		if rec := s.report(t, &alice, catAlpha, goodReport); rec.Code != http.StatusCreated {
			t.Fatalf("request %d: %d %s", i, rec.Code, rec.Body.String())
		}
	}
	rec := s.report(t, &alice, catAlpha, goodReport)
	assertProblem(t, rec, http.StatusTooManyRequests, "RATE_LIMITED")
	if ra := rec.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Fatalf("Retry-After = %q", ra)
	}
	// The key is the user: another member is not limited.
	if rec := s.report(t, &bob, catAlpha, goodReport); rec.Code != http.StatusCreated {
		t.Fatalf("other member: %d %s", rec.Code, rec.Body.String())
	}
	// No user id or address is stored in clear: the bucket is an HMAC.
	var plain int
	if err := s.conn.QueryRow(ctx, `SELECT count(*) FROM rate_limit_buckets WHERE scope = 'write:user' AND position($1::bytea in key_hash) > 0`, []byte(alice.id)).Scan(&plain); err != nil || plain != 0 {
		t.Fatalf("plain user id in a bucket key: %d, %v", plain, err)
	}
}

type failingLimiter struct{}

func (failingLimiter) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return false, 0, fmt.Errorf("limiter store down")
}

// The write limit fails CLOSED (500, nothing stored); the anonymous search limit fails OPEN (200).
func TestLimiterFailureClosedForReportsOpenForSearch(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, failingLimiter{})
	alice := s.signIn(t, "alice", true, 1)

	assertProblem(t, s.report(t, &alice, catAlpha, goodReport), http.StatusInternalServerError, "INTERNAL_ERROR")
	if n := s.reportRows(t); n != 0 {
		t.Fatalf("%d reports stored while the limiter was down", n)
	}
	rec := s.do(t, http.MethodGet, "/api/v1/search?q=pay", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("search with the limiter down: %d %s", rec.Code, rec.Body.String())
	}
}

func TestAnonymousSearchRateLimit(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "alice", true, 1)
	awaitFreshMinute(t)

	search := func(xff string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		return s.do(t, http.MethodGet, "/api/v1/search?q=pay", "", map[string]string{"X-Forwarded-For": xff}, cookies...)
	}
	// The leading entries are client supplied and ignored: only the right-most address (the one the
	// proxy saw) is the identity, so rotating them does not reset the bucket.
	for i := 1; i <= 60; i++ {
		if rec := search(fmt.Sprintf("10.9.9.%d, 203.0.113.%d", i, i)); rec.Code != http.StatusOK {
			t.Fatalf("request %d: %d %s", i, rec.Code, rec.Body.String())
		}
	}
	rec := search("10.1.1.1, 203.0.113.250")
	assertProblem(t, rec, http.StatusTooManyRequests, "RATE_LIMITED")
	if ra := rec.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Fatalf("Retry-After = %q", ra)
	}
	// Another network is a different bucket.
	if rec := search("198.51.100.77"); rec.Code != http.StatusOK {
		t.Fatalf("other network: %d %s", rec.Code, rec.Body.String())
	}
	// A member on the exhausted network is never limited.
	for i := range 65 {
		if rec := search("203.0.113.1", alice.session); rec.Code != http.StatusOK {
			t.Fatalf("member request %d: %d %s", i, rec.Code, rec.Body.String())
		}
	}
}

// unlimited lets every request through, for tests that exercise the read path, not the limit.
type unlimited struct{}

func (unlimited) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}
