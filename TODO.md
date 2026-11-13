# 226os P0 tracker

> **Live status is in this file. History is in [docs/archive/tracker/](docs/archive/tracker/)** (S0 and the checked items: [TODO-archive-2026-11-12.md](docs/archive/tracker/TODO-archive-2026-11-12.md)).
>
> - This is the single live tracker. The plan, rules R1 to R11, blockers (B), contradictions (C), open decisions (O), verification items (V) and risks (K) are in [ROADMAP.md](ROADMAP.md). This file keeps IDs only, never the decision text.
> - Update this file **in the same PR as the work**, one line per change. Never rewrite someone else's checkbox or status; add a dated line under it.
> - When a slice merges, set its status to `DONE (PR #n, YYYY-MM-DD)` and add a line to the Done log. People are named by GitHub login only.
> - Contracts are linked, not copied: [api/openapi.yaml](api/openapi.yaml) (operations, `x-authz`), [docs/SPEC.md](docs/SPEC.md) (BDD), [AUTHORIZATION_MATRIX](docs/spec-kit/AUTHORIZATION_MATRIX.md), [DATABASE_DESIGN](docs/spec-kit/DATABASE_DESIGN.md).

**Legend.** Checkbox: `- [ ]` open, `- [x]` done (with PR number). Slice status: `PLANNED`, `IN PROGRESS`, `IN REVIEW`, `DONE (PR #n, date)`, `BLOCKED (reason)`; S10 only: `CONDITIONAL`, `NOT BUILT (P2)`. Gate results: `PASS`, `FAIL`, `NOT RUN`, `BLOCKED` (`NOT RUN` is never `PASS`). BDD ids such as `US-SOLVE-001/2` number the scenarios of a SPEC.md story in order.

**DoD-common (every slice).** R1 to R11 of ROADMAP section 5 are respected. `make quality` (G1 to G4) is green in CI on the PR head SHA. The PR description lists the RED then GREEN test names, Go coverage at 100.00% (web above 70% where touched), the `make security` route-table output, the story id and reviewer, and `Commands executed` with exit codes. Every new mounted route is in `scripts/public-routes.txt` with a reason, or covered by a negative authz test the gate recognises (B1; routes leave `scripts/unimplemented-routes.txt` as their slice lands). Generated code is never hand-edited; two PRs touching generated files rebase on `develop` and rerun `make generate`.

## Now

One line per person. Each person edits only their own line.

| Login | Next action | Updated |
|---|---|---|
| `Y4NN777` | S0 is merged (PRs #31 to #36, see Done log). Owner tasks: H6 and H7 (Datacloud, DNS), H12 (V2 quotas), issue #46 (collaborators' age keys). Build: S6 moderation, B4 release workflow, the S9 contract PR (R7) and the S9 backend. Review S1 and S0 follow-ups with `NICE-DEV226`. | 2026-11-12 |
| `NICE-DEV226` | Identity confirmed (H1 done). S0 is merged: start S1 (RED tests for the domain and service with fakes, then the handlers on the merged S0 base). Then S3; then help on S5 (default), the S9 second developer, and the E5 backend if time allows. | 2026-11-12 |
| `mibienpanjoe` | Identity confirmed (H1 done). First contract PR #79 and E2 PR #80 are merged. Catalogue reads are in review as PR #83. Search and error reports wait on S1. | 2026-11-13 |
| `PANK4SS` | Identity confirmed (H1 done). **Accept the repository invitation (H2) before pushing.** S7 shell, i18n, shared components and pillar navigation on typed fixtures from `src/web/src/api/schema.d.ts`; SEED fixtures (`local-dev` only, C15). Later E4, the E5 UI, on-demand support. | 2026-11-12 |

## Open CADEV checklist and owner decisions

CADEV rules checklist (added 2026-11-12 from the official rules; the rules text is not copied here). Owner and team tasks.

- [ ] **Admissibility (art. 6): a Datacloud product is actually purchased** and active at submission (H6). Owner.
- [ ] **Admissibility (art. 6): no GPL, AGPL or LGPL component.** Observed facts:
  - the API image no longer contains busybox (removed 2026-11-12: it ships its own `/healthcheck` command, `src/api/cmd/healthcheck`); busybox remains only as the one-shot `acme-init` helper image in Compose (staging and production);
  - the web image is `nginx-unprivileged:stable-alpine` with `apk upgrade` (`src/web/Dockerfile`), so it carries Alpine OS packages;
  - `make licenses` covers only Go modules and npm production dependencies (`scripts/licenses.sh`), not image OS packages.
  - [x] DECIDED 2026-11-12 (owner): runtime base images (busybox, Alpine, distroless, postgres, traefik, mailpit) are treated as deployed container images, not as bundled project code. Residual risk: not confirmed by the organiser; optional follow-up: ask (cadev@systalink.com) before submission.
  - [ ] Optional evidence: produce a licence report of the built container images (`api`, `migrate`, `web`) to back the decision above.
  - [x] DONE 2026-11-12 (owner instruction): busybox removed from the API image; `/healthcheck` replaces the wget probe (image HEALTHCHECK, Compose health check, smoke test). Not removed: `busybox` in the `acme-init` helper and the Alpine and Debian packages of the web, Traefik and PostgreSQL images (covered by the base-image decision above).
  - [x] DECIDED 2026-11-12 (owner): `isbot` (Unlicense, a public-domain dedication) is accepted; Unlicense is already on the allowed list of `scripts/check-licenses.py`. Base-image decision above. Both are recorded in `THIRD_PARTY_NOTICES.md`.
- [ ] **Titularity:** each member declares titularity at submission. Every member.
- [x] **Licences declared and verified (2026-11-13):** the direct-dependency inventory is in `THIRD_PARTY_NOTICES.md`; licences were verified against the upstream licence files on 2026-11-13 (method and limits in the "Verification" section of that file); the brand-asset copyright is an owner decision.
- [x] **AI usage declared:** tools and the parts concerned, in `AI_USAGE.md`. Every member for their own work. Closed by the owner on 2026-11-13.
- [x] **Brand assets generated with AI (closed by the owner, 2026-11-13):** the record in `AI_USAGE.md` is closed; no further licence review.
- [ ] **Technical quality is 30% of the score and includes the team's ability to explain the code:** schedule a code walkthrough with all four members before the submission. Owner sets the slot.
- [ ] **Peer vote is mandatory for every member** between Sun 2026-11-15 18:00 and Mon 2026-11-16 23:59 GMT (H10).
- [ ] The names in `docs/PRD.md` section 12 match the declared team (flag only; ROADMAP section 9).
- [x] Licence decisions are recorded in `THIRD_PARTY_NOTICES.md` and closed by the owner (2026-11-13): isbot (Unlicense) accepted, base images treated as not bundled code, the AI-generated brand assets accepted.

Open owner items outside CADEV (IDs and decisions in ROADMAP section 8, 10 and 11):

- [x] **H4 (DONE 2026-11-12):** `quality` is now a required status check on `develop` and `main`, next to "Validate branch promotion" (strict). Open sub-item: a review rule for `main` once the team pushes (with enforce-admins on, one required approval would block owner promotions until a teammate reviews).
- [ ] **AI review in CI (optional, owner):** `.github/workflows/code-review.yml` is added but OFF. To enable: run `claude setup-token`, store the result as the secret `CLAUDE_CODE_OAUTH_TOKEN`, set the variable `CODE_REVIEW_ENABLED=true` (check Anthropic's terms for a subscription token in shared CI first). Locally, everyone with Claude Code can run `/delegate-review`; `make review` needs an LLM provider.
- [ ] **Review gate (owner):** `.github/workflows/code-review.yml` job `gate` (check "Review gate") and `scripts/review-gate.py`; sensitive paths in `.github/sensitive-paths.txt`. After it is on `develop`: (1) add "Review gate" as a required check on `develop`, (2) watch it in `report` mode (variable `CODE_REVIEW_GATE`, default) over the next PRs, (3) switch to `enforce` once false alarms are known to be rare (before the Sat 14 Nov 18:00 freeze at the latest), `off` is the kill switch. Override a CRITICAL with the label `owner-override`.
- [ ] **H6, H7:** Datacloud account and active product; DNS for `226os.tech` and the staging subdomain (PROPOSAL `staging.226os.tech`, owner to confirm). Nothing is deployed; `PUBLIC_HOST` and `ACME_EMAIL` are empty.
- [x] **H8 (DONE; alias verified by the owner, 2026-11-12):** the Gmail settings live in the staging secrets file (`secrets/staging-test.enc.env`). Gmail rewrites `From` to the authenticated account unless the sender is a verified "Send mail as" alias, the alias for the branded sender `noreply@226os.tech` was already verified under "Send mail as" (owner statement, 2026-11-12). Remaining: repeat the real-inbox test on the staging host once it exists.
- [ ] **H12 / V2:** read the project's real AI Studio limits in the console (V2 OPEN: the 3 probes of `aicheck limits` do not measure them).
- [ ] **H10:** vote reminders, with a named follow-up person (Sun 17:00).
- [x] **H16 (DONE 2026-11-12):** the owner approved PRD amendment 3.2. Still open (H13): the S9 contract PR and the SPEC story for S9 (and S10 if GO), see S9.
- [ ] **H19:** run the reviewed S4 content script on `production-live` (see S4 and the release-day checklist).
- [ ] **Issue #46:** give each collaborator their own age key and re-encrypt `secrets/*.enc.env` for them (after the sync meeting). Owner.
- [ ] **B4:** reviewed release workflow (protected `production` environment, exact `develop` SHA, image digests). The guard in `mk/50-ops.mk` is not weakened. Before Sunday.
- [ ] **O4:** no resend-verification operation exists; the owner rules. **C3 to C10, C12, C13** are open (ROADMAP section 10).

## Shared pieces (O2: each is a proposal; the backend lead confirms or reassigns)

- [x] **Session resolver port with an anonymous default** (S0, PR #33). Used by optional-session reads: `listProjects`, `getProject`, `listContributionNeeds`, `getTrainingSeries`, `getProfile`, `search`, `getSession`.
- [ ] **Session resolver, session-backed implementation** (S1 provides it; S0 only has the anonymous default).
- [ ] **PostgreSQL rate limiter** (`rate_limit_buckets`, ADR-0003 windows). S1 builds it as a reusable port; other scopes: anonymous search 60/IP/min (S2); authenticated writes 60/user/min (S3, S4, S5); GitHub import 10/user/hour (S5); assistant (S9).
- [ ] **`EMAIL_NOT_VERIFIED` check.** S1 exposes `email_verified` on the subject, and each publishing service checks it.
- [ ] **`links[]` relation writer and resolved-links reader** (`content_relations`, ADR-0005 point 13). `mibienpanjoe` builds the reader in S2 and the writer in S4; S3 and S5 reuse them. `links` appears in `QuestionCreate`, `QuestionUpdate`, `ProjectCreate`, `ProjectUpdate`, `ArticleCreate`. 2026-11-13: the reader is in PR #83; the writer is still S4.
- [ ] **`audit_log` writer.** Proposal: S6 builds it; S1 security events also go to `audit_log` (DATABASE_DESIGN), so the lead decides the order.
- [x] **Idempotency-Key** on the six operations (`createQuestion`, `createAnswer`, `createProject`, `createContributionNeed`, `createApplication`, `createLearnArticle`; `castVote` does not bind the header). Built in S0 (PR #34).
- [ ] **Contract PR for the Idempotency response headers** (own PR, R7): declare `Idempotency-Replayed` and `Idempotency-Body-Omitted` on the six bound operations. PROPOSED by S0; the owner approves by merging. OBSERVED 2026-11-12: `api/openapi.yaml` already declares both headers; the owner confirms the approval and ticks this. Until then ADR-0004 point 2 applies (clients ignore unknown headers).
- [x] **FIRST CONTRACT PR (one entry; author `mibienpanjoe`; reviewers `Y4NN777` and `PANK4SS`, R7; then `make generate` for `apigen` and `schema.d.ts`).** PR #79, 2026-11-12. The S5 docs follow-ups in this bullet stay open. Carries:
  - C2: `listLearnArticles` `topic` retyped as the topic code (pattern `^[a-z][a-z0-9-]{1,29}$`, updated description); the same `topic` parameter added to `listTrainingSeries`; an optional `topic` on `ArticleCreate`;
  - C1: `decideApplication` x-authz corrected to owner-only (cited by operationId, not by line);
  - C14b: `profile_visibility` (`public|members|private`, default `public`) in the profile read and update schemas, so `PATCH /me/profile` accepts it, and `deny_status 404` kept on `getProfile`.
  - S4 handlers, E4 topic feeds and E5 wait for it. Docs follow-ups: `AUTHORIZATION_MATRIX.md` and `DATABASE_DESIGN.md` "owner or maintainer" to "owner only" (see S5).
- [ ] **SHIP 4.2 smoke text:** `/healthz` is checked at the edge and `/readyz` inside the container (B2); confirm `docs/SHIP.md` 4.2 says exactly that (S8).

---

## S0 — API core: DONE

- **Owner:** `Y4NN777`. **Reviewer:** `NICE-DEV226`.
- **Status:** DONE (PRs #31 to #36, 2026-11-12). 16 of 16 tasks checked; B1 (authz gate), B2 and B3 DECIDED, mailer and Gmail test mail (H8) done. Full block and notes: [archive](docs/archive/tracker/TODO-archive-2026-11-12.md#s0-api-core-original-lines-76-166).

- [ ] Replace the gate's HTTP probe that expects 501 by a method-set check of the production `apigen.ServerInterface` implementation (a method not overridden is still the `apigen.Unimplemented` one), then drop the probe.
- [ ] Optional (`mibienpanjoe`): a wider `idempotency_keys.response_body` column or explicit header columns (the replayed body is omitted above 16 KiB).
- [ ] C5: `METHOD_NOT_ALLOWED`, the bare 501 of `apigen.Unimplemented` and a 405 without `Allow` are outside the contract enum; the owner rules.

---

## S1 — Identity

- **Owner:** `NICE-DEV226`. **Reviewer:** `mibienpanjoe` (the lead may co-review the auth surface). **Helpers:** `PANK4SS` on demand.
- **Branches:** `feature/identity/auth-sessions`, `feature/identity/profiles`
- **Depends on:** S0 (merged). Argon2id parameters are pinned after the benchmark on the target host (S8, ADR-0003). The `getProfile` and `updateMyProfile` visibility work needs the `profile_visibility` migration (`mibienpanjoe`), and that PR must also update `TestFreshMigrationToHead` (hardcodes the migration count).
- **Operations** (see `x-authz` in [api/openapi.yaml](api/openapi.yaml) by operationId; rules in ADR-0003):
  - [x] `getCsrfToken` (public; pre-session token) — PR pending
  - [x] `signUp` (generic 202, no enumeration, 5/IP/h, sends the verification email) — PR pending
  - [x] `signIn` (generic `INVALID_CREDENTIALS`, equal timing, 5 failures/account and 20/IP per 15 min failing closed, rotates the session) — PR pending
  - [x] `signOut` (revokes only the caller's own session) — PR pending
  - [x] `getSession` (never 401) — PR pending
  - [x] `verifyEmail` (single use, 24 h, 422 `INVALID_TOKEN`) — PR pending
  - [x] `requestPasswordReset` (always 202; 3/account and 10/IP per hour) — PR pending
  - [x] `confirmPasswordReset` (30 min; revokes all the user's sessions) — PR pending
  - [ ] `getMe` (the only place the email is returned)
  - [ ] `updateMyProfile` (own profile; JSON Merge Patch; accepts `profile_visibility`, default `public`)
  - [ ] `getProfile` (C14b: `members` gives anonymous 404, `private` owner only, the 404 body identical to unknown or deleted; moderators no special access; suspended users out of scope for P0)
- [ ] Bylines on published content keep showing username and display name whatever `profile_visibility` says; deleted accounts show "membre supprimé" (ADR-0005 point 15).
- [x] Argon2id m=64 MiB, t=3, p=1 with a semaphore (implemented; the target-host benchmark stays with S8, K6 check pending there).
- [x] Cookie flags per profile (ADR-0003). [x] CSRF with `X-CSRF-Token` plus an `Origin` / `Sec-Fetch-Site` check. [x] Session timeouts: 2 h idle, 7 days absolute.
- [ ] O4: no resend-verification operation exists; ask the owner (flagged in the S1 PR description).
- **Tables:** `users`, `user_credentials`, `profiles`, `user_skills`, `sessions`, `email_verifications`, `password_resets`, `rate_limit_buckets`; reads `countries`, `tags`.
- **Acceptance:** `US-IDENTITY-001/2` (with S7); ADR-0003 "Evidence" (session expiry, rotation, revocation; CSRF cases; enumeration and timing; 429 with `Retry-After`; no token in logs). Negative authz per [AUTHORIZATION_MATRIX](docs/spec-kit/AUTHORIZATION_MATRIX.md) identity rows, plus `getProfile` per `profile_visibility` x caller (anonymous, member, owner, moderator) with identical 404 bodies.
- **Prototype pages:** `connexion`, `inscription`, `verifier-email`, `mot-de-passe-oublie`, `reinitialiser`, `bienvenue`, `profile`, `parametres`.
- **Allowed paths:** `src/api/internal/identity/**`, `db/queries/identity.sql`, `db/queries/platform.sql`, generated `dbgen` (via `make generate`), `scripts/public-routes.txt` (anonymous auth routes, with reasons). **DoD:** DoD-common plus cookie inspection per profile and a redacted email-flow log through Mailpit. **Stop:** if ADR-0003 values would need to change.
- **Status:** IN PROGRESS — `feature/identity/auth-sessions` implements the 8 auth operations end to end (sign-up through password reset, CSRF, session resolver, PostgreSQL rate limits, pgtest integration); `getMe`, `updateMyProfile`, `getProfile` and the profile-visibility matrix remain for `feature/identity/profiles`.

---

## S2 — Discover and search

- **Owner:** `mibienpanjoe` (accountable). **Reviewer:** `NICE-DEV226` (`Y4NN777` if `NICE-DEV226` co-wrote). **Helpers:** `Y4NN777`, `NICE-DEV226` (never migrations), `PANK4SS` on demand.
- **Branches:** `feature/discover/catalogue-read`, `feature/discover/search`. **Depends on:** S0 (merged); the error report and the anonymous-search 429 also need S1.
- **Operations:**
  - [x] `listResources` (public; provenance on each item). PR #83, 2026-11-13.
  - [x] `getResource` (public; resolved related links, visible items only). PR #83, 2026-11-13. Outgoing links are published projects and articles; incoming links are published questions. `content_relations` has no series.
  - [x] Review of PR #83 (2026-11-13): the batched queries for related links, discussions and SDKs keep at most 20 rows per resource (a window per resource), so one resource cannot starve the others on a page (integration test `TestCatalogueBatchedQueriesCapPerResource`, red first). A related resource leaking through an unpublished provider cannot happen: `content_relations_pair_ck` allows no resource-to-resource link. The anonymous `q` throttle waits for the S1 limiter (60/IP/min, see Shared pieces).
  - [x] `listResourceSources` (public; a stale source is never shown as verified). PR #83, 2026-11-13. A source is `verified` only when it is the resource's current verified source. No freshness window is defined, so the read does not invent `stale`.
  - [ ] `createResourceErrorReport` (member; reporter kept private; rate limited)
  - [ ] `search` (`q` 2 to 200, C4; `per_group` 1 to 20; published and visible only; anonymous 60/IP/min)
- [ ] Bounded sqlc queries with `EXPLAIN (ANALYZE, BUFFERS)` evidence and a bilingual relevance dataset (ADR-0005 point 6). 2026-11-13: the catalogue list plan is logged by `TestCatalogueReadsPublishedRowsOnly`. The relevance dataset stays with search.
- [ ] Resolved-links reader (Shared pieces). 2026-11-13: the reader is in `getResource` on this branch. The writer stays in S4. Demo fixtures are in SEED.
- **Tables:** `resource_categories`, `providers`, `resources`, `resource_countries`, `resource_sdk_languages`, `resource_sources`, `resource_tags`, `error_reports`, `countries`, `tags`; search reads questions, answers, projects, learn articles.
- **Acceptance:** `US-DISCOVER-001/1`, search steps of `US-SOLVE-001/1` and `US-LEARN-001/2`. Negative authz: no private or removed content through lists or search; anonymous error report 401; CSRF 403; reporter identity never exposed; rate limit.
- **Prototype pages:** `discover`, `resource`, `recherche`, part of `accueil`. **Allowed paths:** `src/api/internal/discover/**`, `src/api/internal/search/**`, `db/queries/**` (these files), generated `dbgen`, `db/migrations/**` only if needed (R6), `scripts/public-routes.txt`. **DoD:** DoD-common plus EXPLAIN plans. **Stop:** if a query cannot be bounded within ADR-0005.
- **Status:** IN PROGRESS (PR #83, 2026-11-13). `DONE` waits for the merge. Search and error reports stay open.

---

## S3 — Solve

- **Owner:** `NICE-DEV226`. **Reviewer:** `mibienpanjoe`. **Helpers:** `PANK4SS` on demand. **Branches:** `feature/solve/questions-answers`, `feature/solve/votes-accept`. **Depends on:** S1 (Idempotency-Key from S0 is merged).
- **Operations:**
  - [ ] `listQuestions` (published only; newest first with a UUID tie-breaker)
  - [ ] `createQuestion` (verified email, Idempotency-Key, at most 5 tags, `links[]`, snippets never executed)
  - [ ] `getQuestion` (`ETag`; 404 if not published)
  - [ ] `updateQuestion` (owner only, non-owner 403; `If-Match` required: 428 missing, 412 stale; authz and 404 run BEFORE the ETag check)
  - [ ] `listAnswers` (accepted first, then by score)
  - [ ] `createAnswer` (verified email; Idempotency-Key)
  - [ ] `castVote` (one vote per target; own content 403 `SELF_VOTE_FORBIDDEN`; server-side score)
  - [ ] `retractVote` (idempotent)
  - [ ] `acceptAnswer` (question author only; atomically replaces the previous accepted answer)
  - [ ] `unacceptAnswer` (idempotent)
- **Tables:** `questions`, `answers`, `votes`, `question_tags`, `question_sources`, `answer_sources`, `content_relations`, `tags`, `idempotency_keys`.
- **Acceptance:** `US-SOLVE-001/1`, `/2`, `US-IDENTITY-001/1`. Negative authz: writes 401 anonymous, 403 `EMAIL_NOT_VERIFIED`, 403 CSRF; non-owner PATCH 403; Bob's accept 403 with no change; self-vote 403; duplicate vote creates no row; closed or removed question rejects answers; links to private targets or by non-authors rejected.
- **Prototype pages:** `solve`, `question`, part of `accueil`. **Allowed paths:** `src/api/internal/solve/**`, `db/queries/**` (solve files), generated `dbgen`, `scripts/public-routes.txt` (3 public reads). **DoD:** DoD-common plus a vote-score reconciliation test and a concurrent-accept test on real PostgreSQL. **Stop:** any change to `api/openapi.yaml` (R7).
- **Status:** PLANNED

---

## S4 — Learn

- **Owner:** `mibienpanjoe` (accountable). **Reviewer:** `NICE-DEV226` (`Y4NN777` if `NICE-DEV226` co-wrote, unless `Y4NN777` co-wrote too). **Helpers:** `Y4NN777` (default: S4 and the content script, after the Fri 13 Nov 23:00 checkpoint), `NICE-DEV226`, `PANK4SS` on demand.
- **Branch:** `feature/learn/articles-series` (C8). **Depends on:** S0 for reads; S1 for publishing; **the first contract PR (C2) merged**. C3 (`my_progress`) is open.
- **Operations:**
  - [ ] `listLearnArticles` (public; `topic` = topic code, after the first contract PR)
  - [ ] `createLearnArticle` (no session 401; publishing needs a verified email 403, a `draft` does not; Idempotency-Key; `links[]`; keeps the original language)
  - [ ] `getLearnArticle` (original-language body plus `lang`; resolved links)
  - [ ] `listTrainingSeries` (public)
  - [ ] `getTrainingSeries` (public; `my_progress` blocked on C3)
- [ ] **Content script for the free series (B5 DECIDED; H19).** Written by `mibienpanjoe`, helped by `Y4NN777`, **run by `Y4NN777`**. Path `scripts/content/`: plain SQL, one-time, PR-reviewed, neither a seed nor a migration. Creates one `training_series` and 3 to 4 tutorial `learn_articles` referenced by `training_lessons.article_id` (original team text, `provenance='live'`, attributed to the owner's real account looked up by username, never a hard-coded id). Runs with an explicit confirmation like `migrate-up`, after a backup; first on `staging-test`, then on `production-live`. Record each run date in the Done log.
- [ ] `links[]` writer (Shared pieces).
- **Tables:** `learn_articles`, `learn_article_tags`, `training_series`, `training_lessons`, `learn_topics`, `tags`, `content_relations`. `learn_progress` is P1.
- **Acceptance:** `US-LEARN-001/1` to `/4`. Negative authz: anonymous publish 401 and nothing created; unverified publish 403; CSRF 403; a draft visible only to its author; removed content never leaks.
- **Prototype pages:** `apprendre`, `article`, `formation`, `ecrire`. **Allowed paths:** `src/api/internal/learn/**`, `db/queries/**` (learn files), generated `dbgen`, `scripts/public-routes.txt`, `scripts/content/**` (`Y4NN777` co-reviews). **DoD:** DoD-common plus a test that original-language content is preserved. **Stop:** do not merge the S4 handlers before the first contract PR has merged.
- **Status:** PLANNED

---

## S5 — Build

- **Owner:** `mibienpanjoe` (accountable). **Reviewers:** `NICE-DEV226`; when `NICE-DEV226` co-writes, `Y4NN777` reviews and co-reviews the GitHub fetch and anti-SSRF code (K10). **Helpers:** `NICE-DEV226` (default first helper after Solve; the Fri 13 Nov 23:00 checkpoint decides), `Y4NN777`, `PANK4SS` on demand.
- **Branches:** `feature/build/projects-needs`, `feature/build/applications`. **Depends on:** S1. C1 DECIDED: project owner only, no maintainer delegation at P0; the applicant may withdraw.
- **Operations:**
  - [ ] `listProjects` (public projects plus the caller's own private ones)
  - [ ] `createProject` (verified email, Idempotency-Key, GitHub URL allowlist normalized and unique, fetch with timeouts and anti-SSRF, 10 imports/user/hour)
  - [ ] `getProject` (private projects 404 to everyone but the owner)
  - [ ] `updateProject` (owner only; non-owner 403 on public, 404 on private; `github_url` immutable)
  - [ ] `listContributionNeeds` (follows the project's visibility)
  - [ ] `createContributionNeed` (owner, C1; Idempotency-Key)
  - [ ] `createApplication` (verified email; not on your own project; duplicate 409 `DUPLICATE_APPLICATION`; closed need 409 `STATE_CONFLICT`; re-apply only after withdrawal; replay returns the original 201)
  - [ ] `listProjectApplications` (owner, C1; a third party gets 403 or 404 and no count)
  - [ ] `decideApplication` (applicant may withdraw; only the project owner accepts or declines; invalid transition 409; cross-project IDOR 404)
- [ ] Correct the `decideApplication` x-authz ("owner or maintainer" to owner only): carried by the first contract PR (Shared pieces).
- [ ] In the S5 PR, correct `AUTHORIZATION_MATRIX.md` (the three "owner or maintainer" rows for project update, needs and applications) to "owner only" (C1). This narrows ADR-0005 point 10; the owner records it.
- [ ] Docs follow-up (S5 PR or a docs PR): `DATABASE_DESIGN.md` sections 7 and 10 and the `project_maintainers` description still say "owner or maintainer"; correct to owner only (C1). Owner `mibienpanjoe`, reviewed by `NICE-DEV226`.
- **Tables:** `projects`, `project_maintainers`, `project_tags`, `contribution_needs`, `need_skills`, `applications`, `content_relations`, `tags`, `idempotency_keys`, `rate_limit_buckets`.
- **Acceptance:** `US-BUILD-001/1`, `/2`. Negative authz: invalid GitHub URL, SSRF targets and rate limit rejected; non-owner cannot mutate and a private project is 404; a `project_maintainers` row grants nothing; duplicates, closed needs, re-apply after decline rejected; no count leaks; cross-project IDOR blocked.
- **Prototype pages:** `build`, `project`. **Allowed paths:** `src/api/internal/build/**`, `db/queries/**` (build files), generated `dbgen`, `scripts/public-routes.txt`, `docs/spec-kit/AUTHORIZATION_MATRIX.md` (the three C1 rows only; owner review). **DoD:** DoD-common plus an SSRF test suite (private IPs, redirects, non-GitHub hosts) and a double-click application test. **Stop:** any `api/openapi.yaml` change outside a contract PR (R7).
- **Status:** PLANNED

---

## S6 — Moderation (built last)

- **Owner:** `Y4NN777`. **Reviewer:** `mibienpanjoe`. **Helpers:** `PANK4SS` on demand. **Branch:** `feature/community/moderation` (domain choice is a GUESS). **Depends on:** S0 (merged) and S1; targets come from S2 to S5 (S10 if built).
- **Operations:**
  - [ ] `hideContent` (`targetType` question, answer, resource, project, article or relation; content `status='removed'` plus `removed_at`, relation `status='hidden'`; an `audit_log` row in the same transaction; already hidden 409)
  - [ ] `restoreContent` (same rules; not hidden 409)
  - [ ] `resolveErrorReport` (audited; reporter identity never returned)
- [ ] `audit_log` writer (Shared pieces).
- [ ] **Documented one-time moderator promotion** of the owner account: `users.role='moderator'` through an audited operational step (ADR-0003). **No moderator is seeded.**
- [ ] **Emergency SQL runbook** (`status='removed'` plus `removed_at`, or `status='hidden'` for a relation, plus an `audit_log` row). Path `docs/runbooks/moderation.md` is a GUESS the owner confirms; the file does not exist yet.
- [ ] Moderation UI: no prototype page exists; the owner decides between an API demo and small controls built by `PANK4SS`.
- **Tables:** `audit_log`, `users.role`, `error_reports`, every target table. **Acceptance:** ROADMAP section 3, item 9. Negative authz: non-moderator 403, anonymous 401, CSRF 403, audit row in the same transaction (rollback leaves no orphan), the moderator never edits content.
- **Allowed paths:** `src/api/internal/moderation/**`, `db/queries/**` (moderation files), generated `dbgen`, `docs/runbooks/moderation.md`. **DoD:** DoD-common plus the same-transaction audit test. **Stop:** if scope would grow beyond ADR-0003.
- **Status:** PLANNED

---

## S7 — Frontend (prototypes become React pages)

- **Owner:** `PANK4SS`. **Reviewer:** `Y4NN777`. **Branches:** `feature/<page domain>/web-<page>`. **Depends on:** typed fixtures from `src/web/src/api/schema.d.ts` until each endpoint merges. **Order:** shell, i18n and shared components; public reads; identity pages; write flows.
- [ ] **Pillar emphasis (O7):** navigation, home (`accueil`) and FR/EN copy make Learn & Grow and Connect prominent; Solve, Discover and Build stay full pillars in the same navigation, on desktop and mobile (`DESIGN.md` section 5); IA and copy only, no new endpoints; nothing non-functional announced (Connect shows only the E5 directory as live). Interface copy "Afrique de l'Ouest d'abord, puis toute l'Afrique" (FR/EN) is DECIDED.
- [ ] **Shell:** `getSession` with header sign-in and sign-out; FR/EN labels without altering content; a demo badge for `provenance='demo'`; loading, empty and error states.
- [ ] **Profile visibility setting** on `parametres` (C14b) with FR/EN copy saying that bylines still show the name. It slips before the column if the team is behind.
- [ ] **Page map** (a page is done when its operations are wired to the live endpoint): `accueil` (`listQuestions`, `listResources`, `listProjects`, `listLearnArticles`); `discover` (`listResources`; staging and production show only real `live` entries, K39; `local-dev` shows SEED fixtures with a demo badge); `resource` (`getResource`, `listResourceSources`, `createResourceErrorReport`); `recherche` (`search`, plus the S9 panel); `apprendre` (`listLearnArticles`, `listTrainingSeries`, E4 topic feeds after the first contract PR); `article` (`getLearnArticle`, E4 "Poser une question sur cette leçon" opens `solve` with `links[article]`); `formation` (`getTrainingSeries`, the free series from the S4 content script or SEED); `solve` (`listQuestions`, `createQuestion`, the S9 panel, E4 pre-fill); `question` (`getQuestion`, `listAnswers`, `createAnswer`, `castVote`, `retractVote`, `acceptAnswer`, `unacceptAnswer`, `updateQuestion`); `connexion` (`getCsrfToken`, `signIn`); `inscription` (`signUp`); `verifier-email` (`verifyEmail`); `mot-de-passe-oublie` (`requestPasswordReset`); `reinitialiser` (`confirmPasswordReset`); `bienvenue` (`updateMyProfile`, GUESS); `profile` (`getProfile`); `parametres` (`getMe`, `updateMyProfile`, visibility setting); `ecrire` (`createLearnArticle`); `build` (`listProjects`, `createProject`); `project` (`getProject`, `updateProject`, `listContributionNeeds`, `createContributionNeed`, `createApplication`, `listProjectApplications`, `decideApplication`); `connect` (E5 people list; real registered members only, "later" badge removed only once the endpoint is live, `connect-spaces` stays "bientôt"). `auth` is a legacy redirect and `_template` is not a page.
- **Acceptance:** `US-IDENTITY-001/2`, `US-LEARN-001/4`, `US-ACCESS-001/1`, plus each slice's journeys. **Negative cases:** the UI never grants authority; no token in `localStorage`; every unsafe call sends `X-CSRF-Token`. **Allowed paths:** `src/web/**`, `DESIGN.md` (ADR-0008); `pnpm-lock.yaml` one change per PR. **DoD:** DoD-common (web above 70%), Vitest and Testing Library by role and label; ADR-0007 visual, axe and responsive evidence is **NOT RUN** until produced. **Stop:** if a page needs a field not in `schema.d.ts` (R7).
- **Prototype alignment** (PRD amendment 3.2, approved 2026-11-12; owner `PANK4SS`, reviewer `Y4NN777`). `prototypes/html/**` is outside the S7 allowed paths, so the owner confirms these edits are in scope first.
  - [ ] Navigation and home emphasise **Apprendre** and **Connect**; the other pillars stay in the same navigation.
  - [ ] The `connect` page becomes the people-directory preview (country, language, skill filters; only fields the caller may see; no email); spaces stay "bientôt"; mentoring and opportunities "coming soon"; no matchmaking.
  - [ ] Learn topic labels match the 14-topic list and resource category labels the 18-category list (E2), with none removed.
  - [ ] An assistant panel mock in `recherche` and in the `solve` ask flow, marked as prototype: cited sources, "Comment vérifier", an AI label, no vote or accept control, a notice that the question is sent to the provider.
  - [ ] A profile visibility setting on `parametres` (same as above).
  - [ ] Replace fictional people and data with labelled demo data, or mark each remaining fictional item visibly.
  - [ ] Keep the fictional markers honest (banners and badges such as `discover.fictional.*`, "lien fictif", the `FICTIONAL` header comments in `assets/data/*.js`); check the landing script that strips "(fictif)" from labels. No fictional item is presented as real activity.
- **West Africa first** (region and positioning DECIDED 2026-11-12; the rest PROPOSED):
  - [ ] **DECIDED:** the region is a **UI constant in `src/web`**, no migration, no column: UN geoscheme "Western Africa" restricted to our table: `BJ, BF, CV, CI, GM, GH, GN, GW, LR, ML, MR, NE, NG, SN, SL, TG` (16). **To verify against the UN M49 list; the owner may adjust.** Default filter "Afrique de l'Ouest" with a one-click "Toute l'Afrique". Owner `PANK4SS`, reviewer `Y4NN777`.
  - [ ] PROPOSED: Discover entries curated by the team, starting with the West African context (facts and source links only, no copied text).
  - [ ] PROPOSED: the Connect directory highlights West African builders and the diaspora, within the E5 visibility rules.
  - [ ] PROPOSED: the Learn series examples use low-bandwidth and Mobile Money contexts.
- **Status:** PLANNED

---

## S8 — Ops, release and CADEV

- **Owner:** `Y4NN777`. **Reviewer:** `NICE-DEV226` or `mibienpanjoe`. **Helpers:** `PANK4SS` on demand. **Branches:** `chore/ops/staging`, `chore/ops/release-workflow`, `chore/ops/e2e-harness`, `chore/ops/perf-harness`. **Depends on:** H6 to H8.
- [ ] H4, H6, H7, H8 follow-up, H12, B4: see "Open CADEV checklist and owner decisions" above (kept there once).
- [ ] `staging-test` up from pinned images (`make up-staging-test`, `.env.staging-test` outside Git). **No seeds on staging:** only real `live` content (owner correction B). The S4 content script is rehearsed there. Repeat the Gmail real-inbox test on the staging host.
- [ ] Argon2id benchmark on the target host, handed to S1.
- [x] B2 smoke (DECIDED: `/readyz` is NOT routed at the edge). `scripts/smoke.sh <profile> [https://host]` (and `make smoke PROFILE=... BASE_URL=...`, 2026-11-13) checks `/healthz` from outside (certificate verified), the http to https redirect, that `/readyz` is not answered by the API at the edge, and `/readyz` inside the container with `/healthcheck /readyz`; twelve script tests. NOT RUN against a real host (nothing is deployed). Application flows (SHIP 4.3 to 4.5) are not covered.
- [ ] **B4** release workflow: a DRAFT with the steps, the existing tools, what is missing and five owner questions is in `docs/RELEASE_DRAFT.md` (2026-11-13). Not active: no workflow file exists and the production guard is unchanged. Next: owner answers (registry, host access, approver), then a reviewed `.github/workflows/release.yml` (needs H4 environment, H6, H7).
- [ ] Implement `make test-e2e` (Playwright P0 flows, FR/EN, mobile, axe) and `make perf` (dataset, EXPLAIN, API latency, page weight); both are NOT IMPLEMENTED today.
- [ ] Backup and restore drills on `staging-test` and `production-live` (`scripts/backup.sh`, `scripts/restore-check.sh`; the local-dev drill is done, see archive).
- [ ] Rollback drill (redeploy the previous image by digest), with evidence.
- [ ] G7 evidence on the target SHA: build SHA, digests (`make build` writes `reports/image-digests.txt`), audit, backup, smoke, rollback.
- [ ] Production promotion PR from `develop` to `main`, with owner approval.
- [ ] CADEV checklist (above and ROADMAP section 9), including H10 and the PRD section 12 names check. Submission URL `https://226os.tech`.
- **Allowed paths:** `Makefile`, `mk/**`, `scripts/**`, `compose.yaml`, `deploy/**`, `.github/**`, `docs/quality/**`, `docs/evidence/**` (GUESS). **DoD:** each item has evidence in the PR or under the evidence path for the target SHA; `make quality-full` on the `develop` SHA reports G5 and G6 `PASS`, or `NOT RUN` with a dated owner exception. **Stop:** a production deploy without G7 evidence and owner approval is forbidden.
- **Status:** PLANNED

---

## S9 — Grounded AI assistant (P0 by owner decision, amends PRD 5.8)

- **Owner:** `Y4NN777` (contract PR, LLM port, backend). **Second developer:** `NICE-DEV226`, after Solve. **UI:** `PANK4SS`. **Reviewers:** `mibienpanjoe` for the owner's backend PRs; `Y4NN777` and `PANK4SS` for the contract (R7).
- **Branches:** `feature/intelligence/assistant-contract`, then `feature/intelligence/assistant`. **Depends on:** S1 (subject, rate limiter), S2 (search), H12.
- **Provider** (decision 2026-11-12): Google AI Studio, Gemma 4. V3 DECIDED: free tier (development and evaluation; the assistant stays disabled by config until it is integrated). The key lives only in the staging and production env files (`secrets/staging-test.enc.env`, SOPS and age, ADR-0009). The LLM port (`platform/llm`, built, not wired into the router; `aicheck` command) keeps the provider swappable; development and the golden set run against the fake adapter. Terms facts and V1 to V4: ROADMAP section 10 (O5, V-items, K20).
- **Tasks**
  - [x] S9 adapter output cleaning (2026-11-13, `platform/llm`): `CleanAnswer` removes reasoning blocks (`<think>`, `<thinking>`, `<thought>`, also when never closed) and the adapter drops `thought` parts, so only the answer is returned; empty after cleaning is `ErrUnavailable`. Fixtures are synthetic: no real provider sample is committed. **Still to do (owner, NOT RUN):** run `aicheck ping fr` and `aicheck ping en` once, check the real reasoning shape against `CleanAnswer`, and add a recorded sample if it differs. One answer format (cited sources as links, no guarantee, "Comment vérifier") is applied by the S9 backend, not by the adapter.
  - [x] S9 tracking of the free tier (2026-11-13, `platform/llm`): `Tracker` wraps a `Client` per model and counts calls, OK, 429, 500, 503, other 5xx, timeouts, cancellations, rejected, not configured and invalid, with total, mean and max latency; failures log model, class, status and latency only (never prompt, answer or key). `StatusError`/`StatusOf` expose the HTTP status. Not wired into the router yet (S9 backend). Reading the numbers before the demo stays an owner task (V2).
  - [ ] `Y4NN777`: V2 (real quotas, console) and V4 (FR/EN latency and quality on the golden set), due Fri 13 Nov evening. V1 is DONE (`gemma-4-26b-a4b-it` and `gemma-4-31b-it`; staging uses `gemma-4-26b-a4b-it`); V3 DECIDED.
  - [ ] PROPOSED ADR for the AI integration (next free ADR number) deciding: data sent (question text and retrieved public content only), the cost cap or quota guard and where its counter lives (a new table needs a migration from `mibienpanjoe`; C19), timeouts (B3), provider-down and quota-exhausted behaviour, FR/EN behaviour, evaluation set and pass criteria, anonymous access (O5).
  - [ ] SPEC story with BDD and new `AUTHORIZATION_MATRIX.md` rows for the assistant, proposed in the ADR; the owner applies them (H13). PRD amendment 3.2 is approved and applied (see `docs/PRD.md`); the provider decision is `docs/decisions/0011-ai-assistant-provider.md`.
  - [x] Contract PR (branch `feature/intelligence/assistant-contract`, DONE for review): one operation, `askAssistant` (`POST /assistant/answers`, members only, verified email, CSRF), 200 `AssistantAnswer` (outcome `answered` or `no_source`, `generated_by: ai`, sources, `how_to_verify`), 429 and 503; `scripts/unimplemented-routes.txt`, `contract_test.go` (48) and `COVERAGE_SCOPE.md` updated. Original scope: new operation(s) in `api/openapi.yaml` with `x-priority`, `x-authz`, 429 and 503; reviewed by `Y4NN777` and `PANK4SS`; merges before any code (R7). Regenerate `apigen` and `schema.d.ts`; update the operation count in `src/api/internal/platform/apigen/contract_test.go` and the exclusion justification in `docs/quality/COVERAGE_SCOPE.md` (change-controlled: owner, reviewer, date). B1 applies to each new route.
  - [ ] Retrieval through the existing PostgreSQL search (S2) over published and public content only. No vector database.
  - [ ] Answers carry links to the sources used; the assistant refuses when there is no source and never invents guarantees (PRD 5.8).
  - [ ] Retrieved content is treated as untrusted (prompt injection).
  - [ ] A redaction step removes obvious secrets and emails before sending, with tests.
  - [ ] Per-subject rate limit, a quota or cost guard, and a clear "assistant unavailable" response. The page never fails.
  - [ ] No prompt logging that contains personal data or secrets.
  - [ ] A per-environment enable switch in config. The UI shows a visible notice that the question is sent to Google.
  - [ ] **Golden evaluation set:** questions with expected cited sources, on demo seeds (fixed UUIDs), run against the fake adapter in CI. The real-provider run is recorded separately (V4) and is NOT RUN until done.
  - [ ] `PANK4SS`: the assistant panel in `recherche` and in the `solve` ask flow (grounded answer with citations, or a refusal). "Similar questions" and "help phrasing" are **P1**, out of S9 scope.
  - [ ] `THIRD_PARTY_NOTICES.md` and `AI_USAGE.md` entries for the provider (owner).
- **E1 teaching rules** (accepted 2026-11-12; each is an acceptance test and appears in the golden set):
  - [ ] **AI-1** ground or refuse (zero published passages: the provider is **not called**; the answer links to search). [ ] **AI-2** server-validated citations (at least one, or a refusal). [ ] **AI-3** no guarantees about an API or a verified solution. [ ] **AI-4** every answer ends with "Comment vérifier". [ ] **AI-5** uncertainty and source dates stated when sources conflict or are undated. [ ] **AI-9** answers labelled AI, never accepted or voted (no `answers` row, no vote target). [ ] **AI-10** only the question text and public retrieved content are sent, after redaction. AI-6 and AI-7 are P1, out of scope.
- **Other acceptance:** an injected instruction is ignored; provider down gives "assistant unavailable"; the N+1 request gets 429; the golden set passes; failure modes in ROADMAP section 12.2. **Tables:** none planned (a quota counter may need one, C19).
- **Allowed paths:** `src/api/internal/assistant/**` (GUESS), `api/openapi.yaml` (contract PR only), generated `apigen` and `schema.d.ts`, `docs/quality/COVERAGE_SCOPE.md` (count line, owner approval), `docs/decisions/**` (ADR, owner), the S9 UI in `src/web/**`. **Negative authz:** depends on O5; always CSRF on unsafe methods, 429 per subject, no private or removed content in context or citations. **DoD:** DoD-common, golden set green on the fake adapter, and the recorded real-provider run with V4 numbers before the demo. The assistant stays disabled by config until it is integrated.
- **Status:** PLANNED (the LLM port, fake adapter, `aicheck` and the key are done; see archive)

---

## E2 — Taxonomy: add and relabel only

- **Owner:** `mibienpanjoe` (the only migration writer, R6). **Reviewer:** `NICE-DEV226`; `Y4NN777` co-reviews the pgtest change. **Helpers:** `Y4NN777`, `NICE-DEV226` on pgtest updates only, never migrations. **Branch:** `feature/platform/taxonomy-ai-topics`. **Depends on:** nothing; one migration at a time (R6).
- [x] Create the migration with `make migrate-new NAME=<slug>` and **use the number it gives** (C19 DECIDED: `00011` is not reserved; E2 is `00011` and `profile_visibility` `00012`, or the reverse). Forward-only, standard header. (2026-11-12: `00011_taxonomy_ai_topics`, `00012_profile_visibility`. PR #80.)
- [x] Docs follow-up **in the same PR**: correct the statements in `DATABASE_DESIGN.md` that reserve `00011` for the P1 tables (section 5 and its migration plan). Owner `mibienpanjoe`, reviewer `NICE-DEV226`. (2026-11-12.)
- [x] Add 5 learn topics: `learn-to-learn` (Apprendre et rester à jour / Learning & keeping up), `ai-assisted-dev` (Coder avec l’IA sans perdre la main / Coding with AI, keeping your skills), `ai-agents` (Construire et évaluer des agents / Building & evaluating agents), `open-models` (Modèles ouverts et IA en local / Open-weight & local models), `african-languages` (IA et langues africaines / AI & African languages). (2026-11-12.)
- [x] Relabel only (codes and keys unchanged): learn topic `ai` (Comprendre les LLM / Understanding LLMs); `mobile` (Mobile, USSD et faible débit / Mobile, USSD & low bandwidth); `open-source` (Contribuer à l’open source / Contributing to open source); resource category `cloud-hosting` (Hébergement, cloud et GPU en Afrique / Hosting, cloud & GPUs in Africa). (2026-11-12.)
- [x] Add 3 resource categories (each with a `sort_order`, step 10): `ai-models` (Modèles et API d’IA / AI models & APIs), `ai-tooling` (Agents et outillage IA / AI agents & tooling), `ai-datasets` (Jeux de données et évaluation / Datasets & evaluation). (2026-11-12.)
- [x] Result: **14 learn topics and 18 resource categories**. `payments`, `ai`, `mobile-money` and `sms-ussd` stay valid (fixtures use them). (2026-11-12, asserted in pgtest.)
- [x] In the same PR, update the exact lists and counts in `pgtest/migrate_integration_test.go`, and `TestFreshMigrationToHead` (hardcodes the migration count and version list; the new head is 10 plus the new files). (2026-11-12: head is 12.)
- [x] K25: check whether seeds or fixtures tag `ai` content that belongs under `african-languages`. (2026-11-12: no seeds yet. The fixture series "Introduction à l'IA" stays on `ai`.)
- [ ] Docs follow-up, outside this PR: the `DATABASE_DESIGN.md` lines about the lists.
- [x] **Separate migration file: `profile_visibility` (C14b DECIDED).** Written by `mibienpanjoe` only, numbered by `make migrate-new`; may share the PR but is its own file: `ALTER TABLE profiles ADD COLUMN profile_visibility text NOT NULL DEFAULT 'public'` with a named CHECK `IN ('public','members','private')`. Update the pgtest structure and constraint tests in the same PR. Docs follow-up: the `profiles` entry in `DATABASE_DESIGN.md`. (2026-11-12: `00012_profile_visibility`, check `profiles_profile_visibility_ck`. 2026-11-13: sqlc model `Profile.ProfileVisibility` regenerated so the static gate matches the column.)
- Removals are **DEFERRED** until after CADEV and need explicit human approval: `ussd-messaging`, `geo-data`, the `logistics` merge. Taxonomy rows are data: they change through a new forward-only migration or a moderator path, never by editing old migrations.
- **Allowed paths:** new files in `db/migrations/`, `src/api/internal/platform/pgtest/**`. **DoD:** DoD-common; `make test-integration` passes on fresh and upgrade runs; the migration immutability and no-`Down` checks pass (`make lint`).
- **Status:** IN PROGRESS (PR #80, 2026-11-13). `DONE` waits for the merge.

---

## SEED — Original demo fixtures and `make seed`

- **Owner:** `PANK4SS` (original fixture text, SQL in `db/seeds`, `make seed`). **Reviewer:** `mibienpanjoe`; `Y4NN777` also reviews the `mk/50-ops.mk` change (area owner) and curates demo content. **Purpose:** **`local-dev` only** (owner correction B); staging and production never hold demo data. **Branches:** `feature/platform/demo-seeds`, `chore/ops/make-seed`. **Depends on:** migrations 00001 to 00010, E2 (topic `ai-assisted-dev`), C15 (DECIDED).
- **Rules:** deterministic SQL in `db/seeds/` with fixed UUIDs, original team text; every root row `provenance='demo'`, shown as demo. Demo accounts (C15): username prefix `demo-`, email `@demo.226os.invalid`, **no `user_credentials` row**, `local-dev` only. No fake votes, endorsements, verifications or applications. External material only as source links on Discover entries (`resource_sources`), never copied text.
- [ ] Implement `make seed` (NOT IMPLEMENTED today): it **refuses everywhere except `APP_ENV=local-dev`** and fails closed. The current help text in `mk/50-ops.mk` says "local-dev and staging-test only", which contradicts the owner decision; correct the help text and the guard message in the same PR.
- [ ] **Zero demo rows (one entry, merged from SEED and the release-day checklist):** a check that finds zero `provenance='demo'` rows on **both** `staging-test` and `production-live`, run as part of the G7 evidence (DATABASE_DESIGN seed policy, extended to staging). No seeds ever run there.
- [ ] **The Learn free series (local copy):** a demo-fixture `training_series` of 3 to 4 short original lessons, each a `learn_article` of kind `tutorial` referenced by `training_lessons.article_id` (C13). The staging and production copy is the S4 content script.
- [ ] A few Discover fixtures (providers and resources with countries, categories, source links) and Build fixtures.
- [ ] Tests: seed provenance; demo accounts have no credentials; no votes or applications seeded; refusal tests for `staging-test` and `production-live`.
- **Allowed paths:** `db/seeds/**`, `mk/50-ops.mk` (the `seed` target, owner review). **DoD:** DoD-common plus `make seed` evidence on `local-dev`.
- **Status:** PLANNED (assigned; unblocked)

---

## E4 — Zero-migration interaction modes

- **Owner:** `PANK4SS` with `NICE-DEV226`. **Reviewer:** `Y4NN777`. **Branches:** `feature/solve/web-ask-about-lesson`, `feature/learn/web-topic-feeds` (C8). **Depends on:** S3 (`createQuestion`) and a lesson to exist (a member article from S4 or a SEED article); the topic feeds wait for the first contract PR (C2). No new table, no moderation system, no contract change.
- [ ] The "Poser une question sur cette leçon" button on `article` and `formation`: pre-fills the `solve` ask form; `createQuestion` sends `links: [{type: article, id}]` (an allowed relation pair; the question author authors the source item).
- [ ] Read-only topic feeds, after the first contract PR (C2): articles by topic code (`listLearnArticles?topic=<code>`); questions by tag equal to the topic code (`listQuestions?tag=<code>`).
- [ ] Tests: the pre-fill, the link created, an anonymous session redirected to sign-in; web coverage above 70%.
- **Allowed paths:** `src/web/**`; backend only in `src/api/internal/solve/**` if the link handling needs a fix. **DoD:** DoD-common plus an e2e step once S8 delivers the harness.
- **Status:** PLANNED

---

## E5 — Connect: people directory

- **Backend owner:** `NICE-DEV226`, after Solve, if time allows (it slips early). **UI:** `PANK4SS`. **Reviewers:** `mibienpanjoe` (backend), `Y4NN777` (UI); the contract PR needs `Y4NN777` and `PANK4SS` (R7). **Branches:** `feature/connect/profiles-contract`, then `feature/connect/people-directory`.
- **Depends on:** S1 (profiles, session resolver), the `profile_visibility` migration and the first contract PR. C14 DECIDED (ROADMAP section 10): listable only if the user is active, `profile_visibility` allows the caller, `open_to_collab = true` and `open_to_collab_visibility` allows the caller; fields per `field_visibility`; email never shown; filters only on caller-visible fields.
- [ ] **Contract PR first:** a new operation, suggested `GET /profiles` (`listProfiles`): filters `country`, `language`, `skill`; `cursor` and `limit` (20 default, 50 max, above 50 returns 422); `x-priority: P0`; `x-authz` `anonymous|member`; public summary fields only. `make generate`; raise the operation count in `contract_test.go` and `COVERAGE_SCOPE.md` (owner approval). P0 then has one more operation (K29).
- [ ] A new row in `AUTHORIZATION_MATRIX.md`, proposed in the contract PR; the owner applies it.
- [ ] Add the route to `scripts/public-routes.txt` with its reason (B1).
- [ ] Backend: a projection filtered by `field_visibility` per caller; email never selected; only listable profiles (deleted and suspended never); filters on visible fields only; stable order with a UUID tie-breaker; a rate limit per IP and per user (window is the owner's call; ADR-0003 has none for this route).
- [ ] `PANK4SS`: the `connect` page with loading, empty and error states, FR/EN, and a reset-filters action.
- [ ] **Office hours: a process note only.** One named mentor answers questions tagged `permanence` during a fixed weekly hour. The owner writes the note and names the mentor.
- **Not in scope** (coming soon, PRD 5.5): mentoring and matchmaking, CV review, interview preparation, study groups, opportunities.
- **Privacy tests (required):** hidden fields never appear for callers who may not see them; filtering on a hidden field never matches; no `email` key; one test per `profile_visibility` x caller and per `open_to_collab_visibility` x caller; `limit=51` 422; tampered cursor 400; N+1 past the rate limit 429.
- **Allowed paths:** `api/openapi.yaml` (contract PR), generated outputs, `src/api/internal/identity/**` or a new profiles package (the lead decides), `db/queries/**` (profiles), `scripts/public-routes.txt`, `src/web/**`. E5 adds no migration. **DoD:** DoD-common plus the privacy tests on real PostgreSQL and the failure modes in ROADMAP section 12.2.
- **Status:** PLANNED (C14 fully DECIDED; waits for the `profile_visibility` migration and the first contract PR)

---

## S10 — Live channels (CONDITIONAL)

- **Owner:** **unassigned**; conditional on the Sat 14 Nov 12:00 checkpoint. **Migrations:** `mibienpanjoe` (R6). **UI:** `PANK4SS`. **Reviewers:** OPEN. **Branch:** `feature/community/channels`. **Depends on:** a GO at the Sat 14 12:00 checkpoint (S0 to S5 on track), plus S1, S4, S5. Without a GO the status becomes `NOT BUILT (P2)` and this is reported honestly. It displaces other work first in the slip order (O6, ROADMAP section 12.1).
- [ ] PROPOSED ADR for the real-time transport (server-sent events with PostgreSQL `LISTEN/NOTIFY`, no Valkey or S3, ADR-0006; connection budget; the 30 s router timeout, B3; Traefik streaming; message bounds).
- [ ] Contract PR: channel operations, and extending `ModerationTargetType` to cover messages; reviewed by `Y4NN777` and `PANK4SS`; regenerate; update the operation count in `contract_test.go` and `COVERAGE_SCOPE.md` (owner approval); B1 applies.
- [ ] Migration with the next free number from `make migrate-new` (C19), written by `mibienpanjoe`.
- [ ] A channel is attached to an existing object (a project or a Learn series); no free-form chat. Messages persisted in PostgreSQL and delivered live; members only; rate-limited; carry `lang`, never translated.
- [ ] Moderation can hide messages (an S6 extension, audited).
- [ ] `PANK4SS`: a channel panel on the `project` and `formation` pages.
- **Negative authz:** anonymous 401; a private project channel 404 to non-owners; CSRF on posting; rate limits; a hidden message is never delivered. **Allowed paths:** the S10 module under `src/api/internal/**` (GUESS), `api/openapi.yaml` (contract PR), generated outputs, `db/migrations/**` (`mibienpanjoe` only), the channel UI in `src/web/**`. **DoD:** DoD-common plus a stream reconnect test and a moderation-hide test.
- **Status:** CONDITIONAL

---

## Release-day checklist (owner `Y4NN777`; target SHA recorded on each line; S8 holds the build tasks, this is the single deployment list)

- [ ] `develop` SHA frozen at Sat 18:00. `make quality-full` run on it, with its result (G1 to G6).
- [ ] `staging-test` deployed from that SHA by digest (`make build`, then `make up-staging-test`); DNS resolves for `226os.tech` and the staging subdomain first (H7).
- [ ] Staging smoke per `docs/SHIP.md` section 4: HTTPS, `/healthz` at the edge, `/readyz` inside the container (B2), auth, Q&A, vote, search, resource, project, application, anonymous reads, denied mutations, FR/EN, mobile, no secrets in logs.
- [ ] Staging rehearsal: the team's real accounts create real `live` content through the app, and the S4 content script runs on staging with the production procedure; record the date in the Done log.
- [ ] Backup taken and restore verified on a scratch database; rollback drill (previous image redeployed by digest).
- [ ] Zero `provenance='demo'` rows on **both** `staging-test` and `production-live` (see SEED).
- [ ] G7 evidence complete for the SHA (ROADMAP section 1, go/no-go rule); E2 migration is part of the forward migrations on that SHA.
- [ ] Production PR from `develop` to `main` with owner approval. The same SHA is promoted, never rebuilt.
- [ ] Production: backup, then `migrate-up` (`MIGRATE_CONFIRM` and `BACKUP_EVIDENCE`), then deploy through the release workflow (B4), then the smoke test.
- [ ] Owner account promoted to moderator in production (S6 runbook). S9 enabled or disabled per environment.
- [ ] S4 content script run on `production-live` by `Y4NN777` (H19), with confirmation and a backup first; record the run date in the Done log.
- [ ] Production content window: the team's real accounts create real `live` content through the app. Gmail SMTP must work in production, because publishing needs a verified email (K38; alias check in the owner items).
- [ ] CADEV submission of `https://226os.tech` (O1) around Sun 12:00 GMT, before 17:00 (ROADMAP section 9).
- [ ] Vote reminders sent. **Every member votes between Sun 18:00 and Mon 23:59 GMT.**

## Done log

Full log and the checked items: [archive](docs/archive/tracker/TODO-archive-2026-11-12.md). PR numbers below come from `git log` merge commits and the records already in this file.

| Date | Item | PR | By |
|---|---|---|---|
| 2026-11-12 | B0 bootstrap: API skeleton, migrations 00001 to 00010, codegen, web skeleton, Compose, Make, CI | #24 | `Y4NN777` |
| 2026-11-12 | CADEV team registration (H9); H1 identities and H3 ADR-0008 accepted | n/a | `Y4NN777` |
| 2026-11-12 | S0 B1 authz gate and its hardening (three route lists, stray-router scan, RED/GREEN) | #31, #32 | `Y4NN777` |
| 2026-11-12 | S0 api-mount: pgxpool, `/readyz`, apigen under `/api/v1`, validation, cursors, transactions, subject port | #33 | `Y4NN777` |
| 2026-11-12 | S0 Idempotency-Key, ETag helper; B2 and B3 decided | #34 | `Y4NN777` |
| 2026-11-12 | S0 mailer (port, SMTP adapter, fake, Mailpit, mailcheck) and Gmail test mail (H8) | #35 | `Y4NN777` |
| 2026-11-12 | S0 follow-ups: cursor expiry, problem+json timeout, media type, query guard | #36 | `Y4NN777` |
| 2026-11-12 | SOPS and age encrypted secrets (ADR-0009), `.env` deny patterns, owner age recipient | #44, #45, #48 | `Y4NN777` |
| 2026-11-12 | Backup and restore drill on local-dev; worktree helper; secrets init generates all profiles | #47, #49, #50, #52 | `Y4NN777` |
| 2026-11-12 | Staging secrets and AI Studio V1 to V4 results; V3 decided; README and AI wiring | #53, #56 | `Y4NN777` |
| 2026-11-12 | Parallel CI gate groups (ADR-0010); promotions `develop` to `main` and back-merges | #57, #54, #55, #58, #59 | `Y4NN777` |
