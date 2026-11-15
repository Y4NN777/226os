## Notes closed after triage in the review batch

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent, brief by the owner's lead session)"
purpose: "Notes issues (label kind:note) stayed open forever because the settle step never ticked a NOT A DEFECT or DECIDED verdict (rule H5 of #187). New rule (owner, 2026-11-15): for a note, such a verdict ticks the box when it cites a path:line, path:a-b or docs/decisions/NNNN-*.md that the script finds at the tip of develop, and the issue closes as not planned when every box is settled. Findings keep the old rule."
repository_or_source_context: "branch fix/review/notes-close-after-triage from develop 3ca5862"
files_or_sections_affected: ["scripts/findings.py", "scripts/tests/findings-settle-notes_test.sh", "scripts/tests/findings-settle_test.sh (comment)", "docs/quality/REVIEW_CONTEXT.md", ".claude/commands/review-fix.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["RED: findings-settle-notes_test.sh 2 failed (note ticked and issue closed as not planned; mixed issue) before the change", "GREEN: findings-settle-notes_test.sh 12 passed; findings, findings-settle, findings-batch, review-batch-nothing, review-autofix tests pass"]
open_risks: ["NOT RUN on a real runner", "the reference is checked for existence only, not for relevance to the note", "the no-commit path (review-autofix.py nothing) shares settle_issue, so it ticks and closes notes the same way"]
```
