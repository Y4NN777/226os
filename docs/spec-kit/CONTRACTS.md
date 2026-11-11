# 226OS Implementation Contracts

**Status:** PROPOSED / OPEN; human approval required before parallel implementation

Detailed G0 proposals are split into `AUTHORIZATION_MATRIX.md`, `API_CONTRACT.md`, and `DATABASE_DESIGN.md`. This summary does not approve their defaults.

## API contract

- Base path `/api/v1`; JSON UTF-8; RFC3339 UTC timestamps; opaque UUID identifiers.
- Errors use `application/problem+json` (RFC 9457) with stable type, title, status, detail policy, and correlation identifier where appropriate.
- Cursor pagination has a documented maximum; filters are validated; ordering is deterministic.
- Initial resources: sessions, `users/me`, profiles, questions, answers, votes, resources, projects, contribution needs, applications, and search.
- Session cookies are secure, HttpOnly, SameSite-appropriate, rotated/revoked according to the proposed auth design; mutations require CSRF protection.
- Server-side authorization is required for ownership, collaborator, and private application access.
- `api/openapi.yaml`, DTOs, error catalogue, and generated-client checks are required before TASK-API-001 can enter implementation; see `API_CONTRACT.md`.

**OPEN:** exact verbs, schemas, status codes, auth lifecycle, rate limits, idempotency keys, and correlation logging policy. Proposed defaults are in `G0_DECISIONS.md`.

## Data and migration contract

- PostgreSQL is the initial source of truth and search engine.
- SQL migrations are versioned, reviewed, and tested for fresh install, upgrade, and rollback/restore strategy.
- Required invariants include unique normalized email, at most one accepted answer per question, unique user/target vote, normalized public GitHub URL, unique user/need application, bounded statuses, explicit visibility, and non-pending content relations.
- `pgx` and `sqlc` are used; SQL is parameterized; multi-table mutations are transactional.
- Seed data must be deterministic, clearly synthetic, and never presented as real activity, endorsement, or provider verification.

**OPEN:** migration tool, generated-column versus trigger FTS strategy, search dictionaries for French/English, seed ownership, deletion/archive policy, and compatibility policy. Proposed design is in `DATABASE_DESIGN.md`.

## Search contract

- Search spans questions, resources, and projects with explicit bounds, stable ordering, country/tag/type/language filters, PostgreSQL full-text search, GIN indexes, and `pg_trgm` typo tolerance.
- Query cost is bounded and representative plans are captured with `EXPLAIN (ANALYZE, BUFFERS)`.
- No semantic search, summaries, or AI claims are P0. Any future assistance must expose sources and limitations.

**OPEN:** relevance dataset, score/tie-break policy, language tokenization, and measured latency budget approval.

## Frontend contract

- React 19, TypeScript strict, Vite, TanStack Router/Query/Form, Meta Astryx as the primary component foundation, and Tailwind only as complementary layout styling.
- Canonical first routes are Explore, Solve list/detail/create, Discover list/detail, Build list/detail/create/need/application, sign-in/sign-up, and public profile. Exact URL slugs and public/authenticated boundaries are `OPEN`.
- Shared UI must cover loading, empty, error, success, focus, keyboard, dialog, code block, tags, country, technology, source, and relation states.
- FR/EN locale choice must preserve original user content and unsaved form state. Locale persistence and fallback strategy are `OPEN`.
- Required responsive checkpoints: 360, 768, 1024, 1440px and 200% zoom. Core flows must work on throttled mobile networks without essential heavy media.
- Astryx release, CLI syntax, component APIs, theme token mapping, CSS import order, and lockfile are verified only after installation; illustrative snippets in `DESIGN.md` are not normative APIs.

## Security contract

Threat cases include auth/session abuse, CSRF, IDOR, Markdown/XSS, SQL injection, GitHub URL SSRF, quota exhaustion, upload risks if later added, secrets in logs/fixtures, privacy leakage, and unauthorized application access.

Required evidence includes negative authorization tests, input validation tests, CSRF/session tests, XSS/SSRF tests, rate-limit behavior, SAST, SCA, secret scan, and license scan. No code execution from snippets or user Markdown.

**OPEN:** authentication provider/recovery/email verification, data classification and retention, moderation ownership, incident response owner, and exact rate-limit policy. Proposed matrix is in `AUTHORIZATION_MATRIX.md`.

## Operational contract

- Docker multi-stage images; Compose dev/prod distinct; Traefik is the only public edge; PostgreSQL remains private.
- Health and readiness endpoints, structured logs without secrets, resource limits, immutable image references, and encrypted backup/restore evidence are required.
- Valkey and S3-compatible storage remain candidates only; they are not P0 dependencies without an accepted ADR.
- GitHub Actions and Make are the reproducible interfaces for gates; no undocumented success path.

**OPEN:** Datacloud product/account, DNS, storage, budget, CI environment, image registry policy, and deployment approver.
