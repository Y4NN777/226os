package idempotency_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
)

// These tests drive the error and branch handling of PgStore with a scripted
// DB; the SQL itself is proven against real PostgreSQL by
// pgtest/idempotency_integration_test.go.

type rowFn func(dest ...any) error

func (f rowFn) Scan(dest ...any) error { return f(dest...) }

type scriptedDB struct {
	rows  []rowFn // consumed by QueryRow in order
	execs []execResult
	args  [][]any
}

type execResult struct {
	tag pgconn.CommandTag
	err error
}

func (d *scriptedDB) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	d.args = append(d.args, args)
	r := d.rows[0]
	d.rows = d.rows[1:]
	return r
}

func (d *scriptedDB) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	d.args = append(d.args, args)
	r := d.execs[0]
	d.execs = d.execs[1:]
	return r.tag, r.err
}

var (
	noRows   = rowFn(func(...any) error { return pgx.ErrNoRows })
	errBoom  = errors.New("boom")
	now      = time.Date(2026, 11, 12, 10, 0, 0, 123456789, time.UTC)
	someHash = [32]byte{1, 2, 3}
)

// claimed scripts the RETURNING created_at of a successful claim.
func claimed(createdAt time.Time) rowFn {
	return func(dest ...any) error { *(dest[0].(*time.Time)) = createdAt; return nil }
}

// existingRow scripts the SELECT of an existing key.
func existingRow(method, route string, hash [32]byte, state string, status *int16, body []byte) rowFn {
	return func(dest ...any) error {
		*(dest[0].(*string)) = method
		*(dest[1].(*string)) = route
		*(dest[2].(*[]byte)) = hash[:]
		*(dest[3].(*string)) = state
		*(dest[4].(**int16)) = status
		*(dest[5].(*[]byte)) = body
		return nil
	}
}

func ptr[T any](v T) *T { return &v }

func TestPgStoreBegin(t *testing.T) {
	other := [32]byte{9}
	done := func(status int16, body string) rowFn {
		return existingRow("POST", "/r", someHash, "completed", ptr(status), []byte(body))
	}
	tests := []struct {
		name        string
		rows        []rowFn
		want        idempotency.Outcome
		wantAttempt idempotency.Attempt
		wantResp    idempotency.Response
		wantErr     bool
	}{
		{"claimed", []rowFn{claimed(now)}, idempotency.Started, idempotency.Attempt(now.UnixMicro()), idempotency.Response{}, false},
		{"claim error", []rowFn{func(...any) error { return errBoom }}, 0, 0, idempotency.Response{}, true},
		{"different hash", []rowFn{noRows, existingRow("POST", "/r", other, "completed", ptr(int16(201)), nil)}, idempotency.Reused, 0, idempotency.Response{}, false},
		{"different method", []rowFn{noRows, existingRow("PATCH", "/r", someHash, "in_progress", nil, nil)}, idempotency.Reused, 0, idempotency.Response{}, false},
		{"different route", []rowFn{noRows, existingRow("POST", "/other", someHash, "in_progress", nil, nil)}, idempotency.Reused, 0, idempotency.Response{}, false},
		{"in progress", []rowFn{noRows, existingRow("POST", "/r", someHash, "in_progress", nil, nil)}, idempotency.InProgress, 0, idempotency.Response{}, false},
		{"completed without status is not a replay", []rowFn{noRows, existingRow("POST", "/r", someHash, "completed", nil, nil)}, idempotency.InProgress, 0, idempotency.Response{}, false},
		{"replay envelope", []rowFn{noRows, done(201, `{"v": 1, "body": {"id": "q"}, "headers": {"Location": "/api/v1/questions/q", "ETag": "\"x\""}, "content_type": "application/json"}`)},
			idempotency.Replay, 0, idempotency.Response{Status: 201, ContentType: "application/json",
				Headers: map[string]string{"Location": "/api/v1/questions/q", "ETag": `"x"`}, Body: []byte(`{"id": "q"}`)}, false},
		{"replay problem envelope", []rowFn{noRows, done(409, `{"v": 1, "body": {"code": "CONFLICT"}, "content_type": "application/problem+json"}`)},
			idempotency.Replay, 0, idempotency.Response{Status: 409, ContentType: "application/problem+json", Body: []byte(`{"code": "CONFLICT"}`)}, false},
		{"replay body omitted", []rowFn{noRows, done(201, `{"v": 1, "headers": {"Location": "/x"}, "body_omitted": true, "content_type": "application/json"}`)},
			idempotency.Replay, 0, idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"}}, false},
		{"legacy plain object replays as json", []rowFn{noRows, done(201, `{"id": "q", "v": 1}`)},
			idempotency.Replay, 0, idempotency.Response{Status: 201, ContentType: "application/json", Body: []byte(`{"id": "q", "v": 1}`)}, false},
		{"legacy object with a user type field is still json", []rowFn{noRows, done(201, `{"title": "urn:226os:problem:x"}`)},
			idempotency.Replay, 0, idempotency.Response{Status: 201, ContentType: "application/json", Body: []byte(`{"title": "urn:226os:problem:x"}`)}, false},
		{"other envelope version replays as legacy", []rowFn{noRows, done(201, `{"v": 2, "body": {}}`)},
			idempotency.Replay, 0, idempotency.Response{Status: 201, ContentType: "application/json", Body: []byte(`{"v": 2, "body": {}}`)}, false},
		{"legacy NULL body", []rowFn{noRows, existingRow("POST", "/r", someHash, "completed", ptr(int16(204)), nil)},
			idempotency.Replay, 0, idempotency.Response{Status: 204}, false},
		{"read error", []rowFn{noRows, func(...any) error { return errBoom }}, 0, 0, idempotency.Response{}, true},
		{"row vanishes once then is claimed", []rowFn{noRows, noRows, claimed(now)}, idempotency.Started, idempotency.Attempt(now.UnixMicro()), idempotency.Response{}, false},
		{"row keeps vanishing", []rowFn{noRows, noRows, noRows, noRows, noRows, noRows}, 0, 0, idempotency.Response{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, attempt, resp, err := idempotency.NewPgStore(&scriptedDB{rows: tc.rows}).Begin(t.Context(), "u", "k", "POST", "/r", someHash, now)
			if (err != nil) != tc.wantErr || got != tc.want || attempt != tc.wantAttempt {
				t.Fatalf("outcome=%v attempt=%v err=%v, want %v %v (err %v)", got, attempt, err, tc.want, tc.wantAttempt, tc.wantErr)
			}
			if resp.Status != tc.wantResp.Status || resp.ContentType != tc.wantResp.ContentType || string(resp.Body) != string(tc.wantResp.Body) ||
				fmt.Sprint(resp.Headers) != fmt.Sprint(tc.wantResp.Headers) {
				t.Fatalf("response = %+v, want %+v", resp, tc.wantResp)
			}
		})
	}
}

func TestPgStoreBeginPassesTheWindows(t *testing.T) {
	db := &scriptedDB{rows: []rowFn{claimed(now)}}
	if _, _, _, err := idempotency.NewPgStore(db).Begin(t.Context(), "u", "k", "POST", "/r", someHash, now); err != nil {
		t.Fatal(err)
	}
	a := db.args[0]
	if !a[6].(time.Time).Equal(now.Add(24*time.Hour)) || !a[7].(time.Time).Equal(now.Add(-2*time.Minute)) {
		t.Fatalf("expires/abandon args = %v / %v", a[6], a[7])
	}
}

// completeArgs decodes what Complete sends: the full envelope ($4) and the body-less one ($6).
type envelopeView struct {
	V           int               `json:"v"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	Body        json.RawMessage   `json:"body"`
	BodyOmitted bool              `json:"body_omitted"`
}

func decodeEnv(t *testing.T, arg any) envelopeView {
	t.Helper()
	var e envelopeView
	if err := json.Unmarshal(arg.([]byte), &e); err != nil {
		t.Fatalf("envelope %q: %v", arg, err)
	}
	return e
}

func TestPgStoreCompleteWritesAVersionedEnvelope(t *testing.T) {
	ok := execResult{tag: pgconn.NewCommandTag("UPDATE 1")}
	attempt := idempotency.Attempt(now.UnixMicro())
	t.Run("with body", func(t *testing.T) {
		db := &scriptedDB{execs: []execResult{ok}}
		resp := idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"}, Body: []byte(`{"id":"q"}`)}
		if err := idempotency.NewPgStore(db).Complete(t.Context(), "u", "k", attempt, resp, now); err != nil {
			t.Fatal(err)
		}
		a := db.args[0]
		full, omitted := decodeEnv(t, a[3]), decodeEnv(t, a[5])
		if full.V != 1 || full.ContentType != "application/json" || full.Headers["Location"] != "/x" || string(full.Body) != `{"id":"q"}` || full.BodyOmitted {
			t.Fatalf("full envelope = %+v", full)
		}
		if omitted.V != 1 || omitted.Headers["Location"] != "/x" || omitted.Body != nil || !omitted.BodyOmitted {
			t.Fatalf("omitted envelope = %+v", omitted)
		}
		if a[2] != 201 || !a[4].(time.Time).Equal(now) || !a[6].(time.Time).Equal(time.UnixMicro(int64(attempt))) {
			t.Fatalf("args = %v", a)
		}
	})
	t.Run("without body", func(t *testing.T) {
		db := &scriptedDB{execs: []execResult{ok}}
		if err := idempotency.NewPgStore(db).Complete(t.Context(), "u", "k", attempt, idempotency.Response{Status: 204}, now); err != nil {
			t.Fatal(err)
		}
		if e := decodeEnv(t, db.args[0][3]); e.Body != nil || !e.BodyOmitted || e.V != 1 {
			t.Fatalf("envelope = %+v", e)
		}
	})
	for name, body := range map[string]string{"array": `[1,2]`, "scalar": `1`, "null": `null`, "string": `"x"`} {
		t.Run("a non-object body ("+name+") is omitted", func(t *testing.T) {
			db := &scriptedDB{execs: []execResult{ok}}
			if err := idempotency.NewPgStore(db).Complete(t.Context(), "u", "k", attempt, idempotency.Response{Status: 201, Body: []byte(body)}, now); err != nil {
				t.Fatal(err)
			}
			if e := decodeEnv(t, db.args[0][3]); e.Body != nil || !e.BodyOmitted {
				t.Fatalf("envelope = %+v", e)
			}
		})
	}
	t.Run("a header value too long for the column is dropped", func(t *testing.T) {
		db := &scriptedDB{execs: []execResult{ok}}
		resp := idempotency.Response{Status: 201, Headers: map[string]string{"Location": strings.Repeat("a", 4096), "ETag": `"e"`}}
		if err := idempotency.NewPgStore(db).Complete(t.Context(), "u", "k", attempt, resp, now); err != nil {
			t.Fatal(err)
		}
		if e := decodeEnv(t, db.args[0][3]); e.Headers["Location"] != "" || e.Headers["ETag"] != `"e"` {
			t.Fatalf("headers = %v", e.Headers)
		}
	})
	t.Run("a body that is not JSON is an error", func(t *testing.T) {
		db := &scriptedDB{}
		if err := idempotency.NewPgStore(db).Complete(t.Context(), "u", "k", attempt, idempotency.Response{Status: 201, Body: []byte(`{`)}, now); err == nil {
			t.Fatal("want an error")
		}
		if len(db.args) != 0 {
			t.Fatal("nothing must be sent to the database")
		}
	})
}

func TestPgStoreComplete(t *testing.T) {
	pgFail := execResult{err: &pgconn.PgError{Code: "22P05"}}
	ok := execResult{tag: pgconn.NewCommandTag("UPDATE 1")}
	none := execResult{tag: pgconn.NewCommandTag("UPDATE 0")}
	tests := []struct {
		name     string
		body     []byte
		execs    []execResult
		wantErr  error
		wantBody []bool // the envelope sent per Exec call carries the body
	}{
		{"completes", []byte(`{}`), []execResult{ok}, nil, []bool{true}},
		{"lost key", []byte(`{}`), []execResult{none}, idempotency.ErrLostKey, []bool{true}},
		{"refused body is retried without it", []byte(`{}`), []execResult{pgFail, ok}, nil, []bool{true, false}},
		{"retry fails too", []byte(`{}`), []execResult{pgFail, {err: errBoom}}, errBoom, []bool{true, false}},
		{"retry loses the key", []byte(`{}`), []execResult{pgFail, none}, idempotency.ErrLostKey, []bool{true, false}},
		{"driver error is not retried", []byte(`{}`), []execResult{{err: errBoom}}, errBoom, []bool{true}},
		{"database error without body is not retried", nil, []execResult{pgFail}, pgFail.err, []bool{false}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := &scriptedDB{execs: tc.execs}
			err := idempotency.NewPgStore(db).Complete(t.Context(), "u", "k", 7, idempotency.Response{Status: 201, ContentType: "application/json", Body: tc.body}, now)
			if tc.wantErr == nil && err != nil || tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if len(db.args) != len(tc.wantBody) {
				t.Fatalf("exec calls = %d, want %d", len(db.args), len(tc.wantBody))
			}
			for i, withBody := range tc.wantBody {
				if e := decodeEnv(t, db.args[i][3]); (e.Body != nil) != withBody {
					t.Fatalf("call %d: envelope carries body = %v, want %v", i, e.Body != nil, withBody)
				}
			}
		})
	}
}

func TestPgStoreAbandonAndPurge(t *testing.T) {
	t.Run("abandon ok", func(t *testing.T) {
		db := &scriptedDB{execs: []execResult{{tag: pgconn.NewCommandTag("DELETE 1")}}}
		if err := idempotency.NewPgStore(db).Abandon(t.Context(), "u", "k", idempotency.Attempt(now.UnixMicro())); err != nil {
			t.Fatal(err)
		}
		if !db.args[0][2].(time.Time).Equal(time.UnixMicro(now.UnixMicro())) {
			t.Fatalf("attempt arg = %v", db.args[0][2])
		}
	})
	t.Run("abandon by a stale attempt", func(t *testing.T) {
		err := idempotency.NewPgStore(&scriptedDB{execs: []execResult{{tag: pgconn.NewCommandTag("DELETE 0")}}}).Abandon(t.Context(), "u", "k", 1)
		if !errors.Is(err, idempotency.ErrLostKey) {
			t.Fatalf("err = %v, want ErrLostKey", err)
		}
	})
	t.Run("abandon error", func(t *testing.T) {
		if err := idempotency.NewPgStore(&scriptedDB{execs: []execResult{{err: errBoom}}}).Abandon(t.Context(), "u", "k", 1); err == nil {
			t.Fatal("want an error")
		}
	})
	t.Run("purge counts", func(t *testing.T) {
		n, err := idempotency.NewPgStore(&scriptedDB{execs: []execResult{{tag: pgconn.NewCommandTag("DELETE 4")}}}).Purge(t.Context(), now)
		if err != nil || n != 4 {
			t.Fatalf("n=%d err=%v", n, err)
		}
	})
	t.Run("purge error", func(t *testing.T) {
		if _, err := idempotency.NewPgStore(&scriptedDB{execs: []execResult{{err: errBoom}}}).Purge(t.Context(), now); err == nil {
			t.Fatal("want an error")
		}
	})
}

// A value PostgreSQL refuses can sit in a header (a NUL in Location), so the body-less envelope
// that keeps the headers can fail too: the last resort carries no header at all.
func TestPgStoreCompleteFallsBackToAnEnvelopeWithoutHeaders(t *testing.T) {
	pgFail := execResult{err: &pgconn.PgError{Code: "22P05"}}
	ok := execResult{tag: pgconn.NewCommandTag("UPDATE 1")}
	resp := idempotency.Response{Status: 201, ContentType: "application/json", Headers: map[string]string{"Location": "/x"}, Body: []byte(`{"id":"q"}`)}
	tests := []struct {
		name    string
		execs   []execResult
		wantErr bool
	}{
		{"headers dropped on the third try", []execResult{pgFail, pgFail, ok}, false},
		{"minimal envelope refused too", []execResult{pgFail, pgFail, pgFail}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := &scriptedDB{execs: tc.execs}
			err := idempotency.NewPgStore(db).Complete(t.Context(), "u", "k", 7, resp, now)
			if (err != nil) != tc.wantErr || len(db.args) != 3 {
				t.Fatalf("err = %v, exec calls = %d, want 3", err, len(db.args))
			}
			full, omitted, minimal := decodeEnv(t, db.args[0][3]), decodeEnv(t, db.args[1][3]), decodeEnv(t, db.args[2][3])
			if full.Body == nil || omitted.Body != nil || omitted.Headers["Location"] != "/x" || !omitted.BodyOmitted {
				t.Fatalf("tiers 1 and 2: %+v %+v", full, omitted)
			}
			if minimal.V != 1 || minimal.ContentType != "application/json" || !minimal.BodyOmitted || minimal.Body != nil || len(minimal.Headers) != 0 {
				t.Fatalf("minimal envelope = %+v", minimal)
			}
		})
	}
}
