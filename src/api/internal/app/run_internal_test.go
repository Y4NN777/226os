package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fakePool struct {
	closed  bool
	pingErr error
}

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
