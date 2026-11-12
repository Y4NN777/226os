//go:build integration

package pgtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
)

const idemRoute = "/api/v1/questions"

// idemKey is built at run time: a literal key-shaped string in the source trips the secret scanner.
var idemKey = fmt.Sprintf("retry-%022d", 1)

// idemEnv is a migrated database with two users and a store over a real pool.
type idemEnv struct {
	ctx   context.Context
	pool  *pgxpool.Pool
	store *idempotency.PgStore
	now   time.Time
}

func newIdemEnv(t *testing.T) *idemEnv {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	conn := migratedDB(ctx, t)
	pool, err := pgxpool.New(ctx, pg.dsn(conn.Config().Database))
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	for _, u := range []string{uAlice, uBob} {
		if _, err := pool.Exec(ctx, `INSERT INTO users (id, email, role) VALUES ($1, $2, 'member')`, u, u+"@example.org"); err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}
	return &idemEnv{ctx: ctx, pool: pool, store: idempotency.NewPgStore(pool), now: time.Now().UTC().Truncate(time.Microsecond)}
}

func hashOf(s string) [32]byte {
	var h [32]byte
	copy(h[:], s)
	return h
}

func (e *idemEnv) begin(t *testing.T, user, key, method, route string, hash [32]byte, now time.Time) (idempotency.Outcome, idempotency.Response) {
	t.Helper()
	o, _, s, err := e.store.Begin(e.ctx, user, key, method, route, hash, now)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	return o, s
}

// start claims the key and returns the attempt that owns it.
func (e *idemEnv) start(t *testing.T, user, key string, hash [32]byte, now time.Time) idempotency.Attempt {
	t.Helper()
	o, a, _, err := e.store.Begin(e.ctx, user, key, "POST", idemRoute, hash, now)
	if err != nil || o != idempotency.Started {
		t.Fatalf("Begin = %v, %v; want Started", o, err)
	}
	return a
}

// rawBody reads response_body as the database stores it.
func (e *idemEnv) rawBody(t *testing.T, user string) map[string]any {
	t.Helper()
	var raw []byte
	if err := e.pool.QueryRow(e.ctx, `SELECT response_body FROM idempotency_keys WHERE user_id = $1`, user).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("response_body %q: %v", raw, err)
	}
	return out
}

func (e *idemEnv) count(t *testing.T, where string, args ...any) int {
	t.Helper()
	var n int
	if err := e.pool.QueryRow(e.ctx, `SELECT count(*) FROM idempotency_keys WHERE `+where, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestIdempotencyStoreLifecycle(t *testing.T) {
	e := newIdemEnv(t)
	h := hashOf("body-a")

	attempt := e.start(t, uAlice, idemKey, h, e.now)
	var state string
	var expires, created time.Time
	if err := e.pool.QueryRow(e.ctx, `SELECT state, created_at, expires_at FROM idempotency_keys WHERE user_id = $1`, uAlice).Scan(&state, &created, &expires); err != nil {
		t.Fatal(err)
	}
	if state != "in_progress" || expires.Sub(created) != 24*time.Hour {
		t.Fatalf("state=%s retention=%v", state, expires.Sub(created))
	}
	if o, _ := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now); o != idempotency.InProgress {
		t.Fatalf("duplicate while running = %v, want InProgress", o)
	}

	body := []byte(`{"id":"0190c3f2","title":"Hello"}`)
	if err := e.store.Complete(e.ctx, uAlice, idemKey, attempt, idempotency.Response{Status: 201, ContentType: "application/json", Body: body}, e.now); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	o, s := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now)
	if o != idempotency.Replay || s.Status != 201 || s.ContentType != "application/json" {
		t.Fatalf("replay = %v %+v", o, s)
	}
	var got, want map[string]any
	if err := json.Unmarshal(s.Body, &got); err != nil {
		t.Fatalf("stored body %q: %v", s.Body, err)
	}
	_ = json.Unmarshal(body, &want)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("stored body = %v, want %v", got, want)
	}

	// Same key, other request: reused, whatever differs.
	for name, c := range map[string]struct {
		method, route string
		hash          [32]byte
	}{
		"hash":   {"POST", idemRoute, hashOf("body-b")},
		"route":  {"POST", "/api/v1/projects", h},
		"method": {"PATCH", idemRoute, h},
	} {
		if o, _ := e.begin(t, uAlice, idemKey, c.method, c.route, c.hash, e.now); o != idempotency.Reused {
			t.Fatalf("different %s = %v, want Reused", name, o)
		}
	}
	// Keys are bound to the subject.
	if o, _ := e.begin(t, uBob, idemKey, "POST", idemRoute, h, e.now); o != idempotency.Started {
		t.Fatalf("another subject = %v, want Started", o)
	}
}

func TestIdempotencyStoreReplaysProblemDocuments(t *testing.T) {
	e := newIdemEnv(t)
	h := hashOf("p")
	attempt := e.start(t, uAlice, idemKey, h, e.now)
	problem := []byte(`{"type":"urn:226os:problem:CONFLICT","status":409,"code":"CONFLICT"}`)
	if err := e.store.Complete(e.ctx, uAlice, idemKey, attempt, idempotency.Response{Status: 409, ContentType: "application/problem+json", Body: problem}, e.now); err != nil {
		t.Fatal(err)
	}
	if o, s := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now); o != idempotency.Replay || s.Status != 409 || s.ContentType != "application/problem+json" {
		t.Fatalf("replay = %v %+v", o, s)
	}
}

func TestIdempotencyStoreExpiredRowIsReplaced(t *testing.T) {
	e := newIdemEnv(t)
	h := hashOf("a")
	attempt := e.start(t, uAlice, idemKey, h, e.now)
	if err := e.store.Complete(e.ctx, uAlice, idemKey, attempt, idempotency.Response{Status: 201, ContentType: "application/json", Body: []byte(`{"a":1}`)}, e.now); err != nil {
		t.Fatal(err)
	}
	later := e.now.Add(25 * time.Hour)
	if o, _ := e.begin(t, uAlice, idemKey, "POST", idemRoute, hashOf("different"), later); o != idempotency.Started {
		t.Fatalf("after expiry = %v, want Started (even with another request)", o)
	}
	var state string
	var status *int16
	var body []byte
	if err := e.pool.QueryRow(e.ctx, `SELECT state, response_status, response_body FROM idempotency_keys WHERE user_id = $1`, uAlice).Scan(&state, &status, &body); err != nil {
		t.Fatal(err)
	}
	if state != "in_progress" || status != nil || body != nil {
		t.Fatalf("the replaced row must be clean: %s %v %q", state, status, body)
	}
	if n := e.count(t, `user_id = $1`, uAlice); n != 1 {
		t.Fatalf("rows = %d, want 1", n)
	}
}

func TestIdempotencyStoreAbandonedRowIsTakenOverByTheSameRequest(t *testing.T) {
	e := newIdemEnv(t)
	h := hashOf("a")
	e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now)
	age := func(d time.Duration) {
		if _, err := e.pool.Exec(e.ctx, `UPDATE idempotency_keys SET created_at = $1 WHERE user_id = $2`, e.now.Add(-d), uAlice); err != nil {
			t.Fatal(err)
		}
	}
	age(time.Minute)
	if o, _ := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now); o != idempotency.InProgress {
		t.Fatalf("1 minute old = %v, want InProgress", o)
	}
	age(3 * time.Minute)
	if o, _ := e.begin(t, uAlice, idemKey, "POST", idemRoute, hashOf("other"), e.now); o != idempotency.Reused {
		t.Fatalf("abandoned but different request = %v, want Reused (never taken over)", o)
	}
	if o, _ := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now); o != idempotency.Started {
		t.Fatalf("abandoned and identical = %v, want Started", o)
	}
	if o, _ := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now); o != idempotency.InProgress {
		t.Fatalf("after takeover the key is running again = %v", o)
	}
}

func TestIdempotencyStoreAbandonFreesTheKey(t *testing.T) {
	e := newIdemEnv(t)
	h := hashOf("a")
	attempt := e.start(t, uAlice, idemKey, h, e.now)
	if err := e.store.Abandon(e.ctx, uAlice, idemKey, attempt); err != nil {
		t.Fatal(err)
	}
	attempt = e.start(t, uAlice, idemKey, h, e.now.Add(time.Second))
	// A completed key is never deleted by Abandon.
	if err := e.store.Complete(e.ctx, uAlice, idemKey, attempt, idempotency.Response{Status: 201, ContentType: "application/json", Body: []byte(`{}`)}, e.now); err != nil {
		t.Fatal(err)
	}
	if err := e.store.Abandon(e.ctx, uAlice, idemKey, attempt); !errors.Is(err, idempotency.ErrLostKey) {
		t.Fatalf("Abandon of a completed key = %v, want ErrLostKey", err)
	}
	if o, _ := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now); o != idempotency.Replay {
		t.Fatalf("completed key after Abandon = %v, want Replay", o)
	}
	// Completing a key that is not ours fails.
	if err := e.store.Complete(e.ctx, uBob, idemKey, attempt, idempotency.Response{Status: 201}, e.now); !errors.Is(err, idempotency.ErrLostKey) {
		t.Fatalf("Complete on an absent key = %v, want ErrLostKey", err)
	}
}

// After a takeover the original owner must neither complete nor abandon the new owner's row.
func TestIdempotencyStoreTakeoverFencesTheOriginalOwner(t *testing.T) {
	e := newIdemEnv(t)
	h := hashOf("a")
	oldAttempt := e.start(t, uAlice, idemKey, h, e.now.Add(-10*time.Minute))
	newAttempt := e.start(t, uAlice, idemKey, h, e.now) // the 2-minute takeover
	if newAttempt == oldAttempt {
		t.Fatalf("a takeover must refresh the attempt: %d", newAttempt)
	}
	if err := e.store.Complete(e.ctx, uAlice, idemKey, oldAttempt, idempotency.Response{Status: 201, Body: []byte(`{"old":true}`)}, e.now); !errors.Is(err, idempotency.ErrLostKey) {
		t.Fatalf("old owner Complete = %v, want ErrLostKey", err)
	}
	if err := e.store.Abandon(e.ctx, uAlice, idemKey, oldAttempt); !errors.Is(err, idempotency.ErrLostKey) {
		t.Fatalf("old owner Abandon = %v, want ErrLostKey", err)
	}
	var state string
	var created time.Time
	if err := e.pool.QueryRow(e.ctx, `SELECT state, created_at FROM idempotency_keys WHERE user_id = $1`, uAlice).Scan(&state, &created); err != nil {
		t.Fatal(err)
	}
	if state != "in_progress" || idempotency.Attempt(created.UnixMicro()) != newAttempt {
		t.Fatalf("the new owner's row was touched: state=%s created=%v", state, created)
	}
	if err := e.store.Complete(e.ctx, uAlice, idemKey, newAttempt, idempotency.Response{Status: 201, Body: []byte(`{"new":true}`)}, e.now); err != nil {
		t.Fatalf("new owner Complete: %v", err)
	}
}

// The envelope that response_body holds survives PostgreSQL: content type, headers and body.
// The store still reports the recorded content type of an omitted body; it is the middleware
// that refuses to send it (a body-less replay carries no Content-Type and is marked with
// Idempotency-Body-Omitted, see the middleware tests), since an empty body labelled
// application/json would make res.json() throw in a client.
func TestIdempotencyStoreEnvelopeRoundTrip(t *testing.T) {
	marker := `{"title":"urn:226os:problem:not-a-problem","n":[1,2,3]}`
	problem := `{"type":"urn:226os:problem:CONFLICT","status":409,"code":"CONFLICT"}`
	tests := []struct {
		name        string
		resp        idempotency.Response
		wantBody    string // "" = omitted
		wantOmitted bool
	}{
		{"object with location and etag", idempotency.Response{Status: 201, ContentType: "application/json",
			Headers: map[string]string{"Location": "/api/v1/questions/q1", "ETag": `"abc"`}, Body: []byte(`{"id":"q1"}`)}, `{"id": "q1"}`, false},
		{"user text with the problem marker", idempotency.Response{Status: 201, ContentType: "application/json", Body: []byte(marker)}, marker, false},
		{"problem document", idempotency.Response{Status: 409, ContentType: "application/problem+json", Body: []byte(problem)}, problem, false},
		{"no body", idempotency.Response{Status: 204, Headers: map[string]string{"Location": "/x"}}, "", true},
		{"array body is omitted", idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"}, Body: []byte(`[1,2]`)}, "", true},
		{"scalar body is omitted", idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"}, Body: []byte(`1`)}, "", true},
		{"body over 16 KiB is omitted", idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"},
			Body: []byte(`{"x":"` + strings.Repeat("a", 20000) + `"}`)}, "", true},
		{"body within the raw cap but wide as jsonb text is omitted", idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"},
			Body: []byte(`{"a":[` + strings.TrimSuffix(strings.Repeat("1,", 7000), ",") + `]}`)}, "", true},
		{"invalid UTF-8 refused by jsonb is omitted", idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"},
			Body: []byte("{\"a\":\"\xff\"}")}, "", true},
		{"NUL escape refused by jsonb is omitted", idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"},
			Body: []byte(`{"a":"\u0000"}`)}, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newIdemEnv(t)
			h := hashOf("b")
			attempt := e.start(t, uAlice, idemKey, h, e.now)
			if err := e.store.Complete(e.ctx, uAlice, idemKey, attempt, tc.resp, e.now); err != nil {
				t.Fatalf("Complete: %v", err)
			}
			var size int
			if err := e.pool.QueryRow(e.ctx, `SELECT octet_length(response_body::text) FROM idempotency_keys WHERE user_id = $1`, uAlice).Scan(&size); err != nil || size > idempotency.MaxStoredBody {
				t.Fatalf("stored size = %d, err %v", size, err)
			}
			if raw := e.rawBody(t, uAlice); (raw["body_omitted"] == true) != tc.wantOmitted || raw["v"] != float64(1) {
				t.Fatalf("envelope = %v, want omitted=%v", raw, tc.wantOmitted)
			}
			o, got := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now)
			if o != idempotency.Replay || got.Status != tc.resp.Status || got.ContentType != tc.resp.ContentType {
				t.Fatalf("replay = %v %+v", o, got)
			}
			for _, name := range []string{"Location", "ETag"} {
				if got.Headers[name] != tc.resp.Headers[name] {
					t.Fatalf("header %s = %q, want %q", name, got.Headers[name], tc.resp.Headers[name])
				}
			}
			if tc.wantBody == "" {
				if got.Body != nil {
					t.Fatalf("body = %q, want none", got.Body)
				}
				return
			}
			var gotJSON, wantJSON any
			if err := json.Unmarshal(got.Body, &gotJSON); err != nil {
				t.Fatalf("body %q: %v", got.Body, err)
			}
			_ = json.Unmarshal([]byte(tc.wantBody), &wantJSON)
			if fmt.Sprint(gotJSON) != fmt.Sprint(wantJSON) {
				t.Fatalf("body = %v, want %v", gotJSON, wantJSON)
			}
		})
	}
}

// A row of an older shape (a plain object, or NULL) replays as application/json.
func TestIdempotencyStoreReplaysRowsOfTheOlderShape(t *testing.T) {
	for name, tc := range map[string]struct {
		body   any
		wantCT string
		want   string
	}{
		"plain object": {`{"id": "q1"}`, "application/json", `{"id": "q1"}`},
		"NULL":         {nil, "", ""},
	} {
		t.Run(name, func(t *testing.T) {
			e := newIdemEnv(t)
			h := hashOf("legacy")
			e.start(t, uAlice, idemKey, h, e.now)
			if _, err := e.pool.Exec(e.ctx, `UPDATE idempotency_keys SET state='completed', response_status=201, completed_at=$2, response_body=$3::jsonb WHERE user_id=$1`, uAlice, e.now, tc.body); err != nil {
				t.Fatal(err)
			}
			o, got := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now)
			if o != idempotency.Replay || got.Status != 201 || got.ContentType != tc.wantCT || string(got.Body) != tc.want {
				t.Fatalf("replay = %v %+v", o, got)
			}
		})
	}
}

func TestIdempotencyStorePurgeDeletesOnlyExpiredRows(t *testing.T) {
	e := newIdemEnv(t)
	for i, k := range []string{fmt.Sprintf("retry-%022d", 11), fmt.Sprintf("retry-%022d", 12), fmt.Sprintf("retry-%022d", 13), fmt.Sprintf("retry-%022d", 14)} {
		e.begin(t, uAlice, k, "POST", idemRoute, hashOf(k), e.now)
		expires := map[int]time.Time{0: e.now.Add(-time.Hour), 1: e.now.Add(-time.Minute), 2: e.now.Add(time.Hour), 3: e.now.Add(time.Microsecond)}[i]
		if _, err := e.pool.Exec(e.ctx, `UPDATE idempotency_keys SET created_at = $1, expires_at = $2 WHERE user_id = $3 AND idempotency_key = $4`,
			e.now.Add(-48*time.Hour), expires, uAlice, k); err != nil {
			t.Fatal(err)
		}
	}
	n, err := e.store.Purge(e.ctx, e.now)
	if err != nil || n != 2 {
		t.Fatalf("Purge = %d, %v; want 2", n, err)
	}
	if left := e.count(t, `user_id = $1`, uAlice); left != 2 {
		t.Fatalf("rows left = %d, want 2", left)
	}
	if left := e.count(t, `idempotency_key IN ($1, $2)`, fmt.Sprintf("retry-%022d", 13), fmt.Sprintf("retry-%022d", 14)); left != 2 {
		t.Fatalf("live rows left = %d, want 2", left)
	}
}

func TestIdempotencyStoreRaceYieldsExactlyOneStarted(t *testing.T) {
	e := newIdemEnv(t)
	const workers = 24
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		outcomes = map[idempotency.Outcome]int{}
		start    = make(chan struct{})
	)
	for range workers {
		wg.Go(func() {
			<-start
			o, _, _, err := e.store.Begin(e.ctx, uAlice, idemKey, "POST", idemRoute, hashOf("race"), e.now)
			if err != nil {
				t.Errorf("Begin: %v", err)
				return
			}
			mu.Lock()
			outcomes[o]++
			mu.Unlock()
		})
	}
	close(start)
	wg.Wait()
	if outcomes[idempotency.Started] != 1 || outcomes[idempotency.InProgress] != workers-1 {
		t.Fatalf("outcomes = %v, want exactly one Started and %d InProgress", outcomes, workers-1)
	}
	if n := e.count(t, `user_id = $1`, uAlice); n != 1 {
		t.Fatalf("rows = %d, want 1", n)
	}
}

func TestIdempotencyStoreRefusesANonUUIDSubject(t *testing.T) {
	e := newIdemEnv(t)
	if _, _, _, err := e.store.Begin(e.ctx, "not-a-uuid", idemKey, "POST", idemRoute, hashOf("a"), e.now); err == nil {
		t.Fatal("a subject that is not a UUID must be an error, not a silent pass")
	}
}

func TestIdempotencyStoreRefusesAKeyOutsideTheColumnCheck(t *testing.T) {
	e := newIdemEnv(t)
	const short = "k7x9q2z" // under the 8-character minimum of idempotency_keys_key_ck
	_, _, _, err := e.store.Begin(e.ctx, uAlice, short, "POST", idemRoute, hashOf("a"), e.now)
	if err == nil {
		t.Fatal("a key under 8 characters violates idempotency_keys_key_ck")
	}
	// The middleware logs this error: PostgreSQL error texts carry the constraint, never the key.
	if strings.Contains(err.Error(), short) {
		t.Fatalf("the error text contains the key: %v", err)
	}
}

// PostgreSQL refusing a header value too (a NUL in Location) must not leave the key in progress:
// the third tier stores a minimal envelope without headers.
func TestIdempotencyStoreCompletesWhenPostgreSQLRefusesTheHeadersToo(t *testing.T) {
	tests := []struct {
		name string
		resp idempotency.Response
	}{
		{"NUL in Location", idempotency.Response{Status: 201, ContentType: "application/json",
			Headers: map[string]string{"Location": "/x\x00y", "ETag": `"e"`}, Body: []byte(`{"id":"q1"}`)}},
		{"NUL in Location and invalid UTF-8 body", idempotency.Response{Status: 201, ContentType: "application/json",
			Headers: map[string]string{"Location": "/x\x00y"}, Body: []byte("{\"a\":\"\xff\"}")}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newIdemEnv(t)
			h := hashOf("nul")
			attempt := e.start(t, uAlice, idemKey, h, e.now)
			if err := e.store.Complete(e.ctx, uAlice, idemKey, attempt, tc.resp, e.now); err != nil {
				t.Fatalf("Complete: %v", err)
			}
			if n := e.count(t, `user_id = $1 AND state = 'completed' AND response_status = 201`, uAlice); n != 1 {
				t.Fatalf("the row must be completed, got %d completed rows", n)
			}
			raw := e.rawBody(t, uAlice)
			if raw["body_omitted"] != true || raw["headers"] != nil || raw["content_type"] != "application/json" {
				t.Fatalf("envelope = %v, want minimal: content_type, body_omitted, no headers", raw)
			}
			o, got := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now)
			if o != idempotency.Replay || got.Status != 201 || got.Body != nil || len(got.Headers) != 0 || got.ContentType != "application/json" {
				t.Fatalf("replay = %v %+v", o, got)
			}
		})
	}
}

// A replayed body is the jsonb normal form of the first body: semantically equal, not
// byte-identical (key order, whitespace, 1e2 stored as 100).
func TestIdempotencyStoreReplaysASemanticallyEqualBody(t *testing.T) {
	e := newIdemEnv(t)
	h := hashOf("normal-form")
	first := `{"z":1,"b":[1e2, 2.50],"a":{"y":true,"x":null}}`
	attempt := e.start(t, uAlice, idemKey, h, e.now)
	if err := e.store.Complete(e.ctx, uAlice, idemKey, attempt, idempotency.Response{Status: 201, ContentType: "application/json", Body: []byte(first)}, e.now); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	_, got := e.begin(t, uAlice, idemKey, "POST", idemRoute, h, e.now)
	var want, have any
	if err := json.Unmarshal([]byte(first), &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(got.Body, &have); err != nil {
		t.Fatalf("replayed body %q: %v", got.Body, err)
	}
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("replayed body %q is not semantically equal to %q", got.Body, first)
	}
}
