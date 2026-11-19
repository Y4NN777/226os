## Review fix of batch 385: committers seed wording in the tracker

```yaml
date: "2026-11-19"
contributor: "Y4NN777"
tool_or_model: "Claude Code in CI (review-batch), Claude Opus 5.5 orchestrator"
purpose: "Work the 12 items of review batch 385. One fix, documentation only: the 2026-11-18 owner-notes line of TODO.md and the open_risks of its AI-use record said the committers seed sets open_to_collab false and leaves the local-dev directory, while the seed, its generator, the seed README and the pgtest keep open_to_collab true (notes seed-collab-doc-drift and seed-directory-claim-contradiction). The other items were settled without a change: review pipeline, workflow, gate script and seed README paths held by the owner, a REVIEW_CONTEXT decision, and notes that are not defects."
repository_or_source_context: "branch fix/review/batch-2026-11-19 from develop b980515, batch issue 385"
files_or_sections_affected: ["TODO.md", "docs/ai-usage/2026-11-18-owner-notes.md", "docs/ai-usage/2026-11-19-review-fix-batch.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["RED: TODO.md:74 said open_to_collab false and the directory no longer lists the committers, against db/seeds/local-dev/00001_connect_committers.sql:169 (open_to_collab true) and TestLocalDevSeedConnectCommitters (listable == n)", "GREEN: the TODO.md line and the record's open_risks say intent NULL and open_to_collab true; grep finds no remaining claim that the directory no longer lists them", "documentation only: no test applies; make test-unit, make coverage, make test-integration-go, make security NOT RUN (no code changed)", "bash scripts/check-todo-updated.sh: see the report"]
open_risks: []
```
