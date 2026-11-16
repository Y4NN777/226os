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
tests_or_evidence: ["Backtick rule (main session): the plain-word id row and the updated segment row failed first (47 passed, 2 failed); after names_id requires `id` in backticks, findings-settle-notes 49 passed, findings-settle 41, findings-batch 48, review-autofix 205", "RED: findings-settle-notes_test.sh 2 failed (note ticked and issue closed as not planned; mixed issue) before the change", "RED (#266 deep review, rule tightened): findings-settle-notes_test.sh 5 failed (own-file cite, unparsable own line, non-ACCEPTED ADR, ADR not naming the file ticked or closed)", "RED (skeptic gaps): findings-settle-notes_test.sh 3 failed (PROPOSED ADR with a later accepted line, ACCEPTED superseded, ADR naming the file only in longer paths or a URL); GREEN after anchoring the first Status line and a path-boundary match; MUTATION CHECK, not a true RED (code was written before the rows; owner, ADR must name the note id): findings-settle-notes_test.sh 3 failed (ADR naming the file but not the id, longer ids, other case) with the id check disabled; RED (skeptic, second round): rows 69, 70 and the path-mention unit row failed before the fix (3 failed, 45 passed); GREEN: 48 passed (id read from the issue body, file mentions masked, `/` and `.word` continue a token); earlier GREEN: 45 passed after threading the id into verified_reference; GREEN earlier: findings-settle-notes_test.sh 36 passed; findings-settle_test.sh 41 passed; findings, findings-settle, findings-batch, review-batch-nothing, review-autofix tests pass"]
open_risks: ["NOT RUN on a real runner", "the no-commit path (review-autofix.py nothing) shares settle_issue, so it ticks and closes notes the same way"]
```
