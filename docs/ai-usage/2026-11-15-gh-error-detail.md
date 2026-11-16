## GitHub error detail and one retry in the review scripts

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent)"
purpose: "Batch run 43673158317 failed in the publish step with only `FAILED (gh failed)`: no status, no message, a transient error that a rerun cleared. A failed gh call now reports the method, the endpoint without query string, the exit code and the first line of stderr (200 characters at most, secret shapes redacted, no request body), and is retried once after a short delay when that is safe (any GET except a 4xx; a write only after a 5xx, a secondary rate limit, a network error, or a 429 with Retry-After of 10 s at most)."
repository_or_source_context: "branch fix/review/gh-error-detail; owner-approved pipeline improvement"
files_or_sections_affected: ["scripts/review-context.py", "scripts/findings.py", "scripts/tests/review-batch-nothing_test.sh", "TODO.md", "docs/ai-usage/2026-11-15-gh-error-detail.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #232)"]
tests_or_evidence: ["RED first: six new cases in review-batch-nothing_test.sh failed before the change (502 once, GET once, 502 twice, 403, secret in stderr, GET 404)", "GREEN: review-batch-nothing_test.sh, findings-settle_test.sh, findings_test.sh, findings-batch_test.sh, review-autofix_test.sh, review-context_test.sh"]
open_risks: ["not run against the real GitHub API: the exact stderr wording of gh for 5xx and rate limits is matched on `HTTP <status>` and a few phrases", "a write retried after a 5xx may already have gone through: the comment writes stay idempotent by marker on the next run, but a retry inside one run can duplicate a comment"]
```
