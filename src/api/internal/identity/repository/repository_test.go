package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
)

// fakeDB is a scripted DBTX: every Exec/QueryRow pops the next scripted
// result, so each repository method is exercised against exact row shapes.
type fakeDB struct {
	steps     []dbStep
	n         int
	queryRows [][]any // scripted rows for Query (multi-row results)
	queryErr  error
	beginErr  error
	commitErr error
}

type dbStep struct {
	vals []any // QueryRow: the single row; Exec: ignored
	err  error
}

func (f *fakeDB) next(kind string) dbStep {
	if f.n >= len(f.steps) {
		panic("unexpected " + kind)
	}
	s := f.steps[f.n]
	f.n++
	return s
}

func (f *fakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	s := f.next("exec")
	return pgconn.CommandTag{}, s.err
}

func (f *fakeDB) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	return &fakeRows{rows: f.queryRows}, nil
}

func (f *fakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	s := f.next("query row")
	if s.err != nil {
		return fakeRow{err: s.err}
	}
	if len(s.vals) == 0 {
		return fakeRow{err: pgx.ErrNoRows}
	}
	return fakeRow{vals: s.vals}
}

// Begin returns a fakeTx sharing the same scripted steps, so a transaction's
// statements consume the same dbStep sequence.
func (f *fakeDB) Begin(context.Context) (pgx.Tx, error) {
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return fakeTx{db: f}, nil
}

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

func (t fakeTx) Commit(context.Context) error   { return t.db.commitErr }
func (t fakeTx) Rollback(context.Context) error { return nil }

type fakeRow struct {
	vals []any
	err  error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return scanVals(dest, r.vals)
}

// fakeRows is a scripted pgx.Rows for Query (multi-row results).
type fakeRows struct {
	rows [][]any
	i    int
	err  error
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
func (r *fakeRows) Scan(dest ...any) error { return scanVals(dest, r.rows[r.i-1]) }

func scanVals(dest, vals []any) error {
	if len(dest) != len(vals) {
		return errors.New("scan width")
	}
	for i, d := range dest {
		dv := reflect.ValueOf(d).Elem()
		vv := reflect.ValueOf(vals[i])
		if !vv.Type().AssignableTo(dv.Type()) {
			return errors.New("scan type")
		}
		dv.Set(vv)
	}
	return nil
}

var (
	testUUID  = uuid.MustParse("01932f8e-7b1a-7c2d-9e3f-4a5b6c7d8e9f")
	testTime  = time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	testToken = [32]byte{1, 2, 3}
)

func pguuid() pgtype.UUID { return pgtype.UUID{Bytes: testUUID, Valid: true} }
func tsVal(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func createAccountVals(verified bool) []any {
	verifiedAt := pgtype.Timestamptz{}
	if verified {
		verifiedAt = tsVal(testTime)
	}
	return []any{pguuid(), "sam@example.com", verifiedAt, "member", "active", tsVal(testTime)}
}

func acctVals(verified bool) []any {
	return []any{
		pguuid(), "sam@example.com",
		tsVal(testTime),
		"member", "active", tsVal(testTime),
		"$argon2id$v=19$m=65536,t=3,p=1$c2FsdA$a2V5", "sam", "Sam",
	}
}

func sessVals(revoked bool) []any {
	revokedAt := pgtype.Timestamptz{}
	reason := pgtype.Text{}
	if revoked {
		revokedAt = tsVal(testTime)
		reason = pgtype.Text{String: "logout", Valid: true}
	}
	return []any{
		pguuid(), pguuid(), testToken[:], make([]byte, 32),
		tsVal(testTime), tsVal(testTime), tsVal(testTime),
		tsVal(testTime.Add(time.Hour)), tsVal(testTime.Add(24 * time.Hour)),
		revokedAt, reason,
	}
}

func TestCreateAccount(t *testing.T) {
	ctx := context.Background()
	in := service.NewAccount{Email: "sam@example.com", PasswordHash: "phc", Username: "sam", DisplayName: "Sam"}

	// Happy path.
	db := &fakeDB{steps: []dbStep{{vals: createAccountVals(true)}}}
	a, err := New(db).CreateAccount(ctx, in)
	if err != nil || a.Email != "sam@example.com" || a.ID != testUUID.String() || !a.EmailVerified {
		t.Fatalf("account %+v err %v", a, err)
	}

	// Unverified variant: EmailVerifiedAt invalid.
	db = &fakeDB{steps: []dbStep{{vals: createAccountVals(false)}}}
	a, err = New(db).CreateAccount(ctx, in)
	if err != nil || a.EmailVerified {
		t.Fatalf("unverified account %+v err %v", a, err)
	}

	// Constraint mapping.
	for _, tc := range []struct {
		name       string
		step       dbStep
		wantErrIs  error
		wantSubstr string
	}{
		{"email taken", dbStep{err: &pgconn.PgError{Code: "23505", ConstraintName: "users_email_uq"}}, domain.ErrEmailTaken, ""},
		{"username taken", dbStep{err: &pgconn.PgError{Code: "23505", ConstraintName: "profiles_username_uq"}}, domain.ErrUsernameTaken, ""},
		{"other constraint", dbStep{err: &pgconn.PgError{Code: "23505", ConstraintName: "other_uq"}}, nil, "identity create account"},
		{"other sqlstate", dbStep{err: &pgconn.PgError{Code: "23503", ConstraintName: "users_email_uq"}}, nil, "identity create account"},
		{"plain error", dbStep{err: errors.New("conn reset")}, nil, "identity create account"},
	} {
		db := &fakeDB{steps: []dbStep{tc.step}}
		_, err = New(db).CreateAccount(ctx, in)
		if tc.wantErrIs != nil {
			if !errors.Is(err, tc.wantErrIs) {
				t.Fatalf("%s: err %v", tc.name, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.wantSubstr) {
			t.Fatalf("%s: err %v", tc.name, err)
		}
	}
}

func TestAccountLookups(t *testing.T) {
	ctx := context.Background()

	db := &fakeDB{steps: []dbStep{{vals: acctVals(true)}}}
	a, err := New(db).AccountByEmail(ctx, "sam@example.com")
	if err != nil || a.Username != "sam" || a.DisplayName != "Sam" || a.PasswordHash == "" {
		t.Fatalf("account %+v err %v", a, err)
	}

	// An invalid UUID column maps to an empty ID rather than a panic.
	invalid := acctVals(true)
	invalid[0] = pgtype.UUID{}
	db = &fakeDB{steps: []dbStep{{vals: invalid}}}
	a, err = New(db).AccountByEmail(ctx, "sam@example.com")
	if err != nil || a.ID != "" {
		t.Fatalf("invalid uuid: %+v %v", a, err)
	}

	// Not found and driver error.
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"no rows", pgx.ErrNoRows, domain.ErrNotFound},
		{"driver", errors.New("down"), nil},
	} {
		db = &fakeDB{steps: []dbStep{{err: tc.err}}}
		_, err = New(db).AccountByEmail(ctx, "x@y.z")
		if tc.want != nil {
			if !errors.Is(err, tc.want) {
				t.Fatalf("%s: %v", tc.name, err)
			}
		} else if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
	}

	// By ID: valid, malformed and lookup failures.
	db = &fakeDB{steps: []dbStep{{vals: acctVals(true)}}}
	a, err = New(db).AccountByID(ctx, testUUID.String())
	if err != nil || a.ID != testUUID.String() {
		t.Fatalf("by id %+v %v", a, err)
	}
	if _, err = New(&fakeDB{}).AccountByID(ctx, "not-a-uuid"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad uuid: %v", err)
	}
	db = &fakeDB{steps: []dbStep{{err: pgx.ErrNoRows}}}
	if _, err = New(db).AccountByID(ctx, testUUID.String()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no rows: %v", err)
	}
	db = &fakeDB{steps: []dbStep{{err: errors.New("down")}}}
	if _, err = New(db).AccountByID(ctx, testUUID.String()); err == nil {
		t.Fatal("driver error swallowed")
	}
}

func TestCreateSession(t *testing.T) {
	ctx := context.Background()
	sess := domain.Session{
		ID: testUUID.String(), UserID: testUUID.String(),
		TokenHash: testToken, CSRFHash: [32]byte{9},
		CreatedAt: testTime, LastSeenAt: testTime, AuthenticatedAt: testTime,
		IdleExpiresAt: testTime.Add(time.Hour), AbsoluteExpiresAt: testTime.Add(24 * time.Hour),
	}
	if id, err := New(&fakeDB{steps: []dbStep{{}}}).CreateSession(ctx, sess); err != nil || id != sess.ID {
		t.Fatalf("create session %q %v", id, err)
	}
	// Bad ids.
	bad := sess
	bad.ID = "nope"
	if _, err := New(&fakeDB{}).CreateSession(ctx, bad); err == nil {
		t.Fatal("bad session id accepted")
	}
	bad = sess
	bad.UserID = "nope"
	if _, err := New(&fakeDB{}).CreateSession(ctx, bad); err == nil {
		t.Fatal("bad user id accepted")
	}
	// Exec failure.
	db := &fakeDB{steps: []dbStep{{err: errors.New("down")}}}
	if _, err := New(db).CreateSession(ctx, sess); err == nil {
		t.Fatal("exec error swallowed")
	}
}

func TestSessionByTokenHash(t *testing.T) {
	ctx := context.Background()

	// Live session.
	db := &fakeDB{steps: []dbStep{{vals: sessVals(false)}}}
	s, err := New(db).SessionByTokenHash(ctx, testToken)
	if err != nil || s.ID != testUUID.String() || s.UserID != testUUID.String() ||
		s.TokenHash != testToken || s.RevokedAt != nil || s.RevokeReason != "" {
		t.Fatalf("session %+v err %v", s, err)
	}
	// Revoked session carries reason.
	db = &fakeDB{steps: []dbStep{{vals: sessVals(true)}}}
	s, err = New(db).SessionByTokenHash(ctx, testToken)
	if err != nil || s.RevokedAt == nil || s.RevokeReason != "logout" {
		t.Fatalf("revoked session %+v err %v", s, err)
	}
	// Not found and driver error.
	if _, err = New(&fakeDB{steps: []dbStep{{err: pgx.ErrNoRows}}}).SessionByTokenHash(ctx, testToken); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no rows: %v", err)
	}
	if _, err = New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).SessionByTokenHash(ctx, testToken); err == nil {
		t.Fatal("driver error swallowed")
	}
}

func TestSessionMutations(t *testing.T) {
	ctx := context.Background()
	repo := New(&fakeDB{})

	// RevokeSession.
	if err := repo.RevokeSession(ctx, "bad", "logout", testTime); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad id: %v", err)
	}
	ok := &fakeDB{steps: []dbStep{{}}}
	if err := New(ok).RevokeSession(ctx, testUUID.String(), "logout", testTime); err != nil {
		t.Fatal(err)
	}
	if err := New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).RevokeSession(ctx, testUUID.String(), "x", testTime); err == nil {
		t.Fatal("exec error swallowed")
	}

	// RevokeUserSessions: keep id parsing branches.
	if err := New(&fakeDB{}).RevokeUserSessions(ctx, "bad", "r", testTime, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad user: %v", err)
	}
	if err := New(&fakeDB{}).RevokeUserSessions(ctx, testUUID.String(), "r", testTime, "bad-keep"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad keep: %v", err)
	}
	for _, keep := range []string{"", testUUID.String()} {
		ok = &fakeDB{steps: []dbStep{{}}}
		if err := New(ok).RevokeUserSessions(ctx, testUUID.String(), "password_reset", testTime, keep); err != nil {
			t.Fatalf("keep %q: %v", keep, err)
		}
	}
	if err := New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).RevokeUserSessions(ctx, testUUID.String(), "r", testTime, ""); err == nil {
		t.Fatal("exec error swallowed")
	}

	// TouchSession.
	if err := New(&fakeDB{}).TouchSession(ctx, "bad", testTime, testTime); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad id: %v", err)
	}
	ok = &fakeDB{steps: []dbStep{{}}}
	if err := New(ok).TouchSession(ctx, testUUID.String(), testTime, testTime.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).TouchSession(ctx, testUUID.String(), testTime, testTime); err == nil {
		t.Fatal("exec error swallowed")
	}
}

func TestTokenQueries(t *testing.T) {
	ctx := context.Background()
	hash := testToken

	// CreateEmailVerification.
	if err := New(&fakeDB{}).CreateEmailVerification(ctx, "bad", "a@b.c", hash, testTime); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad user: %v", err)
	}
	ok := &fakeDB{steps: []dbStep{{}}}
	if err := New(ok).CreateEmailVerification(ctx, testUUID.String(), "a@b.c", hash, testTime); err != nil {
		t.Fatal(err)
	}
	if err := New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).CreateEmailVerification(ctx, testUUID.String(), "a@b.c", hash, testTime); err == nil {
		t.Fatal("exec error swallowed")
	}

	// VerifyEmail: row, no rows, driver error.
	db := &fakeDB{steps: []dbStep{{vals: []any{pguuid()}}}}
	if err := New(db).VerifyEmail(ctx, hash, testTime); err != nil {
		t.Fatalf("verify email: %v", err)
	}
	if err := New(&fakeDB{steps: []dbStep{{err: pgx.ErrNoRows}}}).VerifyEmail(ctx, hash, testTime); !errors.Is(err, domain.ErrTokenInvalid) {
		t.Fatalf("no rows: %v", err)
	}
	if err := New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).VerifyEmail(ctx, hash, testTime); err == nil {
		t.Fatal("driver error swallowed")
	}

	// CreatePasswordReset.
	if err := New(&fakeDB{}).CreatePasswordReset(ctx, "bad", hash, testTime); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad user: %v", err)
	}
	ok = &fakeDB{steps: []dbStep{{}}}
	if err := New(ok).CreatePasswordReset(ctx, testUUID.String(), hash, testTime); err != nil {
		t.Fatal(err)
	}
	if err := New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).CreatePasswordReset(ctx, testUUID.String(), hash, testTime); err == nil {
		t.Fatal("exec error swallowed")
	}

	// ResetPassword.
	db = &fakeDB{steps: []dbStep{{vals: []any{pguuid()}}}}
	if err := New(db).ResetPassword(ctx, hash, "phc", testTime); err != nil {
		t.Fatalf("reset password: %v", err)
	}
	if err := New(&fakeDB{steps: []dbStep{{err: pgx.ErrNoRows}}}).ResetPassword(ctx, hash, "phc", testTime); !errors.Is(err, domain.ErrTokenInvalid) {
		t.Fatalf("no rows: %v", err)
	}
	if err := New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).ResetPassword(ctx, hash, "phc", testTime); err == nil {
		t.Fatal("driver error swallowed")
	}
}

func TestHit(t *testing.T) {
	ctx := context.Background()
	key := testToken
	now := testTime

	// Allowed.
	db := &fakeDB{steps: []dbStep{{vals: []any{int32(3)}}}}
	ok, retry, err := New(db).Hit(ctx, "signin:ip", key, 5, 15*time.Minute, now)
	if err != nil || !ok || retry != 0 {
		t.Fatalf("hit %+v %v %v", ok, retry, err)
	}
	// Boundary: hits == limit is still allowed (limit is the max count).
	db = &fakeDB{steps: []dbStep{{vals: []any{int32(5)}}}}
	ok, _, err = New(db).Hit(ctx, "signin:ip", key, 5, 15*time.Minute, now)
	if err != nil || !ok {
		t.Fatalf("boundary hit %v %v", ok, err)
	}
	// Exhausted: hits > limit -> denied with retry = next window start - now.
	db = &fakeDB{steps: []dbStep{{vals: []any{int32(6)}}}}
	ok, retry, err = New(db).Hit(ctx, "signin:ip", key, 5, 15*time.Minute, now)
	if err != nil || ok || retry <= 0 {
		t.Fatalf("exhausted %v %v %v", ok, retry, err)
	}
	// Driver error.
	if _, _, err = New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).Hit(ctx, "s", key, 1, time.Minute, now); err == nil {
		t.Fatal("driver error swallowed")
	}
}

func TestHitPurgesExpiredBucketsWhenAWindowOpens(t *testing.T) {
	ctx := context.Background()
	// First hit of a new bucket: the table only grows through new buckets, so
	// that is when old windows are purged (one bounded DELETE).
	db := &fakeDB{steps: []dbStep{{vals: []any{int32(1)}}, {}}}
	ok, _, err := New(db).Hit(ctx, "signin:ip", testToken, 5, 15*time.Minute, testTime)
	if err != nil || !ok || db.n != 2 {
		t.Fatalf("purge not run: ok=%v err=%v steps=%d", ok, err, db.n)
	}
	// A later hit in the same window does not purge.
	db = &fakeDB{steps: []dbStep{{vals: []any{int32(2)}}}}
	if _, _, err = New(db).Hit(ctx, "signin:ip", testToken, 5, 15*time.Minute, testTime); err != nil || db.n != 1 {
		t.Fatalf("unexpected purge: err=%v steps=%d", err, db.n)
	}
	// Housekeeping must never deny a request: the hit is already counted, so a
	// failed purge is logged and the normal result is returned.
	db = &fakeDB{steps: []dbStep{{vals: []any{int32(1)}}, {err: errors.New("down")}}}
	ok, _, err = New(db).Hit(ctx, "signin:ip", testToken, 5, 15*time.Minute, testTime)
	if err != nil || !ok {
		t.Fatalf("a purge failure must not fail the hit: ok=%v err=%v", ok, err)
	}
}

func TestCount(t *testing.T) {
	ctx := context.Background()
	db := &fakeDB{steps: []dbStep{{vals: []any{int32(4)}}}}
	hits, retry, err := New(db).Count(ctx, "signin:account", testToken, 15*time.Minute, testTime)
	if err != nil || hits != 4 || retry <= 0 {
		t.Fatalf("count %d %v %v", hits, retry, err)
	}
	// No row yet: zero hits, not an error.
	db = &fakeDB{steps: []dbStep{{}}}
	if hits, _, err = New(db).Count(ctx, "signin:account", testToken, 15*time.Minute, testTime); err != nil || hits != 0 {
		t.Fatalf("empty window: %d %v", hits, err)
	}
	if _, _, err = New(&fakeDB{steps: []dbStep{{err: errors.New("down")}}}).Count(ctx, "s", testToken, time.Minute, testTime); err == nil {
		t.Fatal("driver error swallowed")
	}
}
