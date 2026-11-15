## Typo-tolerant search (issue #248, item 2)

```yaml
date: "2026-11-15"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Issue #248 item 2. The AI wrote the failing tests first, then: four bounded Similar* queries in db/queries/search.sql (pg_trgm word similarity over the partial trigram indexes of migration 00010, same predicates and filters as full text, negation-only queries read nothing), Repository.Similar, a service fallback when a group's full text finds nothing (GET /search only, 3 characters at least; the assistant's retrieval keeps exact matches), EXPLAIN evidence, and the SPEC, ARCHITECTURE and DATABASE_DESIGN text. It measured word_similarity on sample typos to keep the default 0.6 threshold."
repository_or_source_context: "branch feature/discover/search-typo-tolerance, from develop 25774f2"
files_or_sections_affected: ["db/queries/search.sql", "src/api/internal/platform/dbgen/search.sql.go", "src/api/internal/search/**", "src/api/internal/platform/pgtest/search_typo_integration_test.go", "src/api/internal/platform/pgtest/explain_helpers_integration_test.go", "docs/SPEC.md", "docs/architecture/ARCHITECTURE.md", "docs/spec-kit/DATABASE_DESIGN.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["unit tests RED first; TestSearchToleratesTypos fails with the fallback off (12 group mismatches) and passes with it; plans logged by TestSearchTypoFallbackUsesTheTrigramIndexes; see the pull request for the gate results"]
open_risks: ["swapped letters (pyhton) stay below the 0.6 threshold", "only titles and names are typo tolerant, not bodies"]
```
