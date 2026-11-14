## SEED original demo fixtures and make seed

```yaml
date: "2026-11-14"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer and postgres subagents)"
purpose: "SEED #75: the fixture contract (db/seeds/README.md), the local-dev-only fail-closed make seed target, the five deterministic idempotent seed files (identity, learn series of 4 tutorials, discover, build, optional solve) and their tests."
repository_or_source_context: "branch feature/platform/demo-seeds; TODO.md SEED block; issue #75; db/migrations 00001 to 00013"
files_or_sections_affected: ["db/seeds/**", "mk/50-ops.mk (seed target)", "scripts/tests/seed_test.sh", "scripts/tests/make_test.sh", "src/api/internal/platform/pgtest/seed_integration_test.go", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original fixture text and code, fictional names, example.org URLs, no new dependency"
closure: "open"
human_reviewers: ["mibienpanjoe (SEED reviewer, pending)", "Y4NN777 (mk/50-ops.mk and demo content, pending)"]
tests_or_evidence: ["RED: seed_integration_test.go fails on a scratch seed with a live-provenance user and a vote (TestSeedProvenanceIsDemo, TestSeedHasNoFakeActivity)", "GREEN: go test -race -tags integration -run Seed ./src/api/internal/platform/pgtest/ (7 tests PASS)", "scripts/tests/seed_test.sh 10 passed", "APP_ENV=local-dev make seed: 107 rows, second run 0 rows; APP_ENV=staging-test make seed refused with exit 2"]
open_risks: ["make quality not run end to end locally: make doctor refuses pnpm 9.15.9 (pin 10.32.1)", "the zero-demo-rows check on staging-test and production-live is G7 evidence, after deployment", "path extension beyond db/seeds and mk/50-ops.mk awaits owner approval"]
```
