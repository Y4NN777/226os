package ratelimit

import (
	"bytes"
	"context"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

var errBoom = errors.New("boom: postgres://user:secret@host/db")

type fakeQueries struct {
	hits     []int32
	hitErr   error
	purgeErr error

	hitArgs   []dbgen.HitRateBucketParams
	purgeArgs []pgtype.Timestamptz
	ctxHasDL  []bool
}

func (f *fakeQueries) HitRateBucket(ctx context.Context, a dbgen.HitRateBucketParams) (int32, error) {
	_, dl := ctx.Deadline()
	f.ctxHasDL = append(f.ctxHasDL, dl)
	f.hitArgs = append(f.hitArgs, a)
	if f.hitErr != nil {
		return 0, f.hitErr
	}
	h := f.hits[0]
	f.hits = f.hits[1:]
	return h, nil
}

func (f *fakeQueries) PurgeRateBuckets(ctx context.Context, cutoff pgtype.Timestamptz) error {
	_, dl := ctx.Deadline()
	f.ctxHasDL = append(f.ctxHasDL, dl)
	f.purgeArgs = append(f.purgeArgs, cutoff)
	return f.purgeErr
}

var t0 = time.Date(2026, 11, 13, 10, 0, 20, 0, time.UTC)

func secret() []byte { return []byte("0123456789abcdef0123456789ABCDEF") }

func newTest(q *fakeQueries, logs *bytes.Buffer) *Limiter {
	l := newLimiter(q, secret(), slog.New(slog.NewTextHandler(logs, nil)))
	l.now = func() time.Time { return t0 }
	return l
}

func TestHitFixedWindow(t *testing.T) {
	cases := []struct {
		name      string
		hits      int32
		wantOK    bool
		wantRetry time.Duration
	}{
		{"first hit allowed", 1, true, 0},
		{"at the limit allowed", 3, true, 0},
		{"over the limit denied with the time left in the window", 4, false, 40 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := &fakeQueries{hits: []int32{tc.hits}}
			ok, retry, err := newTest(q, &bytes.Buffer{}).Hit(t.Context(), "search:ip", "203.0.113.0/24", 3, time.Minute)
			if err != nil || ok != tc.wantOK || retry != tc.wantRetry {
				t.Fatalf("Hit = %v, %v, %v; want %v, %v, nil", ok, retry, err, tc.wantOK, tc.wantRetry)
			}
			if got, want := q.hitArgs[0].WindowStart.Time, t0.Truncate(time.Minute); !got.Equal(want) {
				t.Fatalf("window_start = %v, want the aligned window %v", got, want)
			}
			if q.hitArgs[0].Scope != "search:ip" {
				t.Fatalf("scope = %q", q.hitArgs[0].Scope)
			}
			for _, dl := range q.ctxHasDL {
				if !dl {
					t.Fatal("a database call ran without a deadline")
				}
			}
		})
	}
}

func TestHitPurgesOnlyWhenABucketIsCreated(t *testing.T) {
	q := &fakeQueries{hits: []int32{1, 2}}
	l := newTest(q, &bytes.Buffer{})
	for range 2 {
		if _, _, err := l.Hit(t.Context(), "search:ip", "x", 60, time.Minute); err != nil {
			t.Fatal(err)
		}
	}
	if len(q.purgeArgs) != 1 {
		t.Fatalf("purges = %d, want 1 (only the hit that created the bucket)", len(q.purgeArgs))
	}
	if got, want := q.purgeArgs[0].Time, t0.Add(-retention); !got.Equal(want) {
		t.Fatalf("purge cutoff = %v, want %v", got, want)
	}
}

func TestPurgeFailureIsLoggedAndDoesNotFailTheHit(t *testing.T) {
	q := &fakeQueries{hits: []int32{1}, purgeErr: errBoom}
	var logs bytes.Buffer
	ok, _, err := newTest(q, &logs).Hit(t.Context(), "search:ip", "x", 60, time.Minute)
	if err != nil || !ok {
		t.Fatalf("Hit = %v, %v; a purge failure must not fail the hit", ok, err)
	}
	if !strings.Contains(logs.String(), "rate limit purge failed") {
		t.Fatalf("purge failure not logged: %q", logs.String())
	}
	if strings.Contains(logs.String(), "secret") {
		t.Fatalf("log leaks the driver error text: %q", logs.String())
	}
}

func TestHitErrorsAreReturned(t *testing.T) {
	q := &fakeQueries{hitErr: errBoom}
	ok, _, err := newTest(q, &bytes.Buffer{}).Hit(t.Context(), "search:ip", "x", 60, time.Minute)
	if ok || !errors.Is(err, errBoom) {
		t.Fatalf("Hit = %v, %v; want denied and the wrapped cause", ok, err)
	}
}

func TestHitRejectsInvalidParameters(t *testing.T) {
	cases := []struct {
		name   string
		limit  int
		window time.Duration
	}{
		{"zero limit", 0, time.Minute},
		{"zero window", 1, 0},
		{"negative window", 1, -time.Second},
		{"window longer than the purge retention allows", 1, time.Hour + time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := &fakeQueries{}
			ok, _, err := newTest(q, &bytes.Buffer{}).Hit(t.Context(), "search:ip", "x", tc.limit, tc.window)
			if ok || err == nil {
				t.Fatalf("Hit = %v, %v; want an error", ok, err)
			}
			if len(q.hitArgs) != 0 {
				t.Fatal("an invalid call reached the database")
			}
		})
	}
}

func TestBucketKeyUsesAnHKDFSubkeyNotTheRawSecret(t *testing.T) {
	q := &fakeQueries{hits: []int32{1}}
	l := newTest(q, &bytes.Buffer{})
	if _, _, err := l.Hit(t.Context(), "search:ip", "203.0.113.0/24", 60, time.Minute); err != nil {
		t.Fatal(err)
	}
	got := q.hitArgs[0].KeyHash

	// Independent derivation: HKDF-Expand(secret, info, 32), then HMAC over scope|value.
	sub, err := hkdf.Expand(sha256.New, secret(), "226os/ratelimit/bucket", 32)
	if err != nil {
		t.Fatal(err)
	}
	m := hmac.New(sha256.New, sub)
	m.Write([]byte("search:ip|203.0.113.0/24"))
	if want := m.Sum(nil); !bytes.Equal(got, want) {
		t.Fatalf("key = %x, want HMAC(HKDF subkey) %x", got, want)
	}
	raw := hmac.New(sha256.New, secret())
	raw.Write([]byte("search:ip|203.0.113.0/24"))
	if bytes.Equal(got, raw.Sum(nil)) {
		t.Fatal("bucket key is HMAC under the raw secret")
	}
}

func TestBucketKeySeparatesScopesAndValues(t *testing.T) {
	l := newLimiter(&fakeQueries{}, secret(), nil)
	a := l.bucketKey("search:ip", "1.2.3.0/24")
	if bytes.Equal(a, l.bucketKey("error_report:user", "1.2.3.0/24")) || bytes.Equal(a, l.bucketKey("search:ip", "1.2.4.0/24")) {
		t.Fatal("scope and value must both change the bucket key")
	}
	if !bytes.Equal(a, l.bucketKey("search:ip", "1.2.3.0/24")) {
		t.Fatal("bucket key is not deterministic")
	}
}

func TestNew(t *testing.T) {
	l := New(nil, secret(), nil)
	if l == nil || l.now == nil || l.logger == nil {
		t.Fatalf("New = %+v", l)
	}
}
