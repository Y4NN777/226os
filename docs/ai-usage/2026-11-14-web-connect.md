## E5 connect page (people directory)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "The connect page on the merged listProfiles API (#193): filters, paging, states, profile links and visibility-aware fields, built to help the E5 UI owner PANK4SS before the freeze."
repository_or_source_context: "branch feature/connect/web-connect; TODO E5 block; issue #77; prototypes/html/pages/connect.html"
files_or_sections_affected: ["src/web/src/features/connect/**", "src/web/src/app/router.tsx", "src/web/src/app/navigation.ts", "src/web/src/app/Shell.test.tsx", "src/web/src/features/home/**", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (E5 UI owner, pending)"]
tests_or_evidence: ["RED: connect, Shell and Home tests failing before the route existed", "GREEN: lint, format:check, typecheck, coverage exit 0 (476 tests; 95.97/89.2/95.63/97.01)", "#214 follow-up: githubProfileUrl helper, RED then GREEN in src/features/connect and src/features/profile"]
open_risks: ["not exercised against the real backend", "axe, visual and responsive checks NOT RUN"]
```
