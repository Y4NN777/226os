> Archived 2026-11-12. Historical record, not maintained.
> Current source: [AGENTS.md](../../AGENTS.md) and [CLAUDE.md](../../CLAUDE.md).
> Its snapshot of the repository (no `src/`, no Makefile, Claude Code 2.1.322) is out of date.

# 226os Claude Code Harness — Documented v0.1

**Repository examined:** https://github.com/Y4NN777/226os, default branch `main`, 2026-11-11.
**Status:** Harness contract documented for human review; runtime `.claude/` settings, six agents, five workflow skills, ten domain skills and the domain-skill hooks are installed and load-validated on Claude Code 2.1.322 (H0 record in `AI_USAGE.md`). This document does not grant application implementation or release approval.

## Verified baseline
- The remote tree contains PRD/SPEC/TEST/DEV/SHIP/RULES/ARCHITECTURE/DESIGN, AGENTS.md, spec-kit G0 proposals and Claude handoff, plus Git workflow and an existing branch-policy workflow.
- The tree does **not** yet contain src/web, src/api, db migrations, Makefile, Compose/Traefik manifests, OpenAPI file, operational tests or deploy pipelines.
- `docs/spec-kit/CLAUDE_HANDOFF.md` records CONDITIONAL GO for implementation planning; application implementation and release remain gated by conditions and evidence.
- `docs/spec-kit/G0_DECISIONS.md` contains proposed defaults; `docs/decisions/0001-g0-conditional-baseline.md` records which defaults are accepted with conditions and which decisions remain deferred.
- Root README/DESIGN are canonical; engineering docs live under docs/. Avoid a wholesale docs move just for harness setup.
- `docs/GIT_WORKFLOW.md` specifies `develop` as staging/preview and `main` as production; do not reintroduce a required `staging` branch for this harness.

## Design rationale
- A concise CLAUDE.md is an index and operating contract, not another copy of AGENTS.md.
- Six specialized subagents, not a swarm of autonomous editors: advisor (Opus), scout (Haiku), Go implementer (Sonnet), PostgreSQL expert (Sonnet), web implementer (Sonnet), skeptic reviewer (Sonnet).
- Five workflow skills (user-invoked): `/advisor`, `/start-slice`, `/spec-to-task`, `/verify-slice`, `/ship-check`.
- Ten domain skills built from the project docs plus cited external expertise: `engineering-principles`, `go-backend`, `postgres-data`, `api-contract`, `web-frontend`, `security-authz`, `tdd-unit`, `bdd-acceptance`, `integration-testing`, `ci-cd-delivery`. Project docs are binding; external practice is advisory and cited.
- Domain-skill enforcement, three layers:
  1. Implementer and reviewer agents preload their domain skills (`skills:` frontmatter).
  2. `.claude/hooks/domain_skills.py` maps file paths to skills via `.claude/hooks/skill-map.json`. On Read it adds a short directive (once per skill) to load the missing skills; on Edit/Write/MultiEdit/NotebookEdit it denies until every mapped skill is loaded through the Skill tool. Skill text is never inlined: Claude Code truncates long hook output to a ~2 KB preview (observed on 2.1.322), while the Skill tool loads the full skill.
  3. A PostToolUse hook on Skill records loaded skills per session and per subagent (subagent `skills:` preloads count as loaded); a SessionStart hook clears the record on compact/clear so skills are reloaded after context loss. An unreadable skill map or a missing SKILL.md blocks writes.
- Domain skills must not use `paths:` frontmatter: Claude Code then hides them until a matching file exists, which breaks the Skill tool and subagent preloads (observed on 2.1.322). Path routing lives only in `skill-map.json`.
- Hook tests: `python3 -m unittest discover -s .claude/hooks/tests`. Shell writes through Bash bypass the hook; review and CI remain the backstop.
- No Make/CI execution hooks yet, because Make targets and the execution contract do not exist. Add them after bootstrap.
- `.claude/settings.json` only adds narrow read/edit denies for typical local env and secret paths. These rules are defense-in-depth, NOT an OS sandbox; verify via `/permissions` and `/doctor`. No blanket Bash approval, no bypass mode.

## Agent routing
| Work | Owner | Delegate | Model / effort |
|---|---|---|---|
| Goal framing, architectural arbitration, synthesis | Main thread | none | Opus 5.5 / medium by default; high for hard cross-domain decisions |
| Independent architecture opinion | `/advisor` | advisor | Opus 5.5 / high, read-only |
| Targeted repo investigation | main | scout | Haiku 5.5 / low, read-only |
| Approved Go API vertical slice | main | go-implementer | Sonnet 5.5 / medium |
| Approved PostgreSQL schema/migration/query slice | main | postgres-expert | Sonnet 5.5 / high |
| Approved React/TanStack/Astryx vertical slice | main | web-implementer | Sonnet 5.5 / medium |
| Independent review on risky change | main | skeptic | Sonnet 5.5 / high, read-only |

Avoid assigning both implementers overlapping contracts, root config, migration or generated API files. Keep one integrator for shared files. `CLAUDE_CODE_SUBAGENT_MODEL_FORCE` must remain unset: it would override individual Sonnet/Haiku routing.

## Human approval before code
1. Close G0 owner/reviewer roster and acceptance statuses; record accepted ADRs, not just an in-chat assent.
2. Confirm security/auth/data/API decisions and ownership per `G0_DECISIONS.md`.
3. Pin and verify Astryx release and package manager via official package metadata.
4. Confirm P0 nonfunctional scope; keep Valkey/S3 candidate until ADR.
5. Confirm CADEV compliance review for source licensing, AI provenance and third-party dependencies.

## Rollout plan
- **H0** (harness-only): review and commit `CLAUDE.md` and `PROJECT_HARNESS_SPEC.md` on a documentation branch targeting `develop`. Runtime `.claude/` settings, agents and skills remain a follow-up until their exact files are added and validated. No application code.
- **G0**: use Opus main + `/advisor` for open decisions, write proposed ADRs and gather human approval. `ADR-0001` records a conditional baseline; do not silently turn deferred or conditional decisions into unconditional acceptance.
- **B0** (post-G0): Go + web + database + Make + Compose skeleton in bounded tasks. Sonnet workers, Haiku read-only scouting. Define and validate CI quality gates; no fake PASS.
- **B1**: Solve vertical slice, then Discover + Build and relations, followed by FR/EN, responsive, security, performance, CADEV demo.
- **SHIP**: require exact SHA, CI gate reports, Datacloud evidence, rollback and human approval.

## Operational pitfalls
- Do not use experimental agent teams initially. Their concurrency, shared task state and shutdown complexity are unnecessary for the first slices; subagents suffice. Evaluate opt-in later for independent areas.
- Do not turn G0 into a bureaucratic ban on creating the harness itself. Harness setup is allowed when the user explicitly approves the scope; app implementation remains blocked.
- 100% Go statement coverage and >70% on all four frontend metrics are configured gates once code exists; they are not factual results today.
- README public French; technical/agent files English.
- H0 load validation covers registration, deny-rule enforcement and one skill invocation only. Re-validate after CLI upgrades, and confirm provider model access before relying on Opus/Sonnet/Haiku routing.

## Source documentation
- https://code.claude.com/docs/en/sub-agents
- https://code.claude.com/docs/en/skills
- https://code.claude.com/docs/en/settings
- https://code.claude.com/docs/en/permissions
- https://code.claude.com/docs/en/hooks
- https://code.claude.com/docs/en/agent-teams
- https://code.claude.com/docs/en/model-config
- https://blog.buildfastwithai.com/claude-opus-5-prompts-tested