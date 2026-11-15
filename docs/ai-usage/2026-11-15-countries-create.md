## All 54 countries and optional fields on the project creation form

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, web-implementer subagent (Sonnet 5.5)"
purpose: "The Build country filter and the project forms offered 16 West African codes plus the codes already seen, while the database accepts the 54 African countries. One shared list (West Africa first, then the rest, names from the platform) now serves Build, Connect and Discover; the creation form gets the optional stage, licence and country, reusing the owner edit fields, with 422 on /stage, /license_spdx and /country_code shown on the fields."
repository_or_source_context: "branch feature/web/countries-create, PR #246"
files_or_sections_affected: ["src/web/src/components/countries.ts", "src/web/src/features/build/ProjectFields.tsx", "src/web/src/features/build/ProjectForm.tsx", "src/web/src/features/build/ProjectPage.tsx", "src/web/src/features/build/BuildPage.tsx", "src/web/src/features/connect/ConnectPage.tsx", "src/web/src/features/discover/DiscoverPage.tsx"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency added"
closure: "open"
human_reviewers: []
tests_or_evidence: ["vitest run --maxWorkers=1 on src/features/build, connect, discover and src/components/countries.test.ts: 13 files, 177 tests PASS; typecheck, lint, format:check PASS"]
open_risks: ["the 54 codes are copied from db/migrations/00002_reference_taxonomy.sql, not read from the API"]
```
