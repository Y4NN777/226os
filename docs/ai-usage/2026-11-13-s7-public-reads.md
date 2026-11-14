## S7 frontend PR 2: public reads (Découvrir, resource, recherche)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (integrator) and Claude Sonnet 5.5 (web-implementer lanes in separate worktrees)"
purpose: "S7 PR 2: Découvrir with the West Africa default and one-click whole-continent view, resource detail with dated sources and verification status, grouped search with a marked assistant mock; Découvrir becomes a live pillar."
repository_or_source_context: "branch feature/discover/web-public-reads; lane branches feature/discover/web-catalogue and feature/discover/web-search; prototypes/html for layout and copy"
files_or_sections_affected: ["src/web/src/features/discover/**", "src/web/src/features/search/**", "src/web/src/app/{navigation,Shell}.tsx", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["installed @astryxdesign/cli 0.7.14 component docs"]
license_review: "NOT APPLICABLE: original code and text, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (pending)"]
tests_or_evidence: ["RED then GREEN per file (region.test.ts, DiscoverPage.test.tsx, ResourcePage.test.tsx, SearchPage.test.tsx, searchQuery.test.ts)", "pnpm --filter web run coverage exit 0"]
open_risks: ["The assistant answer is a labelled sample until S9 is live.", "West Africa default filters on the client until the API accepts several countries."]
```
