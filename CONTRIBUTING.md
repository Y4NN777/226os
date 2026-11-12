# Contributing to 226os

**Status:** policy in force (ADR-0001 to ADR-0010). Application work happens inside an open slice of [`TODO.md`](TODO.md); human approval is required for the checkpoints listed in [`AGENTS.md`](AGENTS.md).

Language convention: product docs (README, PRD, SPEC, DEV, RULES, TEST, SHIP, ARCHITECTURE, KICKOFF) are in French; engineering and harness docs (this file, AGENTS, CLAUDE, GIT_WORKFLOW, ADRs, secrets) are in English.

## Day one

1. **Identity.** Send the maintainer (GitHub login `Y4NN777`) this record and wait for confirmation. Emails are never written in the repository.

   ```yaml
   github_username: "<your GitHub username>"
   git_user_name: "<your approved commit name>"
   git_user_email: "<approved Git email>"
   ```

   Then set it locally, for example for the owner: `git config --local user.name "Y4NN77"` and `git config --local user.email "<approved Git email>"`. `Y4NN77` is the Git author name, `Y4NN777` is the GitHub login (used for owner, approver and reviewer fields). Historical aliases are normalized through `.mailmap`; do not introduce a new author alias without review. Never use another contributor's identity, a shared account or an agent identity.
2. **Tooling.** `make doctor` checks the toolchain; `make setup` installs Go and pnpm dependencies.
3. **Secrets key** ([ADR-0009](docs/decisions/0009-secrets-with-sops-and-age.md), [`secrets/README.md`](secrets/README.md)). Run `scripts/secrets.sh keygen`: it creates your private age key at `~/.config/sops/age/keys.txt` and prints the PUBLIC key. Keep the private key local, never commit or share it. Send only the public key to the owner (issue #46 tracks collaborator keys). The owner adds it to `secrets/<profile>.recipients` and runs `scripts/secrets.sh rotate <profile>`; until then you cannot decrypt anything.
4. **Run locally.** `make up-local-dev` (it decrypts `secrets/local-dev.enc.env` at runtime, so your key must be a recipient of that profile; if the file does not exist yet, `scripts/secrets.sh init local-dev` creates it).
5. **First change.** Pick your slice in `TODO.md`, branch from `develop`, and follow the rules below.

## Before opening a change

- Read `AGENTS.md`, `docs/PRD.md`, `docs/SPEC.md`, `docs/RULES.md`, and the applicable Spec Kit artifact.
- Identify the stable story/task ID, scope, owner, reviewer, allowed paths, acceptance examples, threat cases, and required gates.
- Branches and promotion: [`docs/GIT_WORKFLOW.md`](docs/GIT_WORKFLOW.md) owns the naming convention and the flow (work branch -> PR `develop` -> PR `main`).
- Do not add product behavior, dependencies, storage services, migrations, public claims, or architecture changes without updating the relevant contract and, where required, an ADR.

## Change discipline

- One slice per pull request. Bundle related docs and chore changes into one PR; avoid PR sprawl, because each PR costs a CI wait.
- Fill in the [pull request template](.github/PULL_REQUEST_TEMPLATE.md) and update `TODO.md` in the same PR.
- Write or update observable BDD/acceptance tests before implementation (RED, then GREEN); preserve negative authorization cases.
- Do not commit credentials, tokens, private data, generated artifacts without a documented source, or copied third-party code.
- Record evidence by target SHA. `NOT RUN` is not `PASS`; no agent may self-approve a gate or exception.
- Never add `Co-authored-by` trailers for AI systems or agents; record AI assistance in `AI_USAGE.md`.

## Reviews

Reviewer pairs and areas are in [ADR-0008](docs/decisions/0008-team-work-areas-and-ownership.md) and [`docs/team/CONTRIBUTORS.md`](docs/team/CONTRIBUTORS.md). Nobody reviews a PR they authored or co-worked on. Human review is mandatory for security, license, accessibility, performance, migration, dependency, release, and production changes.

## Approval boundaries

Human approval is required for product scope, accepted ADRs, authentication/security exceptions, migrations, new dependencies, storage services, secrets, costs, remote access, deployment, and release.

## Branch protection (current state, observed 2026-11-12)

`develop` and `main` require only the status check "Validate branch promotion". The `quality` check is not required yet, and `main` requires 0 approvals. Both are owner settings (planned, see `docs/GIT_WORKFLOW.md`). Direct pushes are prohibited; deployment jobs remain `NOT IMPLEMENTED`.
