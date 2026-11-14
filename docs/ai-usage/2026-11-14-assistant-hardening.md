## S9 assistant hardening: Retry-After, AI-3 enforced, dated sources (AI-5)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (postgres-expert, go-implementer and web-implementer subagents)"
purpose: "The S9 items cut at the 12:00 checkpoint, back in scope by owner decision: Retry-After on every assistant 503, the AI-3 rule enforced on the server, publication dates on assistant sources and in the panels (AI-5), golden cases G14 to G17, and project and article source URLs to their new pages."
repository_or_source_context: "branch feature/intelligence/assistant-hardening; TODO S9 block; ADR-0012; SPEC US-INTEL-001; issue #73"
files_or_sections_affected: ["db/queries/search.sql", "src/api/internal/platform/dbgen/**", "src/api/internal/search/{domain,repository}/**", "src/api/internal/assistant/**", "src/api/internal/platform/pgtest/**", "src/web/src/features/search/**", "docs/decisions/0012-*.md", "TODO.md", "ROADMAP.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["mibienpanjoe (S9 backend reviewer, pending)", "PANK4SS (UI, pending)"]
tests_or_evidence: ["RED then GREEN: TestRemoveGuarantees, TestAskEvery503CarriesRetryAfter, TestUserPromptCarriesEachPassageDate, TestParseVerifyNoticeFollowsTheCitedSources, TestSearchHitsCarryThePublicationDate, golden G14 to G17, AssistantPanel date tests", "Go coverage 100.0 % (6239/6239); lint-go, race unit and integration exit 0"]
open_risks: ["guarantee detection is a FR/EN pattern list: other phrasings pass through", "a bare marker left alone after a removed guarantee keeps its source cited", "the real-provider run (V4) is NOT RUN"]
```
