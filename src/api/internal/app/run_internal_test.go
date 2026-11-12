package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
)

type fakePool struct {
	closed  bool
	pingErr error

	mu     sync.Mutex
	execed []string      // SQL of every Exec, in order
	purged chan struct{} // receives once per purge statement, when non-nil
}

func (f *fakePool) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	f.mu.Lock()
	f.execed = append(f.execed, sql)
	ch := f.purged
	f.mu.Unlock()
	if ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	return pgconn.NewCommandTag("DELETE 0"), nil
}

func (f *fakePool) QueryRow(context.Context, string, ...any) pgx.Row { return nil }

func (f *fakePool) Ping(context.Context) error { return f.pingErr }
func (f *fakePool) Close()                     { f.closed = true }

func fakeOpener(p *fakePool) poolOpener {
	return func(context.Context, string) (dbPool, error) { return p, nil }
}

func validEnv(t *testing.T) func(string) string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{
		"APP_ENV":      "local-dev",
		"API_ADDR":     "127.0.0.1:0",
		"DATABASE_URL": "postgres://u:p@127.0.0.1:1/db",
		"CURSOR_KEY":   base64.StdEncoding.EncodeToString(key),
	}
	return func(k string) string { return env[k] }
}

func TestRunReportsServeFailureAndClosesPool(t *testing.T) {
	var out bytes.Buffer
	pool := &fakePool{}
	fail := func(context.Context, net.Listener, http.Handler, time.Duration) error { return errors.New("boom") }
	code := run(context.Background(), validEnv(t), &out, fail, fakeOpener(pool))
	if code != 1 || !strings.Contains(out.String(), "boom") {
		t.Fatalf("code = %d, log = %s", code, out.String())
	}
	if !pool.closed {
		t.Fatal("pool not closed after serve failure")
	}
}

func TestRunClosesPoolOnGracefulShutdown(t *testing.T) {
	var out bytes.Buffer
	pool := &fakePool{}
	var handler http.Handler
	serve := func(_ context.Context, ln net.Listener, h http.Handler, _ time.Duration) error {
		handler = h
		return ln.Close()
	}
	if code := run(context.Background(), validEnv(t), &out, serve, fakeOpener(pool)); code != 0 {
		t.Fatalf("code = %d, log = %s", code, out.String())
	}
	if !pool.closed {
		t.Fatal("pool not closed on shutdown")
	}
	if handler == nil {
		t.Fatal("router not handed to serve")
	}
}

func TestRunFailsWhenPoolCannotBeOpened(t *testing.T) {
	var out bytes.Buffer
	open := func(context.Context, string) (dbPool, error) { return nil, errors.New("dial postgres://u:p@host") }
	code := run(context.Background(), validEnv(t), &out, nil, open)
	if code != 1 || strings.Contains(out.String(), "postgres://") || !strings.Contains(out.String(), "database pool") {
		t.Fatalf("code = %d, log = %s", code, out.String())
	}
}

func TestNewDepsAlwaysWiresTheIdempotencyStore(t *testing.T) {
	deps := newDeps(&fakePool{})
	if deps.Idempotency == nil || deps.Server == nil || deps.Ready == nil {
		t.Fatalf("incomplete deps: %+v", deps)
	}
}

// run must hand the router a store backed by the pool and start the purge job on it, and the
// purge goroutine must be gone when run returns.
func TestRunWiresTheStoreToThePoolAndStopsThePurgeJob(t *testing.T) {
	var out bytes.Buffer
	pool := &fakePool{purged: make(chan struct{}, 1)}
	serve := func(ctx context.Context, _ net.Listener, _ http.Handler, _ time.Duration) error {
		select {
		case <-pool.purged:
		case <-time.After(10 * time.Second):
			t.Error("the purge job never queried the pool")
		}
		return nil
	}
	if code := run(context.Background(), validEnv(t), &out, serve, fakeOpener(pool)); code != 0 {
		t.Fatalf("code = %d, log = %s", code, out.String())
	}
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if len(pool.execed) == 0 || !strings.Contains(pool.execed[0], "DELETE FROM idempotency_keys") || !pool.closed {
		t.Fatalf("execed = %v closed = %v", pool.execed, pool.closed)
	}
}

func TestStartPurgeStopWaitsForTheJob(t *testing.T) {
	pool := &fakePool{}
	stop := startPurge(context.Background(), idempotency.NewPgStore(pool), slog.New(slog.NewJSONHandler(io.Discard, nil)))
	stop() // returns only once RunPurge has exited
	stop2 := startPurge(t.Context(), idempotency.NewPgStore(pool), slog.New(slog.NewJSONHandler(io.Discard, nil)))
	stop2()
}
