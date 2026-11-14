## Issue #116 after-freeze follow-ups: skills delete scope, C14b matrix, cache headers

```yaml
date: "2026-11-13"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "The after-freeze items of issue #116 (PR #94 audit): (1) DeleteUserSkills wiped every user_skills row, including links to moderator-hidden tags that the read projections filter out — a hidden link was silently destroyed by an unrelated write and restoring the tag did not bring the skill back; the delete now joins tags and removes only active-tag links, proven on real PostgreSQL (link survives the rewrite, skill returns on tag restore). (2) The C14b matrix lacked a moderator row and a byte-for-byte 404 comparison — TestIdentityProfileNotFoundBodiesIdentical runs unknown, deleted, suspended, members-only-for-anonymous and private-for-member-and-moderator through the real router and compares status, Content-Type and the problem body minus request_id and instance. (3) GET /me (email, PII) and GET /profiles/{username} (caller-dependent body) now send Cache-Control: no-store so a future shared cache cannot leak a members-only 200 to an anonymous caller."
repository_or_source_context: "issue #116 (PR #94 audit follow-ups), branch fix/identity/profile-audit-116-after"
files_or_sections_affected: ["db/queries/identity.sql", "src/api/internal/platform/dbgen/identity.sql.go", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/identity/transport/profile.go", "src/api/internal/identity/transport/profile_test.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-13-issue-116-after-freeze.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["go test -race -count=1 ./internal/identity/... PASS", "go test -race -count=1 -tags integration -run 'TestIdentityProfile|TestIdentityProfiles' ./internal/platform/pgtest/ PASS on real PostgreSQL (TestIdentityProfileHiddenTagFiltered extended: hidden link survives, restore brings skill back; TestIdentityProfileNotFoundBodiesIdentical: 7-case matrix)", "make generate regenerated dbgen (comment + scoped DELETE), check-sqlc drift committed", "new TestGetMeAndGetProfileSendNoStore pins the header", "go vet, gofmt clean"]
open_risks: ["the undeclared 422 on getProfile (validator rejects a malformed path username; the contract lists no 422 response) is a contract change left for the owner before/after freeze decision", "resolveSkills still runs outside the write transaction: a tag hidden between resolve and insert is linked but filtered on read — accepted and documented, no leak", "hidden-tag links now accumulate until the tag is restored or purged; no purge exists for user_skills (same as before for active links)"]
```
