## S5 Build PR 2: applications

```yaml
date: "2026-11-14"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "S5 Build, second PR: createApplication (verified email; need locked FOR SHARE; someone else's private project 404, own project and closed need 409 STATE_CONFLICT; INSERT ON CONFLICT DO NOTHING then the existing row locked: a withdrawn application is reopened, anything else 409 DUPLICATE_APPLICATION, so a double click never makes a second row), listProjectApplications (owner only, owner predicate inside the query, 403 public / 404 private with no count), decideApplication (decided on the application row locked with its project: neither applicant nor owner 404, wrong party 403, not pending 409, guarded UPDATE ... WHERE status = 'pending')."
repository_or_source_context: "branch feature/build/applications (stacked on feature/build/projects-needs)"
files_or_sections_affected: ["db/queries/build.sql", "src/api/internal/platform/dbgen/build.sql.go", "src/api/internal/platform/dbgen/build_applications_queries_test.go", "src/api/internal/build/**", "src/api/internal/app/server.go", "src/api/internal/app/build_internal_test.go", "src/api/internal/platform/pgtest/build_applications_integration_test.go", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt", "TODO.md", "docs/ai-usage/2026-11-14-s5-applications.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["NICE-DEV226 (pending)"]
tests_or_evidence: ["RED: the four application pgtest tests answered 501 before the routes were wired", "GREEN: go test -race on internal/build/... and the dbgen contract test; pgtest on real PostgreSQL: lifecycle (apply, replay, duplicate 409, owner read, withdraw, re-apply reopens the row, decline, no re-apply after decline, no second decision, closed need), double click (one 201, one 409, one row), privacy and IDOR (403/404 without data, maintainer row grants nothing, third party and other project's owner 404), concurrent accept and withdraw (one 200, one 409)"]
open_risks: ["an empty message is 422 although the contract allows it (applications_message_ck needs 1 to 2000 characters)", "the decision note is not stored (no column)", "a reopened application keeps its first created_at, so it sorts by its first application"]
```
