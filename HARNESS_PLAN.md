# 226OS Claude Code Harness — Documented v0.1

**Repository examined:** https://github.com/Y4NN777/226os, default branch `main`, 2026-11-11.
**Status:** Harness contract documented for human review; runtime Claude configuration is not installed yet. This document does not grant application implementation or release approval.

## Verified baseline
- The remote tree contains PRD/SPEC/TEST/DEV/SHIP/RULES/ARCHITECTURE/DESIGN, AGENTS.md, spec-kit G0 proposals and Claude handoff, plus Git workflow and an existing branch-policy workflow.
- The tree does **not** yet contain apps/web, apps/api, db migrations, Makefile, Compose/Traefik manifests, OpenAPI file, operational tests or deploy pipelines.
- `docs/spec-kit/CLAUDE_HANDOFF.md` records CONDITIONAL GO for implementation planning; application implementation and release remain gated by conditions and evidence.
- `docs/spec-kit/G0_DECISIONS.md` contains proposed defaults; `docs/decisions/0001-g0-conditional-baseline.md` records which defaults are accepted with conditions and which decisions remain deferred.
- Root README/DESIGN are canonical; engineering docs live under docs/. Avoid a wholesale docs move just for harness setup.
- `docs/GIT_WORKFLOW.md` specifies `develop` as staging/preview and `main` as production; do not reintroduce a required `staging` branch for this harness.

## Design rationale
- A concise CLAUDE.md is an index and operating contract, not another copy of AGENTS.md.
- Five specialized subagents, not a swarm of autonomous editors: advisor (Opus), scout (Haiku), Go implementer (Sonnet), web implementer (Sonnet), skeptic reviewer (Sonnet).
- Five skills: `/advisor`, `/start-slice`, `/spec-to-task`, `/verify-slice`, `/ship-check`.
- No repo-wide production shell hooks initially, because Make targets and execution contract do not yet exist. Add tested PreToolUse/Stop hooks after bootstrap, in a separate reviewed PR.
- `.claude/settings.json` only adds narrow read/edit denies for typical local env and secret paths. These rules are defense-in-depth, NOT an OS sandbox; verify via `/permissions` and `/doctor`. No blanket Bash approval, no bypass mode.

## Agent routing
| Work | Owner | Delegate | Model / effort |
|---|---|---|---|
| Goal framing, architectural arbitration, synthesis | Main thread | none | Opus 5.5 / medium by default; high for hard cross-domain decisions |
| Independent architecture opinion | `/advisor` | advisor | Opus 5.5 / high, read-only |
| Targeted repo investigation | main | scout | Haiku 5.5 / low, read-only |
| Approved Go + database vertical slice | main | go-implementer | Sonnet 5.5 / medium |
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
- **H0** (harness-only): review and commit `CLAUDE.md` and `HARNESS_PLAN.md` on a documentation branch targeting `develop`. Runtime `.claude/` settings, agents and skills remain a follow-up until their exact files are added and validated. No application code.
- **G0**: use Opus main + `/advisor` for open decisions, write proposed ADRs and gather human approval. `ADR-0001` records a conditional baseline; do not silently turn deferred or conditional decisions into unconditional acceptance.
- **B0** (post-G0): Go + web + database + Make + Compose skeleton in bounded tasks. Sonnet workers, Haiku read-only scouting. Define and validate CI quality gates; no fake PASS.
- **B1**: Solve vertical slice, then Discover + Build and relations, followed by FR/EN, responsive, security, performance, CADEV demo.
- **SHIP**: require exact SHA, CI gate reports, Datacloud evidence, rollback and human approval.

## Operational pitfalls
- Do not use experimental agent teams initially. Their concurrency, shared task state and shutdown complexity are unnecessary for the first slices; subagents suffice. Evaluate opt-in later for independent areas.
- Do not turn G0 into a bureaucratic ban on creating the harness itself. Harness setup is allowed when the user explicitly approves the scope; app implementation remains blocked.
- 100% Go statement coverage and >70% on all four frontend metrics are configured gates once code exists; they are not factual results today.
- README public French; technical/agent files English.
- This harness contract is not yet validated by a Claude Code runtime in this repository. Validate with the actual installed CLI version and user's provider model access before relying on model routing or skills.

## Source documentation
- https://code.claude.com/docs/en/sub-agents
- https://code.claude.com/docs/en/skills
- https://code.claude.com/docs/en/settings
- https://code.claude.com/docs/en/permissions
- https://code.claude.com/docs/en/hooks
- https://code.claude.com/docs/en/agent-teams
- https://code.claude.com/docs/en/model-config
- https://blog.buildfastwithai.com/claude-opus-5-prompts-tested
