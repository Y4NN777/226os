## Search candidate bound and cursor digest (#118 before-freeze items)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "Issue #118 before-freeze items: bound the search ranking (tsquery built once, no read for a query without a positive term, at most 500 candidates ranked), worst-case tests that measure rows scanned instead of the presence of a Limit node, a redacted 500 on a query timeout, and a fixed 16-byte filter digest in cursors so they stay within the contract's 512 characters."
repository_or_source_context: "branch fix/discover/search-bounds, issue #118"
files_or_sections_affected: ["db/queries/search.sql", "src/api/internal/platform/dbgen/search.sql.go", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/platform/cursor/cursor.go", "src/api/internal/platform/cursor/cursor_test.go", "src/api/internal/search/repository/repository_test.go", "src/api/internal/search/transport/handler_test.go", "src/api/internal/platform/pgtest/search_read_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-13-search-bounds.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["NICE-DEV226 (pending)"]
tests_or_evidence: ["RED: TestSearchQueriesUseTheGINIndexes scanned 6000 rows for 'bruit' and matched every row for '-zebra'; TestCursorLengthDoesNotGrowWithTheFilters produced a 1222-character cursor", "GREEN: see the PR evidence table"]
open_risks: ["above 500 matches the ranked set is the first 500 rows the index returns, not the 500 best", "a query with only negations now returns no hits instead of every row", "cursors issued before the deploy are refused once (INVALID_CURSOR)"]
```
