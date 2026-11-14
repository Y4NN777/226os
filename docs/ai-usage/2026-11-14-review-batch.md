## AI review: triage the register in batches

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent)"
purpose: "Work the AI-review register (open finding issues, undecided notes) in ranked batches, one fix pull request per batch, instead of one by one."
repository_or_source_context: "branch chore/review/triage-batch; scripts/findings.py; docs/quality/REVIEW_CONTEXT.md"
files_or_sections_affected: ["scripts/findings.py", "scripts/tests/findings-batch_test.sh", ".github/workflows/review-batch.yml", "mk/60-review.mk", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: scripts/tests/findings-batch_test.sh failed 16 of 20 before the subcommand existed", "GREEN: findings-batch_test.sh 26 passed; findings_test.sh 40 passed"]
open_risks: ["NOT RUN against the real GitHub API", "the scheduled trigger runs only from the default branch", "actionlint NOT RUN (not installed)"]
```
