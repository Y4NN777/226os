## Issue #116 before-freeze follow-ups: profile patch validation and violation contract

```yaml
date: "2026-11-13"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "The two before-freeze items of issue #116 (PR #94 audit): (1) NUL or invalid UTF-8 in display_name/bio passed the schema and died on the Postgres 22021 text check as a 500 that also burnt one of the 60 profile:write hits per minute — validateProfileWrite now rejects !utf8.ValidString or a NUL rune with 422 invalid_value on /display_name and /bio; (2) the service emitted violation codes outside the OpenAPI Violation.code enum (invalid_length, too_many, unknown_value) and bare field names — every violation now uses an enum code (too_short, too_long, too_many_items, invalid_format, invalid_value, duplicate) and the JSON pointer of the request member (/skills, /field_visibility/bio, /country…) so the S7 form can map them. pgtest asserts the problem body's violations against the enum and pointer shape on the real router."
repository_or_source_context: "issue #116 (PR #94 audit follow-ups), branch fix/identity/profile-audit-116"
files_or_sections_affected: ["src/api/internal/identity/service/profile.go", "src/api/internal/identity/service/profile_test.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-13-issue-116-profile-validation.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["go test -race -count=1 ./internal/identity/... PASS (new TestUpdateMyProfileViolationsUseContractCodesAndPointers: 16 cases)", "go test -race -count=1 -tags integration -run 'TestIdentityProfiles|TestIdentityProfile' ./internal/platform/pgtest/ PASS on real PostgreSQL (NUL bio -> 422, assertProblemViolations on unknown skill/country)", "go vet clean, gofmt clean"]
open_risks: ["the three after-freeze items of issue #116 stay open: skills DELETE removes hidden-tag links, the C14b matrix lacks a moderator row and a byte-for-byte 404 body comparison, and the Cache-Control/Vary + resolveSkills-outside-transaction notes", "invalid UTF-8 cannot arrive through encoding/json (it coerces to U+FFFD); the utf8.ValidString half of validText is defence in depth for non-HTTP callers"]
```
