package dbgen_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the sqlc-generated wiring (this file is hand-written; the
// generated files are never edited). It pins, per query, the SQL text that is
// sent, the order of the bind arguments and the order of the scan targets
// against the column list, using a fake DBTX. Behaviour against real
// PostgreSQL is covered by the integration tests in the pgtest package.

type call struct {
	sql  string
	args []any
}

type fakeDB struct {
	calls   []call
	rowVals []any // values the fake row scans into the destinations, in order
	rowErr  error
	execErr error
}

func (f *fakeDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	f.calls = append(f.calls, call{sql, args})
	return pgconn.NewCommandTag("DELETE 3"), f.execErr
}

func (f *fakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("not used by the generated queries under test")
}

func (f *fakeDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	f.calls = append(f.calls, call{sql, args})
	return fakeRow{vals: f.rowVals, err: f.rowErr}
}

type fakeRow struct {
	vals []any
	err  error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.vals) {
		return errors.New("scan target count differs from the fake row column count")
	}
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(r.vals[i]))
	}
	return nil
}

// fakeTx is a pgx.Tx that only forwards the DBTX methods to the wrapped fake.
type fakeTx struct {
	pgx.Tx
	db *fakeDB
}

func (t fakeTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.db.Exec(ctx, sql, args...)
}

func (t fakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.db.Query(ctx, sql, args...)
}

func (t fakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.db.QueryRow(ctx, sql, args...)
}

func TestGetCountry(t *testing.T) {
	db := &fakeDB{rowVals: []any{"BF", "Burkina Faso", "Burkina Faso"}}
	got, err := dbgen.New(db).GetCountry(t.Context(), "BF")
	if err != nil {
		t.Fatal(err)
	}
	if want := (dbgen.Country{Code: "BF", NameFr: "Burkina Faso", NameEn: "Burkina Faso"}); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	assertCall(t, db, "FROM countries WHERE code = $1", []any{"BF"})
}

func TestGetCountryPropagatesScanError(t *testing.T) {
	errBoom := errors.New("boom")
	_, err := dbgen.New(&fakeDB{rowErr: errBoom}).GetCountry(t.Context(), "BF")
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
}

func TestGetUserByEmail(t *testing.T) {
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	verified := pgtype.Timestamptz{Time: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC), Valid: true}
	created := pgtype.Timestamptz{Time: time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC), Valid: true}
	// Column order of the generated SELECT: id, email, email_verified_at, role, status, created_at.
	db := &fakeDB{rowVals: []any{id, "alice@example.org", verified, "member", "active", created}}

	got, err := dbgen.New(db).GetUserByEmail(t.Context(), "alice@example.org")
	if err != nil {
		t.Fatal(err)
	}
	want := dbgen.GetUserByEmailRow{ID: id, Email: "alice@example.org", EmailVerifiedAt: verified, Role: "member", Status: "active", CreatedAt: created}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	assertCall(t, db, "FROM users WHERE email = $1", []any{"alice@example.org"})
	if strings.Contains(db.calls[0].sql, "password") {
		t.Fatal("the identity query must never select credentials")
	}
}

func TestDeleteExpiredIdempotencyKeys(t *testing.T) {
	cutoff := pgtype.Timestamptz{Time: time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC), Valid: true}
	db := &fakeDB{}
	if err := dbgen.New(db).DeleteExpiredIdempotencyKeys(t.Context(), cutoff); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "DELETE FROM idempotency_keys WHERE expires_at < $1", []any{cutoff})

	errBoom := errors.New("boom")
	if err := dbgen.New(&fakeDB{execErr: errBoom}).DeleteExpiredIdempotencyKeys(t.Context(), cutoff); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
}

func TestWithTxRunsQueriesOnTheTransaction(t *testing.T) {
	outer, inner := &fakeDB{}, &fakeDB{rowVals: []any{"CI", "Côte d’Ivoire", "Côte d’Ivoire"}}
	q := dbgen.New(outer).WithTx(fakeTx{db: inner})
	if _, err := q.GetCountry(t.Context(), "CI"); err != nil {
		t.Fatal(err)
	}
	if len(outer.calls) != 0 || len(inner.calls) != 1 {
		t.Fatalf("calls: outer=%d inner=%d; want the query to run on the transaction only", len(outer.calls), len(inner.calls))
	}
}

func assertCall(t *testing.T, db *fakeDB, sqlFragment string, args []any) {
	t.Helper()
	if len(db.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(db.calls))
	}
	c := db.calls[0]
	if !strings.Contains(c.sql, sqlFragment) {
		t.Errorf("sql %q does not contain %q", c.sql, sqlFragment)
	}
	if !reflect.DeepEqual(c.args, args) {
		t.Errorf("args = %#v, want %#v", c.args, args)
	}
}
