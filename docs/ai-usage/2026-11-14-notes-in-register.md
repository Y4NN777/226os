## AI review: notes in the findings register

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (main session, no subagents)"
purpose: "The review ledger keeps two kinds of entry: findings (defects) and notes (remarks), each with an evidence tag OBSERVED or INFERRED. Open notes become one 'PR<N> notes' checklist issue per pull request on merge (label kind:note, due after the freeze); the dashboard has a separate Notes section and flags INFERRED notes undecided for 7 days; the review pack lists decided notes (ticked, or closed as not planned) so they are not raised again. A note never triggers a new review or an owner ping. Also adds the owner task for the GitHub API token (S5) to the release-day checklist."
repository_or_source_context: "branch chore/review/notes-in-register; owner request after the review-register work (#113, #115)"
files_or_sections_affected: ["scripts/findings.py", "scripts/review-context.py", "scripts/review-ping.py", "scripts/tests/findings_test.sh", "scripts/tests/review-context_test.sh", "scripts/tests/review-ping_test.sh", ".claude/commands/delegate-review.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: the new cases in findings_test.sh, review-context_test.sh and review-ping_test.sh failed before the change", "GREEN: findings 40 passed, review-context 50 passed, review-ping 13 passed"]
open_risks: ["NOT RUN end to end until the first merged pull request after this PR: the model must actually write kind and evidence in the ledger (an entry without them is read as a finding, so nothing breaks, but no note is filed)", "the kind:note label is created by the issue API on first use; the board 4 Auto-add filter stays label:finding, so notes are not on the board"]
```
