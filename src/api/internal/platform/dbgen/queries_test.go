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
	many    [][]any
	manyErr error
	scanErr error
	rowsErr error
}

func (f *fakeDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	f.calls = append(f.calls, call{sql, args})
	return pgconn.NewCommandTag("DELETE 3"), f.execErr
}

func (f *fakeDB) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	f.calls = append(f.calls, call{sql, args})
	if f.manyErr != nil {
		return nil, f.manyErr
	}
	return &fakeRows{rows: f.many, scanErr: f.scanErr, err: f.rowsErr}, nil
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

type fakeRows struct {
	rows    [][]any
	i       int
	scanErr error
	err     error
}

func (r *fakeRows) Close()                                       {}
func (r *fakeRows) Err() error                                   { return r.err }
func (r *fakeRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeRows) Values() ([]any, error)                       { return nil, nil }
func (r *fakeRows) RawValues() [][]byte                          { return nil }
func (r *fakeRows) Conn() *pgx.Conn                              { return nil }
func (r *fakeRows) TypeMap() *pgtype.Map                         { return nil }
func (r *fakeRows) Next() bool {
	if r.i < len(r.rows) {
		r.i++
		return true
	}
	return false
}
func (r *fakeRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	vals := r.rows[r.i-1]
	if len(dest) != len(vals) {
		return errors.New("scan width")
	}
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(vals[i]))
	}
	return nil
}

func TestCatalogueQueriesScanAndReject(t *testing.T) {
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	ts := pgtype.Timestamptz{Time: time.Unix(10, 0).UTC(), Valid: true}
	text := pgtype.Text{String: "https://example.test", Valid: true}
	listRow := []any{id, "alpha", "Alpha", "api", "mobile-money", "Paiement", []string{"fr"}, text, "verified", ts, ts, id, "acme", "Acme", text, int32(1)}
	db := &fakeDB{many: [][]any{listRow}}
	rows, err := dbgen.New(db).ListCatalogue(t.Context(), dbgen.ListCatalogueParams{RowLimit: 21, Q: text})
	if err != nil || len(rows) != 1 || rows[0].Slug != "alpha" {
		t.Fatalf("rows %+v err %v", rows, err)
	}
	if !strings.Contains(db.calls[0].sql, "LIMIT") {
		t.Fatal("list query is not bounded")
	}
	got, err := dbgen.New(&fakeDB{rowVals: listRow}).GetCatalogueEntry(t.Context(), id)
	if err != nil || got.Slug != "alpha" {
		t.Fatalf("%+v %v", got, err)
	}
	countries, err := dbgen.New(&fakeDB{many: [][]any{{id, "BF"}}}).CatalogueCountries(t.Context(), []pgtype.UUID{id})
	if err != nil || countries[0].CountryCode != "BF" {
		t.Fatalf("%+v %v", countries, err)
	}
	sdks, err := dbgen.New(&fakeDB{many: [][]any{{id, "go", "Go", text}}}).CatalogueSDKs(t.Context(), []pgtype.UUID{id})
	if err != nil || sdks[0].Language != "go" {
		t.Fatalf("%+v %v", sdks, err)
	}
	related, err := dbgen.New(&fakeDB{many: [][]any{{id, "project", id, "A project", "fr"}}}).CatalogueRelated(t.Context(), []pgtype.UUID{id})
	if err != nil || related[0].Title != "A project" {
		t.Fatalf("%+v %v", related, err)
	}
	discussions, err := dbgen.New(&fakeDB{many: [][]any{{id, id, "A question", "fr"}}}).CatalogueDiscussions(t.Context(), []pgtype.UUID{id})
	if err != nil || discussions[0].Title != "A question" {
		t.Fatalf("%+v %v", discussions, err)
	}
	sources, err := dbgen.New(&fakeDB{many: [][]any{{id, "claim", "A claim", "https://claim.example", ts, "verified", id}}}).ListCatalogueSources(t.Context(), id)
	if err != nil || sources[0].Kind != "claim" {
		t.Fatalf("%+v %v", sources, err)
	}

	boom := errors.New("db down")
	calls := []func(*fakeDB) error{
		func(db *fakeDB) error {
			_, err := dbgen.New(db).ListCatalogue(t.Context(), dbgen.ListCatalogueParams{})
			return err
		},
		func(db *fakeDB) error {
			_, err := dbgen.New(db).CatalogueCountries(t.Context(), nil)
			return err
		},
		func(db *fakeDB) error {
			_, err := dbgen.New(db).CatalogueSDKs(t.Context(), nil)
			return err
		},
		func(db *fakeDB) error {
			_, err := dbgen.New(db).CatalogueRelated(t.Context(), nil)
			return err
		},
		func(db *fakeDB) error {
			_, err := dbgen.New(db).CatalogueDiscussions(t.Context(), nil)
			return err
		},
		func(db *fakeDB) error {
			_, err := dbgen.New(db).ListCatalogueSources(t.Context(), id)
			return err
		},
	}
	for _, call := range calls {
		if err := call(&fakeDB{manyErr: boom}); err == nil {
			t.Fatal("expected query error")
		}
		if err := call(&fakeDB{scanErr: boom, many: [][]any{nil}}); err == nil {
			t.Fatal("expected scan error")
		}
		if err := call(&fakeDB{rowsErr: boom}); err == nil {
			t.Fatal("expected rows error")
		}
	}
	if _, err := dbgen.New(&fakeDB{rowErr: boom}).GetCatalogueEntry(t.Context(), id); err == nil {
		t.Fatal("expected scan error")
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
