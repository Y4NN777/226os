---
name: postgres-data
description: 226os PostgreSQL data layer (pgx v5 + sqlc, forward-only versioned SQL migrations, DB-enforced invariants, UUID keys, tsvector/GIN/pg_trgm search, bounded queries, EXPLAIN evidence, synthetic seeds). Use when touching db/**, db/migrations/**, db/queries/**, or sqlc config/generated code.
---
# PostgreSQL Data (pgx + sqlc)

## Scope and triggers
- `db/migrations/**`, `db/queries/**`, `db/seeds/**`, `sqlc.yaml`, repository adapters under `src/api/internal/*/repository/`.
- Status note: almost everything below is PROPOSED / ACCEPTED WITH CONDITIONS (ADR-0001, `docs/decisions/0001-g0-conditional-baseline.md:18`). Final schema, migration tool and FTS strategy wait on ADR-OPEN-005 (`docs/decisions/README.md:34`). Do not present proposals as accepted facts (`docs/spec-kit/TASKS.md:29`).

## Project rules (binding)
- PostgreSQL with FTS, GIN, pg_trgm is the initial data/search system — ACCEPTED BASE-005 (`docs/decisions/README.md:23`); single source of truth, no external secondary index (`docs/architecture/ARCHITECTURE.md:21`).
- Access via `pgx` + `sqlc`, no general-purpose ORM (`docs/DEV.md:7`, `:10`); SQL parameterized only (`docs/DEV.md:45`, `docs/spec-kit/CONTRACTS.md:24`).
- Keep generated sqlc in sync with `db/queries` and migrations; non-drift test (`docs/DEV.md:47`); generated output never hand-edited (`docs/spec-kit/DATABASE_DESIGN.md:43`).
- Migrations traceable and tested, constraints and indexes documented, FTS/GIN/pg_trgm, bounded queries and EXPLAIN plans (`docs/RULES.md:14`).
- Migrations: versioned SQL in `db/migrations/`, immutable after merge, prefer forward-only; destructive changes need expand/contract + backup/restore plan + human-approved ADR (`docs/spec-kit/DATABASE_DESIGN.md:39-40`; G0-DATA-001 `docs/spec-kit/G0_DECISIONS.md:16`).
- Each migration documents purpose, tables/indexes, lock risk, data volume, transaction behaviour, compatibility window, rollback/restore plan, owner (`docs/spec-kit/DATABASE_DESIGN.md:41`).
- CI runs fresh DB and upgrade-from-previous; release does dry run + backup check (`docs/spec-kit/DATABASE_DESIGN.md:42`, `docs/SHIP.md:27`). Never assume an app rollback undoes a data migration (`docs/SHIP.md:38`).
- Destructive migrations require explicit human approval (`AGENTS.md:85`, `:130`).
- Invariants enforced in DB: normalized unique email; one accepted answer per question; unique `(user_id,target_type,target_id)` vote; unique `(user_id,contribution_need_id)` application; normalized GitHub URL; bounded statuses; explicit visibility; no dangling relations (`docs/SPEC.md:108-116`, `docs/spec-kit/CONTRACTS.md:23`, TASK-DATA-001 AC `docs/spec-kit/TASKS.md:54`).
- Separate tables per domain + typed `content_relations`; no polymorphic content table (`docs/architecture/ARCHITECTURE.md:22`, `docs/spec-kit/DATABASE_DESIGN.md:7`, `:48`).
- UUID PKs generated server-side; IDs are not authorization (`docs/spec-kit/DATABASE_DESIGN.md:22`); `timestamptz` UTC `created_at`/`updated_at` (`:23`); explicit status + separate visibility (`:24`).
- No cascading delete by default for user-owned content; prefer archival (`docs/spec-kit/DATABASE_DESIGN.md:25-26`).
- Search (proposed): stored/generated `search_document tsvector` + GIN on questions/resources/projects; `simple` config until bilingual relevance benchmark; `pg_trgm` only on bounded user-facing fields; max query length, bounded results, stable ordering + deterministic tie-breaker (`docs/spec-kit/DATABASE_DESIGN.md:31-34`; G0-DATA-002 `docs/spec-kit/G0_DECISIONS.md:17`).
- Use `websearch_to_tsquery`; verify plans with `EXPLAIN (ANALYZE, BUFFERS)` on representative data (`docs/SPEC.md:120`; TASK-DATA-001 `docs/spec-kit/TASKS.md:56`).
- Seeds deterministic, synthetic, provenance-labelled; never imply real users, endorsements or verification (`docs/spec-kit/DATABASE_DESIGN.md:44`).
- DB on private network only (`docs/RULES.md:31`, `docs/architecture/ARCHITECTURE.md:24`).
- Doc conflicts: `docs/SPEC.md:119` says "migrations SQL up/down" vs forward-only in `docs/spec-kit/G0_DECISIONS.md:16` / `docs/spec-kit/DATABASE_DESIGN.md:52` and "reversible ou plan de restauration" in `docs/SHIP.md:28` -> follow G0-DATA-001 (forward-only, restore/redeploy), flag SPEC.md as stale. `docs/SPEC.md:120` "dictionnaires adaptés" vs `simple` fallback (`docs/spec-kit/DATABASE_DESIGN.md:32`) -> `simple` until benchmark. `db/` location: root `db/` (`docs/DEV.md:26`) vs `src/api/db` (`docs/DEV.md:24`) -> resolve in ADR-OPEN-005.
- Migration tool and PostgreSQL major version: OPEN, pin per ADR (`docs/spec-kit/DATABASE_DESIGN.md:56`).

## Expert practices
- Name constraints explicitly (`questions_one_accepted_answer_uq`) so repositories map SQLSTATE + constraint name to domain errors. [CREATE TABLE](https://www.postgresql.org/docs/current/sql-createtable.html), [Error codes](https://www.postgresql.org/docs/current/errcodes-appendix.html)
- One accepted answer: partial unique index `CREATE UNIQUE INDEX ... ON answers(question_id) WHERE accepted;` (or `accepted_answer_id` FK on `questions`). [Partial indexes](https://www.postgresql.org/docs/current/indexes-partial.html)
- Bounded statuses with `CHECK (status IN (...))` or enums; prefer `CHECK`/lookup tables for easier forward-only evolution. [Constraints](https://www.postgresql.org/docs/current/ddl-constraints.html)
- Normalized email uniqueness on the normalized column/expression (e.g. `lower(email)`); keep normalization rule versioned in one place (`docs/spec-kit/DATABASE_DESIGN.md:27`). [Indexes on expressions](https://www.postgresql.org/docs/current/indexes-expressional.html)
- UUID default `gen_random_uuid()` (core); time-ordered `uuidv7()` only if the pinned PG major provides it — UNVERIFIED for chosen version. [UUID functions](https://www.postgresql.org/docs/current/functions-uuid.html)
- Generated column with explicit two-arg config: `GENERATED ALWAYS AS (setweight(to_tsvector('simple', coalesce(title,'')),'A') || setweight(to_tsvector('simple', coalesce(body,'')),'B')) STORED` + `USING GIN`. Never rely on `default_text_search_config`. [textsearch-tables](https://www.postgresql.org/docs/current/textsearch-tables.html), [textsearch-controls](https://www.postgresql.org/docs/current/textsearch-controls.html)
- `websearch_to_tsquery` never raises syntax errors on user input; still cap input length before calling. Ranking reads each matching tsvector and is costly — bound candidates before `ts_rank`. [textsearch-controls](https://www.postgresql.org/docs/current/textsearch-controls.html)
- `pg_trgm`: `GIN (col gin_trgm_ops)` for `%`, `ILIKE`, `~`; GIN cannot serve `<->` ORDER BY LIMIT (GiST can); short patterns with no trigrams degrade to full-index scans -> enforce a minimum length. [pg_trgm](https://www.postgresql.org/docs/current/pgtrgm.html)
- Keyset (cursor) pagination: `WHERE (sort_key, id) < ($1, $2) ORDER BY sort_key DESC, id DESC LIMIT $3`, backed by a matching composite index; never `OFFSET` for public lists. [Row comparison](https://www.postgresql.org/docs/current/functions-comparisons.html#ROW-WISE-COMPARISON), [Indexes and ORDER BY](https://www.postgresql.org/docs/current/indexes-ordering.html)
- Set `statement_timeout` (role or per-transaction `SET LOCAL`) as a backstop to Go ctx deadlines. [Client config](https://www.postgresql.org/docs/current/runtime-config-client.html)
- Concurrency: use `SELECT ... FOR UPDATE` or the unique constraint + `ON CONFLICT` for check-then-act; never rely on app-only checks. [INSERT ON CONFLICT](https://www.postgresql.org/docs/current/sql-insert.html), [Explicit locking](https://www.postgresql.org/docs/current/explicit-locking.html)
- Index creation on populated tables: `CREATE INDEX CONCURRENTLY` (cannot run inside a transaction block — migration must be non-transactional, document it). Add FK/CHECK as `NOT VALID` then `VALIDATE CONSTRAINT` to shorten locks. [CREATE INDEX](https://www.postgresql.org/docs/current/sql-createindex.html), [ALTER TABLE](https://www.postgresql.org/docs/current/sql-altertable.html)
- Set `lock_timeout` in migrations to fail fast instead of queueing behind long transactions. [Client config](https://www.postgresql.org/docs/current/runtime-config-client.html)
- sqlc: `version: "2"`, `engine: postgresql`, `gen.go.sql_package: pgx/v5`, `emit_interface` only if needed (services define their own ports); schema = migrations dir. Drift check = `sqlc generate` then `git diff --exit-code`; use `sqlc vet` rules for lint. `sqlc diff`/`sqlc verify` behaviour UNVERIFIED on pinned version; `sqlc generate` + `git diff --exit-code` is the guaranteed check. [sqlc config](https://docs.sqlc.dev/en/latest/reference/config.html), [sqlc vet](https://docs.sqlc.dev/en/latest/howto/vet.html)
- pgx: one `pgxpool.Pool` per process (concurrency-safe), configured via `ParseConfig`/`NewWithConfig`; ping on startup; `pool.Close()` on shutdown; transactions via `BeginTx`/`pgx.BeginTxFunc`. [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool), [pgx](https://pkg.go.dev/github.com/jackc/pgx/v5)
- `EXPLAIN (ANALYZE, BUFFERS)` executes the statement — run it only on disposable/representative datasets, wrap writes in `BEGIN ... ROLLBACK`. [EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [Using EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html)

## SOLID / design notes for this layer
- The DB enforces invariants; Go domain validates early for UX. Both exist; the DB is the backstop under concurrency.
- One query file per module (`db/queries/solve/*.sql`); sqlc output stays inside that module's repository package — no cross-module table access.
- Repository adapts sqlc types to domain types (ISP: expose only methods the service port needs).
- Typed `content_relations` with per-pair FKs or checked existence; no free-form `target_type text` without CHECK.

## TDD & tests for this layer
- RED first: a constraint test that inserts a violating row and expects SQLSTATE + constraint name (e.g. second accepted answer -> `23505`).
- Real PostgreSQL only (ephemeral Docker), migrations applied from zero (`docs/TEST.md:16`); no SQL mocks for invariant claims.
- Migration tests: fresh install; upgrade from previous release schema with representative rows; re-run idempotence of the runner.
- Concurrency test: N goroutines vote/apply/accept concurrently -> exactly one row / one accepted answer.
- Search tests: filters (country, tag, type, language), stable order + tie-break, pagination with no dupes/gaps, max length rejected, SQL-injection strings treated as data (`docs/TEST.md:37`), private/unpublished never returned (`docs/spec-kit/AUTHORIZATION_MATRIX.md:23`).
- Evidence: EXPLAIN (ANALYZE, BUFFERS) output per search/list query in `docs/evidence/TASK-DATA-001/` (`docs/spec-kit/TASKS.md:58`).
- sqlc drift check in G1 (`AGENTS.md:47`).

## Anti-patterns (reject in review)
- Editing a merged migration; down migrations that drop data presented as rollback.
- String-concatenated SQL, `fmt.Sprintf` into queries, dynamic `ORDER BY` from user input without allowlist.
- `ON DELETE CASCADE` on user-owned content; polymorphic `content` table; integer public IDs.
- Uniqueness or "one accepted" enforced only in Go.
- `OFFSET` pagination; unbounded `LIMIT`; `ILIKE '%x%'` without trigram index or length floor; `to_tsvector(col)` one-arg form.
- Hand-edited sqlc output; generated code drifting from queries.
- Seeds with real names/emails/providers or claims of verification; secrets in seed or fixture files.
- Running EXPLAIN ANALYZE of writes against shared/staging data.

## Done checklist
- [ ] Migration header filled (purpose, locks, volume, tx mode, compat, restore, owner).
- [ ] Every SPEC §5 invariant touched has a named constraint + RED/GREEN test.
- [ ] Fresh + upgrade migration paths green on real PG.
- [ ] sqlc regenerated, no diff; `sqlc vet` clean.
- [ ] Indexes justify each query; EXPLAIN evidence stored.
- [ ] Queries bounded (limit, length, timeout); stable ordering with UUID tie-break.
- [ ] Seeds synthetic and labelled; no destructive change without approved ADR.

## References
- Deep dive: [references/migrations-and-search.md](references/migrations-and-search.md)
- https://www.postgresql.org/docs/current/textsearch.html · https://www.postgresql.org/docs/current/pgtrgm.html
- https://docs.sqlc.dev/en/latest/ · https://pkg.go.dev/github.com/jackc/pgx/v5
- https://www.postgresql.org/docs/current/ddl-constraints.html · https://www.postgresql.org/docs/current/sql-explain.html
