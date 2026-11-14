## AI review: the ledger records the pull request head

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent of the lead session)"
purpose: "The review ledger recorded GitHub's temporary merge commit, so every push got a full re-review. The pack now prints the pull request head (PR_HEAD_SHA) as `Head commit:`, computes the range against it, and the ledger must record it; decide() returns incremental for any push with new commits, skip only without new commits, first without a usable ledger (with the reason in one line)."
repository_or_source_context: "branch fix/review/ledger-head-sha; runs 43578279388 (#147) and the ledgers of #147 and #149 (sha not in the repository)"
files_or_sections_affected: ["scripts/review-context.py", "scripts/tests/review-context_test.sh", ".github/workflows/code-review.yml", ".claude/commands/delegate-review.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["Y4NN777 (owner, pending)"]
tests_or_evidence: ["RED: 13 of 59 tests fail against the previous script", "GREEN: bash scripts/tests/review-context_test.sh exit 0 (59 passed)"]
open_risks: ["the workflow prompt still says `--to HEAD`; delegate-review.md tells the model to use the pack's Range ends", "each open pull request gets one more full review, since its old ledger points at a vanished merge commit"]
```
