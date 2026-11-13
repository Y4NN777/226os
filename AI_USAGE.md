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

Rule on `contributor` (added 2026-11-12): going forward the field holds the GitHub login. `Y4NN77` (the Git author name) and `Y4NN777` (the GitHub login) denote the same person, the repository owner; older records keep the value they were written with and are not rewritten.

## Review and closure status (2026-11-12)

Every record in this file has `license_review: PENDING` and no human CADEV review is recorded yet; `human_reviewers` is empty or names an owner decision or a reviewer role, not an evidenced licence review. Convention for later records: add a `closure` line (`open` or `closed: <date>, <reviewer login>`) once the licence review and human review of that record are done. This note adds the convention only; no record is closed. Original-work statement: `docs/research/REFERENCES.md`, section "Déclaration d'originalité". Dependency inventory: `THIRD_PARTY_NOTICES.md`.

## Current task record

> Dated historical record (2026-11-11, documentation audit and G0 preparation); it no longer describes the repository, see the dated records below.

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
  - "claude -p probe A (first design): first Write to src/api/ denied by the hook; no file created"
  - "claude -p probe D: Read of src/api/main.go received the hook context naming engineering-principles, go-backend"
  - "claude -p probe E: spawned skeptic subagent received only go-backend (its preloads were recognised via agent_type); hook output over ~2 KB was truncated to a preview, which led to the Skill-tool design"
  - "claude -p probe G (final design): Write to src/api/ denied with the directive; the model loaded engineering-principles and go-backend via the Skill tool; the hook recorded both; the retry was not observed live (covered by unit tests); no file created"
  - "claude -p probe C: all 10 domain skills listed at init; Skill tool invocations recorded by the PostToolUse hook; the subsequent Write passed the hook (probe file then removed)"
  - "claude -p spawning postgres-expert: subagent reported exactly its 5 preloaded skill headings"
open_risks: ["Bash-based file writes bypass the hook", "an Edit of a non-existent file is rejected by Claude Code before PreToolUse fires (harmless: no write happens)", "skills cite documents with known internal conflicts (listed in the PR description)"]
```

## Brand assets (concept D) record

```yaml
date: "2026-11-11"
contributor: "Y4NN77"
tool_or_model: "Google Gemini (image generation, by the repository owner); ImageMagick resizing by Claude Code"
purpose: "226os logo, symbol, app icons; README header and social preview; prototype header and favicons"
repository_or_source_context: "Owner-provided ~/Downloads/226os-concept-d-assets.zip (transparent PNG extracted from an approved four-concept board); Gemini_Generated_Image_*.jpeg not used (checkerboard baked into a JPEG, no real transparency)"
files_or_sections_affected: ["docs/brand/226os-logo.png", "docs/brand/226os-logo-white.png", "docs/brand/226os-social-preview.png", "prototypes/html/assets/brand/*", "README.md header", "prototypes/html/assets/js/shell.js brand"]
third_party_sources_consulted: []
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["visual check of every exported PNG", "prototype header rendered in light and dark"]
open_risks:
  - "AI-generated raster, not vector: approximate edges, needs review at 16/32 px; a clean SVG redraw is recommended before release"
  - "Copyright and trademark status of AI-generated marks must be reviewed with CADEV/legal (G0-COMP-001)"
```

## Slice B0 bootstrap record

```yaml
date: "2026-11-12"
contributor: "Y4NN77"
tool_or_model: "Claude Code (orchestrator session) with Claude subagents: postgres-expert, go-implementer, web-implementer, general-purpose, lead (taxonomy research) and proof:verifier (independent review before each commit)"
purpose: "Slice B0: Go API skeleton, goose migrations 00001-00010 with sqlc and a PostgreSQL 18 harness, OpenAPI code generation (Go and TypeScript), web skeleton with Astryx, Compose profiles and non-root images, Makefile (mk/*.mk) and quality workflow"
repository_or_source_context: "this repository at develop 0372648; owner-approved dependency installs (Go modules, pnpm packages, Docker images)"
files_or_sections_affected: ["go.mod", "go.sum", "src/api/**", "src/web/**", "db/**", "api/oapi-codegen.yaml", "package.json", "pnpm-workspace.yaml", "pnpm-lock.yaml", "compose.yaml", "deploy/**", "Makefile", "mk/**", "scripts/**", ".github/workflows/quality.yml", ".env.example", ".golangci.yml", "docs/quality/COVERAGE_SCOPE.md"]
third_party_sources_consulted:
  - "reference taxonomy research for resource_categories and learn_topics: web sources listed in the PR description (GSMA, OpenUK, KRA eTIMS, TechCabal, and others); figures labelled secondary were not opened by the reviewer"
license_review: PENDING
human_reviewers: ["Y4NN77 (owner decisions: 54-country list, 15 resource categories, 9 learn topics, coverage exclusion for apigen.gen.go, Node 26, licence allowlist, Semgrep)"]
tests_or_evidence: ["make quality exit 0 on the B0 tip (reported by proof:verifier and the unit agents)", "per-unit RED/GREEN records in the unit reports and PR description"]
open_risks:
  - "country names in migration 00002 were transcribed by the agent, not checked against an official ISO or UN source (codes were cross-checked against tzdata)"
  - "taxonomy lists are research proposals approved by the owner, not user-tested"
  - "ADR-0007 visual, axe and responsive evidence is NOT RUN; G5 to G7 NOT RUN"
```

## Product direction documentation record

```yaml
date: "2026-11-12"
contributor: "Y4NN77"
tool_or_model: "Claude Code (orchestrator session) with a Claude subagent (lead) for the documentation sync; a Claude research subagent for the agentic-era topics and the reference study; the grounded assistant planned in S9 will call Google AI Studio (Gemma) at runtime"
purpose: "Align README, PRD amendment 3.2 (proposed), SPEC stories, DATABASE_DESIGN notes, the reference study for 225 Open Source and the prototype disclaimer with the product decisions of 2026-11-12"
repository_or_source_context: "this repository at develop c55fbde; the public home page text of 225os.com was read by the owner and paraphrased in the study; no text, visuals or data from it are copied"
files_or_sections_affected: ["README.md", "docs/PRD.md", "docs/SPEC.md", "docs/spec-kit/DATABASE_DESIGN.md", "docs/research/REFERENCES.md", "prototypes/html/DISCLAIMER.md", "prototypes/html/README.md", "docs/screenshots (removed)", "docs/team/KICKOFF.md", "ROADMAP.md", "TODO.md"]
third_party_sources_consulted:
  - "Stack Overflow Developer Survey 2025, Anthropic and PNAS studies on AI assistance and learning, GSMA 2026 mobile money report, Google AI Studio terms (aistudio.google.com/docs/terms): used as research input for the plan; figures are cited in ROADMAP.md with their source quality"
license_review: PENDING
human_reviewers: ["Y4NN77 (decisions on positioning, content, assistant, Connect directory, profile visibility, demo data policy)"]
tests_or_evidence: ["no gate applies to documentation; consistency checks by search across the changed files", "NOT RUN: make quality is unaffected by these files"]
open_risks:
  - "the 225 Open Source study is based on home page text only; the team must read its terms and complete the study"
  - "the West Africa country list follows the UN geoscheme from memory and is to be verified"
  - "PRD amendment 3.2 is proposed until the owner approves it on the PR"
```

## S0 authz gate (B1) record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5"
purpose: "Implement the B1 route gate: three route lists, test-name existence check, scan for routers built outside httpx.NewRouter"
repository_or_source_context: "TODO.md S0, ROADMAP.md B1, scripts/check-authz-routes.sh, api/openapi.yaml"
files_or_sections_affected: ["scripts/check-authz-routes.sh", "scripts/authz/routes_check.go.txt", "scripts/tests/authz-routes_test.sh", "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: PENDING
human_reviewers: ["NICE-DEV226 (S0 reviewer)"]
tests_or_evidence: ["scripts/tests/authz-routes_test.sh: RED 11 of 13 before the change; second RED 8 of 21 on skeptic bypasses; third RED 3 of 25 on TestAuthz convention, NotFound handler and interface registration; 25 of 25 after"]
open_risks: ["the outside-NewRouter scan is syntactic; middleware in httpx that short-circuits a path is covered by owner review of router.go only", "the silent-pass guards (empty test list, test did not run) have no RED test", "a mapped TestAuthz test is checked for name, operationId prefix and direct t.Skip only, not for what it asserts", "the 501 probe is to be replaced by a method-set check in api-mount (TODO S0)"]
```

## S0 authz gate hardening record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5"
purpose: "Close the open gaps of the B1 gate: Docker coupling, hollow test mappings, silent-pass guard"
repository_or_source_context: "PR #31 review findings, TODO.md S0"
files_or_sections_affected: ["scripts/check-authz-routes.sh", "scripts/authz/routes_check.go.txt", "scripts/tests/authz-routes_test.sh", "src/api/internal/platform/httpx/router_test.go", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: PENDING
human_reviewers: ["NICE-DEV226 (S0 reviewer)"]
tests_or_evidence: ["scripts/tests/authz-routes_test.sh: RED 6 of 40 (Docker coupling, content rules, silent-pass guard), then RED 12 of 53 after the second skeptic review (toolchain-ignored test files, same-declaration, parent path, false positives); GREEN 53 of 53 (about 22 s)"]
open_risks: ["a mapped test is checked syntactically (route named, denial status mentioned, no skip outside testing.Short), not for what it really asserts: unused literals, dead code and an aliased t.Skip still pass and are review matters", "a planted init() or TestMain in package httpx could print a forged PASS line (repo-write attacker, review matter)", "GOPROXY and GOTOOLCHAIN from the caller are not scrubbed", "the 501 probe is replaced in api-mount"]
```

## S0 api-mount record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (orchestration) and go-implementer subagents (Sonnet)"
purpose: "S0 api-mount: pgxpool and /readyz, apigen mount, problem+json error table, cursor codec, spec-driven request validation, transaction runner, subject port, gate adaptation"
repository_or_source_context: "TODO.md S0, ADR-0003, ADR-0004, api/openapi.yaml, scripts/check-authz-routes.sh"
files_or_sections_affected: ["src/api/internal/platform/{apperr,cursor,apispec,db,subject,httpx,config}/**", "src/api/internal/app/**", "src/api/internal/platform/pgtest/tx_integration_test.go", "scripts/{check-authz-routes.sh,check-generated.sh,public-routes.txt,router-middleware.txt,authz/routes_check.go.txt,tests/authz-routes_test.sh}", "compose.yaml", ".env.example", ".claude/settings.json", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["github.com/getkin/kin-openapi (v0.142.0 already in go.sum, promoted from indirect to direct, then bumped to v0.149.0 because govulncheck reported 8 advisories fixed up to v0.149.0)"]
license_review: PENDING
human_reviewers: ["NICE-DEV226 (S0 reviewer)"]
tests_or_evidence: ["per-slice RED then GREEN reported by the implementers; see the PR description"]
open_risks: ["doc.OpenAPI is set to 3.0.3 at runtime so kin-openapi uses its keyword-reporting schema engine for a 3.1 contract", "the Content-Type violation code (C5) and METHOD_NOT_ALLOWED are outside the contract enum", "scripts/verify-containers.sh: the B0 rule 'api must not receive DATABASE_URL' is inverted (api must receive DATABASE_URL and CURSOR_KEY; no other service gets CURSOR_KEY)", "cursors carry no expiry (ADR-0004 point 5 says expired cursors are refused); not decided"]
```

## S0 idempotency and ETag record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (orchestration) and go-implementer subagents (Sonnet)"
purpose: "S0 Idempotency-Key middleware and store, ETag/If-Match helper, review fixes, B2 and B3 proposals"
repository_or_source_context: "TODO.md S0, ADR-0004 points 6 and 7, db/migrations/00004, api/openapi.yaml"
files_or_sections_affected: ["src/api/internal/platform/{idempotency,etag}/**", "src/api/internal/platform/httpx/{router.go,router_test.go,validate_test.go}", "src/api/internal/app/**", "src/api/internal/platform/pgtest/idempotency_integration_test.go", "scripts/{router-middleware.txt,secrets-scan.sh,gitleaks-ignore.txt,verify-containers.sh,tests/authz-routes_test.sh}", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["RFC 9110 sections 5.6.1.2 and 8.8.3 (read from memory of the standard; not fetched)"]
license_review: PENDING
human_reviewers: ["NICE-DEV226 (S0 reviewer)"]
tests_or_evidence: ["skeptic review reproduced a cross-resource replay (hash covered the route template only), fixed test-first; unit and PostgreSQL integration tests per finding"]
open_risks: ["the idempotency row is outside the business transaction (crash window of 2 minutes)", "bodies over the 16 KiB column check replay without a body", "scripts/gitleaks-ignore.txt holds one fingerprint for a dummy value already in pushed history"]
```

## S0 mailer record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (orchestration) and go-implementer subagents (Sonnet)"
purpose: "S0 mailer port, SMTP adapter, fake, Mailpit service, mailcheck command, DoD run"
repository_or_source_context: "TODO.md S0, ADR-0003, ADR-0006, compose.yaml, scripts/verify-containers.sh"
files_or_sections_affected: ["src/api/internal/platform/{mailer,config}/**", "src/api/internal/app/**", "src/api/cmd/mailcheck/**", "compose.yaml", ".env.example", "scripts/verify-containers.sh", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["axllent/mailpit v1.33.1 image (MIT), pinned by digest sha256:aef2aacc...61ce after `docker pull`"]
license_review: PENDING
human_reviewers: ["NICE-DEV226 (S0 reviewer)"]
tests_or_evidence: ["make up-local-dev on an isolated Compose project: /readyz 200 inside the api container, mailcheck delivered a test mail to Mailpit (message count 1)"]
open_risks: ["Gmail STARTTLS with the app password was not exercised (NOT RUN, H8)", "only AUTH PLAIN is supported", "a throwaway volume os226-s0-dod_pgdata remains on the developer machine"]
```

## S0 follow-ups record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (orchestration) and a go-implementer subagent (Sonnet)"
purpose: "Close the open points of the reviews of PRs #33 to #35: cursor expiry, problem+json request timeout, mixed-case media type, subject fallback, idempotency query-parameter guard"
repository_or_source_context: "review threads of PRs #33 and #34, ADR-0004 points 5 and 8, TODO.md S0"
files_or_sections_affected: ["src/api/internal/platform/{cursor,httpx,subject,idempotency}/**", "scripts/router-middleware.txt", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["each item test-first with a RED line; 100.0% statement coverage on cursor, httpx, subject and idempotency; route gate suite 59 passed"]
open_risks: ["a handler that ignores its context deadline still blocks the 503 until it returns (the middleware is synchronous like chi's)", "cursors issued before this change (wire format v1) are refused; none existed in production"]
```

## S9 foundation record (llm port, AI Studio adapter, aicheck)

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (orchestration) and a go-implementer subagent (Sonnet) in an isolated worktree"
purpose: "Provider-agnostic LLM port, a Google AI Studio adapter, AI_* configuration and the aicheck command that lets the owner answer V1, V2 and V4"
repository_or_source_context: "TODO.md S9, the mailer slice as the style reference"
files_or_sections_affected: ["src/api/internal/platform/llm/**", "src/api/internal/platform/config/{ai.go,config.go}", "src/api/internal/app/aicheck.go", "src/api/cmd/aicheck/**", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["Google AI Studio REST API (field names from the task brief, NOT confirmed against Google's documentation)"]
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["each deliverable test-first; 100.0% statement coverage on llm, llm/aistest, config, app, cmd/aicheck; no external network in tests"]
open_risks: ["the Gemini REST field names are unconfirmed until the first real aicheck run", "the adapter has never called the real API (V1 to V4 are the owner's)", "a ListModels page near 1 MiB would be rejected"]
```

## Secrets, backup and CI follow-ups record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5"
purpose: "SOPS and age secrets workflow (ADR-0009) with generated values, backup and restore drill, CI time (script tests no longer run twice), worktree helper, staging secrets file, AI Studio V1 to V4 measurements"
repository_or_source_context: "ADR-0009, docs/SHIP.md, mk/20-gates.mk, TODO.md S8 and S9, PRs #44 to #53"
files_or_sections_affected: ["scripts/{secrets,backup,restore-check,worktree}.sh", "scripts/tests/**", "mk/20-gates.mk", "secrets/*.recipients", "secrets/staging-test.enc.env", "docs/{SHIP,GIT_WORKFLOW}.md", "docs/decisions/0009-secrets-with-sops-and-age.md", "TODO.md", "ROADMAP.md", "AI_USAGE.md"]
third_party_sources_consulted: ["sops and age documentation (sops set --value-stdin)", "third-party model catalogs and Google AI forum posts for Gemma 4 (NOT official, used only to know what to check; every id was confirmed with aicheck on the real key)"]
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["secrets_test 47 passed, backup_test 24 passed, worktree_test 8 passed, make_test 42 passed", "make security PASS", "backup and restore drill on a throwaway local-dev stack: 39 tables and migration version 10 identical after restore", "aicheck models, ping fr, ping en, limits run by the assistant through sops exec-env with the owner's key at the owner's request (the key was never printed)"]
open_risks: ["the encrypted files are readable only with the owner's private age key (collaborators: issue #46)", "staging and production backup drills are NOT RUN", "the K21 EEA rule for the free tier is not yet handled", "the answers of both Gemma 4 models contain reasoning text: the adapter must drop it (TODO S9)"]
```

## CI parallel gate groups record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5, with an independent skeptic subagent review"
purpose: "Run the quality gates as parallel CI groups with a fan-in check, caches and a docs-only scope (ADR-0010)"
repository_or_source_context: "ADR-0006, mk/20-gates.mk, mk/40-quality.mk, .github/workflows/quality.yml, measured CI timings of 2026-11-12"
files_or_sections_affected: [".github/workflows/quality.yml", "mk/20-gates.mk", "mk/40-quality.mk", "Makefile", "scripts/{check-quality-evidence,ci-scope}.sh", "scripts/tests/{make,quality-evidence,ci-scope}_test.sh", "docs/decisions/0010-ci-parallel-gate-groups.md", "README.md", "AI_USAGE.md"]
third_party_sources_consulted: ["GitHub Actions documentation and community guidance on required checks, path filters and aggregator jobs (web search); pinned action versions verified with the GitHub API"]
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["make_test 49 passed, quality-evidence_test 13 passed, ci-scope_test 31 passed, lint-shell OK, actionlint clean", "all 11 groups run locally through make ci-<group> and the fan-in check PASS for 19 steps", "a skeptic subagent review found a confirmed rename hole in the docs scope (fixed with --no-renames -z) and smaller items (fixed)"]
open_risks: ["real GitHub Actions behaviour (matrix, artifacts, caches) is only proved by the CI run of the pull request", "a green quality on a documentation-only pull request means docs only (accepted in ADR-0010)", "the checker and the workflow come from the pull request itself: CODEOWNERS review is the control"]
```

## ADR 0007 to 0010 and decision documents record

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code (orchestrator session) with Claude subagents (lead, skeptic review for ADR-0010); model versions as stated in the neighbouring records"
purpose: "Draft and update ADR-0007 (Astryx pinning, dated 2026-11-11), ADR-0008 (team work areas, CODEOWNERS), ADR-0009 (secrets with SOPS and age) and ADR-0010 (parallel CI gate groups) and the related documents"
repository_or_source_context: "docs/decisions/, docs/team/, CODEOWNERS, ROADMAP.md, TODO.md; commits 738a65c, e0fbfd3, fa17ca5, 20dd18a, b625ffa"
files_or_sections_affected: ["docs/decisions/0007-*.md", "docs/decisions/0008-*.md", "docs/decisions/0009-*.md", "docs/decisions/0010-*.md", "docs/decisions/README.md", "CODEOWNERS", "docs/team/**"]
third_party_sources_consulted: ["sops and age documentation (ADR-0009)", "GitHub Actions documentation on required checks and path filters (ADR-0010)"]
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["documentation: no gate applies to the ADR text; the implementation evidence is in the slice records above (secrets, CI parallel gate groups)"]
open_risks: ["ADR approvals are owner decisions recorded in the files; this record does not prove them"]
```

## Prototype and UX record (Direction v2, design review)

```yaml
date: "2026-11-11"
contributor: "Y4NN77"
tool_or_model: "Claude Code (orchestrator session) with Claude subagents; exact model versions not recorded"
purpose: "HTML prototype v2 (Direction v2 flows, onboarding, premium visual system, French navigation, Apprendre pillar, global search), design review and QA fixes"
repository_or_source_context: "this repository, commits 3af147d, f2f8077, 7688c8a, 73a36fd, 536975d, d9c8783 (dates from git log)"
files_or_sections_affected: ["prototypes/html/**", "DESIGN.md", "docs/PRD.md (amendment 3.1)"]
third_party_sources_consulted: []
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["prototype QA findings closed in d9c8783; no automated gate applies to the static prototype"]
open_risks: ["the prototype is frozen as a reference; DESIGN.md governs the product UI", "prototype icons are an own CSS sprite; no third-party asset recorded"]
```

## Astryx verification and web skeleton record

```yaml
date: "2026-11-12"
contributor: "Y4NN77"
tool_or_model: "Claude Code (orchestrator session) with a web-implementer subagent (Sonnet); exact model versions not recorded"
purpose: "Install and verify Meta Astryx against the registry and CLI (ADR-0007), record the result, and add the React 19 web skeleton with TanStack and a Vitest coverage gate"
repository_or_source_context: "ADR-0007, DESIGN.md, commit cb7e7f9; owner-approved package installs"
files_or_sections_affected: ["src/web/**", "docs/quality/ASTRYX_REPORT.md", "pnpm-lock.yaml", "package.json"]
third_party_sources_consulted: ["npm registry metadata and `pnpm licenses` output for the web production dependencies (summary in docs/quality/ASTRYX_REPORT.md and THIRD_PARTY_NOTICES.md)"]
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["see docs/quality/ASTRYX_REPORT.md; visual, axe and responsive evidence for ADR-0007 is NOT RUN"]
open_risks: ["isbot (Unlicense) accepted by the owner on 2026-11-12; the organiser has not ruled", "DESIGN.md still contains pre-install Astryx statements that the report contradicts"]
```

## Secrets documentation and AI environment record (PR #56) and promotions (PRs #57 to #59)

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5"
purpose: "PR #56: README matches the encrypted-secrets flow, AI variables reach the api container, secrets README note. PR #57 is the CI parallel gate groups work already recorded above. PRs #58 and #59 are promotion and back-merge merge commits with no new authored content."
repository_or_source_context: "commits 7597803, f512164, 0b57cb7; PRs #56 to #59 (merge commits dated 2026-11-12)"
files_or_sections_affected: [".claude/settings.json", ".env.example", "README.md", "compose.yaml", "scripts/verify-containers.sh", "secrets/README.md"]
third_party_sources_consulted: []
license_review: PENDING
human_reviewers: []
tests_or_evidence: ["see the PR descriptions; no separate evidence is claimed here"]
open_risks: ["file list taken from `git show --stat` of commit 7597803 only; the merges of PR #56 were not re-diffed"]
```

## Documentation audit and cleanup record (2026-11-12)

```yaml
date: "2026-11-12"
contributor: "Y4NN777"
tool_or_model: "Claude Code (VS Code extension) with claude-sonnet-5-5 as lead model; 20 read-only subagents (scout, Explore), same model; execution agents on the same branch"
note: "The task specification named claude-opus-5-5 for the lead; the session ran on Sonnet 5.5."
purpose: "Documentation audit, cleanup plan and its execution: coherence fixes, archive of stale documents, CADEV records completed (additive only)"
repository_or_source_context: "tracked repository files, git metadata, read-only gh api GET calls; plan and findings in reports/docs-cleanup-plan.md and reports/docs-audit-findings.md (git-ignored)"
files_or_sections_affected: ["THIRD_PARTY_NOTICES.md", "AI_USAGE.md", "docs/SHIP.md", "docs/research/REFERENCES.md", "other documentation files per the pull request"]
third_party_sources_consulted: []
commands: "git diff/log, gh api GET, make help, make -s ci-groups, make -n quality; no build, test or install"
license_review: "NOT APPLICABLE (no third-party code or generated asset added)"
human_reviewers: []
tests_or_evidence: ["documentation only; consistency checks by search; no quality gate run for this record"]
open_risks: ["status: executed on branch chore/docs/cleanup-and-coherence, not yet merged", "owner review of the plan outcomes pending"]
```

## Catalogue reads (S2, not yet a pull request)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Cursor, Grok 4.7"
purpose: "Published catalogue reads: listResources, getResource, listResourceSources. Search and error reports are not in this change."
repository_or_source_context: "branch feature/discover/catalogue-read from origin/develop"
files_or_sections_affected: ["db/queries/discover.sql", "src/api/internal/platform/dbgen/discover.sql.go", "src/api/internal/discover/**", "src/api/internal/app/server.go", "src/api/internal/app/app.go", "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "src/api/internal/platform/pgtest/catalogue_read_integration_test.go", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: PENDING
closure: open
human_reviewers: []
tests_or_evidence: ["go test -race -count=1 ./src/api/... exit 0", "go test -race -count=1 -tags integration -run TestCatalogueReadsPublishedRowsOnly ./src/api/internal/platform/pgtest/ exit 0", "scripts/check-authz-routes.sh exit 0", "go tool sqlc vet -f db/sqlc.yaml exit 0", "make test-integration, make lint and make coverage were NOT RUN"]
open_risks: ["no freshness window is defined, so a source is never labelled stale", "content_relations cannot express a series link", "a resource with no published provider is absent from the catalogue", "pull request not opened"]
```
