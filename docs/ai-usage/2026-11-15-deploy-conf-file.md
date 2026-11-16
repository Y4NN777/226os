## Deploy settings in versioned files, E2E account in the sops file

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent, brief from the owner's lead session)"
purpose: "The AI moved the non-secret deploy settings (DEPLOY_HOST, DEPLOY_USER, PUBLIC_HOST, DEPLOY_KNOWN_HOSTS) from GitHub variables and secrets to deploy/staging.conf and deploy/production.conf, with a strict parser (scripts/deploy-conf.sh), and the E2E account from GitHub secrets to secrets/staging-test.enc.env (scripts/e2e-env.sh). It updated deploy.yml, deploy-env.sh, the bootstrap message, the tests and the deploy runbook."
repository_or_source_context: "branch feature/ops/deploy-conf-file, from develop 3fe6d93"
files_or_sections_affected: [".github/workflows/deploy.yml", "deploy/staging.conf", "deploy/production.conf", "scripts/deploy-conf.sh", "scripts/e2e-env.sh", "scripts/deploy-env.sh", "scripts/server-bootstrap.sh", "scripts/tests/deploy-conf_test.sh", "scripts/tests/e2e-env_test.sh", "scripts/tests/deploy-env_test.sh", "docs/runbooks/deploy-setup.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["scripts/tests/deploy-conf_test.sh, e2e-env_test.sh, deploy-env_test.sh, e2e-ci-config_test.sh pass; RED before the scripts existed"]
open_risks: ["not run in GitHub Actions: the sops decrypt of the E2E keys and GITHUB_ENV loading are untested end to end", "the E2E keys are not yet in secrets/staging-test.enc.env, so the e2e job fails until they are added"]
```
