## S3 Solve PR 2: votes and accepted answer

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer and skeptic subagents)"
purpose: "S3 PR 2: castVote, retractVote, acceptAnswer and unacceptAnswer on migration 00005, with the score kept in the same transaction as the vote (row lock, then one statement), one accepted answer under concurrency, an independent security review and its two test fixes, and two owner decisions (answer paging under moving votes accepted; the S3 path extension)."
repository_or_source_context: "branch feature/solve/votes-accept; issue #67; TODO S3 block"
files_or_sections_affected: ["src/api/internal/solve/**", "db/queries/solve.sql", "src/api/internal/platform/dbgen/solve.sql.go (generated)", "src/api/internal/platform/dbgen/*_test.go", "src/api/internal/platform/pgtest/solve_votes_integration_test.go", "src/api/internal/app/** (wiring)", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt", "scripts/tests/authz-routes_test.sh", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["NICE-DEV226 (slice owner, pending)", "mibienpanjoe (reviewer, pending)"]
tests_or_evidence: ["RED first per layer (service, transport, repository, sqlc, app wiring); a mutation check removing the target lock made the duplicate-vote and reconciliation tests fail", "GREEN: concurrent accepts leave exactly one accepted answer; vote_score equals the sum of votes after a random burst; concurrent duplicate votes make one row; a removed question's answers are not votable (404)"]
open_risks: ["answer paging can repeat or skip an answer while votes move (accepted)", "a removal racing a vote is safe by the lock but has no test (no P0 path removes content yet)"]
```
