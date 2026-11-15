## Docs refresh, shorter README, demo seeds on staging-test

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session) with project subagents: lead (read-only audit, then doc edits), ci-implementer (make seed), go-implementer (directory filter)"
purpose: "The owner asked for a shorter, more professional README and for the documentation to be brought up to date. A read-only audit listed the stale facts; the owner then decided the updates. README.md was rewritten (standalone, French); every other doc got fact updates only (merged PRs up to #229, slice statuses, live GitHub settings read with gh, broken anchor, ADR range). Owner decision: demo seeds now run in local-dev and staging-test (reverses correction B, within ADR-0005 point 17), never in production-live; make seed accepts staging-test; the Connect directory hides provenance='demo' accounts outside local-dev."
repository_or_source_context: "branch chore/ops/staging-seeds-and-docs-refresh from develop c9e4fe7"
files_or_sections_affected: ["README.md", "ROADMAP.md", "TODO.md", "CONTRIBUTING.md", "DESIGN.md", "docs/**", "prototypes/html/DISCLAIMER.md", "db/seeds/**", "mk/50-ops.mk", "scripts/tests/seed_test.sh", "db/queries/identity.sql", "src/api/internal/platform/dbgen/**", "src/api/internal/identity/**", "src/api/internal/app/**", "src/api/internal/platform/pgtest/directory_integration_test.go"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original text and code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["scripts/tests/seed_test.sh: RED 15/18 before the mk change, GREEN 18/18 after", "go test identity, app, dbgen: PASS; 100.0% statements on the touched packages (agent run); make lint-go PASS; sqlc vet PASS; sqlc regenerate shows no drift", "TestDirectoryHidesDemoAccountsByDefault on real PostgreSQL: FAIL with the filter removed (demo-person listed), PASS with it; go test -race -tags integration -run TestDirectory PASS"]
open_risks: ["full make test-integration not run by hand (memory tight); the commit gate runs the configured checks", "seeding staging-test aborts the whole batch if a seeded tag or username already exists there (db/seeds/README.md section 12)", "the merge-rule contradiction (OWNER_PREFERENCES vs ADR-0008) and C12 are left for the owner"]
```
