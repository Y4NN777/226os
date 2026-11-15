## Deployment pipeline (staging and production)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, CI implementer subagent (Sonnet), brief by the owner's lead session"
purpose: "Wrote the deployment to two fresh servers: the server bootstrap script, the deploy workflow (staging on a green develop push, production on main by the digests tested on staging), the helper scripts and their offline tests, the owner setup runbook and the SHIP and release documents."
repository_or_source_context: "branch feature/ops/deploy-pipeline, from develop f939030"
files_or_sections_affected: [".github/workflows/deploy.yml", "scripts/server-bootstrap.sh", "scripts/deploy-plan.sh", "scripts/deploy-images.sh", "scripts/deploy-env.sh", "scripts/deploy-remote.sh", "scripts/tests/", "docs/SHIP.md", "docs/RELEASE_DRAFT.md", "docs/runbooks/deploy-setup.md", "docs/GIT_WORKFLOW.md", "mk/50-ops.mk"]
third_party_sources_consulted: ["sops v3.13.3 release checksums (pinned binary)"]
license_review: "NOT APPLICABLE: no dependency added to the project; the CI downloads the pinned sops binary (MPL-2.0) at run time"
closure: "open"
human_reviewers: []
tests_or_evidence: ["scripts/tests/{server-bootstrap,deploy-plan,deploy-images,deploy-env,deploy-remote}_test.sh"]
open_risks: ["nothing was run against a server or GitHub: first real run is the test", "the workflow was not checked with actionlint", "GHCR package access for the job token is unverified", "ADR-0009 point 8 needs an amendment"]
```
