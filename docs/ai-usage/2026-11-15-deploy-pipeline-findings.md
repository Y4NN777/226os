## Deploy pipeline findings #267 and #280: attested tested digests, E2E in its own job

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent) and Claude Opus 5.5 (lead session, skeptic review subagent)"
purpose: "The AI fixed two review findings of the deploy pipeline. #267: the build outputs the digests it pushed and refuses a different pre-existing <sha> tag; staging deploys those digests; promote attests them (actions/attest, predicate tree and sha) before tagging tested-<tree>; production verifies the attestation, its signer workflow and the tree (scripts/deploy-verify-tested.sh). #280: the E2E suite moved to its own job and environment staging-e2e (E2E_USER and E2E_PASSWORD only), and the junit report is checked for the account before upload (scripts/e2e-junit-check.sh)."
repository_or_source_context: "branch fix/ops/deploy-pipeline-findings, from develop 628532b"
files_or_sections_affected: [".github/workflows/deploy.yml", "scripts/deploy-images.sh", "scripts/deploy-verify-tested.sh", "scripts/e2e-env.sh", "scripts/e2e-junit-check.sh", "scripts/tests/deploy-images_test.sh", "scripts/tests/deploy-verify-tested_test.sh", "scripts/tests/deploy-conf_test.sh", "scripts/tests/e2e-ci-config_test.sh", "scripts/tests/e2e-env_test.sh", "scripts/tests/e2e-junit-check_test.sh", "docs/runbooks/deploy-setup.md", "docs/SHIP.md", "TODO.md"]
third_party_sources_consulted: ["actions/attest action.yml at 1e69f48 (v4.2.2)", "gh attestation verify manual"]
license_review: "no new dependency; actions/attest is MIT, pinned by SHA"
closure: "open"
human_reviewers: []
tests_or_evidence: ["the 10 deploy-* and e2e-* test files pass, RED first for each new case", "scripts/lint-shell.sh OK", "scripts/check-todo-updated.sh OK", "two independent skeptic reviews; the second: none blocking"]
open_risks: ["nothing ran on GitHub: the real push, attest and gh attestation verify path waits for a staging server", "the --source-ref refs/heads/develop value for a workflow_run signer is assumed, not observed", "a re-run of the build for an already pushed SHA fails closed (push a new commit)", "actionlint NOT RUN"]
```
