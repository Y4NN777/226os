# 226os AI Use and Provenance Record

**Status:** DRAFT; human and CADEV compliance review required

## Policy

AI output is an untrusted proposal. Human owners remain responsible for requirements, architecture, security, tests, licenses, and release claims. Rewriting, translating, or generating code with AI does not remove third-party copyright or license obligations.

Do not submit secrets, credentials, private user data, unpublished tokens, or confidential external material to AI systems. Do not use AI output as evidence that a command, test, source review, performance measurement, or security check occurred.

## Required record for each material AI-assisted contribution

```yaml
date: "YYYY-MM-DD"
contributor: ""
tool_or_model: ""
purpose: ""
repository_or_source_context: ""
files_or_sections_affected: []
third_party_sources_consulted: []
license_review: PENDING
human_reviewers: []
tests_or_evidence: []
open_risks: []
```

## Current task record

- Task: documentation audit and G0 Spec Kit preparation.
- AI assistance: repository reading, bounded specialist review, and Markdown drafting.
- Application implementation: none.
- Commands/builds/tests/deployments: none in the audit phase; Git publication was separately authorized and recorded outside application evidence.
- Third-party code integrated: none observed in this documentation set.
- Compliance status: `PENDING` human/CADEV review.

## H0 Claude runtime harness record

```yaml
date: "2026-11-11"
contributor: "Y4NN77"
tool_or_model: "Claude Code 2.1.322, Claude Opus 5.5 (validation probes on Haiku)"
purpose: "Add and validate the .claude/ runtime harness; reconcile main into develop history"
repository_or_source_context: "PROJECT_HARNESS_SPEC.md, AGENTS.md, CLAUDE.md, docs/GIT_WORKFLOW.md"
files_or_sections_affected: [".claude/settings.json", ".claude/agents/*.md", ".claude/skills/*/SKILL.md", "PROJECT_HARNESS_SPEC.md", "docs/spec-kit/CLAUDE_HANDOFF.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: PENDING
human_reviewers: []
tests_or_evidence:
  - "settings.json deny rule enforced live: a write to .env.probe was blocked by the harness"
  - "claude -p --output-format stream-json init event: agents [advisor, go-implementer, scout, skeptic, web-implementer]; skills/slash_commands [advisor, ship-check, spec-to-task, start-slice, verify-slice]"
  - "claude -p '/spec-to-task US-NOPE-999': skill invoked, stopped at step 1, no files changed"
  - "/doctor, /permissions, /agents interactive checks: NOT RUN"
open_risks: ["provider access to each routed model not verified"]
```

## Domain skills and hooks record

```yaml
date: "2026-11-11"
contributor: "Y4NN77"
tool_or_model: "Claude Code 2.1.322, Claude Opus 5.5 (orchestration), general-purpose research subagents, Haiku validation probes"
purpose: "Add ten domain skills, the postgres-expert agent, and the domain-skill hooks"
repository_or_source_context: "AGENTS.md, DESIGN.md, docs/{DEV,TEST,RULES,SPEC,SHIP,GIT_WORKFLOW}.md, docs/spec-kit/*, docs/decisions/*, docs/quality/COVERAGE_SCOPE.md"
files_or_sections_affected: [".claude/skills/<10 domain skills>/**", ".claude/agents/*.md", ".claude/hooks/**", ".claude/settings.json", "CLAUDE.md", "PROJECT_HARNESS_SPEC.md"]
third_party_sources_consulted: ["Official docs cited inline in each SKILL.md (Go, Chi, pgx, sqlc, PostgreSQL, OpenAPI 3.1, RFC 9457, OWASP, React, TanStack, WCAG, Playwright, testcontainers-go, GitHub Actions, Docker, Traefik)"]
license_review: PENDING
human_reviewers: []
tests_or_evidence:
  - "python3 -m unittest discover -s .claude/hooks/tests: 26 tests OK"
  - "claude -p probe A (first design): first Write to apps/api/ denied by the hook; no file created"
  - "claude -p probe D: Read of apps/api/main.go received the hook context naming engineering-principles, go-backend"
  - "claude -p probe E: spawned skeptic subagent received only go-backend (its preloads were recognised via agent_type); hook output over ~2 KB was truncated to a preview, which led to the Skill-tool design"
  - "claude -p probe G (final design): Write to apps/api/ denied with the directive; the model loaded engineering-principles and go-backend via the Skill tool; the hook recorded both; the retry was not observed live (covered by unit tests); no file created"
  - "claude -p probe C: all 10 domain skills listed at init; Skill tool invocations recorded by the PostToolUse hook; the subsequent Write passed the hook (probe file then removed)"
  - "claude -p spawning postgres-expert: subagent reported exactly its 5 preloaded skill headings"
open_risks: ["Bash-based file writes bypass the hook", "an Edit of a non-existent file is rejected by Claude Code before PreToolUse fires (harmless: no write happens)", "skills cite documents with known internal conflicts (listed in the PR description)"]
```
