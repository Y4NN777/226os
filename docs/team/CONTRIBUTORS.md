# 226os contributors and work areas

**Status:** DRAFT, pending ADR-0008 acceptance by the owner. Only GitHub logins are recorded here. Real names and emails stay with the maintainer and are never committed.

## Contributors

| GitHub login | GitHub access (2026-11-12) | Identity record confirmed by the maintainer | Work area (ADR-0008) |
|---|---|---|---|
| `Y4NN777` | admin (owner) | yes (identity `Y4NN77`) | ops and platform; backend lead (also codes); ADRs, risky decisions, releases |
| `NICE-DEV226` | write | no | backend |
| `mibienpanjoe` | write | no | backend |
| `PANK4SS` | invitation pending | no | frontend (implements the existing prototypes); also backend integration and support |

## Work areas

| Area | People | Reviewed by |
|---|---|---|
| Ops and platform | `Y4NN777` | a backend developer |
| Backend | `NICE-DEV226`, `mibienpanjoe`, lead `Y4NN777` | the lead or the other backend developer |
| Frontend | `PANK4SS` | the owner |

The split of backend sub-areas between the two backend developers is decided per slice by the backend lead and written on the PR. The owner may ask anyone for temporary help in any area; the helper is named on the PR and the area owner stays accountable.

## Before the first commit

1. Send the maintainer your GitHub username, approved Git `user.name` and `user.email` (see `CONTRIBUTING.md`) and wait for confirmation.
2. Set them locally: `git config --local user.name "<approved name>"` and `git config --local user.email "<approved email>"`.
3. Start from `develop` and name the branch `feature|fix|refactor|test|docs|chore/<domain>/<short-name>`.
4. Run `make doctor` and `make quality` before opening a PR; CI runs the same `make quality`.
5. Never add an AI agent as author or `Co-authored-by`; record AI assistance in `AI_USAGE.md`.
