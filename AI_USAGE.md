# 226os AI Use and Provenance Record

**Status:** records dated up to the owner's closure on 2026-11-13 are closed (see "Review and closure status"); each later record carries its own `closure` field and stays `open` until the owner closes it. On 2026-11-15 the owner closed every record whose pull request is merged into `develop` (`human_reviewers` names the merge pull request; reviewers still marked pending were dropped, no line-by-line review is claimed); the owner accepted the licence of the four French lessons (`2026-11-13-s4-learn.md`), and the records missing for the owner's merged pull requests #88, #123 and #235 were written after the fact and say so. The records for #121 and #136 are their contributor's to write (issue #242). `2026-11-15-docs-refresh-staging-seeds.md` is closed with its follow-up #241. Records are as written; licences were not independently verified.

## Policy

AI output is an untrusted proposal. Human owners remain responsible for requirements, architecture, security, tests, licenses, and release claims. Rewriting, translating, or generating code with AI does not remove third-party copyright or license obligations.

Do not submit secrets, credentials, private user data, unpublished tokens, or confidential external material to AI systems. Do not use AI output as evidence that a command, test, source review, performance measurement, or security check occurred.

## Required record for each material AI-assisted contribution

**Where to write it (since 2026-11-13):** one file per record in `docs/ai-usage/` (see its README), never an edit of this file: every pull request used to append at the end here, so every merge conflicted with every other open pull request. The block at the end of this file is generated from those files by `make ai-usage` (the owner runs it before the submission). The records above the block are the historical ones and stay as written.

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

**Licence verification, 2026-11-15 (owner `Y4NN777`).** The historical records above stay as written; this paragraph gives each licence outcome instead of editing them.

- **Dependencies:** the third-party licences these records rely on were read from upstream in #88 (2026-11-13; method in `THIRD_PARTY_NOTICES.md`, "Verification"). Two direct Go modules were missing from that table and are now listed: `github.com/google/uuid` v1.6.0 and `golang.org/x/crypto` v0.58.0. #89 promoted both from indirect to direct, and both are BSD-3-Clause, read from their LICENSE files on 2026-11-15.
- **"S1 identity: domain, password hashing and auth service core"** (`PENDING`): VERIFIED. It is the change that made `google/uuid` and `golang.org/x/crypto` direct dependencies (BSD-3-Clause, above).
- **"Search core and catalogue follow-ups (S2)"** and **"Error reports, search rate limit and shared limiter (S2)"** (`PENDING: owner review`): NOT APPLICABLE. Their pull requests (#92 and #96) changed no dependency file and cite no third-party source.
- **"AI code review tooling (2026-11-13)"** ("not re-verified from source"): VERIFIED. `open-code-review` is Apache-2.0, read from its upstream LICENSE file on 2026-11-15. It runs through `npx` and is not distributed.
- **The records closed as `ACCEPTED_BY_OWNER (2026-11-13): not independently verified`:** their dependency licences are covered by #88 and the line above. What that wording still covers is the originality of the AI-written text and code, which no licence check can prove. It stays an owner acceptance.
- **"Brand assets (concept D)":** the copyright and trademark status of AI-generated marks stays an owner and legal decision (G0-COMP-001). It is not a dependency licence and cannot be verified from a licence file.

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
date: "2026-11-13"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "S1 Identity profiles slice, feature/identity/profiles branch: profile read models (own + public with account status join), JSON Merge Patch update with tri-state PatchField (absent vs explicit null), per-field CASE WHEN guarded UPDATE plus transactional skill replacement (row lock then separate DELETE/INSERT on fresh snapshots, ON CONFLICT DO NOTHING), C14b visibility matrix (page + per-field, anonymous/member/owner), skill slug resolution against active tags with active-only reads, bio_language derivation to satisfy profiles_bio_language_ck, rate limit profile:write 60/user/min, transport merge-patch decoding with key-presence detection relying on the embedded OpenAPI validator for non-nullable nulls, unit tests on fakes, repository scripted-DBTX tests, real-PostgreSQL pgtest end-to-end including concurrent PATCH serialization and hidden-tag filtering."
repository_or_source_context: "branch feature/identity/profiles, TODO.md S1"
files_or_sections_affected: ["src/api/internal/identity/**", "src/api/internal/app/server.go", "src/api/internal/app/server_identity_test.go", "src/api/internal/platform/dbgen/identity.sql.go", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "db/queries/identity.sql", "scripts/public-routes.txt", "scripts/authz-tests.txt", "scripts/unimplemented-routes.txt", "scripts/tests/authz-routes_test.sh", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency (no go.mod, go.sum, package.json or lockfile change in PR #94; third_party_sources_consulted is empty)"
closure: "closed by the owner, 2026-11-13"
human_reviewers: ["mibienpanjoe (reviewed, changes requested 2026-11-13; his points verified in the code by the owner, no second round)", "Y4NN777 (owner closure, 2026-11-13; merged without a second review round, leftovers fixed in the follow-up PR)"]
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

## AI review: gates in report mode (2026-11-13)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Sonnet 5.5"
purpose: "PR 3 of the review plan: three deterministic checks that replace what the AI review kept nagging about, all in report mode. Review gate: stray paths (.github/stray-paths.txt, read from the base branch: .devin, .cursor, .idea, .vscode, and .gitignore on feature and fix branches; label stray-ok exempts) and finding-to-test (a pull request that closes a finding issue must change a test file; label no-test exempts), with their own switch GATE_NEW (repository variable CODE_REVIEW_GATE_NEW: off, report, enforce). check-docs: an AI_USAGE.md record dated after the HEAD commit date (UTC, one day of tolerance) is a WARN; --strict-dates makes it a failure."
repository_or_source_context: "plan revision 4 (owner), PR #94 (a .devin file and a .gitignore edit in a profiles pull request; an AI_USAGE record dated the next day)"
files_or_sections_affected: ["scripts/review-gate.py", "scripts/tests/review-gate_test.sh", "scripts/check-docs.py", "scripts/tests/check-docs_test.sh", ".github/stray-paths.txt", ".github/workflows/code-review.yml", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["scripts/tests/review-gate_test.sh 39 passed and scripts/tests/check-docs_test.sh 16 passed, RED first (the new checks did not exist)", "make lint-scripts and python3 -I scripts/check-docs.py: see the pull request", "NOT RUN until the first live pull request: the two new gate checks against the real GitHub API (labels, pull request body, the stray-paths list on the base branch)"]
open_risks: ["Both gate checks are report only: they print a NOTE and never fail until the owner sets CODE_REVIEW_GATE_NEW=enforce (after 10 pull requests with no false positive).", "Finding-to-test reads the issue numbers after Closes/Fixes/Resolves in the pull request body: a different wording is not seen.", "In enforce mode a CRITICAL verdict is read at the next gate run (a push, a label or a review event), because the gate no longer waits for the review."]
```

## AI review: findings register, dashboard and memory loop (2026-11-13)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Sonnet 5.5"
purpose: "PR 2 of the review plan: when a pull request is merged, the findings the review left open become GitHub issues (CRITICAL and HIGH one each, at most 5; everything else one follow-ups checklist per pull request), once, by marker; one Findings dashboard ranks what to fix next (due, severity, area, age) from issue state only; the context pack lists the tracked and the not-planned findings so the reviewer does not raise them again; the INSERT pointer names the unique key found in db/migrations; the PR template gets a Review focus section."
repository_or_source_context: "plan revision 4 (owner), PR #112 (PR 1); the fake gh of scripts/tests/findings_test.sh is stateful so idempotency is really tested"
files_or_sections_affected: ["scripts/findings.py", "scripts/tests/findings_test.sh", "scripts/review-context.py", "scripts/tests/review-context_test.sh", "scripts/tests/fixtures/review-context/**", ".github/workflows/code-review.yml", ".github/PULL_REQUEST_TEMPLATE.md", "mk/60-review.mk", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["scripts/tests/findings_test.sh 23 passed and scripts/tests/review-context_test.sh 44 passed, RED first (no script, no sections, no key names)", "make lint-scripts and python3 -I scripts/check-docs.py: see the pull request", "NOT RUN until the first merge after this PR: the harvest job on a real merged pull request, board 4 Auto-add for issues created by GITHUB_TOKEN (the owner enables the Auto-add workflow with filter label:finding in the project settings), the dashboard issue creation (make findings or scripts/findings.py dashboard once, then pin it)"]
open_risks: ["Notes of the AI review on #114 and #115 were all addressed in this branch (the AI review said INFERRED, the owner wants them fixed): GATE_NEW=enforce now blocks on its own whatever GATE_MODE says; a negation (does not close #N) is not a closing keyword and Fixes: #N is one; the AI_USAGE date check uses the last non-merge commit (CI checks out a merge commit); an @mention in ledger text is neutralised; an issue whose marker was edited away is recognised by its title, and titles now carry the PR number so that check is per pull request; a finding whose file changed after the review is still filed but says so (GitHub compare API); a blocking GATE_NEW check still prints what the other checks found; the skipped gate job on closed is explained in the workflow. Tests added for each.", "GITHUB_TOKEN cannot write to a user project: until the Auto-add workflow of board 4 is enabled, the issues exist but are not on the board.", "Labels are created by the issue API on first use with default colours.", "The register trusts the model's ledger only through a strict schema; a malformed ledger is skipped and logged."]
```

## S7 API gaps: provenance, country lists, mail links

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (integrator) and Claude Sonnet 5.5 (go-implementer in a separate worktree)"
purpose: "Close three gaps found while building S7: Resource.provenance in the contract and responses, a comma-separated country list on listResources (union, max 20), and verify/reset mails carrying a #token= link built from an optional APP_BASE_URL."
repository_or_source_context: "branch fix/platform/s7-api-gaps, owner decision 2026-11-13 (one backend PR), based on PR #96"
files_or_sections_affected: ["api/openapi.yaml", "generated outputs (apigen, apispec, dbgen, src/web/src/api/schema.d.ts)", "db/queries/discover.sql", "src/api/internal/discover/**", "src/api/internal/identity/service/**", "src/api/internal/platform/config/**", "src/api/internal/app/**", "src/api/internal/platform/pgtest/**", "compose.yaml", ".env.example", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (pending)"]
tests_or_evidence: ["RED then GREEN: TestListResourcesEncodesAPage, TestListPassesTheCountryListAsAnArray, TestLoadAppBaseURL, TestVerificationMailCarriesALinkWhenABaseURLIsSet, TestResetMailCarriesALinkWhenABaseURLIsSet", "integration on real PostgreSQL: TestResourceProvenanceIsServed, TestResourceCountryListIsAUnion, TestResourceCountryListPaginatesStably, TestResourceCountryListRefusesBadInput", "make generate twice, no drift; make test-unit, make coverage (Go 100.0%), make test-integration-go exit 0"]
open_risks: ["No test fails if app.go stops passing WithBaseURL (wiring line only).", "Two filter-hash collision tests now use category instead of country (country is validated)."]
```

<!-- ai-usage:generated:start (do not edit by hand: `make ai-usage` rewrites everything between the two markers from docs/ai-usage/*.md) -->

## ADR-0003 amended: anonymous search limited per client network

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (owner's lead session, no subagent)"
purpose: "Owner decision of 2026-11-13 (item 1 of the PR96 audit, #117). ADR-0003 said anonymous search is limited to 60 per IP per minute, while the code limits per client network (/24 for IPv4, /48 for IPv6). The AI drafted the one-line amendment of ADR-0003: the per-network rule, the reason (it bounds the cost of an anonymous scan per network and cannot be dodged by changing the last byte), the trade-off (people behind one shared network share the budget) and when to revisit it. The owner decided the rule."
repository_or_source_context: "branch docs/decisions/search-limit-network, PR #123 (merged 2026-11-13)"
files_or_sections_affected: ["docs/decisions/0003-authentication-sessions-csrf-rate-limits.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: documentation only, no dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #123)"]
tests_or_evidence: ["scripts/check-docs.py PASS (PR #123)"]
open_risks: ["record written on 2026-11-15, after the merge: none was filed at the time"]
```

## AI-use records: one file per record, AI_USAGE.md generated

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Sonnet 5.5"
purpose: "Remove the conflict every pull request had on AI_USAGE.md (all of them appended at the end, so after any merge every other open pull request conflicted: five at once on 2026-11-13). A record is now one file in docs/ai-usage/; scripts/ai-usage.py rewrites a generated block at the end of AI_USAGE.md from those files (make ai-usage, make ai-usage-check); the date check of scripts/check-docs.py reads the record files too. The five conflicting pull requests (#99 to #102, #111) were brought up to date by a mechanical merge that kept develop's AI_USAGE.md and each branch's own new records."
repository_or_source_context: "pull requests #99, #100, #101, #102, #111 (conflicts only in AI_USAGE.md)"
files_or_sections_affected: ["scripts/ai-usage.py", "scripts/tests/ai-usage_test.sh", "scripts/check-docs.py", "scripts/tests/check-docs_test.sh", "docs/ai-usage/**", "AI_USAGE.md", "mk/60-review.mk", ".github/PULL_REQUEST_TEMPLATE.md", "AGENTS.md", "CLAUDE.md", "CONTRIBUTING.md", "docs/GIT_WORKFLOW.md", "docs/SHIP.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #120)"]
tests_or_evidence: ["scripts/tests/ai-usage_test.sh 16 passed and scripts/tests/check-docs_test.sh 20 passed, RED first (no script; the date check ignored the record files)", "make lint-scripts and python3 -I scripts/check-docs.py: see the pull request"]
open_risks: ["AI_USAGE.md is only as complete as its last build: the owner must run make ai-usage before the submission (SHIP.md says so); make docs-check runs the stale check on a pull request into main (not before, by design).", "Records already appended to AI_USAGE.md stay where they are; only new records use the files."]
```

## Issue #116 after-freeze follow-ups: skills delete scope, C14b matrix, cache headers

```yaml
date: "2026-11-13"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "The after-freeze items of issue #116 (PR #94 audit): (1) DeleteUserSkills wiped every user_skills row, including links to moderator-hidden tags that the read projections filter out — a hidden link was silently destroyed by an unrelated write and restoring the tag did not bring the skill back; the delete now joins tags and removes only active-tag links, proven on real PostgreSQL (link survives the rewrite, skill returns on tag restore). (2) The C14b matrix lacked a moderator row and a byte-for-byte 404 comparison — TestIdentityProfileNotFoundBodiesIdentical runs unknown, deleted, suspended, members-only-for-anonymous and private-for-member-and-moderator through the real router and compares status, Content-Type and the problem body minus request_id and instance. (3) GET /me (email, PII) and GET /profiles/{username} (caller-dependent body) now send Cache-Control: no-store so a future shared cache cannot leak a members-only 200 to an anonymous caller."
repository_or_source_context: "issue #116 (PR #94 audit follow-ups), branch fix/identity/profile-audit-116-after"
files_or_sections_affected: ["db/queries/identity.sql", "src/api/internal/platform/dbgen/identity.sql.go", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/identity/transport/profile.go", "src/api/internal/identity/transport/profile_test.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-13-issue-116-after-freeze.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #130)"]
tests_or_evidence: ["go test -race -count=1 ./internal/identity/... PASS", "go test -race -count=1 -tags integration -run 'TestIdentityProfile|TestIdentityProfiles' ./internal/platform/pgtest/ PASS on real PostgreSQL (TestIdentityProfileHiddenTagFiltered extended: hidden link survives, restore brings skill back; TestIdentityProfileNotFoundBodiesIdentical: 7-case matrix)", "make generate regenerated dbgen (comment + scoped DELETE), check-sqlc drift committed", "new TestGetMeAndGetProfileSendNoStore pins the header", "go vet, gofmt clean"]
open_risks: ["the undeclared 422 on getProfile (validator rejects a malformed path username; the contract lists no 422 response) is a contract change left for the owner before/after freeze decision", "resolveSkills still runs outside the write transaction: a tag hidden between resolve and insert is linked but filtered on read — accepted and documented, no leak", "hidden-tag links now accumulate until the tag is restored or purged; no purge exists for user_skills (same as before for active links)"]
```

## Issue #116 before-freeze follow-ups: profile patch validation and violation contract

```yaml
date: "2026-11-13"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "The two before-freeze items of issue #116 (PR #94 audit): (1) NUL or invalid UTF-8 in display_name/bio passed the schema and died on the Postgres 22021 text check as a 500 that also burnt one of the 60 profile:write hits per minute — validateProfileWrite now rejects !utf8.ValidString or a NUL rune with 422 invalid_value on /display_name and /bio; (2) the service emitted violation codes outside the OpenAPI Violation.code enum (invalid_length, too_many, unknown_value) and bare field names — every violation now uses an enum code (too_short, too_long, too_many_items, invalid_format, invalid_value, duplicate) and the JSON pointer of the request member (/skills, /field_visibility/bio, /country…) so the S7 form can map them. pgtest asserts the problem body's violations against the enum and pointer shape on the real router."
repository_or_source_context: "issue #116 (PR #94 audit follow-ups), branch fix/identity/profile-audit-116"
files_or_sections_affected: ["src/api/internal/identity/service/profile.go", "src/api/internal/identity/service/profile_test.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-13-issue-116-profile-validation.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #127)"]
tests_or_evidence: ["go test -race -count=1 ./internal/identity/... PASS (new TestUpdateMyProfileViolationsUseContractCodesAndPointers: 16 cases)", "go test -race -count=1 -tags integration -run 'TestIdentityProfiles|TestIdentityProfile' ./internal/platform/pgtest/ PASS on real PostgreSQL (NUL bio -> 422, assertProblemViolations on unknown skill/country)", "go vet clean, gofmt clean"]
open_risks: ["the three after-freeze items of issue #116 stay open: skills DELETE removes hidden-tag links, the C14b matrix lacks a moderator row and a byte-for-byte 404 body comparison, and the Cache-Control/Vary + resolveSkills-outside-transaction notes", "invalid UTF-8 cannot arrive through encoding/json (it coerces to U+FFFD); the utf8.ValidString half of validText is defence in depth for non-HTTP callers"]
```

## S4 Learn: articles, series, links writer and the free-series content script

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "S4 Learn: listLearnArticles, createLearnArticle (401 anonymous, 403 publishing without a verified email, drafts allowed, shared write:user budget failing closed, Idempotency-Key, original text stored unchanged), getLearnArticle (draft or archived only to its author), listTrainingSeries, getTrainingSeries (no my_progress, C3); the article links[] writer (canonical question/resource/project -> article, target locked FOR SHARE in the creating transaction); the one-time free series content script and its gated wrapper; the S1/S2 tracker ticks."
repository_or_source_context: "branch feature/learn/articles-series"
files_or_sections_affected: ["db/queries/learn.sql", "src/api/internal/platform/dbgen/learn.sql.go", "src/api/internal/platform/dbgen/learn_queries_test.go", "src/api/internal/learn/**", "src/api/internal/app/server.go", "src/api/internal/app/app.go", "src/api/internal/app/learn_internal_test.go", "src/api/internal/platform/pgtest/learn_integration_test.go", "src/api/internal/platform/pgtest/errorreport_integration_test.go", "scripts/content/**", "scripts/tests/free-series_test.sh", "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt", "TODO.md", "docs/ai-usage/2026-11-13-s4-learn.md"]
third_party_sources_consulted: []
license_review: "ACCEPTED_BY_OWNER (2026-11-15): the four French lessons of scripts/content/free-series.sql, not independently verified"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #111)"]
tests_or_evidence: ["RED: service tests failed to compile before the service existed; pgtest Learn tests answered 501 with the handler unwired", "make generate, lint-go, lint-contract, lint-web, lint-scripts, typecheck, docs-check PASS", "make test-unit PASS; make coverage PASS: Go first-party 100.0% (4061/4061)", "make test-integration-go PASS (TestLearn*, TestFreeSeriesContentScript, EXPLAIN plans logged by TestLearnListPlansAreBounded)", "make security-authz, security-sast, security-govulncheck, security-secrets, licenses PASS", "scripts/tests/free-series_test.sh 13/13"]
open_risks: ["the composition root (src/api/internal/app), pgtest, dbgen tests and scripts/tests are outside the S4 allowed paths as written; S2 used the same files", "series have no tags and no relations at P0 (no tables): the API returns empty arrays", "links to a series or an article are refused with 422 (content_relations has no such pair)", "article bodies above about 16 KiB replay with Idempotency-Body-Omitted (idempotency column limit)", "deleted authors are not shown as 'membre supprimé' yet (S1 bylines item)", "the content script has not run on staging-test or production-live"]
```

## S7 frontend PR 3: landing, member home and parametres

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (integrator, parametres) and Claude Sonnet 5.5 (landing and home lanes in separate worktrees)"
purpose: "S7 PR 3: public landing page and member home on a session-aware index route, profile visibility setting on parametres (C14b), settings moved to /parametres."
repository_or_source_context: "branch feature/platform/web-home-settings; lane branches feature/platform/web-landing and feature/platform/web-home-feed; prototypes/html for copy"
files_or_sections_affected: ["src/web/src/features/{landing,home,settings}/**", "src/web/src/app/**", "src/web/src/i18n/**", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["installed @astryxdesign/cli 0.7.14 component docs"]
license_review: "NOT APPLICABLE: original code and text, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #102)"]
tests_or_evidence: ["RED then GREEN (ProfileVisibility.test.tsx, LandingPage.test.tsx, HomePage.test.tsx, translate.test.ts key-collision check)", "Real-browser journeys on make up-local-dev with Mailpit: sign-up, verify, sign-in, sign-out, wrong password, reset, sign-in with the new password; axe on 13 pages; no horizontal scroll at 360/768/1024/1440 and 200% zoom (56/56)"]
open_risks: ["Demo examples on the member home stay until listLearnArticles, listQuestions and listProjects are live.", "The local display-name field on parametres is a prototype leftover kept for now (owner decision)."]
```

## S7 frontend PR 2: public reads (Découvrir, resource, recherche)

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (integrator) and Claude Sonnet 5.5 (web-implementer lanes in separate worktrees)"
purpose: "S7 PR 2: Découvrir with the West Africa default and one-click whole-continent view, resource detail with dated sources and verification status, grouped search with a marked assistant mock; Découvrir becomes a live pillar."
repository_or_source_context: "branch feature/discover/web-public-reads; lane branches feature/discover/web-catalogue and feature/discover/web-search; prototypes/html for layout and copy"
files_or_sections_affected: ["src/web/src/features/discover/**", "src/web/src/features/search/**", "src/web/src/app/{navigation,Shell}.tsx", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["installed @astryxdesign/cli 0.7.14 component docs"]
license_review: "NOT APPLICABLE: original code and text, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #101)"]
tests_or_evidence: ["RED then GREEN per file (region.test.ts, DiscoverPage.test.tsx, ResourcePage.test.tsx, SearchPage.test.tsx, searchQuery.test.ts)", "pnpm --filter web run coverage exit 0"]
open_risks: ["The assistant answer is a labelled sample until S9 is live.", "West Africa default filters on the client until the API accepts several countries."]
```

## S7 frontend PR 1: shell and identity pages

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (integrator) and Claude Sonnet 5.5 (web-implementer lanes in separate worktrees)"
purpose: "S7 PR 1: CSRF-aware API helper (X-CSRF-Token on every unsafe call, token in memory only, one retry after CSRF_FAILED), problem+json mapping, session hooks, Astryx AppShell with pillar emphasis and coming-soon areas, shared states, and the six identity pages wired to the live S1 operations."
repository_or_source_context: "branch feature/platform/web-shell-identity; lane branch feature/identity/web-account-pages; prototypes/html for layout and copy"
files_or_sections_affected: ["src/web/src/**", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["installed @astryxdesign/cli 0.7.14 component docs", "openapi-fetch 0.17 type declarations"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #100)"]
tests_or_evidence: ["RED then GREEN per file with Vitest (api/http.test.ts, app/Shell.test.tsx, features/identity/*.test.tsx)", "pnpm --filter web run lint/format:check/typecheck/test/coverage exit 0; coverage 96.94/87.29/98.75/98.38"]
open_risks: ["Visual, axe and responsive evidence (ADR-0007) not produced in jsdom.", "Mails carry a bare token until APP_BASE_URL links land (fix/platform/s7-api-gaps)."]
```

## Search candidate bound and cursor digest (#118 before-freeze items)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "Issue #118 before-freeze items: bound the search ranking (tsquery built once, no read for a query without a positive term, at most 500 candidates ranked), worst-case tests that measure rows scanned instead of the presence of a Limit node, a redacted 500 on a query timeout, and a fixed 16-byte filter digest in cursors so they stay within the contract's 512 characters."
repository_or_source_context: "branch fix/discover/search-bounds, issue #118"
files_or_sections_affected: ["db/queries/search.sql", "src/api/internal/platform/dbgen/search.sql.go", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/platform/cursor/cursor.go", "src/api/internal/platform/cursor/cursor_test.go", "src/api/internal/search/repository/repository_test.go", "src/api/internal/search/transport/handler_test.go", "src/api/internal/platform/pgtest/search_read_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-13-search-bounds.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #126)"]
tests_or_evidence: ["RED: TestSearchQueriesUseTheGINIndexes scanned 6000 rows for 'bruit' and matched every row for '-zebra'; TestCursorLengthDoesNotGrowWithTheFilters produced a 1222-character cursor", "GREEN: see the PR evidence table"]
open_risks: ["above 500 matches the ranked set is the first 500 rows the index returns, not the 500 best", "a query with only negations now returns no hits instead of every row", "cursors issued before the deploy are refused once (INVALID_CURSOR)"]
```

## Traefik forwarded headers pinned (#117 item 2)

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "Issue #117 item 2: pin forwardedHeaders.insecure=false on every Traefik entrypoint (local-dev, staging-test, production-live) and enforce it, plus no trusted upstream proxy, in scripts/verify-containers.sh; record the staging spoofed X-Forwarded-For smoke as a release-checklist step for the owner."
repository_or_source_context: "branch chore/ops/traefik-forwarded-headers, issue #117"
files_or_sections_affected: ["compose.yaml", "scripts/verify-containers.sh", "TODO.md", "docs/ai-usage/2026-11-13-traefik-forwarded-headers.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: configuration and test only, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #128)"]
tests_or_evidence: ["RED: scripts/verify-containers.sh failed on the 3 profiles (entrypoints web and websecure did not pin forwardedHeaders.insecure=false)", "GREEN: scripts/verify-containers.sh ALL PASS", "the pinned traefik:v3.6 image accepts --entrypoints.web.forwardedHeaders.insecure=false and rejects an unknown forwardedHeaders field"]
open_risks: ["the end-to-end spoofed X-Forwarded-For smoke through Traefik is not run (no deployed host yet): release-checklist step"]
```

## Third-party licences verified instead of accepted unverified

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (owner's lead session, no subagent)"
purpose: "#87 had closed the licences as accepted without verification. The AI read each licence from its upstream source on 2026-11-13: the GitHub licence API of each repository, and the licence file itself where the API answered NOASSERTION (PostgreSQL, Node.js, goose). This covered the container images, the Go tool entries, go.yaml.in/yaml/v3 (Apache-2.0 with MIT libyaml portions), the CI and Make tools, claude-code-action and the 431 npm development dependencies (none strong copyleft). It wrote THIRD_PARTY_NOTICES.md with a Verification section stating the method and its limits. golangci-lint and ShellCheck (GPL-3.0) and Semgrep (LGPL-2.1) are recorded as build and CI tools, not distributed or linked. The AI-generated brand assets stay an owner decision."
repository_or_source_context: "branch chore/docs/verify-licences, PR #88 (merged 2026-11-13)"
files_or_sections_affected: ["THIRD_PARTY_NOTICES.md", "TODO.md"]
third_party_sources_consulted: ["GitHub licence API of each upstream repository", "upstream licence files of PostgreSQL, Node.js and goose"]
license_review: "VERIFIED: read from the upstream licence API or licence file (method in THIRD_PARTY_NOTICES.md, Verification)"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #88)"]
tests_or_evidence: ["THIRD_PARTY_NOTICES.md Verification section; runtime Go and npm dependencies still come from make licenses"]
open_risks: ["record written on 2026-11-15, after the merge: none was filed at the time", "the brand assets' licence stays an owner decision"]
```

## Contract violation codes and pointers in Learn and discover reports

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "Align the Learn and discover error-report 422 violations with the contract, as #127 did for identity (issue #116 rule): JSON-pointer fields for body members and Violation.code enum values instead of 'invalid'; query parameters keep their names. Also marks S4 DONE in TODO.md (PR #111 merged)."
repository_or_source_context: "branch fix/platform/violation-codes; review of PR #127"
files_or_sections_affected: ["src/api/internal/learn/service/service.go", "src/api/internal/learn/service/service_test.go", "src/api/internal/discover/service/reports.go", "src/api/internal/discover/service/reports_test.go", "src/api/internal/discover/service/service.go", "src/api/internal/platform/pgtest/learn_integration_test.go", "src/api/internal/platform/pgtest/errorreport_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-13-violation-codes.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #129)"]
tests_or_evidence: ["RED: TestCreateViolationsUseContractCodesAndPointers and TestCreateErrorReportValidation failed on 'invalid' codes and bare field names", "GREEN: see the PR evidence table"]
open_risks: ["identity's password and new_password violations still use 'invalid' (identity area, flagged for NICE-DEV226)"]
```

## Contract: an application message is not empty (#135)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session, no subagent)"
purpose: "Owner decision on #135: the contract allowed an empty ApplicationCreate.message while applications_message_ck and the service required 1 to 2000 characters. The contract now says minLength 1; a test pins it."
repository_or_source_context: "branch fix/build/application-message-min; issue #135 (PR132 notes); review batch #157 item 10"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apispec/openapi.gen.yaml (generated)", "src/api/internal/platform/apispec/apispec_test.go", "src/api/internal/build/service/applications.go (comment)", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #164)"]
tests_or_evidence: ["RED: TestApplicationMessageIsNotEmpty failed on minLength 0", "GREEN: apispec, build packages and TestBuildApplication* (real PostgreSQL) pass; make lint-contract and docs-check exit 0"]
open_risks: ["none: the server already refused an empty message with 422; only the documented contract changes"]
```

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
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #195)"]
tests_or_evidence: ["RED then GREEN: TestRemoveGuarantees, TestAskEvery503CarriesRetryAfter, TestUserPromptCarriesEachPassageDate, TestParseVerifyNoticeFollowsTheCitedSources, TestSearchHitsCarryThePublicationDate, golden G14 to G17, AssistantPanel date tests", "Go coverage 100.0 % (6239/6239); lint-go, race unit and integration exit 0"]
open_risks: ["guarantee detection is a FR/EN pattern list: other phrasings pass through", "a bare marker left alone after a removed guarantee keeps its source cited", "the real-provider run (V4) is NOT RUN"]
```

## Auto-fix runner: sandbox dependencies and label permissions

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent), lead session Claude Opus 5.5"
purpose: "Owner-approved pipeline fix after the first live auto-fix run (PR #199, run 43625357930): the fixer could not run a test or commit because Claude Code's Linux Bash sandbox needs bubblewrap and socat, and the labels autofix-N and needs-owner failed with HTTP 403 because labelling a pull request needs pull-requests: write. The two fixer jobs install the two packages from the runner's apt before the Claude step; the decide job gets pull-requests: write; the labels written after the push move from the publish job (which checks the branch out) to a new label job with no checkout and no secret."
repository_or_source_context: "branch fix/review/autofix-runner"
files_or_sections_affected: [".github/workflows/review-autofix.yml", ".github/workflows/review-batch.yml", "scripts/tests/review-runner_test.sh", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency in the repository (two distro packages installed on the runner at run time)"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #221)"]
tests_or_evidence: ["RED: scripts/tests/review-runner_test.sh failed 4 assertions on develop f0b4e79", "GREEN: review-runner_test.sh 9 passed, review-agent-triggers_test.sh, make lint-scripts, make docs-check and a YAML parse of both workflows"]
open_risks: ["NOT RUN on a real runner: the install step and the label job are checked statically only until the next auto-fix run", "actionlint is not installed locally: NOT RUN", "review-automerge.yml writes no label from its YAML (the script labels with its own token), so it is unchanged"]
```

## Review batch: the "nothing to fix" path

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session and the ci-implementer subagent)"
purpose: "A review batch whose fixer makes no commit (run 43633446506, batch #226) settled nothing and looped forever. The trusted publish job now applies the fixer's verdicts with the rules of settle, posts one verdict table on the batch issue, ticks the items with a verdict and closes the batch, and new batches leave the owner-held items out until they change. Also two MEDIUM findings from the review of #223: FIXED ELSEWHERE naming the source's own pull request, and a CRITICAL whose override cannot be read."
repository_or_source_context: "branch fix/review/batch-nothing-path from develop 68b3312 (after #223); batch issue #226; owner decision 2026-11-14 ~19:55 and ~20:05 GMT"
files_or_sections_affected: [".github/workflows/review-batch.yml (publish job)", "scripts/review-autofix.py", "scripts/findings.py", "scripts/review-gate.py", "scripts/tests/review-batch-nothing_test.sh", "scripts/tests/findings-batch_test.sh", "scripts/tests/findings-settle_test.sh", "scripts/tests/review-gate_test.sh", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
human_reviewers: ["Y4NN777 (pipeline owner, merge click)", "Y4NN777 (owner closure, 2026-11-15; merged in #229)"]
closure: "closed by the owner (2026-11-15)"
tests_or_evidence: ["RED then GREEN: review-batch-nothing_test.sh (new), findings-batch_test.sh (held items), findings-settle_test.sh (own source PR), review-gate_test.sh (events read error)", "make lint-scripts", "make docs-check", "YAML parse of review-batch.yml"]
open_risks: ["NOT RUN on a real runner", "the publish job is not in the review-batch-write concurrency group: a batch run reading at the same moment can race the close", "batch #226 stays open until one manual dispatch after the merge", "FIXED ELSEWHERE is refused when the source was filed after the fixing pull request merged (fails safe, a human checks)"]
```

## E5: Connect people directory (contract and backend)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead + go-implementer (Sonnet)"
purpose: "E5 Connect, read-only people directory: the listProfiles contract (GET /profiles, 49 operations, 45 P0), the authorization matrix row, the public route entry, and the backend in the identity module (visibility-filtered SQL projection and filters, keyset cursor, rate limit) with unit, handler and real-PostgreSQL privacy tests."
repository_or_source_context: "branch feature/connect/people-directory; TODO.md E5; ROADMAP E5 and C14b; ADR-0003, ADR-0004, ADR-0005"
files_or_sections_affected: ["api/openapi.yaml", "db/queries/identity.sql", "src/api/internal/identity/**", "src/api/internal/app/**", "src/api/internal/platform/{apigen,apispec,dbgen,httpx,pgtest}/**", "src/web/src/api/schema.d.ts", "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "docs/spec-kit/AUTHORIZATION_MATRIX.md", "docs/quality/COVERAGE_SCOPE.md", "TODO.md", "ROADMAP.md", "docs/SPEC.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code and text, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["mibienpanjoe (backend)", "Y4NN777 (contract and matrix, R7)", "Y4NN777 (owner closure, 2026-11-15; merged in #193)"]
tests_or_evidence: ["RED: 20 service tests failed on a stub before the implementation (assertion failures)", "GREEN: go test -race ./... exit 0; scripts/coverage-go.sh 100.0% (6328/6328)", "go test -race -tags integration ./src/api/internal/platform/pgtest -run TestDirectory exit 0, EXPLAIN plans logged"]
open_risks: ["The 60 per minute window is not in ADR-0003: the owner confirms it", "Demo (provenance demo) accounts are not excluded from the directory outside local-dev: the owner decides (K39)"]
```

## ADR-0013 default branch develop, pipeline trust anchor

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead + ci-implementer (Sonnet)"
purpose: "ADR-0013 (default branch develop so review stages 3 to 5 run), the docs and comments that followed from it, and a static guard test keeping jobs of the review-agent environment out of pull request events."
repository_or_source_context: "branch chore/review/default-branch-develop; owner decisions of 2026-11-14; TODO AI review block; GitHub documentation on workflow triggers and the default branch"
files_or_sections_affected: ["docs/decisions/0013-default-branch-develop.md", "docs/decisions/README.md", "docs/GIT_WORKFLOW.md", "docs/quality/REVIEW_CONTEXT.md", ".github/workflows/review-autofix.yml", ".github/workflows/review-automerge.yml", ".github/workflows/review-batch.yml", "scripts/close-merged-issues.py", "scripts/tests/review-agent-triggers_test.sh", "TODO.md"]
third_party_sources_consulted: ["https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows", "https://github.blog/changelog/2025-11-07-actions-pull_request_target-and-environment-branch-protections-changes/", "https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository"]
license_review: "NOT APPLICABLE: original text and code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #185)"]
tests_or_evidence: ["scripts/tests/make_test.sh pins test-harness (#184): RED on a copy of mk/ with test-harness and the harness group removed (48 passed, 4 failed; the old test passed 49/0), GREEN on develop 52 passed", "RED: scripts/tests/review-agent-triggers_test.sh on a pull_request_target fixture exits 1", "GREEN: same test on .github/workflows exits 0", "RED (fail-closed pass): 18 new fixtures (expression, anchor and alias environments, flow-mapping and odd-indent jobs, unsafe if: forms, always() with needs, unreadable on:, workflow_call) accepted by the first version, 38 of 38 checks GREEN after the rewrite; mutation: review-batch.yml copy without the plan guard is refused", "RED (review of #185): 4 new fixtures (multi-line if:, multi-line && guard, multi-line if: in a needed job, quote-protected hash) accepted by the previous guard, 47 of 47 GREEN after the fix", "RED (review of #185 round 2): 1 new fixture (plain scalar with a hash before a quote, guard text after it) accepted by the previous guard (50 passed, 1 failed), 51 of 51 GREEN after the fix; the old cmt_single expectation (hash inside a quote of a plain scalar is not a comment) was wrong per YAML 1.2 and is now expect_good; the quoted-continuation fixture was already refused (conjuncts() rejects any quote-leading if:), so it is a pin, not a RED", "actionlint: not installed (command -v actionlint empty), NOT RUN"]
open_risks: ["stages 3 to 5 NOT RUN until the owner switches the default branch", "REVIEW_AGENT_TOKEN is the owner's own token (owner decision 2026-11-14, ADR-0013): actions taken with it are recorded as the owner's; three jobs read it (automerge, publish of review-autofix.yml and review-batch.yml), none approves", "review-automerge.yml lines 13 and 47 said only the auto-merge job reads the token: corrected in the same PR", "scripts/findings.py docstring (~lines 415-416) said closing keywords act only on main: fixed in the register-detail PR (chore/review/register-detail, commit 9958fc9)"]
```

## Demo gaps A, B and C

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "OpenAI Codex CLI"
purpose: "Close the demo gaps selected by the owner before the 20:00 GMT freeze: rate limits, frontend connections and, if green before the cut line, moderation."
repository_or_source_context: "branch fix/platform/demo-gaps; owner plan twinkling-strolling-parrot.md; TODO S2, S6 and S7"
files_or_sections_affected: ["src/api/**", "src/web/**", "docs/runbooks/moderation.md", "TODO.md", "ROADMAP.md", "docs/architecture/ARCHITECTURE.md", "mk/50-ops.mk"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["PANK4SS", "mibienpanjoe", "Y4NN777 (owner closure, 2026-11-15; merged in #220)"]
tests_or_evidence: ["RED then GREEN focused tests for rate limits and frontend gaps", "moderation RED: error_report.resolved audit action; GREEN: error_report.resolve plus unit, authz and real-PostgreSQL tests", "required Go, web and docs gates recorded in the pull request"]
open_risks: ["manual visual and responsive checks are NOT RUN until recorded"]
```

## Docs state sync after the 12:00 checkpoint

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead agent"
purpose: "Find stale statements in the trackers and project docs and align them with the merged pull requests and the owner decisions of the 12:00 checkpoint (ROADMAP section 14); draft the issue updates for the owner to apply. Documentation only: no scope, contract or decision change."
repository_or_source_context: "branch chore/docs/state-sync from develop a71aa9e; JSON dumps of the open issues, the last 60 merged pull requests and the open pull requests, supplied by the lead session; scripts/unimplemented-routes.txt, scripts/authz-tests.txt, pgtest migrate tests and docs/ai-usage records read for evidence"
files_or_sections_affected: ["TODO.md (Now row of Y4NN777, owner checklist H16 and PRD names, Shared pieces, S1, S3, S5, S6, S7, S9, SEED, E4, E5, Done log)", "ROADMAP.md (sections 1, 4, 8 H13, 10 C10 O5 O6, 11 K29, 13)", "README.md (Statut, Contribuer)", "docs/team/KICKOFF.md (sections 1, 4, 5)", "docs/SPEC.md (US-LEARN-002 note, US-INTEL-001 ADR status, section 9)", "docs/spec-kit/AUTHORIZATION_MATRIX.md (roles line, status line)", "docs/spec-kit/DATABASE_DESIGN.md (status lines, migration 00013)", "docs/spec-kit/TRACEABILITY.md (intelligence row, approved rows note)"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: project documentation only, no third-party text"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #186)"]
tests_or_evidence: ["NOT RUN: make docs-check (the agent has no shell; the lead session runs it before the commit)"]
open_risks: ["the By column of the new Done log rows comes from docs/ai-usage and AI_USAGE.md contributor fields and the slice owners, not from the GitHub PR authors (not in the dumps)", "ADR-0013 is not referenced: its file is written in another pull request"]
```

## Environment hosts in the encrypted secrets (staging and production)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session, no subagent)"
purpose: "With the owner's explicit permission, set APP_BASE_URL and PUBLIC_HOST in the staging secrets and create the production secrets file, using SOPS without reading or printing any value (sops set; scripts/secrets.sh init generates the database password, DATABASE_URL and CURSOR_KEY). Records the staging host decision (H7)."
repository_or_source_context: "branch chore/ops/environment-hosts; issue #109 section 1; ADR-0009; TODO H6/H7"
files_or_sections_affected: ["secrets/staging-test.enc.env (encrypted)", "secrets/production-live.enc.env (encrypted, new)", "TODO.md", ".claude/settings.json (deny list: reading the encrypted secrets/*.enc.env is allowed, ciphertext and SOPS metadata only, so the session can check names and recipients; the age key directory **/sops/age/** and /dev/shm/** are denied; every Edit deny on secrets stays)"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: configuration, no dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner, secrets)", "Y4NN777 (owner closure, 2026-11-15; merged in #180)"]
tests_or_evidence: ["scripts/secrets.sh check: check-secrets-encrypted PASS", "make security-secrets: no leaks found (working tree and history of HEAD)", "no secret value was displayed: sops set and init print nothing but a 'Created' line"]
open_risks: ["226os.tech and staging.226os.tech have no DNS A record yet", "AI_API_KEY stays empty until the owner fills it; ACME_EMAIL (both files) and production SMTP were set by the owner with hidden input, never seen by the session", "production secrets are encrypted to the current recipients only: add the host age key, rotate, and regenerate POSTGRES_PASSWORD and CURSOR_KEY before deploying"]
```

## #125 copy-edit pass on the free-series lessons

```yaml
date: "2026-11-14"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "Issue #125 team review of the four AI-drafted lessons in scripts/content/free-series.sql before the staging rehearsal. The lessons were rewritten for engineers by mibienpanjoe on 2026-11-14 (record 2026-11-14-free-series-lessons.md); this pass is a French copy-edit of that text: the whole file was read, two corrections applied in lesson 1 ('effort consenti' -> 'effort de raisonnement'; tokens are text fragments, images and audio only enter through multimodal models). Everything else read as idiomatic, natural French; no structural or factual change."
repository_or_source_context: "issue #125; branch docs/learn/free-series-copyedit; scripts/content/free-series.sql lessons 1-4"
files_or_sections_affected: ["scripts/content/free-series.sql", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: edits to original team text, no new content or dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["owner approval of the final text still open on #125", "Y4NN777 (owner closure, 2026-11-15; merged in #204)"]
tests_or_evidence: ["bash scripts/tests/free-series_test.sh: 13 passed, 0 failed", "read-through of all four lessons; edits confined to lesson 1"]
open_risks: ["text remains AI-drafted under mibienpanjoe's direction with this copy-edit; PANK4SS's read pass and the owner's final approval are still open on #125"]
```

## Free Learn series: the four lessons rewritten for engineers (#125)

```yaml
date: "2026-11-14"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 then Claude Fable 5.2 (main session, no subagents)"
purpose: "Issue #125: the four lessons of scripts/content/free-series.sql were judged outdated and too shallow for an engineering audience. The contributor set the direction (address engineers; cover harnesses, agentic systems, always-on agents and models that are no longer text-only). AI-drafted: the full French text of the four lessons, the series title and summary, the four lesson titles and summaries. Human part: the direction above, then a full read and approval of the text by the contributor, without line edits. Structure changes: series retitled 'Des LLM aux agents : l'ingénierie de l'IA aujourd'hui' (the idempotency guard no longer depends on the title: it refuses when a platform series on topic ai already exists, after the review finding deep-free-series-guard-title), level beginner -> intermediate, tags ia/llm/prompt/debutant -> ia/llm/agents/harness. Still 4 tutorial lessons, language fr, topic ai."
repository_or_source_context: "branch docs/learn/free-series-lessons; issue #125 (from PR #111, S4 Learn)"
files_or_sections_affected: ["scripts/content/free-series.sql", "src/api/internal/platform/pgtest/learn_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-14-free-series-lessons.md"]
third_party_sources_consulted: []
license_review: "Original text written for 226os; no passage copied or translated from a third-party source. Vendor-neutral: no product or model named, except the open protocol MCP and the AGENTS.md file convention."
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["mibienpanjoe (read and approved the text, 2026-11-14)", "Y4NN777 (owner closure, 2026-11-15; merged in #190)"]
tests_or_evidence: ["scripts/tests/free-series_test.sh 13 passed, 0 failed", "RED: TestFreeSeriesContentScript failed (run after a retitled series: nil) with the title-only guard; GREEN after the guard on origin and topic (ephemeral PostgreSQL)", "excalibur run: verified, known debt in forge-check", "screenshots read after loading the series with the script: code block, inline code, lists and table render correctly"]
open_risks: ["rendering was checked on a throwaway local stack only (temporary PostgreSQL, API and Vite dev server; not the nginx and Traefik front): lessons 1 to 4 at 390 px or 1280 px; the body no longer repeats the lesson title as a first heading (the page already shows it); the lesson 4 table scrolls sideways inside its own box on a phone", "technical statements about models and agents are general and undated; they need a re-read when the field moves"]
```

## Declare the 422 on getProfile in the contract

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (main session, no subagents)"
purpose: "Issue #116 owner decision: the request validator already answers 422 to a malformed path username, but the contract listed no 422 on getProfile. The response is now declared ($ref Unprocessable) and the committed generated outputs are refreshed (embedded spec copy, web schema types). A test pins it."
repository_or_source_context: "branch fix/identity/profile-422-contract; issue #116"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apispec/openapi.gen.yaml", "src/api/internal/platform/apispec/apispec_test.go", "src/web/src/api/schema.d.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #134)"]
tests_or_evidence: ["RED: TestGetProfileDeclaresTheValidatorsUnprocessable failed before the contract change", "GREEN: apispec, apigen and identity tests pass; make lint-contract passes (redocly valid, migration policy OK)"]
open_risks: ["a contract change (R7): the web client types change by one response line, no behaviour change on the server"]
```

## S5 optional GitHub token for the repository check

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer subagent)"
purpose: "An optional GITHUB_TOKEN so project registration stops sharing the unauthenticated 60 requests/hour/IP GitHub quota, with redaction, a fixed host and a one-time unauthenticated retry when the token is rejected."
repository_or_source_context: "branch feature/build/github-token; TODO S5 choices (no token at first); K10 anti-SSRF rules"
files_or_sections_affected: ["src/api/internal/build/github/**", "src/api/internal/platform/config/**", "src/api/internal/app/**", "src/api/internal/build/build_test.go", "compose.yaml", ".env.example", "scripts/compose.sh", "scripts/tests/compose-wrapper_test.sh", "secrets/staging-test.enc.env", "secrets/production-live.enc.env", "README.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (K10 co-review)", "Y4NN777 (owner closure, 2026-11-15; merged in #208)"]
tests_or_evidence: ["RED: compile failures before New(token, logger) and LoadGitHubToken existed", "GREEN: TestRepoSendsTheTokenOnlyWhenSet, TestRepoRetriesOnceWithoutTheTokenOn401, TestClientFormattingHidesTheToken, TestLoadGitHubTokenValues; lint-go, race unit, pgtest Build/Project exit 0; Go coverage 100.0 % (6271/6271)"]
open_risks: ["not tested against the real GitHub API with a token", "a revoked token silently falls back to the shared 60/h quota (warn log github_token_rejected)", "the token is tied to the owner's account until a GitHub App replaces it"]
```

## Harness tests in CI

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent)"
purpose: "Fix the agent inventory test for the new ci-implementer agent, and run the harness tests (.claude/hooks/tests) in CI through a new make target test-harness, a step of make quality and of the CI group harness."
repository_or_source_context: "branch chore/claude/ci-implementer-inventory; owner-approved pipeline task; ADR-0010 (CI groups); mk/40-quality.mk"
files_or_sections_affected: [".claude/hooks/tests/test_domain_skills.py", "mk/20-gates.mk", "mk/40-quality.mk", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #183)"]
tests_or_evidence: ["RED: python3 -m unittest discover -s .claude/hooks/tests failed, ci-implementer missing from the expected set", "GREEN: same command exit 0 (28 tests)"]
open_risks: ["the harness CI group is NOT RUN in CI until this pull request runs", "scripts/tests/make_test.sh was not touched: it pins the repo group steps, so the step lives in its own group"]
```

## AI review: notes in the findings register

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (main session, no subagents)"
purpose: "The review ledger keeps two kinds of entry: findings (defects) and notes (remarks), each with an evidence tag OBSERVED or INFERRED. Open notes become one 'PR<N> notes' checklist issue per pull request on merge (label kind:note, due after the freeze); the dashboard has a separate Notes section and flags INFERRED notes undecided for 7 days; the review pack lists decided notes (ticked, or closed as not planned) so they are not raised again. A note never triggers a new review or an owner ping. Also adds the owner task for the GitHub API token (S5) to the release-day checklist."
repository_or_source_context: "branch chore/review/notes-in-register; owner request after the review-register work (#113, #115)"
files_or_sections_affected: ["scripts/findings.py", "scripts/review-context.py", "scripts/review-ping.py", "scripts/tests/findings_test.sh", "scripts/tests/review-context_test.sh", "scripts/tests/review-ping_test.sh", ".claude/commands/delegate-review.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #133)"]
tests_or_evidence: ["RED: the new cases in findings_test.sh, review-context_test.sh and review-ping_test.sh failed before the change", "GREEN: findings 40 passed, review-context 50 passed, review-ping 13 passed"]
open_risks: ["NOT RUN end to end until the first merged pull request after this PR: the model must actually write kind and evidence in the ledger (an entry without them is read as a finding, so nothing breaks, but no note is filed)", "the kind:note label is created by the issue API on first use; the board 4 Auto-add filter stays label:finding, so notes are not on the board"]
```

## Contract: link types narrowed to the stored pairs

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session, no subagent)"
purpose: "Owner decision on S3: the contract offered question and series link targets that content_relations cannot store (422 at runtime). QuestionCreate and QuestionUpdate now take QuestionLinkRef (resource, project, article) and LinkRef drops series; the generated Go and TypeScript types follow."
repository_or_source_context: "branch fix/solve/question-link-types; db/migrations/00009_content_relations.sql (content_relations_pair_ck); TODO S3 and S4 deviations"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apigen/apigen.gen.go (generated)", "src/api/internal/platform/apispec/openapi.gen.yaml (generated)", "src/api/internal/platform/apispec/apispec_test.go", "src/web/src/api/schema.d.ts (generated)", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #141)"]
tests_or_evidence: ["RED: TestLinkTypesMatchTheStoredPairs failed on develop's embedded spec (LinkRef items, series present)", "GREEN: the same test and TestEmbeddedSpecMatchesContract pass after scripts/check-generated.sh generate"]
open_risks: ["a client that sent series or a question-to-question link now fails schema validation (422) instead of the service's 422; no web code sends them"]
```

## Review register: self-contained issues (detail, fix, source, permalink)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent), lead session Claude Opus 5.5; the follow-up round on the independent review of a32ec20 (H1-H7, M1-M6) by the ci-implementer subagent on Claude Opus 5.5"
purpose: "Owner-approved pipeline task: the AI-review ledger and the issues built from it lose the proof, the recommended fix and the exact comment (issue #172 against the last review comment of PR #168); a stage 1 note repeated by the deep review was re-tagged source deep (PR #180). Entries now carry detail and fix, the script sets source, first_sha and location, texts are neutralised and secret shapes redacted, issues are self-contained, duplicates (same id, file and line) across pull requests are one batch row; a missing detail or fix never hides a verdict, and the fixer receives the context. Follow-up (independent review of a32ec20, owner's safe subset): the deep review covers the whole pull request when the last deep review did not pass on the last reviewed commit; the no-click batch auto-merge on sensitive paths is withdrawn and the pipeline's own tests are pipeline and owner paths; only `fixed` clears a HIGH or CRITICAL; settle closes or ticks only FIXED and FIXED ELSEWHERE; fixer commit messages and batch bodies neutralise every issue reference form; readers and writers take the last ledger block and an unusable ledger keeps the previous one; unknown status is open and unknown sev MEDIUM everywhere; a deep CRITICAL brings its file and line; a re-list never lowers an unfixed finding; an over-declared deep level is NOT RUN; forged ledgers in free text are defanged; Deep is read from the header."
repository_or_source_context: "branch chore/review/register-detail; the verbatim last review comment of PR #168 is a test fixture"
files_or_sections_affected: [".claude/agents/{go,web,ci}-implementer.md, postgres-expert.md, skeptic.md (owner 2026-11-14: full Claude Code primitives - LSP, Monitor, TaskStop, TodoWrite, WebFetch, WebSearch, ToolSearch, NotebookEdit - and the rule to use them instead of Bash for reading, searching and editing)", ".claude/commands/review-fix.md (the pipeline tests are not fixer paths)", ".claude/commands/deep-review.md (the Deep range)", "scripts/review-comment.py", "scripts/review-automerge.py", "scripts/review-context.py", "scripts/findings.py", "scripts/review-autofix.py", "scripts/review-ping.py", "scripts/tests/ (findings, findings-batch, findings-settle, review-comment, review-comment-levels.py, review-autofix, review-automerge, review-context, review-ping, fixtures/review-register)", ".github/workflows/code-review.yml", ".github/pipeline-paths.txt", "docs/decisions/0013-default-branch-develop.md (schedule sentence and line references only)", ".claude/commands/delegate-review.md", ".claude/commands/deep-review.md", ".claude/commands/review-fix.md", ".github/workflows/review-batch.yml", ".github/workflows/review-autofix.yml", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency; secret shapes written from public token formats, no copied rule file"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #187)"]
tests_or_evidence: ["RED: 26 review-comment, 16 findings, 3 findings-batch, 3 findings-settle and 8 review-autofix assertions failed before the change", "GREEN: the eleven script tests, make lint-scripts, make docs-check and a YAML parse of both workflows", "follow-up round RED: level table 65 failures over 28 explicit literal rows, review-automerge 10, review-autofix 9, review-context 6, findings-settle 4, review-ping 3 assertions failed before the change; GREEN: every review and findings script test, review-agent-triggers, close-merged-issues, make lint-scripts, make docs-check and a YAML parse of code-review.yml"]
open_risks: ["the location check is deterministic text matching: an entry whose detail names no backticked identifier is only checked for file and line range", "a repository-code step in the fixer job can still rewrite finding-context.json, but the job then replaces it with an empty context (hash of the trusted step, held as a runner step output)", "an entry the model lists again with another file or line takes the head as first_sha (the permalink stays correct, the first raised date moves)", "the deep range wiring in code-review.yml is checked statically only: a real deep run on a pull request is NOT RUN", "scripts/tests/fixtures/** and close-merged-issues_test.sh stay writable by the fixer", "NOT A DEFECT and DECIDED sources stay open and can come back in the next batch until a human settles them", "an untrimmed sev (for example ' HIGH ') counts as MEDIUM, so an answered one stops blocking", "a cited file with a space or a bracket in its path is dropped (empty file)"]
```

## Review pipeline follow-ups (#212)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and the ci-implementer and skeptic subagents"
purpose: "The #212 follow-ups of the independent review of #187: the gate decides on the head's ledger, verified owner trust, a comment that never loses a review or a CRITICAL, a literal level table, small hardening and an HMAC-signed path back to no-click batch auto-merge."
repository_or_source_context: "branch fix/review/issue-212-follow-ups; issue #212; PR #187 (merged); plan reviewed by the owner"
files_or_sections_affected: ["scripts/review-gate.py", "scripts/review-automerge.py", "scripts/review-comment.py", "scripts/review-ping.py", "scripts/review-autofix.py", "scripts/findings.py", ".github/workflows/code-review.yml", ".github/workflows/review-autofix.yml", ".github/workflows/review-batch.yml", ".github/workflows/review-automerge.yml", "scripts/tests/review-*", "scripts/tests/findings*", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (pipeline owner, merge click)", "Y4NN777 (owner closure, 2026-11-15; merged in #223)"]
tests_or_evidence: ["RED then GREEN per group in review-gate_test, review-automerge_test, review-comment_test, review-ping_test, findings-batch_test, review-autofix_test; level-table mutation check (blocking() ignoring answered HIGH gives 10 failures)", "make lint-scripts"]
open_risks: ["note-settle-trusts-fixer-verdict not changed (owner decision)", "branch protection reading the latest Review gate run is INFERRED", "REVIEW_PIPELINE_HMAC_KEY is an owner secret; until it exists the batch exemption stays off"]
```

## Review batch 2026-11-14 fixes (issue #157)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent of the owner's session)"
purpose: "Triage of the 15 items of review batch #157: each source read against the code and docs/quality/REVIEW_CONTEXT.md, then fixed (RED test first), answered as not a defect or already decided with a file:line, or left for an owner decision. Fixes: a resource's related links showed private projects (#118, live since S5 writes resource -> project links), and the display-name minimum aligned to the contract (#146, owner decision). #156 is left to its fix by NICE-DEV226 (#160)."
repository_or_source_context: "branch fix/review/batch-2026-11-14; issues #157, #156, #118, #117 and the PR notes issues #135, #138, #140, #142, #143, #145, #146, #148"
files_or_sections_affected: ["db/queries/solve.sql", "db/queries/discover.sql", "src/api/internal/platform/dbgen/solve.sql.go (generated)", "src/api/internal/platform/dbgen/discover.sql.go (generated)", "src/api/internal/platform/pgtest/solve_integration_test.go", "src/api/internal/platform/pgtest/catalogue_read_integration_test.go", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #163)"]
tests_or_evidence: ["RED: TestSolveEditKeepsALinkToAHiddenTag failed on the old query (tags after the restore [go], want [go tag-cache])", "RED: TestCatalogueRelatedShowsPublicProjectsOnly failed on the old query (private project shown on the public resource)", "GREEN: both pass with -race -tags integration after the query changes"]
open_risks: ["the other batch items are decisions or owner questions, not code changes; humans tick or close each source"]
```

## AI review: triage the register in batches

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent)"
purpose: "Work the AI-review register (open finding issues, undecided notes) in ranked batches, one fix pull request per batch, instead of one by one."
repository_or_source_context: "branch chore/review/triage-batch; scripts/findings.py; docs/quality/REVIEW_CONTEXT.md"
files_or_sections_affected: ["scripts/findings.py", "scripts/tests/findings-batch_test.sh", ".github/workflows/review-batch.yml", "mk/60-review.mk", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #150)"]
tests_or_evidence: ["RED: scripts/tests/findings-batch_test.sh failed 16 of 20 before the subcommand existed", "GREEN: findings-batch_test.sh 26 passed; findings_test.sh 40 passed"]
open_risks: ["NOT RUN against the real GitHub API", "the scheduled trigger runs only from the default branch", "actionlint NOT RUN (not installed)"]
```

## AI review: auto-fix in CI and one fix pull request per batch

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent)"
purpose: "Run the review loop in CI with little owner time: Claude Code fixes open MEDIUM and HIGH findings of a pull request (stage 3) and opens one fix pull request per review batch (stage 5), as a normal project session, behind deterministic decisions and a push guard."
repository_or_source_context: "branch chore/review/ci-autofix-batch (stacked on fix/review/ledger-head-sha); anthropics/claude-code-action v1.0.275 inputs read at the pinned commit (setting sources default to user, project, local)"
files_or_sections_affected: [".claude/commands/review-fix.md", ".github/workflows/review-autofix.yml", ".github/workflows/review-batch.yml", "scripts/review-autofix.py", "scripts/tests/review-autofix_test.sh", ".github/workflows/review-merged.yml", ".claude/agents/ci-implementer.md", "scripts/findings.py", "scripts/tests/findings-settle_test.sh", "scripts/close-merged-issues.py", "scripts/tests/close-merged-issues_test.sh", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: ["https://github.com/anthropics/claude-code-action (action.yml and base-action/src/parse-sdk-options.ts at 4d75e5f390ee76b2e09c1980ecc99ff3a2e088c8, read only)"]
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #170)"]
tests_or_evidence: ["RED: scripts/tests/review-autofix_test.sh 40 failed, 3 passed before the script existed", "GREEN: review-autofix_test.sh 43 passed; after the security split (fixer-run wrapper, .npmrc and owner preferences guarded) RED 9 failed, then GREEN 66 passed", "format-patch, am and check round trip simulated locally: PASS", "settle: findings-settle_test.sh RED 13 failed, then GREEN 19 passed; findings-batch_test.sh 26 and findings_test.sh 40 still pass", "body with Batch and Sources lines: RED 2 failed, then GREEN 71 passed", "close-merged-issues_test.sh 14 passed (written before the script, RED run not recorded)", "security review fixes: RED 14 forbidden paths and the rename bypass passed check, 2 settle and 1 close-merged cases failed, the token message failed 2; GREEN review-autofix_test.sh 123, findings-settle_test.sh 21, close-merged-issues_test.sh 15, findings-batch 26, findings 40", "config restore from develop simulated locally: PASS", "PR 170 review fixes: RED 30 guard and prepare cases, 3 settle cases; GREEN review-autofix_test.sh 157, findings-settle_test.sh 24", "PR 170 second review: RED 13 deploy and generated paths, 2 prepare cases; GREEN review-autofix_test.sh 171", "YAML parse of the review workflows PASS"]
open_risks: ["NOT RUN against GitHub: workflow_run and schedule run only from the default branch copy", "actionlint NOT RUN (not installed)", "make runs pull request code while the Claude OAuth token is in the action step (same-repository branches only; CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1 is best effort, no bubblewrap)", "network isolation (unshare -rn) covers go test and pnpm only, and only where the runner allows unprivileged user namespaces", "make targets run pull request code during the model session and could write a Claude configuration after prepare; --setting-sources project limits what is read, hot reload of settings UNVERIFIED", "--settings as a claude_args flag for attribution: UNVERIFIED that the pinned action passes it through", "scripts/review-ping.py clears needs-owner when nothing CRITICAL or HIGH is open; the attempt cap rests on the autofix-2 label"]
```

## AI review: deep review (stage 2), auto-merge (stage 4), NOT RUN fallback

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent of the lead session)"
purpose: "Run the review and merge process in CI with little owner time. An empty or invalid review output now writes the marker comment with `Triage: NOT RUN | <reason>` and the previous ledger carried forward (NOT RUN is never a pass; the gate reads it as NOT RUN instead of an older verdict). Stage 2: a read-only adversarial deep review (Opus, .claude/commands/deep-review.md, skeptic mandate) runs after the review when a P0 path is touched and merges its findings into the same comment and ledger (source deep). Stage 4: scripts/review-automerge.py and .github/workflows/review-automerge.yml merge a pull request into develop with a merge commit only when every required check, the review, the deep review and the owner rules agree, then update the other open pull requests."
repository_or_source_context: "branch chore/review/ci-deep-review-merge, stacked on fix/review/ledger-head-sha (PR 159); branch protection of develop read through the API (required checks Validate branch promotion, quality, Review gate; strict false; conversation resolution required)"
files_or_sections_affected: ["scripts/review-comment.py", "scripts/review-gate.py", "scripts/review-automerge.py", ".github/pipeline-paths.txt", "scripts/review-strip-pr.sh", "scripts/review-context.py", ".claude/commands/delegate-review.md", "scripts/tests/review-strip-pr_test.sh", "scripts/tests/review-comment_test.sh", "scripts/tests/review-gate_test.sh", "scripts/tests/review-automerge_test.sh", ".github/workflows/code-review.yml", ".github/workflows/review-automerge.yml", ".github/deep-review-paths.txt", ".claude/commands/deep-review.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #168)"]
tests_or_evidence: ["RED: review-comment 16 of 28 fail, review-gate 1 of 49, review-automerge 34 of 38 (no script)", "GREEN: review-comment 30, review-gate 50, review-automerge 44 passed (after merging the PR 159 hardening)", "PR 168 review (4 MEDIUM): RED review-comment 4 of 39, review-automerge 12 of 50, review-strip-pr 1 of 4 failed; GREEN 39, 50 and 4 passed", "PR 168 second review (2 MEDIUM): RED review-comment 3 of 41, review-automerge 13 of 53; GREEN 41 and 53 passed", "one comment per run (owner decision): RED review-comment 11 of 46, review-context 1 of 73, review-gate 1 of 51; GREEN 46, 73, 51, review-automerge 54"]
open_risks: ["auto-merge runs from main (workflow_run uses the default branch): NOT RUN end to end before the promotion", "the deep job and the NOT RUN fallback run only once their scripts are on develop (they are read from the base commit)", "a CRITICAL replaced by a later NOT RUN no longer fails the gate (the gate fails open on NOT RUN); auto-merge still refuses", "update-branch after each merge costs one CI run and one review per open pull request"]
```

## Review batch #174 and the review pile fixed

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead with go-implementer, postgres-expert and web-implementer (Sonnet)"
purpose: "Fix or decide every item of review batch #174 and of the open notes issues outside it (owner directive: everything fixed or decided with a reason), including the owner-approved contract change viewer_vote and viewer_application."
repository_or_source_context: "branch fix/review/batch-2026-11-14-b; batch #174; issues #117, #118, #151, #155, #158, #162, #165, #167, #171, #175, #181, #182"
files_or_sections_affected: ["api/openapi.yaml and generated code", "db/queries/discover.sql, solve.sql, build.sql", "src/api/internal/{search,discover,solve,build,assistant,platform}/**", "src/web/src/features/{solve,build,search,learn,settings,discover}/**", "src/web/src/test/**", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency added"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner)", "Y4NN777 (owner closure, 2026-11-15; merged in #199)"]
tests_or_evidence: ["every FIXED item has a test that failed first (RED) then passed (GREEN); names in the pull request body", "make generate PASS (no drift)", "Go coverage 100.0% (6295/6295)", "make test-integration-go PASS", "web test 450 passed", "make lint-scripts: re-run after this tracker update"]
open_risks: ["the question ETag is caller-scoped: a question loaded signed out then edited after sign-in is refetched (sign-in clears the cache)", "a hidden tag slug answers 422, which reveals that a hidden tag exists under that slug"]
```

## Review fix for PR 199: snippet-nested-tag

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code in CI (review fixer), Claude Opus 5.5 lead with go-implementer (Sonnet)"
purpose: "Fix the AI-review finding snippet-nested-tag: one htmlTag pass in search Snippet let a nested tag rebuild a tag."
repository_or_source_context: "pull request 199, finding snippet-nested-tag"
files_or_sections_affected: ["src/api/internal/search/domain/domain.go", "src/api/internal/search/domain/domain_test.go", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency added"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner)", "Y4NN777 (owner closure, 2026-11-15; merged in #199)"]
tests_or_evidence: ["RED then GREEN: TestSnippetIsPlainTextOfAtMost300Runes/nested_tags_do_not_rebuild_a_tag and /a_tag_hidden_by_a_tag_does_not_survive", "go test ./src/api/internal/search/... PASS", "make lint: gofmt, go vet, staticcheck, golangci-lint, sqlc vet, OpenAPI, migration policy, ESLint and Prettier PASS; ShellCheck and later steps NOT RUN (no Docker daemon in the fixer sandbox)", "make coverage: every package at 100.0% (search/domain included) except src/api/cmd/api, which fails only because the runner's GITHUB_TOKEN env var is rejected by the config check (environment, unrelated to this change)"]
open_risks: []
```

## Review fix for PR 199: snippet-markdown-rebuilds-tag

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code in CI (review fixer), Claude Opus 5.5 lead with go-implementer (Sonnet)"
purpose: "Fix the AI-review finding snippet-markdown-rebuilds-tag: the markdown passes ran after the tag loop in search Snippet, so inline code or emphasis could rebuild a tag."
repository_or_source_context: "pull request 199, finding snippet-markdown-rebuilds-tag"
files_or_sections_affected: ["src/api/internal/search/domain/domain.go", "src/api/internal/search/domain/domain_test.go", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency added"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner)", "Y4NN777 (owner closure, 2026-11-15; merged in #199)"]
tests_or_evidence: ["RED then GREEN: TestSnippetIsPlainTextOfAtMost300Runes/inline_code_cannot_rebuild_a_tag, /inline_code_hiding_a_tag_does_not_survive, /emphasis_cannot_rebuild_a_tag, /bold_cannot_rebuild_a_tag, /strikethrough_cannot_rebuild_a_tag", "go test ./src/api/internal/search/... PASS", "make lint: gofmt, go vet, staticcheck, golangci-lint, sqlc vet, OpenAPI, migration policy, ESLint and Prettier PASS; ShellCheck and later steps NOT RUN (no Docker daemon in the fixer sandbox)", "make coverage: every package at 100.0% (search/domain included) except src/api/cmd/api, which fails only because the runner's GITHUB_TOKEN env var is rejected by the config check (environment, unrelated to this change)"]
open_risks: []
```

## AI review: ledger head, one comment, deterministic preview, P0 calibration

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (subagent of the lead session)"
purpose: "The review ledger recorded GitHub's temporary merge commit, so every push got a full re-review. The pack now prints the pull request head (PR_HEAD_SHA) as `Head commit:`, computes the range against it, and the ledger must record it; decide() returns incremental for any push with new commits, skip only without new commits, first without a usable ledger (with the reason in one line). Same pull request: the workflow runs the open-code-review preview itself and the comment carries a fixed `Tool:` line; one review comment per pull request, written by scripts/review-comment.py by marker (older bot Triage duplicates minimized), with the gate and the ledger reader preferring the marker comment, else the most recently updated one; an INFERRED concern with a plausible failing scenario on a P0 path is a MEDIUM finding. After an independent review of PR 159: the review step has no git or preview tool (git diff/show/log accept --output), the base scripts run from a copy under RUNNER_TEMP, open-code-review is installed with --ignore-scripts and empty npm config outside the checkout, the owner preferences come from the base branch, the ledger sha is pinned to the head, a merge of the base branch is left out of the range, and a missing review text gives a `Triage: NOT RUN` comment."
repository_or_source_context: "branch fix/review/ledger-head-sha; runs 43578279388 (#147) and 43581882270 (#153, 6 turns, 0 permission denials, preview not run); the ledgers of #147 and #149 (sha not in the repository); the duplicate Triage comments of #100 to #102; the pinned claude-code-action source (structured_output with --json-schema)"
files_or_sections_affected: ["scripts/review-context.py", "scripts/review-comment.py", "scripts/review-gate.py", "scripts/tests/review-context_test.sh", "scripts/tests/review-comment_test.sh", "scripts/tests/review-gate_test.sh", ".github/workflows/code-review.yml", ".github/sensitive-paths.txt", ".claude/commands/delegate-review.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #159)"]
tests_or_evidence: ["RED: 13 of 59 tests fail against the previous script", "GREEN: bash scripts/tests/review-context_test.sh exit 0 (59 passed)", "RED then GREEN: review-context 8 new failures, review-gate 3, review-comment (no script); GREEN 68, 48 and 13 passed", "PR 159 review fixes: RED 4 (context), 3 (comment), 1 (gate); GREEN 72, 16, 49"]
open_risks: ["each open pull request gets one more full review, since its old ledger points at a vanished merge commit", "this pull request's own review run uses develop's scripts, so it posts no comment (no review-comment.py on the base yet)", "structured output and the CI preview are NOT RUN in CI until merged"]
```

## S3 Solve PR 1: questions and answers

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer and skeptic subagents)"
purpose: "S3 PR 1: the six question and answer operations (listQuestions, createQuestion, getQuestion, updateQuestion, listAnswers, createAnswer) on the existing tables of migration 00005, with the platform packages (etag, idempotency, cursor, ratelimit, subject, apperr), an independent security review and its fixes, and three owner decisions (relevance order with q, ETag on editable fields, snippet and link limits recorded as deviations)."
repository_or_source_context: "branch feature/solve/questions-answers; issue #67; TODO S3 block; models src/api/internal/build and src/api/internal/learn"
files_or_sections_affected: ["src/api/internal/solve/**", "db/queries/solve.sql", "src/api/internal/platform/dbgen/solve.sql.go (generated)", "src/api/internal/platform/dbgen/solve_queries_test.go", "src/api/internal/platform/pgtest/solve_integration_test.go", "src/api/internal/app/** (wiring)", "scripts/unimplemented-routes.txt", "scripts/public-routes.txt", "scripts/authz-tests.txt", "scripts/tests/authz-routes_test.sh", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #147)"]
tests_or_evidence: ["RED: the ten TestSolve* integration tests answered 501 before the handlers; each review fix and owner decision had its own failing test first", "GREEN: make generate, test-unit, coverage (Go first-party 100.0%), test-integration-go and security exit 0"]
open_risks: ["the unit tests of the first commit were written after the implementation (integration tests were RED first)", "snippets and question-link limits are deviations until a migration or contract PR after the freeze", "the answer cursor assumes a stable vote_score until PR 2"]
```

## S3 Solve PR 2: votes and accepted answer

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer and skeptic subagents)"
purpose: "S3 PR 2: castVote, retractVote, acceptAnswer and unacceptAnswer on migration 00005, with the score kept in the same transaction as the vote (row lock, then one statement), one accepted answer under concurrency, an independent security review and its two test fixes, and two owner decisions (answer paging under moving votes accepted; the S3 path extension)."
repository_or_source_context: "branch feature/solve/votes-accept; issue #67; TODO S3 block"
files_or_sections_affected: ["src/api/internal/solve/**", "db/queries/solve.sql", "src/api/internal/platform/dbgen/solve.sql.go (generated)", "src/api/internal/platform/dbgen/*_test.go", "src/api/internal/platform/pgtest/solve_votes_integration_test.go", "src/api/internal/app/** (wiring)", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt", "scripts/tests/authz-routes_test.sh", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #153)"]
tests_or_evidence: ["RED first per layer (service, transport, repository, sqlc, app wiring); a mutation check removing the target lock made the duplicate-vote and reconciliation tests fail", "GREEN: concurrent accepts leave exactly one accepted answer; vote_score equals the sum of votes after a random burst; concurrent duplicate votes make one row; a removed question's answers are not votable (404)"]
open_risks: ["answer paging can repeat or skip an answer while votes move (accepted)", "a removal racing a vote is safe by the lock but has no test (no P0 path removes content yet)"]
```

## S5 Build PR 2: applications

```yaml
date: "2026-11-14"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "S5 Build, second PR: createApplication (verified email; need locked FOR SHARE; someone else's private project 404, own project and closed need 409 STATE_CONFLICT; INSERT ON CONFLICT DO NOTHING then the existing row locked: a withdrawn application is reopened, anything else 409 DUPLICATE_APPLICATION, so a double click never makes a second row), listProjectApplications (owner only, owner predicate inside the query, 403 public / 404 private with no count), decideApplication (decided on the application row locked with its project: neither applicant nor owner 404, wrong party 403, not pending 409, guarded UPDATE ... WHERE status = 'pending')."
repository_or_source_context: "branch feature/build/applications (stacked on feature/build/projects-needs)"
files_or_sections_affected: ["db/queries/build.sql", "src/api/internal/platform/dbgen/build.sql.go", "src/api/internal/platform/dbgen/build_applications_queries_test.go", "src/api/internal/build/**", "src/api/internal/app/server.go", "src/api/internal/app/build_internal_test.go", "src/api/internal/platform/pgtest/build_applications_integration_test.go", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt", "TODO.md", "docs/ai-usage/2026-11-14-s5-applications.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #132)"]
tests_or_evidence: ["RED: the four application pgtest tests answered 501 before the routes were wired", "GREEN: go test -race on internal/build/... and the dbgen contract test; pgtest on real PostgreSQL: lifecycle (apply, replay, duplicate 409, owner read, withdraw, re-apply reopens the row, decline, no re-apply after decline, no second decision, closed need), double click (one 201, one 409, one row), privacy and IDOR (403/404 without data, maintainer row grants nothing, third party and other project's owner 404), concurrent accept and withdraw (one 200, one 409)"]
open_risks: ["an empty message is 422 although the contract allows it (applications_message_ck needs 1 to 2000 characters)", "the decision note is not stored (no column)", "a reopened application keeps its first created_at, so it sorts by its first application"]
```

## S5 Build PR 1: projects and contribution needs

```yaml
date: "2026-11-14"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "S5 Build, first PR: listProjects (public plus the caller's own private projects, q/technology/skill/open_needs filters, keyset cursor bound to viewer and filters), createProject (verified email, GitHub URL normaliser and allowlist, duplicate URL 409 checked before any GitHub call, write:user and 10-per-hour import budgets failing closed, read-only GitHub adapter with a post-DNS dial check refusing non-public addresses and non-443 ports, no redirects, no proxy, timeouts and a 1 MiB body cap), getProject (private 404 to everyone but the owner), updateProject (owner only on the locked row: 403 public, 404 private; github_url immutable; technologies replaced, links added), listContributionNeeds (follows project visibility), createContributionNeed (owner only, verified email); C1 owner-only corrections in AUTHORIZATION_MATRIX.md and DATABASE_DESIGN.md."
repository_or_source_context: "branch feature/build/projects-needs"
files_or_sections_affected: ["db/queries/build.sql", "src/api/internal/platform/dbgen/build.sql.go", "src/api/internal/platform/dbgen/build_queries_test.go", "src/api/internal/build/**", "src/api/internal/app/server.go", "src/api/internal/app/app.go", "src/api/internal/app/build_internal_test.go", "src/api/internal/platform/pgtest/build_integration_test.go", "src/api/internal/platform/pgtest/build_stack_integration_test.go", "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt", "docs/spec-kit/AUTHORIZATION_MATRIX.md", "docs/spec-kit/DATABASE_DESIGN.md", "TODO.md", "docs/ai-usage/2026-11-14-s5-projects-needs.md"]
third_party_sources_consulted: ["GitHub REST API repository endpoint shape (name, full_name, description, private), from general knowledge; no code copied"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #131)"]
tests_or_evidence: ["RED: domain normaliser tests failed on behaviour against a stub; GitHub adapter tests failed against a stub; the Build pgtest tests answered 501 before the handler was wired", "GREEN: go test -race on internal/build/... and the dbgen contract test", "pgtest on real PostgreSQL: registration, idempotent replay, duplicate 409 without GitHub call, 422/503 mapping, link rollback, concurrent registration (one 201, one 409, one row), visibility matrix, maintainer row grants nothing, needs, filters, 10/hour import limit, pagination, EXPLAIN plans", "make generate, lint-go, lint-contract, security-authz, security-sast, docs-check, coverage (Go first-party 100.0%)"]
open_risks: ["GitHub is called without a token: 60 requests per hour per server IP, shared by every user; a token is an owner secret and decision", "repository renames are refused (redirects are not followed); the owner re-submits the new URL", "a private project's GitHub URL still blocks the same URL for others (existence oracle accepted at P0, DATABASE_DESIGN section 10)", "the composition root, pgtest, dbgen tests and route lists are outside the S5 allowed paths as written (same as S4; owner agreement asked on the PR)"]
```

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
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #161)"]
tests_or_evidence: ["RED: the wiring test and every review fix (17 redaction cases, nested delimiters, links in output, malformed markers, Retry-After)", "GREEN: make generate, test-unit, coverage (Go 100.0%), test-integration-go, security exit 0"]
open_risks: ["real provider not run (V2, V4)", "unit tests of the first commit written after the code", "no UI panel yet, so the 'sent to Google' notice does not exist"]
```

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
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (SPEC and matrix, H13)", "Y4NN777 (owner closure, 2026-11-15; merged in #176)"]
tests_or_evidence: ["RED: 6 tests failed before the code (recherche short query; solve notice, request, no_source, 503, English)", "GREEN: pnpm lint, format:check, typecheck, coverage exit 0 (345 tests; 96.72/90.29/96.79/97.78)"]
open_risks: ["SPEC owner items: the disabled and limiter-failure 503 send no Retry-After although TODO S9 says 30; AI-3 and AI-5 are enforced by the prompt only; published_at never filled; series in the source enum never retrieved; ADR-0012 still PROPOSED"]
```

## SEED original demo fixtures and make seed

```yaml
date: "2026-11-14"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer and postgres subagents)"
purpose: "SEED #75: the fixture contract (db/seeds/README.md), the local-dev-only fail-closed make seed target, the five deterministic idempotent seed files (identity, learn series of 4 tutorials, discover, build, optional solve) and their tests."
repository_or_source_context: "branch feature/platform/demo-seeds; TODO.md SEED block; issue #75; db/migrations 00001 to 00013"
files_or_sections_affected: ["db/seeds/**", "mk/50-ops.mk (seed target)", "scripts/tests/seed_test.sh", "scripts/tests/make_test.sh", "src/api/internal/platform/pgtest/seed_integration_test.go", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original fixture text and code, fictional names, example.org URLs, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #215)"]
tests_or_evidence: ["RED: seed_integration_test.go fails on a scratch seed with a live-provenance user and a vote (TestSeedProvenanceIsDemo, TestSeedHasNoFakeActivity)", "GREEN: go test -race -tags integration -run Seed ./src/api/internal/platform/pgtest/ (7 tests PASS)", "scripts/tests/seed_test.sh 10 passed", "APP_ENV=local-dev make seed: 107 rows, second run 0 rows; APP_ENV=staging-test make seed refused with exit 2"]
open_risks: ["make quality not run end to end locally: make doctor refuses pnpm 9.15.9 (pin 10.32.1)", "the zero-demo-rows check on staging-test and production-live is G7 evidence, after deployment", "path extension beyond db/seeds and mk/50-ops.mk awaits owner approval"]
```

## parametres: the display name is saved on the account

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "First #109 follow-up, owner decision: parametres kept the display name only in browser storage and never called the API. The form now loads it with getMe and saves it with updateMyProfile through the CSRF-aware helper, shows server violations on the field, and the old storage key is removed."
repository_or_source_context: "branch feature/identity/web-settings-profile; issue #109; a note from the independent review of PR #100"
files_or_sections_affected: ["src/web/src/features/settings/**", "src/web/src/app/storage.ts", "src/web/src/main.tsx", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #144)"]
tests_or_evidence: ["RED: 7 tests failed first (load from the API, sign-in invite, PATCH with CSRF and JSON body, header refresh with no storage copy, server violation on the field, legacy key removal, storage unavailable)", "GREEN: pnpm lint, format:check, typecheck, coverage exit 0 (97.63/91.59/98.94/98.83)"]
open_risks: ["the client keeps a 2-character minimum while the contract allows 1", "the start-up cleanup call in main.tsx has no test (the function it calls is tested)"]
```

## Sign-in account limit under concurrency (issue #97)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer subagent)"
purpose: "Issue #97: the per-account sign-in limit was check-then-act, so concurrent wrong passwords all reached the password check. The owner chose to count the attempt up front with one atomic hit and give it back on a successful sign-in."
repository_or_source_context: "branch fix/identity/signin-limit-race; issue #97; ADR-0003 rate limits"
files_or_sections_affected: ["src/api/internal/identity/**", "db/queries/identity.sql", "src/api/internal/platform/dbgen/identity.sql.go (generated)", "src/api/internal/platform/dbgen/queries_test.go", "src/api/internal/app/server_identity_test.go", "src/api/internal/platform/pgtest/signin_limit_integration_test.go", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #139)"]
tests_or_evidence: ["RED: TestIdentitySignInAccountLimitHoldsUnderConcurrency failed on the old code (20 password checks, want 5)", "GREEN: the same test (5 INVALID_CREDENTIALS, 15 RATE_LIMITED) and TestIdentitySignInSuccessGivesTheHitBack pass with -race"]
open_risks: ["more than 5 simultaneous correct sign-ins to one account can briefly get 429", "a give-back skipped under a concurrent bucket change keeps the count strict, never lower"]
```

## S3 follow-up: hidden tag links survive question edits (issue #156)

```yaml
date: "2026-11-14"
contributor: "NICE-DEV226"
tool_or_model: "Devin (SWE-2 Flash)"
purpose: "Fix the defect found in review of #147 and registered as issue #156: DeleteSolveQuestionTags deleted every question_tags row, so editing a question's tags silently dropped links to tags a moderator had set hidden (invisible on reads, impossible for the author to re-send). The delete now joins tags and removes only status='active' links, mirroring DeleteProjectTags (#131) and the #116 user_skills fix (#130)."
repository_or_source_context: "branch fix/solve/hidden-tag-links from develop 22a337b; issue #156; db/queries/solve.sql; pgtest solve_integration_test.go"
files_or_sections_affected: ["db/queries/solve.sql", "src/api/internal/platform/dbgen/solve.sql.go (regenerated, not hand-edited)", "src/api/internal/platform/pgtest/solve_integration_test.go", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #160)"]
tests_or_evidence: ["RED: TestSolveEditKeepsAHiddenTagLink FAILS on the pre-fix query (0 link rows left)", "GREEN: same test PASS with the scoped delete; TestSolveEditKeepsAHiddenLinkHidden still PASS; go test -race -tags integration on real PostgreSQL"]
open_risks: ["InsertSolveQuestionTags still has no status filter: a client that guesses a hidden slug can re-link it; reads filter it anyway, same exposure as build's InsertProjectTags — accepted, noted in issue #156"]
```

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
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #169)"]
tests_or_evidence: ["RED: 10 AssistantPanel tests failed against the demo panel", "GREEN: pnpm lint, format:check, typecheck, coverage exit 0 (338 tests; 96.76/90.2/96.77/97.77)"]
open_risks: ["the solve ask-flow panel is not built", "the real provider stays NOT RUN (V2, V4); the assistant is off until AI_ENABLED is set"]
```

## S7 build, project and profile pages

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer and skeptic subagents)"
purpose: "The build, project and profile pages on the merged S5 and S1 APIs: project list and registration, project detail with contribution needs and applications (owner decisions, applicant withdrawal), and the public profile with the completion checklist on the member's own profile only."
repository_or_source_context: "branch feature/build/web-build-profile-pages; issue #109 section 5; TODO S1, S5 choices and deviations, S7 page map; prototypes/html/pages/{build,project,profile}.html"
files_or_sections_affected: ["src/web/src/features/build/**", "src/web/src/features/profile/**", "src/web/src/app/router.tsx", "src/web/src/app/navigation.ts", "src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #177)"]
tests_or_evidence: ["RED: BuildPage 13/13, ProjectPage 21/21, ProfilePage 15/15 failing before the routes existed", "GREEN: lint, format:check, typecheck, coverage exit 0 (424 tests; 96.07/88.82/95.45/97.06)", "skeptic review: no security finding; test gaps F1-F3 and the application replay guard F4 fixed in 61ade18"]
open_risks: ["updateProject has no If-Match in the contract, so concurrent owner edits are last-write-wins", "an applicant can withdraw only right after applying: no operation lists a member's own applications", "axe, visual and responsive checks NOT RUN"]
```

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
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #213)"]
tests_or_evidence: ["RED: connect, Shell and Home tests failing before the route existed", "GREEN: lint, format:check, typecheck, coverage exit 0 (476 tests; 95.97/89.2/95.63/97.01)", "#214 follow-up: githubProfileUrl helper, RED then GREEN in src/features/connect and src/features/profile"]
open_risks: ["not exercised against the real backend", "axe, visual and responsive checks NOT RUN"]
```

## S7 ecrire page (issue #124)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "The ecrire page on createLearnArticle under the three #124 constraints: question, resource and project links only (pre-filled from the route), articles created as published, and a body-omitted idempotent replay that follows Location."
repository_or_source_context: "branch feature/learn/web-ecrire; issue #124; TODO S4 known deviations; prototypes/html/pages/ecrire.html"
files_or_sections_affected: ["src/web/src/features/learn/**", "src/web/src/app/router.tsx", "src/web/src/app/Shell.tsx", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #154)"]
tests_or_evidence: ["RED: 23 tests failing before the route existed", "GREEN: lint, format:check, typecheck, coverage exit 0 (329 tests; 96.75/90.26/96.86/97.79)"]
open_risks: ["the 14 topic labels are a client constant copied from migrations 00002 and 00011 (no endpoint)", "no article page exists yet, so success links to the home page"]
```

## S7 learn pages: apprendre, article, formation

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "The apprendre, article and formation pages on the merged S4 API (listLearnArticles with the topic filter, listTrainingSeries, getLearnArticle, getTrainingSeries), the « Poser une question sur cette leçon » action to /solve?link=article:<id>, and the ecrire success link to the new article page."
repository_or_source_context: "branch feature/learn/web-learn-pages; issue #109 section 5; TODO S4 known deviations and S7 page map; prototypes/html/pages/{apprendre,article,formation}.html"
files_or_sections_affected: ["src/web/src/features/learn/**", "src/web/src/app/router.tsx", "src/web/src/app/navigation.ts", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #173)"]
tests_or_evidence: ["RED: vitest src/features/learn 23 of 46 failing before the pages existed", "GREEN: lint, format:check, typecheck, coverage exit 0 (362 tests; 96.74/90.52/96.47/97.79)"]
open_risks: ["/build/<id> and /profils/<username> links resolve once the build and profile pages merge", "axe, visual and responsive checks NOT RUN"]
```

## S7 live home page, resource error report and E4 completion

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 lead + web-implementer (Sonnet)"
purpose: "The accueil sections on live data (listLearnArticles, listQuestions, listResources, listProjects) with the fixtures removed, the « Signaler une erreur » form on the resource page (createResourceErrorReport), the ask-about-the-lesson button on the series page, the read-only questions-by-topic-tag feed on apprendre, and the E4 tests (pre-fill, link sent by createQuestion, anonymous visitor)."
repository_or_source_context: "branch feature/web/live-home-and-e4; TODO S7 page map and E4; api/openapi.yaml read-only"
files_or_sections_affected: ["src/web/src/features/home/**", "src/web/src/features/discover/**", "src/web/src/features/learn/**", "src/web/src/features/solve/AskAboutLesson.test.tsx", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #196)"]
tests_or_evidence: ["RED before each item: failing Vitest tests (home 5, error report 7, series 1, topic feed 3); the E4 tests are characterisation tests of merged behaviour and passed at first run", "GREEN, pnpm --filter web run typecheck, lint, format:check: exit 0 each", "GREEN, pnpm --filter web run test: exit 0, 44 files, 451 tests", "pnpm --filter web run coverage: exit 1, NOT a PASS: two runs on an overloaded host (load average above 50 from other sessions) timed out 53 to 54 tests in unrelated files while the same 451 tests passed in the test run; no coverage figure was produced, re-run on a quiet host"]
open_risks: ["error report: the contract has no duplicate code for createResourceErrorReport, so there is no duplicate message", "axe, visual and responsive checks NOT RUN"]
```

## S7 solve and question pages

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "The solve and question pages on the live S3 PR 1 operations: list with search and tag filter, ask form (Idempotency-Key per submission, one snippet, pre-filled link only), question with answers (accepted first), answer form, owner edit with If-Match and 412 reload; vote and accept controls hidden for the session on 501 until S3 PR 2. Also the Vitest timeout raised to 15 s (owner decision)."
repository_or_source_context: "branch feature/solve/web-solve-question; TODO S7 page map and S3 deviations; prototypes/html/pages/solve.html and question.html for layout and copy"
files_or_sections_affected: ["src/web/src/features/solve/**", "src/web/src/app/router.tsx", "src/web/src/app/navigation.ts", "src/web/src/app/Shell.test.tsx", "src/web/src/i18n/messages.ts", "src/web/vite.config.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #149)"]
tests_or_evidence: ["RED: 30 failing tests before the routes existed", "GREEN: pnpm lint, format:check, typecheck, coverage exit 0 (301 tests; 97.17/90.92/96.95/98.07)"]
open_risks: ["the home feed's question items are demo fixtures and do not link to the question page yet", "the edit form covers title, body and tags only"]
```

## Web UI audit follow-ups (#192)

```yaml
date: "2026-11-14"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents), Playwright for the screenshots"
purpose: "The contributor asked to finish the follow-ups of the 14 Nov UI audit listed in issue #192. The AI wrote the fixes and tests for items 5, 7, 9, 12, 13, 14, 15 and 16 and the dev-server sign-in 403 (item 6 was already fixed by #196): live landing pillars link to their pages and share one card style; Résoudre's search row and tag placeholder; « Réponses : 0 » dropped; profile details before the checklist, full-width checklist fields, a visible progress track on the muted card (theme override of progress-bar-track, found by reading the Astryx ProgressBar styles); a one-line empty state on a bare public profile; visible labels on the theme and language switches; the password show toggle laid over the end of the field so it is as wide as the e-mail field; the Vite proxy keeps the browser Host for the API's same-origin check."
repository_or_source_context: "branch fix/platform/web-ui-audit-followups; S7 web pages; issue #192"
files_or_sections_affected: ["src/web/src/features/identity/PasswordField.tsx", "src/web/src/features/landing/LandingPage.tsx", "src/web/src/features/landing/LandingPage.test.tsx", "src/web/src/features/landing/messages.ts", "src/web/src/features/profile/Checklist.tsx", "src/web/src/features/profile/ProfilePage.tsx", "src/web/src/features/profile/ProfilePage.test.tsx", "src/web/src/features/profile/messages.ts", "src/web/src/features/settings/SettingsPage.tsx", "src/web/src/features/settings/SettingsPage.test.tsx", "src/web/src/features/solve/SolvePage.tsx", "src/web/src/features/solve/SolvePage.test.tsx", "src/web/src/features/solve/messages.ts", "src/web/src/styles/index.css", "src/web/src/theme/theme226os.ts", "src/web/vite.config.ts", "TODO.md", "docs/ai-usage/2026-11-14-web-ui-audit-followups.md"]
third_party_sources_consulted: ["@astryxdesign/core 0.7.14 TextInput, IconButton, ProgressBar and defineTheme sources and types (node_modules), read for the missing end slot, the track colour and the component override syntax"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #209)"]
tests_or_evidence: ["Order: the fixes were written first, the tests after", "new tests: profile empty-state line (shown to a visitor, hidden when something is filled, replaced by the checklist on the own profile, details before checklist); « Réponses : 0 » hidden while « Réponses : 3 » shows; visible Thème and Langue labels", "changed test: LandingPage.test.tsx expects one « Bientôt » instead of four and links for Apprendre, Résoudre and Construire", "existing sign-in, sign-up and reset tests still drive the show-password toggle", "excalibur run: verified (lint, prettier, tsc, vitest); feature tests 356 passed", "browser: sign-in at 1280 and 390 px with and without errors (both fields 408 px, toggle clear of the error icon), landing pillars, Résoudre at 1280 and 390 px, Réglages, a bare public profile, a 0 % ProgressBar on a muted card before and after"]
open_risks: ["the member's own checklist was not looked at signed in (no test account used); the 0 % track was checked on an isolated ProgressBar", "the password toggle position is measured in a layout effect and has no jsdom assertion (layout only)", "the dev-server sign-in 403 fix was not re-run in a browser this session; the real local-dev stack behind Traefik is not checked", "the browser checks ran on the Vite dev server with English pages (French copy is covered by the tests only), not on the built image"]
```

## Web UI audit and its safe fixes

```yaml
date: "2026-11-14"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Fable 5.2 (main session, no subagents), Playwright for the screenshots"
purpose: "The contributor asked for an inspection of the web UI (wrong parts, incoherences, details to improve), then chose which findings to fix. The AI captured 20 pages in French at 1280 px and 390 px, signed out and signed in, on a throwaway local stack (temporary PostgreSQL, the API and the Vite dev server, a fictional account), listed 16 findings, and wrote the fixes and tests for the 7 the contributor selected (1, 2, 3, 4, 8, 10, 11). The 9 others are in issue #192. Fixes: phone navigation back in the Astryx drawer (Shell passed a SideNav as mobileNav.content, so it was laid out under the content and the toggle did nothing); not-found page framed in a main landmark; recherche button no longer truncated; landing search field no wider than a phone; sign-in prompt on the question page primary like the others; article body 720 px at most. Finding 10 (uneven phone padding) needed no change of its own: it disappears with the navigation fix."
repository_or_source_context: "branch fix/platform/web-ui-audit; S7 web pages; follow-up issue #192"
files_or_sections_affected: ["src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "src/web/src/app/App.test.tsx", "src/web/src/features/NotFound.tsx", "src/web/src/features/search/SearchPage.tsx", "src/web/src/features/landing/LandingPage.tsx", "src/web/src/features/solve/QuestionPage.tsx", "src/web/src/features/learn/ArticlePage.tsx", "src/web/src/styles/index.css", "TODO.md", "docs/ai-usage/2026-11-14-web-ui-audit.md"]
third_party_sources_consulted: ["@astryxdesign/core 0.7.14 AppShell source (node_modules), read to find how mobileNav is rendered"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #194)"]
tests_or_evidence: ["Order: the fixes were written first, the tests after; the two new tests were then run against the previous Shell.tsx and NotFound.tsx and both failed (RED), and pass with the fixes (GREEN)", "Shell.test.tsx: 'keeps the navigation in a drawer on a phone, opened by the toggle'; App.test.tsx: the not-found page sits in a main landmark", "vitest 432 passed (42 files); eslint, prettier --check and tsc --noEmit PASS", "before and after screenshots read at 390 px and 1280 px"]
open_risks: ["the search button, the landing field width, the primary sign-in prompt and the article width have no automated test (layout only, checked on screenshots)", "the open drawer's content is not asserted in jsdom (no modal dialog there); it was checked in the browser", "checked on the Vite dev server, not on the built image behind nginx and Traefik; English pages not re-captured after the fixes"]
```

## Accessibility pass fixes (#109 §6)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked to fix the three defects found by the manual keyboard and accessibility-tree pass for issue #109 §6. The AI wrote RED tests first, then the fixes: useHashToken follows hashchange so a mail link opened in a tab already on the page verifies the token (VerifyEmailPage, once per token) or pre-fills the code (ResetPasswordPage); the language items of the account menu and the anonymous options menu are menuitemradio with the current language checked (Astryx DropdownMenuRadioGroup); the mobile navigation drawer is an Astryx MobileNav with aria-modal true. It also added two TODO.md lines (S5 updateProject without If-Match, S7 summary of the pass)."
repository_or_source_context: "branch fix/web/a11y-pass-109; S7 web pages; issue #109"
files_or_sections_affected: ["src/web/src/features/identity/useHashToken.ts", "src/web/src/features/identity/VerifyEmailPage.tsx", "src/web/src/features/identity/ResetPasswordPage.tsx", "src/web/src/features/identity/VerifyEmailPage.test.tsx", "src/web/src/features/identity/ResetPasswordPage.test.tsx", "src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-a11y-pass-109.md"]
third_party_sources_consulted: ["@astryxdesign/core installed type declarations and sources (DropdownMenu radio group and item, MobileNav, SideNav render context, AppShell mobileNav config), read from node_modules"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #233)"]
tests_or_evidence: ["Order: RED tests first (failing for the missing behaviour), then the fixes", "new tests: verify and reset pages react to a fragment set in a tab already open; language menuitemradio checked in both menus; drawer dialog Navigation has aria-modal true", "changed test: the English language item is now a menuitemradio", "pnpm --filter web typecheck, lint, format:check and test all exit 0 (510 tests)"]
open_risks: ["the drawer was not re-checked in a real browser: jsdom has no showModal, so the test stands one in; focus trap and Escape rely on Astryx MobileNav, unchanged", "the drawer is now built from our own MobileNav content instead of the AppShell auto drawer; its links reuse the SideNav drawer-content mode", "no real screen reader run"]
```

## ADR-0015 draft: auto-fix by proof

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) with a research subagent"
purpose: "After the auto-fix failures seen on #252, a research subagent surveyed deployed auto-fix systems and the program-repair literature; the lead drafted ADR-0015 from that report and the owner's answers (a plausible fix waits for a human review; prose findings go to the batch)."
repository_or_source_context: "scripts/review-autofix.py, review-comment.py, findings.py, ADR-0014, the #252 run history"
files_or_sections_affected: ["docs/decisions/0015-autofix-by-proof.md", "docs/decisions/README.md", "docs/ai-usage/2026-11-15-adr-0015-autofix.md"]
third_party_sources_consulted: ["papers and documentation listed in the ADR's Sources section"]
license_review: "NOT APPLICABLE: original text, no code or dependency"
closure: "open"
human_reviewers: ["Y4NN777 (pending)"]
tests_or_evidence: ["make docs-check"]
open_risks: ["some figures come from secondary summaries, said so in the ADR", "status PROPOSED until the owner settles the two open points"]
```

## ADR-0015 phase 0: attempts per finding and head, path prefilter, locked tests, status comment, batch trigger

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent) and Claude Opus 5.5 (lead session, skeptic review subagent)"
purpose: "The AI implemented phase 0 of ADR-0015 in the review pipeline: auto-fix attempts counted per (finding, head SHA) in one bot status comment, written by `review-autofix.py record` when the run ends, with a cap on aborted runs; a path prefilter that sends findings outside the fixer allowlist to the owner; existing tests, fixtures, test infrastructure and test configs locked at the trusted push, new test files under a lexical scan; a preflight job that refuses a run without REVIEW_AGENT_TOKEN before the model; NO-OP / NOT RUN wording; the batch also runs after Review merged."
repository_or_source_context: "branch feature/review/adr-0015-phase-0, from develop 84922c2"
files_or_sections_affected: ["scripts/review-autofix.py", ".github/workflows/review-autofix.yml", ".github/workflows/review-batch.yml", "scripts/tests/review-autofix_test.sh", "scripts/tests/review-autofix-record_test.sh", "scripts/tests/review-batch-trigger_test.sh", "docs/decisions/0015-autofix-by-proof.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["the 17 review-* and findings* test files pass, RED first for each new case", "scripts/lint-shell.sh OK", "check-docs OK", "check-todo-updated OK", "three skeptic review rounds with a bypass probe; the last bypasses (TestMain, test configs, test infrastructure) fixed and tested"]
open_risks: ["nothing ran on GitHub: the record job, the preflight job, the status comment edit and the workflow_run conclusion filter are tested with a fake gh only", "the scan of NEW test files is lexical and best effort: a weak assertion in a new test is not detected", "the preflight job is a fourth reader of REVIEW_AGENT_TOKEN (ADR-0013 says three)", "actionlint NOT RUN"]
```

## ADR-0015 phase 1: classes, reproducers, static oracles, routing by class

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (review-steward subagent and lead session)"
purpose: "The AI implemented units 1 and 2 of ADR-0015 phase 1 in the review pipeline: a `class` (executable, static, prose) and a `rule` on each ledger entry, validated by `review-comment.py` per owner decision D1 (a downgrade only, missing is prose, carried forward with the re-list); a prose finding never blocks the merge unless it is CRITICAL; an answered or accepted prose HIGH is filed for the batch by `findings.py harvest`; a classless entry and a prose HIGH on a P0 path keep blocking (owner answers); the reproducer job, the read-only reproducer check, the reproducer hash rule and the static oracle before and after the fix (units 3 to 5); the status comment rows with the class, the repro state and the reproducer hash prefix, and the routing by class in `decide_pr` (units 6 and 7: prose to the batch, a classless entry keeps the phase 0 lane)."
repository_or_source_context: "branch feature/review/adr-0015-phase-1, from develop 213bf3d"
files_or_sections_affected: ["scripts/review-autofix.py", ".github/workflows/review-autofix.yml", ".claude/commands/review-repro.md", "scripts/tests/review-repro_test.sh", "scripts/tests/review-autofix_test.sh", "scripts/tests/review-automerge_test.sh", "scripts/review-comment.py", "scripts/review-context.py", "scripts/findings.py", ".github/workflows/code-review.yml", ".claude/commands/delegate-review.md", ".claude/commands/deep-review.md", "scripts/tests/review-comment_test.sh", "scripts/tests/review-comment-levels.py", "scripts/tests/findings_test.sh", "docs/decisions/0015-autofix-by-proof.md", "docs/quality/REVIEW_CONTEXT.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["the review-* and findings* test files pass, RED first for each new case", "routing by class and the status proof columns tested in review-autofix_test.sh, with a scratch probe of malformed classes and forged repro facts", "the level table and the fixtures that relied on a classless finding blocking now name a blocking class, the classless case is tested as prose"]
open_risks: ["nothing ran on GitHub", "under D1 every finding outside src/ and db/queries/ is prose and blocks nothing unless CRITICAL", "eslint and tsc rule ids are checked by shape only", "the repro-check and oracle jobs run pull request code, so their results could be forged by that code (accepted by the owner, answer (f))", "the reproducer, oracle and publish wiring is tested statically only", "actionlint NOT RUN", "a prose HIGH on a P0 path keeps blocking the merge but is never given to the fixer: it waits for the batch or the owner"]
```

## Connect directory: public GitHub developers from committers.top (ADR-0016)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (implementer subagent of the lead session, acting as go-implementer, postgres-expert and web-implementer)"
purpose: "The owner approved Part 2 of the Connect plan: the directory also lists public GitHub developers ranked by committers.top as labelled external entries, with member opt-in unchanged. The AI wrote ADR-0016, migration 00014 (github_developers, github_developer_optouts, profiles_github_lower_idx), the sqlc queries (db/queries/connect.sql), the identity/committers package (fixed-host fetch client with the Build SSRF dial guard, a YAML parser with field validation, a daily job that skips an unchanged snapshot and keeps the previous rows on any error), the CONNECT_GITHUB_IMPORT setting, the contract change (GET /profiles source and external, PublicProfile.ranking), the service, repository and transport code, and the connect and profile web changes, each test first."
repository_or_source_context: "branch feature/connect/github-developers from origin/develop 213bf3d; E5 Connect; plan Part 2"
files_or_sections_affected: ["docs/decisions/0016-connect-github-developers.md", "docs/decisions/README.md", "db/migrations/00014_github_developers.sql", "db/queries/connect.sql", "src/api/internal/identity/committers/**", "src/api/internal/identity/domain/githubdev.go", "src/api/internal/identity/domain/profile.go", "src/api/internal/identity/repository/githubdev.go", "src/api/internal/identity/service/directory.go", "src/api/internal/identity/service/profile.go", "src/api/internal/identity/service/service.go", "src/api/internal/identity/transport/directory.go", "src/api/internal/identity/transport/profile.go", "src/api/internal/platform/config/connect.go", "src/api/internal/platform/config/config.go", "src/api/internal/app/app.go", "src/api/internal/platform/pgtest/githubdev_*_integration_test.go", "src/api/internal/platform/pgtest/migrate_integration_test.go", "src/api/internal/platform/pgtest/errorreport_integration_test.go", "api/openapi.yaml and generated outputs", "src/web/src/features/connect/**", "src/web/src/features/profile/**", "docs/spec-kit/AUTHORIZATION_MATRIX.md", "scripts/public-routes.txt", ".env.example", "TODO.md"]
third_party_sources_consulted: ["committers.top location file format (ashkulz/committers.top, gh-pages, _data/locations/<slug>.yml, MIT), through the parser of PR #326", "go.yaml.in/yaml/v3 already in go.mod (now used by first-party runtime code)"]
license_review: "NOT APPLICABLE for code: original code, no new dependency. Data: committers.top is MIT; only public GitHub fields are stored (ADR-0016 point 7)."
closure: "open"
human_reviewers: []
tests_or_evidence: ["Order: RED test first for each deliverable, then the code (logs kept in the session)", "Go unit: parser with \\u escapes, field validation, top 500, skip unchanged, keep on failure, SSRF dial guard, redirect and size cap, ticker start, stop and cancel; service merge, source filter, cursor across segments and forged cursors, ranking visibility", "pgtest: constraints of 00014, replace in one transaction, keyset across countries, dedup against listable members with a visible github per audience, opt-out, members page byte-identical for source=members, ranking only with a visible github, EXPLAIN plans bounded by Limit on 5000 rows", "web: connect external rows, labels, footer, source filter, avatar host check, profile rank (FR and EN)"]
open_risks: ["the opt-out channel is a new issue on the public repository until the owner names another", "the web change was not compared with a prototype capture in a real browser", "the first import runs a minute after start and calls raw.githubusercontent.com from staging and production (10 small GETs a day)"]
```

## Local-dev Connect seed from committers.top

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "The AI wrote a stdlib generator that turns committers.top country files (the `users` list) into a local-dev seed of real public GitHub developers for the Connect directory, the make seed change that loads db/seeds/local-dev/ for APP_ENV=local-dev only, and the tests."
repository_or_source_context: "branch feature/connect/seed-committers, from develop 213bf3d"
files_or_sections_affected: ["scripts/content/committers-seed.py", "db/seeds/local-dev/00001_connect_committers.sql", "mk/50-ops.mk", "scripts/tests/seed_test.sh", "scripts/tests/committers-seed_test.sh", "scripts/tests/fixtures/committers-bf.yml", "src/api/internal/platform/pgtest/local_dev_seed_integration_test.go", "db/seeds/README.md", "TODO.md"]
third_party_sources_consulted: ["committers.top data (github.com/ashkulz/committers.top, MIT, gh-pages, _data/locations/*.yml)"]
license_review: "data from an MIT repository; real public GitHub profile data, local-dev only by owner decision; no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["seed_test.sh 20/20 (RED first for the local-dev directory)", "committers-seed_test.sh 17/17 (RED first)", "TestLocalDevSeed* pass against PostgreSQL; the taken-name test fails when the skip clause is removed", "existing TestSeed* unchanged and passing"]
open_risks: ["real people's names in a local-dev database: never load db/seeds/local-dev/ elsewhere", "the data ages: regenerate with the generator", "make seed end to end against the local-dev stack NOT RUN in this session"]
```

## Shorter UI copy, part 2 (shared, solve, build, discover, connect, profile)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner confirmed the style guide and its decisions. The AI applied the rewritten strings of the shared i18n catalogue and of the solve, build, discover, connect and profile catalogues (FR and EN), deleted the explanatory ledes, not-found bodies, Connect spaces teasers and dead keys with their renders, moved « réglages » to « paramètres », and updated the expected strings of the tests."
repository_or_source_context: "branch feature/web/copy-rewrite-2, cut from feature/web/prototype-fidelity; S7 web pages; owner copy brief of 2026-11-15"
files_or_sections_affected: ["src/web/src/i18n/messages.ts", "src/web/src/features/solve/", "src/web/src/features/build/messages.ts", "src/web/src/features/discover/", "src/web/src/features/connect/", "src/web/src/features/profile/messages.ts", "src/web/src/features/settings/", "src/web/src/features/NotFound.tsx", "TODO.md", "docs/ai-usage/2026-11-15-copy-rewrite-2.md"]
third_party_sources_consulted: ["none beyond the style guide approved by the owner"]
license_review: "NOT APPLICABLE: original text, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["only expected strings changed in tests; the one check of the deleted sign-in note on the question page was removed, not replaced", "scoped queries where two controls now share a name: the Enregistrer buttons of the settings page (by section) and the Effacer les filtres buttons of the build empty state", "pnpm typecheck, lint, format:check and test --maxWorkers=1 results are in the PR"]
open_risks: ["copy is not reviewed by a native French editor", "the deleted not-found bodies leave the title and the back link alone on those pages", "shell.search is « Rechercher sur le site » / « Search the site », not the table value, so that the top bar, the home feed search and the learn search are search landmarks with different names", "an accepted answer now carries the pill « Résolue » (solve.pill.resolved), the same string as a resolved question"]
```

## Shorter UI copy, part 1 (landing, home, learn, search, identity)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked for shorter, less AI-sounding UI copy and confirmed a style guide. The AI applied the rewritten strings of the landing, home, learn, search and identity catalogues (FR and EN), deleted the explanatory ledes and notes with their renders (HomePage section sub, landing pillar and catalogue ledes, learn and write ledes, invitation body, not-found bodies, sign-in and sign-up card ledes, welcome intro), shortened the learn empty states added by the filters pass, and updated the expected strings of the tests."
repository_or_source_context: "branch feature/web/copy-rewrite, cut from feature/web/prototype-fidelity; S7 web pages; owner copy brief of 2026-11-15"
files_or_sections_affected: ["src/web/src/features/landing/", "src/web/src/features/home/", "src/web/src/features/learn/", "src/web/src/features/search/", "src/web/src/features/identity/", "src/web/src/features/solve/SolvePage.test.tsx", "src/web/src/features/solve/AskAboutLesson.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-copy-rewrite.md"]
third_party_sources_consulted: ["none beyond the style guide approved by the owner"]
license_review: "NOT APPLICABLE: original text, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["only expected strings changed in tests; the two checks of the deleted sign-in and sign-up ledes were removed, not replaced", "LearnPage 'offers writing to a member': the header action and the invitation now share the label « Write an article », so the single findByRole became findAllByRole with length 2 and an href check on each (stricter)", "solve tests SolvePage and AskAboutLesson assert strings of the search and learn catalogues and were updated", "pnpm typecheck, lint and format:check exit 0; pnpm test --maxWorkers=1 leaves only the known Shell.test.tsx banner failure (see the PR)"]
open_risks: ["copy is not reviewed by a native French editor", "catalogues outside these five (src/i18n/messages.ts and others) are part 2", "the assistant notice no longer says that only the question is sent", "learn empty-state body « Retirez un filtre. » sits above the « Effacer les filtres » button"]
```

## All 54 countries and optional fields on the project creation form

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, web-implementer subagent (Sonnet 5.5)"
purpose: "The Build country filter and the project forms offered 16 West African codes plus the codes already seen, while the database accepts the 54 African countries. One shared list (West Africa first, then the rest, names from the platform) now serves Build, Connect and Discover; the creation form gets the optional stage, licence and country, reusing the owner edit fields, with 422 on /stage, /license_spdx and /country_code shown on the fields."
repository_or_source_context: "branch feature/web/countries-create, PR #246"
files_or_sections_affected: ["src/web/src/components/countries.ts", "src/web/src/features/build/ProjectFields.tsx", "src/web/src/features/build/ProjectForm.tsx", "src/web/src/features/build/ProjectPage.tsx", "src/web/src/features/build/BuildPage.tsx", "src/web/src/features/connect/ConnectPage.tsx", "src/web/src/features/discover/DiscoverPage.tsx"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency added"
closure: "open"
human_reviewers: []
tests_or_evidence: ["vitest run --maxWorkers=1 on src/features/build, connect, discover and src/components/countries.test.ts: 13 files, 177 tests PASS; typecheck, lint, format:check PASS"]
open_risks: ["the 54 codes are copied from db/migrations/00002_reference_taxonomy.sql, not read from the API"]
```

## Data gaps wiring: show the data PR #238 serves (S7)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, web-implementer subagent of the lead session"
purpose: "The merged API (PR #238) now serves project stage, licence and country, a required total on every page, sort and filter parameters, and username filters. The AI wired them into the web pages the way the prototypes show them: Build rows and project page (stage, country, licence, stage and country filters, sort select, owner edit of stage, licence and country with 422 mapped to fields), Solve and Discover (sort select, result count from the total), Learn (count from the totals, no more « N premiers »), and the profile activity tabs (Questions, Réponses, Projets). Tests were written first (RED) for every visible element and URL parameter. Search keeps its per-group counts: /search serves no total."
repository_or_source_context: "branch feature/web/data-gaps-wiring (stacked on feature/web/prototype-fidelity, PR #246); S7 web; prototypes/html as visual reference only (DISCLAIMER.md)"
files_or_sections_affected:
  [
    "src/web/src/components/SortSelect.tsx",
    "src/web/src/app/router.tsx",
    "src/web/src/features/build/**",
    "src/web/src/features/solve/**",
    "src/web/src/features/discover/**",
    "src/web/src/features/learn/**",
    "src/web/src/features/profile/**",
    "TODO.md",
    "docs/ai-usage/2026-11-15-data-gaps-wiring.md",
  ]
third_party_sources_consulted:
  [
    "@astryxdesign/core installed sources and docs (Selector, TabList, Tab) read from node_modules",
  ]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence:
  [
    "RED first: 43 new tests failed before the code (route search validation, row meta, counts from the total, sort and filter parameters in the URL and the request, restart at the first page, owner edit with null clearing and 422 mapping, profile tabs)",
    "changed tests: the Discover and Learn count tests now expect the API total instead of « Les N premiers »; page fixtures carry a total",
    "targeted vitest runs and pnpm typecheck, lint, format:check: see the PR",
  ]
open_risks:
  [
    "the country list of the filters and of the owner form is West Africa plus the codes already seen, not every country the API knows",
    "the project creation form does not yet offer stage, licence or country",
    "no captures: the shared local-dev API was built before PR #238 and was not restarted",
    "axe and screen reader NOT RUN on the new controls and tabs",
  ]
```

## Deploy settings in versioned files, E2E account in the sops file

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent, brief from the owner's lead session)"
purpose: "The AI moved the non-secret deploy settings (DEPLOY_HOST, DEPLOY_USER, PUBLIC_HOST, DEPLOY_KNOWN_HOSTS) from GitHub variables and secrets to deploy/staging.conf and deploy/production.conf, with a strict parser (scripts/deploy-conf.sh), and the E2E account from GitHub secrets to secrets/staging-test.enc.env (scripts/e2e-env.sh). It updated deploy.yml, deploy-env.sh, the bootstrap message, the tests and the deploy runbook."
repository_or_source_context: "branch feature/ops/deploy-conf-file, from develop 3fe6d93"
files_or_sections_affected: [".github/workflows/deploy.yml", "deploy/staging.conf", "deploy/production.conf", "scripts/deploy-conf.sh", "scripts/e2e-env.sh", "scripts/deploy-env.sh", "scripts/server-bootstrap.sh", "scripts/tests/deploy-conf_test.sh", "scripts/tests/e2e-env_test.sh", "scripts/tests/deploy-env_test.sh", "docs/runbooks/deploy-setup.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["scripts/tests/deploy-conf_test.sh, e2e-env_test.sh, deploy-env_test.sh, e2e-ci-config_test.sh pass; RED before the scripts existed"]
open_risks: ["not run in GitHub Actions: the sops decrypt of the E2E keys and GITHUB_ENV loading are untested end to end", "the E2E keys are not yet in secrets/staging-test.enc.env, so the e2e job fails until they are added"]
```

## Emergency production deploy without the staging gate

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (systalink session); design agreed with the lead session (Claude Opus 5.5)"
purpose: "The AI wrote a separate emergency workflow that deploys production without the staging tested-<tree> gate, on the owner's decision: manual dispatch on main with a reason and the confirmation EMERGENCY (validated by scripts/deploy-plan.sh), own build and scan, images tagged emergency-<sha>-<run id>-<attempt> and deployed by the pushed digests behind the production environment, never verified, attested or tagged tested-*, concurrency shared with deploy.yml."
repository_or_source_context: "branch feature/deploy/emergency-path, from develop after #303"
files_or_sections_affected: [".github/workflows/deploy-emergency.yml", ".github/workflows/deploy-production.yml", "scripts/deploy-plan.sh", "scripts/tests/deploy-plan_test.sh", "scripts/tests/deploy-emergency_test.sh", "scripts/tests/deploy-images_test.sh", "docs/runbooks/deploy-setup.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["deploy-plan_test 19/19 and deploy-emergency_test 16/16, RED first", "deploy-images_test 18/18 with the two emergency-tag cases", "all deploy-* and e2e-* test files pass", "scripts/lint-shell.sh OK", "scripts/check-docs.py PASS"]
open_risks: ["nothing ran on GitHub: the first real emergency run is the first proof", "no E2E before an emergency deploy (accepted by the owner), only the smoke test after it", "the shared deploy-production.yml (workflow_call) is used by the emergency workflow only; deploy.yml keeps its own production steps until it is migrated", "actionlint NOT RUN"]
```

## Deploy pipeline findings #267 and #280: attested tested digests, E2E in its own job

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent) and Claude Opus 5.5 (lead session, skeptic review subagent)"
purpose: "The AI fixed two review findings of the deploy pipeline. #267: the build outputs the digests it pushed and refuses a different pre-existing <sha> tag; staging deploys those digests; promote attests them (actions/attest, predicate tree and sha) before tagging tested-<tree>; production verifies the attestation, its signer workflow and the tree (scripts/deploy-verify-tested.sh). #280: the E2E suite moved to its own job and environment staging-e2e (E2E_USER and E2E_PASSWORD only), and the junit report is checked for the account before upload (scripts/e2e-junit-check.sh)."
repository_or_source_context: "branch fix/ops/deploy-pipeline-findings, from develop 628532b"
files_or_sections_affected: [".github/workflows/deploy.yml", "scripts/deploy-images.sh", "scripts/deploy-verify-tested.sh", "scripts/e2e-env.sh", "scripts/e2e-junit-check.sh", "scripts/tests/deploy-images_test.sh", "scripts/tests/deploy-verify-tested_test.sh", "scripts/tests/deploy-conf_test.sh", "scripts/tests/e2e-ci-config_test.sh", "scripts/tests/e2e-env_test.sh", "scripts/tests/e2e-junit-check_test.sh", "docs/runbooks/deploy-setup.md", "docs/SHIP.md", "TODO.md"]
third_party_sources_consulted: ["actions/attest action.yml at 1e69f48 (v4.2.2)", "gh attestation verify manual"]
license_review: "no new dependency; actions/attest is MIT, pinned by SHA"
closure: "open"
human_reviewers: []
tests_or_evidence: ["the 10 deploy-* and e2e-* test files pass, RED first for each new case", "scripts/lint-shell.sh OK", "scripts/check-todo-updated.sh OK", "two independent skeptic reviews; the second: none blocking"]
open_risks: ["nothing ran on GitHub: the real push, attest and gh attestation verify path waits for a staging server", "the --source-ref refs/heads/develop value for a workflow_run signer is assumed, not observed", "a re-run of the build for an already pushed SHA fails closed (push a new commit)", "actionlint NOT RUN"]
```

## Deployment pipeline (staging and production)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, CI implementer subagent (Sonnet), brief by the owner's lead session"
purpose: "Wrote the deployment to two fresh servers: the server bootstrap script, the deploy workflow (staging on a green develop push, production on main by the digests tested on staging), the helper scripts and their offline tests, the owner setup runbook and the SHIP and release documents."
repository_or_source_context: "branch feature/ops/deploy-pipeline, from develop f939030"
files_or_sections_affected: [".github/workflows/deploy.yml", "scripts/server-bootstrap.sh", "scripts/deploy-plan.sh", "scripts/deploy-images.sh", "scripts/deploy-env.sh", "scripts/deploy-remote.sh", "scripts/tests/", "docs/SHIP.md", "docs/RELEASE_DRAFT.md", "docs/runbooks/deploy-setup.md", "docs/GIT_WORKFLOW.md", "mk/50-ops.mk"]
third_party_sources_consulted: ["sops v3.13.3 release checksums (pinned binary)"]
license_review: "NOT APPLICABLE: no dependency added to the project; the CI downloads the pinned sops binary (MPL-2.0) at run time"
closure: "open"
human_reviewers: []
tests_or_evidence: ["scripts/tests/{server-bootstrap,deploy-plan,deploy-images,deploy-env,deploy-remote}_test.sh"]
open_risks: ["nothing was run against a server or GitHub: first real run is the test", "the workflow was not checked with actionlint", "GHCR package access for the job token is unverified", "ADR-0009 point 8 needs an amendment"]
```

## Docs refresh, shorter README, demo seeds on staging-test

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session) with project subagents: lead (read-only audit, then doc edits), ci-implementer (make seed), go-implementer (directory filter)"
purpose: "The owner asked for a shorter, more professional README and for the documentation to be brought up to date. A read-only audit listed the stale facts; the owner then decided the updates. README.md was rewritten (standalone, French); every other doc got fact updates only (merged PRs up to #229, slice statuses, live GitHub settings read with gh, broken anchor, ADR range). Owner decision: demo seeds now run in local-dev and staging-test (reverses correction B, within ADR-0005 point 17), never in production-live; make seed accepts staging-test; the Connect directory hides provenance='demo' accounts outside local-dev. ADR-0014 records the review rule the pipeline already applies (auto-merge into develop; the owner steps in on CRITICAL, open HIGH, sensitive and pipeline paths), written from the review scripts' docstrings, and AGENTS.md, DEV.md and GIT_WORKFLOW.md point to it."
repository_or_source_context: "branch chore/ops/staging-seeds-and-docs-refresh from develop c9e4fe7"
files_or_sections_affected: ["README.md", "ROADMAP.md", "TODO.md", "CONTRIBUTING.md", "DESIGN.md", "docs/**", "prototypes/html/DISCLAIMER.md", "db/seeds/**", "mk/50-ops.mk", "scripts/tests/seed_test.sh", "db/queries/identity.sql", "src/api/internal/platform/dbgen/**", "src/api/internal/identity/**", "src/api/internal/app/**", "src/api/internal/platform/pgtest/directory_integration_test.go", "docs/decisions/0014-ai-review-merges-develop.md", "AGENTS.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original text and code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #237 and its follow-up #241)"]
tests_or_evidence: ["scripts/tests/seed_test.sh: RED 15/18 before the mk change, GREEN 18/18 after", "go test identity, app, dbgen: PASS; 100.0% statements on the touched packages (agent run); make lint-go PASS; sqlc vet PASS; sqlc regenerate shows no drift", "TestDirectoryHidesDemoAccountsByDefault on real PostgreSQL: FAIL with the filter removed (demo-person listed), PASS with it; go test -race -tags integration -run TestDirectory PASS", "TestSeedReusesLiveTagsBySlug: FAIL before (tags_slug_uq 23505 on a live go tag), PASS after the tag links resolve by slug; all TestSeed* and seed_test.sh PASS"]
open_risks: ["full make test-integration not run by hand (memory tight); the commit gate runs the configured checks", "C12 is left for the owner"]
```

## Playwright end-to-end suite for the deployed staging site

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent, brief from the owner's lead session)"
purpose: "The AI wrote a Playwright suite (src/web/e2e, src/web/playwright.config.ts) that runs against a deployed site: anonymous flows on the five pillars, search and FR/EN switch, mobile navigation, a signed-in question and answer flow driven by E2E_USER and E2E_PASSWORD, and a CSRF and Set-Cookie smoke. It also implemented the make test-e2e recipe, updated scripts/tests/make_test.sh for it, and documented the suite in docs/TEST.md and docs/DEV.md. The owner approved adding @playwright/test as a dev dependency."
repository_or_source_context: "branch feature/web/e2e-staging, from develop f939030"
files_or_sections_affected: ["src/web/e2e/", "src/web/playwright.config.ts", "src/web/package.json", "src/web/tsconfig.json", "src/web/eslint.config.js", "src/web/.prettierignore", "pnpm-lock.yaml", ".gitignore", "mk/40-quality.mk", "scripts/tests/make_test.sh", "docs/TEST.md", "docs/DEV.md", "THIRD_PARTY_NOTICES.md"]
third_party_sources_consulted: ["Playwright documentation conventions (getByRole, projects, trace on first retry)"]
license_review: "@playwright/test 1.64.1 is Apache-2.0 (LICENSE and package.json of the installed package); it is on the allowed list and is a dev dependency, not shipped"
closure: "open"
human_reviewers: []
tests_or_evidence: ["suite run against the local-dev stack (older develop build): see the PR for the per-test result", "make test-e2e without E2E_BASE_URL exits non-zero with NOT RUN; scripts/tests/make_test.sh 51 passed"]
open_risks: ["not yet run against staging", "the mobile bottom-bar test fails on a build without the prototype shell", "the signed-in flow needs a verified staging account in E2E_USER and E2E_PASSWORD"]
```

## Build pages follow their prototypes (fidelity pass)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked every page to match its HTML prototype visually. The AI rebuilt /build and /build/<id> after prototypes/html/pages/build.html and project.html, on Astryx components and Graphite & Lime tokens, using only contract fields: dense project rows (monogram, title link, two-line excerpt, technology tags with +n, meta line with owner, update date, private badge and origin question, open-needs state pill), the q and open_needs filters with a filtered empty state, local breadcrumbs, a project header with monogram and pills, related-content cards, the description with its language badge, needs as cards with state pills and one sign-in hint for visitors, owner applications with status pills and icon decisions, and side cards (about, technologies, maintainer). Behaviour, routes, viewer_application, idempotency and CSRF are unchanged."
repository_or_source_context: "branch feature/web/fidelity-build; S7 web pages; owner fidelity brief of 2026-11-15"
files_or_sections_affected: ["src/web/src/features/build/BuildPage.tsx", "src/web/src/features/build/ProjectPage.tsx", "src/web/src/features/build/Needs.tsx", "src/web/src/features/build/Applications.tsx", "src/web/src/features/build/present.tsx", "src/web/src/features/build/queries.ts", "src/web/src/features/build/messages.ts", "src/web/src/features/build/build.css", "src/web/src/features/build/BuildPage.test.tsx", "src/web/src/features/build/ProjectPage.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-build.md"]
third_party_sources_consulted: ["@astryxdesign/core installed type declarations (List, Badge, Icon registry, Breadcrumbs, TextInput, CheckboxInput, Link, Heading, Avatar), read from node_modules", "prototypes/html (pages, CSS, JS, CONVENTIONS.md, DISCLAIMER.md)"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: RED tests first (5 failing for the missing rows, filters, header, related cards and sign-in hint), then the pages", "new tests: row pill, private badge, update date, origin question link and content language; English pill; q and open_needs sent and reset; project header pill, private badge and jump link; description language; about, technologies and maintainer cards with country name; related kind outside the link name; no repository link without a GitHub address; one sign-in hint for visitors; breadcrumbs", "no existing test was changed", "captures of prototype and app at 1280x800 and 375x812, light and dark, kept in the lead session scratchpad"]
open_risks: ["filters are local state, not in the URL: /build has no validateSearch and router.tsx is outside the area", "the panels are white surfaces meant for the new shell that puts content on the page background; on the old white content panel they do not stand out", "owner and member views were checked by tests only (no seeded credentials for a browser capture)", "local breadcrumbs, pills, rows and monograms duplicate what the shared src/ui pieces will provide"]
```

## Page fidelity pass: Connect, profile and settings

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus (web-implementer subagent of the lead session)"
purpose: "The owner asked that every page match its HTML prototype. The AI rebuilt the look of /connect (prototype filter bar, dense people rows with avatar, country code and name, languages, skills and GitHub; the spaces block kept and labelled « Bientôt » once, per the DISCLAIMER), /profils/<username> (profile hero with avatar, handle, icon meta line, bio, skills, collaboration pill and GitHub; own profile: edit link, privacy note and checklist meter with a done/total count) and /parametres (section nav and one panel per section: Profil, Préférences, Confidentialité). Tests first for each new visible element; behaviour, routes, visibility rules and the checklist forms are unchanged."
repository_or_source_context: "branch feature/web/fidelity-connect; S7 web pages; prototypes/html/pages/{connect,profile,parametres}.html"
files_or_sections_affected: ["src/web/src/features/connect/ConnectPage.tsx", "src/web/src/features/connect/PersonRow.tsx", "src/web/src/features/connect/messages.ts", "src/web/src/features/connect/ConnectPage.test.tsx", "src/web/src/features/profile/ProfilePage.tsx", "src/web/src/features/profile/Checklist.tsx", "src/web/src/features/profile/icons.tsx", "src/web/src/features/profile/messages.ts", "src/web/src/features/profile/ProfilePage.test.tsx", "src/web/src/features/settings/SettingsPage.tsx", "src/web/src/features/settings/ProfileVisibility.tsx", "src/web/src/features/settings/settings.css", "src/web/src/features/settings/SettingsPage.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-connect.md"]
third_party_sources_consulted: ["@astryxdesign/core installed docs and sources (List, Item, Stack, StackItem, Card, Badge, Avatar, Icon, ProgressBar, Button), read from node_modules", "icon paths copied from the repository's own prototypes/html/assets/css/icons.css"]
license_review: "NOT APPLICABLE: original code and in-repository icon paths, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: RED tests first for the Connect rows and spaces block, the profile hero (meta line, edit link, privacy note, English) and the settings sections (nav anchors, named regions, English), then the code", "pnpm --filter web typecheck, lint and format:check exit 0; the full suite (vitest run in src/web, the script behind pnpm test) exits 0 with 519 tests, while the pnpm wrapper itself was stopped by SIGTERM on the shared machine before any test ran", "captures in the lead's scratchpad: prototype light at 1280 and 375; app signed out, light at 1280 and 375, dark at 1280"]
open_risks: ["signed-in views (own profile with checklist, settings forms) were checked by tests only, not captured in a browser", "settings strings live in profile/messages.ts under profile.settings.* because settings has no catalogue of its own and src/i18n/messages.ts was out of scope", "the settings nav has no current-section highlight (all sections are on the page; no scroll spy)", "page title stays « Réglages » (prototype « Paramètres »)"]
```

## Control fidelity: heights, search fields, phone rows, breadcrumbs and recherche match the prototypes

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "An independent screen judge listed the five largest visual gaps between the app and the HTML prototypes. The AI fixed them: Astryx control sizes set once in the theme tokens (44 px default, 36 small, 52 large, as the prototype), search icon inside the field on solve and discover (no separate button), small dark field labels, the language tag kept in the top-right corner of solve rows on a phone, the four solve states fitting a 375 px strip, a back link to the parent in place of the wrapping breadcrumb on a phone, tinted square vote buttons, the recherche start view with topic chips and a placeholder, and the signed-out sign-in row aligned to the start edge on a phone. RED tests first for the behaviour changes."
repository_or_source_context: "branch feature/web/fidelity-controls; S7 web pages; prototypes/html/assets/css/{tokens,components,layout}.css and the pages solve, question, recherche, connexion, discover, build, apprendre"
files_or_sections_affected: ["src/web/src/theme/theme226os.ts", "src/web/src/styles/ui.css", "src/web/src/styles/shell.css", "src/web/src/components/Breadcrumbs.tsx", "src/web/src/components/List.tsx", "src/web/src/features/solve/SolvePage.tsx", "src/web/src/features/solve/QuestionRow.tsx", "src/web/src/features/solve/Interactions.tsx", "src/web/src/features/solve/solve.css", "src/web/src/features/discover/DiscoverPage.tsx", "src/web/src/features/learn/LearnPage.tsx", "src/web/src/features/search/SearchPage.tsx", "src/web/src/features/search/search.css", "src/web/src/features/search/messages.ts", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-controls.md"]
third_party_sources_consulted: ["@astryxdesign/core installed sources (SizeContext, theme tokens, TextInput, Button, Icon registry), read from node_modules"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED first: 5 new tests failed (phone back link, corner tag class, recherche topics, solve and discover without a search button) before the code", "changed tests: the solve and discover search tests no longer click a search button (it is gone, Enter submits); the theme test gained the three control sizes", "captures of prototype and app at 1280 and 375 px in the session scratchpad (compare-after)", "pnpm typecheck, lint, format:check and test --maxWorkers=1 in src/web"]
open_risks: ["the control sizes change every Astryx button, input, selector and tab, so pages not captured (settings, learn forms, connect) are unchecked", "the solve search form has a hidden submit button so Enter works from the tag field; there is no visible submit", "topic chips are fixed demo tag slugs, not served data"]
```

## Discover and resource pages follow their prototypes

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner found the built Discover pages far from the HTML prototypes and asked for visual fidelity. The AI rebuilt the markup of /discover and /discover/<id> inside src/web/src/features/discover only: compact list rows (monogram tile, title link in an h3, provider, category label and type, one-line summary, country codes with hidden names, source count, status pill with icon and verified date), the 18 resource category labels from migrations 00002 and 00011 with a category filter, a results count, an empty state with a reset action; on the listing page a header (breadcrumb, monogram, kind and category, provider, status pill, verified and updated dates, docs button, « Signaler une erreur » icon that opens the report panel), then coverage, limitations, sources and related tiles. Styling is a feature CSS file on Astryx tokens. RED tests first, then the code."
repository_or_source_context: "branch feature/web/fidelity-discover; S7 web pages; prototypes/html/pages/discover.html and resource.html"
files_or_sections_affected: ["src/web/src/features/discover/DiscoverPage.tsx", "src/web/src/features/discover/ResourcePage.tsx", "src/web/src/features/discover/ResourceRow.tsx", "src/web/src/features/discover/ResourceThumb.tsx", "src/web/src/features/discover/CountryCodes.tsx", "src/web/src/features/discover/VerificationBadge.tsx", "src/web/src/features/discover/ErrorReport.tsx", "src/web/src/features/discover/categories.ts", "src/web/src/features/discover/messages.ts", "src/web/src/features/discover/discover.css", "src/web/src/features/discover/DiscoverPage.test.tsx", "src/web/src/features/discover/ResourcePage.test.tsx", "src/web/src/features/discover/ResourceThumb.test.ts", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-discover.md"]
third_party_sources_consulted: ["@astryxdesign/core 0.7.14 installed docs and sources (Badge, Icon, IconButton, Button, Selector, SegmentedControl, TextInput, Heading, EmptyState, VisuallyHidden, List and Item), read from node_modules", "prototypes/html (pages, CSS, page scripts, CONVENTIONS.md, DISCLAIMER.md)"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: 11 failing tests in src/features/discover before the code, then all green", "second round after review, RED 3 then green: the catalogue keeps h1 > hidden h2 « Résultats » > h3 rows; the report thanks survive closing and reopening the panel; the country select says « Choisir un pays » in the West Africa view and « Tous les pays » for the continent", "guard test (green before and after): the empty state offers the filter reset", "new tests: row heading, provider line, category label or raw code, codes with hidden names, count wording (singular, plural, first n while pages remain), category filter sent to the API, empty-state reset, header kind and category, status pill, verified and updated dates, docs and report actions, related tiles, monogram initials", "changed tests: provider line without « Par », « 2 sources » instead of « Sources : 2 », coverage shows each country code and name instead of one joined string (prototype markup)", "captures of prototype and app at 1280 and 375 px, light and dark, in the session scratchpad"]
open_risks: ["panels use the surface colour; on the current white main area they read as plain sections until the shell gives main the body background", "the shell may also render breadcrumbs: the listing page keeps its own until the lead swaps in the shared ones", "no real screen reader run; the row title link covers the row through a CSS overlay (checked in the browser: a click point in the summary and in the status column both resolve to the title link)", "the feature CSS is loaded before astryx.css, so it only targets its own discover- classes"]
```

## The S7 pages share one header, one scale and the shared pieces

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked that the app pages look like one product and close the remaining gaps with the HTML prototypes. Seven earlier passes had built each page group with local copies of the same pieces. The AI extended the shared components (PageHeader as a header element with tile, eyebrow, subtitle, meta and title language; Section; LinkedItems; LangCode; Avatar monogram tiles; StatusBadge kinds open, closed, claimed, disputed, pending, private; CountryCode with a visible name; TagList with a +n tail; ListRow subtitle and title language; people and solve glyphs in components/icons) and moved discover, solve, build, connect, profile and parametres onto them, with breadcrumbs on every page and one scale in styles/ui.css (h1, lead, section h2, rows, pills, panels). The local copies and their CSS were deleted. RED tests first for the shared pieces, then the pages."
repository_or_source_context: "branch feature/web/fidelity-harmonize from the integration branch at 341d497; S7 web pages; prototypes/html (tokens.css, base.css, layout.css, components.css, pages)"
files_or_sections_affected: ["src/web/src/components/PageHeader.tsx", "src/web/src/components/Section.tsx", "src/web/src/components/LinkedItems.tsx", "src/web/src/components/LangCode.tsx", "src/web/src/components/Avatar.tsx", "src/web/src/components/StatusBadge.tsx", "src/web/src/components/CountryCode.tsx", "src/web/src/components/Tag.tsx", "src/web/src/components/List.tsx", "src/web/src/components/icons.tsx", "src/web/src/components/ui.test.tsx", "src/web/src/styles/ui.css", "src/web/src/i18n/messages.ts", "src/web/src/features/discover/**", "src/web/src/features/solve/**", "src/web/src/features/build/**", "src/web/src/features/connect/**", "src/web/src/features/profile/**", "src/web/src/features/settings/SettingsPage.tsx", "src/web/src/features/settings/SettingsPage.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-harmonize.md"]
third_party_sources_consulted: ["@astryxdesign/core 0.7.14 installed types and sources (Heading, Avatar sizes and initials, Icon colours, Badge), read from node_modules", "prototypes/html (CSS tokens and components, page markup, App.statePill and App.thumb in assets/js/app.js)"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: ui.test.tsx failed on the missing Section, LinkedItems and LangCode modules, then green; the discover breadcrumb test failed against the 341d497 page (no « Fil d’Ariane » navigation), then green", "new tests: header element with tile, eyebrow, subtitle, meta and title language; Section region, aside, lede and title note; LinkedItems names; LangCode hint; tile initials and tone (moved from ResourceThumb.test.ts); six new pill kinds; CountryCode visible name read once; TagList +n; ListRow subtitle and language; breadcrumbs (Accueil and the section) on every page group", "changed test: the question page asserted a back link named « Questions »; it now asserts the breadcrumb trail (Accueil, Résoudre to /solve, the question as current page), since the trail replaces the link", "fix: the solve.state labels the apprendre list uses were restored, the merged branch did not typecheck without them", "gates at 178ed9c: typecheck exit 0, lint exit 0, format:check exit 0, full Vitest suite with one worker 47 files and 602 tests passed, exit 0", "after the captures: row titles bold and meta links at the meta size (Astryx link inner span), h3 at 16px inside sections, build keeps its visible « Projets » heading as build.html does", "captures of app and prototype at 1280 and 375, light, from the worktree's Vite server on the local-dev API, plus a computed-style table (h1 32px at 1280 and 26px at 375, lead 16px, section h2 20px, rows 12px 16px, pills 24px high at 12px, panels 12px radius on the surface, no horizontal scroll) in the session scratchpad"]
open_risks: ["solve keeps its en-GB date format against dateStyle medium on the other pages in English", "the settings section nav and the solve ask view's back button stay page-local by design", "no real screen reader run; row links in meta lines and asides were hit-tested in the browser above the row overlay, GitHub links on Connect were not present in the demo data", "the learn, home, landing, search and identity pages were out of scope and still use their own pieces"]
```

## Page fidelity pass: landing, member home, search and account pages

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked that every page match its HTML prototype visually. For `/` (landing and member home), `/recherche` and the account pages the AI rebuilt the layout after prototypes/html (index, accueil, recherche, connexion, inscription, mot-de-passe-oublie, reinitialiser, verifier-email, bienvenue): dense list rows with lead, title link, meta line and state pills, section heads with icon and see-all link, hero, icon pillar cards and catalogue cards, search groups, the account card head with title and lead, the forgot-password link on the password label line and welcome choice tiles. Only contract fields are shown; behaviour, routes and tests are kept. It wrote RED tests for the new visible elements first, then the code, and added one TODO.md line."
repository_or_source_context: "branch feature/web/fidelity-home-auth; S7 web pages; prototypes/html (read only)"
files_or_sections_affected: ["src/web/src/features/home/**", "src/web/src/features/landing/**", "src/web/src/features/search/SearchPage.tsx", "src/web/src/features/search/SearchPage.test.tsx", "src/web/src/features/search/messages.ts", "src/web/src/features/search/search.css", "src/web/src/features/identity/**", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-home-auth.md"]
third_party_sources_consulted: ["@astryxdesign/core installed type declarations and astryx.css token names, read from node_modules", "prototypes/html icon paths (the project's own prototype, copied into FeedIcon.tsx)"]
license_review: "NOT APPLICABLE: original code and the project's own prototype assets, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: RED tests first (each failed on the previous page), then the pages", "new tests: home rows (pills, counts, see-all links, never Vérifié when unverified), landing catalogue cards, search title and complete-group count, sign-in lead and forgot link placement, sign-up lead, welcome tile lines", "changed tests: sign-in h1 is now « Content de vous revoir » / « Welcome back » (prototype copy; the document title stays « Connexion »)", "captures of prototype and app at 1280 and 375 px, light and dark, in the session scratchpad"]
open_risks: ["the member home was captured with a stubbed session (dev-only Vite middleware, not committed), not a real sign-in", "the logo above the account card and its legal footer are in app/AuthLayout.tsx, outside this branch: the lead applies them", "the local rows, pills and icons duplicate what the shell branch adds in ui/ and components/: the lead swaps them", "no axe or screen-reader run on the new markup"]
```

## Page fidelity: the apprendre filter area

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked to close every remaining gap with the HTML prototypes. For /apprendre the AI wrote RED tests first, then added the prototype filter area inside features/learn: type chips (Astryx SegmentedControl), a text search, the Sujet, Niveau and Langue selects, a results line, one list (series rows, then article rows) and a reset, all kept in the URL. Only the filters the contract serves are offered: listLearnArticles takes q, type, topic, level and lang; listTrainingSeries takes topic, level and lang only."
repository_or_source_context: "branch feature/web/fidelity-learn-filters from the integration branch 341d497; S7 web pages; prototypes/html/pages/apprendre.html and assets/js/pages/learn.js"
files_or_sections_affected: ["src/web/src/features/learn/LearnPage.tsx", "src/web/src/features/learn/LearnPage.test.tsx", "src/web/src/features/learn/queries.ts", "src/web/src/features/learn/routeSearch.ts", "src/web/src/features/learn/messages.ts", "src/web/src/features/learn/learn.css", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-learn-filters.md"]
third_party_sources_consulted: ["@astryxdesign/core installed sources and docs (SegmentedControl, Selector, TextInput, VisuallyHidden), read from node_modules"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: RED tests first (17 new or changed tests failing), then the code", "new tests: chips, search and selects in one area with no chip counts; each filter sends its query parameter to the endpoint that serves it and not to the other; URL read and write; unknown values dropped; one reset clears every filter; a filter change restarts at the first page (no cursor); empty state per filter and per type; English copy", "changed tests (owner fidelity decision): the reset is « Réinitialiser » instead of « Tous les sujets »; the featured series is listed again in the list, as in the prototype; the unfiltered empty state speaks of « contenu »", "fixed on the way: the topic question feed used solve.state.* keys removed by the solve pass (typecheck and two tests failed on 341d497)", "captures of prototype and app at 1280 and 375 px in the session scratchpad", "pnpm typecheck, lint and format:check exit 0; full test run with --maxWorkers=1"]
open_risks: ["the series endpoint has no text search: while q is set the series rows are left out and a note says so; under Formations the search field is hidden", "two cursors cannot be merged honestly, so the list shows series first, then articles, each with its own « Afficher plus » button", "Astryx field labels are larger than the prototype's small muted labels", "at 375 px the four segments scroll inside their strip instead of wrapping like the prototype chips"]
```

## Learn pages match their HTML prototypes (fidelity pass)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked that every page match its HTML prototype visually. The AI rebuilt the layout of apprendre, article, formation and ecrire from prototypes/html/pages/{apprendre,article,formation,ecrire}.html on Astryx components and theme tokens, with contract fields only: featured course with its lessons, dense rows with type badge and icon, author, level, language code and date, topic questions as rows, an invitation for members; article pills, byline with avatar, country and date, Markdown body with copyable code and an outline; course hero with facts and numbered lessons; ecrire breadcrumb, form card, tips and footer. RED tests first, then the markup. Learn-local icons, row pieces and learn.css, to be swapped for the shared ones by the lead."
repository_or_source_context: "branch feature/web/fidelity-learn; S7 web pages; owner fidelity brief of 2026-11-15"
files_or_sections_affected: ["src/web/src/features/learn/LearnPage.tsx", "src/web/src/features/learn/ArticlePage.tsx", "src/web/src/features/learn/SeriesPage.tsx", "src/web/src/features/learn/WritePage.tsx", "src/web/src/features/learn/ArticleForm.tsx", "src/web/src/features/learn/present.tsx", "src/web/src/features/learn/icons.tsx", "src/web/src/features/learn/learn.css", "src/web/src/features/learn/messages.ts", "src/web/src/features/learn/LearnPage.test.tsx", "src/web/src/features/learn/ArticlePage.test.tsx", "src/web/src/features/learn/SeriesPage.test.tsx", "src/web/src/features/learn/WritePage.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-learn.md"]
third_party_sources_consulted: ["@astryxdesign/core installed docs and sources (Badge, Avatar, Icon, Link, Button, Markdown, Outline, CodeBlock, VisuallyHidden), read from node_modules", "icon paths from the project's own prototypes/html/assets/css/icons.css"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED then GREEN per page: new tests failed on the old pages, pass on the new ones (article, formation and ecrire RED taken by stashing the new page under memory pressure)", "changed tests: apprendre rows show the author name without « Par » and no tags (prototype row), « Leçons : 2 » became « 2 leçons »; the code-block check matches the code element by its whole text (highlighting splits it)", "pnpm --filter web typecheck, lint and format:check exit 0; learn tests plus translate, AskAboutLesson and ConnectPage: 92 passed", "captures prototype vs app at 1280 and 375 px in the session scratchpad"]
open_risks: ["the current shell paints the content area with the surface colour, so the surface cards and lists read flat until the new shell (feature/web/prototype-shell puts content on the page background) is merged", "progress ticks left out although the contract has Lesson.my_progress and a progress endpoint (a write flow, out of a visual pass)", "type, level and language filters left out (contract-backed, new behaviour)", "no real screen reader run"]
```

## Fidelity leftovers: Discover type chips, #ask on /solve, landing landmarks

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "Three leftovers of the prototype-fidelity pass. The Discover type chips showed truncated labels at tablet width: they now take their own grid row and scroll inside their strip if still too narrow. The shell link « Poser une question » (/solve#ask) did nothing when already on /solve: SolvePage now reacts to the hash changing, opens the ask view and focuses its first field, and clears the hash when the view closes so the link works again. The landing section heads are div elements instead of header elements, so they cannot count as banner landmarks."
repository_or_source_context: "branch feature/web/fidelity-leftovers; S7 web pages; prototypes/html/pages/discover.html"
files_or_sections_affected: ["src/web/src/features/discover/DiscoverPage.tsx", "src/web/src/features/discover/discover.css", "src/web/src/features/solve/SolvePage.tsx", "src/web/src/features/solve/SolvePage.test.tsx", "src/web/src/features/landing/LandingPage.tsx", "src/web/src/features/landing/LandingPage.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-leftovers.md"]
third_party_sources_consulted: ["none"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: navigating to /solve#ask while on the list did not open the ask view (1 failing test), then GREEN", "landing banner test (one banner landmark) already passed in jsdom before the change, kept as a guard", "captures of /discover at 375, 800, 1024 and 1280 px in the session scratchpad"]
open_risks: ["the prototype has no type chips (a select in a collapsible panel); the app keeps its segmented control", "the type chips fix is CSS only and checked by capture, not by a unit test"]
```

## Tablet icon rail (S7 fidelity pass)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked every page to match its prototype. From 768 to 1023 px the prototype shows a 92 px icon rail (icon over a short label); the app kept the 232 px sidebar, which crowded Découvrir. The AI wrote RED tests first, then a RailNav in the shell (flat list of areas, same order, aria-current on the current one, same « Navigation principale » label) chosen by a matchMedia hook on the tablet range, its CSS, the rail-wide footer and brand offsets, and viewport-fit=cover in index.html for the bottom bar's safe-area padding."
repository_or_source_context: "branch feature/web/fidelity-rail from feature/web/prototype-fidelity; S7 web shell; prototypes/html/assets/css/layout.css (.nav-item rail rules)"
files_or_sections_affected: ["src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "src/web/src/app/navigation.ts", "src/web/src/styles/shell.css", "src/web/index.html", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-rail.md"]
third_party_sources_consulted: ["none beyond the project prototype CSS and the installed Astryx AppShell"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED first: the rail test failed (the sidebar group titles were still shown)", "new tests: rail links, order and decorative icons; only the current area has aria-current; one primary navigation and no bottom bar; no rail above 1023 px", "captures of /discover and / at 800x1000 and 1024x800, light and dark, next to the prototype (scratch, not committed)", "gates in src/web: typecheck 0, lint 0, format:check 0, vitest run --maxWorkers=1 (see PR)"]
open_risks: ["the Découvrir type filter still truncates at 800 px (« S... », « T... ») because it is a feature-owned control, not the shell", "rail labels are 11 px, as in the prototype, below the 16 px body rule", "no real screen reader run"]
```

## Shell gaps after the prototype shell (S7 fidelity pass)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked to close every remaining difference between the app shell and the HTML prototype shell. The AI wrote RED tests first, then: the phone bottom bar (Accueil, Apprendre, Résoudre, Découvrir, « Plus ») that opens the Astryx MobileNav drawer from the left with the rest of the navigation and the account entries; the « Poser une question » button for members, linking to /solve#ask; the footer with 226os, Licence and Contribuer; « Paramètres » instead of « Réglages »; the phone search as the second line of the top bar; the 226os logo above a 400 px borderless card on the account pages (Brand moved to components/); the breadcrumb convention recorded on PageHeader with an order test. Follow-ups: the footer links open in a new tab like every external link of the app (Astryx Link, « nouvel onglet » label), the drawer account group is named « Compte », and a scroll padding keeps scrolled-to controls above the phone bar. It also added one TODO.md line and merged the updated integration branch (conflicts in icons.tsx and TODO.md, both sides kept)."
repository_or_source_context: "branch feature/web/fidelity-shell-gaps from feature/web/prototype-fidelity; S7 web shell; prototypes/html/assets/js/shell.js and assets/css/layout.css"
files_or_sections_affected: ["src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "src/web/src/app/AuthLayout.tsx", "src/web/src/app/AuthLayout.test.tsx", "src/web/src/app/navigation.ts", "src/web/src/components/Brand.tsx", "src/web/src/components/icons.tsx", "src/web/src/components/PageHeader.tsx", "src/web/src/components/ui.test.tsx", "src/web/src/styles/shell.css", "src/web/src/i18n/messages.ts", "src/web/src/i18n/translate.test.ts", "src/web/src/i18n/I18nProvider.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-shell-gaps.md"]
third_party_sources_consulted: ["@astryxdesign/core 0.7.14 installed type declarations and sources (AppShell mobileNav config and mobile context, MobileNav side, MobileNavToggle, Button, Card variants, TopNav), read from node_modules", "icon paths copied from prototypes/html/assets/css/icons.css (project-owned)"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: RED tests first (15 failing for the missing behaviour), then the code", "new tests: bottom bar items, hrefs, aria-current and decorative icons; no bar above the phone breakpoint; « Plus » opens the drawer (aria-expanded, aria-controls, aria-modal, data-side start); a drawer link and Escape close it; visitor and member account entries in the drawer, sign-out from it; ask button for members only, href /solve#ask; footer contentinfo with only existing documents; « Paramètres » in the menu and the document title; PageHeader order; account page logo and dark variant", "changed tests: the phone drawer test opens it with « Plus » instead of the removed top-bar toggle and checks a drawer-only link (the bar now shows Résoudre); the top bar is found by its label « Barre du site » because the landing page's own <header> elements are also banner landmarks in jsdom; the « Réglages » expectations became « Paramètres » (owner decision)", "captures at 1280 and 375, light and dark, signed out and signed in (scratch Vite config and session stub, not committed), next to accueil.html, discover.html and connexion.html; drawer opened in a real browser: left side, focus inside, Escape closes", "final captures at 375, 768 and 1280, light and dark, signed out and in, after the merge", "gates on the merged state (src/web): format:check 0, lint 0, typecheck 0, vitest run --maxWorkers=1 0 (49 files, 619 tests)"]
open_risks: ["index.html lacks viewport-fit=cover, so env(safe-area-inset-bottom) is 0 on iOS", "following « Poser une question » while already on /solve does not open the ask view: SolvePage reads the hash only in its initial state (features/solve, another agent)", "the footer links leave the site for the public GitHub repository", "no real screen reader run", "from 768 to 1023 px the prototype shows a 92 px icon rail; the app keeps the 232 px sidebar (not built in this pass)", "in Chromium focus() already scrolls clear of the bar, so the browser check did not isolate what the scroll padding adds"]
```

## Page fidelity: solve and question match their prototypes

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked that every S7 page match its HTML prototype visually. For /solve and /questions/<id> the AI wrote RED tests first, then rebuilt the markup and styles inside features/solve: dense list rows (score and answer count with icons, title link covering the row, plain-text excerpt, state pill, #tags, initials avatar, author country code, date, original language), a state filter on the listQuestions state parameter, the tag filter behind a filter toggle, the ask form as its own view with a back link and the assistant panel beside it, and the question page with header meta, vote column, answer cards with the accepted pill, answer sources as safe external links, context and linked items, and a sign-in invitation. Behaviour, routes, query keys, CSRF, idempotency and the assistant behaviour are unchanged."
repository_or_source_context: "branch feature/web/fidelity-solve; S7 web pages; prototypes/html/pages/solve.html and question.html"
files_or_sections_affected: ["src/web/src/features/solve/SolvePage.tsx", "src/web/src/features/solve/QuestionPage.tsx", "src/web/src/features/solve/QuestionRow.tsx", "src/web/src/features/solve/QuestionForm.tsx", "src/web/src/features/solve/Interactions.tsx", "src/web/src/features/solve/present.tsx", "src/web/src/features/solve/icons.tsx", "src/web/src/features/solve/solve.css", "src/web/src/features/solve/routeSearch.ts", "src/web/src/features/solve/queries.ts", "src/web/src/features/solve/messages.ts", "src/web/src/features/solve/SolvePage.test.tsx", "src/web/src/features/solve/QuestionPage.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-solve.md"]
third_party_sources_consulted: ["@astryxdesign/core installed type declarations and sources (Avatar, Icon registry, SegmentedControl, VisuallyHidden, Link, TextInput, token CSS), read from node_modules"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: RED tests first (9 failing), then the markup", "new tests: dense row content (score, answers, excerpt, country, date, language, lang attribute), state filter sent to the server, ask view with focus on its heading and a way back, question header meta, context and linked items, answer sources (http(s) only), sign-in note", "changed tests: h1 is now Questions; the byline shows the name without « Par »; the accepted state reads « Réponse acceptée »; a zero answer count is shown, as in the prototype; the tag field sits behind the Filtres toggle", "captures of prototype and app at 1280 and 375 px, light and dark, in the session scratchpad", "pnpm --filter web typecheck, lint and format:check exit 0; test exit 0 (46 files, 514 tests, run with --maxWorkers=2 after two memory-guard kills)"]
open_risks: ["the signed-in captures (ask view, answer form) used a dev-server stub of GET /auth/session with the synthetic test member; nothing was posted, so publishing was checked in jsdom only", "captures were taken on the current shell (white content panel); the new shell puts content on the page background, where the white cards are meant to stand out", "the state filter is new on the page (served by the contract, shown by the prototype); drop it if the owner does not want it", "at 375 px the four state labels scroll inside their strip (« Résolues » is off screen) to keep the larger target size", "the #ask view opens from the URL hash only on page load and does not set the hash itself", "the prototype's two-step ask form (title, then details) and its similar-questions panel are not built: the assistant panel keeps that place", "inline SVG glyphs in features/solve/icons.tsx stand in for the shared icon set"]
```

## GitHub error detail and one retry in the review scripts

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent)"
purpose: "Batch run 43673158317 failed in the publish step with only `FAILED (gh failed)`: no status, no message, a transient error that a rerun cleared. A failed gh call now reports the method, the endpoint without query string, the exit code and the first line of stderr (200 characters at most, secret shapes redacted, no request body), and is retried once after a short delay when that is safe (any GET except a 4xx; a write only after a 5xx, a secondary rate limit, a network error, or a 429 with Retry-After of 10 s at most)."
repository_or_source_context: "branch fix/review/gh-error-detail; owner-approved pipeline improvement"
files_or_sections_affected: ["scripts/review-context.py", "scripts/findings.py", "scripts/tests/review-batch-nothing_test.sh", "TODO.md", "docs/ai-usage/2026-11-15-gh-error-detail.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #232)"]
tests_or_evidence: ["RED first: six new cases in review-batch-nothing_test.sh failed before the change (502 once, GET once, 502 twice, 403, secret in stderr, GET 404)", "GREEN: review-batch-nothing_test.sh, findings-settle_test.sh, findings_test.sh, findings-batch_test.sh, review-autofix_test.sh, review-context_test.sh"]
open_risks: ["not run against the real GitHub API: the exact stderr wording of gh for 5xx and rate limits is matched on `HTTP <status>` and a few phrases", "a write retried after a 5xx may already have gone through: the comment writes stay idempotent by marker on the next run, but a retry inside one run can duplicate a comment"]
```

## Landing v2: landing-only header and footer, Africa globe hero, photo sections, scroll reveals (Geist, landing only)

```yaml
date: "2026-11-15"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent) led by Claude Opus 5.5"
purpose: "Build version 2 of the public landing after the owner's feedback: landing-only header and footer, full-viewport hero with one call to action and a large Africa-centred dotted globe with arcs between ten African cities, five distinct sections, Unsplash photos and decorative portrait circles, a reusable data-reveal scroll system, count-up of real totals, demo content excluded, FR and EN in one language each."
repository_or_source_context: "branch feature/landing/v2 (follows the merged PR #307); src/web/src/features/landing, src/web/src/app/Shell.tsx; captures of the dev server reading GET /api/v1 of production anonymously"
files_or_sections_affected: ["src/web/src/features/landing/**", "src/web/src/app/Shell.tsx", "src/web/src/app/*.test.tsx", "src/web/src/components/icons.tsx", "src/web/src/styles/shell.css", "src/web/public/landing/**", "TODO.md"]
third_party_sources_consulted: ["https://github.com/vercel/geist-font v1.9.2 (Geist[wght].woff2, OFL-1.1, licence file copied)", "https://github.com/nvkelso/natural-earth-vector ne_110m_land.geojson (public domain), used offline only to compute dot coordinates", "Unsplash photos listed with photographer and page in src/web/public/landing/CREDITS.md"]
license_review: "Geist: SIL OFL-1.1; Natural Earth: public domain; photos: Unsplash License (free use, no attribution required), pages checked on 2026-11-15; no new npm dependency. Owner to confirm the photo licence."
closure: "open: waits for the owner's decision and review"
human_reviewers: []
tests_or_evidence: ["RED: LandingPage.test.tsx rewritten first, 20 of 28 tests failed against the v1 page", "GREEN: pnpm exec vitest run src/features/landing (28 PASS); whole web suite, typecheck, lint, format:check see the final report", "Playwright captures 1440, 768 and 390 px, dark and light, FR and EN, hero and full page, no horizontal scroll", "pixel-sampled contrast audit of every text node, worst real case 5.02:1"]
open_risks: ["owner decision: Geist vs DESIGN.md section 4 (system fonts)", "owner decision: « Membres » on the landing vs « Connect » in the app", "owner decision: Unsplash photos and licence", "the member avatar chip was dropped, so the landing no longer depends on #306", "totals of a list longer than one page include demo items (the server total is used)", "no axe run (@axe-core/playwright not installed), no screen reader run", "the compact menu starts below 1024 px, not 768 px, because the full bar does not fit at 768 px"]
```

## Notes closed after triage in the review batch

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent, brief by the owner's lead session)"
purpose: "Notes issues (label kind:note) stayed open forever because the settle step never ticked a NOT A DEFECT or DECIDED verdict (rule H5 of #187). New rule (owner, 2026-11-15): for a note, such a verdict ticks the box when it cites a path:line, path:a-b or docs/decisions/NNNN-*.md that the script finds at the tip of develop, and the issue closes as not planned when every box is settled. Findings keep the old rule."
repository_or_source_context: "branch fix/review/notes-close-after-triage from develop 3ca5862"
files_or_sections_affected: ["scripts/findings.py", "scripts/tests/findings-settle-notes_test.sh", "scripts/tests/findings-settle_test.sh (comment)", "docs/quality/REVIEW_CONTEXT.md", ".claude/commands/review-fix.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["Backtick rule (main session): the plain-word id row and the updated segment row failed first (47 passed, 2 failed); after names_id requires `id` in backticks, findings-settle-notes 49 passed, findings-settle 41, findings-batch 48, review-autofix 205", "RED: findings-settle-notes_test.sh 2 failed (note ticked and issue closed as not planned; mixed issue) before the change", "RED (#266 deep review, rule tightened): findings-settle-notes_test.sh 5 failed (own-file cite, unparsable own line, non-ACCEPTED ADR, ADR not naming the file ticked or closed)", "RED (skeptic gaps): findings-settle-notes_test.sh 3 failed (PROPOSED ADR with a later accepted line, ACCEPTED superseded, ADR naming the file only in longer paths or a URL); GREEN after anchoring the first Status line and a path-boundary match; MUTATION CHECK, not a true RED (code was written before the rows; owner, ADR must name the note id): findings-settle-notes_test.sh 3 failed (ADR naming the file but not the id, longer ids, other case) with the id check disabled; RED (skeptic, second round): rows 69, 70 and the path-mention unit row failed before the fix (3 failed, 45 passed); GREEN: 48 passed (id read from the issue body, file mentions masked, `/` and `.word` continue a token); earlier GREEN: 45 passed after threading the id into verified_reference; GREEN earlier: findings-settle-notes_test.sh 36 passed; findings-settle_test.sh 41 passed; findings, findings-settle, findings-batch, review-batch-nothing, review-autofix tests pass"]
open_risks: ["NOT RUN on a real runner", "the no-commit path (review-autofix.py nothing) shares settle_issue, so it ticks and closes notes the same way"]
```

## P0 review notes #155 and #162 (vote quota order, assistant fair share)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (go-implementer subagent)"
purpose: "Owner decision to fix three review notes. The AI wrote the failing tests first, then: (1) a vote on an unknown or removed target is checked before the write budget is spent (a new read query, the same 404); (3) a fair share of the global assistant guard (3 of 12 provider calls an hour per member, ADR-0012 amended). Note note-answers-cursor-dup was analysed and not changed: it needs a contract or order change, so it waits for the owner."
repository_or_source_context: "branch fix/api/p0-notes-155-162, from develop 2af62a2"
files_or_sections_affected: ["src/api/internal/solve/**", "db/queries/solve.sql", "src/api/internal/platform/dbgen/solve.sql.go", "src/api/internal/assistant/service/**", "src/api/internal/platform/pgtest/*_integration_test.go", "docs/decisions/0012-ai-assistant-integration.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["go test -race ./..., coverage-go, go test -tags integration (see the pull request for the exit codes)"]
open_risks: ["note-answers-cursor-dup is not fixed", "a member now gets at most 3 sourced assistant answers an hour"]
```

## Profile photo from the member's GitHub avatar

```yaml
date: "2026-11-15"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent) led by Claude Opus 5.5"
purpose: "A profile photo taken from the member's GitHub avatar: the ProfilePhoto component (validated login, lazy image, no referrer, initials fallback), its use in the profile hero and the Connect rows, the help line on the GitHub field, the CSP img-src entry and their tests."
repository_or_source_context: "branch feature/profile/web-github-photo; TODO.md S7 block; api/openapi.yaml read only (github field of MyProfile, PublicProfile, ProfileSummary)"
files_or_sections_affected: ["src/web/src/components/ProfilePhoto.tsx", "src/web/src/components/ProfilePhoto.test.tsx", "src/web/src/features/profile/**", "src/web/src/features/connect/**", "src/web/nginx/security-headers.conf", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency; pictures are loaded at run time from GitHub, not stored or redistributed"
closure: "open: waits for the owner's decision and review"
human_reviewers: []
tests_or_evidence: ["RED: ProfilePhoto.test.tsx failed on the missing module; the ProfilePage and ConnectPage photo and hint tests failed before the wiring", "GREEN: see the commit report (vitest, typecheck, lint, prettier)"]
open_risks: ["privacy: GitHub sees the viewer's IP address when a photo loads (no referrer sent); the owner decides whether to accept it", "the shell avatar stays initials: the session user carries no github field and no request was added", "CSP img-src conflicts with the other branch that adds opengraph.githubassets.com: the lead resolves it at integration"]
```

## Project thumbnails and the card view of the Build, Discover and Learn lists

```yaml
date: "2026-11-15"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Card view by default on Build, Discover and Apprendre with a Cartes / Liste switch (view search param, remembered per browser), and a thumbnail for each project in Build (list rows and project page) from the repository's GitHub social preview, with the monogram tile as fallback; no contract change, no migration, no server storage."
repository_or_source_context: "branch feature/build/web-project-thumbnails; TODO.md S5 block; DESIGN.md (iconography, low bandwidth); src/web/src/components/Avatar.tsx"
files_or_sections_affected: ["src/web/src/features/build/ProjectThumb.tsx", "src/web/src/features/build/ProjectThumb.test.tsx", "src/web/src/features/build/BuildPage.tsx", "src/web/src/features/build/ProjectPage.tsx", "src/web/src/features/build/build.css", "src/web/nginx/security-headers.conf", "src/web/src/components/ViewToggle.tsx", "src/web/src/components/CardGrid.tsx", "src/web/src/components/useListView.ts", "src/web/src/components/ViewToggle.test.tsx", "src/web/src/components/CardGrid.test.tsx", "src/web/src/components/useListView.test.tsx", "src/web/src/styles/ui.css", "src/web/src/i18n/messages.ts", "src/web/src/features/build/routeSearch.ts", "src/web/src/features/build/BuildPage.view.test.tsx", "src/web/src/features/discover/DiscoverPage.tsx", "src/web/src/features/discover/routeSearch.ts", "src/web/src/features/discover/DiscoverPage.view.test.tsx", "src/web/src/features/learn/LearnPage.tsx", "src/web/src/features/learn/present.tsx", "src/web/src/features/learn/routeSearch.ts", "src/web/src/features/learn/LearnPage.view.test.tsx", "TODO.md"]
third_party_sources_consulted: ["GitHub social preview images (opengraph.githubassets.com)"]
license_review: "NOT APPLICABLE: no new dependency; images are loaded from GitHub at view time, not copied or redistributed"
closure: "open: waits for the owner's decision and review"
human_reviewers: []
tests_or_evidence: ["RED: vitest run src/features/build/ProjectThumb.test.tsx fails (module ./ProjectThumb not found)", "GREEN: vitest run src/features/build: 7 files, 107 tests pass", "RED (card view): vitest run on the new view tests fails (modules ./ViewToggle, ./CardGrid, ./useListView not found; no Cartes / Liste buttons on the three pages)", "GREEN (card view): the same files pass, then the build, discover, learn and components suites", "pnpm run typecheck, pnpm run lint, prettier: PASS"]
open_risks: ["privacy: GitHub receives the viewer's IP for each preview (no referrer sent)", "bandwidth: one preview per project card or row, lazy-loaded", "CSP img-src now allows opengraph.githubassets.com", "Learn cards keep the existing icon tile (KindThumb), not the monogram: Learn rows have no monogram", "reverses the earlier no-remote-picture choice for projects: owner decision"]
```

## API data gaps for the prototype pages

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and the go-implementer subagent"
purpose: "Serve the data the HTML prototypes show and the API did not (owner decision 2026-11-15 ~07:00 GMT, R7 for these items, no migration): Project stage, license_spdx and country_code on reads and writes with list filters; a total on listResources, listProjects, listQuestions, listLearnArticles and listTrainingSeries; sort on listResources (name, verified), listProjects (updated, created, needs) and listQuestions (new, votes, unanswered) bound to the cursor; the profile activity filters owner, author and answers_by under the getProfile visibility rule."
repository_or_source_context: "branch feature/api/prototype-data-gaps from origin/develop; ADR-0004 (cursors, problems), ADR-0005 (visibility), migration 00007 (projects columns), prototypes/html/pages/{discover,build,project,profile,apprendre,solve}.html for the sort options and counters"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apigen/apigen.gen.go", "src/api/internal/platform/apispec/openapi.gen.yaml", "src/web/src/api/schema.d.ts", "db/queries/{build,discover,learn,solve}.sql", "src/api/internal/platform/dbgen/**", "src/api/internal/{build,discover,learn,solve}/**", "src/api/internal/platform/pgtest/{project_facts,list_totals,list_sort,profile_activity}_integration_test.go", "docs/SPEC.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["Y4NN777 (owner, pending)"]
tests_or_evidence: ["RED then GREEN per layer (repository, service, transport) for each item; dbgen bind-order tests for every changed query", "real-PostgreSQL integration: TestBuildProjectFacts, TestListTotals*, TestListSort*, TestProfileActivity* (walk every page, total equals the items returned, cursor of another sort is 400 INVALID_CURSOR, hidden or private username is an empty page)", "EXPLAIN (ANALYZE, BUFFERS) on 3000 projects and 3000 questions: every list and count under 8 ms, in shared buffers (lead report)"]
open_risks: ["no index serves the new orders (projects updated_at, questions vote_score, resources verified_at) nor the open-needs order: sequential scan plus top-N sort, fine for P0 volumes; an index needs a migration (mibienpanjoe)", "the listProjects default order moved from published_at to updated_at (owner decision)", "the web typings now require total, stage, license_spdx and country_code: web fixtures may need the new fields"]
```

## Prototype fidelity integration: merges, copy rewrite tooling, fidelity report and README screenshots

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) with web-implementer subagents and two proof:screen-judge reviews"
purpose: "The lead session integrated the page-fidelity branches (shell, rail, harmonize, learn filters, home-auth, leftovers, controls, copy rewrite parts 1 and 2) into feature/web/prototype-fidelity and resolved their conflicts. It also applied the owner-approved copy rewrite by script (only where a string still had its audited value), restored the assistant privacy fact, captured the app next to the prototypes, had an independent reviewer judge them twice, and wrote the fidelity report, the copy style guide and the README screenshots section."
repository_or_source_context: "branch feature/web/prototype-fidelity; prototypes/html/pages; docs/quality/ASTRYX_REPORT.md section 8"
files_or_sections_affected: ["docs/quality/PROTOTYPE_FIDELITY.md", "docs/quality/evidence/fidelity/2026-11-15/", "docs/quality/ASTRYX_REPORT.md", "docs/team/COPY_STYLE.md", "docs/team/OWNER_PREFERENCES.md", "README.md", "TODO.md", "src/web/src/components/{List,PageHeader,StatusBadge,icons}.tsx", "src/web/src/features/{connect,settings,solve,discover,search}/", "src/web/src/i18n/messages.ts"]
third_party_sources_consulted: ["copy research sources listed in docs/team/COPY_STYLE.md (Microsoft fr-FR localization guide, GitHub Primer, Google developer style guide, Atlassian, Mailchimp, OQLF, Mastodon and Discourse FR translations)"]
license_review: "NOT APPLICABLE: original text and code, no dependency added; screenshots are of this project"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["web gates on the merged branch: typecheck, lint, format:check PASS; Vitest 48 files, 646 tests PASS", "two independent screen-judge passes, verdicts in docs/quality/PROTOTYPE_FIDELITY.md", "58 captures in docs/quality/evidence/fidelity/2026-11-15/"]
open_risks: ["PageHeader is a div, not a header element, so it is not an extra banner landmark; tests locate .ui-page-header", "no signed-in captures and no dark captures of detail pages", "the copy has not been reviewed by a native French editor"]
```

## Prototype shell and shared page pieces (S7 phase 0)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, web-implementer subagent of the lead session"
purpose: "Owner decision (2026-11-15, about 05:10 GMT): every page must match its HTML prototype. Phase 0 only: the AI rebuilt the app shell after prototypes/html (shell.js, layout.css) on the Astryx AppShell, TopNav and SideNav (logo linking home, site search in the top bar going to /recherche?q=, light/dark toggle, a Préférences menu with language and theme radios, Se connecter and Créer un compte for visitors, the account menu behind the member's initials, sidebar icons, active fill, « Bientôt » style, 1200 px content column, phone header and search), mapped the remaining Graphite & Lime values onto Astryx token slots (status fills, icon tints, skeleton, the strong input border), wrote the shared pieces PageHeader, Breadcrumbs, ResultsMeta, List and ListRow, StatusBadge, Meta and MetaItem, Avatar, CountryCode, Tag and TagList, EmptyState, gave ErrorState the prototype .state look and LoadingState list placeholder rows, put the 404 page on the empty state, and wrote src/web/scripts/capture-fidelity.sh to screenshot app routes next to their prototype page. Tests were written first (RED) for the shell and the components. Navigation labels and order stay those of the product (O7), not the prototype's."
repository_or_source_context: "branch feature/web/prototype-shell (stacked on fix/web/a11y-pass-109, PR #233); S7 web; prototypes/html as visual reference only (DISCLAIMER.md)"
files_or_sections_affected: ["src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "src/web/src/app/App.test.tsx", "src/web/src/app/navigation.ts", "src/web/src/components/icons.tsx", "src/web/src/components/PageHeader.tsx", "src/web/src/components/Breadcrumbs.tsx", "src/web/src/components/ResultsMeta.tsx", "src/web/src/components/List.tsx", "src/web/src/components/StatusBadge.tsx", "src/web/src/components/Meta.tsx", "src/web/src/components/Avatar.tsx", "src/web/src/components/CountryCode.tsx", "src/web/src/components/Tag.tsx", "src/web/src/components/EmptyState.tsx", "src/web/src/components/ErrorState.tsx", "src/web/src/components/LoadingState.tsx", "src/web/src/components/ui.test.tsx", "src/web/src/features/NotFound.tsx", "src/web/src/features/solve/QuestionPage.test.tsx", "src/web/src/theme/ThemeProvider.tsx", "src/web/src/theme/theme226os.ts", "src/web/src/i18n/messages.ts", "src/web/src/styles/index.css", "src/web/src/styles/shell.css", "src/web/src/styles/ui.css", "src/web/src/assets/brand/226os-logo.png", "src/web/src/assets/brand/226os-logo-white.png", "src/web/scripts/capture-fidelity.sh", "TODO.md", "docs/ai-usage/2026-11-15-prototype-shell.md"]
third_party_sources_consulted: ["@astryxdesign/core installed docs, type declarations and sources (AppShell, useAppShellMobile, TopNav, SideNav, DropdownMenu, IconButton, TextInput, Avatar, Badge, Breadcrumbs, EmptyState, Icon, Skeleton, Link, VisuallyHidden, input border token) read from node_modules and `astryx docs icons`", "gstack browse binary (local tool) for the screenshots"]
license_review: "NOT APPLICABLE: original code, no new dependency. The icon paths come from the project's own prototype icon set (prototypes/html/assets/css/icons.css) and the logo files from prototypes/html/assets/brand (AI-generated, already recorded)"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED first: Shell tests (logo link, top bar search to /recherche?q= and without a query, theme toggle aria-pressed and storage, nav icons aria-hidden, aria-current on the active item, Créer un compte link, theme radios, « Bientôt » item not a link) and ui.test.tsx failed before the code", "changed tests: the anonymous menu trigger is now « Préférences » / Preferences (was « Plus d’options »), the top bar search is a search field (was a link to /recherche), « Créer un compte » is looked up inside the banner, the QuestionPage raw-HTML test looks for img/script inside main (the shell logo is now an img)", "screenshots of / and /discover against accueil.html and discover.html at 1280x800 and 375x812, light and dark, with capture-fidelity.sh against the worktree dev server", "pnpm --filter web typecheck, lint, format:check and test: see the PR"]
open_risks: ["the shell still differs from the prototype in places: no phone bottom bar, no « Poser une question » create button, the phone search scrolls with the page instead of staying in the sticky header, the drawer opens on the Astryx side and the sidebar keeps the Progresser and Piliers section titles", "global token changes (status fills, icon tints, strong input borders) also restyle Astryx badges and inputs on pages not touched here", "ErrorState is no longer an Astryx Banner: its look changed on every page that uses it", "no real screen reader run; axe NOT RUN"]
```

## README brought up to date with develop

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (owner's lead session, no subagent)"
purpose: "The README described develop as of the afternoon of 14 Nov. Checked against develop, the AI updated it: Connect (#193, #220), moderation (S6: API and runbook) and the local demo data (#215) are delivered; only the four P1 operations still answer 501 (scripts/unimplemented-routes.txt); /connect is live with country, language and skill filters and the member opt-in, while mentoring and spaces stay « Bientôt »; error reports and moderation are described; an optional make seed step is added; one sentence describes the automated review flow."
repository_or_source_context: "branch docs/readme/state-2026-11-15, PR #235 (merged 2026-11-15)"
files_or_sections_affected: ["README.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: documentation only, no dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #235)"]
tests_or_evidence: ["checked against develop and scripts/unimplemented-routes.txt; scripts/check-docs.py PASS (PR #235)"]
open_risks: ["record written on 2026-11-15, after the merge: none was filed at the time", "the README was rewritten again by #237 the same day"]
```

## Review fix of batch 334: freeze time in the release checklist

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code in CI (review-batch), Claude Opus 5.5 orchestrator"
purpose: "Work the 15 items of review batch 334. One fix: the release-day checklist in TODO.md still gave the Sat 22:00 GMT freeze; a dated line now records the owner's move to Sun 15 Nov 12:00 GMT (note phase-doc-sync). The other items were settled without a change: two notes on a file removed by another PR, design points of ADR-0015 and deploy notes held by the owner, and two web test notes that are not defects."
repository_or_source_context: "branch fix/review/batch-2026-11-15, batch issue 334"
files_or_sections_affected: ["TODO.md", "docs/ai-usage/2026-11-15-review-fix-batch.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["RED: grep 'freeze moved to Sun 15 Nov 12:00 GMT' TODO.md found 0 lines before the edit", "GREEN: the same grep finds the new dated line under the release-day checklist", "bash scripts/check-todo-updated.sh: NOT RUN in the fixer (command not allowed; CI lint-scripts runs it)", "make lint, make test-unit, make coverage: NOT RUN (documentation-only change, no code touched)"]
open_risks: []
```

## Review fix of PR 252: E2E specs aligned with the rendered shell

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code in CI (review-autofix), Claude Opus 5.5 orchestrator, web-implementer subagent"
purpose: "Fix review finding e2e-fails-on-this-tree: the E2E language test looked for role menuitem while the shell renders menuitemradio, and the mobile bottom-bar test needs a « Navigation rapide » bar not on this tree. The other findings of the run (scripts/secrets.sh, scripts/deploy-images.sh, .github/workflows/deploy.yml) are owner paths and were left to the owner."
repository_or_source_context: "branch feature/ops/deploy-pipeline, PR 252"
files_or_sections_affected: ["src/web/e2e/anonymous.spec.ts", "src/web/e2e/mobile.spec.ts", "TODO.md", "docs/ai-usage/2026-11-15-review-fix-pr252.md"]
third_party_sources_consulted: []
license_review: "PENDING"
closure: "open"
human_reviewers: []
tests_or_evidence: ["RED: the E2E tests themselves are the defect; their failure is shown statically (Shell.tsx renders DropdownMenuRadioItem; Shell.test.tsx asserts menuitemradio; no « Navigation rapide » in src/web/src). Playwright NOT RUN in the fixer (no browser, no backend, not allowlisted)", "GREEN: pnpm --filter web run test (Shell.test.tsx, 510 tests) PASS; lint PASS; typecheck PASS; format:check PASS; E2E NOT RUN", "make lint: see the commit run"]
open_risks: ["the E2E fix is not confirmed by a Playwright run", "the bottom-bar test is test.fixme until feature/web/prototype-fidelity merges; remove the fixme then"]
```

## Review fix of PR 286: the typo fallback keeps negations

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code in CI (review-autofix), Claude Opus 5.5 orchestrator, go-implementer subagent"
purpose: "Fix review finding typo-fallback-ignores-negation: the pg_trgm fallback compared the raw query to titles, and pg_trgm ignores '-', so 'paiemnt -mobile' returned the titles the negation excluded. The search service now skips the fallback when the query negates a word in websearch_to_tsquery syntax. Second pass (findings negation-phrase-not-detected, negation-check-misses-operator-forms): the check also trims the websearch operator characters before and after the '-', so a negated phrase or group and '!-word' skip the fallback too."
repository_or_source_context: "branch feature/discover/search-typo-tolerance, PR 286"
files_or_sections_affected: ["src/api/internal/search/service/service.go", "src/api/internal/search/service/fuzzy_test.go", "src/api/internal/platform/pgtest/search_typo_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-15-review-fix-pr286.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["RED: TestFuzzySearchSkipsTheFallbackWhenTheQueryNegatesAWord failed before the fix (5 negation cases ran the fallback)", "GREEN: go test -race ./src/api/internal/search/service/... PASS", "integration case 'paiemnt -mobile' in TestSearchToleratesTypos: NOT RUN in the fixer (no Docker)", "make lint, make test-unit, make coverage: see the commit run", "second pass RED: the negated phrase, negated group, negated open group, negated quoted word after the hyphen and negation after an operator cases of TestFuzzySearchSkipsTheFallbackWhenTheQueryNegatesAWord ran the fallback", "second pass GREEN: go test -run TestFuzzySearch -race ./src/api/internal/search/service/... PASS"]
open_risks: ["the integration case is not confirmed against a real database in the fixer", "a quoted '-word' inside a phrase also skips the fallback (conservative)"]
```

## Review follow-ups of PR #285 and PR #286 (issues #290 to #293)

```yaml
date: "2026-11-15"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Review follow-ups of #285 and #286. The AI checked websearch_to_tsquery on real PostgreSQL, found the Go negation heuristic wrong on four forms, wrote the failing integration test first, then moved the negation rule into the Similar* queries (no fallback when the tsquery holds a NOT) and removed the heuristic. On identity it added a warning when a resend mail fails (RED first), corrected the comments about concurrent resends, and wrote the accepted residual risk of #290 in ADR-0003, as chosen by the contributor."
repository_or_source_context: "branch fix/discover/review-followups-248, from develop c88634d"
files_or_sections_affected: ["db/queries/search.sql", "db/queries/identity.sql", "src/api/internal/platform/dbgen/**", "src/api/internal/search/service/**", "src/api/internal/identity/service/**", "src/api/internal/platform/pgtest/search_typo_integration_test.go", "docs/decisions/0003-*.md", "docs/architecture/ARCHITECTURE.md", "docs/spec-kit/DATABASE_DESIGN.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["TestSearchTypoFallbackFollowsThePostgreSQLNegation RED on the four forms, then GREEN; TestResendEmailVerificationLogsAFailedMail RED then GREEN; see the pull request for the gate results"]
open_risks: ["the #290 acceptance is a security acceptance; the owner can overrule it"]
```

## Review pipeline steward agent and its skill

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session)"
purpose: "The AI wrote a project agent dedicated to the maintenance and evolution of the AI review pipeline (.claude/agents/review-steward.md) and its preloaded skill (.claude/skills/review-pipeline/SKILL.md: map of the stages, ledger fields, invariants, failure history, test patterns, ADR-0015 roadmap), routed the pipeline files to that skill in the skill map, and pointed ci-implementer and CLAUDE.md to the new agent."
repository_or_source_context: "branch feature/harness/review-steward, from develop"
files_or_sections_affected: [".claude/agents/review-steward.md", ".claude/skills/review-pipeline/SKILL.md", ".claude/hooks/skill-map.json", ".claude/hooks/tests/test_domain_skills.py", ".claude/agents/ci-implementer.md", "CLAUDE.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["python3 -I -m unittest discover -s .claude/hooks/tests: OK (28 tests), RED first on the new skill-map cases", "check-docs PASS"]
open_risks: ["the skill describes ADR-0015 phase 0 rules from PR #309; it says to check that #309 is merged before relying on them"]
```

## Typo-tolerant search (issue #248, item 2)

```yaml
date: "2026-11-15"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Issue #248 item 2. The AI wrote the failing tests first, then: four bounded Similar* queries in db/queries/search.sql (pg_trgm word similarity over the partial trigram indexes of migration 00010, same predicates and filters as full text, negation-only queries read nothing), Repository.Similar, a service fallback when a group's full text finds nothing (GET /search only, 3 characters at least; the assistant's retrieval keeps exact matches), EXPLAIN evidence, and the SPEC, ARCHITECTURE and DATABASE_DESIGN text. It measured word_similarity on sample typos to keep the default 0.6 threshold."
repository_or_source_context: "branch feature/discover/search-typo-tolerance, from develop 25774f2"
files_or_sections_affected: ["db/queries/search.sql", "src/api/internal/platform/dbgen/search.sql.go", "src/api/internal/search/**", "src/api/internal/platform/pgtest/search_typo_integration_test.go", "src/api/internal/platform/pgtest/explain_helpers_integration_test.go", "docs/SPEC.md", "docs/architecture/ARCHITECTURE.md", "docs/spec-kit/DATABASE_DESIGN.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["unit tests RED first; TestSearchToleratesTypos fails with the fallback off (12 group mismatches) and passes with it; plans logged by TestSearchTypoFallbackUsesTheTrigramIndexes; see the pull request for the gate results"]
open_risks: ["swapped letters (pyhton) stay below the 0.6 threshold", "only titles and names are typo tolerant, not bodies"]
```

## Two slow Build tests made fast

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "Two Build tests timed out under full-suite load in CI. The AI measured them, found that the cost was the accessible-name role query over the whole page after opening a Selector (about 600 ms per pick, more with 54 countries), and added a test helper that opens the Selector by keyboard and finds the option by text inside the listbox. The create-form test was also split in two (stage and licence, then country), with the same assertions, and its 30 s timeout was removed. No assertion was removed or loosened and no timeout was raised."
repository_or_source_context: "branch fix/web/slow-build-tests, from develop 1045d41"
files_or_sections_affected: ["src/web/src/test/select.ts", "src/web/src/features/build/BuildPage.data.test.tsx", "src/web/src/features/build/ProjectForm.data.test.tsx"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: test code only, no dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["vitest --maxWorkers=1 on both files and on src/features/build: PASS; typecheck, lint, format:check PASS"]
open_risks: ["the helper looks the option up by text, not by accessible name"]
```

## Email verification resend (issue #248, item 1)

```yaml
date: "2026-11-15"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Issue #248 item 1. The AI wrote the failing tests first, then: the resendEmailVerification contract operation (owner-approved, R7), the identity service and handler (always 202, 3 per account and 10 per IP per hour counted before the lookup, token drawn before the lookup, token write and mail off the request path), CreateEmailVerification now burns every earlier open code of the account in the same statement, and the « Renvoyer le code » action on /verifier-email. Items 2 (trigram search) and 3 (learn_progress) are not in this change."
repository_or_source_context: "branch feature/identity/verification-resend, from develop b4a891c"
files_or_sections_affected: ["api/openapi.yaml", "generated apigen, apispec, schema.d.ts, dbgen", "db/queries/identity.sql", "src/api/internal/identity/**", "src/api/internal/app/server.go", "src/api/internal/platform/pgtest/identity_integration_test.go", "scripts/public-routes.txt", "src/web/src/features/identity/**", "docs/decisions/0003-*.md", "docs/SPEC.md", "docs/architecture/ARCHITECTURE.md", "docs/quality/COVERAGE_SCOPE.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["service, transport, server and Vitest tests RED first; the integration test fails on the old CreateEmailVerification (2 open codes) and passes on the new one; see the pull request for the gate results"]
open_risks: ["the old code dies at the resend, so a mail still in transit with the older code no longer works"]
```

<!-- ai-usage:generated:end -->
