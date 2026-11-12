---
name: integration-testing
description: 226os integration and contract tests (G3/G4). Covers real ephemeral PostgreSQL, fresh migrations, isolated deterministic fixtures, httptest against a real DB, OpenAPI conformance and drift, database invariants, and the negative authorization matrix. Use when touching **/integration/**, **/*_integration_test.go, tests/contract/**.
---
# Integration, Contract and Invariant Tests (G3, authZ part of G4)

## Scope and triggers
- `**/integration/**`, `**/*_integration_test.go`, `tests/contract/**`. Schema, query and contract authoring belong to the `postgres-data` and `api-contract` skills. This skill owns the tests that prove them.
- Not here: pure unit logic (`tdd-unit`), browser journeys (`bdd-acceptance`), pipeline wiring (`ci-cd-delivery`).

## Project rules (binding)
- Integration means a real, ephemeral PostgreSQL under Docker plus HTTP tests. It is mandatory. Contract means real API responses checked against OpenAPI and the generated client. It is mandatory. `docs/TEST.md:16-17`.
- G3 is real PostgreSQL, API contract conformance and invariants. Its evidence is the suite logs and the OpenAPI diff. `AGENTS.md:49`, `docs/TEST.md:46`.
- 100% coverage does not replace DB integration tests. `docs/RULES.md:20`, `docs/TEST.md:9`.
- Invariants are enforced in the database: unique normalised email, one accepted answer, unique vote `(user_id,target_type,target_id)`, unique application `(user_id,contribution_need_id)`, no dangling relations, and the others listed. `docs/SPEC.md:110-116`, `docs/archive/spec-kit/CONTRACTS.md:23`. The data slice acceptance criteria are in its block in `TODO.md`.
- Migrations are versioned, immutable after merge, and forward-only. CI exercises a fresh DB and an upgrade from the previous schema. `docs/spec-kit/DATABASE_DESIGN.md:39-42`, `docs/archive/spec-kit/G0_DECISIONS.md:16`.
- Seeds are deterministic, synthetic, provenance-labelled, and never implied to be real. `docs/spec-kit/DATABASE_DESIGN.md:44`, `docs/archive/spec-kit/CONTRACTS.md:25`.
- sqlc output stays in sync, with a no-drift test. `docs/DEV.md:47`. SQL is parameterised only, with a context deadline on DB access. `docs/DEV.md:45`.
- OpenAPI 3.1 is the shared contract and drift is not tolerated. `docs/RULES.md:15`. Errors use RFC 9457 `application/problem+json`. `docs/SPEC.md:99`, `docs/archive/spec-kit/G0_DECISIONS.md:15`.
- Status mapping (ADR-0004): 400/401/403/404/409/413/422/429/500/503, with `Retry-After` on 429/503. `docs/spec-kit/API_CONTRACT.md` and ADR-0004.
- The authZ negative matrix covers anonymous, member, owner, moderator and third party for every sensitive row. `docs/spec-kit/AUTHORIZATION_MATRIX.md:21-42`, `docs/TEST.md:36`.
- Also required: CSRF rejection, unknown fields, oversized body, invalid cursor, replay, idempotency key with the same or a different hash, and redaction. `docs/spec-kit/API_CONTRACT.md`, required contract tests.
- Search is bounded with a stable tie-breaker and `EXPLAIN (ANALYZE, BUFFERS)` on representative data. `docs/SPEC.md:120-121`, `docs/spec-kit/DATABASE_DESIGN.md:34`.
- Run via `make test-integration` after reading its recipe. Never `docker compose down -v` outside approved scope. `AGENTS.md:120-125`.

## Expert practices
- The in-repo harness is `src/api/internal/platform/pgtest` (read `doc.go` first): it starts one throwaway digest-pinned PostgreSQL container through the docker CLI and always removes it. Reuse it rather than adding a second mechanism. For reference, the alternatives are testcontainers-go `postgres.Run(ctx, <image>, ...)` with explicit wait strategies (`postgres.BasicWaitStrategies()`) or a Compose service with a healthcheck. Pin images by digest. [testcontainers-go postgres](https://golang.testcontainers.org/modules/postgres/)
- Start one container per test package (`TestMain`), apply migrations from scratch with the real migration tool, then isolate per test.
- Per-test isolation: (a) open a transaction, run the test, roll back. This works only when the code under test accepts a `pgx.Tx`/`DBTX` and does not commit itself. (b) Clone a template DB per test (`CREATE DATABASE t_x TEMPLATE migrated`). (c) Use testcontainers `Snapshot`/`Restore`, avoiding the `postgres` DB name and registering the `pgx` driver. [testcontainers-go postgres](https://golang.testcontainers.org/modules/postgres/)
- Use (b) or (c) whenever the code commits, uses advisory locks, or opens several connections. Transactional rollback hides commit-time constraint errors (deferred constraints) and serialisation conflicts.
- Test invariants at the DB layer by violating them with raw SQL and asserting the SQLSTATE (`23505` unique, `23503` FK, `23514` check). Then test that the service maps each to the RFC 9457 `409`/`422`.
- Concurrency: race two goroutines on "accept answer" and "duplicate application/vote" against the real DB, then assert exactly one success. Unit fakes cannot prove this.
- HTTP: build the real router with real repositories and drive it through `httptest.NewServer`. Assert status, `Content-Type`, problem `type/code`, headers (`Retry-After`, cookies), and the DB state afterwards. [net/http/httptest](https://pkg.go.dev/net/http/httptest) (UNVERIFIED: not fetched)
- Contract: validate every response from the integration suite against `api/openapi.yaml` with a schema validator (tool per ADR, UNVERIFIED). Regenerate server/client stubs in CI and fail on `git diff --exit-code`.
- Make the authZ matrix table-driven from the `AUTHORIZATION_MATRIX.md` rows. Each `D` cell is a test that asserts the status and that nothing changed. `404` vs `403` follows the "intentionally undisclosed" rule (`API_CONTRACT.md:57`). Private lists must leak no counts (`AUTHORIZATION_MATRIX.md:39`). [OWASP WSTG authorization testing](https://owasp.org/www-project-web-security-testing-guide/) (UNVERIFIED: not fetched)
- Determinism: fixed UUIDs/timestamps from a seeded generator, an injected clock, `ORDER BY` with a tie-breaker in every list query, and no reliance on serial IDs or insertion order.
- Never point tests at a shared, staging or production DB. Credentials are throwaway and generated per run, never committed.

## Patterns
```go
//go:build integration

func TestMain(m *testing.M) { os.Exit(runWithPostgres(m)) } // starts PG, migrates fresh, snapshots

func TestAcceptAnswer_AuthzMatrix(t *testing.T) {
	cases := []struct {
		name       string
		actor      fixture.Actor
		wantStatus int
	}{
		{"anonymous", fixture.Anonymous, http.StatusUnauthorized},
		{"member non-author (Bob)", fixture.Bob, http.StatusForbidden}, // SPEC.md:24-28
		{"question owner (Alice)", fixture.Alice, http.StatusOK},
		// moderator row: per docs/spec-kit/AUTHORIZATION_MATRIX.md and ADR-0003 (hide/restore content, resolve error reports only)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := freshDB(t) // template clone or Restore in t.Cleanup
			srv := httptest.NewServer(app.Router(db))
			t.Cleanup(srv.Close)
			q, a := fixture.QuestionWithAnswer(t, db, fixture.Alice)
			before := acceptedAnswer(t, db, q)
			res := tc.actor.Do(t, srv, http.MethodPost, "/api/v1/questions/"+q+"/accept", body(a)) // with CSRF token
			assertProblemOr(t, res, tc.wantStatus)
			validateAgainstOpenAPI(t, res)
			if tc.wantStatus != http.StatusOK && acceptedAnswer(t, db, q) != before {
				t.Fatal("state mutated on denied request")
			}
		})
	}
}
```

## Anti-patterns (reject in review)
- SQLite/in-memory or mocked `pgx` presented as an "integration" test.
- A shared long-lived DB, test order dependence, or `TRUNCATE` races between parallel packages.
- Schema built from `schema.sql` instead of the real migrations. No fresh-plus-upgrade path.
- Editing a merged migration. Destructive migrations without an ADR or backup plan.
- Asserting only the status code: no problem body check, no DB-unchanged check, no OpenAPI validation.
- AuthZ tests covering only the happy path or only `anonymous`.
- Random data without a fixed seed. Lists asserted without a deterministic order.
- Leaking SQL, stack traces or existence of private objects in error bodies, untested (`API_CONTRACT.md:56`).
- Real-looking personal data or provider endorsements in seeds.

## Done checklist
- [ ] Fresh-DB migration and upgrade-from-previous both run in the suite.
- [ ] Every touched invariant has a DB-level violation test and a service/HTTP mapping test.
- [ ] AuthZ matrix rows for touched endpoints are fully covered, with DB-unchanged assertions.
- [ ] All integration responses pass OpenAPI validation. Generation diff is empty. sqlc no-drift passes.
- [ ] `make test-integration`: exit code, log path, PG image digest, target SHA recorded.
- [ ] EXPLAIN evidence attached for changed search queries (G6 input).
- [ ] Evidence under `docs/evidence/<TASK-ID>/`. Status reported as `PASS/FAIL/NOT RUN/BLOCKED`. Missing or unreviewed `api/openapi.yaml` or migrations means `BLOCKED`.

## References
- `docs/TEST.md`, `docs/SPEC.md`, `docs/spec-kit/{API_CONTRACT,AUTHORIZATION_MATRIX,DATABASE_DESIGN}.md`, archived `docs/archive/spec-kit/{CONTRACTS,G0_DECISIONS}.md`, `docs/decisions/` (ADR-0003, 0004, 0005)
- https://golang.testcontainers.org/modules/postgres/
- https://pkg.go.dev/net/http/httptest (UNVERIFIED) · https://owasp.org/www-project-web-security-testing-guide/ (UNVERIFIED)
- https://www.rfc-editor.org/rfc/rfc9457 (UNVERIFIED)
