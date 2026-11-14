package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/moderation/domain"
)

const (
	testActor  = "018f0000-0000-7000-8000-000000000001"
	testTarget = "018f0000-0000-7000-8000-000000000002"
)

var errStore = errors.New("db down")

type dbStep struct {
	vals []any
	err  error
}

type fakeDB struct {
	t                     *testing.T
	steps                 []dbStep
	n                     int
	beginErr, commitErr   error
	committed, rolledBack bool
}

func (f *fakeDB) next(kind string) dbStep {
	f.t.Helper()
	if f.n >= len(f.steps) {
		f.t.Fatalf("unexpected %s call %d", kind, f.n+1)
	}
	s := f.steps[f.n]
	f.n++
	return s
}

func (f *fakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	s := f.next("exec")
	return pgconn.NewCommandTag("UPDATE 1"), s.err
}

func (f *fakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

func (f *fakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	s := f.next("query row")
	return fakeRow(s)
}

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
func (t fakeTx) Commit(context.Context) error {
	if t.db.commitErr == nil {
		t.db.committed = true
	}
	return t.db.commitErr
}
func (t fakeTx) Rollback(context.Context) error { t.db.rolledBack = true; return nil }

type fakeRow struct {
	vals []any
	err  error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.vals) {
		return fmt.Errorf("scan width %d, want %d", len(dest), len(r.vals))
	}
	for i := range dest {
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(r.vals[i]))
	}
	return nil
}

func fake(t *testing.T, steps ...dbStep) *fakeDB { return &fakeDB{t: t, steps: steps} }
func row(vals ...any) dbStep                     { return dbStep{vals: vals} }
func execOK() dbStep                             { return dbStep{} }
func failDB(err error) dbStep                    { return dbStep{err: err} }
func pgID(raw string) pgtype.UUID                { return pgtype.UUID{Bytes: uuid.MustParse(raw), Valid: true} }

func TestChangeEveryTargetInBothDirections(t *testing.T) {
	for _, kind := range []string{"question", "answer", "resource", "project", "article", "relation"} {
		t.Run(kind+" hide", func(t *testing.T) {
			f := fake(t, row(visibleState(kind)), execOK(), execOK())
			if err := New(f).Hide(t.Context(), kind, testTarget, testActor, "spam", "req"); err != nil || !f.committed || f.n != 3 {
				t.Fatalf("err=%v committed=%v calls=%d", err, f.committed, f.n)
			}
		})
		t.Run(kind+" restore", func(t *testing.T) {
			f := fake(t, row(hiddenState(kind)), execOK(), execOK())
			if err := New(f).Restore(t.Context(), kind, testTarget, testActor, "appel", "req"); err != nil || !f.committed || f.n != 3 {
				t.Fatalf("err=%v committed=%v calls=%d", err, f.committed, f.n)
			}
		})
	}
}

func TestChangeFailuresRollback(t *testing.T) {
	tests := []struct {
		name, kind, target, actor string
		steps                     []dbStep
		beginErr, commitErr       error
		want                      error
	}{
		{name: "bad target", kind: "question", target: "bad", actor: testActor},
		{name: "bad actor", kind: "question", target: testTarget, actor: "bad"},
		{name: "begin", kind: "question", target: testTarget, actor: testActor, beginErr: errStore, want: errStore},
		{name: "missing", kind: "question", target: testTarget, actor: testActor, steps: []dbStep{failDB(pgx.ErrNoRows)}, want: domain.ErrNotFound},
		{name: "lock", kind: "question", target: testTarget, actor: testActor, steps: []dbStep{failDB(errStore)}, want: errStore},
		{name: "unknown kind", kind: "planet", target: testTarget, actor: testActor, want: domain.ErrNotFound},
		{name: "wrong state", kind: "question", target: testTarget, actor: testActor, steps: []dbStep{row("removed")}, want: domain.ErrStateConflict},
		{name: "update", kind: "question", target: testTarget, actor: testActor, steps: []dbStep{row("published"), failDB(errStore)}, want: errStore},
		{name: "audit", kind: "question", target: testTarget, actor: testActor, steps: []dbStep{row("published"), execOK(), failDB(errStore)}, want: errStore},
		{name: "commit", kind: "question", target: testTarget, actor: testActor, steps: []dbStep{row("published"), execOK(), execOK()}, commitErr: errStore, want: errStore},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := fake(t, tc.steps...)
			f.beginErr, f.commitErr = tc.beginErr, tc.commitErr
			err := New(f).Hide(t.Context(), tc.kind, tc.target, tc.actor, "spam", "req")
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || f.committed {
				t.Fatalf("err=%v want=%v committed=%v", err, tc.want, f.committed)
			}
		})
	}
}

func TestResolveReportFlows(t *testing.T) {
	created := pgtype.Timestamptz{Time: time.Date(2026, 11, 14, 17, 0, 0, 0, time.UTC), Valid: true}
	reportRow := []any{pgID(testTarget), pgID(testActor), "outdated", "received", created}
	resolvedRow := []any{pgID(testTarget), pgID(testActor), "outdated", "resolved", created}
	for _, status := range []string{"resolved", "rejected"} {
		f := fake(t, row(reportRow...), row(resolvedRow...), execOK())
		got, err := New(f).ResolveReport(t.Context(), testTarget, testActor, status, "note", "req")
		if err != nil || got.ID != testTarget || got.ResourceID != testActor || got.Kind != "outdated" || got.CreatedAt != created.Time || !f.committed {
			t.Fatalf("got=%+v err=%v committed=%v", got, err, f.committed)
		}
	}
	tests := []struct {
		name, report, actor string
		steps               []dbStep
		beginErr, commitErr error
		want                error
	}{
		{name: "bad report", report: "bad", actor: testActor},
		{name: "bad actor", report: testTarget, actor: "bad"},
		{name: "begin", report: testTarget, actor: testActor, beginErr: errStore, want: errStore},
		{name: "missing", report: testTarget, actor: testActor, steps: []dbStep{failDB(pgx.ErrNoRows)}, want: domain.ErrNotFound},
		{name: "lock", report: testTarget, actor: testActor, steps: []dbStep{failDB(errStore)}, want: errStore},
		{name: "terminal", report: testTarget, actor: testActor, steps: []dbStep{row(pgID(testTarget), pgID(testActor), "outdated", "resolved", created)}, want: domain.ErrStateConflict},
		{name: "resolve", report: testTarget, actor: testActor, steps: []dbStep{row(reportRow...), failDB(errStore)}, want: errStore},
		{name: "audit", report: testTarget, actor: testActor, steps: []dbStep{row(reportRow...), row(resolvedRow...), failDB(errStore)}, want: errStore},
		{name: "commit", report: testTarget, actor: testActor, steps: []dbStep{row(reportRow...), row(resolvedRow...), execOK()}, commitErr: errStore, want: errStore},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := fake(t, tc.steps...)
			f.beginErr, f.commitErr = tc.beginErr, tc.commitErr
			_, err := New(f).ResolveReport(t.Context(), tc.report, tc.actor, "resolved", "note", "req")
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || f.committed {
				t.Fatalf("err=%v want=%v committed=%v", err, tc.want, f.committed)
			}
		})
	}
}

func TestRequiredIDAndIDString(t *testing.T) {
	id, err := requiredID(testTarget)
	if err != nil || idString(id) != testTarget {
		t.Fatalf("id=%+v err=%v", id, err)
	}
}

func visibleState(kind string) string {
	if kind == "relation" {
		return "active"
	}
	return "published"
}
func hiddenState(kind string) string {
	if kind == "relation" {
		return "hidden"
	}
	return "removed"
}
