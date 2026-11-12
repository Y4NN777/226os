//go:build integration

package pgtest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Y4NN777/226os/src/api/internal/platform/db"
)

// countRows returns how many rows of table match the key column value.
func countRows(ctx context.Context, t *testing.T, conn *pgx.Conn, q, arg string) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(ctx, q, arg).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// TestTxCommitsTwoTableWrite writes countries and tags in one transaction.
func TestTxCommitsTwoTableWrite(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn := migratedDB(ctx, t)

	err := db.NewRunner(conn).InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO countries (code, name_fr, name_en) VALUES ('ZZ', 'Zedland', 'Zedland')`); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO tags (slug, label) VALUES ('tx-commit', 'Tx commit')`)
		return err
	})
	if err != nil {
		t.Fatalf("InTx: %v", err)
	}
	if n := countRows(ctx, t, conn, `SELECT count(*) FROM countries WHERE code = $1`, "ZZ"); n != 1 {
		t.Fatalf("countries rows = %d, want 1", n)
	}
	if n := countRows(ctx, t, conn, `SELECT count(*) FROM tags WHERE slug = $1`, "tx-commit"); n != 1 {
		t.Fatalf("tags rows = %d, want 1", n)
	}
}

// TestTxRollsBackWhenSecondStatementFails: the tags insert violates
// tags_slug_ck, so the earlier countries insert must not survive.
func TestTxRollsBackWhenSecondStatementFails(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn := migratedDB(ctx, t)

	err := db.NewRunner(conn).InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO countries (code, name_fr, name_en) VALUES ('ZY', 'Zedland', 'Zedland')`); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO tags (slug, label) VALUES ('BAD SLUG', 'Bad')`)
		return err
	})
	if err == nil {
		t.Fatal("InTx = nil, want the check violation")
	}
	var pe = pgErr(err)
	if pe == nil || pe.Code != "23514" || pe.ConstraintName != "tags_slug_ck" {
		t.Fatalf("err = %v, want 23514 tags_slug_ck", err)
	}
	if errors.Is(err, pgx.ErrTxClosed) {
		t.Fatalf("unexpected tx closed error: %v", err)
	}
	if n := countRows(ctx, t, conn, `SELECT count(*) FROM countries WHERE code = $1`, "ZY"); n != 0 {
		t.Fatalf("countries rows = %d after rollback, want 0", n)
	}
	if n := countRows(ctx, t, conn, `SELECT count(*) FROM tags WHERE slug = $1`, "BAD SLUG"); n != 0 {
		t.Fatalf("tags rows = %d after rollback, want 0", n)
	}
}
