# ADR-0006 — Toolchain, environments, and quality commands

- **Status:** ACCEPTED
- **Date:** 2026-11-11
- **Approver:** `Y4NN777`
- **Closes:** ADR-OPEN-006 (cache and object storage stay excluded) and ADR-OPEN-007 together with the hosting decision recorded in ADR-0001 (Datacloud)
- **Refines:** G0-OPS-001, G0-OPS-002
- **Related:** `docs/DEV.md`, `docs/TEST.md`, `docs/SHIP.md`, `docs/GIT_WORKFLOW.md`, `docs/spec-kit/QUALITY_RELEASE.md`, [ADR-0007](0007-astryx-version-pinning-and-verification.md)
- **Amended by:** [ADR-0009](0009-secrets-with-sops-and-age.md) (secrets in Git, encrypted) and [ADR-0010](0010-ci-parallel-gate-groups.md) (CI as parallel groups)
- **Metadata note 2026-11-12:** approver spelling normalised to the GitHub login (`Y4NN777`); the decision text is unchanged.

## Context

No Makefile, workflow, Compose file or lockfile exists yet. G1 to G6 need one reproducible interface, and `AGENTS.md` allows only reviewed Make targets for execution. The environments must map onto the branch model (`develop` is staging/preview, `main` is production).

## Options considered

- **Package manager:** pnpm (chosen), npm, yarn. pnpm gives a strict, reproducible lockfile and fast CI installs.
- **Environment model:** Compose profiles in one tree (chosen) vs fully separate Compose files.
- **Quality entrypoint:** Make (chosen) vs task runners vs CI YAML only. Make is mandated by `AGENTS.md`.
- **Cache / object storage:** Valkey or S3 (rejected for P0) vs PostgreSQL-only (chosen).

## Decision

**Toolchain.** **Go = latest stable release at bootstrap; Node = the current Active LTS at bootstrap**; no other candidates. pnpm is pinned through the `packageManager` field (corepack). Versions are pinned in `go.mod` (`go` and `toolchain` lines), `.nvmrc` and `engines`. **The bootstrap PR records the exact versions.** `pnpm-lock.yaml` is committed and CI uses `pnpm install --frozen-lockfile`. Vite builds the web app. The Astryx pin is decided in ADR-0007.

**Containers.** Docker multi-stage images for `api` (static Go binary, minimal base, non-root user, read-only root filesystem where possible) and `web` (static assets on an unprivileged web server). Base and final images are pinned by digest; healthchecks, resource limits, structured logs without secrets. Registry **GHCR**, immutable tags by commit SHA, deployment by digest. **Traefik** is the only public edge; PostgreSQL stays on a private network, never published. **Valkey and S3 are excluded at P0** (a new ADR is needed to adopt either).

**Environments.** Compose profiles **`local-dev`, `staging-test`, `production-live`** correspond to **the local machine, the `develop` branch, and the `main` branch**, in **one profile-based Compose tree**. **Datacloud is the host for `staging-test` and `production-live`.** Actual deployment still needs Datacloud account/product, DNS, budget and approver evidence; that is G7 execution evidence, not an open decision.

| Compose profile | Meaning | Git | Host |
|---|---|---|---|
| `local-dev` | developer machine | any branch | local Docker |
| `staging-test` | shared staging/preview of the `develop` SHA | `develop` | Datacloud |
| `production-live` | production | `main` | Datacloud |

**This satisfies the intent of `docs/SHIP.md` "Compose local et prod séparés"**: local and production are separate profiles with separate env files, hardened service definitions and no shared secrets; they are not separate file trees. `docs/SHIP.md` is updated accordingly. A profile selects which services run; it is not a safety boundary. A required `APP_ENV` value must match the profile, and destructive or live-affecting targets check it and refuse otherwise. Per-environment config comes from `.env.example` plus secrets stored outside Git; secrets are never in images or Compose files. (Amended by [ADR-0009](0009-secrets-with-sops-and-age.md): encrypted secrets may be committed to Git, plaintext never.)

**Makefile as the single entrypoint.** Recipes use `SHELL := bash`, `.SHELLFLAGS := -euo pipefail -c`, `.DELETE_ON_ERROR`, `.PHONY`. Fail-closed rules: a missing tool or unimplemented target exits non-zero with `NOT IMPLEMENTED`/`NOT RUN`, never 0; no `|| true`, no `-` prefixed lines, no echo of PASS without running the check; versions pinned; no network downloads at gate time beyond locked dependency fetch.

| Target | Does | Gate |
|---|---|---|
| `doctor` | verifies pinned tool versions, Docker, env; prerequisite of the others | all |
| `setup` | installs from lockfiles only (`go mod download`, `pnpm install --frozen-lockfile`); needs execution approval | prerequisite |
| `generate` | oapi-codegen (ogen fallback, ADR-0004), openapi-typescript, sqlc; then fails on `git diff --exit-code` | G1 |
| `lint` | gofmt, vet, staticcheck, golangci-lint, ESLint, Prettier check, OpenAPI lint, migration checks (no `Down`, immutability) | G1 |
| `typecheck` | `tsc --noEmit` strict; Go build | G1 |
| `test-unit` | `go test -race`, Vitest | G2 |
| `coverage` | Go first-party statements exactly 100.00%; frontend statements/branches/functions/lines each above 70%; fails if `docs/quality/COVERAGE_SCOPE.md` is absent | G2 |
| `test-integration` | ephemeral real PostgreSQL 18, fresh and upgrade migrations, HTTP and OpenAPI contract, permission and invariant tests | G3 |
| `security` | negative authz suite, govulncheck, JS dependency audit, secret scan, SAST, image scan | G4 |
| `licenses` | runtime dependency licence inventory against the allowed list | G4/G7 |
| `test-e2e` | Playwright P0 flows, FR/EN, mobile, axe | G5 |
| `perf` | reproducible dataset, EXPLAIN plans, API latency, page weight | G6 |
| `quality` | aggregate of **G1 to G4** (what a PR must prove) | G1 to G4 |
| `quality-full` | `quality` plus `test-e2e` and `perf` (**G5 and G6**) for the `develop` SHA before promotion | G1 to G6 |
| `build` | multi-stage images tagged with the SHA, digests recorded | G7 input |
| `up-<profile>` / `down-<profile>` | start/stop a profile; `down` never removes volumes | ops |
| `migrate-up`, `migrate-status`, `migrate-new` | goose forward-only; `migrate-up` outside `local-dev` requires explicit environment confirmation and a recorded backup | G3/G7 |
| `seed` | demo seeds, `local-dev` and `staging-test` only | local/staging |
| `db-reset` | drops and recreates the DB; runs only when `APP_ENV=local-dev` and the DB host is the local Compose service, otherwise exits non-zero | local only |

Both `make quality` and `make quality-full` are kept. G7 (SHA, image digest, staging smoke, backup/restore, rollback, human approval) is a runbook and release-workflow concern, not part of `quality`. `up-production-live` is never run from a developer machine.

**CI.** `.github/workflows/quality.yml` runs **`make quality`** on pull requests into `develop` and `main`; `permissions: contents: read`, actions pinned by SHA, no secrets for forked PRs, concurrency cancel, evidence (logs, coverage, reports) uploaded tied to the commit SHA. Its job name is the required status check. It has no deploy step. (Amended by [ADR-0010](0010-ci-parallel-gate-groups.md): CI runs the gates as parallel groups under the same required check; `make quality` stays the local definition of done.) Deploy workflows for `staging-test` and `production-live` remain `NOT IMPLEMENTED` until TASK-OPS-001; production uses the protected `production` environment and an artifact built from the validated `develop` SHA.

## Consequences

- **Security:** pinned, non-root, digest-addressed images and a minimal-permission CI reduce supply-chain risk; `db-reset` and `down` guards prevent data loss. Make targets are not safe by default, so each recipe is reviewed before first execution.
- **Data:** only `local-dev` may be reset; staging and production get forward migrations and backups only.
- **Ops:** Datacloud hosts both deployed profiles; a release needs the G7 evidence (account/product, DNS/TLS, budget, backup/restore, approver). Until then G7 stays `NOT RUN` or `BLOCKED`.
- **Cost:** GHCR and GitHub Actions on the existing plan; hosting cost is part of the Datacloud budget evidence.

## Evidence / verification required

`make doctor` and each target reviewed and run once in `local-dev` with exit codes; a deliberately failing check proves `make quality` fails; `db-reset` refusal test outside `local-dev`; a CI run on a PR with the SHA; image digests recorded; non-root verification (`docker inspect`); frozen lockfile install; exact Go, Node and pnpm versions recorded in the bootstrap PR. Gates G1 to G6; G7 later.
