## ADR-0013 default branch develop, pipeline trust anchor

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead + ci-implementer (Sonnet)"
purpose: "ADR-0013 (default branch develop so review stages 3 to 5 run), the docs and comments that followed from it, and a static guard test keeping jobs of the review-agent environment out of pull request events."
repository_or_source_context: "branch chore/review/default-branch-develop; owner decisions of 2026-11-14; TODO AI review block; GitHub documentation on workflow triggers and the default branch"
files_or_sections_affected: ["docs/decisions/0013-default-branch-develop.md", "docs/decisions/README.md", "docs/GIT_WORKFLOW.md", "docs/quality/REVIEW_CONTEXT.md", ".github/workflows/review-autofix.yml", ".github/workflows/review-automerge.yml", ".github/workflows/review-batch.yml", "scripts/close-merged-issues.py", "scripts/tests/review-agent-triggers_test.sh", "TODO.md"]
third_party_sources_consulted: ["https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows", "https://github.blog/changelog/2025-11-07-actions-pull_request_target-and-environment-branch-protections-changes/", "https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository"]
license_review: "NOT APPLICABLE: original text and code, no new dependency"
closure: "open"
human_reviewers: ["Y4NN777 (owner, pending)"]
tests_or_evidence: ["RED: scripts/tests/review-agent-triggers_test.sh on a pull_request_target fixture exits 1", "GREEN: same test on .github/workflows exits 0", "RED (fail-closed pass): 18 new fixtures (expression, anchor and alias environments, flow-mapping and odd-indent jobs, unsafe if: forms, always() with needs, unreadable on:, workflow_call) accepted by the first version, 38 of 38 checks GREEN after the rewrite; mutation: review-batch.yml copy without the plan guard is refused", "actionlint: not installed (command -v actionlint empty), NOT RUN"]
open_risks: ["stages 3 to 5 NOT RUN until the owner switches the default branch", "owner of REVIEW_AGENT_TOKEN undecided", "scripts/findings.py docstring (~lines 415-416) still says closing keywords act only on main"]
```
