## S9 grounded assistant backend

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer and skeptic subagents)"
purpose: "askAssistant over the platform search with the fake and AI Studio adapters: members only, redaction before sending, untrusted passages in a delimited data section, server-side citation validation, no provider call without sources, per-member and global limits failing closed, 503 with Retry-After, a 13-case golden set on the fake adapter, ADR-0012; then the fixes of an independent security review."
repository_or_source_context: "branch feature/intelligence/assistant; TODO S9 block; ROADMAP O5; ADR-0011"
files_or_sections_affected: ["src/api/internal/assistant/**", "src/api/internal/app/** (wiring)", "src/api/internal/platform/pgtest/** (golden set)", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt", "docs/decisions/0012-ai-assistant-integration.md", "docs/decisions/README.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["mibienpanjoe (S9 backend reviewer, pending)"]
tests_or_evidence: ["RED: the wiring test and every review fix (17 redaction cases, nested delimiters, links in output, malformed markers, Retry-After)", "GREEN: make generate, test-unit, coverage (Go 100.0%), test-integration-go, security exit 0"]
open_risks: ["real provider not run (V2, V4)", "unit tests of the first commit written after the code", "no UI panel yet, so the 'sent to Google' notice does not exist"]
```
