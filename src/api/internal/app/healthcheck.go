package app

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
)

const (
	healthUsage   = "usage: healthcheck [/healthz | /readyz | help]\n"
	healthTimeout = 3 * time.Second
)

// HealthCheck probes the API listening in the same container and returns a process exit code: 0 when the
// path answers 2xx, 1 when it does not answer or answers otherwise, 2 for a usage or address problem.
// The distroless runtime has no shell and no wget, so the image health check and the smoke test call
// this command instead. Only /healthz (default) and /readyz are allowed; nothing from the response body
// is printed. The address is the one the API itself uses (API_ADDR, default :8080), probed on loopback.
func HealthCheck(ctx context.Context, args []string, getenv func(string) string, client *http.Client, errOut io.Writer) int {
	path := "/healthz"
	switch {
	case len(args) == 1 && args[0] == "help":
		say(errOut, "%s", healthUsage)
		return 0
	case len(args) == 1 && (args[0] == "/healthz" || args[0] == "/readyz"):
		path = args[0]
	case len(args) != 0:
		say(errOut, "%s", healthUsage)
		return 2
	}
	host, port, err := net.SplitHostPort(config.ListenAddr(getenv))
	if err != nil {
		say(errOut, "healthcheck: API_ADDR is not host:port\n")
		return 2
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()
	// The URL is built from a fixed scheme, a loopback or configured host and one of two fixed paths.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort(host, port)+path, nil)
	if err != nil {
		say(errOut, "healthcheck: invalid probe URL\n")
		return 2
	}
	resp, err := client.Do(req)
	if err != nil {
		say(errOut, "healthcheck: %s did not answer\n", path)
		return 1
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		say(errOut, "healthcheck: %s answered %d\n", path, resp.StatusCode)
		return 1
	}
	return 0
}

// HealthCheckMain runs HealthCheck and calls exit only on failure (exit(0) would be a no-op).
func HealthCheckMain(ctx context.Context, args []string, getenv func(string) string, client *http.Client, errOut io.Writer, exit func(int)) {
	if code := HealthCheck(ctx, args, getenv, client, errOut); code != 0 {
		exit(code)
	}
}
