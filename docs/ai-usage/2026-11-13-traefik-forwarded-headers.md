## Traefik forwarded headers pinned (#117 item 2)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "Issue #117 item 2: pin forwardedHeaders.insecure=false on every Traefik entrypoint (local-dev, staging-test, production-live) and enforce it, plus no trusted upstream proxy, in scripts/verify-containers.sh; record the staging spoofed X-Forwarded-For smoke as a release-checklist step for the owner."
repository_or_source_context: "branch chore/ops/traefik-forwarded-headers, issue #117"
files_or_sections_affected: ["compose.yaml", "scripts/verify-containers.sh", "TODO.md", "docs/ai-usage/2026-11-13-traefik-forwarded-headers.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: configuration and test only, no new dependency"
closure: "open"
human_reviewers: ["Y4NN777 (ops area, pending)"]
tests_or_evidence: ["RED: scripts/verify-containers.sh failed on the 3 profiles (entrypoints web and websecure did not pin forwardedHeaders.insecure=false)", "GREEN: scripts/verify-containers.sh ALL PASS", "the pinned traefik:v3.6 image accepts --entrypoints.web.forwardedHeaders.insecure=false and rejects an unknown forwardedHeaders field"]
open_risks: ["the end-to-end spoofed X-Forwarded-For smoke through Traefik is not run (no deployed host yet): release-checklist step"]
```
