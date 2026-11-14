## Review batch 2026-11-14 fixes (issue #157)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent of the owner's session)"
purpose: "Triage of the 15 items of review batch #157: each source read against the code and docs/quality/REVIEW_CONTEXT.md, then fixed (RED test first), answered as not a defect or already decided with a file:line, or left for an owner decision. Fixes: a resource's related links showed private projects (#118, live since S5 writes resource -> project links), and the display-name minimum aligned to the contract (#146, owner decision). #156 is left to its fix by NICE-DEV226 (#160)."
repository_or_source_context: "branch fix/review/batch-2026-11-14; issues #157, #156, #118, #117 and the PR notes issues #135, #138, #140, #142, #143, #145, #146, #148"
files_or_sections_affected: ["db/queries/solve.sql", "db/queries/discover.sql", "src/api/internal/platform/dbgen/solve.sql.go (generated)", "src/api/internal/platform/dbgen/discover.sql.go (generated)", "src/api/internal/platform/pgtest/solve_integration_test.go", "src/api/internal/platform/pgtest/catalogue_read_integration_test.go", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: TestSolveEditKeepsALinkToAHiddenTag failed on the old query (tags after the restore [go], want [go tag-cache])", "RED: TestCatalogueRelatedShowsPublicProjectsOnly failed on the old query (private project shown on the public resource)", "GREEN: both pass with -race -tags integration after the query changes"]
open_risks: ["the other batch items are decisions or owner questions, not code changes; humans tick or close each source"]
```
