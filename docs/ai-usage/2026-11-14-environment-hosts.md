## Environment hosts in the encrypted secrets (staging and production)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session, no subagent)"
purpose: "With the owner's explicit permission, set APP_BASE_URL and PUBLIC_HOST in the staging secrets and create the production secrets file, using SOPS without reading or printing any value (sops set; scripts/secrets.sh init generates the database password, DATABASE_URL and CURSOR_KEY). Records the staging host decision (H7)."
repository_or_source_context: "branch chore/ops/environment-hosts; issue #109 section 1; ADR-0009; TODO H6/H7"
files_or_sections_affected: ["secrets/staging-test.enc.env (encrypted)", "secrets/production-live.enc.env (encrypted, new)", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: configuration, no dependency"
closure: "open"
human_reviewers: ["Y4NN777 (owner, secrets)"]
tests_or_evidence: ["scripts/secrets.sh check: check-secrets-encrypted PASS", "make security-secrets: no leaks found (working tree and history of HEAD)", "no secret value was displayed: sops set and init print nothing but a 'Created' line"]
open_risks: ["staging.226os.tech has no DNS A record yet", "SMTP, AI_API_KEY and ACME_EMAIL stay empty until the owner fills them"]
```
