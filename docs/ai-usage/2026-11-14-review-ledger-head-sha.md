## AI review: ledger head, one comment, deterministic preview, P0 calibration

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent of the lead session)"
purpose: "The review ledger recorded GitHub's temporary merge commit, so every push got a full re-review. The pack now prints the pull request head (PR_HEAD_SHA) as `Head commit:`, computes the range against it, and the ledger must record it; decide() returns incremental for any push with new commits, skip only without new commits, first without a usable ledger (with the reason in one line). Same pull request: the workflow runs the open-code-review preview itself and the comment carries a fixed `Tool:` line; one review comment per pull request, written by scripts/review-comment.py by marker (older bot Triage duplicates minimized), with the gate and the ledger reader preferring the marker comment, else the most recently updated one; an INFERRED concern with a plausible failing scenario on a P0 path is a MEDIUM finding."
repository_or_source_context: "branch fix/review/ledger-head-sha; runs 43578279388 (#147) and 43581882270 (#153, 6 turns, 0 permission denials, preview not run); the ledgers of #147 and #149 (sha not in the repository); the duplicate Triage comments of #100 to #102; the pinned claude-code-action source (structured_output with --json-schema)"
files_or_sections_affected: ["scripts/review-context.py", "scripts/review-comment.py", "scripts/review-gate.py", "scripts/tests/review-context_test.sh", "scripts/tests/review-comment_test.sh", "scripts/tests/review-gate_test.sh", ".github/workflows/code-review.yml", ".claude/commands/delegate-review.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["Y4NN777 (owner, pending)"]
tests_or_evidence: ["RED: 13 of 59 tests fail against the previous script", "GREEN: bash scripts/tests/review-context_test.sh exit 0 (59 passed)", "RED then GREEN: review-context 8 new failures, review-gate 3, review-comment (no script); GREEN 68, 48 and 13 passed"]
open_risks: ["each open pull request gets one more full review, since its old ledger points at a vanished merge commit", "this pull request's own review run uses develop's scripts, so it posts no comment (no review-comment.py on the base yet)", "structured output and the CI preview are NOT RUN in CI until merged"]
```
