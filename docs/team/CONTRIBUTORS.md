# 226os contributors and work areas

**Status:** ACCEPTED on 2026-11-12, with ADR-0008 accepted by the owner. Only GitHub logins are recorded here. Real names and emails stay with the maintainer and are never committed.

## Contributors

| GitHub login | GitHub access (2026-11-12) | Identity record confirmed by the maintainer | Work area (ADR-0008) |
|---|---|---|---|
| `Y4NN777` | admin (owner) | yes (identity `Y4NN77`) | ops and platform; backend lead (also codes); ADRs, risky decisions, releases. Slices: S0, S6, S8, S9; helper on S2, S4 (and the content script, which the owner runs), S5 and E2; reviews frontend PRs |
| `NICE-DEV226` | write | yes (confirmed by the owner 2026-11-12) | backend. Slices: S1, then S3, then the E5 backend if time allows; S9 second developer; helper on S2, S4, S5 and E2 (default after Solve: S5); reviews `mibienpanjoe` and S0 |
| `mibienpanjoe` | write | yes (confirmed by the owner 2026-11-12) | backend, and the single migration writer. Slices: S2, S4 (including the content script and the first contract PR), S5, the E2 migration and the `profile_visibility` migration (accountable even when helped); reviews `NICE-DEV226`, SEED, S6 and S9 |
| `PANK4SS` | invitation pending (must accept before pushing) | yes (confirmed by the owner 2026-11-12) | frontend (implements the existing prototypes). Slices: S7 (shell, i18n, pillar navigation, all pages), E4 (with the Solve owner), the E5 UI, and SEED (`local-dev` fixtures only); supports any slice on demand (backend integration, typed client wiring, fixtures, small fixes), noted as co-worker on the PR |

## Work areas

| Area | People | Reviewed by |
|---|---|---|
| Ops and platform | `Y4NN777` | a backend developer |
| Backend | `NICE-DEV226`, `mibienpanjoe`, lead `Y4NN777` | the lead or the other backend developer |
| Frontend | `PANK4SS` | the owner |

The split of backend sub-areas between the two backend developers is decided per slice by the backend lead and written on the PR. The owner may ask anyone for temporary help in any area; the helper is named on the PR and the area owner stays accountable.

**Slice assignments, reviewer pairs and helpers (owner instruction, 2026-11-12).** The live list is in `TODO.md` and `ROADMAP.md` section 4.

- **Reviewer pairs.**
  - `NICE-DEV226` and `mibienpanjoe` review each other.
  - The owner reviews `PANK4SS`.
  - A backend developer reviews the owner: `NICE-DEV226` for S0, and `mibienpanjoe` for S6 and S9.
  - Nobody reviews a PR they authored or co-worked on.
- **Helpers on `mibienpanjoe`'s slices** (S2, S4, S5, E2, the content script): `Y4NN777` and `NICE-DEV226`.
  - A helper is named as co-worker on the PR and may write code on that branch.
  - `mibienpanjoe` stays accountable and is the only person who adds migrations.
- S10 (live channels) is unassigned and conditional on the Sat 14 Nov 12:00 checkpoint.

## Before the first commit

1. Send the maintainer your GitHub username, approved Git `user.name` and `user.email` (see `CONTRIBUTING.md`) and wait for confirmation.
2. Set them locally: `git config --local user.name "<approved name>"` and `git config --local user.email "<approved email>"`.
3. Start from `develop` and name the branch `feature|fix|refactor|test|docs|chore/<domain>/<short-name>`.
4. Run `make doctor` and `make quality` before opening a PR; CI runs the same `make quality`.
5. Never add an AI agent as author or `Co-authored-by`; record AI assistance in `AI_USAGE.md`.
