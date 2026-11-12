package app_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/app"
)

func envOf(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

// serveStatus starts a server answering status on every path and returns its host:port.
func serveStatus(t *testing.T, status int) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
	t.Cleanup(srv.Close)
	return srv.Listener.Addr().String()
}

func TestHealthCheckSucceedsOnTwoHundred(t *testing.T) {
	var e bytes.Buffer
	for _, args := range [][]string{nil, {"/healthz"}, {"/readyz"}} {
		code := app.HealthCheck(context.Background(), args, envOf(map[string]string{"API_ADDR": serveStatus(t, http.StatusOK)}), http.DefaultClient, &e)
		if code != 0 || e.Len() != 0 {
			t.Fatalf("args %v: code = %d err = %q", args, code, e.String())
		}
	}
}

func TestHealthCheckFailsOnNon2xx(t *testing.T) {
	var e bytes.Buffer
	code := app.HealthCheck(context.Background(), nil, envOf(map[string]string{"API_ADDR": serveStatus(t, http.StatusServiceUnavailable)}), http.DefaultClient, &e)
	if code != 1 || !strings.Contains(e.String(), "503") {
		t.Fatalf("code = %d err = %q", code, e.String())
	}
}

func TestHealthCheckFailsWhenNothingListens(t *testing.T) {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	var e bytes.Buffer
	code := app.HealthCheck(context.Background(), nil, envOf(map[string]string{"API_ADDR": addr}), http.DefaultClient, &e)
	if code != 1 || e.Len() == 0 {
		t.Fatalf("code = %d err = %q", code, e.String())
	}
}

func TestHealthCheckUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"/metrics"}, {"/healthz", "/readyz"}} {
		var e bytes.Buffer
		code := app.HealthCheck(context.Background(), args, envOf(nil), http.DefaultClient, &e)
		if code != 2 || !strings.Contains(e.String(), "usage: healthcheck") {
			t.Fatalf("args %v: code = %d err = %q", args, code, e.String())
		}
	}
}

func TestHealthCheckHelpNeedsNoNetwork(t *testing.T) {
	var e bytes.Buffer
	if code := app.HealthCheck(context.Background(), []string{"help"}, envOf(nil), http.DefaultClient, &e); code != 0 || !strings.Contains(e.String(), "usage: healthcheck") {
		t.Fatalf("code = %d err = %q", code, e.String())
	}
}

func TestHealthCheckRejectsAnInvalidAddress(t *testing.T) {
	var e bytes.Buffer
	code := app.HealthCheck(context.Background(), nil, envOf(map[string]string{"API_ADDR": "no-port"}), http.DefaultClient, &e)
	if code != 2 || !strings.Contains(e.String(), "API_ADDR") {
		t.Fatalf("code = %d err = %q", code, e.String())
	}
}

func TestHealthCheckUsesLoopbackForWildcardAddresses(t *testing.T) {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	_ = srv.Listener.Close()
	srv.Listener = ln
	srv.Start()
	t.Cleanup(srv.Close)
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	for _, addr := range []string{":" + port, "0.0.0.0:" + port, "[::]:" + port} {
		var e bytes.Buffer
		if code := app.HealthCheck(context.Background(), nil, envOf(map[string]string{"API_ADDR": addr}), http.DefaultClient, &e); code != 0 {
			t.Fatalf("addr %q: code = %d err = %q", addr, code, e.String())
		}
	}
}

func TestHealthCheckMainExitsOnlyOnFailure(t *testing.T) {
	var e bytes.Buffer
	app.HealthCheckMain(context.Background(), []string{"help"}, envOf(nil), http.DefaultClient, &e, func(int) { t.Fatal("exit called on success") })
	got := -1
	app.HealthCheckMain(context.Background(), []string{"/metrics"}, envOf(nil), http.DefaultClient, &e, func(c int) { got = c })
	if got != 2 {
		t.Fatalf("exit code = %d", got)
	}
}

func TestHealthCheckRejectsAnUnusableHost(t *testing.T) {
	var e bytes.Buffer
	code := app.HealthCheck(context.Background(), nil, envOf(map[string]string{"API_ADDR": "bad host:80"}), http.DefaultClient, &e)
	if code != 2 || !strings.Contains(e.String(), "invalid probe URL") {
		t.Fatalf("code = %d err = %q", code, e.String())
	}
}
