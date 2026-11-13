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

func TestGetAccountByEmail(t *testing.T) {
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	verified := pgtype.Timestamptz{Time: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC), Valid: true}
	created := pgtype.Timestamptz{Time: time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC), Valid: true}
	// Column order of the generated SELECT: id, email, email_verified_at, role,
	// status, created_at, password_hash, username, display_name.
	db := &fakeDB{rowVals: []any{id, "alice@example.org", verified, "member", "active", created,
		"$argon2id$v=19$m=65536,t=3,p=1$c2FsdA$a2V5", "alice", "Alice"}}

	got, err := dbgen.New(db).GetAccountByEmail(t.Context(), "alice@example.org")
	if err != nil {
		t.Fatal(err)
	}
	want := dbgen.GetAccountByEmailRow{ID: id, Email: "alice@example.org", EmailVerifiedAt: verified,
		Role: "member", Status: "active", CreatedAt: created,
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=1$c2FsdA$a2V5", Username: "alice", DisplayName: "Alice"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	assertCall(t, db, "WHERE u.email = $1", []any{"alice@example.org"})
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

func TestSearchQueriesBindArgumentsScanAndReject(t *testing.T) {
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	country := pgtype.Text{String: "BF", Valid: true}
	lang := pgtype.Text{String: "fr", Valid: true}
	row := []any{id, "Titre", "Extrait", "fr"}
	type searchCall func(*fakeDB) ([]string, error)
	calls := map[string]searchCall{
		"questions": func(db *fakeDB) ([]string, error) {
			rows, err := dbgen.New(db).SearchQuestions(t.Context(), dbgen.SearchQuestionsParams{Q: "mot", Country: country, Lang: lang, RowLimit: 6})
			return titles(len(rows), func(i int) string { return rows[i].Title }), err
		},
		"resources": func(db *fakeDB) ([]string, error) {
			rows, err := dbgen.New(db).SearchResources(t.Context(), dbgen.SearchResourcesParams{Q: "mot", Country: country, Lang: lang, RowLimit: 6})
			return titles(len(rows), func(i int) string { return rows[i].Title }), err
		},
		"projects": func(db *fakeDB) ([]string, error) {
			rows, err := dbgen.New(db).SearchProjects(t.Context(), dbgen.SearchProjectsParams{Q: "mot", Country: country, Lang: lang, RowLimit: 6})
			return titles(len(rows), func(i int) string { return rows[i].Title }), err
		},
		"learn": func(db *fakeDB) ([]string, error) {
			rows, err := dbgen.New(db).SearchLearn(t.Context(), dbgen.SearchLearnParams{Q: "mot", Lang: lang, RowLimit: 6})
			return titles(len(rows), func(i int) string { return rows[i].Title }), err
		},
	}
	for name, run := range calls {
		t.Run(name, func(t *testing.T) {
			db := &fakeDB{many: [][]any{row}}
			got, err := run(db)
			if err != nil || len(got) != 1 || got[0] != "Titre" {
				t.Fatalf("rows %v err %v", got, err)
			}
			sql := db.calls[0].sql
			if !strings.Contains(sql, "LIMIT") || !strings.Contains(sql, "ts_rank") || !strings.Contains(sql, "websearch_to_tsquery('simple', f_unaccent(") {
				t.Fatalf("search query is not bounded, ranked and accent-folded:\n%s", sql)
			}
			wantArgs := []any{"mot", country, lang, int32(6)}
			if name == "learn" {
				wantArgs = []any{"mot", lang, int32(6)}
			}
			if !reflect.DeepEqual(db.calls[0].args, wantArgs) {
				t.Fatalf("args = %#v, want %#v", db.calls[0].args, wantArgs)
			}
			boom := errors.New("db down")
			if _, err := run(&fakeDB{manyErr: boom}); err == nil {
				t.Fatal("expected query error")
			}
			if _, err := run(&fakeDB{scanErr: boom, many: [][]any{nil}}); err == nil {
				t.Fatal("expected scan error")
			}
			if _, err := run(&fakeDB{rowsErr: boom}); err == nil {
				t.Fatal("expected rows error")
			}
		})
	}
}

func titles(n int, at func(int) string) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = at(i)
	}
	return out
}

// Identity queries (S1): contract tests pin the SQL text, bind order and scan
// order of every generated statement.
func TestIdentityQueries(t *testing.T) {
	ctx := t.Context()
	uid := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	uid2 := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	ts := pgtype.Timestamptz{Time: time.Unix(10, 0).UTC(), Valid: true}
	text := pgtype.Text{String: "logout", Valid: true}
	hash := make([]byte, 32)
	accountCols := []any{uid, "a@b.co", ts, "member", "active", ts,
		"$argon2id$v=19$m=65536,t=3,p=1$c2FsdA$a2V5", "alice", "Alice"}

	// CreateAccount :one (CTE over users + credentials + profiles).
	db := &fakeDB{rowVals: []any{uid, "a@b.co", pgtype.Timestamptz{}, "member", "active", ts}}
	if _, err := dbgen.New(db).CreateAccount(ctx, dbgen.CreateAccountParams{
		Email: "a@b.co", PasswordHash: "phc", Username: "alice", DisplayName: "Alice",
	}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "INSERT INTO profiles", []any{"a@b.co", "phc", "alice", "Alice"})

	// GetAccountByID :one — same row shape as GetAccountByEmail.
	db = &fakeDB{rowVals: accountCols}
	if _, err := dbgen.New(db).GetAccountByID(ctx, uid); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "WHERE u.id = $1", []any{uid})

	// CreateSession :exec.
	db = &fakeDB{}
	if err := dbgen.New(db).CreateSession(ctx, dbgen.CreateSessionParams{
		ID: uid, UserID: uid2, TokenHash: hash, CsrfTokenHash: hash,
		IdleExpiresAt: ts, AbsoluteExpiresAt: ts,
	}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "INSERT INTO sessions", []any{uid, uid2, hash, hash, ts, ts})

	// GetSessionByTokenHash :one — 11 columns.
	sessCols := []any{uid, uid2, hash, hash, ts, ts, ts, ts, ts, pgtype.Timestamptz{}, pgtype.Text{}}
	db = &fakeDB{rowVals: sessCols}
	if _, err := dbgen.New(db).GetSessionByTokenHash(ctx, hash); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "FROM sessions\nWHERE token_hash = $1", []any{hash})

	// Revoked variant of the same row.
	sessCols[9] = ts
	sessCols[10] = text
	db = &fakeDB{rowVals: sessCols}
	if _, err := dbgen.New(db).GetSessionByTokenHash(ctx, hash); err != nil {
		t.Fatal(err)
	}

	// Session mutations :exec.
	for name, run := range map[string]func(*dbgen.Queries) error{
		"RevokeSession": func(q *dbgen.Queries) error {
			return q.RevokeSession(ctx, dbgen.RevokeSessionParams{ID: uid, RevokedReason: text, RevokedAt: ts})
		},
		"RevokeUserSessions": func(q *dbgen.Queries) error {
			return q.RevokeUserSessions(ctx, dbgen.RevokeUserSessionsParams{
				UserID: uid, RevokedReason: text, RevokedAt: ts, Column4: uid2})
		},
		"TouchSession": func(q *dbgen.Queries) error {
			return q.TouchSession(ctx, dbgen.TouchSessionParams{ID: uid, LastSeenAt: ts, IdleExpiresAt: ts})
		},
		"CreateEmailVerification": func(q *dbgen.Queries) error {
			return q.CreateEmailVerification(ctx, dbgen.CreateEmailVerificationParams{
				UserID: uid, Email: "a@b.co", TokenHash: hash, ExpiresAt: ts})
		},
		"CreatePasswordReset": func(q *dbgen.Queries) error {
			return q.CreatePasswordReset(ctx, dbgen.CreatePasswordResetParams{
				UserID: uid, TokenHash: hash, ExpiresAt: ts})
		},
	} {
		db = &fakeDB{}
		if err := run(dbgen.New(db)); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(db.calls) != 1 {
			t.Fatalf("%s: calls = %d", name, len(db.calls))
		}
	}

	// Consume queries :one returning user_id.
	for name, run := range map[string]func(*dbgen.Queries) (pgtype.UUID, error){
		"VerifyEmail": func(q *dbgen.Queries) (pgtype.UUID, error) {
			return q.VerifyEmail(ctx, dbgen.VerifyEmailParams{TokenHash: hash, ConsumedAt: ts})
		},
		"ResetPassword": func(q *dbgen.Queries) (pgtype.UUID, error) {
			return q.ResetPassword(ctx, dbgen.ResetPasswordParams{
				TokenHash: hash, ConsumedAt: ts, PasswordHash: "phc"})
		},
	} {
		db = &fakeDB{rowVals: []any{uid}}
		got, err := run(dbgen.New(db))
		if err != nil || got != uid {
			t.Fatalf("%s: %v %v", name, got, err)
		}
		if len(db.calls) != 1 {
			t.Fatalf("%s: calls = %d", name, len(db.calls))
		}
	}

	// HitRateLimit :one returning hits.
	db = &fakeDB{rowVals: []any{int32(7)}}
	hits, err := dbgen.New(db).HitRateLimit(ctx, dbgen.HitRateLimitParams{
		Scope: "signin:ip", KeyHash: hash, WindowStart: ts})
	if err != nil || hits != 7 {
		t.Fatalf("hits = %d, %v", hits, err)
	}
	assertCall(t, db, "INSERT INTO rate_limit_buckets", []any{"signin:ip", hash, ts})

	// ---------- profiles ----------

	// Shared profile projection: 20 columns ending with the skills array.
	profCols := func() []any {
		return []any{
			uid, "ann", "Ann", text, text, []string{"fr"}, text, text, text,
			true, "public", "members", "private", "public", "members", "public", "public",
			ts, ts, []string{"go"},
		}
	}

	// GetProfileByUserID :one.
	db = &fakeDB{rowVals: profCols()}
	if _, err := dbgen.New(db).GetProfileByUserID(ctx, uid); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "FROM profiles p", []any{uid})

	// GetProfileByUsername :one — 22 columns: account_status and joined_at
	// sit between updated_at and skills.
	cols := profCols()
	skills := cols[len(cols)-1]
	cols = cols[:len(cols)-1]
	cols = append(cols, "active", ts, skills)
	db = &fakeDB{rowVals: cols}
	if _, err := dbgen.New(db).GetProfileByUsername(ctx, "ann"); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "JOIN users u ON u.id = p.user_id", []any{"ann"})

	// ResolveSkillSlugs :many — happy path plus the three error branches.
	db = &fakeDB{many: [][]any{{uid, "go"}, {uid2, "rust"}}}
	rows, err := dbgen.New(db).ResolveSkillSlugs(ctx, []string{"go", "rust"})
	if err != nil || len(rows) != 2 {
		t.Fatalf("slugs = %v, %v", rows, err)
	}
	assertCall(t, db, "FROM tags WHERE slug = ANY", []any{[]string{"go", "rust"}})

	db = &fakeDB{manyErr: errors.New("conn reset")}
	if _, err := dbgen.New(db).ResolveSkillSlugs(ctx, []string{"go"}); err == nil {
		t.Fatal("query error not propagated")
	}
	db = &fakeDB{many: [][]any{{uid, "go"}}, scanErr: errors.New("scan boom")}
	if _, err := dbgen.New(db).ResolveSkillSlugs(ctx, []string{"go"}); err == nil {
		t.Fatal("scan error not propagated")
	}
	db = &fakeDB{many: [][]any{{uid, "go"}}, rowsErr: errors.New("rows boom")}
	if _, err := dbgen.New(db).ResolveSkillSlugs(ctx, []string{"go"}); err == nil {
		t.Fatal("rows error not propagated")
	}

	// UpdateProfile :one — the guarded merge patch statement (31 bind args;
	// assert the statement shape and the user id rather than every flag).
	db = &fakeDB{rowVals: []any{uid}}
	if _, err := dbgen.New(db).UpdateProfile(ctx, dbgen.UpdateProfileParams{UserID: uid}); err != nil {
		t.Fatal(err)
	}
	if len(db.calls) != 1 || !strings.Contains(db.calls[0].sql, "UPDATE profiles SET") ||
		db.calls[0].args[0] != uid {
		t.Fatalf("update profile call %+v", db.calls)
	}

	// CountRateLimit :one.
	db = &fakeDB{rowVals: []any{int32(3)}}
	got, err := dbgen.New(db).CountRateLimit(ctx, dbgen.CountRateLimitParams{
		Scope: "signin:account", KeyHash: hash, WindowStart: ts})
	if err != nil || got != 3 {
		t.Fatalf("count = %d, %v", got, err)
	}
	assertCall(t, db, "SELECT hits FROM rate_limit_buckets", []any{"signin:account", hash, ts})

	// PurgeRateLimitBuckets :exec.
	db = &fakeDB{}
	if err := dbgen.New(db).PurgeRateLimitBuckets(ctx, ts); err != nil {
		t.Fatalf("purge: %v", err)
	}
	assertCall(t, db, "DELETE FROM rate_limit_buckets", []any{ts})
}
