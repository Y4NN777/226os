# 226os P0 tracker

> **Rules for this file**
>
> - This is the single live tracker. The plan, rules R1 to R11, blockers (B), contradictions (C), open decisions (O) and risks (K) are in [ROADMAP.md](ROADMAP.md).
> - Update this file **in the same PR as the work**, one line per change.
> - **Never rewrite someone else's checkbox or status.** Add a dated line under it instead.
> - When a slice merges, set its status to `DONE (PR #n, YYYY-MM-DD)` and add a line to the Done log.
> - People are named by GitHub login only.
> - Contracts are linked, not copied:
>   - [api/openapi.yaml](api/openapi.yaml) (operations, `x-authz`)
>   - [docs/SPEC.md](docs/SPEC.md) (BDD)
>   - [docs/spec-kit/AUTHORIZATION_MATRIX.md](docs/spec-kit/AUTHORIZATION_MATRIX.md)
>   - [docs/spec-kit/DATABASE_DESIGN.md](docs/spec-kit/DATABASE_DESIGN.md)

## Status legend

| Item | Values |
|---|---|
| Checkbox | `- [ ]` open, `- [x]` done (with PR number) |
| Slice status | `PLANNED`, `IN PROGRESS`, `IN REVIEW`, `DONE (PR #n, date)`, `BLOCKED (reason)` |
| S10 only | `CONDITIONAL`, `NOT BUILT (P2)` |
| Gate results | `PASS`, `FAIL`, `NOT RUN`, `BLOCKED`. `NOT RUN` is never `PASS`. |

BDD references such as `US-SOLVE-001/2` number the scenarios of a SPEC.md story in order. SPEC.md has no scenario ids of its own.

**Definition of done for every slice (DoD-common):**

- R1 to R11 from ROADMAP section 5 are respected.
- `make quality` is green in CI on the PR head SHA. It covers `doctor generate lint typecheck test-unit coverage test-integration security licenses`, which are gates G1 to G4.
- The PR description includes:
  - the RED test names, then GREEN;
  - Go coverage at 100.00% (and web above 70% where touched);
  - the `make security` route-table output;
  - the story id and reviewer;
  - `Commands executed` with exit codes.
- Every new mounted route is either in `scripts/public-routes.txt` with a reason, or covered by negative authz tests that the gate recognises (B1).
- Generated code is never hand-edited. Data slices regenerate sqlc into `src/api/internal/platform/dbgen/**` with `make generate`. Two PRs that touch the same generated files must rebase on `develop` and rerun `make generate` before merging, or G1 drift fails.

## Now

One line per person. Each person edits only their own line.

| Login | Next action | Updated |
|---|---|---|
| `Y4NN777` | Do H1, H3, H4 and H6 (ROADMAP section 8). Then S0: the authz gate extension (B1) first, then pgxpool, `/readyz` and the apigen mount. By Fri afternoon: rule on ADR-0009 (H15), C15, C16, C2 and C14; choose the sources (H17); name the S12 builder (O10) and the E5 backend owner (O9). | 2026-11-12 |
| `NICE-DEV226` | Send the identity record (H1). Start S1 RED tests for the domain and service with fakes. Mount nothing before S0 merges. | 2026-11-12 |
| `mibienpanjoe` | Send the identity record (H1). Write the S2 sqlc queries with local EXPLAIN. Raise C2 (Learn `topic`) with the owner. Open the E2 taxonomy migration PR (`make migrate-new`, plus the pgtest lists). | 2026-11-12 |
| `PANK4SS` | Accept the invitation (H2) and send the identity record (H1). Build the S7 shell, i18n and shared components on typed fixtures from `src/web/src/api/schema.d.ts`. E4 comes later (after S3, and after C2 for the feeds). Then the pillar navigation and copy (S7), and the `connect` page once the E5 contract merges. | 2026-11-12 |
| S12 builder (OPEN) | Wait for ADR-0009 and the owner's naming (O10). Then build the `src/api/cmd/ingest` skeleton over the service layer, with the GitHub connector first. | 2026-11-12 |

## Shared pieces (O2: where each lives is decided by the backend lead at kickoff)

Each line below is a **proposal**. The lead confirms or reassigns it.

- [ ] **Session resolver with an anonymous default.**
  - Proposal: S0 adds a `Subject` port with an anonymous default, and S1 provides the session-backed implementation.
  - Used by optional-session reads: `listProjects`, `getProject`, `listContributionNeeds`, `getTrainingSeries`, `getProfile`, `search`, `getSession`.
- [ ] **PostgreSQL rate limiter** (`rate_limit_buckets`, ADR-0003 windows).
  - Proposal: S1 builds it as a reusable port.
  - Other scopes that use it:
    - anonymous search, 60 per IP per minute (S2);
    - authenticated writes, 60 per user per minute (S3, S4, S5);
    - GitHub import, 10 per user per hour (S5);
    - assistant (S9).
- [ ] **`EMAIL_NOT_VERIFIED` check.** S1 exposes `email_verified` on the subject, and each publishing service checks it.
- [ ] **`links[]` relation writer and resolved-links reader** (`content_relations`, ADR-0005 point 13).
  - Proposal: `mibienpanjoe` builds the reader in S2 (`getResource` resolves links) and the writer in S4, and S3 and S5 reuse them.
  - The schema `links` field appears in `QuestionCreate`, `QuestionUpdate`, `ProjectCreate`, `ProjectUpdate` and `ArticleCreate` (`api/openapi.yaml:3313,3346,3890,3911,4236`).
- [ ] **`audit_log` writer.** Proposal: S6 builds it. S1 security events also go to `audit_log` (`docs/spec-kit/DATABASE_DESIGN.md:332`), so the lead decides the order.
- [ ] **Idempotency-Key.** S0 owns it and it must merge before S3, S4 and S5 writes.
  - It covers six operations: `createQuestion`, `createAnswer`, `createProject`, `createContributionNeed`, `createApplication`, `createLearnArticle`.
  - `castVote` does not bind the header.

---

## S0 — API core

- **Owner:** `Y4NN777`
- **Reviewer:** `NICE-DEV226` (proposed; the owner's PRs are reviewed by a backend developer)
- **Branches:** `feature/platform/authz-gate`, `feature/platform/api-mount`, `feature/platform/idempotency`, `feature/platform/mailer`
- **Depends on:** H1 and H4
- **Operations:** none of its own. It mounts the generated `apigen` interface (47 operations).

**Tasks**

- [ ] **B1.** Extend `scripts/check-authz-routes.sh` and `scripts/authz/routes_check.go.txt` so that a protected route passes only when it is bound to named negative tests that ran and passed, or mount only implemented operations. Owner review. The gate must not be weakened.
- [ ] **O3.** Decide how the 4 P1 routes are handled: kept at 501 and covered by the gate, or not mounted.
- [ ] Add `pgxpool` and `DATABASE_URL` to the config (`src/api/internal/platform/config`), and pass `DATABASE_URL` to `api` in `compose.yaml` (today `compose.yaml:90` omits it).
- [ ] Add `GET /readyz` (DB ping with a deadline) to `scripts/public-routes.txt` with its reason. Decide edge routing for B2 (`deploy/traefik/dynamic.edge.yml:6`).
- [ ] Mount `apigen` with `HandlerWithOptions` under `/api/v1`, with a problem+json `ErrorHandlerFunc`:
  - missing required `If-Match` returns 428 `PRECONDITION_REQUIRED`;
  - bad parameters return 400 `MALFORMED_REQUEST` or 422 `VALIDATION_ERROR` (ADR-0004);
  - `METHOD_NOT_ALLOWED` is outside the enum (C5).
- [ ] If the `httpx.NewRouter` signature changes, update the overlay in the same PR (`routes_check.go.txt:32`).
- [ ] Request validation:
  - `additionalProperties: false` and `DisallowUnknownFields`;
  - `maxLength` and `maxItems`;
  - 1 MiB body limit, returning 413 `PAYLOAD_TOO_LARGE`;
  - `Content-Type: application/json` on mutations (the error code for this is open, C5).
- [ ] Signed opaque cursors: HMAC key from env, sort key plus UUID plus filter hash, 400 `INVALID_CURSOR`. `limit` defaults to 20 and returns 422 above 50.
- [ ] `ETag` and `If-Match` helper returning 412 and 428. It is used by `updateQuestion` (S3).
- [ ] Idempotency-Key on the six operations listed in Shared pieces:
  - stored in `idempotency_keys` for 24 hours;
  - bound to subject, route and request hash;
  - same key and hash replays the stored response;
  - same key with a different hash returns 422 `IDEMPOTENCY_KEY_REUSED`;
  - a request still in flight returns 409 `IDEMPOTENCY_IN_PROGRESS`;
  - a purge job removes expired rows.
- [ ] Transaction runner (multi-table mutations in one transaction).
- [ ] Domain error to problem+json mapping table (one place).
- [ ] Mailer port, SMTP adapter and fake:
  - a digest-pinned Mailpit service in the `local-dev` profile of `compose.yaml`;
  - `SMTP_*` variables in `.env.example` (their current presence is UNVERIFIED);
  - the Gmail app password only in the staging and production env files.
- [ ] Per-route timeout decision for B3 (`httpx/router.go:14,23`; matters for S9 and S10).
- [ ] Session resolver port with an anonymous default (Shared pieces).

**Tables:** `idempotency_keys` (migration `00004`).

**Acceptance cases** (ADR-0004, "Evidence"):

- malformed JSON;
- unknown field;
- oversized body;
- invalid, tampered or mismatched cursor;
- `limit` 51 returns 422;
- idempotency with the same and with a different hash;
- concurrent duplicate;
- missing `If-Match`, and `If-Match` mismatch;
- redaction (no SQL or hosts in 500 responses);
- `/readyz` fails when the DB is down.

**Allowed paths:**

- `src/api/cmd/**`
- `src/api/internal/app/**`
- `src/api/internal/platform/**` (apigen is generated only)
- `scripts/**`
- `compose.yaml`
- `deploy/**`
- `.env.example`
- `mk/**` (one change per PR)

**Negative authz:** the gate itself (B1). Every mounted non-public route must stay red until its tests exist.

**DoD:** DoD-common, plus a `make up-local-dev` run where `/readyz` returns 200 and Mailpit receives a test mail (the evidence is the log in the PR).

**Stop condition:** stop and ask the owner if B1 cannot be solved without weakening the gate.

**Status:** PLANNED

---

## S1 — Identity

- **Owner:** `NICE-DEV226`
- **Reviewer:** `Y4NN777` (lead; auth surface, ADR-0008 "Consequences")
- **Branches:** `feature/identity/auth-sessions`, `feature/identity/profiles`
- **Depends on:** S0. Argon2id parameters are pinned only after the benchmark on the target host (S8, ADR-0003).

**Operations (11).** See `x-authz` in `api/openapi.yaml:44-435`.

- [ ] `getCsrfToken` `GET /auth/csrf` (public; pre-session token plus `__Host-226os_csrf` or `226os_csrf`)
- [ ] `signUp` `POST /auth/sign-up`
  - generic 202, no enumeration;
  - 5 per IP per hour;
  - sends the verification email.
- [ ] `signIn` `POST /auth/sign-in`
  - generic `INVALID_CREDENTIALS` with equal timing;
  - 5 failures per account and 20 per IP per 15 minutes, failing closed;
  - rotates the session.
- [ ] `signOut` `POST /auth/sign-out` (revokes only the caller's own session)
- [ ] `getSession` `GET /auth/session` (never 401)
- [ ] `verifyEmail` `POST /auth/verify-email` (single use, 24 hours, 422 `INVALID_TOKEN`)
- [ ] `requestPasswordReset` `POST /auth/password-reset/request` (always 202; 3 per account and 10 per IP per hour)
- [ ] `confirmPasswordReset` `POST /auth/password-reset/confirm` (30 minutes; revokes all of the user's sessions)
- [ ] `getMe` `GET /me` (the only place the email is returned)
- [ ] `updateMyProfile` `PATCH /me/profile` (own profile only; JSON Merge Patch)
- [ ] `getProfile` `GET /profiles/{username}` (field visibility; email never returned; private profile 404)

**Also:**

- [ ] Argon2id m=64 MiB, t=3, p=1, with a semaphore. Benchmark it, and check K6 (128M container limit).
- [ ] Cookie flags per profile (ADR-0003).
- [ ] CSRF with `X-CSRF-Token` plus an `Origin` / `Sec-Fetch-Site` check.
- [ ] Session timeouts: 2 hours idle, 7 days absolute.
- [ ] O4: no resend-verification operation exists; ask the owner.

**Tables:** `users`, `user_credentials`, `profiles`, `user_skills`, `sessions`, `email_verifications`, `password_resets`, `rate_limit_buckets`. It also reads `countries` and `tags` (`00002` to `00004`; sqlc in `db/queries/identity.sql` and `platform.sql`).

**Acceptance:**

- `US-IDENTITY-001/2` (language switch, with S7);
- ADR-0003 "Evidence": session expiry, rotation and revocation; CSRF missing, wrong, cross-origin and pre-session cases; enumeration and timing; 429 with `Retry-After`; the token never appears in logs.

**Prototype pages:** `connexion`, `inscription`, `verifier-email`, `mot-de-passe-oublie`, `reinitialiser`, `bienvenue`, `profile`, `parametres` (`auth.html` is a legacy redirect).

**Allowed paths:** `src/api/internal/identity/**`, `db/queries/identity.sql`, `db/queries/platform.sql`, `src/api/internal/platform/dbgen/**` (generated by `make generate` only), and `scripts/public-routes.txt` (anonymous auth routes only, with reasons).

**Negative authz** (`AUTHORIZATION_MATRIX.md:25-28`):

- nobody can revoke another user's session;
- anonymous `GET /me` and `PATCH /me/profile` return 401;
- cross-user access is impossible;
- a private profile returns 404 and the email never leaks;
- rate limits, generic errors and CSRF are enforced on the anonymous endpoints.

**DoD:** DoD-common, plus cookie inspection per profile and a redacted email-flow log through Mailpit.

**Stop condition:** stop if ADR-0003 values would need to change.

**Status:** PLANNED

---

## S2 — Discover and search, demo seeds

- **Owner:** `mibienpanjoe`
- **Reviewers:** `NICE-DEV226` (proposed), and `Y4NN777` for the seed content and for `mk/50-ops.mk` (owner area; the owner curates demo content, `DATABASE_DESIGN.md:1299`)
- **Branches:** `feature/discover/catalogue-read`, `feature/discover/search`, `feature/discover/demo-seeds`
- **Depends on:** S0. The error report and the anonymous-search 429 also need S1.

**Operations (5)**

- [ ] `listResources` `GET /resources` (public; provenance on each item)
- [ ] `getResource` `GET /resources/{resourceId}` (public; resolved related links, visible items only)
- [ ] `listResourceSources` `GET /resources/{resourceId}/sources` (public; a stale source is never shown as verified)
- [ ] `createResourceErrorReport` `POST /resources/{resourceId}/error-reports` (member; reporter kept private; rate limited)
- [ ] `search` `GET /search`
  - `q` length 2 to 200 (C4: ADR-0005 says 100);
  - `per_group` 1 to 20;
  - only published and visible content;
  - anonymous callers limited to 60 per IP per minute.

**Also**

- [ ] Demo seeds in `db/seeds/`:
  - fixed UUIDs and `provenance='demo'`;
  - `@demo.226os.invalid` accounts and `demo-` usernames;
  - `example.org` URLs only;
  - no fake endorsements.
- [ ] `make seed` (`mk/50-ops.mk:50`) refuses unless `APP_ENV` is `local-dev` or `staging-test`. Add a check that `production-live` has zero demo rows (`DATABASE_DESIGN.md:1296`).
- [ ] Bounded sqlc queries with `EXPLAIN (ANALYZE, BUFFERS)` evidence, plus a bilingual relevance dataset (ADR-0005 point 6).
- [ ] Resolved-links reader (Shared pieces, proposal).

**Tables:** `resource_categories`, `providers`, `resources`, `resource_countries`, `resource_sdk_languages`, `resource_sources`, `resource_tags`, `error_reports`, `countries`, `tags`. Search reads `questions`, `answers`, `projects` and `learn_articles`.

**Acceptance:** `US-DISCOVER-001/1`. The search steps of `US-SOLVE-001/1` and `US-LEARN-001/2`.

**Prototype pages:** `discover`, `resource`, `recherche`, and part of `accueil`.

**Allowed paths:**

- `src/api/internal/discover/**`
- `src/api/internal/search/**`
- `db/queries/**` (discover and search files)
- `src/api/internal/platform/dbgen/**` (generated only)
- `db/seeds/**`
- `db/migrations/**` (only if a migration is needed; R6)
- `mk/50-ops.mk` (the `seed` target only, with owner review)
- `scripts/public-routes.txt` (the 4 public reads, with reasons)

**Negative authz** (`AUTHORIZATION_MATRIX.md:24,36`):

- private, unpublished or removed content never leaks through lists or search;
- an anonymous error report returns 401;
- a CSRF failure returns 403;
- the reporter's identity is never exposed;
- the rate limit applies.

**DoD:** DoD-common, plus `make seed` evidence on `local-dev`, EXPLAIN plans in the PR, and the seed provenance test.

**Stop condition:** stop if a query cannot be bounded within ADR-0005.

**Status:** PLANNED

---

## S3 — Solve

- **Owner:** `NICE-DEV226`
- **Reviewer:** `mibienpanjoe` (proposed)
- **Branches:** `feature/solve/questions-answers`, `feature/solve/votes-accept`
- **Depends on:** S1, and Idempotency-Key from S0.

**Operations (10)**

- [ ] `listQuestions` `GET /questions` (published only; newest first with a UUID tie-breaker)
- [ ] `createQuestion` `POST /questions`
  - verified email required;
  - Idempotency-Key;
  - at most 5 tags;
  - `links[]`;
  - snippets are never executed.
- [ ] `getQuestion` `GET /questions/{questionId}` (`ETag`; 404 if not published)
- [ ] `updateQuestion` `PATCH /questions/{questionId}`
  - owner only;
  - a non-owner gets 403;
  - `If-Match` is required (428 if missing, 412 if stale).
- [ ] `listAnswers` `GET /questions/{questionId}/answers` (accepted answer first, then by score)
- [ ] `createAnswer` `POST /questions/{questionId}/answers` (verified email; Idempotency-Key)
- [ ] `castVote` `POST /votes`
  - one vote per target;
  - voting on your own content returns 403 `SELF_VOTE_FORBIDDEN`;
  - the score is computed on the server.
- [ ] `retractVote` `DELETE /votes` (idempotent)
- [ ] `acceptAnswer` `POST /questions/{questionId}/accept` (question author only; atomically replaces the previous accepted answer)
- [ ] `unacceptAnswer` `POST /questions/{questionId}/unaccept` (idempotent)

**Tables:** `questions`, `answers`, `votes`, `question_tags`, `question_sources`, `answer_sources`, `content_relations`, `tags`, `idempotency_keys` (`00005`, `00009`).

**Acceptance:** `US-SOLVE-001/1`, `US-SOLVE-001/2`, `US-IDENTITY-001/1`.

**Prototype pages:** `solve` (list and ask form), `question`, and part of `accueil`.

**Allowed paths:** `src/api/internal/solve/**`, `db/queries/**` (solve files), `src/api/internal/platform/dbgen/**` (generated only), and `scripts/public-routes.txt` (the 3 public reads).

**Negative authz** (`AUTHORIZATION_MATRIX.md:29-35,49`):

- every write returns 401 for anonymous callers, 403 `EMAIL_NOT_VERIFIED` for unverified ones, and 403 on a CSRF failure;
- a non-owner `PATCH` returns 403;
- when Bob tries to accept, he gets 403 and nothing changes;
- a self-vote returns 403;
- a duplicate vote creates no second row;
- a closed or removed question rejects new answers;
- a link to a private or unpublished target is rejected, and so is a link created by someone who is not the author.

**DoD:** DoD-common, plus a vote-score reconciliation test and a concurrent-accept test on real PostgreSQL.

**Stop condition:** stop on any change to `api/openapi.yaml` (R7).

**Status:** PLANNED

---

## S4 — Learn

- **Owner:** `mibienpanjoe`
- **Reviewer:** `NICE-DEV226` (proposed)
- **Branch:** `feature/learn/articles-series` (C8: `learn` is not in `docs/GIT_WORKFLOW.md:38`; owner ruling)
- **Depends on:** S0 for the reads; S1 for publishing. **Resolve C2 first**, and C3 too.

**Operations (5)**

- [ ] `listLearnArticles` `GET /learn/articles` (public; the `topic` filter is blocked on C2)
- [ ] `createLearnArticle` `POST /learn/articles`
  - no session returns 401;
  - publishing needs a verified email (403); saving a `draft` does not;
  - Idempotency-Key;
  - `links[]`;
  - keeps the original language.
- [ ] `getLearnArticle` `GET /learn/articles/{articleId}` (original-language body plus `lang`; resolved links)
- [ ] `listTrainingSeries` `GET /learn/series` (public)
- [ ] `getTrainingSeries` `GET /learn/series/{seriesId}` (public; `my_progress` is blocked on C3)

**Also**

- [ ] Seed one public, free AI training series with lessons, marked `provenance='demo'`.
- [ ] `links[]` writer (Shared pieces, proposal).

**Tables:** `learn_articles`, `learn_article_tags`, `training_series`, `training_lessons`, `learn_topics`, `tags`, `content_relations` (`00008`). `learn_progress` is P1.

**Acceptance:** `US-LEARN-001/1`, `/2`, `/3`, `/4`.

**Prototype pages:** `apprendre`, `article`, `formation`, `ecrire`.

**Allowed paths:** `src/api/internal/learn/**`, `db/queries/**` (learn files), `src/api/internal/platform/dbgen/**` (generated only), `db/seeds/**`, and `scripts/public-routes.txt` (the 4 public reads).

**Negative authz** (`AUTHORIZATION_MATRIX.md:24,48`):

- an anonymous publish returns 401 and creates nothing;
- an unverified publish returns 403;
- a CSRF failure returns 403;
- a draft is never visible to anyone but its author;
- removed content never leaks.

**DoD:** DoD-common, plus a test that original-language content is preserved.

**Stop condition:** stop until C2 is ruled.

**Status:** PLANNED

---

## S5 — Build

- **Owner:** `mibienpanjoe`
- **Reviewers:** `NICE-DEV226` (proposed); `Y4NN777` co-reviews the GitHub fetch and anti-SSRF code (K10)
- **Branches:** `feature/build/projects-needs`, `feature/build/applications`
- **Depends on:** S1. **Resolve C1 (maintainer rights) first.**

**Operations (9)**

- [ ] `listProjects` `GET /projects` (public projects plus the caller's own private ones)
- [ ] `createProject` `POST /projects`
  - verified email required;
  - Idempotency-Key;
  - GitHub URL allowlist, normalized and unique;
  - fetch with timeouts and anti-SSRF;
  - 10 imports per user per hour.
- [ ] `getProject` `GET /projects/{projectId}` (private projects return 404 to everyone except the owner)
- [ ] `updateProject` `PATCH /projects/{projectId}`
  - owner only (C1);
  - a non-owner gets 403 on a public project and 404 on a private one;
  - `github_url` cannot change.
- [ ] `listContributionNeeds` `GET /projects/{projectId}/contribution-needs` (follows the project's visibility)
- [ ] `createContributionNeed` `POST /projects/{projectId}/contribution-needs` (owner, C1; Idempotency-Key)
- [ ] `createApplication` `POST /contribution-needs/{needId}/applications`
  - verified email; not allowed on your own project;
  - a duplicate returns 409 `DUPLICATE_APPLICATION`;
  - a closed need returns 409 `STATE_CONFLICT`;
  - re-applying is allowed only after a withdrawal;
  - an Idempotency-Key replay returns the original 201.
- [ ] `listProjectApplications` `GET /projects/{projectId}/applications` (owner, C1; Charlie gets 403 or 404 and no count)
- [ ] `decideApplication` `PATCH /applications/{applicationId}`
  - the applicant can withdraw;
  - the owner or a maintainer can accept or decline;
  - an invalid transition returns 409;
  - a cross-project IDOR returns 404.

**Tables:** `projects`, `project_maintainers`, `project_tags`, `contribution_needs`, `need_skills`, `applications`, `content_relations`, `tags`, `idempotency_keys`, `rate_limit_buckets` (`00007`).

**Acceptance:** `US-BUILD-001/1`, `US-BUILD-001/2`.

**Prototype pages:** `build`, `project`.

**Allowed paths:** `src/api/internal/build/**`, `db/queries/**` (build files), `src/api/internal/platform/dbgen/**` (generated only), and `scripts/public-routes.txt` (the 3 public reads).

**Negative authz** (`AUTHORIZATION_MATRIX.md:39-46`):

- invalid GitHub URL, SSRF targets and the rate limit are rejected;
- a non-owner cannot mutate a project, and a private project returns 404;
- only the owner (or a maintainer, per C1) publishes a need;
- duplicate applications, closed needs, and re-applying after a decline are rejected;
- a third party cannot read applications, and nothing leaks a count;
- cross-project IDOR is blocked.

**DoD:** DoD-common, plus an SSRF test suite (private IPs, redirects, non-GitHub hosts) and a double-click application test.

**Stop condition:** stop until C1 is ruled.

**Status:** PLANNED

---

## S6 — Moderation (built last)

- **Owner:** `Y4NN777`
- **Reviewer:** `mibienpanjoe` (proposed)
- **Branch:** `feature/community/moderation` (domain choice is a GUESS; `community` is in `docs/GIT_WORKFLOW.md:38`)
- **Depends on:** S0 and S1. Its targets come from S2 to S5 (S10 if built).

**Operations (3)**

- [ ] `hideContent` `POST /moderation/{targetType}/{targetId}/hide`
  - `targetType` is one of `question`, `answer`, `resource`, `project`, `article`, `relation`;
  - content gets `status='removed'` and `removed_at`; a relation gets `status='hidden'` (`DATABASE_DESIGN.md:33,37`);
  - an `audit_log` row is written in the same transaction;
  - hiding content that is already hidden returns 409.
- [ ] `restoreContent` `POST /moderation/{targetType}/{targetId}/restore` (same rules; content that is not hidden returns 409)
- [ ] `resolveErrorReport` `POST /moderation/error-reports/{reportId}/resolve` (audited; the reporter's identity is never returned)

**Also**

- [ ] `audit_log` writer (Shared pieces).
- [ ] **Documented one-time moderator promotion** of the owner account: set `users.role='moderator'` through an audited operational step (ADR-0003). **No moderator is seeded.**
- [ ] **Emergency SQL runbook.** It sets `status='removed'` and `removed_at` (or `status='hidden'` for a relation) and writes an `audit_log` row. The brief said "status hidden", but for content that value exists only on `content_relations` and `tags`.
- [ ] Moderation UI: no prototype page exists for it. The owner decides between a demo through the API and small controls on content pages built by `PANK4SS`.

**Tables:** `audit_log`, `users.role`, `error_reports`, and every target table (`00004`, `00005` to `00009`).

**Acceptance:** ROADMAP section 3, item 9.

**Prototype pages:** none.

**Allowed paths:** `src/api/internal/moderation/**`, `db/queries/**` (moderation files), `src/api/internal/platform/dbgen/**` (generated only), and a single runbook file, `docs/runbooks/moderation.md` (path is a GUESS; the owner confirms it).

**Negative authz** (`AUTHORIZATION_MATRIX.md:50-51`):

- a non-moderator gets 403 and an anonymous caller 401;
- a CSRF failure returns 403;
- the audit row is written in the same transaction, and a rollback leaves no orphan;
- the moderator never edits content.

**DoD:** DoD-common, plus a test that the audit row is written in the same transaction.

**Stop condition:** stop if moderation scope would grow beyond ADR-0003.

**Status:** PLANNED

---

## S7 — Frontend (prototypes become React pages)

- **Owner:** `PANK4SS`
- **Reviewer:** `Y4NN777`
- **Branches:** `feature/<page domain>/web-<page>`, for example `feature/discover/web-resource`
- **Depends on:** typed fixtures from `src/web/src/api/schema.d.ts` until each endpoint merges, then switch to the live endpoint.

**Order:** shell, i18n and shared components; then the public reads; then the identity pages; then the write flows.

- [ ] **Pillar emphasis** (owner, 2026-11-12; O7):
  - navigation, home (`accueil`) and FR/EN copy make **Learn & Grow** and **Connect** prominent;
  - Solve, Discover and Build stay full pillars in the same navigation, with no pillar erased or demoted (ROADMAP O7 and section 12.4);
  - an information-architecture and copy change only, with no new endpoints;
  - nothing non-functional is announced: Connect shows only the E5 directory as live.
- [ ] Shell:
  - `getSession` with header sign-in and sign-out;
  - FR/EN labels without altering content;
  - a demo badge for `provenance='demo'`;
  - loading, empty and error states.

| Page (`prototypes/html/pages/`) | Operations | Slice | Done |
|---|---|---|---|
| `accueil` | `listQuestions`, `listResources`, `listProjects`, `listLearnArticles` | S2 to S5 | [ ] |
| `discover` | `listResources`; demo data comes from S12 aggregation (providers, docs links, countries, categories), labelled "via <source>" | S2, S12 | [ ] |
| `resource` | `getResource`, `listResourceSources`, `createResourceErrorReport`; source attribution shown | S2, S12 | [ ] |
| `recherche` | `search` (plus the S9 panel) | S2, S9 | [ ] |
| `apprendre` | `listLearnArticles`, `listTrainingSeries`; demo data = S12 aggregated articles and curated paths ("via <source>"); E4 topic feeds (`listLearnArticles?topic=`, `listQuestions?tag=`; blocked on C2) | S4, S12, E4 | [ ] |
| `article` | `getLearnArticle` ("via <source>" for aggregated items, "by member" otherwise); E4 "Poser une question sur cette leçon" opens `solve`, which calls `createQuestion` with `links[article]` | S4, S12, E4 | [ ] |
| `formation` | `getTrainingSeries`: a curated learning path of attributed external lessons, aggregated by S12; E4 ask button per lesson | S4, S12, E4 | [ ] |
| `solve` | `listQuestions`, `createQuestion` (ask form, the S9 panel, and the E4 pre-fill from a lesson) | S3, S9, E4 | [ ] |
| `question` | `getQuestion`, `listAnswers`, `createAnswer`, `castVote`, `retractVote`, `acceptAnswer`, `unacceptAnswer`, `updateQuestion` | S3 | [ ] |
| `connexion` | `getCsrfToken`, `signIn` | S1 | [ ] |
| `inscription` | `signUp` | S1 | [ ] |
| `verifier-email` | `verifyEmail` | S1 | [ ] |
| `mot-de-passe-oublie` | `requestPasswordReset` | S1 | [ ] |
| `reinitialiser` | `confirmPasswordReset` | S1 | [ ] |
| `bienvenue` | `updateMyProfile` (onboarding; GUESS from the prototype) | S1 | [ ] |
| `profile` | `getProfile` | S1 | [ ] |
| `parametres` | `getMe`, `updateMyProfile` | S1 | [ ] |
| `ecrire` | `createLearnArticle` | S4 | [ ] |
| `build` | `listProjects`, `createProject`; demo data = S12 aggregated public GitHub projects ("via GitHub"). Whether aggregated projects accept needs or applications depends on C15. | S5, S12 | [ ] |
| `project` | `getProject`, `updateProject`, `listContributionNeeds`, `createContributionNeed`, `createApplication`, `listProjectApplications`, `decideApplication` | S5 | [ ] |
| `connect` | E5 people list (the new operation, suggested `listProfiles` `GET /profiles`; name fixed in the contract PR) for the `connect-discovery` section with its country, language and skill filters. Its "later" badge (`connect.html:42`) and the `data-proto` preview note (`:27`) are removed only once the endpoint is live. The `connect-spaces` block stays "bientôt" (`:66-72`), with no fake spaces. Mentoring and opportunities are shown as coming soon (PRD 5.5). **Demo data:** the people list can show only real registered members, because S12 ingests no personal data and creates no users. The brief says Connect relies on aggregated data, so the owner must rule (C18) on what aggregated data, if any, appears here (for example linked aggregated projects). | E5, S12 | [ ] |
| `auth` | legacy redirect to `connexion` | - | n/a |
| `_template` | template, not a page | - | n/a |

**Acceptance:** `US-IDENTITY-001/2`, `US-LEARN-001/4`, `US-ACCESS-001/1`, plus each slice's journeys through the UI.

**Allowed paths:** `src/web/**` and `DESIGN.md` (ADR-0008). `pnpm-lock.yaml` takes one change per PR.

**Negative cases:**

- the UI never grants authority; role flags only hide controls;
- no token is stored in `localStorage`;
- every unsafe call sends `X-CSRF-Token`.

**DoD:** DoD-common (web above 70% on the four metrics), plus Vitest and Testing Library by role and label. ADR-0007 visual, axe and responsive evidence is **NOT RUN** until it is produced.

**Stop condition:** stop if a page needs a field that is not in `schema.d.ts` (R7: contract PR first).

**Status:** PLANNED

---

## S8 — Ops, release and CADEV

- **Owner:** `Y4NN777`
- **Reviewer:** `NICE-DEV226` or `mibienpanjoe`
- **Branches:** `chore/ops/staging`, `chore/ops/release-workflow`, `chore/ops/e2e-harness`, `chore/ops/perf-harness`
- **Depends on:** H6 to H8

**Tasks**

- [ ] H1, H3, H4: identities confirmed, ADR-0008 decided, branch protection with the `quality` check required on `develop` and `main`.
- [ ] Datacloud account and active product (H6). DNS `226os.tech` (H7). TLS through Traefik. The DB is never published.
- [ ] `staging-test` up from pinned images (`make up-staging-test`, `.env.staging-test` outside Git). `make seed` on staging if O1 chooses staging.
- [ ] Gmail SMTP on staging with a real-inbox test (H8).
- [ ] Argon2id benchmark on the target host, handed to S1.
- [ ] B2: `/readyz` reachable for the smoke check (route it, or check from the host).
- [ ] **B4:** a reviewed release workflow (protected `production` environment, exact `develop` SHA, image digests). The guard in `mk/50-ops.mk:28-32` is not weakened.
- [ ] Implement `make test-e2e` (`mk/40-quality.mk:18`): Playwright P0 flows, FR/EN, mobile, axe.
- [ ] Implement `make perf` (`mk/40-quality.mk:21`): dataset, EXPLAIN, API latency, page weight.
- [ ] Backup and restore drill (`pg_dump -Fc`, restore verified), with evidence.
- [ ] Rollback drill (redeploy the previous image by digest), with evidence.
- [ ] G7 evidence on the target SHA: build SHA, digests (`make build`, which writes `reports/image-digests.txt`), audit, backup, smoke, rollback.
- [ ] Production promotion PR from `develop` to `main`, with owner approval.
- [ ] CADEV checklist (ROADMAP section 9), including H9 and H10, and the check that names in PRD section 12 match the declared team (flag only).

**Allowed paths:** `Makefile`, `mk/**`, `scripts/**`, `compose.yaml`, `deploy/**`, `.github/**`, `docs/quality/**`, and `docs/evidence/**` (path GUESS).

**DoD:** each item has evidence in the PR or under the evidence path for the target SHA. `make quality-full` on the `develop` SHA reports G5 and G6 `PASS`, or `NOT RUN` with a dated owner exception.

**Stop condition:** a production deploy without G7 evidence and owner approval is forbidden.

**Status:** PLANNED

---

## S9 — Grounded AI assistant (P0 by owner decision, amends PRD 5.8)

- **Owner:** `Y4NN777` (contract, ADR, port)
- **Second developer:** **OPEN** (`NICE-DEV226` or `mibienpanjoe`, once their first slice merges)
- **UI:** `PANK4SS`
- **Reviewers:** a backend developer for the owner's PRs; `Y4NN777` and `PANK4SS` for the contract
- **Branches:** `feature/intelligence/assistant-contract`, then `feature/intelligence/assistant`
- **Depends on:** S1 (subject, rate limiter), S2 (search), and H12
- **Provider** (owner decision 2026-11-12; supersedes the OpenAI decision): Google AI Studio, model Gemma 4, free tier.
  - The key lives only in the staging and production env files, never in Git, CI logs or images.
  - The LLM port keeps the provider swappable.
  - Development and the golden evaluation set run against the fake adapter.

**Facts from AI Studio terms (OFFICIAL)** — `https://aistudio.google.com/docs/terms`, "Effective March 23, 2026", re-fetched 2026-11-12. Details are in ROADMAP section 10 under O5 and in K21.

- **Free tier (Unpaid Services).**
  - Content is used to improve Google products.
  - Human reviewers may read API input and output.
  - The terms say not to submit personal information.
- **EEA, Switzerland and UK.** Only Paid Services may be offered to users there.

**Tasks**

- [ ] `Y4NN777`: create the AI Studio API key and store it in the staging env file.
- [ ] `Y4NN777`: verify V1 to V4 in AI Studio and Google's docs. **Due Fri 13 Nov evening.**
  - V1: the exact Gemma 4 model id and its availability. Ids reported by third-party aggregators are NOT CONFIRMED.
  - V2: the project's real limits. They are UNKNOWN, and third-party figures conflict.
  - V3: a free project or a billing-linked one, given the EEA rule.
  - V4: FR/EN latency and quality on the golden set.
- [ ] PROPOSED ADR for the AI integration (next free ADR number). It must decide:
  - the data sent: the question text and retrieved public content only; no emails, private data or secrets;
  - the cost cap or quota guard, and where its counter lives (a new table needs a migration from `mibienpanjoe`; C11);
  - timeouts (B3);
  - what happens when the provider is down or the quota is exhausted;
  - FR/EN behaviour;
  - the evaluation set and its pass criteria;
  - whether anonymous users have access (O5).
- [ ] PRD amendment 3.2 proposal text inside the ADR; the owner applies it to `docs/PRD.md` (H13).
- [ ] SPEC story with BDD, and new `AUTHORIZATION_MATRIX.md` rows, proposed in the ADR; the owner applies them.
- [ ] Contract PR: new operation(s) in `api/openapi.yaml` (names TBD in the ADR) with `x-priority`, `x-authz`, 429 and 503. Reviewed by `Y4NN777` and `PANK4SS`; it merges before any code (R7).
  - Regenerate `src/api/internal/platform/apigen/**` and `src/web/src/api/schema.d.ts` with `make generate`.
  - Update the operation count (47) in `src/api/internal/platform/apigen/contract_test.go`.
  - Amend the exclusion justification in `docs/quality/COVERAGE_SCOPE.md:17`, which cites 47 operations. That file is change-controlled, so the change needs owner, reviewer and date.
  - Each new operation is a newly mounted route, so B1 applies to it.
- [ ] LLM port with a fake adapter. Tests make no network calls, and Go coverage stays at 100.00%. The real adapter is selected by configuration.
- [ ] Retrieval through the existing PostgreSQL search (S2), over published and public content only, with the top results as context. No vector database.
- [ ] Answers carry links to the sources used. The assistant refuses when there is no source, and never invents guarantees about an API or about whether a solution is verified (PRD 5.8).
- [ ] Retrieved content is treated as untrusted (prompt injection).
- [ ] A redaction step removes obvious secrets and emails before sending, with tests.
- [ ] Per-subject rate limit, a quota or cost guard, and a clear "assistant unavailable" response when the provider is down or the quota is exhausted. The page never fails.
- [ ] No prompt logging that contains personal data or secrets.
- [ ] A per-environment enable switch in config. The UI shows a visible notice that the question is sent to Google.
- [ ] **Golden evaluation set**: questions with their expected cited sources, built on demo seeds (fixed UUIDs). It runs against the fake adapter in CI. The real-provider run is recorded separately (V4) and is NOT RUN until done.
- [ ] `PANK4SS`: the assistant panel in `recherche`, and in the `solve` ask flow (similar questions, help phrasing the question).
- [ ] `THIRD_PARTY_NOTICES.md` and `AI_USAGE.md` entries for the provider (owner).

**E1 teaching rules** (owner-accepted 2026-11-12). Each rule is an acceptance test and appears in the golden set.

- [ ] **AI-1 Ground or refuse.** If zero published passages are retrieved, the provider is **not called**. The answer says there is no source and links to search.
- [ ] **AI-2 Server-validated citations.** A citation outside the retrieved set is rejected or stripped. Every answer carries at least one citation, or it is a refusal.
- [ ] **AI-3** No guarantees about an API or about a verified solution (PRD 5.8).
- [ ] **AI-4** Every answer ends with a "Comment vérifier" section.
- [ ] **AI-5** When sources conflict or are undated, the answer states its uncertainty explicitly and gives the source dates.
- [ ] **AI-9** Answers are labelled as AI and can never be accepted or voted on as an answer (no `answers` row, no vote target).
- [ ] **AI-10** Only the question text and public retrieved content are sent to the provider, after redaction.
- AI-6 (hint-first) and AI-7 (ask for your attempt) are **P1, out of scope**.

**Other acceptance** (to be fixed in the SPEC story):

- an injected instruction in content is ignored;
- when the provider is down, the response is "assistant unavailable";
- the N+1 request gets 429;
- the golden set passes.
- Failure modes: ROADMAP section 12.2.

**Tables:** none planned. A cost or quota counter may need one (C11).

**Allowed paths:**

- `src/api/internal/assistant/**` (name GUESS)
- `api/openapi.yaml` (contract PR only), `src/api/internal/platform/apigen/**` and `src/web/src/api/schema.d.ts` (generated only), and `docs/quality/COVERAGE_SCOPE.md` (the count line only, with owner approval)
- `docs/decisions/**` (ADR, owner)
- the S9 UI in `src/web/**` (`PANK4SS`)

**Negative authz:** depends on O5. In all cases: CSRF on unsafe methods, 429 per subject, and no private or removed content in context or citations.

**DoD:** DoD-common, plus the golden set green on the fake adapter, plus the recorded real-provider run with V4 numbers before the demo.

**Stop condition:** stop if V3 (the EEA rule) is not decided before the staging demo. Without that decision, the assistant stays disabled by config.

**Status:** PLANNED

---

## E2 — Taxonomy: add and relabel only

- **Owner:** `mibienpanjoe` (the only migration writer, R6)
- **Reviewer:** `Y4NN777`, because pgtest is in the platform area
- **Branch:** `feature/platform/taxonomy-ai-topics`
- **Depends on:** nothing. One migration at a time (R6).

**Tasks**

- [ ] Create a new migration with `make migrate-new NAME=<slug>`. Use the number it gives; **not** 00011, which is reserved (C11). It is forward-only, with the standard header.
- [ ] Add 5 learn topics:
  - `learn-to-learn` (Apprendre et rester à jour / Learning & keeping up)
  - `ai-assisted-dev` (Coder avec l’IA sans perdre la main / Coding with AI, keeping your skills)
  - `ai-agents` (Construire et évaluer des agents / Building & evaluating agents)
  - `open-models` (Modèles ouverts et IA en local / Open-weight & local models)
  - `african-languages` (IA et langues africaines / AI & African languages)
- [ ] Relabel only; codes and primary keys stay the same:
  - learn topic `ai`: Comprendre les LLM / Understanding LLMs
  - learn topic `mobile`: Mobile, USSD et faible débit / Mobile, USSD & low bandwidth
  - learn topic `open-source`: Contribuer à l’open source / Contributing to open source
  - resource category `cloud-hosting`: Hébergement, cloud et GPU en Afrique / Hosting, cloud & GPUs in Africa
- [ ] Add 3 resource categories (each needs a `sort_order`; the existing step is 10):
  - `ai-models` (Modèles et API d’IA / AI models & APIs)
  - `ai-tooling` (Agents et outillage IA / AI agents & tooling)
  - `ai-datasets` (Jeux de données et évaluation / Datasets & evaluation)
- [ ] Result: **14 learn topics and 18 resource categories**. `payments`, `ai`, `mobile-money` and `sms-ussd` stay valid, because fixtures use them (`pgtest/fixtures_integration_test.go:65-66`).
- [ ] In the same PR, update the exact lists and counts in `src/api/internal/platform/pgtest/migrate_integration_test.go` (`:140-150`, `:180` want 15, `:219` want 9).
- [ ] K25: check whether seeds or fixtures tag `ai` content that really belongs under `african-languages`.
- [ ] Docs follow-up, outside this PR: the `DATABASE_DESIGN.md` lines about the lists.

Removals are **DEFERRED** until after CADEV and need explicit human approval: `ussd-messaging`, `geo-data`, the `logistics` merge.

Taxonomy rows are data. Later they evolve through S12 ingestion classification or a moderator path, never by editing old migrations (ADR-0009 point 9, PROPOSED).

**Allowed paths:** `db/migrations/<new file>` and `src/api/internal/platform/pgtest/migrate_integration_test.go`.

**DoD:**

- DoD-common.
- `make test-integration` passes on fresh and upgrade runs.
- The migration immutability and no-`Down` checks pass (`make lint`).

**Status:** PLANNED

---

## S12 — Content ingestion and curated learning paths (replaces S11)

- **Builder:** **OPEN** (O10; a backend developer).
- **Design:** `Y4NN777` owns the allowlist and ADR-0009.
- **Migration:** `mibienpanjoe` (R6).
- **Reviewers:** `Y4NN777` (sources, licences, legal), plus the other backend developer (code).
- **Branches:** `feature/platform/provenance-aggregated` (migration), `feature/platform/ingest-cli`, `chore/ops/make-ingest`.
- **Depends on:**
  - ADR-0009 accepted (H15);
  - the C15 ruling (owner foreign key for aggregated rows) and the C16 ruling (review step);
  - H17 (sources chosen);
  - E2 (the taxonomy codes used by classification);
  - the S2, S4 and S5 services.

**Tasks**

- [ ] `Y4NN777`: write `docs/decisions/0009-aggregated-content-and-ingestion.md`; it exists as PROPOSED. Accept, amend or reject it (H15). Acceptance resolves B5 and K24.
- [ ] `Y4NN777`: choose the initial sources (H17) and commit the allowlist. The path and format are OPEN (O11; proposal: `db/ingest/sources.yaml`, GUESS).
- [ ] **Each source's licence, robots.txt and terms: to verify per source.** One checkbox per source, added when the allowlist lands. None is assumed.
- [ ] One `THIRD_PARTY_NOTICES.md` entry per source (CADEV audit). The owner edits it.
- [ ] **(g) Migration.** Next number from `make migrate-new`, not 00011. Written by `mibienpanjoe`. The rows below were verified against the schema.

  | Table | Provenance CHECK today | Change | Source metadata today |
  |---|---|---|---|
  | `tags` | `live`, `demo` (`00002`) | add `aggregated` | none |
  | `providers` | `live`, `demo` (`00006`) | add `aggregated` | `website_url` only |
  | `resources` | `live`, `demo` (`00006`) | add `aggregated` | `resource_sources` (`url`, `title`, `kind`, `checked_at`), with no licence, content hash or source name |
  | `projects` | `live`, `demo` (`00007`) | add `aggregated` | `github_url`, `license_spdx`, with no source name, `fetched_at` or hash |
  | `learn_articles` | `live`, `demo` (`00008`) | add `aggregated` | none |
  | `training_series` | `live`, `demo` (`00008`) | add `aggregated` | none |
  | `users`, `questions`, `answers` | `live`, `demo` | **no change** (member-only) | - |

  - Add the missing source metadata: source name, source URL, licence, `fetched_at`, content hash. Per-table columns or one shared table is an ADR-0009 choice.
  - Add a CHECK that `provenance='aggregated'` implies a source URL and `fetched_at`.
  - In the same PR, extend the pgtest provenance tests: `aggregated` is accepted on the six tables and rejected on `users`, `questions` and `answers`. The existing `users_provenance_ck` case (`constraints_integration_test.go:46`) stays.
  - If a new table is added, also update the table list in `migrate_integration_test.go:19-22`.
- [ ] **(a) The CLI `src/api/cmd/ingest`.** It writes **through the service layer, never raw SQL**, so invariants hold.
  - No service writes `providers`, `resources`, `training_series` or `training_lessons` today. The matrix leaves catalogue loading "outside the API" (`AUTHORIZATION_MATRIX.md:37`). Those write services must be added inside their domain packages, coordinated with the S2 and S4 owners (C17).
  - Connectors, in priority order:
    - [ ] GitHub REST API for open-source projects. Follows PRD 5.3 "import par URL GitHub" and SPEC section 7: quotas, timeouts, anti-SSRF.
    - [ ] RSS/Atom for articles.
    - [ ] Allowlisted pages or data files for API and tool directories: providers, docs links, countries, categories.
- [ ] **(b) Processing pipeline.**
  - Normalise.
  - Dedupe by canonical URL plus content hash.
  - Classify to a topic or category with rules first. LLM-assisted classification comes later, behind the S9 port, on public content only.
  - Write with `provenance='aggregated'`, the source name, source URL, licence and `fetched_at`.
  - Review per C16: proposed as review of the allowlist and snapshot in a PR.
  - Publish.
- [ ] **Curated learning paths.** A `training_series` (`origin='platform'`, `00008_learn.sql:75`) whose `training_lessons.article_id` points to aggregated `learn_articles`: attributed external lessons, a short excerpt and the canonical link. A short intro and a verify step are optional glue; authorship is not required. Example: "Coder avec l’IA sans perdre la main".
- [ ] **(c) Pinned snapshot.** A deterministic snapshot for the demo that records the source list and content hashes (path GUESS: `db/ingest/snapshots/`).
- [ ] **(d) The `make ingest` target** in `mk/50-ops.mk` (owner area; owner review; ADR-0006 table).
  - It fails closed.
  - Outside `local-dev` it needs explicit confirmation (`INGEST_CONFIRM=<profile>`) and backup evidence, like `migrate-up`.
  - It is never silent in production.
  - `make seed` stays fake demo fixtures only, `local-dev` and `staging-test` only, never in production.
- [ ] **(e) Scraping rules**, which are acceptance criteria, each with a test:
  - [ ] allowlist only;
  - [ ] each source's licence, robots.txt and terms respected;
  - [ ] store facts, a short excerpt and the canonical link with attribution; full text only where the licence allows;
  - [ ] an identifiable user-agent;
  - [ ] rate limits and timeouts per source;
  - [ ] a takedown path (moderator hide plus allowlist removal; a re-run never resurrects the item);
  - [ ] no personal data ingested;
  - [ ] HTML sanitised and treated as untrusted, including when the assistant later uses it.
- [ ] **(f) UI** (`PANK4SS`, S7): "via <source>" with a link for aggregated content, "by member" for member content. Solve (questions and answers) and identity stay member-only. No fake users or votes anywhere.

**Failure modes:** ROADMAP section 12.2 (S12 rows).

**Allowed paths:**

- `src/api/cmd/ingest/**`
- the write services in the domain packages under `src/api/internal/**` (coordinated with their slice owners)
- `db/queries/**` (ingest-related)
- `src/api/internal/platform/dbgen/**` (generated)
- `db/migrations/<new>` (`mibienpanjoe` only)
- `src/api/internal/platform/pgtest/**` (the provenance tests)
- the allowlist and snapshot paths (O11)
- `mk/50-ops.mk` (owner review)

**DoD:**

- DoD-common.
- Connector tests run against recorded fixtures, with no network in `make quality`.
- `make ingest` on `local-dev`, then on `staging-test`, with the snapshot hash recorded.
- Per-source licence checkboxes are ticked, with evidence.

**Status:** PLANNED (builder OPEN; blocked on ADR-0009, C15 and C16)

---

## E4 — Zero-migration interaction modes

- **Owner:** `PANK4SS` with `NICE-DEV226` (the Solve owner)
- **Reviewer:** `Y4NN777`
- **Branches:** `feature/solve/web-ask-about-lesson`, `feature/learn/web-topic-feeds` (C8)
- **Depends on:** S3 (`createQuestion`) and S12 or S4 articles. **The topic feeds wait for the C2 ruling**: questions have no topic, and `listQuestions` filters only by `tag` (`api/openapi.yaml:462`).

**No new table, no new moderation system, no contract change.**

**Tasks**

- [ ] The "Poser une question sur cette leçon" button on `article` and `formation`.
  - It pre-fills the `solve` ask form.
  - On submit, `createQuestion` sends `links: [{type: article, id}]`.
  - That relation pair is allowed (`db/migrations/00009_content_relations.sql:52`), and the question author is the author of the source item.
- [ ] Read-only topic feeds: filtered views of questions and articles per topic, using `listLearnArticles?topic=` and `listQuestions?tag=`. **Blocked on C2.**
- [ ] Tests: the pre-fill, the link created, and a question with an anonymous session redirects to sign-in. Web coverage stays above 70%.

**Allowed paths:** `src/web/**`. Backend changes only in `src/api/internal/solve/**`, and only if the link handling needs a fix.

**DoD:** DoD-common, plus an e2e step in `make test-e2e` once S8 delivers the harness.

**Status:** PLANNED

---

## E5 — Connect: people directory

- **Backend owner:** **OPEN** (O9). Proposed: `mibienpanjoe` or `NICE-DEV226`, after their first slice merges.
- **UI:** `PANK4SS`
- **Reviewers:** the backend lead (`Y4NN777`) or the other backend developer. The contract PR needs `Y4NN777` and `PANK4SS` (R7).
- **Branches:** `feature/connect/profiles-contract`, then `feature/connect/people-directory` (`connect` is in `docs/GIT_WORKFLOW.md:38`)
- **Depends on:** S1 (profiles, session resolver) and **C14** (an owner ruling on what makes a profile listable).

**Tasks**

- [ ] **Contract PR first:** a new operation in `api/openapi.yaml`, suggested `GET /profiles`.
  - Filters: `country` (CountryCode), `language` and `skill` (TagSlug).
  - Standard `cursor` and `limit` (default 20, maximum 50, above 50 returns 422).
  - `x-priority: P0`, and an `x-authz` allowing `anonymous|member`.
  - A response of public summary fields only.
  - Run `make generate` to regenerate `src/api/internal/platform/apigen/**` and `src/web/src/api/schema.d.ts`.
  - Raise the operation count by one in `src/api/internal/platform/apigen/contract_test.go`, and in `docs/quality/COVERAGE_SCOPE.md:17` with owner approval. That is 47 to 48 if this contract PR merges before the S9 one; otherwise add one to whatever the count is then.
  - **P0 then has 44 operations.**
- [ ] A new row in `AUTHORIZATION_MATRIX.md`, proposed in the contract PR. The owner applies it, because that file is outside this tracker's scope.
- [ ] Add `GET /api/v1/profiles` to `scripts/public-routes.txt`, with its reason (public by design). B1 applies.
- [ ] Backend: a projection filtered by `field_visibility`, depending on the caller (anonymous or member).
  - Email is never selected.
  - Only listable profiles appear (C14). Deleted and suspended users never appear.
  - Filters match only fields visible to the caller.
  - Stable order with a UUID tie-breaker.
  - A rate limit per IP and per user. The window is the owner's call, because ADR-0003 has none for this route.
- [ ] `PANK4SS`: the `connect` page (see the S7 page map), with loading, empty and error states, FR/EN, and a reset-filters action.
- [ ] **Office hours: a process note only.** One named mentor answers questions tagged `permanence` during a fixed weekly hour. The owner writes the note and names the mentor. It needs no code.

**Not in scope**, shown as coming soon (PRD 5.5): mentoring and matchmaking, CV review, interview preparation, study groups, opportunities.

**Negative and privacy tests** (required):

- a hidden field (`members` or `private`) never appears for a caller who may not see it;
- filtering on a hidden field never matches;
- no `email` key ever appears;
- a non-listable or private profile is absent from the list, and `getProfile` returns 404 for it (matrix row `GET /profiles/{username}`);
- `limit=51` returns 422;
- a tampered cursor returns 400;
- the N+1 request past the rate limit gets 429.

**Allowed paths:**

- `api/openapi.yaml` (contract PR)
- the generated outputs listed above
- `src/api/internal/identity/**`, or a new profiles package (the lead decides)
- `db/queries/**` (profiles files)
- `src/api/internal/platform/dbgen/**` (generated only)
- `scripts/public-routes.txt`
- `src/web/**` (`PANK4SS`)

A migration is needed only if C14 is ruled as a new column, and only `mibienpanjoe` writes it (R6).

**DoD:** DoD-common, plus the privacy tests above on real PostgreSQL, plus the failure modes in ROADMAP section 12.2.

**Status:** PLANNED (owner OPEN; blocked on C14)

---

## S10 — Live channels (CONDITIONAL)

- **Owner:** **OPEN**
- **Migrations:** `mibienpanjoe` (R6)
- **UI:** `PANK4SS`
- **Reviewers:** OPEN
- **Branch:** `feature/community/channels`
- **Depends on:** a **GO at the Sat 14 12:00 checkpoint**, which requires S0 to S5 on track. It also needs S1, S4 and S5. Without a GO, the status becomes `NOT BUILT (P2)` and this is reported honestly.

**Displaces if built** (O6, GUESS): Saturday afternoon capacity for S6 and S7 integration, the G5 and G6 harnesses, the backup and restore drill, the bug-fix buffer before the 18:00 freeze, and the E4-type work. In the slip order, S10 is the first item (ROADMAP section 12.1).

**Tasks**

- [ ] PROPOSED ADR for the real-time transport:
  - server-sent events with PostgreSQL `LISTEN/NOTIFY`, no Valkey or S3 (ADR-0006);
  - the connection budget;
  - the 30 s router timeout (B3);
  - Traefik streaming;
  - message bounds.
- [ ] Contract PR: channel operations, and extending `ModerationTargetType` (`api/openapi.yaml:2454`) to cover messages. Reviewed by `Y4NN777` and `PANK4SS`.
  - Regenerate `src/api/internal/platform/apigen/**` and `src/web/src/api/schema.d.ts`.
  - Update the operation count in `contract_test.go` and in `docs/quality/COVERAGE_SCOPE.md:17` (owner approval).
  - B1 applies to every new operation.
- [ ] Migration with the next free number (C11; `00011` is reserved), written by `mibienpanjoe`.
- [ ] A channel is attached to an existing object (a project or a Learn series). There is no free-form chat.
- [ ] Messages are persisted in PostgreSQL and delivered live. Only members may use channels. Rate-limited. Messages carry `lang` and are never translated.
- [ ] Moderation can hide messages (an S6 extension, audited).
- [ ] `PANK4SS`: a channel panel on the `project` and `formation` pages.

**Negative authz:**

- anonymous callers get 401;
- a private project channel returns 404 to non-owners;
- CSRF is checked on posting;
- rate limits apply;
- a hidden message is never delivered.

**Allowed paths:** the S10 module under `src/api/internal/**` (name GUESS); `api/openapi.yaml` (contract PR); `src/api/internal/platform/apigen/**`, `src/api/internal/platform/dbgen/**` and `src/web/src/api/schema.d.ts` (generated only); `db/migrations/**` (`mibienpanjoe` only); and the channel UI in `src/web/**`.

**DoD:** DoD-common, plus a stream reconnect test and a moderation-hide test.

**Status:** CONDITIONAL

---

## Deployment checklist (owner `Y4NN777`, target SHA recorded on each line)

- [ ] `develop` SHA frozen at Sat 18:00. `make quality-full` run on it, with its result (G1 to G6).
- [ ] `staging-test` deployed from that SHA by digest (`make build`, then `make up-staging-test`).
- [ ] Staging smoke per `docs/SHIP.md` section 4: HTTPS, `/healthz`, `/readyz`, auth, Q&A, vote, search, resource, project, application, anonymous reads, denied mutations, FR/EN, mobile, no secrets in logs.
- [ ] Backup taken and restore verified on a scratch database.
- [ ] Rollback drill: the previous image is redeployed by digest.
- [ ] G7 evidence complete for the SHA (ROADMAP section 1, go/no-go rule).
- [ ] Production PR from `develop` to `main` with owner approval. The same SHA is promoted, never rebuilt.
- [ ] Production: backup, then `migrate-up` (`MIGRATE_CONFIRM` and `BACKUP_EVIDENCE`), then deploy through the release workflow (B4), then the smoke test.
- [ ] Owner account promoted to moderator in production (S6 runbook). A check confirms there are zero demo rows in production.
- [ ] S9 enabled or disabled per environment, according to V3.
- [ ] S12: `make ingest` run on `staging-test` from the pinned snapshot, with its hash recorded. Then the same snapshot on `production-live`, with explicit confirmation and a backup first, and the same hash recorded. A check confirms zero `demo` rows in production; `aggregated` rows are allowed under ADR-0009.
- [ ] E2 migration applied (it is part of the forward migrations on the target SHA).
- [ ] CADEV submission around Sun 12:00 GMT, before 17:00 (ROADMAP section 9).
- [ ] Vote reminders sent. **Every member votes between Sun 18:00 and Mon 23:59 GMT.**

## Done log

| Date | Item | PR | By |
|---|---|---|---|
| 2026-11-12 | B0 bootstrap merged: API skeleton, migrations 00001 to 00010, codegen, web skeleton, Compose, Make, CI | #24 | `Y4NN777` |
