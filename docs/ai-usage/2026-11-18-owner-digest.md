## Review batch: owner-decision digest

```yaml
date: "2026-11-18"
contributor: "Y4NN777"
tool_or_model: "Claude Code (Claude Opus 5.5), review-steward agent"
purpose: "Build the missing owner-decision loop of the review batch: items a no-commit batch run gave a verdict (batch-handled) stayed open with nobody told. findings.py digest keeps one issue listing them by verdict; findings.py digest-apply acts on a source only for rows the owner ticked (sender and previous body of the issues: edited event); new workflow review-digest.yml. Tests first."
repository_or_source_context: "branch feature/review/owner-digest from develop 7ecee08"
files_or_sections_affected: ["scripts/findings.py", "scripts/tests/findings-digest_test.sh", ".github/workflows/review-digest.yml", "docs/quality/REVIEW_CONTEXT.md", "TODO.md", "docs/ai-usage/2026-11-18-owner-digest.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["RED: findings-digest_test.sh failed (no digest subcommand)", "GREEN: findings-digest_test.sh 42 passed", "every review-*_test.sh and findings*_test.sh, lint-shell, check-docs, check-todo-updated, make lint-scripts (see the report)"]
open_risks: ["NOT RUN on GitHub: the issues: edited event payload (sender, changes.body.from) and the default-branch copy of the workflow take effect only after the merge into develop", "an owner edit made through another tool that rewrites the whole body still counts as the owner's tick only for rows that edit changed from unticked to ticked"]
```
