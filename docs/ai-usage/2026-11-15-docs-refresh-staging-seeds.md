## Docs refresh, shorter README, demo seeds on staging-test

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session) with project subagents: lead (read-only audit, then doc edits), ci-implementer (make seed), go-implementer (directory filter)"
purpose: "The owner asked for a shorter, more professional README and for the documentation to be brought up to date. A read-only audit listed the stale facts; the owner then decided the updates. README.md was rewritten (standalone, French); every other doc got fact updates only (merged PRs up to #229, slice statuses, live GitHub settings read with gh, broken anchor, ADR range). Owner decision: demo seeds now run in local-dev and staging-test (reverses correction B, within ADR-0005 point 17), never in production-live; make seed accepts staging-test; the Connect directory hides provenance='demo' accounts outside local-dev. ADR-0014 records the review rule the pipeline already applies (auto-merge into develop; the owner steps in on CRITICAL, open HIGH, sensitive and pipeline paths), written from the review scripts' docstrings, and AGENTS.md, DEV.md and GIT_WORKFLOW.md point to it."
repository_or_source_context: "branch chore/ops/staging-seeds-and-docs-refresh from develop c9e4fe7"
files_or_sections_affected: ["README.md", "ROADMAP.md", "TODO.md", "CONTRIBUTING.md", "DESIGN.md", "docs/**", "prototypes/html/DISCLAIMER.md", "db/seeds/**", "mk/50-ops.mk", "scripts/tests/seed_test.sh", "db/queries/identity.sql", "src/api/internal/platform/dbgen/**", "src/api/internal/identity/**", "src/api/internal/app/**", "src/api/internal/platform/pgtest/directory_integration_test.go", "docs/decisions/0014-ai-review-merges-develop.md", "AGENTS.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original text and code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #237 and its follow-up #241)"]
tests_or_evidence: ["scripts/tests/seed_test.sh: RED 15/18 before the mk change, GREEN 18/18 after", "go test identity, app, dbgen: PASS; 100.0% statements on the touched packages (agent run); make lint-go PASS; sqlc vet PASS; sqlc regenerate shows no drift", "TestDirectoryHidesDemoAccountsByDefault on real PostgreSQL: FAIL with the filter removed (demo-person listed), PASS with it; go test -race -tags integration -run TestDirectory PASS", "TestSeedReusesLiveTagsBySlug: FAIL before (tags_slug_uq 23505 on a live go tag), PASS after the tag links resolve by slug; all TestSeed* and seed_test.sh PASS"]
open_risks: ["full make test-integration not run by hand (memory tight); the commit gate runs the configured checks", "C12 is left for the owner"]
```
