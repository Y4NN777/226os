## AI review: deep review (stage 2), auto-merge (stage 4), NOT RUN fallback

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent of the lead session)"
purpose: "Run the review and merge process in CI with little owner time. An empty or invalid review output now writes the marker comment with `Triage: NOT RUN | <reason>` and the previous ledger carried forward (NOT RUN is never a pass; the gate reads it as NOT RUN instead of an older verdict). Stage 2: a read-only adversarial deep review (Opus, .claude/commands/deep-review.md, skeptic mandate) runs after the review when a P0 path is touched and merges its findings into the same comment and ledger (source deep). Stage 4: scripts/review-automerge.py and .github/workflows/review-automerge.yml merge a pull request into develop with a merge commit only when every required check, the review, the deep review and the owner rules agree, then update the other open pull requests."
repository_or_source_context: "branch chore/review/ci-deep-review-merge, stacked on fix/review/ledger-head-sha (PR 159); branch protection of develop read through the API (required checks Validate branch promotion, quality, Review gate; strict false; conversation resolution required)"
files_or_sections_affected: ["scripts/review-comment.py", "scripts/review-gate.py", "scripts/review-automerge.py", ".github/pipeline-paths.txt", "scripts/review-strip-pr.sh", "scripts/tests/review-strip-pr_test.sh", "scripts/tests/review-comment_test.sh", "scripts/tests/review-gate_test.sh", "scripts/tests/review-automerge_test.sh", ".github/workflows/code-review.yml", ".github/workflows/review-automerge.yml", ".github/deep-review-paths.txt", ".claude/commands/deep-review.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["Y4NN777 (owner, pending)"]
tests_or_evidence: ["RED: review-comment 16 of 28 fail, review-gate 1 of 49, review-automerge 34 of 38 (no script)", "GREEN: review-comment 30, review-gate 50, review-automerge 44 passed (after merging the PR 159 hardening)", "PR 168 review (4 MEDIUM): RED review-comment 4 of 39, review-automerge 12 of 50, review-strip-pr 1 of 4 failed; GREEN 39, 50 and 4 passed", "PR 168 second review (2 MEDIUM): RED review-comment 3 of 41, review-automerge 13 of 53; GREEN 41 and 53 passed"]
open_risks: ["auto-merge runs from main (workflow_run uses the default branch): NOT RUN end to end before the promotion", "the deep job and the NOT RUN fallback run only once their scripts are on develop (they are read from the base commit)", "a CRITICAL replaced by a later NOT RUN no longer fails the gate (the gate fails open on NOT RUN); auto-merge still refuses", "update-branch after each merge costs one CI run and one review per open pull request"]
```
