# ADR-0010 — CI runs the quality gates as parallel groups

- **Status:** ACCEPTED
- **Date:** 2026-11-12
- **Approver:** `Y4NN777`, repository owner (asked for the CI to be split into steps that run in parallel, 2026-11-12)
- **Scope:** how `.github/workflows/quality.yml` runs the gates G1 to G4 of [ADR-0006](0006-toolchain-environments-quality-commands.md)
- **Amends:** the sentence in ADR-0006 that CI runs `make quality` as one job. `make quality` is unchanged and stays the sequential local definition of done.
- **Related:** `mk/40-quality.mk`, `scripts/quality.sh`, `scripts/check-quality-evidence.sh`, `scripts/ci-scope.sh`, [ADR-0009](0009-secrets-with-sops-and-age.md)

## Context

`make quality` ran as one sequential job in CI. Measured on the runs of 2026-11-12: 434 s to 511 s, of which `lint` 133 to 199 s, `test-integration` 113 to 220 s (the script tests had been run twice, fixed in #51), `generate` 2 to 78 s (a cold Go build cache), `test-unit` 53 s. The workflow ran on pull requests only, so no cache was ever written for the base branch and every pull request started cold. The recipes also mixed unrelated domains (Go, web, SQL, OpenAPI, scripts, containers), which made the single job the critical path.

## Decision

1. **Split the gates into independent steps.** `lint` becomes `lint-go`, `lint-contract` (sqlc vet, OpenAPI, migration policy), `lint-web` and `lint-scripts`; `test-integration` becomes `test-integration-go` and `test-integration-containers`. `make lint` and `make test-integration` still run all of them, one after the other.
2. **Run them as parallel CI groups.** `mk/40-quality.mk` lists the groups (`make ci-groups`) and the steps of each (`make ci-<group>`, through the same fail-closed driver). The workflow builds its job matrix from `make ci-groups`, so a new group cannot be forgotten in CI. A test (`scripts/tests/make_test.sh`) compares `make ci-steps` with `make quality-leaf-steps`: the groups cover exactly the steps of `make quality`.
3. **One check means "everything ran".** The job `quality` downloads the summary of every group and runs `scripts/check-quality-evidence.sh`: it fails unless every expected step has a PASS for the tested commit, no summary has a FAIL or NOT RUN line, and every summary ends with its own PASS line. A failed, cancelled or missing group, a failed scope job, or a summary for another commit all fail it. The job name stays `quality`; the evidence artifacts are named `ci-evidence-<group>-<commit sha>` and are overwritten when a single job is re-run.
4. **Documentation-only changes skip the code gates, visibly.** `scripts/ci-scope.sh` reads `git diff --name-only --no-renames -z` (both sides of a move are listed, so a file moved out of the code tree into `docs/` is still `full`) and answers `docs` only when every changed path is a Markdown file at the repository root, or a document or image under `docs/` (except `docs/quality/`, which `coverage-go.sh` reads) or `prototypes/`, or `LICENSE` or `NOTICE`. Anything else, an empty list or any doubt is `full`. If a gate starts reading a path that would match, it must be added to the `full` cases. The `repo` group (secret scan, encrypted-file check, licences) always runs. Pull requests into `main`, pushes and manual runs are always `full`. The `quality` job prints NOT RUN for the skipped gates and adds a notice annotation to the check; it never claims they passed. The owner accepts that a green `quality` on a documentation-only pull request means "docs only"; promotion into `main` is always full.
5. **Caches only speed up compilation.** Go modules and the Go build cache are cached per group and on `go.sum`; the pnpm store on `pnpm-lock.yaml`. The golangci-lint cache is deliberately not restored, because it holds linter results. Tests run with `-count=1` and no coverage, test or linter result is cached, so the evidence is produced by the run itself. The workflow also runs on pushes to `develop`, which validates what was merged and writes the caches that later pull requests restore.

## Consequences

- **Speed:** the wall time becomes the slowest group plus the setup of one job, instead of the sum. The number is MEASURED on the first runs after this change and reported in the pull request that introduced it; it is not estimated here.
- **Cost:** more job-minutes in total (each group repeats the checkout and the toolchain setup), less wall time. Runner concurrency limits apply.
- **Quality:** the gates are unchanged; only where they run changes. `make quality` is still the local definition of done.
- **Not done, on purpose:** per-domain path skipping for code changes (for example, skipping the Go groups for a web-only change). The OpenAPI contract and the container checks couple the domains, and a wrong skip is a silent PASS. It can be added later with measured evidence.

## Evidence / verification required

`scripts/tests/make_test.sh` (the split loses no check, the groups equal the leaf steps, an unknown group fails), `scripts/tests/quality-evidence_test.sh` (missing step, FAIL, NOT RUN, another commit, cut-off summary, no summary), `scripts/tests/ci-scope_test.sh`, `actionlint` on the workflow, and the CI run of the pull request that introduced it.
