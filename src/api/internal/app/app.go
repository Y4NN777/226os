// Package app is the composition root: it wires config, logging and transport.
package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Y4NN777/226os/src/api/internal/build"
	"github.com/Y4NN777/226os/src/api/internal/build/github"
	"github.com/Y4NN777/226os/src/api/internal/discover"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	"github.com/Y4NN777/226os/src/api/internal/learn"
	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
	"github.com/Y4NN777/226os/src/api/internal/platform/ratelimit"
	"github.com/Y4NN777/226os/src/api/internal/search"
	"github.com/Y4NN777/226os/src/api/internal/solve"
)

const (
	shutdownTimeout = 10 * time.Second
	// idempotencyPurgeEvery is how often expired Idempotency-Key rows are deleted (24 h retention).
	idempotencyPurgeEvery = time.Hour
)

// Run starts the API and blocks until ctx is cancelled. It returns a process exit code.
func Run(ctx context.Context, getenv func(string) string, out io.Writer) int {
	return run(ctx, getenv, out, httpx.Serve, openPool)
}

type serveFunc func(context.Context, net.Listener, http.Handler, time.Duration) error

// dbPool is the pool surface the composition root needs: readiness, the idempotency
// store queries and shutdown.
type dbPool interface {
	httpx.Pinger
	idempotency.DB
	repository.Beginner
	Close()
}

// newDeps builds the router collaborators; the idempotency store is always the PostgreSQL one.
func newDeps(pool dbPool, opts ...Option) httpx.Deps {
	return httpx.Deps{Server: NewServer(opts...), Ready: pool, Idempotency: idempotency.NewPgStore(pool)}
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
	codec, err := cursor.New(cfg.CursorKey)
	if err != nil {
		logger.Error("invalid cursor configuration")
		return 1
	}
	mail := newMailer(cfg, logger)
	secureCookies := cfg.AppEnv != "local-dev"
	identSvc := service.New(repository.New(pool), mail, cfg.CursorKey, service.WithBaseURL(cfg.AppBaseURL))
	authHandler := identitytransport.New(identSvc, secureCookies, logger)
	limiter := ratelimit.New(pool, cfg.CursorKey, logger)
	catalogue := discover.NewHandler(pool, codec, limiter, logger)
	deps := newDeps(pool,
		WithMailer(mail),
		WithCatalogue(catalogue),
		WithErrorReports(catalogue),
		WithIdentity(authHandler),
		WithSearch(search.NewHandler(pool, limiter, logger)),
		WithLearn(learn.NewHandler(pool, codec, limiter, logger)),
		WithBuild(build.NewHandler(pool, github.New(cfg.GitHubToken, logger), codec, limiter, logger)),
		WithAssistant(newAssistantHandler(cfg.AI, pool, limiter, logger)),
		WithSolve(solve.NewHandler(pool, codec, limiter, logger)),
	)
	deps.Subjects = identitytransport.NewResolver(identSvc, secureCookies)
	deps.CSRF = identitytransport.CSRFMiddleware(identSvc, secureCookies, logger)
	// Deferred after pool.Close, so it runs first: the purge job stops before the pool closes.
	stopPurge := startPurge(ctx, deps.Idempotency, logger)
	defer stopPurge()
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Addr)
	if err != nil {
		logger.Error("listen failed", "error", err)
		return 1
	}
	logger.Info("api started", "addr", ln.Addr().String(), "app_env", cfg.AppEnv)
	if err := serve(ctx, ln, httpx.NewRouter(logger, deps), shutdownTimeout); err != nil {
		logger.Error("api stopped", "error", err)
		return 1
	}
	logger.Info("api stopped")
	return 0
}

// startPurge runs the expired-key purge until the returned function is called, which waits for it.
func startPurge(ctx context.Context, store idempotency.Store, logger *slog.Logger) (stop func()) {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Go(func() { idempotency.RunPurge(ctx, store, idempotencyPurgeEvery, time.Now, logger) })
	return func() {
		cancel()
		wg.Wait()
	}
}

// Main runs the API and calls exit only on failure (exit(0) would be a no-op).
func Main(ctx context.Context, getenv func(string) string, out io.Writer, exit func(int)) {
	if code := Run(ctx, getenv, out); code != 0 {
		exit(code)
	}
}
