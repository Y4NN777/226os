# 226os AI Use and Provenance Record

**Status:** records dated up to the owner's closure on 2026-11-13 are closed (see "Review and closure status"); each later record carries its own `closure` field and stays `open` until the owner closes it. Records are as written; licences were not independently verified.

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
license_review: ""   # NOT APPLICABLE, a verified licence, or ACCEPTED_BY_OWNER (date)
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
closure: ""          # closed by the owner (date)
tests_or_evidence: []
open_risks: []
```

Rule on `contributor` (added 2026-11-12): going forward the field holds the GitHub login. `Y4NN77` (the Git author name) and `Y4NN777` (the GitHub login) denote the same person, the repository owner; older records keep the value they were written with and are not rewritten.

## Review and closure status

**Closed on 2026-11-13 by the owner (`Y4NN777`).** Until then every record carried `license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"` and an empty or role-only `human_reviewers`. The owner closed them all by decision, with no further review: `license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"` records that decision, it is **not** the result of a licence review; `human_reviewers` names the owner as the person who closed the record, and no line-by-line human review is claimed; `closure` says when. The `open_risks` of each record stay as written because they are still true statements. Records added later fill the same three fields at the time they are written.

## Current task record

> Dated historical record (2026-11-11, documentation audit and G0 preparation); it no longer describes the repository, see the dated records below.

- Task: documentation audit and G0 Spec Kit preparation.
- AI assistance: repository reading, bounded specialist review, and Markdown drafting.
- Application implementation: none.
- Commands/builds/tests/deployments: none in the audit phase; Git publication was separately authorized and recorded outside application evidence.
- Third-party code integrated: none observed in this documentation set.
- Compliance status: closed by the owner on 2026-11-13 (no further review; not independently verified).

## H0 Claude runtime harness record

```yaml
date: "2026-11-11"
contributor: "Y4NN77"
tool_or_model: "Claude Code 2.1.322, Claude Opus 5.5 (validation probes on Haiku)"
purpose: "Add and validate the .claude/ runtime harness; reconcile main into develop history"
repository_or_source_context: "PROJECT_HARNESS_SPEC.md, AGENTS.md, CLAUDE.md, docs/GIT_WORKFLOW.md"
files_or_sections_affected: [".claude/settings.json", ".claude/agents/*.md", ".claude/skills/*/SKILL.md", "PROJECT_HARNESS_SPEC.md", "docs/spec-kit/CLAUDE_HANDOFF.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
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
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["documentation only; consistency checks by search; no quality gate run for this record"]
open_risks: ["status: executed on branch chore/docs/cleanup-and-coherence, not yet merged", "owner review of the plan outcomes pending"]
```

## First contract: topic code, owner-only decideApplication, profile_visibility

```yaml
date: "2026-11-12"
contributor: "mibienpanjoe"
tool_or_model: "Cursor, Grok 4.7"
purpose: "Apply the decided first-contract edits in api/openapi.yaml (C2 topic code, C1 owner-only decideApplication, C14b profile_visibility) and regenerate the Go and TypeScript types."
repository_or_source_context: "TODO.md FIRST CONTRACT PR; ROADMAP.md C1, C2, C14b; branch feature/platform/first-contract"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apigen/first_contract_test.go", "src/api/internal/platform/apigen/apigen.gen.go", "src/api/internal/platform/apispec/openapi.gen.yaml", "src/web/src/api/schema.d.ts"]
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
closure: "closed by the owner, 2026-11-13"
tests_or_evidence: ["go test -count=1 -run TestFirstContractDecisions ./src/api/internal/platform/apigen/: RED exit 1, then GREEN exit 0", "go test -count=1 ./src/api/internal/platform/apigen/: exit 0", "scripts/check-generated.sh generate: exit 0", "pnpm exec redocly lint api/openapi.yaml: exit 0"]
open_risks: ["handlers, the profile_visibility migration, and the response topic fields are out of scope"]
```

## Tracker line for the first contract pull request

```yaml
date: "2026-11-12"
contributor: "mibienpanjoe"
tool_or_model: "Cursor, Grok 4.7"
purpose: "Mark the first contract pull request in TODO.md. The contract diff itself is the earlier commit on this branch."
repository_or_source_context: "TODO.md header: update the tracker in the same PR. PR #79."
files_or_sections_affected: ["TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
closure: "closed by the owner, 2026-11-13"
tests_or_evidence: ["no product code changed; the box names PR #79 and the S5 docs follow-ups stay open"]
open_risks: ["the pull request is in review, not merged"]
```

## E2 taxonomy and profile visibility record

```yaml
date: "2026-11-12"
contributor: "mibienpanjoe"
tool_or_model: "Cursor, Grok 4.7"
purpose: "E2 forward-only migrations: five learn topics, three resource categories, four relabels, and the profiles.profile_visibility column, with pgtest fresh and upgrade coverage"
repository_or_source_context: "TODO.md E2, docs/spec-kit/DATABASE_DESIGN.md section 5, db/migrations/00002_reference_taxonomy.sql, db/migrations/00003_identity.sql"
files_or_sections_affected:
  - "db/migrations/00011_taxonomy_ai_topics.sql"
  - "db/migrations/00012_profile_visibility.sql"
  - "src/api/internal/platform/pgtest/migrate_integration_test.go"
  - "src/api/internal/platform/pgtest/constraints_integration_test.go"
  - "src/api/internal/platform/pgtest/fixtures_integration_test.go"
  - "scripts/tests/make_test.sh"
  - "scripts/check-todo-updated.sh"
  - "scripts/tests/check-todo-updated_test.sh"
  - "mk/20-gates.mk"
  - ".majesty/checks.json"
  - "docs/spec-kit/DATABASE_DESIGN.md"
  - "TODO.md"
  - "AI_USAGE.md"
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
closure: "closed by the owner, 2026-11-13"
tests_or_evidence:
  - "MIGRATIONS_BASE_REF=origin/develop scripts/check-migrations.sh: OK (12 files)"
  - "make test-integration: exit 0 (go test -race -tags integration ./... including pgtest, db-reset 42 passed, verify-containers ALL PASS)"
  - "make lint: exit 0 on the taxonomy files before the tracker check was added"
  - "scripts/tests/check-todo-updated_test.sh: 6 passed, 0 failed (a slice change without TODO.md fails; TODO.md moving with the slice passes)"
  - "scripts/check-todo-updated.sh on this tree: OK (TODO.md moves with 9 slice files against origin/develop)"
  - "scripts/lint-shell.sh: OK (47 scripts). Full make lint was not re-run after the tracker check was wired"
open_risks:
  - "DATABASE_DESIGN.md list counts for the immutable 00002 file are unchanged; current totals are asserted in pgtest"
  - "the local skill-guard hook edits are not part of this slice and must stay out of the commit"
  - "not committed and not reviewed"
```

## Regenerated profile model for PR #80

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Cursor, Grok 4.7"
purpose: "Commit the sqlc model field that migration 00012 adds, so the static gate's generate check matches the index"
repository_or_source_context: "PR #80, CI run 43334086904, gate static"
files_or_sections_affected: ["src/api/internal/platform/dbgen/models.go", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
closure: "closed by the owner, 2026-11-13"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["go tool sqlc generate -f db/sqlc.yaml then git diff: one added field, Profile.ProfileVisibility. make lint and the CI re-run are NOT RUN locally"]
open_risks: ["the generated field is appended where sqlc places an ALTER-added column, at the end of the Profile struct"]

```

## S9 assistant contract (2026-11-13)

```yaml
tool_or_model: "Claude Code (VS Code extension) with claude-sonnet-5-5"
purpose: "S9 contract PR: one OpenAPI operation, askAssistant (POST /assistant/answers), with tests first; generated outputs refreshed; dependent documents aligned; Gmail alias note (H8) recorded"
repository_or_source_context: "api/openapi.yaml, docs/SPEC.md US-INTEL-001, ADR-0011, TODO.md S9 block"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apigen/apigen.gen.go", "src/api/internal/platform/apispec/openapi.gen.yaml", "src/web/src/api/schema.d.ts", "src/api/internal/platform/apigen/assistant_contract_test.go", "src/api/internal/platform/apigen/contract_test.go", "scripts/unimplemented-routes.txt", "docs/quality/COVERAGE_SCOPE.md", "docs/SPEC.md", "docs/spec-kit/AUTHORIZATION_MATRIX.md", "TODO.md", "ROADMAP.md"]
third_party_sources_consulted: []
commands: "go test (apigen), make lint-contract, make security-authz, scripts/check-generated.sh generate"
license_review: "NOT APPLICABLE (no third-party code or asset added; generated code comes from the pinned generators)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["assistant_contract_test.go and contract_test.go (48 operations) pass; lint-contract and security-authz pass; make quality as a whole NOT RUN locally"]
open_risks: ["design choices to confirm in review: members only with a verified email, no anonymous access (O5 open), no Idempotency-Key, 503 reuses SERVICE_UNAVAILABLE, the 'AI-6/AI-7' P1 modes are out of scope"]

```

## AI code review tooling (2026-11-13)

```yaml
tool_or_model: "Claude Code (VS Code extension) with claude-sonnet-5-5 (setup); open-code-review 1.13.8 (the reviewer tool, LLM provider chosen by each user)"
purpose: "Set up advisory AI code review: project rules, a pinned local wrapper (make review), a Claude Code command (/delegate-review) adapted from open-code-review, and a Claude Code GitHub Actions workflow (OAuth token) that is off by default"
repository_or_source_context: "alibaba/open-code-review README, action.yml and configuration docs (read as untrusted external text); no code copied"
files_or_sections_affected: [".opencodereview/rule.json", "scripts/review.sh", "scripts/tests/review_test.sh", "mk/60-review.mk", ".claude/commands/delegate-review.md", ".github/workflows/code-review.yml", "docs/DEV.md", "THIRD_PARTY_NOTICES.md", "TODO.md"]
third_party_sources_consulted: ["https://github.com/alibaba/open-code-review (Apache-2.0; delegate-review.md adapted, header records the commit and changes)", "https://github.com/anthropics/claude-code-action (README and docs read as untrusted external text)"]
commands: "scripts/tests/review_test.sh, make help, scripts/lint-shell.sh, make_test.sh; the reviewer tool itself was NOT RUN"
license_review: "Apache-2.0 per package metadata and repository page; not re-verified from source"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["review_test.sh 9 passed (fake npx); lint-shell OK; make_test 49 passed"]
open_risks: ["the reviewers send diffs to an external provider (Anthropic for /delegate-review and the CI workflow); reviews are AI output to be checked by a human; the CI workflow and /delegate-review are NOT RUN until the owner enables them; Anthropic terms for a subscription token in shared CI are unverified"]
```

## S9 adapter output cleaning and free-tier tracking (2026-11-13)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code (VS Code extension) with claude-sonnet-5-5"
purpose: "S9 adapter groundwork in platform/llm, test first: remove reasoning text from model answers (CleanAnswer, thought parts) and track 429/500/503/timeouts and latency per model (Tracker, StatusError)"
repository_or_source_context: "TODO.md S9 block, ROADMAP V1 to V4 results, existing platform/llm and aistest packages"
files_or_sections_affected: ["src/api/internal/platform/llm/answer.go", "src/api/internal/platform/llm/answer_test.go", "src/api/internal/platform/llm/tracker.go", "src/api/internal/platform/llm/tracker_test.go", "src/api/internal/platform/llm/llm.go", "src/api/internal/platform/llm/aistudio.go", "TODO.md"]
third_party_sources_consulted: []
commands: "go test -race -cover (llm: 100% statements), go vet, golangci-lint, staticcheck, go test -race ./... (all packages pass)"
license_review: "NOT APPLICABLE (no third-party code or asset added)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["answer_test.go and tracker_test.go failed first (undefined symbols), then passed; llm coverage 100%"]
open_risks: ["test fixtures are synthetic: the real Gemma 4 reasoning shape was NOT observed here (needs the owner's aicheck run); nothing is wired into the API yet"]
```

## S8 smoke script and release workflow draft (2026-11-13)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code (VS Code extension) with claude-sonnet-5-5"
purpose: "S8: post-deploy smoke script (scripts/smoke.sh, make smoke) written test first, and a non-active draft of the release workflow (docs/RELEASE_DRAFT.md)"
repository_or_source_context: "docs/SHIP.md sections 2 to 5, ROADMAP B2 and B4, docs/decisions/0006, scripts/compose.sh, scripts/build-images.sh"
files_or_sections_affected: ["scripts/smoke.sh", "scripts/tests/smoke_test.sh", "mk/50-ops.mk", "docs/RELEASE_DRAFT.md", "docs/SHIP.md", "docs/DEV.md", "TODO.md"]
third_party_sources_consulted: []
commands: "scripts/tests/smoke_test.sh (12 passed, fake curl and fake compose), scripts/lint-shell.sh, make_test.sh; the smoke script itself was NOT RUN against a real host"
license_review: "NOT APPLICABLE (no third-party code or asset added)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["smoke_test.sh failed first (script missing), then passed; ShellCheck clean"]
open_risks: ["the script has only run against fakes: the real curl output formats and a real stack are untested", "the release draft is a proposal with open owner questions; no workflow file was added"]
```

## Review gate (2026-11-13)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code (VS Code extension) with claude-sonnet-5-5"
purpose: "A required check that lets the AI review stop an OBSERVED CRITICAL (owner override, fail open) and asks the owner's approval only on sensitive paths; report mode first. Written test first."
repository_or_source_context: "the first review output on PR #85, ADR-0008, GIT_WORKFLOW, branch protection settings read with gh api"
files_or_sections_affected: ["scripts/review-gate.py", "scripts/tests/review-gate_test.sh", ".github/workflows/code-review.yml", ".github/sensitive-paths.txt", ".claude/commands/delegate-review.md", "docs/DEV.md", "docs/GIT_WORKFLOW.md", "TODO.md"]
third_party_sources_consulted: []
commands: "scripts/tests/review-gate_test.sh (26 passed, fake gh), scripts/lint-shell.sh; the gate was NOT RUN on GitHub yet"
license_review: "NOT APPLICABLE (no third-party code or asset added)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["review-gate_test.sh failed first (script missing), then passed"]
open_risks: ["the workflow YAML of a pull request runs from the PR itself, so only a human reading changes to .github can stop a PR from weakening the job (the script and the list come from the base)", "a false CRITICAL blocks one PR until the owner labels it", "report mode first: not enforced yet"]
```

## Catalogue reads (S2)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Cursor, Grok 4.7"
purpose: "Published catalogue reads: listResources, getResource, listResourceSources. Search and error reports are not in this change."
repository_or_source_context: "branch feature/discover/catalogue-read, PR #83"
files_or_sections_affected: ["db/queries/discover.sql", "src/api/internal/platform/dbgen/discover.sql.go", "src/api/internal/discover/**", "src/api/internal/app/server.go", "src/api/internal/app/app.go", "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "src/api/internal/platform/pgtest/catalogue_read_integration_test.go", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
closure: "closed by the owner, 2026-11-13"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["go test -race -count=1 ./src/api/... exit 0", "go test -race -count=1 -tags integration -run TestCatalogueReadsPublishedRowsOnly ./src/api/internal/platform/pgtest/ exit 0", "scripts/check-authz-routes.sh exit 0", "go tool sqlc vet -f db/sqlc.yaml exit 0", "make test-integration, make lint and make coverage were NOT RUN"]
open_risks: ["no freshness window is defined, so a source is never labelled stale", "content_relations cannot express a series link", "a resource with no published provider is absent from the catalogue"]
```

## Tracker line for the catalogue pull request

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Cursor, Grok 4.7"
purpose: "Record PR #83 on the three catalogue operations. Search and error reports stay open."
repository_or_source_context: "PR #83"
files_or_sections_affected: ["TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
closure: "closed by the owner, 2026-11-13"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["no product code in this commit"]
open_risks: []
```

## Catalogue lint-go fix

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Cursor, Grok 4.7"
purpose: "Clear the lint-go failure on PR #83: wrap catalogue query errors with %w, and give catalogue test requests a context."
repository_or_source_context: "PR #83, Quality run 43355277357, gate lint-go"
files_or_sections_affected: ["src/api/internal/discover/repository/repository.go", "src/api/internal/discover/transport/handler_test.go", "src/api/internal/app/catalogue_internal_test.go", "src/api/internal/platform/pgtest/catalogue_read_integration_test.go", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-13): not independently verified"
closure: "closed by the owner, 2026-11-13"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-13)"]
tests_or_evidence: ["make lint-go exit 0", "go test -race -count=1 ./src/api/internal/discover/... ./src/api/internal/app/ exit 0. The catalogue integration test was not re-run."]
open_risks: ["Intended, not a defect: Get translates pgx.ErrNoRows to domain.ErrNotFound and returns it without wrapping, so no driver type leaks past the repository (go-backend rule). Every other driver error stays in the chain via %w."]
```

## Search core and catalogue follow-ups (S2)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session) with Claude Sonnet subagents (go-implementer, skeptic review)"
purpose: "GET /search over questions, resources, projects and learn (published and visible only, C4 PROPOSED option 1), plus the PR #83 review follow-ups: length-prefixed cursor filter hash, interleaved related links, EXPLAIN plans for the catalogue q and sdk filters, and a 422 for NUL or invalid UTF-8 in the catalogue q."
repository_or_source_context: "branch feature/discover/search"
files_or_sections_affected: ["db/queries/search.sql", "src/api/internal/platform/dbgen/search.sql.go", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/search/**", "src/api/internal/discover/service/**", "src/api/internal/discover/repository/**", "src/api/internal/app/app.go", "src/api/internal/app/server.go", "src/api/internal/app/search_internal_test.go", "src/api/internal/platform/pgtest/search_read_integration_test.go", "src/api/internal/platform/pgtest/explain_helpers_integration_test.go", "src/api/internal/platform/pgtest/catalogue_read_integration_test.go", "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "PENDING: owner review"
closure: "open"
human_reviewers: ["NICE-DEV226 (pending)"]
tests_or_evidence: ["make generate, lint-go, lint-contract, typecheck, test-unit, coverage (Go 100.0%) and test-integration-go PASS on the pre-merge tree", "EXPLAIN (ANALYZE, BUFFERS) logged per group on 6001 rows: GIN index for a rare term, about 5 ms for a common term and a negation-only query"]
open_risks: ["no anonymous 60/IP/min limit on search yet (next S2 PR, on the S1 rate limiter)", "ts_rank scores every match before the top-N cut; bounded by LIMIT and the 2 s per-query timeout", "answers are not searched (no answers group in the contract)", "the cursor embeds the raw filter text, so a long accented q can exceed the 512-character cursor limit (pre-existing)"]
```

## S1 identity: domain, password hashing and auth service core

```yaml
date: "2026-11-13"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "S1 Identity slice, auth-sessions branch: Argon2id password hashing (PHC, OWASP floor, timing-equalized dummy), session domain model, sign-up/sign-in/session/CSRF/verify/reset use cases behind a Store port, sqlc queries and repository adapter (atomic account-creation CTE), HTTP transport with per-profile cookies, session-backed subject resolver, synchronizer-token CSRF middleware, PostgreSQL rate limits, app wiring, unit tests on fakes plus real-PostgreSQL pgtest integration."
repository_or_source_context: "branch feature/identity/auth-sessions, TODO.md S1"
files_or_sections_affected: ["src/api/internal/identity/**", "src/api/internal/app/**", "src/api/internal/platform/httpx/**", "src/api/internal/platform/dbgen/**", "src/api/internal/platform/pgtest/**", "db/queries/identity.sql", "scripts/public-routes.txt", "scripts/authz-tests.txt", "scripts/router-middleware.txt", "scripts/unimplemented-routes.txt", "scripts/tests/authz-routes_test.sh", "TODO.md", "go.mod", "go.sum", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "PENDING"
closure: "open"
human_reviewers: ["mibienpanjoe (assigned reviewer, pending PR)"]
tests_or_evidence: ["go test -race -count=1 ./src/api/internal/identity/... exit 0", "go test -race -count=1 -tags integration ./src/api/internal/platform/pgtest/ exit 0", "scripts/coverage-go.sh coverage: go first-party statement coverage 100.0% (2705/2705), PASS", "make coverage exit 0 (web 99.27% statements)", "make test-integration exit 0", "make lint-scripts security-sast exit 0 (gosec 0 issues)", "scripts/check-generated.sh check exit 0", "make quality full rerun on the PR head SHA pending at the time of this entry"]
open_risks: ["rate-limit bucket keys reuse CURSOR_KEY as HMAC key pending a dedicated secret decision", "CSRF tokens are derived (HMAC of stored token hash / pre-session secret) rather than stored, per ADR-0003 rotation requirement", "no resend-verification operation exists in the contract (open question for the owner, flagged in the PR)", "getMe, updateMyProfile, getProfile and the profile-visibility matrix remain for feature/identity/profiles", "Argon2id parameters await the target-host benchmark at S8 (K6 container check)"]
```

## S1 identity: review follow-ups (rate-limit purge, key separation)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Sonnet 5.5"
purpose: "Close the follow-ups left open by the review of PR #89: purge finished rate_limit_buckets windows when a new bucket opens (bounded DELETE with SKIP LOCKED; a failed purge is logged, never denies the request), separate HKDF-Expand subkeys for rate-limit bucket keys and CSRF tokens, one session lookup per request through a per-request memo in platform/subject, sign-in counting only failures per ADR-0003 (read before the lookup, written on failure, same for known and unknown emails), and the reset-request token write plus mail moved off the request path so a store error and the response time no longer tell a known account from an unknown one."
repository_or_source_context: "branch feature/identity/review-followups, issue #65"
files_or_sections_affected: ["src/api/internal/identity/service/service.go", "src/api/internal/identity/repository/repository.go", "db/queries/identity.sql", "src/api/internal/platform/dbgen/identity.sql.go", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "src/api/internal/platform/subject/subject.go", "src/api/internal/identity/transport/**", "src/api/internal/app/server_identity_test.go", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["RFC 5869 (HKDF-Expand, implemented with crypto/hmac)"]
license_review: "ACCEPTED_BY_OWNER (2026-11-13): original work, no third-party code"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: TestHitPurgesExpiredBucketsWhenAWindowOpens, TestSubkeysAreDerivedPerPurpose, TestSessionIsResolvedOncePerRequest (3 lookups, want 2), TestSignInSuccessDoesNotCountAgainstTheAccountLimit (6th success 429), TestMemo*, TestRequestPasswordResetRateLimitAndErrors (500 became a silent 202; expectation changed because a 500 on the known path is an enumeration oracle, ADR-0003)", "go test -race -tags integration -run Identity ./src/api/internal/platform/pgtest/ exit 0 (real PostgreSQL, incl. TestIdentitySignInCountsOnlyFailures)", "go test -race ./src/api/internal/identity/... ./src/api/internal/app/... exit 0", "go test -tags integration -run TestIdentityRateLimitPurgesOldWindows ./src/api/internal/platform/pgtest/ exit 0 (real PostgreSQL)", "make coverage: go first-party 100.0% (2705/2705)", "make lint exit 0, scripts/check-generated.sh check exit 0, check-docs PASS"]
open_risks: ["ui_lang has no column and csrf_token_hash is never read: both need a migration (mibienpanjoe), tracked on issue #65.", "Sign-in failure counting is check-then-increment: parallel wrong guesses can slip a few attempts past the account bucket; the per-IP bucket (atomic) and the Argon2id semaphore bound it.", "A reset token write that fails in the background sends no mail and the person asks again; it is not logged.", "Existing CSRF tokens and rate-limit buckets are invalidated once at deploy (derived keys differ); harmless before launch."]
```

## S1 identity profiles: getMe, updateMyProfile, getProfile

```yaml
date: "2026-11-14"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "S1 Identity profiles slice, feature/identity/profiles branch: profile read models (own + public with account status join), JSON Merge Patch update with tri-state PatchField (absent vs explicit null), per-field CASE WHEN guarded UPDATE plus transactional skill replacement (row lock then separate DELETE/INSERT on fresh snapshots, ON CONFLICT DO NOTHING), C14b visibility matrix (page + per-field, anonymous/member/owner), skill slug resolution against active tags with active-only reads, bio_language derivation to satisfy profiles_bio_language_ck, rate limit profile:write 60/user/min, transport merge-patch decoding with key-presence detection relying on the embedded OpenAPI validator for non-nullable nulls, unit tests on fakes, repository scripted-DBTX tests, real-PostgreSQL pgtest end-to-end including concurrent PATCH serialization and hidden-tag filtering."
repository_or_source_context: "branch feature/identity/profiles, TODO.md S1"
files_or_sections_affected: ["src/api/internal/identity/**", "src/api/internal/app/server.go", "src/api/internal/app/server_identity_test.go", "src/api/internal/platform/dbgen/identity.sql.go", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "db/queries/identity.sql", "scripts/public-routes.txt", "scripts/authz-tests.txt", "scripts/unimplemented-routes.txt", "scripts/tests/authz-routes_test.sh", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "PENDING"
closure: "open"
human_reviewers: ["mibienpanjoe (assigned reviewer, pending PR)"]
tests_or_evidence: ["go test -race -count=1 ./src/api/internal/identity/... ./src/api/internal/app/ exit 0", "go test -race -count=1 -tags integration ./src/api/internal/platform/pgtest/ exit 0 (TestIdentityProfilesEndToEnd, TestIdentityProfileHiddenTagFiltered, TestIdentityProfileConcurrentSkillReplace: 10 conflicting + 5 identical concurrent rounds on two real connections, final set always exact)", "scripts/coverage-go.sh coverage: go first-party statement coverage 100.0%", "go tool golangci-lint run (default and --build-tags integration): 0 issues", "make lint-contract exit 0 (sqlc vet LIMIT on ResolveSkillSlugs)", "make lint-scripts exit 0 (authz-routes self-test probes moved to ListQuestions)", "scripts/check-authz-routes.sh PASS", "gofmt clean", "go vet clean", "review round: go test -count=1 -run 'TestUpdateMyProfileLanguages' ./internal/identity/service/ exit 0 (fr-BF -> 422 languages), go test -count=1 -run 'TestUpdateMyProfile' ./internal/identity/transport/ exit 0 (null-non-nullable pinned)"]
open_risks: ["bio_language is always derived as 'en' when bio is set: the contract exposes no bio language field (owner decision recorded in PR)", "open_to_collab_visibility column exists but the contract field_visibility has no member for it: stays 'public', flagged for owner decision", "languages: server rejects region subtags (fr-BF) with 422 on purpose while the contract still accepts them; owner owns the contract+DB widening after the freeze (TODO.md known deviation)", "checklist_missing no longer lists open_to_collab: false is a deliberate answer and no answered flag exists", "make quality full rerun pending on the PR head SHA"]
```

## Identity test reset-mail race (#91)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Fix the flaky identity integration test: mails are sent off the request path, so mailToken now waits for the next mail after a given count instead of returning the latest one; the sibling-token test waits for both reset mails."
repository_or_source_context: "issue #91, branch fix/identity/reset-mail-race"
files_or_sections_affected: ["src/api/internal/platform/pgtest/identity_integration_test.go", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE (test code only)"
closure: "open"
human_reviewers: ["NICE-DEV226 (pending)"]
tests_or_evidence: ["before: go test -tags integration -run TestIdentitySignInRotationAndReset failed 3/3 on clean develop 512a290", "after: go test -count=1 -tags integration -run TestIdentity ./src/api/internal/platform/pgtest/ passed 8/8", "go test -race -count=1 -tags integration ./src/api/internal/platform/pgtest/ ok", "make lint-go PASS"]
open_risks: []
```

## Error reports, search rate limit and shared limiter (S2)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session) with Claude Sonnet subagents (go-implementer, skeptic review)"
purpose: "createResourceErrorReport (verified member, reporter private, duplicate open report returns the existing one, 60/user/min failing closed), anonymous search limit of 60/IP/min failing open, shared platform/ratelimit limiter and httpx.ClientIPKey, migration 00013 error_reports.source_id with a composite FK."
repository_or_source_context: "branch feature/discover/error-reports"
files_or_sections_affected: ["db/migrations/00013_error_report_source.sql", "db/queries/ratelimit.sql", "db/queries/discover.sql", "src/api/internal/platform/ratelimit/**", "src/api/internal/platform/httpx/clientip*.go", "src/api/internal/platform/dbgen/**", "src/api/internal/discover/**", "src/api/internal/search/**", "src/api/internal/app/**", "src/api/internal/platform/pgtest/**", "scripts/authz-tests.txt", "scripts/unimplemented-routes.txt", "scripts/public-routes.txt", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "PENDING: owner review"
closure: "open"
human_reviewers: ["NICE-DEV226 (pending)", "Y4NN777 (migration and shared limiter, pending)"]
tests_or_evidence: ["make generate, lint, typecheck, test-unit, test-integration-go, security PASS (subagent run)", "make coverage PASS: Go first-party 100.0% (3079/3079), re-run in the main session"]
open_risks: ["identity still has its own limiter until it migrates to platform/ratelimit", "rate-limit integration tests depend on the real clock (they wait for a fresh minute)", "whitespace-only report bodies are rejected with 422, stricter than the schema's minLength 1", "a repeated open report returns the existing report and drops the new body and source_id", "the client IP trusts the last X-Forwarded-For entry; correct only while Traefik is the sole ingress with default forwardedHeaders (not pinned in compose.yaml)", "/24 and /48 keying means users behind carrier NAT share 60 searches per minute"]
```

## AI review: minimal fix, project memory and pointers (2026-11-13)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Sonnet 5.5 (with an Opus plan review)"
purpose: "PR 1 of the review plan: the gate stops waiting for the AI review, one comment writer per pull request, a review only when needed (first, incremental on a still-open CRITICAL or HIGH, else skip; the label ai-review always reviews), scripts read from the BASE branch, a context pack with the project memory, the ledger and pointers found by script on the changed lines of .sql and Go files (INSERT without ON CONFLICT, DELETE and INSERT on one table in one statement, X-Forwarded-For reads, error text in log attributes), a logic lens in the review prompt, a fenced JSON ledger, and a deterministic owner ping (label needs-owner and a review request) instead of an @-mention. Reason: the review ran on every push (9 comments on #89 and #94), repeated itself, ignored answers and never read .sql files, so it missed both #94 blockers."
repository_or_source_context: "PR #89, #90, #94, #96 review comments; .github/workflows/code-review.yml; .claude/commands/delegate-review.md; the real #94 heads 469984c, 9d12650 and 5c350d5 as test fixtures"
files_or_sections_affected: ["docs/quality/REVIEW_CONTEXT.md", "scripts/review-context.py", "scripts/review-ping.py", "scripts/tests/review-context_test.sh", "scripts/tests/review-ping_test.sh", "scripts/tests/fixtures/review-context/**", ".opencodereview/context-map.json", ".opencodereview/rule.json", ".claude/commands/delegate-review.md", ".github/workflows/code-review.yml", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["Addy Osmani, Code Review in the Age of AI (principles only)", "O'Reilly Radar, Agentic code review (principles only)"]
license_review: "NOT APPLICABLE: original code and text, no new dependency; the sources above were read for ideas, nothing was copied"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["scripts/tests/review-context_test.sh and scripts/tests/review-ping_test.sh: RED first (no pointers, no decision line, no script), then GREEN; the fixtures are the real #94 heads, #96 and #110", "Acceptance run (real reviewer agents, new pack, scratch worktrees built from the real heads): #94 first head 469984c with the later decisions stripped from the memory: BOTH blockers flagged (skills double-submit 500 at identity.sql:154, languages pattern drift at service/profile.go:31); #94 fixed head 5c350d5: no SQL blocker; #110: Triage NONE, empty Read first. The run also found two defects of the script itself, fixed with tests (an empty pull request answer crashed the pack; a ledger written with another summary label or key was not read)", "scripts/tests/review-context_test.sh 37 passed, review-ping_test.sh 12 passed; make lint-scripts and python3 -I scripts/check-docs.py: see the pull request", "NOT RUN until the first live pull request after the merge: the BASE-read script path, first|incremental|skip on real events, gh pr comment --edit-last --create-if-none on the runner, the needs-owner label and the review request"]
open_risks: ["The pointers are regular expressions on changed lines: they will miss cases and add some noise; they point, they never judge.", "The review job now also runs on synchronize (it decides in seconds whether to skip), so a skipped run still costs a runner start.", "REVIEW_CONTEXT.md only helps if the owner adds a line when a decision is made."]
```
