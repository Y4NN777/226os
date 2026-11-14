## Review register: self-contained issues (detail, fix, source, permalink)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent), lead session Claude Opus 5.5"
purpose: "Owner-approved pipeline task: the AI-review ledger and the issues built from it lose the proof, the recommended fix and the exact comment (issue #172 against the last review comment of PR #168); a stage 1 note repeated by the deep review was re-tagged source deep (PR #180). Entries now carry detail and fix, the script sets source, first_sha and location, texts are neutralised and secret shapes redacted, issues are self-contained, duplicates across pull requests are one batch row, and the fixer receives the context."
repository_or_source_context: "branch chore/review/register-detail; the verbatim last review comment of PR #168 is a test fixture"
files_or_sections_affected: ["scripts/review-comment.py", "scripts/review-context.py", "scripts/findings.py", "scripts/review-autofix.py", "scripts/tests/ (findings, findings-batch, findings-settle, review-comment, review-autofix, fixtures/review-register)", ".claude/commands/delegate-review.md", ".claude/commands/deep-review.md", ".claude/commands/review-fix.md", ".github/workflows/review-batch.yml", ".github/workflows/review-autofix.yml", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency; secret shapes written from public token formats, no copied rule file"
closure: "open"
human_reviewers: ["Y4NN777 (pipeline owner, pending)"]
tests_or_evidence: ["RED: 26 review-comment, 16 findings, 3 findings-batch, 3 findings-settle and 8 review-autofix assertions failed before the change", "GREEN: the eleven script tests, make lint-scripts, make docs-check and a YAML parse of both workflows"]
open_risks: ["code-review.yml (not in this change) does not set LOCATION_REPO: the deep job finds the head under pr/, the review job in the checkout; a case neither finds is recorded as location: unverified, never a wrong permalink", "the finding context is written before repository code runs but lives in the same directory as the base documents, which repository code can still reach (same trust level as those documents)", "an entry the model lists again with another file or line takes the head as first_sha (the permalink stays correct, the 'first raised' date moves)"]
```
