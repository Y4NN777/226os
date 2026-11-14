## Environment hosts in the encrypted secrets (staging and production)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session, no subagent)"
purpose: "With the owner's explicit permission, set APP_BASE_URL and PUBLIC_HOST in the staging secrets and create the production secrets file, using SOPS without reading or printing any value (sops set; scripts/secrets.sh init generates the database password, DATABASE_URL and CURSOR_KEY). Records the staging host decision (H7)."
repository_or_source_context: "branch chore/ops/environment-hosts; issue #109 section 1; ADR-0009; TODO H6/H7"
files_or_sections_affected: ["secrets/staging-test.enc.env (encrypted)", "secrets/production-live.enc.env (encrypted, new)", "TODO.md", ".claude/settings.json (deny list: reading the encrypted secrets/*.enc.env is allowed, ciphertext and SOPS metadata only, so the session can check names and recipients; the age key directory **/sops/age/** and /dev/shm/** are denied; every Edit deny on secrets stays)"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: configuration, no dependency"
closure: "open"
human_reviewers: ["Y4NN777 (owner, secrets)"]
tests_or_evidence: ["scripts/secrets.sh check: check-secrets-encrypted PASS", "make security-secrets: no leaks found (working tree and history of HEAD)", "no secret value was displayed: sops set and init print nothing but a 'Created' line"]
open_risks: ["226os.tech and staging.226os.tech have no DNS A record yet", "AI_API_KEY stays empty until the owner fills it; ACME_EMAIL (both files) and production SMTP were set by the owner with hidden input, never seen by the session", "production secrets are encrypted to the current recipients only: add the host age key, rotate, and regenerate POSTGRES_PASSWORD and CURSOR_KEY before deploying"]
```
