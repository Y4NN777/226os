## parametres: the display name is saved on the account

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "First #109 follow-up, owner decision: parametres kept the display name only in browser storage and never called the API. The form now loads it with getMe and saves it with updateMyProfile through the CSRF-aware helper, shows server violations on the field, and the old storage key is removed."
repository_or_source_context: "branch feature/identity/web-settings-profile; issue #109; a note from the independent review of PR #100"
files_or_sections_affected: ["src/web/src/features/settings/**", "src/web/src/app/storage.ts", "src/web/src/main.tsx", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (S7 reviewer, pending)"]
tests_or_evidence: ["RED: 7 tests failed first (load from the API, sign-in invite, PATCH with CSRF and JSON body, header refresh with no storage copy, server violation on the field, legacy key removal, storage unavailable)", "GREEN: pnpm lint, format:check, typecheck, coverage exit 0 (97.63/91.59/98.94/98.83)"]
open_risks: ["the client keeps a 2-character minimum while the contract allows 1", "the start-up cleanup call in main.tsx has no test (the function it calls is tested)"]
```
