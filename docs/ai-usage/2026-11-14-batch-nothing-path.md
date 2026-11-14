## Review batch: the "nothing to fix" path

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session and the ci-implementer subagent)"
purpose: "A review batch whose fixer makes no commit (run 43633446506, batch #226) settled nothing and looped forever. The trusted publish job now applies the fixer's verdicts with the rules of settle, posts one verdict table on the batch issue, ticks the items with a verdict and closes the batch, and new batches leave the owner-held items out until they change. Also two MEDIUM findings from the review of #223: FIXED ELSEWHERE naming the source's own pull request, and a CRITICAL whose override cannot be read."
repository_or_source_context: "branch fix/review/batch-nothing-path from develop 68b3312 (after #223); batch issue #226; owner decision 2026-11-14 ~19:55 and ~20:05 GMT"
files_or_sections_affected: [".github/workflows/review-batch.yml (publish job)", "scripts/review-autofix.py", "scripts/findings.py", "scripts/review-gate.py", "scripts/tests/review-batch-nothing_test.sh", "scripts/tests/findings-batch_test.sh", "scripts/tests/findings-settle_test.sh", "scripts/tests/review-gate_test.sh", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
human_reviewers: ["Y4NN777 (pipeline owner, merge click)"]
closure: "open"
tests_or_evidence: ["RED then GREEN: review-batch-nothing_test.sh (new), findings-batch_test.sh (held items), findings-settle_test.sh (own source PR), review-gate_test.sh (events read error)", "make lint-scripts", "make docs-check", "YAML parse of review-batch.yml"]
open_risks: ["NOT RUN on a real runner", "the publish job is not in the review-batch-write concurrency group: a batch run reading at the same moment can race the close", "batch #226 stays open until one manual dispatch after the merge", "FIXED ELSEWHERE is refused when the source was filed after the fixing pull request merged (fails safe, a human checks)"]
```
