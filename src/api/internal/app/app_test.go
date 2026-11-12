package app_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/app"
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

const testPassword = "s3cr3t-pw-xyz"

// testEnv is a valid environment: the database URL points at a closed port so the lazy
// pool never connects, and the cursor key is generated per run.
func testEnv(t *testing.T, addr string) map[string]string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return map[string]string{
		"APP_ENV":      "local-dev",
		"API_ADDR":     addr,
		"DATABASE_URL": "postgres://u:" + testPassword + "@127.0.0.1:1/db",
		"CURSOR_KEY":   base64.StdEncoding.EncodeToString(key),
	}
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
	env := testEnv(t, addr)
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
	resp, err := get(ctx, "http://"+addr+"/readyz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var body struct{ Code string }
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable || body.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("readyz = %d %+v, want 503 for an unreachable database", resp.StatusCode, body)
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
	env := testEnv(t, ln.Addr().String())
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

func TestRunFailsOnInvalidPoolConfigWithoutLeakingSecrets(t *testing.T) {
	env := testEnv(t, "127.0.0.1:0")
	env["DATABASE_URL"] += "?pool_max_conns=not-a-number"
	var out bytes.Buffer
	if code := app.Run(context.Background(), func(k string) string { return env[k] }, &out); code != 1 {
		t.Fatalf("exit code = %d, logs = %s", code, out.String())
	}
	for _, leak := range []string{testPassword, "u:", "127.0.0.1:1", "pool_max_conns", env["CURSOR_KEY"]} {
		if strings.Contains(out.String(), leak) {
			t.Fatalf("log leaks %q: %s", leak, out.String())
		}
	}
	if !strings.Contains(out.String(), "database pool") {
		t.Fatalf("log = %s", out.String())
	}
}

func TestServerLeavesEveryOperationUnimplemented(t *testing.T) {
	rec := httptest.NewRecorder()
	app.NewServer().GetCsrfToken(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d", rec.Code)
	}
}
