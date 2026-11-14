## S7 solve and question pages

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "The solve and question pages on the live S3 PR 1 operations: list with search and tag filter, ask form (Idempotency-Key per submission, one snippet, pre-filled link only), question with answers (accepted first), answer form, owner edit with If-Match and 412 reload; vote and accept controls hidden for the session on 501 until S3 PR 2. Also the Vitest timeout raised to 15 s (owner decision)."
repository_or_source_context: "branch feature/solve/web-solve-question; TODO S7 page map and S3 deviations; prototypes/html/pages/solve.html and question.html for layout and copy"
files_or_sections_affected: ["src/web/src/features/solve/**", "src/web/src/app/router.tsx", "src/web/src/app/navigation.ts", "src/web/src/app/Shell.test.tsx", "src/web/src/i18n/messages.ts", "src/web/vite.config.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (S7 reviewer, pending)"]
tests_or_evidence: ["RED: 30 failing tests before the routes existed", "GREEN: pnpm lint, format:check, typecheck, coverage exit 0 (301 tests; 97.17/90.92/96.95/98.07)"]
open_risks: ["the home feed's question items are demo fixtures and do not link to the question page yet", "the edit form covers title, body and tags only"]
```
