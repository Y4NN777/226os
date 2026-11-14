## S7 live home page, resource error report and E4 completion

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead + web-implementer (Sonnet)"
purpose: "The accueil sections on live data (listLearnArticles, listQuestions, listResources, listProjects) with the fixtures removed, the « Signaler une erreur » form on the resource page (createResourceErrorReport), the ask-about-the-lesson button on the series page, the read-only questions-by-topic-tag feed on apprendre, and the E4 tests (pre-fill, link sent by createQuestion, anonymous visitor)."
repository_or_source_context: "branch feature/web/live-home-and-e4; TODO S7 page map and E4; api/openapi.yaml read-only"
files_or_sections_affected: ["src/web/src/features/home/**", "src/web/src/features/discover/**", "src/web/src/features/learn/**", "src/web/src/features/solve/AskAboutLesson.test.tsx", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (S7 reviewer, pending)"]
tests_or_evidence: ["RED before each item: failing Vitest tests (home 5, error report 7, series 1, topic feed 3); the E4 tests are characterisation tests of merged behaviour and passed at first run", "GREEN: see the gate results in the pull request"]
open_risks: ["error report: the contract has no duplicate code for createResourceErrorReport, so there is no duplicate message", "axe, visual and responsive checks NOT RUN"]
```
