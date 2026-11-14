Triage: NONE | blast radius: this PR only (review pipeline scripts and workflow)
<!-- 226os-ai-review -->
Tool: PASS (8 files)

No finding above a note. The range is a small, well-tested step: one comment per run, the newest created marker comment read by the gate, context, auto-merge and deep merge, and older bot comments collapsed.

Notes:
- `unreachable-code-after-return` (OBSERVED): `scripts/review-comment.py:291` returns 0 at the end of `main()`, and the old minimize loop stays below it (lines 292 to 302). That loop refers to `target`, which no longer exists. It never runs, but it duplicates `create_and_collapse` and would raise a NameError if someone removes the early `return`. Delete it.
- `carry-forward-keeps-omitted-open` (INFERRED): `carry_forward` (`scripts/review-comment.py:238`) re-adds every previous entry the model leaves out, unless its status is `fixed`. An open HIGH or MEDIUM that the model silently drops after a fix therefore stays `open` indefinitely and keeps blocking auto-merge. The prompt tells the model to list fixed findings, so this relies on the model following that instruction.
- `deep-stale-run-overwrite` (carried forward, INFERRED, still open): the new stale check in `deep_main` skips the edit when a newer review comment exists. A newer push that starts no deep job still produces a review comment, so this note is now largely handled. A late older deep job no longer rewrites the newer comment.

Fixed since the last run: none of the previous findings were touched by these commits.

Coverage: 8 of the tool's files reviewed, plus 4 files I read myself (Step 1b: `.github/workflows/code-review.yml`, `.claude/commands/*.md`, `docs/quality/REVIEW_CONTEXT.md`, `docs/ai-usage/*`). Range: `217da017e453f0b175a9a890add84dc758035dac..40fc44bd7a159f9a995e70ddccabc105bb5be10f`. Gates and tests were NOT RUN by this review; rely on CI.

<details><summary>Review ledger</summary>

```json
{"sha": "40fc44bd7a159f9a995e70ddccabc105bb5be10f", "findings": [{"id": "deep-stale-run-overwrite", "kind": "note", "evidence": "INFERRED", "sev": "LOW", "file": ".github/workflows/code-review.yml", "line": 423, "scenario": "An older deep job could rewrite the comment's findings; the new stale check in deep_main largely covers it.", "status": "open"}, {"id": "unreachable-code-after-return", "kind": "note", "evidence": "OBSERVED", "sev": "LOW", "file": "scripts/review-comment.py", "line": 291, "scenario": "The old minimize loop stays after the return in main(), refers to an undefined target and never runs.", "status": "open"}, {"id": "carry-forward-keeps-omitted-open", "kind": "note", "evidence": "INFERRED", "sev": "LOW", "file": "scripts/review-comment.py", "line": 238, "scenario": "An open finding the model omits after a fix is carried forward as open and keeps blocking auto-merge.", "status": "open"}]}
```

</details>

