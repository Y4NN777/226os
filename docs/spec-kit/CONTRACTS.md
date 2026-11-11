# 226os Implementation Contracts

**Status:** ACCEPTED (ADR-0003 to ADR-0007, 2026-11-11). Parallel implementation may start from these contracts; execution evidence is G1 to G7.

The detail lives in `AUTHORIZATION_MATRIX.md`, `API_CONTRACT.md`, `DATABASE_DESIGN.md` and `api/openapi.yaml`. This summary only points to them.

## API contract

- Base path `/api/v1`; JSON UTF-8; RFC3339 UTC timestamps; opaque UUID identifiers.
- Errors use `application/problem+json` (RFC 9457) with stable type, title, status, detail policy, and correlation identifier where appropriate.
- Cursor pagination has a documented maximum; filters are validated; ordering is deterministic.
- Initial resources: `/auth/*`, `/me`, `/profiles/{username}`, questions, answers, votes, resources, projects, contribution needs, applications, `/learn/*`, `/moderation/*` and `/search` (paths as in `api/openapi.yaml`).
- Session cookies are `__Host-226os_session` (deployed) or `226os_session` (local-dev), HttpOnly, Secure when deployed, SameSite=Lax, rotated/revoked per ADR-0003; unsafe methods require the `X-CSRF-Token` header.
- Server-side authorization is required for ownership, maintainer and accepted-applicant ("collaborator" relationship) rights, and for private application access.
- `api/openapi.yaml`, DTOs, error catalogue, and generated-client checks are required before TASK-API-001 can enter implementation; see `API_CONTRACT.md`.

Verbs, schemas, status codes, auth lifecycle, rate limits and idempotency are decided in ADR-0003 and ADR-0004; `request_id` is the correlation identifier.

## Data and migration contract

- PostgreSQL 18 is the initial source of truth and search engine; primary keys default to native `uuidv7()`.
- SQL migrations are goose forward-only (no `Down`), sequential `00001_*`, immutable after merge, under root `db/`, and tested for fresh install, upgrade and restore.
- Required invariants include unique normalized email, at most one accepted answer per question, unique user/target vote, normalized public GitHub URL, unique user/need application, bounded statuses, explicit visibility, and non-pending content relations.
- `pgx` and `sqlc` are used; SQL is parameterized; multi-table mutations are transactional.
- Seed data must be deterministic, clearly synthetic, and never presented as real activity, endorsement, or provider verification.

Migration tool, generated-column search vectors, French accent handling, seeds, deletion/anonymization and status model are decided in ADR-0005; the design is in `DATABASE_DESIGN.md`.

## Search contract

- Search spans questions, resources, projects and Learn articles with explicit bounds, stable ordering, country/tag/type/language filters, PostgreSQL full-text search (`simple` plus an immutable `unaccent` wrapper), GIN indexes, and `pg_trgm` typo tolerance.
- Query cost is bounded and representative plans are captured with `EXPLAIN (ANALYZE, BUFFERS)`.
- No semantic search, summaries, or AI claims are P0. Any future assistance must expose sources and limitations.

The relevance dataset and measured latency budgets are G6 evidence; they may tune ranking but do not reopen the design.

## Frontend contract

- React 19, TypeScript strict, Vite, TanStack Router/Query/Form, Meta Astryx as the primary component foundation, and Tailwind only as complementary layout styling.
- Canonical first routes are Explore, Solve list/detail/create, Discover list/detail, Build list/detail/create/need/application, sign-in/sign-up, and public profile. URL slugs follow the API resource names; public reads need no session and every mutation needs one.
- Shared UI must cover loading, empty, error, success, focus, keyboard, dialog, code block, tags, country, technology, source, and relation states.
- FR/EN locale choice must preserve original user content and unsaved form state. The chosen locale is persisted in the browser and falls back to the original content language for user content.
- Required responsive checkpoints: 360, 768, 1024, 1440px and 200% zoom. Core flows must work on throttled mobile networks without essential heavy media.
- The Astryx release is pinned at bootstrap and its CLI syntax, component APIs, theme token mapping and CSS import order are verified after installation (ADR-0007); illustrative snippets in `DESIGN.md` are not normative APIs.

## Security contract

Threat cases include auth/session abuse, CSRF, IDOR, Markdown/XSS, SQL injection, GitHub URL SSRF, quota exhaustion, upload risks if later added, secrets in logs/fixtures, privacy leakage, and unauthorized application access.

Required evidence includes negative authorization tests, input validation tests, CSRF/session tests, XSS/SSRF tests, rate-limit behavior, SAST, SCA, secret scan, and license scan. No code execution from snippets or user Markdown.

Authentication, recovery, email verification, rate limits and moderation scope are decided in ADR-0003; ownership of incident response is the interim owner (ADR-0002). The matrix is `AUTHORIZATION_MATRIX.md`.

## Operational contract

- Docker multi-stage images; one profile-based Compose tree with distinct `local-dev`, `staging-test` and `production-live` profiles (ADR-0006); Traefik is the only public edge; PostgreSQL remains private.
- Health and readiness endpoints, structured logs without secrets, resource limits, immutable image references, and encrypted backup/restore evidence are required.
- Valkey and S3-compatible storage are excluded at P0; adopting either needs a new ADR.
- GitHub Actions and Make are the reproducible interfaces for gates; no undocumented success path.

Datacloud is the host for `staging-test` and `production-live`; GHCR is the registry. Account/product, DNS, budget and approver evidence are G7 execution evidence.
