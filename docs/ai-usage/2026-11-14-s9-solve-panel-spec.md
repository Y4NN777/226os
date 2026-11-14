## S9: assistant panel in solve, SPEC story and matrix row

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session), Claude Sonnet 5.5 (web-implementer subagent) and the lead agent (documentation)"
purpose: "The assistant panel in the solve ask flow (shared with recherche: the draft title and body as the question, disabled under 3 characters, posting stays separate; recherche no longer sends a query under 3 characters), plus the S9 SPEC story US-INTEL-001 (acceptance criteria and BDD scenarios) and the exact askAssistant row of the authorization matrix (H13, for the owner's review)."
repository_or_source_context: "branch feature/intelligence/web-solve-assistant; issue #73; ADR-0012; PR #169 review note"
files_or_sections_affected: ["src/web/src/features/search/**", "src/web/src/features/solve/**", "docs/SPEC.md", "docs/spec-kit/AUTHORIZATION_MATRIX.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code and text, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (S7 reviewer, pending)", "Y4NN777 (SPEC and matrix, H13)"]
tests_or_evidence: ["RED: 6 tests failed before the code (recherche short query; solve notice, request, no_source, 503, English)", "GREEN: pnpm lint, format:check, typecheck, coverage exit 0 (345 tests; 96.72/90.29/96.79/97.78)"]
open_risks: ["SPEC owner items: the disabled and limiter-failure 503 send no Retry-After although TODO S9 says 30; AI-3 and AI-5 are enforced by the prompt only; published_at never filled; series in the source enum never retrieved; ADR-0012 still PROPOSED"]
```
