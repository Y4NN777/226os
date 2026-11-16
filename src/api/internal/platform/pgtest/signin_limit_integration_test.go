//go:build integration

package pgtest

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Y4NN777/226os/src/api/internal/app"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

// newIdentityPoolStack is newIdentityStack over a connection pool, so requests
// can really run concurrently (a single pgx.Conn serializes them).
func newIdentityPoolStack(ctx context.Context, t *testing.T) *identityStack {
	t.Helper()
	conn := migratedDB(ctx, t)
	t.Cleanup(func() { _ = conn.Close(ctx) })
	pool, err := pgxpool.New(ctx, conn.Config().ConnString())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	mail := mailer.NewFake()
	svc := service.New(repository.New(pool), mail, []byte("0123456789abcdef0123456789abcdef"),
		service.WithBackground(func(f func()) { f() }))
	deps := httpx.Deps{Server: app.NewServer(app.WithIdentity(identitytransport.New(svc, false, nil))), Ready: pool}
	deps.Subjects = identitytransport.NewResolver(svc, false)
	deps.CSRF = identitytransport.CSRFMiddleware(svc, false, nil)
	return &identityStack{
		router: httpx.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), deps),
		mail:   mail,
		conn:   conn,
	}
}

// signInFrom posts a sign-in from the /24 of 10.<n>.0.1 so the per-IP limit
// (a different bucket per /24) never interferes with the per-account one.
func (s *identityStack) signInFrom(t *testing.T, n int, email, pw string, c *http.Cookie, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	return s.do(t, http.MethodPost, "/api/v1/auth/sign-in",
		`{"email":"`+email+`","password":"`+pw+`"}`,
		map[string]string{"X-CSRF-Token": csrf, "X-Forwarded-For": fmt.Sprintf("10.%d.0.1", n)}, c)
}

func accountBucketHits(ctx context.Context, t *testing.T, s *identityStack) (rows, hits int) {
	t.Helper()
	if err := s.conn.QueryRow(ctx,
		`SELECT count(*), COALESCE(sum(hits), 0) FROM rate_limit_buckets WHERE scope = 'signin:account'`).Scan(&rows, &hits); err != nil {
		t.Fatal(err)
	}
	return rows, hits
}

// Issue #97: the per-account limit must hold under concurrency. 20 parallel
// wrong-password sign-ins at one account (distinct client IPs) may let exactly
// 5 reach the password check (INVALID_CREDENTIALS); the other 15 are 429.
func TestIdentitySignInAccountLimitHoldsUnderConcurrency(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 180*time.Second)
	defer cancel()
	s := newIdentityPoolStack(ctx, t)
	s.signUp(t, "race@example.com", "raceuser")
	cookie, csrf := s.csrfToken(t)

	const attempts = 20
	codes := make([]int, attempts)
	retry := make([]string, attempts)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			rec := s.signInFrom(t, i+1, "race@example.com", "wrong-password-123", cookie, csrf)
			codes[i], retry[i] = rec.Code, rec.Header().Get("Retry-After")
		}()
	}
	close(start)
	wg.Wait()

	var denied, limited int
	for i, c := range codes {
		switch c {
		case http.StatusUnauthorized:
			denied++
		case http.StatusTooManyRequests:
			limited++
			if retry[i] == "" {
				t.Fatalf("attempt %d: 429 without Retry-After", i)
			}
		default:
			t.Fatalf("attempt %d: unexpected status %d", i, c)
		}
	}
	if denied != 5 || limited != attempts-5 {
		t.Fatalf("password checks reached = %d (want 5), 429 = %d (want %d)", denied, limited, attempts-5)
	}
}

// A success gives its up-front hit back: only failures stay counted, so 4
// failures, a success, then failures still allow the documented 5 in total.
func TestIdentitySignInSuccessGivesTheHitBack(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	s := newIdentityPoolStack(ctx, t)
	s.signUp(t, "giveback@example.com", "givebackuser")
	cookie, csrf := s.csrfToken(t)
	in := func(n int, pw string) int {
		return s.signInFrom(t, n, "giveback@example.com", pw, cookie, csrf).Code
	}

	// A lone success leaves no trace: the row is removed (hits must stay > 0).
	if c := in(1, "correct-horse-battery"); c != http.StatusOK {
		t.Fatalf("success: %d", c)
	}
	if rows, hits := accountBucketHits(ctx, t, s); rows != 0 || hits != 0 {
		t.Fatalf("after a lone success: rows=%d hits=%d, want 0 and 0", rows, hits)
	}

	for i := 0; i < 4; i++ {
		if c := in(2, "wrong-password-123"); c != http.StatusUnauthorized {
			t.Fatalf("failure %d: %d", i+1, c)
		}
	}
	if c := in(3, "correct-horse-battery"); c != http.StatusOK {
		t.Fatalf("success after 4 failures: %d", c)
	}
	if _, hits := accountBucketHits(ctx, t, s); hits != 4 {
		t.Fatalf("after 4 failures and a success: hits=%d, want 4", hits)
	}
	if c := in(4, "wrong-password-123"); c != http.StatusUnauthorized {
		t.Fatalf("5th failure: %d", c)
	}
	if c := in(5, "wrong-password-123"); c != http.StatusTooManyRequests {
		t.Fatalf("6th attempt: %d, want 429", c)
	}
}
