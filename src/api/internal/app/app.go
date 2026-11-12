// Package app is the composition root: it wires config, logging and transport.
package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

const shutdownTimeout = 10 * time.Second

// Run starts the API and blocks until ctx is cancelled. It returns a process exit code.
func Run(ctx context.Context, getenv func(string) string, out io.Writer) int {
	return run(ctx, getenv, out, httpx.Serve)
}

type serveFunc func(context.Context, net.Listener, http.Handler, time.Duration) error

func run(ctx context.Context, getenv func(string) string, out io.Writer, serve serveFunc) int {
	logger := slog.New(slog.NewJSONHandler(out, nil))
	cfg, err := config.Load(getenv)
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		return 1
	}
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Addr)
	if err != nil {
		logger.Error("listen failed", "error", err)
		return 1
	}
	logger.Info("api started", "addr", ln.Addr().String(), "app_env", cfg.AppEnv)
	if err := serve(ctx, ln, httpx.NewRouter(logger), shutdownTimeout); err != nil {
		logger.Error("api stopped", "error", err)
		return 1
	}
	logger.Info("api stopped")
	return 0
}

// Main runs the API and calls exit only on failure (exit(0) would be a no-op).
func Main(ctx context.Context, getenv func(string) string, out io.Writer, exit func(int)) {
	if code := Run(ctx, getenv, out); code != 0 {
		exit(code)
	}
}
