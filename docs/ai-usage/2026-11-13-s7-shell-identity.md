## S7 frontend PR 1: shell and identity pages

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (integrator) and Claude Sonnet 5.5 (web-implementer lanes in separate worktrees)"
purpose: "S7 PR 1: CSRF-aware API helper (X-CSRF-Token on every unsafe call, token in memory only, one retry after CSRF_FAILED), problem+json mapping, session hooks, Astryx AppShell with pillar emphasis and coming-soon areas, shared states, and the six identity pages wired to the live S1 operations."
repository_or_source_context: "branch feature/platform/web-shell-identity; lane branch feature/identity/web-account-pages; prototypes/html for layout and copy"
files_or_sections_affected: ["src/web/src/**", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["installed @astryxdesign/cli 0.7.14 component docs", "openapi-fetch 0.17 type declarations"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (pending)"]
tests_or_evidence: ["RED then GREEN per file with Vitest (api/http.test.ts, app/Shell.test.tsx, features/identity/*.test.tsx)", "pnpm --filter web run lint/format:check/typecheck/test/coverage exit 0; coverage 96.94/87.29/98.75/98.38"]
open_risks: ["Visual, axe and responsive evidence (ADR-0007) not produced in jsdom.", "Mails carry a bare token until APP_BASE_URL links land (fix/platform/s7-api-gaps)."]
```
