# Contributing to 226OS

**Status:** DRAFT; human approval required before implementation

## Before opening a change

- Read `AGENTS.md`, `docs/PRD.md`, `docs/SPEC.md`, `docs/RULES.md`, and the applicable Spec Kit artifact.
- Configure the repository Git identity before committing: `git config --local user.name "Y4NN77"` and `git config --local user.email "axel.studiesmail@gmail.com"`.
- Use the exact project identity `Y4NN77 <axel.studiesmail@gmail.com>` for authored and committed project changes. Historical aliases are normalized through `.mailmap`; do not introduce a new author alias without review.
- Identify the stable story/task ID, scope, owner, reviewer, allowed paths, acceptance examples, threat cases, and required gates.
- Follow the branch and promotion rules in [`docs/GIT_WORKFLOW.md`](docs/GIT_WORKFLOW.md): work branch -> PR `develop` -> PR `staging` -> PR `main`.
- Name branches `feature|fix|hotfix|refactor|test|docs|chore/<domain>/<short-name>`; use a real domain such as `solve`, `discover`, `build`, `identity`, `access`, `platform`, `ops`, or `docs`.
- Do not add product behavior, dependencies, storage services, migrations, public claims, or architecture changes without updating the relevant contract and, where required, an ADR.

## Change discipline

- Keep pull requests small and bounded to one vertical slice or documentation decision.
- Write or update observable BDD/acceptance tests before implementation; preserve negative authorization cases.
- Do not commit credentials, tokens, private data, generated artifacts without a documented source, or copied third-party code.
- Record evidence by target SHA. `NOT RUN` is not `PASS`; no agent may self-approve a gate or exception.
- Reviewers must be distinct from the author where practical and must inspect security, license, accessibility, and performance implications relevant to the change.

## Approval boundaries

Human approval is required for product scope, accepted ADRs, authentication/security exceptions, migrations, new dependencies, storage services, secrets, costs, remote access, deployment, and release. This draft does not grant those approvals.

## Proposed workflow

`GOAL -> EXPLORE -> SPECIFY -> PLAN -> RED -> GREEN -> REFACTOR -> VERIFY -> REVIEW -> SHIP`

The Git delivery flow is:

```text
feature|fix|hotfix|refactor|test|docs|chore/<domain>/<name>
	-> PR develop (Preview)
	-> PR staging (Staging)
	-> PR main (Production)
```

Direct pushes to `develop`, `staging`, and `main` are prohibited once GitHub branch protection is configured. The current repository is stopped at the G0 human approval gate; deployment jobs remain `NOT IMPLEMENTED`.
