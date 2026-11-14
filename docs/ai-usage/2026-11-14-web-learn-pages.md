## S7 learn pages: apprendre, article, formation

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "The apprendre, article and formation pages on the merged S4 API (listLearnArticles with the topic filter, listTrainingSeries, getLearnArticle, getTrainingSeries), the « Poser une question sur cette leçon » action to /solve?link=article:<id>, and the ecrire success link to the new article page."
repository_or_source_context: "branch feature/learn/web-learn-pages; issue #109 section 5; TODO S4 known deviations and S7 page map; prototypes/html/pages/{apprendre,article,formation}.html"
files_or_sections_affected: ["src/web/src/features/learn/**", "src/web/src/app/router.tsx", "src/web/src/app/navigation.ts", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (S7 reviewer, pending)"]
tests_or_evidence: ["RED: vitest src/features/learn 23 of 46 failing before the pages existed", "GREEN: lint, format:check, typecheck, coverage exit 0 (362 tests; 96.74/90.52/96.47/97.79)"]
open_risks: ["/build/<id> and /profils/<username> links resolve once the build and profile pages merge", "axe, visual and responsive checks NOT RUN"]
```
