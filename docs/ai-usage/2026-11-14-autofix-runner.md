## Auto-fix runner: sandbox dependencies and label permissions

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent), lead session Claude Opus 5.5"
purpose: "Owner-approved pipeline fix after the first live auto-fix run (PR #199, run 43625357930): the fixer could not run a test or commit because Claude Code's Linux Bash sandbox needs bubblewrap and socat, and the labels autofix-N and needs-owner failed with HTTP 403 because labelling a pull request needs pull-requests: write. The two fixer jobs install the two packages from the runner's apt before the Claude step; the decide job gets pull-requests: write; the labels written after the push move from the publish job (which checks the branch out) to a new label job with no checkout and no secret."
repository_or_source_context: "branch fix/review/autofix-runner"
files_or_sections_affected: [".github/workflows/review-autofix.yml", ".github/workflows/review-batch.yml", "scripts/tests/review-runner_test.sh", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency in the repository (two distro packages installed on the runner at run time)"
closure: "open"
human_reviewers: ["Y4NN777 (pipeline owner, pending)"]
tests_or_evidence: ["RED: scripts/tests/review-runner_test.sh failed 4 assertions on develop f0b4e79", "GREEN: review-runner_test.sh 9 passed, review-agent-triggers_test.sh, make lint-scripts, make docs-check and a YAML parse of both workflows"]
open_risks: ["NOT RUN on a real runner: the install step and the label job are checked statically only until the next auto-fix run", "actionlint is not installed locally: NOT RUN", "review-automerge.yml writes no label from its YAML (the script labels with its own token), so it is unchanged"]
```
