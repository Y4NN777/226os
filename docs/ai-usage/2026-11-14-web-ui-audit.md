## Web UI audit and its safe fixes

```yaml
date: "2026-11-14"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Fable 5.2 (main session, no subagents), Playwright for the screenshots"
purpose: "The contributor asked for an inspection of the web UI (wrong parts, incoherences, details to improve), then chose which findings to fix. The AI captured 20 pages in French at 1280 px and 390 px, signed out and signed in, on a throwaway local stack (temporary PostgreSQL, the API and the Vite dev server, a fictional account), listed 16 findings, and wrote the fixes and tests for the 7 the contributor selected (1, 2, 3, 4, 8, 10, 11). The 9 others are in issue #192. Fixes: phone navigation back in the Astryx drawer (Shell passed a SideNav as mobileNav.content, so it was laid out under the content and the toggle did nothing); not-found page framed in a main landmark; recherche button no longer truncated; landing search field no wider than a phone; sign-in prompt on the question page primary like the others; article body 720 px at most. Finding 10 (uneven phone padding) needed no change of its own: it disappears with the navigation fix."
repository_or_source_context: "branch fix/platform/web-ui-audit; S7 web pages; follow-up issue #192"
files_or_sections_affected: ["src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "src/web/src/app/App.test.tsx", "src/web/src/features/NotFound.tsx", "src/web/src/features/search/SearchPage.tsx", "src/web/src/features/landing/LandingPage.tsx", "src/web/src/features/solve/QuestionPage.tsx", "src/web/src/features/learn/ArticlePage.tsx", "src/web/src/styles/index.css", "TODO.md", "docs/ai-usage/2026-11-14-web-ui-audit.md"]
third_party_sources_consulted: ["@astryxdesign/core 0.7.14 AppShell source (node_modules), read to find how mobileNav is rendered"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: the fixes were written first, the tests after; the two new tests were then run against the previous Shell.tsx and NotFound.tsx and both failed (RED), and pass with the fixes (GREEN)", "Shell.test.tsx: 'keeps the navigation in a drawer on a phone, opened by the toggle'; App.test.tsx: the not-found page sits in a main landmark", "vitest 432 passed (42 files); eslint, prettier --check and tsc --noEmit PASS", "before and after screenshots read at 390 px and 1280 px"]
open_risks: ["the search button, the landing field width, the primary sign-in prompt and the article width have no automated test (layout only, checked on screenshots)", "the open drawer's content is not asserted in jsdom (no modal dialog there); it was checked in the browser", "checked on the Vite dev server, not on the built image behind nginx and Traefik; English pages not re-captured after the fixes"]
```
