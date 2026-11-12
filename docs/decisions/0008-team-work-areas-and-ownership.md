# ADR-0008 — Team work areas and ownership for P0

- **Status:** ACCEPTED
- **Date:** 2026-11-12
- **Approver:** `Y4NN777`, repository owner (accepted 2026-11-12; slice assignments, helpers and reviewer pairs are kept in `ROADMAP.md` section 4 and `TODO.md`)
- **Scope:** who owns which work area during the rest of P0 now that the team has four members; amends the interim ownership of [ADR-0002](0002-interim-single-maintainer-ownership.md)
- **Related:** [ADR-0002](0002-interim-single-maintainer-ownership.md), `docs/GIT_WORKFLOW.md`, `docs/DEV.md` section 6, `TODO.md`, `.github/CODEOWNERS`, `docs/team/CONTRIBUTORS.md`
- **Metadata note 2026-11-12:** approver spelling normalised to the GitHub login (`Y4NN777`); the decision text is unchanged.

## Context

ADR-0002 made `Y4NN777` the interim owner of every role and says roles are assigned when P0 completes. The team is now four people (GitHub access: `Y4NN777` admin, `NICE-DEV226` and `mibienpanjoe` write, `PANK4SS` invited). Slice B0 (bootstrap) is merged, and the next tasks (TASK-DATA-001, TASK-SEC-001, TASK-API-001, TASK-WEB-001, TASK-OPS-001) can run in parallel only if each has one owner and shared files have one editor at a time.

## Decision

1. **Three work areas, owner decisions on 2026-11-12.** Names are GitHub logins and are listed in `docs/team/CONTRIBUTORS.md` and `.github/CODEOWNERS`.

   | Area | People | Paths | Tasks |
   |---|---|---|---|
   | Ops and platform | `Y4NN777` (owner) | `scripts/**`, `mk/**`, `Makefile`, `compose.yaml`, `deploy/**`, `.github/**`, `src/api/internal/platform/**` | TASK-OPS-001, CI, Compose, release evidence |
   | Backend (lead: `Y4NN777`, who also codes) | `NICE-DEV226`, `mibienpanjoe`, with the lead; `PANK4SS` helps on backend integration and support | `src/api/**` (except platform), `db/**`, `api/openapi.yaml` | TASK-SEC-001, TASK-DATA-001, TASK-API-001: identity and API core, data and Discover, then the Solve, Learn and Build slices |
   | Frontend (product side, with the owner) | `PANK4SS` | `src/web/**`, `DESIGN.md` | TASK-WEB-001: implement the existing HTML prototypes (`prototypes/html`) in the web app, plus the ADR-0007 visual, accessibility and responsive evidence |

   `PANK4SS` is given the frontend because of a strong UX eye and the prototypes that already exist; the same person also supports backend integration (typed API client, wiring screens to endpoints) and general support. That does not give write ownership of backend paths.

   The split of the backend sub-areas (identity and API core, data and Discover, community slices) between the two backend developers is decided by the backend lead per slice and is recorded as the slice owner on each PR. Nothing here fixes it in advance.

   **Reviews.** A PR is never reviewed by its author. Backend PRs are reviewed by the lead or the other backend developer. Frontend PRs are reviewed by the owner. The owner's own PRs (ops and platform) are reviewed by a backend developer. A change to `api/openapi.yaml` needs the backend lead and the frontend owner, since both sides consume it.

   **Help from anyone.** The owner may ask any contributor for temporary help in any area. The helper is named in the PR as co-worker; the area owner stays accountable and the PR is reviewed as above. This does not change CODEOWNERS.

2. **Shared files have one editor at a time.**
   - `api/openapi.yaml`: a contract change merges as its own PR before the Go and web code that uses it.
   - `db/migrations/**`: only the backend area adds migrations, one developer at a time, so numbers are never assigned twice.
   - `go.mod`, `go.sum`, `pnpm-lock.yaml`, `mk/**`: one change per PR, merged quickly.

3. **Owner decisions stay with `Y4NN777`:** ADRs, new dependencies, migrations on live data, secrets, costs, remote access, production approval, and the G4 and G7 security and legal reviews. This ADR does not waive any gate or human checkpoint.

4. **Identity.** Each person records their GitHub username and approved Git identity with the maintainer before their first commit, as `CONTRIBUTING.md` requires, and commits only under their own identity. Emails are never recorded in the repository. No AI agent is an author or `Co-authored-by`; AI assistance is recorded in `AI_USAGE.md`.

5. **Branches and reviews** follow `docs/GIT_WORKFLOW.md`: `feature/<domain>/<name>` from `develop`, one slice per branch, CI `quality` green before merge into `develop`. With more than one maintainer, a PR into `develop` needs a review from the reviewer of the area it touches.

6. **ADR-0002 is amended, not repealed:** its interim rules (owner approval sufficient for `main`) stay until the owner sets a different promotion rule. This ADR replaces only the "one owner for every area" part and the "revisit at P0 completion" timing.

## Options considered

- **Keep one owner until P0 ends (ADR-0002 as is):** simple, but four people cannot work in parallel without clear ownership of shared files. Rejected for the remaining P0 work.
- **Split by PRD pillar only (Solve, Discover, Build, Learn):** matches the product but not the shared infrastructure (auth, data, web shell), which would have no owner. Rejected.
- **Three areas as above (chosen):** ops and platform, backend, frontend. Each shared surface has one owner, and the backend lead assigns backend slices per PR.

## Consequences

- **Security:** the backend lead owns authentication and authorization design and reviews slices touching them. The independent security review of G4 and G7 is still required.
- **Data:** the backend lead is the single writer of migrations (the lead may delegate a migration to one developer at a time), which are forward-only and immutable once merged.
- **Ops:** CI `quality` stays the required check for every PR. Branch protection and required reviews on `develop` and `main` must be configured by the owner in GitHub.
- **Cost:** none. Risk: the owner is lead of backend and ops at once, so review capacity is the bottleneck; the other backend developer reviews to relieve it.

## Evidence / verification required

- `docs/team/CONTRIBUTORS.md` and `.github/CODEOWNERS` list the people per area, with identity records confirmed by the maintainer.
- Each slice PR names its owner and reviewer from different areas.
- Branch protection and required reviews are enabled for `develop` and `main` (owner action, not verified here).
