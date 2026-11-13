## AI-use records: one file per record, AI_USAGE.md generated

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Sonnet 5.5"
purpose: "Remove the conflict every pull request had on AI_USAGE.md (all of them appended at the end, so after any merge every other open pull request conflicted: five at once on 2026-11-13). A record is now one file in docs/ai-usage/; scripts/ai-usage.py rewrites a generated block at the end of AI_USAGE.md from those files (make ai-usage, make ai-usage-check); the date check of scripts/check-docs.py reads the record files too. The five conflicting pull requests (#99 to #102, #111) were brought up to date by a mechanical merge that kept develop's AI_USAGE.md and each branch's own new records."
repository_or_source_context: "pull requests #99, #100, #101, #102, #111 (conflicts only in AI_USAGE.md)"
files_or_sections_affected: ["scripts/ai-usage.py", "scripts/tests/ai-usage_test.sh", "scripts/check-docs.py", "scripts/tests/check-docs_test.sh", "docs/ai-usage/**", "AI_USAGE.md", "mk/60-review.mk", ".github/PULL_REQUEST_TEMPLATE.md", "AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/GIT_WORKFLOW.md", "docs/SHIP.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["scripts/tests/ai-usage_test.sh 16 passed and scripts/tests/check-docs_test.sh 20 passed, RED first (no script; the date check ignored the record files)", "make lint-scripts and python3 -I scripts/check-docs.py: see the pull request"]
open_risks: ["AI_USAGE.md is only as complete as its last build: the owner must run make ai-usage before the submission (SHIP.md says so) and make ai-usage-check must pass.", "Records already appended to AI_USAGE.md stay where they are; only new records use the files."]
```
