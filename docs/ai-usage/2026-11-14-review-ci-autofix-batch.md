## AI review: auto-fix in CI and one fix pull request per batch

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent)"
purpose: "Run the review loop in CI with little owner time: Claude Code fixes open MEDIUM and HIGH findings of a pull request (stage 3) and opens one fix pull request per review batch (stage 5), as a normal project session, behind deterministic decisions and a push guard."
repository_or_source_context: "branch chore/review/ci-autofix-batch (stacked on fix/review/ledger-head-sha); anthropics/claude-code-action v1.0.275 inputs read at the pinned commit (setting sources default to user, project, local)"
files_or_sections_affected: [".claude/commands/review-fix.md", ".github/workflows/review-autofix.yml", ".github/workflows/review-batch.yml", "scripts/review-autofix.py", "scripts/tests/review-autofix_test.sh", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: ["https://github.com/anthropics/claude-code-action (action.yml and base-action/src/parse-sdk-options.ts at 4d75e5f390ee76b2e09c1980ecc99ff3a2e088c8, read only)"]
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: scripts/tests/review-autofix_test.sh 40 failed, 3 passed before the script existed", "GREEN: review-autofix_test.sh 43 passed; after the security split (fixer-run wrapper, .npmrc and owner preferences guarded) RED 9 failed, then GREEN 66 passed", "format-patch, am and check round trip simulated locally: PASS", "YAML parse of the three review workflows PASS"]
open_risks: ["NOT RUN against GitHub: workflow_run and schedule run only from the default branch copy", "actionlint NOT RUN (not installed)", "make runs pull request code while the Claude OAuth token is in the action step (same-repository branches only)", "scripts/review-ping.py clears needs-owner when nothing CRITICAL or HIGH is open; the attempt cap rests on the autofix-2 label"]
```
