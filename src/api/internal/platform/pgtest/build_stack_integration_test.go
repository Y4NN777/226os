//go:build integration

package pgtest

import (
	"io"
	"log/slog"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/app"
	"github.com/Y4NN777/226os/src/api/internal/build"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
	"github.com/Y4NN777/226os/src/api/internal/platform/ratelimit"
)

// newBuildStack is the S2 stack's database and identity with the Build handler wired over the real
// rate limiter and idempotency store; gh stands in for the GitHub API (no network in tests).
func newBuildStack(t *testing.T, gh *fakeGitHub) *s2Stack {
	t.Helper()
	s := newS2Stack(t.Context(), t, nil)
	svc := service.New(repository.New(s.pool), s.mail, s2Secret)
	codec, err := cursor.New(s2Secret)
	if err != nil {
		t.Fatal(err)
	}
	deps := httpx.Deps{
		Server: app.NewServer(
			app.WithIdentity(identitytransport.New(svc, false, nil)),
			app.WithBuild(build.NewHandler(s.pool, gh, codec, ratelimit.New(s.pool, s2Secret, nil), nil)),
		),
		Ready:       s.pool,
		Idempotency: idempotency.NewPgStore(s.pool),
	}
	deps.Subjects = identitytransport.NewResolver(svc, false)
	deps.CSRF = identitytransport.CSRFMiddleware(svc, false, nil)
	s.router = httpx.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), deps)
	return s
}
