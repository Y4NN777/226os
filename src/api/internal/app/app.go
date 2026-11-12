// Package app is the composition root: it wires config, logging and transport.
package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

const shutdownTimeout = 10 * time.Second

// Run starts the API and blocks until ctx is cancelled. It returns a process exit code.
func Run(ctx context.Context, getenv func(string) string, out io.Writer) int {
	return run(ctx, getenv, out, httpx.Serve, openPool)
}

type serveFunc func(context.Context, net.Listener, http.Handler, time.Duration) error

// dbPool is the pool surface the composition root needs: readiness and shutdown.
type dbPool interface {
	httpx.Pinger
	Close()
}

// poolOpener builds the database pool; injectable so tests need no database.
type poolOpener func(ctx context.Context, url string) (dbPool, error)

// openPool creates a lazy pgx pool: it validates the URL but does not connect.
func openPool(ctx context.Context, url string) (dbPool, error) {
	p, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func run(ctx context.Context, getenv func(string) string, out io.Writer, serve serveFunc, open poolOpener) int {
	logger := slog.New(slog.NewJSONHandler(out, nil))
	cfg, err := config.Load(getenv)
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		return 1
	}
	pool, err := open(ctx, cfg.DatabaseURL)
	if err != nil {
		// Fixed message on purpose: driver errors can echo the connection URL and its password.
		logger.Error("invalid database pool configuration")
		return 1
	}
	defer pool.Close()
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Addr)
	if err != nil {
		logger.Error("listen failed", "error", err)
		return 1
	}
	logger.Info("api started", "addr", ln.Addr().String(), "app_env", cfg.AppEnv)
	if err := serve(ctx, ln, httpx.NewRouter(logger, httpx.Deps{Server: NewServer(), Ready: pool}), shutdownTimeout); err != nil {
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
