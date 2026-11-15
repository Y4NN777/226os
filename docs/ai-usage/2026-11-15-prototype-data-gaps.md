## API data gaps for the prototype pages

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and the go-implementer subagent"
purpose: "Serve the data the HTML prototypes show and the API did not (owner decision 2026-11-15 ~07:00 GMT, R7 for these items, no migration): Project stage, license_spdx and country_code on reads and writes with list filters; a total on listResources, listProjects, listQuestions, listLearnArticles and listTrainingSeries; sort on listResources (name, verified), listProjects (updated, created, needs) and listQuestions (new, votes, unanswered) bound to the cursor; the profile activity filters owner, author and answers_by under the getProfile visibility rule."
repository_or_source_context: "branch feature/api/prototype-data-gaps from origin/develop; ADR-0004 (cursors, problems), ADR-0005 (visibility), migration 00007 (projects columns), prototypes/html/pages/{discover,build,project,profile,apprendre,solve}.html for the sort options and counters"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apigen/apigen.gen.go", "src/api/internal/platform/apispec/openapi.gen.yaml", "src/web/src/api/schema.d.ts", "db/queries/{build,discover,learn,solve}.sql", "src/api/internal/platform/dbgen/**", "src/api/internal/{build,discover,learn,solve}/**", "src/api/internal/platform/pgtest/{project_facts,list_totals,list_sort,profile_activity}_integration_test.go", "docs/SPEC.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["Y4NN777 (owner, pending)"]
tests_or_evidence: ["RED then GREEN per layer (repository, service, transport) for each item; dbgen bind-order tests for every changed query", "real-PostgreSQL integration: TestBuildProjectFacts, TestListTotals*, TestListSort*, TestProfileActivity* (walk every page, total equals the items returned, cursor of another sort is 400 INVALID_CURSOR, hidden or private username is an empty page)", "EXPLAIN (ANALYZE, BUFFERS) on 3000 projects and 3000 questions: every list and count under 8 ms, in shared buffers (lead report)"]
open_risks: ["no index serves the new orders (projects updated_at, questions vote_score, resources verified_at) nor the open-needs order: sequential scan plus top-N sort, fine for P0 volumes; an index needs a migration (mibienpanjoe)", "the listProjects default order moved from published_at to updated_at (owner decision)", "the web typings now require total, stage, license_spdx and country_code: web fixtures may need the new fields"]
```
