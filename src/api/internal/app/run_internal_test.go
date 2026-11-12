package app

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRunReportsServeFailure(t *testing.T) {
	env := map[string]string{"APP_ENV": "local-dev", "API_ADDR": "127.0.0.1:0"}
	var out bytes.Buffer
	fail := func(context.Context, net.Listener, http.Handler, time.Duration) error { return errors.New("boom") }
	code := run(context.Background(), func(k string) string { return env[k] }, &out, fail)
	if code != 1 || !strings.Contains(out.String(), "boom") {
		t.Fatalf("code = %d, log = %s", code, out.String())
	}
}
