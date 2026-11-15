## Control fidelity: heights, search fields, phone rows, breadcrumbs and recherche match the prototypes

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "An independent screen judge listed the five largest visual gaps between the app and the HTML prototypes. The AI fixed them: Astryx control sizes set once in the theme tokens (44 px default, 36 small, 52 large, as the prototype), search icon inside the field on solve and discover (no separate button), small dark field labels, the language tag kept in the top-right corner of solve rows on a phone, the four solve states fitting a 375 px strip, a back link to the parent in place of the wrapping breadcrumb on a phone, tinted square vote buttons, the recherche start view with topic chips and a placeholder, and the signed-out sign-in row aligned to the start edge on a phone. RED tests first for the behaviour changes."
repository_or_source_context: "branch feature/web/fidelity-controls; S7 web pages; prototypes/html/assets/css/{tokens,components,layout}.css and the pages solve, question, recherche, connexion, discover, build, apprendre"
files_or_sections_affected: ["src/web/src/theme/theme226os.ts", "src/web/src/styles/ui.css", "src/web/src/styles/shell.css", "src/web/src/components/Breadcrumbs.tsx", "src/web/src/components/List.tsx", "src/web/src/features/solve/SolvePage.tsx", "src/web/src/features/solve/QuestionRow.tsx", "src/web/src/features/solve/Interactions.tsx", "src/web/src/features/solve/solve.css", "src/web/src/features/discover/DiscoverPage.tsx", "src/web/src/features/learn/LearnPage.tsx", "src/web/src/features/search/SearchPage.tsx", "src/web/src/features/search/search.css", "src/web/src/features/search/messages.ts", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-controls.md"]
third_party_sources_consulted: ["@astryxdesign/core installed sources (SizeContext, theme tokens, TextInput, Button, Icon registry), read from node_modules"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED first: 5 new tests failed (phone back link, corner tag class, recherche topics, solve and discover without a search button) before the code", "changed tests: the solve and discover search tests no longer click a search button (it is gone, Enter submits); the theme test gained the three control sizes", "captures of prototype and app at 1280 and 375 px in the session scratchpad (compare-after)", "pnpm typecheck, lint, format:check and test --maxWorkers=1 in src/web"]
open_risks: ["the control sizes change every Astryx button, input, selector and tab, so pages not captured (settings, learn forms, connect) are unchecked", "the solve search form has a hidden submit button so Enter works from the tag field; there is no visible submit", "topic chips are fixed demo tag slugs, not served data"]
```
