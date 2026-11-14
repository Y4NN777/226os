## S9 assistant panel in recherche

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "The assistant panel in recherche on the live askAssistant API: a visible notice that the question is sent to Google before the first question, the answer as plain text with an AI label, cited sources and Comment vérifier, no vote or accept control, a refusal with a link to search, the 429 wait, and the panel hidden after a 503; the demo sample and its fixture are removed."
repository_or_source_context: "branch feature/intelligence/web-assistant-panel; TODO S9 block; ADR-0012; PR #161 (backend)"
files_or_sections_affected: ["src/web/src/features/search/**", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (S7 reviewer, pending)"]
tests_or_evidence: ["RED: 10 AssistantPanel tests failed against the demo panel", "GREEN: pnpm lint, format:check, typecheck, coverage exit 0 (338 tests; 96.76/90.2/96.77/97.77)"]
open_risks: ["the solve ask-flow panel is not built", "the real provider stays NOT RUN (V2, V4); the assistant is off until AI_ENABLED is set"]
```
