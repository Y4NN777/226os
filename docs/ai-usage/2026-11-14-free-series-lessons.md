## Free Learn series: the four lessons rewritten for engineers (#125)

```yaml
date: "2026-11-14"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 then Claude Fable 5.2 (main session, no subagents)"
purpose: "Issue #125: the four lessons of scripts/content/free-series.sql were judged outdated and too shallow for an engineering audience. The contributor set the direction (address engineers; cover harnesses, agentic systems, always-on agents and models that are no longer text-only). AI-drafted: the full French text of the four lessons, the series title and summary, the four lesson titles and summaries. Human part: the direction above, then a full read and approval of the text by the contributor, without line edits. Structure changes: series retitled 'Des LLM aux agents : l'ingénierie de l'IA aujourd'hui' (the idempotency guard follows the title), level beginner -> intermediate, tags ia/llm/prompt/debutant -> ia/llm/agents/harness. Still 4 tutorial lessons, language fr, topic ai."
repository_or_source_context: "branch docs/learn/free-series-lessons; issue #125 (from PR #111, S4 Learn)"
files_or_sections_affected: ["scripts/content/free-series.sql", "TODO.md", "docs/ai-usage/2026-11-14-free-series-lessons.md"]
third_party_sources_consulted: []
license_review: "Original text written for 226os; no passage copied or translated from a third-party source. Vendor-neutral: no product or model named, except the open protocol MCP and the AGENTS.md file convention."
closure: "open"
human_reviewers: ["mibienpanjoe (read and approved the text, 2026-11-14)", "pending: PR review, and the owner's approval of the final text (issue #125)"]
tests_or_evidence: ["scripts/tests/free-series_test.sh 13 passed, 0 failed", "go test -tags integration ./internal/platform/pgtest/ -run TestFreeSeriesContentScript PASS (ephemeral PostgreSQL)", "excalibur run: verified, known debt in forge-check"]
open_risks: ["NOT LOOKED AT: the lessons were not viewed rendered on the article page; they are about four times longer than the previous text and add a fenced code block (lesson 2) and a Markdown table (lesson 4), to check on staging during the rehearsal", "the guard matches the new title: if the previous series was already run on an environment, this script would create a second series there (staging not checked)", "technical statements about models and agents are general and undated; they need a re-read when the field moves"]
```
