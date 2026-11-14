## S3 Solve PR 1: questions and answers

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer and skeptic subagents)"
purpose: "S3 PR 1: the six question and answer operations (listQuestions, createQuestion, getQuestion, updateQuestion, listAnswers, createAnswer) on the existing tables of migration 00005, with the platform packages (etag, idempotency, cursor, ratelimit, subject, apperr), an independent security review and its fixes, and three owner decisions (relevance order with q, ETag on editable fields, snippet and link limits recorded as deviations)."
repository_or_source_context: "branch feature/solve/questions-answers; issue #67; TODO S3 block; models src/api/internal/build and src/api/internal/learn"
files_or_sections_affected: ["src/api/internal/solve/**", "db/queries/solve.sql", "src/api/internal/platform/dbgen/solve.sql.go (generated)", "src/api/internal/platform/dbgen/solve_queries_test.go", "src/api/internal/platform/pgtest/solve_integration_test.go", "src/api/internal/app/** (wiring)", "scripts/unimplemented-routes.txt", "scripts/public-routes.txt", "scripts/authz-tests.txt", "scripts/tests/authz-routes_test.sh", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["NICE-DEV226 (slice owner, pending)", "mibienpanjoe (reviewer, pending)"]
tests_or_evidence: ["RED: the ten TestSolve* integration tests answered 501 before the handlers; each review fix and owner decision had its own failing test first", "GREEN: make generate, test-unit, coverage (Go first-party 100.0%), test-integration-go and security exit 0"]
open_risks: ["the unit tests of the first commit were written after the implementation (integration tests were RED first)", "snippets and question-link limits are deviations until a migration or contract PR after the freeze", "the answer cursor assumes a stable vote_score until PR 2"]
```
