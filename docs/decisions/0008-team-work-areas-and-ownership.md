# ADR-0008 — Team work areas and ownership for P0

- **Status:** PROPOSED (not in force until the owner sets it to ACCEPTED)
- **Date:** 2026-11-12
- **Approver:** `Y4NN77`, repository owner (pending)
- **Scope:** who owns which work area during the rest of P0 now that the team has four members; amends the interim ownership of [ADR-0002](0002-interim-single-maintainer-ownership.md)
- **Related:** [ADR-0002](0002-interim-single-maintainer-ownership.md), `docs/GIT_WORKFLOW.md`, `docs/DEV.md` section 6, `docs/spec-kit/TASKS.md`, `.github/CODEOWNERS`, `docs/team/CONTRIBUTORS.md`

## Context

ADR-0002 made `Y4NN77` the interim owner of every role and says roles are assigned when P0 completes. The team is now four people (GitHub access: `Y4NN777` admin, `NICE-DEV226` and `mibienpanjoe` write, `PANK4SS` invited). Slice B0 (bootstrap) is merged, and the next tasks (TASK-DATA-001, TASK-SEC-001, TASK-API-001, TASK-WEB-001, TASK-OPS-001) can run in parallel only if each has one owner and shared files have one editor at a time.

## Decision (proposed)

1. **Four work areas**, each with one owner and one reviewer from a neighbouring area. Owners are filled in `docs/team/CONTRIBUTORS.md` and `.github/CODEOWNERS` when this ADR is accepted.

   | Area | Paths (owner edits) | Tasks |
   |---|---|---|
   | A. Identity and platform | `apps/api/internal/platform/**`, identity module, `scripts/**`, `mk/**`, `compose.yaml`, `deploy/**`, `.github/**` | TASK-SEC-001, API core (error mapper, validation, idempotency, ETag), TASK-OPS-001 |
   | B. Data and Discover | `db/**` (all migrations, seeds, queries, sqlc), Discover module | TASK-DATA-001, Discover slices |
   | C. Frontend | `apps/web/**`, `DESIGN.md` | TASK-WEB-001, ADR-0007 visual, accessibility and responsive evidence |
   | D. Community slices and quality | Solve, Learn, Build modules, `docs/quality/**`, e2e harness | Solve, Learn, Build slices, `make test-e2e` |

   Reviewer pairs: A with B, C with D. A reviewer is never the author.

2. **Shared files have one editor at a time.**
   - `api/openapi.yaml`: a contract change merges as its own PR before the Go and web code that uses it.
   - `db/migrations/**`: only area B adds migrations, so numbers are never assigned twice.
   - `go.mod`, `go.sum`, `pnpm-lock.yaml`, `mk/**`: one change per PR, merged quickly.

3. **Owner decisions stay with `Y4NN77`:** ADRs, new dependencies, migrations on live data, secrets, costs, remote access, production approval, and the G4 and G7 security and legal reviews. This ADR does not waive any gate or human checkpoint.

4. **Identity.** Each person records their GitHub username and approved Git identity with the maintainer before their first commit, as `CONTRIBUTING.md` requires, and commits only under their own identity. Emails are never recorded in the repository. No AI agent is an author or `Co-authored-by`; AI assistance is recorded in `AI_USAGE.md`.

5. **Branches and reviews** follow `docs/GIT_WORKFLOW.md`: `feature/<domain>/<name>` from `develop`, one slice per branch, CI `quality` green before merge into `develop`. With more than one maintainer, a PR into `develop` needs a review from the reviewer of the area it touches.

6. **ADR-0002 is amended, not repealed:** its interim rules (owner approval sufficient for `main`) stay until the owner sets a different promotion rule. This ADR replaces only the "one owner for every area" part and the "revisit at P0 completion" timing.

## Options considered

- **Keep one owner until P0 ends (ADR-0002 as is):** simple, but four people cannot work in parallel without clear ownership of shared files. Rejected for the remaining P0 work.
- **Split by PRD pillar only (Solve, Discover, Build, Learn):** matches the product but not the shared infrastructure (auth, data, web shell), which would have no owner. Rejected.
- **Four areas as above (chosen):** each shared surface has one owner, and domain slices stay with areas B and D.

## Consequences

- **Security:** area A owns authentication and authorization; area B and D review slices touching them. The independent security review of G4 and G7 is still required.
- **Data:** area B is the single writer of migrations, which are forward-only and immutable once merged.
- **Ops:** CI `quality` stays the required check for every PR. Branch protection and required reviews on `develop` and `main` must be configured by the owner in GitHub.
- **Cost:** none. Risk: area D carries the most slices and depends on area A and B delivering first.

## Evidence / verification required

- `docs/team/CONTRIBUTORS.md` and `.github/CODEOWNERS` list one owner per area, with identity records confirmed by the maintainer.
- Each slice PR names its owner and reviewer from different areas.
- Branch protection and required reviews are enabled for `develop` and `main` (owner action, not verified here).
