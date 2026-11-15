## Data gaps wiring: show the data PR #238 serves (S7)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, web-implementer subagent of the lead session"
purpose: "The merged API (PR #238) now serves project stage, licence and country, a required total on every page, sort and filter parameters, and username filters. The AI wired them into the web pages the way the prototypes show them: Build rows and project page (stage, country, licence, stage and country filters, sort select, owner edit of stage, licence and country with 422 mapped to fields), Solve and Discover (sort select, result count from the total), Learn (count from the totals, no more « N premiers »), and the profile activity tabs (Questions, Réponses, Projets). Tests were written first (RED) for every visible element and URL parameter. Search keeps its per-group counts: /search serves no total."
repository_or_source_context: "branch feature/web/data-gaps-wiring (stacked on feature/web/prototype-fidelity, PR #246); S7 web; prototypes/html as visual reference only (DISCLAIMER.md)"
files_or_sections_affected:
  [
    "src/web/src/components/SortSelect.tsx",
    "src/web/src/app/router.tsx",
    "src/web/src/features/build/**",
    "src/web/src/features/solve/**",
    "src/web/src/features/discover/**",
    "src/web/src/features/learn/**",
    "src/web/src/features/profile/**",
    "TODO.md",
    "docs/ai-usage/2026-11-15-data-gaps-wiring.md",
  ]
third_party_sources_consulted:
  [
    "@astryxdesign/core installed sources and docs (Selector, TabList, Tab) read from node_modules",
  ]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence:
  [
    "RED first: 43 new tests failed before the code (route search validation, row meta, counts from the total, sort and filter parameters in the URL and the request, restart at the first page, owner edit with null clearing and 422 mapping, profile tabs)",
    "changed tests: the Discover and Learn count tests now expect the API total instead of « Les N premiers »; page fixtures carry a total",
    "targeted vitest runs and pnpm typecheck, lint, format:check: see the PR",
  ]
open_risks:
  [
    "the country list of the filters and of the owner form is West Africa plus the codes already seen, not every country the API knows",
    "the project creation form does not yet offer stage, licence or country",
    "no captures: the shared local-dev API was built before PR #238 and was not restarted",
    "axe and screen reader NOT RUN on the new controls and tabs",
  ]
```
