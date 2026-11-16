## Automated back-merge of main into develop

```yaml
date: "2026-11-16"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "The AI wrote the workflow and the script that open the main -> develop back-merge pull request after every Deploy or Deploy (emergency) run on main (GIT_WORKFLOW step 6), with REVIEW_AGENT_TOKEN so the required checks run, auto-merge as a merge commit, NO-OP when develop already contains main, and their tests."
repository_or_source_context: "branch feature/ops/sync-main-into-develop, from develop 994f1b3"
files_or_sections_affected: [".github/workflows/sync-main-into-develop.yml", "scripts/sync-main-into-develop.sh", "scripts/tests/sync-main-into-develop_test.sh", "docs/GIT_WORKFLOW.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["sync-main-into-develop_test.sh 14/14, RED first", "review-agent-triggers_test.sh 51/51 with the new workflow"]
open_risks: ["nothing ran on GitHub: workflow_run takes effect only once merged into develop", "REVIEW_AGENT_TOKEN has no Workflows permission: GitHub may refuse to point the branch at a main commit that changes workflow files; the run then fails visibly and the back-merge is opened by hand", "actionlint NOT RUN"]
```
