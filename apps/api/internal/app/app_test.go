package app_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/apps/api/internal/app"
)

func listen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln
}

func get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func freeAddr(t *testing.T) string {
	t.Helper()
	ln := listen(t)
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func TestRunServesHealthzUntilCancelled(t *testing.T) {
	addr := freeAddr(t)
	env := map[string]string{"APP_ENV": "local-dev", "API_ADDR": addr}
	ctx, cancel := context.WithCancel(context.Background())
	var out bytes.Buffer
	done := make(chan int, 1)
	go func() { done <- app.Run(ctx, func(k string) string { return env[k] }, &out) }()

	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := get(ctx, "http://"+addr+"/healthz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never became ready: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	if code := <-done; code != 0 {
		t.Fatalf("exit code = %d, logs = %s", code, out.String())
	}
}

func TestRunFailsOnInvalidConfig(t *testing.T) {
	var out bytes.Buffer
	code := app.Run(context.Background(), func(string) string { return "" }, &out)
	if code != 1 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(out.String(), "APP_ENV") {
		t.Fatalf("log = %s", out.String())
	}
}

func TestRunFailsWhenAddressUnavailable(t *testing.T) {
	ln := listen(t)
	defer func() { _ = ln.Close() }()
	env := map[string]string{"APP_ENV": "local-dev", "API_ADDR": ln.Addr().String()}
	var out bytes.Buffer
	if code := app.Run(context.Background(), func(k string) string { return env[k] }, &out); code != 1 {
		t.Fatalf("exit code = %d", code)
	}
}

func TestMainExitsOnlyOnFailure(t *testing.T) {
	var got []int
	exit := func(c int) { got = append(got, c) }
	var out bytes.Buffer

	app.Main(context.Background(), func(string) string { return "" }, &out, exit)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("exit calls = %v", got)
	}
}
