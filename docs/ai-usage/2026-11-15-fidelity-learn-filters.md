## Page fidelity: the apprendre filter area

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked to close every remaining gap with the HTML prototypes. For /apprendre the AI wrote RED tests first, then added the prototype filter area inside features/learn: type chips (Astryx SegmentedControl), a text search, the Sujet, Niveau and Langue selects, a results line, one list (series rows, then article rows) and a reset, all kept in the URL. Only the filters the contract serves are offered: listLearnArticles takes q, type, topic, level and lang; listTrainingSeries takes topic, level and lang only."
repository_or_source_context: "branch feature/web/fidelity-learn-filters from the integration branch 341d497; S7 web pages; prototypes/html/pages/apprendre.html and assets/js/pages/learn.js"
files_or_sections_affected: ["src/web/src/features/learn/LearnPage.tsx", "src/web/src/features/learn/LearnPage.test.tsx", "src/web/src/features/learn/queries.ts", "src/web/src/features/learn/routeSearch.ts", "src/web/src/features/learn/messages.ts", "src/web/src/features/learn/learn.css", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-learn-filters.md"]
third_party_sources_consulted: ["@astryxdesign/core installed sources and docs (SegmentedControl, Selector, TextInput, VisuallyHidden), read from node_modules"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: RED tests first (17 new or changed tests failing), then the code", "new tests: chips, search and selects in one area with no chip counts; each filter sends its query parameter to the endpoint that serves it and not to the other; URL read and write; unknown values dropped; one reset clears every filter; a filter change restarts at the first page (no cursor); empty state per filter and per type; English copy", "changed tests (owner fidelity decision): the reset is « Réinitialiser » instead of « Tous les sujets »; the featured series is listed again in the list, as in the prototype; the unfiltered empty state speaks of « contenu »", "fixed on the way: the topic question feed used solve.state.* keys removed by the solve pass (typecheck and two tests failed on 341d497)", "captures of prototype and app at 1280 and 375 px in the session scratchpad", "pnpm typecheck, lint and format:check exit 0; full test run with --maxWorkers=1"]
open_risks: ["the series endpoint has no text search: while q is set the series rows are left out and a note says so; under Formations the search field is hidden", "two cursors cannot be merged honestly, so the list shows series first, then articles, each with its own « Afficher plus » button", "Astryx field labels are larger than the prototype's small muted labels", "at 375 px the four segments scroll inside their strip instead of wrapping like the prototype chips"]
```
