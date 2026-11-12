package idempotency_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
)

type purgeStore struct {
	fakeStore
	calls chan time.Time
	n     int64
	err   error
}

func (p *purgeStore) Purge(_ context.Context, now time.Time) (int64, error) {
	p.calls <- now
	return p.n, p.err
}

func TestRunPurgeRunsImmediatelyThenOnTheTickAndStopsOnCancel(t *testing.T) {
	store := &purgeStore{calls: make(chan time.Time, 8), n: 3}
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	fixed := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	ctx, cancel := context.WithCancel(t.Context())
	var wg sync.WaitGroup
	wg.Go(func() {
		idempotency.RunPurge(ctx, store, 5*time.Millisecond, func() time.Time { return fixed }, logger)
	})
	for range 2 { // the immediate run and at least one tick
		if got := <-store.calls; !got.Equal(fixed) {
			t.Fatalf("purge cutoff = %v, want the injected clock %v", got, fixed)
		}
	}
	cancel()
	wg.Wait() // returns: no goroutine leak
	if !strings.Contains(logs.String(), `"deleted":3`) {
		t.Fatalf("expected the count at debug level: %s", logs.String())
	}
}

func TestRunPurgeLogsFailuresAtWarn(t *testing.T) {
	store := &purgeStore{calls: make(chan time.Time, 8), err: errors.New("db down")}
	var logs bytes.Buffer
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		idempotency.RunPurge(ctx, store, time.Hour, time.Now, slog.New(slog.NewJSONHandler(&logs, nil)))
	}()
	<-store.calls
	cancel()
	<-done
	if !strings.Contains(logs.String(), `"level":"WARN"`) || !strings.Contains(logs.String(), "db down") {
		t.Fatalf("logs: %s", logs.String())
	}
}

func TestRunPurgeWithACancelledContextDoesNothing(t *testing.T) {
	store := &purgeStore{calls: make(chan time.Time, 1)}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	idempotency.RunPurge(ctx, store, time.Hour, time.Now, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if len(store.calls) != 0 {
		t.Fatal("a cancelled context must not start a purge")
	}
}
