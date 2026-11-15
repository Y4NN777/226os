## Review follow-ups of PR #285 and PR #286 (issues #290 to #293)

```yaml
date: "2026-11-15"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Review follow-ups of #285 and #286. The AI checked websearch_to_tsquery on real PostgreSQL, found the Go negation heuristic wrong on four forms, wrote the failing integration test first, then moved the negation rule into the Similar* queries (no fallback when the tsquery holds a NOT) and removed the heuristic. On identity it added a warning when a resend mail fails (RED first), corrected the comments about concurrent resends, and wrote the accepted residual risk of #290 in ADR-0003, as chosen by the contributor."
repository_or_source_context: "branch fix/discover/review-followups-248, from develop c88634d"
files_or_sections_affected: ["db/queries/search.sql", "db/queries/identity.sql", "src/api/internal/platform/dbgen/**", "src/api/internal/search/service/**", "src/api/internal/identity/service/**", "src/api/internal/platform/pgtest/search_typo_integration_test.go", "docs/decisions/0003-*.md", "docs/architecture/ARCHITECTURE.md", "docs/spec-kit/DATABASE_DESIGN.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["TestSearchTypoFallbackFollowsThePostgreSQLNegation RED on the four forms, then GREEN; TestResendEmailVerificationLogsAFailedMail RED then GREEN; see the pull request for the gate results"]
open_risks: ["the #290 acceptance is a security acceptance; the owner can overrule it"]
```
