## Review batch #174 and the review pile fixed

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead with go-implementer, postgres-expert and web-implementer (Sonnet)"
purpose: "Fix or decide every item of review batch #174 and of the open notes issues outside it (owner directive: everything fixed or decided with a reason), including the owner-approved contract change viewer_vote and viewer_application."
repository_or_source_context: "branch fix/review/batch-2026-11-14-b; batch #174; issues #117, #118, #151, #155, #158, #162, #165, #167, #171, #175, #181, #182"
files_or_sections_affected: ["api/openapi.yaml and generated code", "db/queries/discover.sql, solve.sql, build.sql", "src/api/internal/{search,discover,solve,build,assistant,platform}/**", "src/web/src/features/{solve,build,search,learn,settings,discover}/**", "src/web/src/test/**", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency added"
closure: "open"
human_reviewers: ["Y4NN777 (owner)"]
tests_or_evidence: ["every FIXED item has a test that failed first (RED) then passed (GREEN); names in the pull request body", "make generate PASS (no drift)", "Go coverage 100.0% (6295/6295)", "make test-integration-go PASS", "web test 450 passed", "make lint-scripts: re-run after this tracker update"]
open_risks: ["the question ETag is caller-scoped: a question loaded signed out then edited after sign-in is refetched (sign-in clears the cache)", "a hidden tag slug answers 422, which reveals that a hidden tag exists under that slug"]
```
