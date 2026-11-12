# 226os contributors and work areas

**Status:** DRAFT, pending ADR-0008 acceptance by the owner. Only GitHub logins are recorded here. Real names and emails stay with the maintainer and are never committed.

## Contributors

| GitHub login | GitHub access (2026-11-12) | Identity record confirmed by the maintainer | Work area (ADR-0008) |
|---|---|---|---|
| `Y4NN777` | admin (owner) | yes (identity `Y4NN77`) | owner of ADRs, risky decisions and releases |
| `NICE-DEV226` | write | no | to be assigned by the owner |
| `mibienpanjoe` | write | no | to be assigned by the owner |
| `PANK4SS` | invitation pending | no | to be assigned by the owner |

## Work areas

| Area | Owner | Reviewer |
|---|---|---|
| A. Identity and platform | _unassigned_ | owner of B |
| B. Data and Discover | _unassigned_ | owner of A |
| C. Frontend | _unassigned_ | owner of D |
| D. Community slices and quality | _unassigned_ | owner of C |

## Before the first commit

1. Send the maintainer your GitHub username, approved Git `user.name` and `user.email` (see `CONTRIBUTING.md`) and wait for confirmation.
2. Set them locally: `git config --local user.name "<approved name>"` and `git config --local user.email "<approved email>"`.
3. Start from `develop` and name the branch `feature|fix|refactor|test|docs|chore/<domain>/<short-name>`.
4. Run `make doctor` and `make quality` before opening a PR; CI runs the same `make quality`.
5. Never add an AI agent as author or `Co-authored-by`; record AI assistance in `AI_USAGE.md`.
