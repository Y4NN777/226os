## Landing talents across West Africa, demo captures

```yaml
date: "2026-11-18"
contributor: "Y4NN777"
tool_or_model: "Claude Code (Claude Opus 5.5, web-implementer run driven by the lead session)"
purpose: "The AI widened the landing Talents query to every imported West African country (no country filter), added the country name next to each rank, pointed the full ranking link to /connect?source=github, updated the tests first, and retook the landing device mockups on local-dev demo content."
repository_or_source_context: "branch feature/web/talents-west-africa, from develop 258f472"
files_or_sections_affected: ["src/web/src/features/landing/**", "src/web/public/landing/mock-*.jpg", "src/web/public/landing/CREDITS.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency; the mockups are captures of the project's own pages on demo seed content"
closure: "open"
human_reviewers: []
tests_or_evidence: ["vitest src/features/landing: 3 RED then 41/41 GREEN", "web gates and repository checks listed in the PR", "landing captures at 1440 and 390 px with a mocked /profiles answer"]
open_risks: ["the Talents captures use a mocked /profiles answer, not imported committers.top data", "a login ranked in two countries can appear twice (ADR-0016, accepted)"]
```
