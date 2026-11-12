package db_test

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/Y4NN777/226os/src/api/internal/platform/db"
)

type fakeTx struct {
	pgx.Tx
	commitErr, rollbackErr error
	commits, rollbacks     int
	rollbackCtxErr         error
	rollbackHasDeadline    bool
}

func (f *fakeTx) Commit(context.Context) error { f.commits++; return f.commitErr }

func (f *fakeTx) Rollback(ctx context.Context) error {
	f.rollbacks++
	f.rollbackCtxErr = ctx.Err()
	_, f.rollbackHasDeadline = ctx.Deadline()
	return f.rollbackErr
}

type fakeBeginner struct {
	tx       *fakeTx
	beginErr error
}

func (f *fakeBeginner) Begin(context.Context) (pgx.Tx, error) {
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return f.tx, nil
}

var errBoom = errors.New("boom")

func TestInTxCommitsWhenFnSucceeds(t *testing.T) {
	tx := &fakeTx{}
	r := db.NewRunner(&fakeBeginner{tx: tx})
	if err := r.InTx(context.Background(), func(context.Context, pgx.Tx) error { return nil }); err != nil {
		t.Fatalf("err = %v", err)
	}
	if tx.commits != 1 || tx.rollbacks != 0 {
		t.Fatalf("commits=%d rollbacks=%d", tx.commits, tx.rollbacks)
	}
}

func TestInTxRollsBackAndKeepsFnError(t *testing.T) {
	tx := &fakeTx{}
	r := db.NewRunner(&fakeBeginner{tx: tx})
	err := r.InTx(context.Background(), func(context.Context, pgx.Tx) error { return errBoom })
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want errBoom", err)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Fatalf("commits=%d rollbacks=%d", tx.commits, tx.rollbacks)
	}
}

func TestInTxPanicRollsBackThenRepanics(t *testing.T) {
	tx := &fakeTx{}
	r := db.NewRunner(&fakeBeginner{tx: tx})
	defer func() {
		if got := recover(); got != "kaboom" {
			t.Fatalf("recovered %v", got)
		}
		if tx.rollbacks != 1 || tx.commits != 0 {
			t.Fatalf("commits=%d rollbacks=%d", tx.commits, tx.rollbacks)
		}
	}()
	_ = r.InTx(context.Background(), func(context.Context, pgx.Tx) error { panic("kaboom") })
	t.Fatal("expected panic")
}

func TestInTxBeginFailure(t *testing.T) {
	r := db.NewRunner(&fakeBeginner{beginErr: errBoom})
	called := false
	err := r.InTx(context.Background(), func(context.Context, pgx.Tx) error { called = true; return nil })
	if !errors.Is(err, errBoom) || called {
		t.Fatalf("err=%v called=%v", err, called)
	}
}

func TestInTxCommitFailureIsWrapped(t *testing.T) {
	tx := &fakeTx{commitErr: errBoom}
	r := db.NewRunner(&fakeBeginner{tx: tx})
	err := r.InTx(context.Background(), func(context.Context, pgx.Tx) error { return nil })
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}

func TestInTxRollbackFailureIsJoined(t *testing.T) {
	errRb := errors.New("rollback failed")
	tx := &fakeTx{rollbackErr: errRb}
	r := db.NewRunner(&fakeBeginner{tx: tx})
	err := r.InTx(context.Background(), func(context.Context, pgx.Tx) error { return errBoom })
	if !errors.Is(err, errBoom) || !errors.Is(err, errRb) {
		t.Fatalf("err = %v, want both", err)
	}
}

func TestInTxRollbackUsesLiveContextWhenRequestCancelled(t *testing.T) {
	tx := &fakeTx{}
	r := db.NewRunner(&fakeBeginner{tx: tx})
	ctx, cancel := context.WithCancel(context.Background())
	err := r.InTx(ctx, func(context.Context, pgx.Tx) error { cancel(); return ctx.Err() })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if tx.rollbacks != 1 || tx.rollbackCtxErr != nil || !tx.rollbackHasDeadline {
		t.Fatalf("rollbacks=%d ctxErr=%v deadline=%v", tx.rollbacks, tx.rollbackCtxErr, tx.rollbackHasDeadline)
	}
}

// runtime.Goexit (for example t.FailNow inside fn) neither returns nor panics: the transaction must
// still be rolled back, or it leaks until the connection closes.
func TestInTxRollsBackOnGoexit(t *testing.T) {
	tx := &fakeTx{}
	runner := db.NewRunner(&fakeBeginner{tx: tx})
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = runner.InTx(context.Background(), func(context.Context, pgx.Tx) error {
			runtime.Goexit()
			return nil
		})
	}()
	<-done
	if tx.rollbacks != 1 || tx.commits != 0 {
		t.Fatalf("rollbacks = %d, commits = %d; want 1 and 0", tx.rollbacks, tx.commits)
	}
}
