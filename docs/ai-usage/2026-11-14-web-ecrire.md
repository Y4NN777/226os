## S7 ecrire page (issue #124)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "The ecrire page on createLearnArticle under the three #124 constraints: question, resource and project links only (pre-filled from the route), articles created as published, and a body-omitted idempotent replay that follows Location."
repository_or_source_context: "branch feature/learn/web-ecrire; issue #124; TODO S4 known deviations; prototypes/html/pages/ecrire.html"
files_or_sections_affected: ["src/web/src/features/learn/**", "src/web/src/app/router.tsx", "src/web/src/app/Shell.tsx", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (S7 reviewer, pending)"]
tests_or_evidence: ["RED: 23 tests failing before the route existed", "GREEN: lint, format:check, typecheck, coverage exit 0 (329 tests; 96.75/90.26/96.86/97.79)"]
open_risks: ["the 14 topic labels are a client constant copied from migrations 00002 and 00011 (no endpoint)", "no article page exists yet, so success links to the home page"]
```
