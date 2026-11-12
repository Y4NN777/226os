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
