//go:build integration

package pgtest

import (
	"bytes"
	"context"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
	"github.com/Y4NN777/226os/src/api/internal/platform/ratelimit"
)

func limiterPool(ctx context.Context, t *testing.T) (*pgxpool.Pool, *pgx.Conn) {
	t.Helper()
	conn := migratedDB(ctx, t)
	pool, err := pgxpool.New(ctx, conn.Config().ConnString())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool, conn
}

func TestRateLimiterFixedWindowAndRetryAfter(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	pool, conn := limiterPool(ctx, t)
	l := ratelimit.New(pool, s2Secret, nil)
	awaitFreshMinute(t)

	for i := 1; i <= 3; i++ {
		ok, retry, err := l.Hit(ctx, "search:ip", "203.0.113.0/24", 3, time.Minute)
		if err != nil || !ok || retry != 0 {
			t.Fatalf("hit %d = %v, %v, %v; want allowed", i, ok, retry, err)
		}
	}
	ok, retry, err := l.Hit(ctx, "search:ip", "203.0.113.0/24", 3, time.Minute)
	if err != nil || ok || retry <= 0 || retry > time.Minute {
		t.Fatalf("4th hit = %v, %v, %v; want denied with 0 < retry <= 1m", ok, retry, err)
	}
	// Other values and other scopes have their own buckets.
	for _, c := range [][2]string{{"search:ip", "198.51.100.0/24"}, {"error_report:user", "203.0.113.0/24"}} {
		if ok, _, err := l.Hit(ctx, c[0], c[1], 3, time.Minute); err != nil || !ok {
			t.Fatalf("%v = %v, %v; want its own bucket", c, ok, err)
		}
	}

	var hits int
	var start time.Time
	if err := conn.QueryRow(ctx, `SELECT hits, window_start FROM rate_limit_buckets WHERE scope = 'search:ip' ORDER BY hits DESC LIMIT 1`).Scan(&hits, &start); err != nil {
		t.Fatal(err)
	}
	if hits != 4 || !start.Equal(start.Truncate(time.Minute)) || time.Since(start) > time.Minute+time.Second {
		t.Fatalf("bucket hits %d window_start %v; want 4 hits in the current aligned minute", hits, start)
	}

	// A bucket of an earlier window does not count toward the current one.
	if _, err := conn.Exec(ctx, `INSERT INTO rate_limit_buckets (scope, key_hash, window_start, hits) VALUES ('search:ip', $1, $2, 1000)`,
		hmacKey("search:ip", "192.0.2.0/24"), start.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if ok, _, err := l.Hit(ctx, "search:ip", "192.0.2.0/24", 3, time.Minute); err != nil || !ok {
		t.Fatalf("hit after an exhausted earlier window = %v, %v; want allowed", ok, err)
	}
}

// hmacKey is the documented derivation, computed independently of the package.
func hmacKey(scope, value string) []byte {
	const info = "226os/ratelimit/bucket"
	sub, err := hkdf.Expand(sha256.New, s2Secret, info, 32)
	if err != nil {
		panic(err)
	}
	m := hmac.New(sha256.New, sub)
	m.Write([]byte(scope + "|" + value))
	return m.Sum(nil)
}

func TestRateLimiterStoresAnHKDFSubkeyHMACNotTheRawSecret(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	pool, conn := limiterPool(ctx, t)
	if _, _, err := ratelimit.New(pool, s2Secret, nil).Hit(ctx, "search:ip", "203.0.113.0/24", 5, time.Minute); err != nil {
		t.Fatal(err)
	}
	var key []byte
	if err := conn.QueryRow(ctx, `SELECT key_hash FROM rate_limit_buckets WHERE scope = 'search:ip'`).Scan(&key); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(key, hmacKey("search:ip", "203.0.113.0/24")) {
		t.Fatalf("key_hash = %x, want HMAC under the HKDF subkey %x", key, hmacKey("search:ip", "203.0.113.0/24"))
	}
	raw := hmac.New(sha256.New, s2Secret)
	raw.Write([]byte("search:ip|203.0.113.0/24"))
	if bytes.Equal(key, raw.Sum(nil)) {
		t.Fatal("key_hash is an HMAC under the raw secret")
	}
	if bytes.Contains(key, []byte("203.0.113")) {
		t.Fatal("the address is stored in clear")
	}
}

func TestRateLimiterIsExactUnderConcurrency(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	pool, _ := limiterPool(ctx, t)
	l := ratelimit.New(pool, s2Secret, nil)
	awaitFreshMinute(t)

	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 40 {
		wg.Go(func() {
			ok, _, err := l.Hit(ctx, "search:ip", "203.0.113.0/24", 10, time.Minute)
			if err != nil {
				t.Error(err)
			}
			if ok {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	if allowed.Load() != 10 {
		t.Fatalf("allowed = %d of 40, want exactly the limit (10)", allowed.Load())
	}
}

func TestRateLimiterPurgesExpiredBucketsInBoundedBatches(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	pool, conn := limiterPool(ctx, t)
	if _, err := conn.Exec(ctx, `
		INSERT INTO rate_limit_buckets (scope, key_hash, window_start)
		SELECT 'search_old', decode(lpad(to_hex(g), 64, '0'), 'hex'), now() - interval '5 hours' FROM generate_series(1, 600) g`); err != nil {
		t.Fatal(err)
	}
	// A bucket of a window that ended 90 minutes ago is inside the 2 h retention and must survive.
	if _, err := conn.Exec(ctx, `INSERT INTO rate_limit_buckets (scope, key_hash, window_start) VALUES ('search_recent', decode(repeat('aa', 32), 'hex'), now() - interval '90 minutes')`); err != nil {
		t.Fatal(err)
	}
	count := func(scope string) (n int) {
		if err := conn.QueryRow(ctx, `SELECT count(*) FROM rate_limit_buckets WHERE scope = $1`, scope).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	l := ratelimit.New(pool, s2Secret, nil)
	// The hit that creates a bucket purges one batch of at most 500 rows.
	if _, _, err := l.Hit(ctx, "search:ip", "a", 5, time.Minute); err != nil {
		t.Fatal(err)
	}
	if got := count("search_old"); got != 100 {
		t.Fatalf("expired buckets left = %d, want 100 (600 minus one batch of 500)", got)
	}
	if got := count("search_recent"); got != 1 {
		t.Fatalf("a bucket within the retention was purged")
	}
	// A hit on an existing bucket does not purge.
	if _, _, err := l.Hit(ctx, "search:ip", "a", 5, time.Minute); err != nil {
		t.Fatal(err)
	}
	if got := count("search_old"); got != 100 {
		t.Fatalf("a repeat hit purged: %d left", got)
	}
}

// failPurge fails the purge DELETE only, like a lock timeout on the housekeeping statement.
type failPurge struct{ dbgen.DBTX }

func (f failPurge) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "DELETE FROM rate_limit_buckets") {
		return pgconn.CommandTag{}, errors.New("canceling statement due to lock timeout (postgres://u:pw@db)")
	}
	return f.DBTX.Exec(ctx, sql, args...)
}

func TestRateLimiterPurgeFailureIsLoggedAndDoesNotFailTheHit(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	pool, conn := limiterPool(ctx, t)
	var logs bytes.Buffer
	l := ratelimit.New(failPurge{pool}, s2Secret, slog.New(slog.NewTextHandler(&logs, nil)))

	ok, _, err := l.Hit(ctx, "search:ip", "203.0.113.0/24", 5, time.Minute)
	if err != nil || !ok {
		t.Fatalf("Hit = %v, %v; a purge failure must not fail the hit", ok, err)
	}
	var hits int
	if err := conn.QueryRow(ctx, `SELECT hits FROM rate_limit_buckets WHERE scope = 'search:ip'`).Scan(&hits); err != nil || hits != 1 {
		t.Fatalf("hit not recorded: %d, %v", hits, err)
	}
	if !strings.Contains(logs.String(), "rate limit purge failed") || strings.Contains(logs.String(), "pw") {
		t.Fatalf("log = %q; want the failure logged without the driver text", logs.String())
	}
}

func TestRateLimiterScopesPassTheTableCheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	pool, _ := limiterPool(ctx, t)
	l := ratelimit.New(pool, s2Secret, nil)
	for _, scope := range []string{"search:ip", "error_report:user"} {
		if _, _, err := l.Hit(ctx, scope, "x", 1, time.Minute); err != nil {
			t.Fatalf("scope %q: %v", scope, err)
		}
	}
	// A scope outside the CHECK is a database error, reported, never a silent allow.
	ok, _, err := l.Hit(ctx, "Bad Scope", "x", 1, time.Minute)
	if ok || err == nil || !strings.Contains(fmt.Sprint(err), "rate limit hit") {
		t.Fatalf("Hit with an invalid scope = %v, %v; want denied with an error", ok, err)
	}
}
