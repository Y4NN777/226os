// Package db holds the transaction runner used by services for multi-table
// mutations (docs/SPEC.md: multi-table writes are transactional).
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Beginner starts a transaction. *pgxpool.Pool satisfies it.
type Beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Runner runs functions inside a transaction.
type Runner struct{ b Beginner }

// NewRunner returns a Runner over b.
func NewRunner(b Beginner) *Runner { return &Runner{b: b} }

// rollbackTimeout bounds a rollback issued after the request context may
// already be cancelled.
const rollbackTimeout = 5 * time.Second

// InTx begins a transaction, runs fn and commits when fn returns nil. When fn
// returns an error the transaction is rolled back and that error is returned
// (wrapped with %w; joined with the rollback error if the rollback also
// fails). When fn panics the transaction is rolled back and the panic is
// re-raised. A commit failure is returned wrapped.
//
// The rollback uses a context detached from ctx cancellation (with a short
// timeout) so a cancelled request still releases the connection cleanly.
func (r *Runner) InTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) (err error) {
	tx, err := r.b.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	rollback := func() error {
		rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
		defer cancel()
		return tx.Rollback(rbCtx)
	}

	// A panic or runtime.Goexit inside fn (for example t.FailNow in a test) neither returns nor
	// commits: roll back, and let the panic or the exit carry on.
	finished := false
	defer func() {
		if !finished {
			_ = rollback() // the panic (or exit) is the primary failure
		}
	}()

	fnErr := fn(ctx, tx)
	finished = true
	if err := fnErr; err != nil {
		err = fmt.Errorf("transaction: %w", err)
		if rbErr := rollback(); rbErr != nil {
			return errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
