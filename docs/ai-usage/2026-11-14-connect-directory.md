## E5: Connect people directory (contract and backend)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead + go-implementer (Sonnet)"
purpose: "E5 Connect, read-only people directory: the listProfiles contract (GET /profiles, 49 operations, 45 P0), the authorization matrix row, the public route entry, and the backend in the identity module (visibility-filtered SQL projection and filters, keyset cursor, rate limit) with unit, handler and real-PostgreSQL privacy tests."
repository_or_source_context: "branch feature/connect/people-directory; TODO.md E5; ROADMAP E5 and C14b; ADR-0003, ADR-0004, ADR-0005"
files_or_sections_affected: ["api/openapi.yaml", "db/queries/identity.sql", "src/api/internal/identity/**", "src/api/internal/app/**", "src/api/internal/platform/{apigen,apispec,dbgen,httpx,pgtest}/**", "src/web/src/api/schema.d.ts", "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "docs/spec-kit/AUTHORIZATION_MATRIX.md", "docs/quality/COVERAGE_SCOPE.md", "TODO.md", "ROADMAP.md", "docs/SPEC.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code and text, no new dependency"
closure: "open"
human_reviewers: ["mibienpanjoe (backend)", "Y4NN777 (contract and matrix, R7)"]
tests_or_evidence: ["RED: 20 service tests failed on a stub before the implementation (assertion failures)", "GREEN: go test -race ./... exit 0; scripts/coverage-go.sh 100.0% (6328/6328)", "go test -race -tags integration ./src/api/internal/platform/pgtest -run TestDirectory exit 0, EXPLAIN plans logged"]
open_risks: ["The 60 per minute window is not in ADR-0003: the owner confirms it", "Demo (provenance demo) accounts are not excluded from the directory outside local-dev: the owner decides (K39)"]
```
