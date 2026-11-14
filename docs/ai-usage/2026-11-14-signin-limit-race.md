## Sign-in account limit under concurrency (issue #97)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer subagent)"
purpose: "Issue #97: the per-account sign-in limit was check-then-act, so concurrent wrong passwords all reached the password check. The owner chose to count the attempt up front with one atomic hit and give it back on a successful sign-in."
repository_or_source_context: "branch fix/identity/signin-limit-race; issue #97; ADR-0003 rate limits"
files_or_sections_affected: ["src/api/internal/identity/**", "db/queries/identity.sql", "src/api/internal/platform/dbgen/identity.sql.go (generated)", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/app/server_identity_test.go", "src/api/internal/platform/pgtest/signin_limit_integration_test.go", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: TestIdentitySignInAccountLimitHoldsUnderConcurrency failed on the old code (20 password checks, want 5)", "GREEN: the same test (5 INVALID_CREDENTIALS, 15 RATE_LIMITED) and TestIdentitySignInSuccessGivesTheHitBack pass with -race"]
open_risks: ["more than 5 simultaneous correct sign-ins to one account can briefly get 429", "a give-back skipped under a concurrent bucket change keeps the count strict, never lower"]
```
