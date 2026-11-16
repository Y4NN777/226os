## Email verification resend (issue #248, item 1)

```yaml
date: "2026-11-15"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Issue #248 item 1. The AI wrote the failing tests first, then: the resendEmailVerification contract operation (owner-approved, R7), the identity service and handler (always 202, 3 per account and 10 per IP per hour counted before the lookup, token drawn before the lookup, token write and mail off the request path), CreateEmailVerification now burns every earlier open code of the account in the same statement, and the « Renvoyer le code » action on /verifier-email. Items 2 (trigram search) and 3 (learn_progress) are not in this change."
repository_or_source_context: "branch feature/identity/verification-resend, from develop b4a891c"
files_or_sections_affected: ["api/openapi.yaml", "generated apigen, apispec, schema.d.ts, dbgen", "db/queries/identity.sql", "src/api/internal/identity/**", "src/api/internal/app/server.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "scripts/public-routes.txt", "src/web/src/features/identity/**", "docs/decisions/0003-*.md", "docs/SPEC.md", "docs/architecture/ARCHITECTURE.md", "docs/quality/COVERAGE_SCOPE.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["service, transport, server and Vitest tests RED first; the integration test fails on the old CreateEmailVerification (2 open codes) and passes on the new one; see the pull request for the gate results"]
open_risks: ["the old code dies at the resend, so a mail still in transit with the older code no longer works"]
```
