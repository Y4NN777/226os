---
name: go-backend
description: 226os Go modular monolith rules and practices (Chi transport-only handlers, domain/service/repository boundaries, context deadlines, error wrapping, slog logging, config, concurrency, package layout, 100% statement coverage). Use when touching src/api/**.
---
# Go Backend (modular monolith)

## Scope and triggers
- `src/api/**` (`cmd/`, `internal/<module>/...`). Pair with `postgres-data` for `db/**`, `api-contract` for `api/openapi.yaml`, `security-authz` for identity/auth paths.
- Status: the API exists under `src/api` (see Layout below). Work only inside an open slice: its block in `TODO.md` names owner, reviewer and allowed paths (ADR-0008). Ask the user to run `/start-slice` if none is open.

## Project rules (binding)
- Stack: Go, Chi, `pgx`, `sqlc`, modular monolith, REST, OpenAPI 3.1 (`docs/DEV.md:7`; ACCEPTED as BASE-004 `docs/decisions/README.md:22`).
- No general-purpose ORM, external search engine, microservices, Redis or Kubernetes without an approved ADR (`docs/DEV.md:10`, `docs/architecture/ARCHITECTURE.md:27`).
- Business logic never lives in HTTP handlers (`docs/DEV.md:10`, `docs/RULES.md:12`); handlers carry transport concerns only, invariants live in the domain/service boundary.
- Packages per domain: `domain`, `service`, `repository`, `transport`; interfaces at the boundary, dependencies point toward business logic (`docs/DEV.md:44`).
- Modules: identity, solve, discover, build, community, connect, language, intelligence (`docs/architecture/ARCHITECTURE.md:10`); a module never imports a neighbour's HTTP transport, share through explicit interfaces and limited models (`docs/architecture/ARCHITECTURE.md:20`).
- Do not materialize empty P1/P2 module folders (`docs/SPEC.md:96`).
- Context with deadline on DB access, HTTP requests, external integrations; typed errors with wrapping; parameterized SQL only (`docs/DEV.md:45`, `docs/RULES.md:12`).
- `go fmt`, `go vet`, `staticcheck`, `golangci-lint` at pinned versions; `go test -race`; PG integration on Docker (`docs/DEV.md:46`).
- Separate configuration, application and handlers; never log secrets (`docs/DEV.md:48`, `docs/RULES.md:25`).
- API owns invariants and permissions; access checks as close as possible to the business operation (`docs/architecture/ARCHITECTURE.md:19`, `:23`).
- Multi-table mutations are transactional (`docs/SPEC.md:119`, `docs/archive/spec-kit/CONTRACTS.md:24`).
- Health/readiness endpoints and structured logs without secrets are required (`docs/archive/spec-kit/CONTRACTS.md:57`).
- GitHub public API access is limited, read-only, with quotas, timeouts and anti-SSRF (`docs/architecture/ARCHITECTURE.md:13`, `docs/SPEC.md:126`).
- Coverage: 100.00% of instrumentable first-party Go statements via `go test -covermode=atomic -coverprofile=coverage.out ./...` + `go tool cover -func` (`docs/TEST.md:6`, `AGENTS.md:48`). Run through `make coverage` (`scripts/coverage-go.sh`). Exclusions only via the `coverage-exclude-go` block of `docs/quality/COVERAGE_SCOPE.md`; read the block for what is excluded today. Never add an exclusion yourself: it is an owner-reviewed file.
- Run only reviewed Make targets after reading their recipe (`AGENTS.md:120-121`). Go toolchain and linter versions are pinned in `go.mod`, `mk/00-config.mk` and the toolchain files (ADR-0006); do not restate them.

## Expert practices
- Wire everything in `cmd/api/main.go`: load config, build `pgxpool`, construct repositories -> services -> handlers, mount Chi router; no package-level mutable globals or `init()` side effects. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Pass `ctx context.Context` as first parameter; never store it in a struct. [CodeReviewComments#contexts](https://go.dev/wiki/CodeReviewComments#contexts), [context pkg](https://pkg.go.dev/context)
- Derive DB/external deadlines with `context.WithTimeout` from the request context; always `defer cancel()`. [context](https://pkg.go.dev/context)
- Define interfaces in the consuming package (service declares the `Repository` it needs); return concrete types. [CodeReviewComments#interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
- Wrap with `fmt.Errorf("op: %w", err)`; classify with `errors.Is`/`errors.As`; define domain sentinels (`ErrNotFound`, `ErrForbidden`, `ErrConflict`, `ErrValidation`) in `domain`; map to HTTP status only in `transport`. [Go blog: errors](https://go.dev/blog/go1.13-errors), [errors pkg](https://pkg.go.dev/errors)
- Error strings lowercase, no trailing punctuation; never discard errors with `_`; indent error flow, early return. [CodeReviewComments](https://go.dev/wiki/CodeReviewComments#error-strings)
- Translate `pgx.ErrNoRows` and constraint violations (`*pgconn.PgError` codes `23505`, `23503`, `23514`) inside `repository`, never leak driver types upward. [pgconn](https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn), [PG error codes](https://www.postgresql.org/docs/current/errcodes-appendix.html)
- Use stdlib `log/slog` JSON handler; attach `request_id` from Chi `middleware.RequestID`; log attributes, never raw request bodies, cookies, tokens, passwords, emails. [log/slog](https://pkg.go.dev/log/slog)
- Middleware order: RequestID -> client IP (trusted-proxy aware) -> structured logger -> Recoverer -> Timeout -> body-size limit -> auth/session -> CSRF -> routes. Do not use deprecated `middleware.RealIP` (spoofable); use the `ClientIPFrom*` variant matching Traefik. [chi middleware](https://pkg.go.dev/github.com/go-chi/chi/v5/middleware)
- Bound request bodies with `http.MaxBytesReader` (-> 413) and decode with `json.Decoder.DisallowUnknownFields()` for mutations (unknown fields rejected, `docs/spec-kit/API_CONTRACT.md:10`). [net/http MaxBytesReader](https://pkg.go.dev/net/http#MaxBytesReader), [encoding/json](https://pkg.go.dev/encoding/json#Decoder.DisallowUnknownFields)
- Set `http.Server` `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`; implement graceful shutdown with `signal.NotifyContext` + `srv.Shutdown(ctx)`. [net/http Server](https://pkg.go.dev/net/http#Server)
- Config: parse env once into a typed struct, validate at startup, fail fast; secrets only from env/mounted files, never defaults in code; `String()`/`LogValue()` redacts secrets. [slog LogValuer](https://pkg.go.dev/log/slog#LogValuer)
- Concurrency: every goroutine has a clear exit (ctx cancellation); prefer `errgroup` for fan-out with cancellation; guard shared state with `sync.Mutex` or confine it; run `-race`. [CodeReviewComments#goroutine-lifetimes](https://go.dev/wiki/CodeReviewComments#goroutine-lifetimes), [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup)
- Don't `panic` for normal errors; Recoverer is the last line, not control flow. [CodeReviewComments#dont-panic](https://go.dev/wiki/CodeReviewComments#dont-panic)
- Package names short, singular, meaningful; no `util`, `common`, `types`, `api`. [CodeReviewComments#package-names](https://go.dev/wiki/CodeReviewComments#package-names), [Effective Go](https://go.dev/doc/effective_go#names)
- `internal/` only hides packages from code outside `src/api/`; it does NOT stop `internal/solve` importing `internal/identity/transport`. Enforce the module boundary (`docs/architecture/ARCHITECTURE.md:20`) with an import-boundary linter (golangci-lint depguard, configured in `.golangci.yml`) plus review. [Go internal packages](https://go.dev/doc/go1.4#internalpackages), [Go modules layout](https://go.dev/doc/modules/layout)

## SOLID / design notes for this layer
- Layout as it exists today (check with `ls src/api/cmd src/api/internal` before adding anything):
```text
src/api/cmd/{api,aicheck,mailcheck}/   # entrypoints; cmd/api is the composition root
src/api/internal/app/                  # application wiring and server startup
src/api/internal/platform/             # cross-cutting, no business rules:
  apigen apispec apperr config cursor db dbgen etag httpx idempotency llm mailer pgtest subject
```
Domain modules are not created yet. When a slice adds one, use `src/api/internal/<module>/{domain,service,repository,transport}` (domain: entities, invariants, sentinel errors, no IO; service: use cases, authz call, tx orchestration; repository: sqlc-backed adapters and error translation; transport: Chi handlers, DTO mapping, problem+json) and do not create empty modules. The layer between transport and domain is named `service`. `db/` is the repository root `db/` (ADR-0005); sqlc output goes to `src/api/internal/platform/dbgen` and is never hand-edited.
- SRP: handler = decode/validate shape, call one service method, encode. Service = one use case. Repository = SQL only.
- DIP/ISP: service owns tiny interfaces (`QuestionStore{ Get(ctx,id) ...}`), not the whole `sqlc.Querier`; repository implements them.
- OCP: add a module/route by mounting a new `transport.Routes(r chi.Router)`; never edit neighbour modules.
- Transactions: service calls a `TxRunner.WithTx(ctx, func(q Store) error)` port; repository implements with `pgx.BeginTxFunc`/`BeginTx`, so services stay driver-agnostic.
- Cross-module reads go through a narrow exported interface (e.g. identity's `Subject` lookup), never another module's tables via its repository.

## TDD & tests for this layer
- RED first (`docs/TEST.md:29-31`): write the failing service test from the Given/When/Then in `docs/SPEC.md` before any handler.
- Domain/service: table-driven tests (`tests := []struct{name string; ...}` + `t.Run`) with hand-written fakes of the consumer interfaces; cover every error branch (100% statements, `docs/TEST.md:6`).
- Transport: `httptest.NewRecorder` + real Chi router; assert status, `Content-Type: application/problem+json`, no leaked internals, 413 on oversized body, 400/422 on malformed/unknown fields.
- Integration (`docs/TEST.md:16`): real ephemeral PostgreSQL in Docker, migrations applied, run service+repository end to end; no DB mocks for transaction/invariant claims (`docs/TEST.md:9`).
- Negative authz matrix per route: anonymous, member, owner, third party, moderator (`docs/TEST.md:36`).
- Concurrency: `go test -race ./...`; add a concurrent-request test for invariants like single accepted answer.
- Each scenario includes nominal, invalid, unauthorized, duplicate/idempotent, DB/network failure (`docs/TEST.md:26`).

## Anti-patterns (reject in review)
- SQL, authz decisions, or business branching inside handlers; handler calling `sqlc.Queries` directly.
- `context.Background()`/`TODO()` inside request paths; ctx stored in structs; no deadline on DB/HTTP calls.
- `err.Error()` string matching; returning `pgx`/`pgconn` errors past repository; swallowing errors; `panic` for validation.
- Logging request bodies, cookies, `Authorization`, CSRF tokens, passwords, emails, reset tokens.
- Owner/user ID taken from body/query instead of session subject.
- A module importing a neighbour's `transport` or `repository` package.
- Package-level DB pool/config globals; `init()` doing IO; `util`/`common` packages.
- Fire-and-forget goroutines without cancellation; unbounded worker spawn per request.
- New dependency (ORM, Redis/Valkey, router, validation lib) without ADR (`docs/DEV.md:64`).
- Coverage exclusions not in `COVERAGE_SCOPE.md`; mocks used to fake a PASS (`AGENTS.md:74`).

## Done checklist
- [ ] Slice opened with `/start-slice`; allowed paths respected.
- [ ] RED evidence recorded, then GREEN; tests not weakened.
- [ ] Handlers transport-only; service holds invariants; repository translates errors.
- [ ] Every DB/external call has a ctx deadline; server timeouts + graceful shutdown set.
- [ ] Errors wrapped `%w`, mapped once to problem+json; redaction test exists.
- [ ] `gofmt`, `go vet`, `staticcheck`, `golangci-lint` clean; `go test -race` green.
- [ ] Coverage 100.00% statements on first-party scope; any exclusion reviewed in `COVERAGE_SCOPE.md`.
- [ ] Real-PG integration + negative authz tests green (G3/G4).
- [ ] Closeout reports PASS/FAIL/NOT RUN per gate with exact commands (`AGENTS.md:33`).

## References
- Deep dive: [references/patterns.md](references/patterns.md)
- https://go.dev/doc/effective_go · https://go.dev/wiki/CodeReviewComments · https://go.dev/blog/go1.13-errors
- https://pkg.go.dev/log/slog · https://pkg.go.dev/context · https://pkg.go.dev/net/http
- https://pkg.go.dev/github.com/go-chi/chi/v5 · https://pkg.go.dev/github.com/go-chi/chi/v5/middleware
- https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool · https://go.dev/doc/modules/layout
