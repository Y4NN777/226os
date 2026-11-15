## Emergency production deploy without the staging gate

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (systalink session); design agreed with the lead session (Claude Opus 5.5)"
purpose: "The AI wrote a separate emergency workflow that deploys production without the staging tested-<tree> gate, on the owner's decision: manual dispatch on main with a reason and the confirmation EMERGENCY (validated by scripts/deploy-plan.sh), own build and scan, images tagged emergency-<sha>-<run id>-<attempt> and deployed by the pushed digests behind the production environment, never verified, attested or tagged tested-*, concurrency shared with deploy.yml."
repository_or_source_context: "branch feature/deploy/emergency-path, from develop after #303"
files_or_sections_affected: [".github/workflows/deploy-emergency.yml", "scripts/deploy-plan.sh", "scripts/tests/deploy-plan_test.sh", "scripts/tests/deploy-emergency_test.sh", "scripts/tests/deploy-images_test.sh", "docs/runbooks/deploy-setup.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["deploy-plan_test 19/19 and deploy-emergency_test 16/16, RED first", "deploy-images_test 18/18 with the two emergency-tag cases", "all deploy-* and e2e-* test files pass", "scripts/lint-shell.sh OK", "scripts/check-docs.py PASS"]
open_risks: ["nothing ran on GitHub: the first real emergency run is the first proof", "no E2E before an emergency deploy (accepted by the owner), only the smoke test after it", "the deploy steps are copied from deploy.yml: a later change there must be mirrored here until a shared workflow_call exists", "actionlint NOT RUN"]
```
