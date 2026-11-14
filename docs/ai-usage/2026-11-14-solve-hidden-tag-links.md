## S3 follow-up: hidden tag links survive question edits (issue #156)

```yaml
date: "2026-11-14"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "Fix the defect found in review of #147 and registered as issue #156: DeleteSolveQuestionTags deleted every question_tags row, so editing a question's tags silently dropped links to tags a moderator had set hidden (invisible on reads, impossible for the author to re-send). The delete now joins tags and removes only status='active' links, mirroring DeleteProjectTags (#131) and the #116 user_skills fix (#130)."
repository_or_source_context: "branch fix/solve/hidden-tag-links from develop 22a337b; issue #156; db/queries/solve.sql; pgtest solve_integration_test.go"
files_or_sections_affected: ["db/queries/solve.sql", "src/api/internal/platform/dbgen/solve.sql.go (regenerated, not hand-edited)", "src/api/internal/platform/pgtest/solve_integration_test.go", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: TestSolveEditKeepsAHiddenTagLink FAILS on the pre-fix query (0 link rows left)", "GREEN: same test PASS with the scoped delete; TestSolveEditKeepsAHiddenLinkHidden still PASS; go test -race -tags integration on real PostgreSQL"]
open_risks: ["InsertSolveQuestionTags still has no status filter: a client that guesses a hidden slug can re-link it; reads filter it anyway, same exposure as build's InsertProjectTags — accepted, noted in issue #156"]
```
