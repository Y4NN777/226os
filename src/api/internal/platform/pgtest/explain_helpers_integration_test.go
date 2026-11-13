//go:build integration

package pgtest

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

type recordedQuery struct {
	sql  string
	args []any
}

// recordingDB forwards to db and remembers every statement the repositories send, so a test can run
// EXPLAIN (ANALYZE, BUFFERS) on the exact SQL and arguments of a real request instead of a copy.
type recordingDB struct {
	db    dbgen.DBTX
	mu    sync.Mutex
	calls []recordedQuery
}

func (r *recordingDB) record(sql string, args []any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, recordedQuery{sql, args})
}

func (r *recordingDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	r.record(sql, args)
	return r.db.Exec(ctx, sql, args...)
}

func (r *recordingDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	r.record(sql, args)
	return r.db.Query(ctx, sql, args...)
}

func (r *recordingDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	r.record(sql, args)
	return r.db.QueryRow(ctx, sql, args...)
}

// reset forgets the recorded statements.
func (r *recordingDB) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = nil
}

// first returns the first recorded statement.
func (r *recordingDB) first(t *testing.T) recordedQuery {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.calls) == 0 {
		t.Fatal("no statement was recorded")
	}
	return r.calls[0]
}

// explainPlan runs EXPLAIN (ANALYZE, BUFFERS) on a read statement and returns the plan text.
// ANALYZE executes the statement: only call it on read queries against the throwaway test database.
func explainPlan(ctx context.Context, t *testing.T, conn *pgx.Conn, sql string, args ...any) string {
	t.Helper()
	rows, err := conn.Query(ctx, "EXPLAIN (ANALYZE, BUFFERS) "+sql, args...)
	if err != nil {
		t.Fatalf("explain: %v", err)
	}
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		plan.WriteString(line)
		plan.WriteByte('\n')
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return plan.String()
}
