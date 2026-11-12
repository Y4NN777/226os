# 226os P0 roadmap to the CADEV submission

> **How to use this file.** It is the stable plan from Thu 2026-11-12 to Sun 2026-11-15. The live task state is in [TODO.md](TODO.md). Update this file in the same PR as the work that changes it, one line per change. Never rewrite someone else's entry. Add decisions only to the dated log at the end. People are named by GitHub login only.

> **Live status is in [TODO.md](TODO.md). History is in [docs/archive/tracker/](docs/archive/tracker/)** (what was moved out of this file: [ROADMAP-archive-2026-11-12.md](docs/archive/tracker/ROADMAP-archive-2026-11-12.md)).
>
> **Sources.** Every fact cites a repository file. **GUESS** marks an estimate with no repository source. **INFERRED** marks a conclusion drawn from files but not executed. All times are GMT.

## 1. Deadlines and scope

| What | When (GMT) | Source |
|---|---|---|
| CADEV submission, hard limit | Sun 2026-11-15 17:00 | `docs/SHIP.md` section 6 |
| Planned submission | Sun 2026-11-15 about 12:00 | owner decision, 2026-11-12 |
| Progress checkpoint (drop decisions, S10 go/no-go) | Sat 2026-11-14 12:00 | owner decision, 2026-11-12 |
| Feature freeze on `develop` | Sat 2026-11-14 18:00 | owner decision, 2026-11-12 |
| **Every member votes**, or the team's project is excluded | Sun 2026-11-15 18:00 to Mon 2026-11-16 23:59 | `docs/SHIP.md` section 6 |

**Scope**

- **The full P0.** These are the 43 operations with `x-priority: P0` in `api/openapi.yaml`, moderation included, and 44 once the E5 contract PR is accepted. No cut line is agreed in advance.
- **S9, grounded AI assistant.** It is P0 by an owner decision taken 2026-11-12 ("AI first"). It amends `docs/PRD.md` 5.8, where the sourced assistant is P2. Its operations are **not yet** in `api/openapi.yaml`. It may not be claimed until the owner applies PRD amendment 3.2 and a SPEC story exists (`docs/PRD.md`, "Traçabilité produit"; `docs/SPEC.md` section 1).
- **S10, live channels.** Conditional ("channels if on track"). It is built only if the Sat 12:00 checkpoint shows S0 to S5 on track. Otherwise it stays P2 and is reported as not built.
- **Product-review additions** (accepted by the owner on 2026-11-12):
  - **E1:** S9 teaching rules AI-1 to AI-5, AI-9 and AI-10 become acceptance tests plus the golden set.
  - **E2:** taxonomy additions and relabels through one new migration.
  - **Content is original** (owner decision 2026-11-12; it supersedes aggregation-first). No third-party content is ingested. Demo-day content comes from two places:
    - (a) member content created through the app by the team's real accounts (`live`);
    - (b) small deterministic demo fixtures in `db/seeds`, original text written by the team, `provenance='demo'`, visibly marked, **`local-dev` only**. This is for local development and e2e. ADR-0005 point 17 allows `local-dev` and `staging-test`; the owner chooses not to use staging seeds, so no ADR change is needed (owner correction 2026-11-12).
    - `staging-test` and `production-live` are similar real environments and **neither has demo data**. Both hold only real `live` content.
    - External material remains only as references (`docs/research/REFERENCES.md` style) and as source links on Discover entries.
    - The Learn free series is a small demo-fixture series of 3 to 4 short original lessons (US-LEARN-001).
  - **E4:** zero-migration interaction modes. An "ask about this lesson" button and read-only topic feeds.
  - **E5:** Connect as a read-only people directory. This is a **new operation**, suggested `GET /profiles`, so it is a contract change. Once its contract PR is accepted, P0 has **44 operations**. Office hours are a process note only.
  - **Region:** public positioning is "Afrique de l'Ouest d'abord, puis toute l'Afrique" (owner decision 2026-11-12). The default filter is "Afrique de l'Ouest", with a one-click "Toute l'Afrique". The region is a UI constant in `src/web`, with no migration and no column. See TODO S7.
  - **Pillars:** "Learn & Grow" and "Connect" are made prominent in navigation, home and wording only. No pillar is erased or demoted. The PRD amendment proposal is OPEN (O7).
  - See section 12 for what slips first.
- **Out of scope:**
  - the 4 operations with `x-priority: P1`: `GET` and `POST /resources/{resourceId}/reviews`, and `GET` and `PUT /learn/series/{seriesId}/lessons/{lessonId}/progress`;
  - any P1 or P2 feature, which is never claimed as done (`docs/PRD.md`, "Non-objectifs de CADEV").

**Go/no-go must be honest.**

- `docs/SHIP.md` section 7 allows `GO` only when G0 to G7 have passed.
- A check that did not run counts as not passed, unless a human records a dated exception with its duration and risk.
- No exception is possible for an exposed secret, a confirmed critical vulnerability or likely data loss (`docs/TEST.md:52`).
- Submitting the project does not make it a GO.

## 2. Starting point (2026-11-12, from the code)

S0 (API core) is **merged** (PRs #31 to #36; B0 is PR #24). The API mounts `GET /healthz`, `GET /readyz` and the generated server under `/api/v1`; operations without a handler answer 501 and are listed in `scripts/unimplemented-routes.txt`. `app.Server` opens a `pgxpool`; the platform layer provides the mailer (SMTP adapter, fake, Mailpit in `local-dev`), idempotency, ETag, signed cursors, a transaction runner, the subject port (anonymous default) and an LLM port with a Google AI Studio adapter that is not wired into the router. No business-domain module exists yet.

Secrets are encrypted in Git with SOPS and age (ADR-0009) and decrypted at runtime by `scripts/compose.sh`. CI runs the gates as parallel groups with a fan-in `quality` job (ADR-0010). **Gates:** G1 to G4 run in CI; G5, G6 and G7 are **NOT RUN** (`make test-e2e`, `make perf` and `make seed` exit with `NOT IMPLEMENTED`; there is no deploy workflow, B4). **Nothing is deployed**; `226os.tech` shows the provider's placeholder page.

## 3. What must be true on Sunday

Each item below is a demo journey from `docs/SPEC.md` section 2 or a smoke step from `docs/SHIP.md` section 4. A reference such as `US-SOLVE-001/2` is numbered by the order of scenarios in SPEC.md; SPEC.md has no scenario ids of its own.

1. **HTTPS and health.** The public URL serves HTTPS with a valid certificate, and `/healthz` and `/readyz` respond (SHIP 4.1, 4.2).
2. **Anonymous visitor** (US-DISCOVER-001/1, US-LEARN-001/1, US-IDENTITY-001/1):
   - filters APIs by country and category, and sees each entry's provenance;
   - reads an article, a tutorial and the free AI series without an account;
   - searches across questions, resources, projects and Learn;
   - follows the links from a question to its resource and its project.
3. **Account.** A member signs up, verifies their email through a real inbox, signs in and signs out (SHIP 4.3).
4. **Solve.** Alice publishes a question, Bob answers and votes, and Alice accepts one answer. Bob's own attempt to accept gets 403 and changes nothing (US-SOLVE-001/1, /2).
5. **Build.**
   - Alice publishes a project and a Go need.
   - Bob applies. A second click creates nothing.
   - Alice sees the application.
   - Charlie gets 403 if the project is visible to him, or 404 if it is private (US-BUILD-001/1, /2).
6. **Learn.** Alice publishes an article. An anonymous publish gets 401 and creates nothing (US-LEARN-001/2, /3).
7. **FR/EN.** Switching language changes the labels but never the original content, and each content item shows its language (US-LEARN-001/4, US-IDENTITY-001/2).
8. **Mobile on a throttled network.** Explore, Solve and Build stay usable, with clear loading, empty and error states (US-ACCESS-001/1).
9. **Moderation.** The owner account, promoted to moderator, hides and restores content, and each action writes an `audit_log` row (`docs/spec-kit/AUTHORIZATION_MATRIX.md`, "Moderator scope at P0").
10. **S9 assistant, subject to K21.** V3 is DECIDED (AI Studio free tier), but the EEA, Switzerland and UK rule (K21) is not handled, so the assistant stays disabled by config until the owner rules. Once enabled, it is available in search and in the ask-a-question flow. It answers only from platform content, with links to the sources it used, and it refuses when it has no source. It never invents a guarantee about an API or about whether a solution is verified (`docs/PRD.md` 5.8). The golden evaluation set has been run, and its result is recorded.
11. **Original content and E4, at `https://226os.tech` (O1):**
    - Discover, Build and Learn show only real `live` content created through the app by the team's real accounts.
    - The free series (3 to 4 short original lessons) arrives through the reviewed content script (B5) and is readable without an account.
    - Discover entries carry source links (metadata and a link, the PRD core), never copied third-party text.
    - Staging and production contain no demo data, no fake users and no fake votes.
    - From any article, "Poser une question sur cette leçon" opens the Solve form pre-filled and links the new question to the article. That covers member-written articles, series lessons, and demo-fixture articles in `local-dev`.
12. **E5 and pillars.**
    - The Connect page lists people. Anyone, signed in or not, can filter by country, language and skill. The list shows only fields that each profile allows, never an email and never a non-listable profile.
    - The "spaces" block stays labelled "bientôt". Mentoring and opportunities are shown as coming soon (PRD 5.5).
    - Navigation and home make Learn & Grow and Connect prominent. Solve, Discover and Build stay full pillars, reachable from the same navigation (section 12.4).
13. **S10, only if built.** A member posts in a project or series channel and other members see the message live. A moderator can hide it. If S10 is not built, it is presented as P2.
14. **Gates.** G1 to G4 are green on the target SHA. G5, G6 and G7 each have evidence or a dated human exception (see section 1). The logs contain no secrets (SHIP 4.5).
15. **CADEV.** The obligations in section 9 are met.

## 4. People and slices

The slices were assigned by owner decisions on 2026-11-12. The review rules come from ADR-0008, **ACCEPTED by the owner on 2026-11-12**. The detail of each slice is in [TODO.md](TODO.md).

| Slice | Content | Owner | Ops | Starts after |
|---|---|---|---:|---|
| S0 | API core: DB pool, `/readyz`, apigen mount, validation, cursors, ETag, Idempotency-Key, transactions, mailer. **DONE (PRs #31 to #36, 2026-11-12)** | `Y4NN777` | 0 | - |
| S1 | Identity: auth (8), me (2), profiles (1) | `NICE-DEV226` | 11 | S0 |
| S2 | Discover and search (demo fixtures moved to SEED) | `mibienpanjoe` (accountable); helpers `Y4NN777`, `NICE-DEV226` | 5 | S0 (the error report also needs S1) |
| S3 | Solve | `NICE-DEV226` | 10 | S1 |
| S4 | Learn, including the one-time production content script for the free series (B5) | `mibienpanjoe` (accountable); helpers `Y4NN777` (S4 and content script), `NICE-DEV226`; script run by `Y4NN777` | 5 | S0 (publishing also needs S1) |
| S5 | Build | `mibienpanjoe` (accountable); helpers `NICE-DEV226` (first, after Solve), `Y4NN777` | 9 | S1 |
| S6 | Moderation, built last | `Y4NN777` | 3 | S0, S1; targets from S2 to S5 |
| S7 | Frontend: the prototypes become React pages | `PANK4SS` | - | typed fixtures first, then each merged endpoint |
| S8 | Ops, release and CADEV submission | `Y4NN777` | - | Datacloud, DNS, SMTP |
| S9 | Grounded AI assistant (P0 by owner decision) | `Y4NN777` (contract PR, LLM port, backend); second developer `NICE-DEV226` after Solve; UI `PANK4SS` | new, TBD | S1, S2; owner verification V1 to V4 (H12) |
| S10 | Live channels (**CONDITIONAL**) | **unassigned**, conditional on Sat 14 Nov 12:00; migrations `mibienpanjoe` | new, TBD | Sat 12:00 GO, S1, S4, S5 |
| E2 | Taxonomy: +5 learn topics, +3 resource categories, 4 relabels, no deletions, one new migration | `mibienpanjoe`, the only migration writer; helpers `Y4NN777` and `NICE-DEV226` on the pgtest updates only, never on the migration | 0 | none (migration only) |
| SEED | Demo fixtures for **`local-dev` only**, for local development and e2e: `db/seeds` and `make seed` (NOT IMPLEMENTED today; refuses everywhere except `local-dev`), demo accounts as decided in C15, and a local copy of the 3 to 4 lesson series | `PANK4SS` (reviewed by `mibienpanjoe`; the `mk/50-ops.mk` change also by `Y4NN777`, area owner) | 0 | migrations 00001 to 00010 and E2; C15 DECIDED |
| E4 | "Ask about this lesson" button, read-only topic feeds | `PANK4SS` with `NICE-DEV226` (Solve owner) | 0 | S3 `createQuestion`; any article (S4, the content script, or SEED in `local-dev`); the C2 contract PR before the feeds (C2 DECIDED) |
| E5 | Connect people directory (new operation, suggested `GET /profiles`), `connect` page, office-hours process note | backend `NICE-DEV226` after Solve, if time allows (it slips early); UI `PANK4SS` | +1 (44 total once accepted) | contract PR; S1 (profiles); the `profile_visibility` migration (C14 DECIDED) |
| IA | Pillar emphasis: navigation, home and FR/EN copy promote Learn & Grow and Connect | `PANK4SS` | 0 | none (no new endpoints) |

**Reviewer pairs (owner instruction 2026-11-12).** A reviewer is never the author.

- `NICE-DEV226` and `mibienpanjoe` review each other.
- `Y4NN777` reviews `PANK4SS`.
- A backend developer reviews `Y4NN777`: `NICE-DEV226` for S0, and `mibienpanjoe` for S6 and S9, because `NICE-DEV226` co-develops S9.
- When a helper co-writes a PR, that helper cannot review it:
  - S5 PRs co-written by `NICE-DEV226` are reviewed by `Y4NN777`;
  - S4 and content-script PRs co-written by `Y4NN777` are reviewed by `NICE-DEV226`.

**Helpers** (owner amendment 2026-11-12).

- `Y4NN777` and `NICE-DEV226` are named helpers on `mibienpanjoe`'s slices: S2, S4, S5, E2 and the content script.
- A helper is named on the PR as co-worker and may write code on that slice's branch. `mibienpanjoe` stays accountable.
- **Only `mibienpanjoe` writes migrations; helpers never add one.**
- `PANK4SS` supports the whole team on demand: backend integration, typed client wiring, fixtures, and small fixes on any slice when an area owner asks. Each support request is noted on the PR as co-worker. This does not change frontend ownership.

**Capacity note (GUESS).** `mibienpanjoe` carries the most (S2, S4, S5, the E2 and `profile_visibility` migrations, the content script); `NICE-DEV226` has S1 and S3 and helps on S5 after Solve; `Y4NN777` helps on S4 and the content script. The **Fri 13 Nov 23:00 checkpoint** (S4 and S5 progress) decides who helps first. C14b adds about half a day across three owners (GUESS) and slips before the column if the team is behind. Operation counts per slice come from `api/openapi.yaml`; the slip order is in section 12.1.

## 5. Rules for every slice (R1 to R11, referenced from TODO.md)

- **R1 Test-first.** Write a failing test (RED), make it pass (GREEN), then refactor without changing behaviour. Invariants and authorization need real PostgreSQL integration tests (`docs/TEST.md`, `AGENTS.md`).
- **R2 Go coverage.** First-party Go statements must reach exactly 100.00%. The only exclusion is `src/api/internal/platform/apigen/apigen.gen.go` (`docs/quality/COVERAGE_SCOPE.md`). External services such as SMTP and the LLM sit behind ports with fakes, so tests use no network.
- **R3 Frontend coverage.** Statements, branches, functions and lines must each be strictly above 70%.
- **R4 CI before merge.** `make quality` (G1 to G4) must be green in CI before a merge into `develop`. Never weaken a gate, skip a test or change a threshold.
- **R5 Authz** (see section 10 for blockers B1 to B5).
  - Add a mounted route to `scripts/public-routes.txt` only if it is public by design.
  - Every member, owner or moderator route needs negative authz tests taken from `docs/spec-kit/AUTHORIZATION_MATRIX.md`.
  - The gate (B1, implemented in S0) fails on any mounted route that is in none of the three route lists.
- **R6 Migrations.** Only `mibienpanjoe` writes migrations, one at a time. They are forward-only and immutable after merge (ADR-0005; ADR-0008 lets the lead delegate this).
- **R7 Shared files.**
  - `api/openapi.yaml`, `go.mod`/`go.sum`, `pnpm-lock.yaml` and `mk/**` take one change per PR.
  - An OpenAPI change merges first, in its own PR, reviewed by `Y4NN777` and `PANK4SS`.
- **R8 Branches.** Use `feature/<domain>/<name>` from `develop` and open the PR into `develop`. Nobody reviews their own PR.
- **R9 Commits.**
  - Commit under your own confirmed Git identity.
  - No AI attribution: no `Co-authored-by` naming an agent and no "Generated with" footer.
  - Record any AI help in `AI_USAGE.md`.
- **R10 Data and secrets.**
  - Demo data is visibly marked (`provenance='demo'`) and never runs in `production-live`.
  - No secrets in Git, logs, fixtures, prompts or screenshots. This covers SMTP, the cursor key and the LLM API key.
- **R11 Honest status.** Never claim a P1 or P2 feature. Report a check that did not run as `NOT RUN`, never `PASS`.

Reviews (ADR-0008, accepted 2026-11-12):

- Backend PRs: the lead or the other backend developer.
- Frontend PRs: `Y4NN777`.
- The owner's own PRs: a backend developer.
- `api/openapi.yaml`: `Y4NN777` and `PANK4SS`.

## 6. Timeline (GMT, checkpoints only; owners and tasks are in TODO.md)

- **Thu 12 evening.** Checkpoint: S0 merged (DONE, PRs #31 to #36). Open: the first contract PR (`mibienpanjoe`), the E2 and `profile_visibility` migrations, S1 RED tests, the S7 shell, SEED design.
- **Fri 13, 12:00.** Checkpoint: public reads merged (`listResources`, `getResource`, `listResourceSources`, `search`; the S4 reads in review). `Y4NN777`: staging, DNS, Argon2id benchmark on the target host, the S9 ADR and contract PR.
- **Fri 13, end of day.** Checkpoint: S1 merged; V2 and K21 handled (V1 and V3 done).
- **Fri 13, 23:00.** Checkpoint: helpers. S4 and S5 progress decides who helps first (default: `NICE-DEV226` on S5, `Y4NN777` on S4 and the content script).
- **Sat 14, 12:00.** Checkpoint (progress): the owner decides any drop or exception from measured progress (merged PRs, CI status) and whether S10 is a GO. The decisions go in section 14. Morning work: S3, S4 publish and the content script, S9 port and retrieval, `make test-e2e` harness.
- **Sat 14, until 18:00.** S6 and S9 merged, S3 and S5 merged, `make perf` harness, backup and restore drill on staging, the content script rehearsed on `staging-test`, the team's real accounts create `live` content on staging, SEED finished, E4, E5 backend if time allows.
- **Sat 14, 18:00: freeze.** `develop` takes only fixes for failing gates or smoke checks (S9 and S10 included). Evening: `make quality-full` on the `develop` SHA, rollback drill, S9 golden run, G7 evidence, ADR-0007 evidence.
- **Sun 15, morning.** `Y4NN777` runs the release: production PR `develop` to `main`, backup, `migrate-up`, deploy the exact SHA, smoke, moderator promotion, content script on production (H19), **submits `https://226os.tech` around 12:00** (O1). Content window for all four right after the deploy (K38).
- **Sun 15, 17:00: hard submission limit.** **Sun 18:00 to Mon 23:59: every member votes (H10).**

## 7. Dependencies and critical path

**Critical path.** Each step waits for the one before it:

1. S0 (DONE) and the first contract PR (C1, C2, C14b);
2. S1 (identity), which blocks every member write;
3. the member writes in S3, S4 and S5 (S6 last, S9 after S1 and S2);
4. the S7 write flows on the live endpoints;
5. the freeze (Sat 18:00);
6. G5, G6 and G7 evidence;
7. production deploy (needs H6, H7 and B4);
8. submission.

**Parallel paths.**

- Datacloud (H6) leads to DNS (H7), staging, and the Argon2id benchmark that S1 needs before pinning its parameters (ADR-0003 "Evidence").
- H12 and V2 lead to S9; the S9 ADR and contract PR lead to its UI panel. K21 gates enabling S9.
- E2 and `profile_visibility` (migrations by `mibienpanjoe`) lead to SEED, E4 and E5. The first contract PR leads to S4 handlers, the E4 topic feeds and the E5 contract PR.
- S10 only after a GO at the Sat 12:00 checkpoint (S0 to S5 on track).

**What waits on S1.** Public reads need no login, but `listProjects`, `getProject`, `listContributionNeeds`, `getTrainingSeries`, `getProfile` and `search` accept an optional session (`security: [{}, cookieAuth]`), so they use the session resolver with its anonymous default (merged in S0).

## 8. Human tasks and lead time

Owner personal tasks still open (owner instruction 2026-11-12): **H6 and H7** (Datacloud, DNS), **H12** (V2 quotas and K21 EEA decision; V1 and V3 are done), **H10** (vote reminders), the **Gmail sender alias check** for `noreply@226os.tech` (H8 follow-up) and issue #46 (collaborators' age keys). H8 itself is DONE. H19 is the owner running the reviewed content script. Closed rows (H1, H3, H9, H11, H18) are in the archive.

| Id | Task | Who | Start by | Lead time |
|---|---|---|---|---|
| H2 | `PANK4SS` accepts the repository invitation (pending, `docs/team/CONTRIBUTORS.md`) | `PANK4SS` | Thu evening | minutes (GUESS) |
| H4 | **DONE 2026-11-12:** `quality` is a required status check on `develop` and `main` next to "Validate branch promotion" (strict). **Open sub-item:** a review rule for `main` once the team pushes (`main` requires 0 approvals; with enforce-admins on, one approval would block owner promotions) | `Y4NN777` | done | - |
| H5 | Rulings on the items still open in section 10 (C3 to C10, C12, C13, O4 to O6; O7 proposed; B4 scheduled). Decided so far: B1, B2, B3, B5, C1, C2, C14, C15, C19, O1 | `Y4NN777` | Thu evening | - |
| H6 | **OPEN.** Datacloud account and an **active commercial product**, SSH key, ports 80/443, storage (`docs/SHIP.md` section 3; a CADEV condition) | `Y4NN777` | Thu evening | UNVERIFIED; hours to a day (GUESS) |
| H7 | **OPEN.** DNS: `226os.tech` is held by the owner but nothing is deployed on it. When the host IP is known, add A records for the apex and the staging subdomain (PROPOSAL `staging.226os.tech`, owner to confirm). TLS (Traefik, Let's Encrypt) only after DNS resolves (`docs/SHIP.md`) | `Y4NN777` | when the host IP is known | minutes to hours (GUESS) |
| H8 | Gmail SMTP app password in the staging and production env files only. **DONE 2026-11-12:** `mailcheck` sent a real mail through `smtp.gmail.com:587` (STARTTLS); the settings are in the staging secrets file. **Follow-up OPEN:** verify the Gmail "Send mail as" alias for the branded sender `noreply@226os.tech` (owner decision 2026-11-12), otherwise Gmail rewrites `From`; repeat the test on the staging host | `Y4NN777` | Fri morning | an app password needs 2-step verification (GUESS; verify) |
| H10 | Vote reminders to all four members, plus a named follow-up person (SHIP section 6) | `Y4NN777` | Sun 17:00 | - |
| H12 | S9: AI Studio key (DONE, in `secrets/staging-test.enc.env`); V1 DONE, V3 DECIDED; **V2 and K21 OPEN** | `Y4NN777` | Fri 13 evening | - |
| H13 | Apply PRD amendment 3.2 and add a SPEC story for S9 (and S10 if GO); the owner applies these | `Y4NN777` | Fri evening | - |
| H16 | The PRD amendment for the pillars (O7): amendment 3.2. **DONE 2026-11-12:** approved by the owner | `Y4NN777` | done | - |
| H19 | Run the one-time S4 content script on `production-live` after its PR is reviewed, with confirmation and a backup first; record the run date in the TODO Done log | `Y4NN777` (runner) | Sun morning, after the production deploy | - |

## 9. CADEV obligations (`docs/SHIP.md` section 6)

- [x] The team has 2 to 5 members, a declared representative and country, an agreed split, and every member registered. DONE 2026-11-12 (H9), as reported by the owner.
- [x] The names in `docs/PRD.md` section 12 match the declared team: the PRD now states four members, as in `docs/team/CONTRIBUTORS.md` (checked 2026-11-12).
- [ ] Git URL, a **public demo URL** (`https://226os.tech`, once `production-live` is deployed; O1), an **active Datacloud product**, the repository and its documentation, and a secured demo account if needed.
- [ ] The project is original and was built during the contest. `THIRD_PARTY_NOTICES.md` and `AI_USAGE.md` are complete, dependency provenance is recorded, and no licence forbidden by the rules is used. The S9 provider's terms and SDK licence, if any, are added.
- [ ] Submitted before Sun 2026-11-15 17:00 GMT.
- [ ] Each member votes between Sun 2026-11-15 18:00 and Mon 2026-11-16 23:59 GMT.

## 10. Blockers, contradictions, open decisions

The original analysis blocks, rejected options and long reasoning are in the [archive](docs/archive/tracker/ROADMAP-archive-2026-11-12.md). Each item here keeps its status, decision and consequences.

### Blockers

**B1. The authz gate. DECIDED 2026-11-12 and implemented in S0 (PRs #31, #32).** The generated server is mounted through `HandlerWithOptions`; every mounted route must be in exactly one of `scripts/public-routes.txt`, `scripts/unimplemented-routes.txt` (501; starts with all operations and shrinks as each slice lands) or `scripts/authz-tests.txt` (route mapped to a negative authz Go test whose existence the gate checks). The gate also fails on routes or handlers built outside `httpx.NewRouter`. RED/GREEN proof exists. Consequences: O3 closed (the P1 routes stay in the unimplemented list); each slice PR moves its routes out of `unimplemented-routes.txt`; K3 is closed. Open follow-up (TODO S0): replace the 501 HTTP probe with a method-set check of the production `apigen.ServerInterface` implementation.

**B2. `/readyz` at the edge. DECIDED (owner approved, 2026-11-12): NOT routed.** `/readyz` is unauthenticated and pings the database, so a public route would be a database-state oracle. The smoke test runs it inside the container (`docker compose exec api /healthcheck /readyz`); the compose healthcheck keeps `/healthz`; the edge routes only `/api/` and `/healthz` (`deploy/traefik/dynamic.edge.yml`). Verified trap: at the edge `/readyz` falls through to the `web` catch-all and answers `200` with the SPA `index.html`, so never probe it through the public host. S8 keeps `docs/SHIP.md` 4.2 aligned (`/healthz` at the edge, `/readyz` inside the container).

**B3. Per-route timeout. DECIDED (owner approved, 2026-11-12).** The 30 s `requestTimeout` (`httpx.NewRouter`) stays the default. A route that needs longer (S9 assistant, S10 stream) is declared in ONE table in `httpx` and applied by a per-operation middleware installed through the apigen `Middlewares` slice (chi `r.Use` runs before routing); streams also extend the write deadline with `http.ResponseController`. Nothing is built until an S9 or S10 operation needs it. Done in S0: a timeout answers a problem+json 503 `SERVICE_UNAVAILABLE` with `Retry-After`.

**B4. Production has no deploy path yet. OPEN (scheduled: S8, before Sunday; K22).** `make up-production-live` runs only from an automated release job (`CI=true`, no interactive terminal, `PRODUCTION_UP_CONFIRM=production-live`), and there is no deploy workflow (`mk/50-ops.mk`, ADR-0006 "CI" keeps deploy workflows NOT IMPLEMENTED until TASK-OPS-001). S8 delivers a reviewed release workflow with a protected `production` environment, the exact `develop` SHA and image digests. The guard is not weakened.

**B5. Real lesson content has no path into production. DECIDED for CADEV (2026-11-12).** The free series (US-LEARN-001) reaches production through a **one-time, PR-reviewed, owner-run content script** in `scripts/content/`: neither a seed nor a migration; explicit confirmation like `migrate-up` and a backup first; lessons attributed to the owner's real account looked up by username, `provenance='live'`; run date recorded in the TODO Done log; an S4 task (`mibienpanjoe` writes, `Y4NN777` runs, H19); rehearsed on `staging-test`. A restricted create endpoint stays a post-CADEV option. The Discover catalogue has the same gap (no write operation, K39): the owner decides whether the script pattern covers a few original entries. Options (a) migration, (b) script, (c) endpoint: archive.

### Contradictions (open unless marked)

- **C1. Maintainer rights. DECIDED (owner approved).** Project actions belong to the **project owner only** at P0 (edit, needs, list applications, accept or decline); no maintainer delegation; the applicant may withdraw. The S5 PR corrects `AUTHORIZATION_MATRIX.md` (three rows) and the docs follow-ups in `DATABASE_DESIGN.md` sections 7 and 10; the operation `decideApplication` still carries an "owner or maintainer" x-authz that the first contract PR corrects (R7). This narrows ADR-0005 point 10; the owner records it.
- **C2. Learn `topic`. DECIDED (owner approved).** The topic filter uses the topic code (`learn_topics.code`, pattern `^[a-z][a-z0-9-]{1,29}$`), in the first contract PR (`mibienpanjoe`; `Y4NN777` and `PANK4SS` review): retype `listLearnArticles` `topic`, add `topic` to `listTrainingSeries`, add an optional `topic` to `ArticleCreate`, then `make generate`. Mapping rule: questions have no topic and `listQuestions` filters only by `tag`; a topic feed shows articles by topic code and questions whose tag equals the topic code. S4 handlers and E4 feeds wait for that PR.
- **C3. `my_progress` without a progress table.** `getTrainingSeries` promises `my_progress` with a session, but `learn_progress` is P1 (a later migration, number from `make migrate-new`). OPEN.
- **C4. Search query length.** `q` has `maxLength: 200` in the contract but ADR-0005 point 6 bounds queries at 100. OPEN.
- **C5. Error codes.** The router emits `METHOD_NOT_ALLOWED` (the `MethodNotAllowed` handler in `httpx.NewRouter`), which is not in the `ErrorCode` enum, and no code exists for a wrong `Content-Type` (ADR-0004 point 8). OPEN.
- **C6. Ownership. RESOLVED 2026-11-12.** ADR-0008 is ACCEPTED and the four-member team is stated in `docs/team/CONTRIBUTORS.md` and the PRD; migrations belong to `mibienpanjoe` as the lead's delegation (R6).
- **C7. Stale `db/` status. RESOLVED.** `docs/spec-kit/DATABASE_DESIGN.md` is already correct (B0 created the migrations).
- **C8. Branch domains.** `docs/GIT_WORKFLOW.md` lists allowed domains without `learn`; the CI regex accepts any lowercase domain (`branch-policy.yml`). OPEN (owner ruling for `learn`).
- **C9. Stale x-authz `matrix:` labels** in `api/openapi.yaml` (for example `POST /sessions`, `GET /users/me`, `PATCH /profiles/me`). OPEN.
- **C10. S9 priority.** PRD 5.8 puts the sourced assistant at P2; the owner made it P0 (H13). OPEN until amendment 3.2 is approved.
- **C12. Article-to-question link.** The database allows only the canonical pair `('question','article')` (`00009_content_relations.sql`), while ADR-0005 point 13 lets only the source author create a relation and `ArticleCreate.links[]` accepts questions. OPEN.
- **C13. `LinkRef` accepts `series`, but `content_relations` has no series column.** This is why each lesson of the demo series is a tutorial `learn_article` referenced by `training_lessons.article_id`. OPEN.
- **C14. Profile visibility and listing. DECIDED 2026-11-12 (owner approved; residual points jointly with the main session).**
  - (a) A profile is listed for a caller only if the user is active, `open_to_collab = true` and the caller may see the flag under `open_to_collab_visibility` (anonymous: `public`; members: `public` and `members`; `private` never). The predicate and every filter use caller-visible fields only (a hidden field counts as absent). Email is never listed.
  - (b) `profiles.profile_visibility text NOT NULL DEFAULT 'public'`, CHECK `public|members|private`, in a new migration written only by `mibienpanjoe` (number from `make migrate-new`; may share the E2 PR as a separate file; pgtest structure tests change in the same PR). `getProfile`: `public` to all; `members` gives anonymous 404; `private` owner only; the 404 is identical to an unknown or deleted username; moderators get no special access. Bylines on published content always show username and display name; deleted accounts show "membre supprimé" (ADR-0005 point 15). The directory requires `profile_visibility` to allow the caller **and** (a). Contract: carried by the first contract PR (TODO Shared pieces). Owners: S1 (`NICE-DEV226`) read, write and tests; S7 (`PANK4SS`) settings UI; E5 depends on the migration. Suspended users: out of scope for P0.
- **C15. Demo fixtures need authors. DECIDED.** Non-real demo accounts: `demo-` prefix, email on the reserved invalid domain (`@demo.226os.invalid`), no `user_credentials` row (cannot sign in), `provenance='demo'`, visible badge, **`local-dev` only** (owner correction B: stricter than ADR-0005 point 17, so no ADR change). No fake votes, endorsements or verifications. "No real users" means never impersonate real people.
- **C19 (merges C11). Migration `00011` is no longer reserved. DECIDED 2026-11-12.** `DATABASE_DESIGN.md` section 5 and ADR-0005 point 16 reserved `00011` for the deferred P1 tables, but `scripts/migrate-new.sh` creates the highest number plus one and `scripts/check-migrations.sh` requires a sequential sequence with no gap. Use the number `make migrate-new` gives (E2 is `00011` and `profile_visibility` `00012`, or the reverse); P1, S9 and S10 tables take the next free number. No script or gate change. Docs follow-up in the same PR as the migration: correct the reservation statements in `DATABASE_DESIGN.md` (`mibienpanjoe`, reviewed by `NICE-DEV226`); the owner records that ADR-0005 point 16 is no longer fixed.

### Open decisions

- **O1. The public demo URL. DECIDED 2026-11-12.** The CADEV URL is **`https://226os.tech`**, the `production-live` deployment, populated before submission with real `live` content created through the app by the team's real accounts. Nothing is deployed there today. `staging-test` (PROPOSAL `staging.226os.tech`, owner to confirm) has **no demo data** and is used for rehearsals. Consequences: the free series arrives through the B5 script; a content window is needed on production, which needs Gmail SMTP working there (K38); no seeds ever run in production.
- **O2. Where the shared pieces live.** Listed in TODO.md "Shared pieces": rate limiter, session resolver, `links[]` writer and reader, audit-log writer, `EMAIL_NOT_VERIFIED` check. The backend lead decides at kickoff.
- **O3. The P1 routes. CLOSED by B1** (mounted, 501, listed in `scripts/unimplemented-routes.txt`).
- **O4. Resending the verification email.** ADR-0003 sets a resend rate limit but no P0 operation resends. OPEN.
- **O5. S9 questions the PROPOSED ADR must decide.** The provider is decided (Google AI Studio, Gemma 4, section 14). Open: exactly what is sent (question text and retrieved public content only; no emails, private data or secrets), the cost cap or quota guard and where its counter lives, timeouts (B3), the fallback ("assistant unavailable"; the page never fails), FR/EN behaviour, the evaluation set and pass criteria, anonymous access and the per-subject rate limit. The second developer is `NICE-DEV226` (after Solve). Design rules proposed for the ADR: a visible notice that the question is sent to Google, redaction of secrets and emails with tests, a per-environment switch, a provider switch through the LLM port if the free tier is unsuitable.
  - **Facts from the AI Studio terms** (OFFICIAL, `https://aistudio.google.com/docs/terms`, "Effective March 23, 2026", re-fetched 2026-11-12; the S9 ADR is their long-term home): on the free tier (Unpaid Services) content is used to improve Google products, human reviewers may read input and output, and personal information must not be submitted; Paid Services do not use content for improvement and log it for a limited period; a project counts as paid only with an active Cloud Billing account; **only Paid Services may be offered to users in the EEA, Switzerland and the UK** (K21).
  - **NOT CONFIRMED:** Gemma free-tier rate limits (shown per project in the console; third-party figures conflict); further sources `https://ai.google.dev/gemini-api/docs/logs-policy` and `https://ai.google.dev/gemini-api/docs/rate-limits` were not re-fetched.
  - **Verification items (owner `Y4NN777`):** **V1 DONE 2026-11-12:** `aicheck models` lists `gemma-4-26b-a4b-it` and `gemma-4-31b-it` (both `generateContent`, 262144 input, 32768 output tokens); staging uses `gemma-4-26b-a4b-it`. **V2 OPEN:** the project's real limits are read in the console; the 3 probes of `aicheck limits` do not measure them (`gemma-4-26b-a4b-it` 3 of 3 ok, `gemma-4-31b-it` 2 of 3 unavailable). **V3 DECIDED 2026-11-12: free tier;** the EEA consequence (K21) is INFERRED and OPEN. **V4:** both models answer in French and English but put their reasoning in the answer text (the adapter must drop it); FR/EN latency and quality on the golden set are still to be measured before the demo.
- **O6. S10 builder and displacement.** Nobody owns S10. If built it takes Saturday capacity from S6 and S7 integration, the G5 and G6 harnesses, the backup drill, the bug-fix buffer and E4-type work (GUESS; not quantified).
- **O7. Prominent pillars.** Learn & Grow and Connect are made prominent in navigation, home and wording only; no PRD pillar is erased or demoted (nine domains, four differentiators stay; one ecosystem, section 12.4). PRD amendment 3.2 is **approved** (owner, 2026-11-12, H16). The amendment wording is unchanged here.
- **O9. E5 backend owner. Assigned:** `NICE-DEV226`, after Solve, if time allows (slips early, section 12.1).
- **O10. SEED owner. Assigned:** `PANK4SS` (original fixtures, `db/seeds`, `make seed`); reviewed by `mibienpanjoe`, and `Y4NN777` for `mk/50-ops.mk`. The production series is the S4 content script (B5).
- **Licences (DECIDED 2026-11-12, owner):** `isbot` (Unlicense) accepted; runtime base images (Alpine and Debian packages, the `acme-init` busybox helper and the other images; busybox was removed from the API image on 2026-11-12) are treated as deployed container images, not bundled project code (CADEV art. 6). Residual risk: the organiser has not ruled. Recorded in `THIRD_PARTY_NOTICES.md`; optional follow-ups in TODO "Open CADEV checklist".

## 11. Risk register

Closed or merged out of this table (see archive): K3 (B1 implemented in S0), K36 (merged into K12), K37 (C15 decided).

| Id | Risk | Level | Mitigation |
|---|---|---|---|
| K1 | Capacity: 19 to 21 operations per backend developer, at 100% coverage with negative authz tests, in about 2.5 days | high (GUESS) | Sat 12:00 checkpoint, small PRs, the owner can ask for help (ADR-0008) |
| K2 | Owner review bottleneck: the owner reviews frontend, S1 and much more while building S0, S6, S8 and S9 | high | Backend developers review each other (ADR-0008 "Consequences") |
| K22 | B4: no release workflow, so production cannot be started under the existing guard | high | B4 OPEN: S8 builds the workflow by Sat; rehearse it on staging first |
| K4 | Gmail SMTP daily sending limits and spam filtering | medium (GUESS) | Real inbox test on staging; Mailpit locally |
| K5 | Datacloud account or product lead time | high (GUESS) | Start Thu evening (H6) |
| K6 | Argon2id m=64 MiB (ADR-0003) against the 128M API memory limit (`compose.yaml`): two concurrent hashes may exceed it | medium (INFERRED) | Size the hashing semaphore or the limit; benchmark on the host |
| K7 | The 54 country names in `db/migrations/00002_reference_taxonomy.sql` were not checked here against an ISO source. Migrations are immutable, so a fix needs a new migration. | low | Owner checks before the demo |
| K8 | `/healthz` is not in `api/openapi.yaml`, but the web client calls it (`src/web/src/api/health.ts`) | low | Owner ruling: document it as an infrastructure route |
| K9 | The CSP has `style-src 'unsafe-inline'` (`src/web/nginx/security-headers.conf:2`) | medium | G4 review item; record as residual risk or remove |
| K10 | `createProject` fetches GitHub with anti-SSRF controls (`api/openapi.yaml`), a new outbound surface | medium | S5 threat model, owner co-review |
| K11 | The rate-limit table is UNLOGGED, so counters reset after a crash | low | Accepted (ADR-0003) |
| K12 | Thin content on the production demo URL (O1 DECIDED: production); also staging and production hold only real content written by the team (K36, merged) | high | The team creates real `live` content through the app on staging (rehearsal) and production (content window) before submission; the series arrives through the content script (H19) |
| K38 | The production content window is short: production deploys Sun morning, content must be created before the ~12:00 submission, and publishing needs a verified email, so Gmail SMTP must work in production | high (GUESS on timing) | Deploy production as early as G7 allows; rehearse the content run on staging Sat evening; test SMTP in production first |
| K39 | Discover has no write operation (`AUTHORIZATION_MATRIX.md`), so the team cannot create Discover entries "through the app" in production | high | Owner decides whether the content-script pattern (B5) also covers a few original Discover entries |
| K13 | S9 provider (Google AI Studio, Gemma 4): V1 DONE (two model ids confirmed with `aicheck`); V2 real limits are OPEN (probes only; third-party figures conflict); the free tier showed 429, 500 and 503 errors | high | Owner reads the console limits (V2); the fake adapter means work does not wait; the provider can be switched by configuration; S9 tracks the error counts |
| K21 | **EEA rule (OFFICIAL, AI Studio terms):** only Paid Services may be offered to users in the EEA, Switzerland or the UK. INFERRED: a French UI and a public demo make EEA users likely, and the free tier cannot serve them. The free tier also allows human review and product use of the data, and forbids personal information. | high | V3 is DECIDED (free tier), so EEA handling is OPEN: the owner decides a restriction or a switch for EEA, Swiss and UK users (or a billing-linked project; cost to be measured). Until then the assistant is disabled by config. |
| K14 | S9 key handling and quota or cost: key leak, quota exhausted during the demo | high | Key only in the env secret; per-subject rate limit; quota or cost guard; "assistant unavailable" degradation; no prompt logging of personal data or secrets |
| K15 | S9 answer quality: hallucination, invented API guarantees (PRD 5.8) | high | Answers must cite retrieved sources or refuse; golden evaluation set as acceptance |
| K16 | S9 prompt injection through retrieved user content | high | Retrieved content treated as untrusted data; injection cases in the golden set |
| K17 | S9 and S10 both raise moderation load: AI answers quoting bad content, chat abuse | medium | S10 messages covered by moderation hide; S9 cites only published, non-removed content |
| K18 | S9 and S10 compete with S0 to S8 for the same four people | high (GUESS; not quantified) | S10 only if the Sat 12:00 checkpoint is GO; S9 second developer only after their first slice merges |
| K19 | S10 live delivery: a `LISTEN` holds a dedicated DB connection; the 30 s router timeout (B3); Traefik streaming behaviour unverified | medium (INFERRED) | S10 ADR covers the transport, connection budget and timeouts |
| K23 | E1, E2, E4, E5, SEED and the pillar change raise the total scope with no added people | high (GUESS; not quantified) | The slip order in section 12.1, decided at the Sat 12:00 checkpoint |
| K28 | E5 privacy: a hidden field or a non-listable profile leaks through the list or the filters; scraping of the directory | high | Visibility-filtered projection and filters; email never selected; bounded `limit` (max 50, ADR-0004); a rate limit per IP and per user (window to be set by the owner, ADR-0003 has none for this route); privacy tests |
| K29 | E5 adds an operation, so `apigen` regenerates and the 47-operation count in `contract_test.go` and `docs/quality/COVERAGE_SCOPE.md:17` changes; B1 applies to the new route | medium | Handled in the E5 contract PR, with owner approval for `COVERAGE_SCOPE.md` |
| K24 | B5: no content path into production | medium (B5 DECIDED for CADEV: the content script) | Script reviewed in a PR, rehearsed on staging, run by the owner (H19) |
| K25 | The relabel of learn topic `ai` changes its meaning, from "IA appliquée et langues africaines" (`db/migrations/00002_reference_taxonomy.sql`) to "Comprendre les LLM". Content already tagged `ai` (seeds, fixtures) may need re-tagging to `african-languages`. | low (INFERRED) | Check seeds and fixtures in the E2 PR |
| K26 | E2 changes the reference lists, so `migrate_integration_test.go` (exact lists and counts at `:140-150`, `:180`, `:219`) fails unless it is updated in the same PR | medium | Same PR; owner review, because pgtest is in the platform area |
| K20 | S9 sends user questions and platform content to a third-party provider. On the free tier, the content may be used to improve Google products and may be read by human reviewers (OFFICIAL). | high | The ADR fixes what is sent: the question text and retrieved public content, never emails, private data or secrets. The UI tells the user. |

## 12. Product review findings (2026-11-12)

### 12.1 The plan is bound by capacity

E1 to E5 and the pillar change raise the total scope with the same four people (K23).

If progress is behind at the Sat 12:00 checkpoint, the suggested order of what slips first is below. It is a GUESS, and the owner decides.

1. S10 never starts.
2. The E4 topic feeds.
3. The E5 people directory. The `connect` page then stays "bientôt" and is reported as not built.
4. Extra demo fixtures beyond the minimum (`local-dev`). The minimum is the 3 to 4 lesson series plus a few Discover and Build entries.
5. Only if behind: the profile visibility setting UI (S7). The `profile_visibility` column and its API semantics stay; the column never slips before the UI.
6. The S6 moderation endpoints. The fallback is the documented SQL runbook, and they are reported as not built.

### 12.2 Failure modes (S9, SEED and E5)

| Failure | User sees | System does | Test |
|---|---|---|---|
| Provider down or quota exhausted | "Assistant unavailable"; the page still works | No retry storm; 503 problem; logged without the prompt | Fake adapter returns an error or a quota response |
| Retrieval returns no published passage (AI-1) | "No source found", with a link to search | The provider is **not called** | Assert the fake adapter receives zero calls |
| Citation not in the retrieved set (AI-2) | Only valid citations; refusal if none remain | Strips or rejects the citation; needs at least one | Fake answer cites an unknown id |
| Prompt injection in retrieved content | A normal grounded answer, or a refusal | Retrieved text treated as untrusted data | Golden-set case with a planted instruction |
| Secrets or emails in the question (AI-10) | The answer; a notice says the question is sent to Google | Redacted before sending; never logged | Assert the redacted payload at the fake adapter |
| SEED: `make seed` run outside `local-dev` | A refusal | Exits non-zero before touching the database | Refusal tests for `APP_ENV=staging-test` and `production-live` |
| SEED: demo rows reach staging or production | Never happens | A check finds zero `provenance='demo'` rows on **both** staging and production (`docs/spec-kit/DATABASE_DESIGN.md`, extended to staging) | The check in the G7 evidence, for both environments |
| SEED: a fixture is mistaken for real activity | A visible demo badge | `provenance='demo'` rendered as a badge; no fake votes or applications | UI test for the badge; a seed test that no votes or applications are seeded |
| E5: a hidden field or a non-listable profile appears in the list | Never happens | Listed only if the user is active, `profile_visibility` allows the caller, `open_to_collab = true`, and `open_to_collab_visibility` allows the caller; filters only on caller-visible fields; email never selected | Privacy tests per `profile_visibility` × caller and per `open_to_collab_visibility` × caller (anonymous, member); filtering on a hidden field does not match; no email key |
| S1: direct read of a hidden profile | The same 404 as an unknown username | `members` profile → 404 for anonymous callers; `private` → 404 for everyone but the owner; the 404 body is identical to unknown or deleted; moderators get no special access | `getProfile` tests per `profile_visibility` × caller (anonymous, member, owner, moderator); 404 bodies compared byte for byte, apart from `request_id` and `instance` |
| E5: listing abuse or scraping | 429 with `Retry-After` past the limit | Bounded `limit` (above 50 returns 422), cursor only, per-IP and per-user rate limit | Limit-51 test; N+1 rate-limit test |
| E5: empty result | A clear empty state, with a reset-filters action | 200 with `items: []` and `next_cursor: null` | Contract test and a web empty-state test |

### 12.3 Not in scope

- Hint-first mode (AI-6, P1).
- Ask-for-your-attempt (AI-7, P1).
- Study groups.
- Mentoring and mentor matching, CV review, interview preparation, and opportunities (internships, jobs, missions). These are shown honestly as coming soon; no matchmaking is announced (PRD 5.5). Office hours exist only as a process note (E5).
- Community spaces. The `connect` page "spaces" block stays "bientôt", with no fake spaces.
- Certificates.
- Progress tracking (`learn_progress`, P1).
- Taxonomy removals (`ussd-messaging`, `geo-data`, the `logistics` merge). These are deferred until after CADEV and need explicit human approval.
- Translation of content.
- Live chat, unless S10 is greenlit.
- **Third-party content ingestion and scraping**, and an `aggregated` provenance value (owner decision 2026-11-12). External material stays only as references and as source links on Discover entries.

### 12.4 Coherence spine: one ecosystem

All pillars share the same spine. That is why making two of them prominent removes nothing. The PRD calls this "Un seul écosystème connecté", sharing identity, search, tags, profiles and contextual links (`docs/PRD.md`, section 4 "Principes produit").

- **Identity everywhere.** One account and profile, used by Solve, Learn, Build and Connect (S1).
- **One search over everything.** `GET /search` covers questions, resources, projects and Learn (S2).
- **Shared tags and topics.** One `tags` namespace, plus `learn_topics` and `resource_categories` (E2). Taxonomy rows are data: they change only through a new forward-only migration or a later moderator path, never by editing old migrations.
- **Typed content relations.** `content_relations` links questions, resources, projects and articles (S2 to S5, E4).
- **The assistant** is grounded on all published content (S9).
- **The people directory** links people to their contributions (E5).
- **Provenance labels.** A demo badge on `provenance='demo'` fixtures, across every pillar. Discover entries show their documentary sources as links.

## 13. NOT RUN (as of 2026-11-12)

- **Gates.** G5 e2e, G6 perf and G7 release. `make test-e2e`, `make perf` and `make seed` are NOT IMPLEMENTED. There is no deploy or release workflow (B4). Nothing is deployed.
- **Frontend evidence.** The ADR-0007 visual, axe and responsive evidence, and frontend coverage of the new pages.
- **Before pinning parameters.** The Argon2id benchmark on the target host (ADR-0003).
- **Operations.** Staging deployment, TLS and DNS; backup and restore drills on staging and production (the local-dev drill ran); the rollback drill; the SMTP test on the staging host (the real-inbox test from a developer machine is done, H8).
- **Search.** The relevance dataset and EXPLAIN evidence (ADR-0005).
- **Reviews.** Per-surface threat models; the human and legal licence review (G4/G7).
- **S9.** V1, V3 and V4 were measured with `aicheck` on the real key (2026-11-12): V1 done, V3 decided, V2 probes only, V4 not yet on the golden set. NOT RUN: the V2 console limits, the adapter against the real API beyond `aicheck`, the FR/EN latency and quality measurement on the golden set, and the golden set itself (CI runs it only against the fake adapter).
- **S10.** Everything; it is conditional.
- **E2.** The migration and the updated pgtest lists, on fresh and upgrade runs.
- **SEED.** The fixtures and `make seed` (NOT IMPLEMENTED).
- **E4.** All of it. **E5.** The contract PR, the backend, the privacy tests and the `connect` page. **Pillar change.** The navigation, home and copy change.

## 14. Decisions log

Superseded entries (OpenAI provider, V1 to V4 findings note, E3/S11, pillars first wording, aggregation first, no-third-party entry, the retracted C14 note, and two O7 notes) are in the [archive](docs/archive/tracker/ROADMAP-archive-2026-11-12.md). Entries are never rewritten; notes say what supersedes them.

| Date | Decision | By |
|---|---|---|
| 2026-11-12 | Scope is the full P0: 43 operations, moderation included. No cut line is agreed in advance; drops are decided only at the Sat 14 12:00 checkpoint, from measured progress. Freeze Sat 14 18:00. Submission around Sun 15 12:00 GMT. | `Y4NN777` |
| 2026-11-12 | Email: Gmail SMTP for `staging-test` and `production-live`. The app password lives only in the env file, never in Git. Mailpit in `local-dev` (ADR-0003). | `Y4NN777` |
| 2026-11-12 | Idempotency-Key is implemented for the six operations that bind the header: `createQuestion`, `createAnswer`, `createProject`, `createContributionNeed`, `createApplication`, `createLearnArticle`. | `Y4NN777` |
| 2026-11-12 | Moderation stays in P0 and is built last. No fake moderator is seeded; the owner account is promoted through a documented one-time step. The emergency fallback is a documented SQL runbook. | `Y4NN777` |
| 2026-11-12 | The Datacloud account and product are handled by the owner, `Y4NN777`. | `Y4NN777` |
| 2026-11-12 | Slices S0 to S8 and their owners are as in section 4. Only `mibienpanjoe` writes migrations. Review rules are as in section 5. | `Y4NN777` |
| 2026-11-12 | **AI first, channels if on track.** S9 is P0 and amends PRD 5.8; the owner applies amendment 3.2 from the S9 ADR. S10 is built only if the Sat 12:00 checkpoint shows S0 to S5 on track; otherwise it stays P2 and is reported as such. The LLM provider and cost are open owner decisions. | `Y4NN777` |
| 2026-11-12 | **Supersedes the OpenAI entry.** S9 provider: Google AI Studio, model Gemma 4, free tier; the owner holds the access. The free tier was later clarified as development and evaluation only, with production use needing V3. The AI Studio API key lives only in the staging and production env files, never in Git, CI logs or images. The LLM port keeps the provider swappable. Development and the golden evaluation set run against the fake adapter. V1 to V6 were open verification items; **superseded by V1 to V4 (section 10, O5; V1 DONE, V3 DECIDED, V2 OPEN).** Fallback: switch provider through the port by configuration; an alternative paid key is possible, at the owner's decision. | `Y4NN777` |
| 2026-11-12 | **E1 accepted.** S9 acceptance tests and golden set:<br>- AI-1: ground or refuse. With zero retrieved passages, the provider is not called and the answer links to search.<br>- AI-2: citations are validated on the server, with at least one per answer.<br>- AI-3: no guarantees about an API or a verified solution.<br>- AI-4: every answer ends with "Comment vérifier".<br>- AI-5: uncertainty is stated, with source dates.<br>- AI-9: answers are labelled AI and can never be accepted or voted as an answer.<br>- AI-10: only the question text and public retrieved content are sent, after redaction.<br>AI-6 (hint-first) and AI-7 (ask for your attempt) are P1. | `Y4NN777` |
| 2026-11-12 | **E2 accepted.** Taxonomy changes are add and relabel only, in one new migration by `mibienpanjoe`. Its number comes from `make migrate-new`; 00011 is reserved. **Superseded by C19 (2026-11-12): 00011 is no longer reserved.**<br>- +5 learn topics: `learn-to-learn`, `ai-assisted-dev`, `ai-agents`, `open-models`, `african-languages`.<br>- Relabels: `ai`, `mobile`, `open-source`, and the category `cloud-hosting`.<br>- +3 categories: `ai-models`, `ai-tooling`, `ai-datasets`.<br>Result: 14 topics and 18 categories. `payments`, `ai`, `mobile-money` and `sms-ussd` stay valid. The pgtest lists change in the same PR; the DATABASE_DESIGN lists are a docs follow-up. Removals are deferred until after CADEV and need explicit approval. | `Y4NN777` |
| 2026-11-12 | **E4 accepted.** An "ask about this lesson" button, using the existing `links[]` on question creation, and read-only topic feeds. No new table and no new moderation system. Owner: `PANK4SS` with the Solve owner. The topic feeds wait for the C2 ruling. | `Y4NN777` |
| 2026-11-12 | **E5 accepted: Connect as a people directory.**<br>- A new read-only operation, suggested `GET /profiles`, with country, language and skill filters and cursor pagination.<br>- Public fields only, filtered by `field_visibility`. Email is never returned, and private profiles are never listed.<br>- Anonymous callers and members are both allowed.<br>- This is a contract change: the contract PR comes first (backend lead plus `PANK4SS`), with `make generate`. P0 then has 44 operations.<br>- The route goes in `scripts/public-routes.txt`.<br>- The `connect` page gets the people list. Spaces stays "bientôt".<br>- Office hours: a process note only.<br>- Mentoring, CV review, interview preparation, study groups and opportunities are not in scope and are shown as coming soon.<br>- The backend owner is OPEN.<br>C14 (no profile-level visibility column) needs a ruling first. | `Y4NN777` |
| 2026-11-12 | **Pillars, wording replaced (O7).** Learn & Grow and Connect are prominent in navigation, home and wording only. No PRD pillar is erased or demoted; all nine domains and four differentiators stay. The amendment text stays a proposal that says "prominence, not removal". The coherence spine is in section 12.4. | `Y4NN777` |
| 2026-11-12 | **CADEV team registration DONE (H9).** | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner instruction: close) B1, the authz gate.**<br>- Mount through `HandlerWithOptions`.<br>- Every mounted route must be in `scripts/public-routes.txt`, in `scripts/unimplemented-routes.txt` (starts with all 47 and shrinks), or in `scripts/authz-tests.txt` (route mapped to a negative authz Go test, whose existence is checked with `go test -list`).<br>- The gate fails on a route in none of the three, or on a handler built outside `httpx.NewRouter`.<br>- RED/GREEN proof is required.<br>**Reason:** mounts the contract without weakening the gate. **Consequence:** O3 is closed, and each slice moves its routes out of the unimplemented list. **APPROVED as written** (owner correction A). | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner instruction: close) C15, fixture authors.**<br>- Non-real demo accounts: a `demo-` prefix, an email on the reserved invalid domain, no password hash so they cannot sign in, `provenance='demo'`, a visible badge.<br>- No fake votes, endorsements or verifications.<br>- "No real users" in ADR-0005 means never impersonate real people.<br>**Reason:** already planned by ADR-0005 point 17 and `DATABASE_DESIGN.md`. **Consequence:** SEED is unblocked and goes to `PANK4SS`, reviewed by `mibienpanjoe`. | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner instruction: close) O1 and B5, the URL and the series.**<br>- The CADEV URL is `production-live` on the project domain, populated with real `live` content created by the team's real accounts through the app before submission.<br>- The free series reaches production through a one-time, PR-reviewed, owner-run content script in `scripts/content/`. It is neither a seed nor a migration. It needs explicit confirmation, attributes the lessons to the owner's real account looked up by username, uses `provenance='live'`, and its run date goes in the Done log.<br>- It is an S4 task: `mibienpanjoe` writes it and `Y4NN777` runs it.<br>- A restricted create endpoint stays a post-CADEV option.<br>**Reason:** does not depend on whether staging counts for CADEV. **Consequence:** K38 (short content window), K39 (no Discover write path). | `Y4NN777` |
| 2026-11-12 | **Assignments** (every OPEN owner replaced):<br>- `Y4NN777`: S0, S6, S8, S9 (second developer `NICE-DEV226`, after Solve), and frontend reviews.<br>- `NICE-DEV226`: S1, then S3, then the E5 backend if time allows.<br>- `mibienpanjoe`: S2, S4 (with the content script), S5, and the E2 migration as the single migration writer.<br>- `PANK4SS`: S7 (pillar navigation), E4, the E5 UI and SEED, plus on-demand support.<br>- S10 is unassigned and conditional.<br>Reviewer pairs are as in section 4. | `Y4NN777` |
| 2026-11-12 | **Helpers (owner amendment).**<br>- `Y4NN777` and `NICE-DEV226` are named helpers on `mibienpanjoe`'s S2, S4, S5, E2 and content script. They are co-workers on the PR. `mibienpanjoe` stays accountable and is still the only migration writer.<br>- Default split after Solve: `NICE-DEV226` helps on S5, `Y4NN777` helps on S4 and the content script.<br>- The Fri 13 Nov 23:00 checkpoint decides who helps first. This replaces the "`NICE-DEV226` takes Build" fallback.<br>- `PANK4SS` supports anyone on demand, as a co-worker on the PR, without changing frontend ownership. | `Y4NN777` |
| 2026-11-12 | **Owner correction B, fixtures.**<br>- `staging-test` and `production-live` are similar real environments, and neither has demo data. Only `local-dev` has demo fixtures.<br>- `make seed` refuses everywhere except `local-dev`. This is stricter than ADR-0005 point 17, which allows `local-dev` and `staging-test`; the owner chooses not to use staging seeds, so **no ADR change is needed**.<br>- Demo accounts exist only in `local-dev`. SEED serves local development and e2e.<br>- The deployment checklist checks zero `demo` rows on both staging and production.<br>- The free-series content script runs on staging and on production with the same procedure. | `Y4NN777` |
| 2026-11-12 | **Owner correction C, domain.**<br>- `226os.tech` exists and the owner holds it, but nothing is deployed on it; it shows the hosting provider's placeholder page.<br>- The CADEV URL is `https://226os.tech` once `production-live` is deployed.<br>- Staging uses a subdomain: PROPOSAL `staging.226os.tech`, owner to confirm.<br>- H7: A records for the apex and the staging subdomain once the Datacloud host IP is known. TLS follows only after DNS resolves. | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner approved) C1.**<br>- Project actions belong to the project owner only: edit, needs, list applications, accept and decline.<br>- No maintainer delegation at P0. The applicant may still withdraw.<br>- The S5 PR corrects matrix lines 40, 42 and 45.<br>**Reason:** matches OpenAPI x-authz, and no operation creates a maintainer.<br>**Consequences:**<br>- S5 authz is unblocked.<br>- the operation `decideApplication` still carries an x-authz that says "owner or maintainer" and needs a contract correction.<br>- This narrows ADR-0005 point 10, which the owner records as an amendment. | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner approved) C2.**<br>- The Learn topic filter uses the topic code (`learn_topics.code`, pattern `^[a-z][a-z0-9-]{1,29}$`).<br>- Its own contract PR: retype the `listLearnArticles` topic, add a topic to `listTrainingSeries` (none today), and add an optional `topic` to `ArticleCreate`, then `make generate`. Author `mibienpanjoe`; reviewers `Y4NN777` and `PANK4SS`. It merges before the S4 handlers and the E4 feeds.<br>**Mapping rule:** feeds show articles by topic code and questions by tag; a topic page may list questions tagged with the topic code.<br>**Consequence:** S4 and E4 are unblocked after that PR. | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner approved) C14.**<br>- A profile is listable only if its user is active and has set `open_to_collab = true`.<br>- Fields are shown per `field_visibility` for the caller. Email is never shown.<br>- Filters apply only to fields the caller can see.<br>- No column for the listing rule itself. **Superseded by C14b (below): `profiles.profile_visibility` is added through one migration.** `GET /profiles/{username}` keeps its own rule.<br>**Consequence:** the E5 contract PR and backend are unblocked. **Residual:** `open_to_collab_visibility` and the undefined `private_profile` rule (section 10, C14). | `Y4NN777` |
| 2026-11-12 | **B2, B3 and B4 are scheduled, not blocking today.** They are tasks in S0 and S8 that must close before Sunday. | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner approved) B2 and B3.**<br>- B2: `/readyz` is not routed at the edge; the smoke test runs it inside the container, and the compose healthcheck keeps `/healthz`. At the edge `/readyz` returns the SPA `index.html` with a 200, so it must never be probed through the public host.<br>- B3: the 30 s timeout stays the default; a per-route exception table plus a per-operation middleware is added by the first S9 or S10 slice that needs it. A timeout already answers a problem+json 503. | `Y4NN777` |
| 2026-11-12 | **H1 DONE.** The owner confirmed the identity records of `NICE-DEV226`, `mibienpanjoe` and `PANK4SS`; no names or emails are recorded. `PANK4SS` still needs to accept the repository invitation (H2) before pushing. | `Y4NN777` |
| 2026-11-12 | **H3 DONE. ADR-0008 ACCEPTED** by the owner. The main session edits the ADR file. | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner and main session, jointly) C14a, `open_to_collab` visibility.**<br>- A profile is listed for a caller only if `open_to_collab = true` and the caller may see that flag: anonymous callers see `public`, members see `public` and `members`, and `private` is never listed.<br>- The predicate and all filters use caller-visible fields only.<br>- Negative tests per value × caller. | `Y4NN777` + main session |
| 2026-11-12 | **DECIDED (owner and main session, jointly) C14b, profile-level visibility column.**<br>- `profiles.profile_visibility text NOT NULL DEFAULT 'public'`, CHECK `public\|members\|private`. A new migration by `mibienpanjoe` only (numbered by `make migrate-new`), which may share a PR with E2 as a separate file; the pgtest structure tests change in the same PR; the `DATABASE_DESIGN.md` line is a docs follow-up.<br>- `getProfile`: `public` is visible to all; `members` gives anonymous callers 404; `private` is visible to the owner only. The 404 is identical to an unknown or deleted username.<br>- Moderators have no special profile access at P0.<br>- Bylines on published content always show the username and display name; deleted accounts show "membre supprimé" (ADR-0005 point 15).<br>- The directory requires `profile_visibility` to allow the caller **and** C14a.<br>- First contract PR: `profile_visibility` in the profile read and update schemas, `PATCH /me/profile` accepts it, and `deny_status 404` stays with these semantics.<br>- S1 (`NICE-DEV226`) does the read, write and tests. S7 (`PANK4SS`) adds the settings UI. E5 depends on the migration.<br>**Suspended users:** out of scope for P0 (see the next entry). | `Y4NN777` + main session |
| 2026-11-12 | **Suspended users in `getProfile`: out of scope for P0.** No suspension feature exists to test. Removed from every open list. | `Y4NN777` |
| 2026-11-12 | **DECIDED: the region.**<br>- Public positioning: "Afrique de l'Ouest d'abord, puis toute l'Afrique".<br>- The region is the UN geoscheme "Western Africa", restricted to the countries in our table: BJ, BF, CV, CI, GM, GH, GN, GW, LR, ML, MR, NE, NG, SN, SL, TG. That is 16, to be verified against the UN M49 list; the owner may adjust.<br>- It is a UI constant in `src/web`, with no migration and no column.<br>- The default filter is "Afrique de l'Ouest", with a one-click "Toute l'Afrique". | `Y4NN777` |
| 2026-11-12 | **Reviewer fixes.**<br>- (a) C19: the reserved `00011` against the no-gap migration gate; task for `Y4NN777`.<br>- (b) `TestFreshMigrationToHead` hardcodes 10 migrations; added to the E2 and `profile_visibility` PR file list.<br>- (c) `mk/50-ops.mk` mentions staging seeds; the SEED task makes `make seed` refuse everywhere except `local-dev`.<br>- (d) `DATABASE_DESIGN.md` sections 7 and 10 still say "owner or maintainer"; docs follow-up in S5.<br>- (e) C14's "no column" text aligned with `profile_visibility`.<br>- (f) O7 and H16: amendment 3.2 proposed in the docs PR.<br>- (g) "Free tier" reworded: development and evaluation only; production needs V3.<br>- (h) S9 suggestions (similar questions, help phrasing a question) marked P1.<br>- (i) ADR-0008 "PROPOSED" mentions removed. | `Y4NN777` |
| 2026-11-12 | **DECIDED C19 and C11: `00011` is no longer reserved.**<br>- **Reasons:** the gate is sequential with no gap, nothing depends on the reservation, and the P1 tables do not exist yet.<br>- Use the numbers from `make migrate-new`: E2 gets `00011` and `profile_visibility` gets `00012`, or the reverse depending on merge order. Both by `mibienpanjoe`. No script or gate change.<br>- Docs follow-up in the same PR: `DATABASE_DESIGN.md` section 5 and its migration plan (`mibienpanjoe`, reviewed by `NICE-DEV226`).<br>- E2 and `profile_visibility` are no longer blocked. | `Y4NN777` |
| 2026-11-12 | **DECIDED V3: AI Studio free tier.** Staging model `gemma-4-26b-a4b-it` (V1, V2 probes and V4 measured with `aicheck` on the real key: about 6 s per answer, 3 of 3 probes ok; `gemma-4-31b-it` was slower and 2 of 3 probes unavailable). Consequences recorded as tasks in TODO S9: drop the model's reasoning text and apply one answer format; track 429, 500, 503 and timeouts. The K21 EEA rule is INFERRED to apply and is not yet handled. | `Y4NN777` |
| 2026-11-12 | **DECIDED (owner): closing the open decisions.**<br>- PRD amendment 3.2 approved (H16 done).<br>- `isbot` (Unlicense) accepted; base images treated as deployed container images, not bundled code (residual risk: no organiser ruling).<br>- Gmail alias domain: `noreply@226os.tech`.<br>- H4: `quality` made a required check on `develop` and `main`.<br>- `.claude/settings.json` `secrets/**` rules: the narrowed rules (`*.enc.env`, `*.recipients`) are reviewed and kept; no extra deny added. | `Y4NN777` |
