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

## H0/G0 Claude prompt (XML)

Use the following prompt to start Claude Code as the 226OS lead orchestrator:

```xml
<mission>
	<role>Act as the 226OS lead technical orchestrator.</role>
	<primary_goal>Validate H0 harness readiness and prepare G0 readiness.</primary_goal>
	<language>English for repository artifacts; concise reporting.</language>
</mission>

<required_inspection>
	<file>AGENTS.md</file>
	<file>CLAUDE.md</file>
	<file>docs/spec-kit/CLAUDE_HANDOFF.md</file>
	<file>docs/spec-kit/G0_DECISIONS.md</file>
	<file>docs/spec-kit/G0_APPROVAL_TEMPLATE.md</file>
	<file>docs/spec-kit/TASKS.md</file>
	<file>docs/GIT_WORKFLOW.md</file>
	<item>Current Git branch, status, remotes, and recent history</item>
</required_inspection>

<execution_order>
	<step id="H0">Validate the Claude harness before application work.</step>
	<step id="G0">Identify G0 decisions, approvals, owners, reviewers, and blockers.</step>
	<step id="STOP">Stop before application implementation unless the explicit conditions below are satisfied.</step>
</execution_order>

<harness_validation>
	<requirement>Verify that .claude/settings.json loads on the installed Claude Code version.</requirement>
	<requirement>Verify that custom agents load and their scopes are effective.</requirement>
	<requirement>Verify that project skills load and can be invoked.</requirement>
	<requirement>Use the installed Claude Code validation commands or UI where available.</requirement>
	<rule>Do not claim that a configuration works without observed validation evidence.</rule>
	<rule>If a runtime file is absent, report it as missing; do not infer successful loading.</rule>
</harness_validation>

<delegation>
	<scout model="Haiku" mode="read-only">Perform one bounded inventory of repository harness files, G0 artifacts, and missing configuration. Return paths, observed facts, and gaps only. Do not edit files or run remote mutations.</scout>
	<advisor model="Opus" condition="only-if-needed">Ask for an opinion only when a genuinely disputed architecture decision is found. Do not invoke advisor for routine inventory, formatting, or already-decided policy.</advisor>
	<rule>Do not use parallel autonomous editors or overlapping implementation agents.</rule>
</delegation>

<g0_review>
	<identify>Open G0 decisions and their current status.</identify>
	<identify>Missing named owners and distinct reviewers.</identify>
	<identify>Exact next human approvals, conditions, and evidence required.</identify>
	<identify>Decisions that are conditionally accepted versus deferred.</identify>
</g0_review>

<git_plan>
	<rule>Use the existing workflow: work branch -&gt; PR develop -&gt; PR main.</rule>
	<rule>develop is staging/preview; main is production.</rule>
	<rule>Preserve the tested commit SHA through promotion.</rule>
	<deliverable>Propose the smallest branch and commit plan for harness-only changes.</deliverable>
</git_plan>

<prohibited_scope>
	<item>Do not create apps/.</item>
	<item>Do not create database migrations or application database code.</item>
	<item>Do not create infra/.</item>
	<item>Do not install dependencies.</item>
	<item>Do not create CI jobs.</item>
	<item>Do not start application implementation until explicit G0 app-implementation approval is recorded.</item>
	<item>Do not deploy, access secrets, or mutate external infrastructure.</item>
</prohibited_scope>

<report>
	<include>Observed facts with paths and commands actually run.</include>
	<include>Missing evidence and unresolved decisions.</include>
	<include>Named owner/reviewer gaps.</include>
	<include>Minimal harness branch/commit plan.</include>
	<include>Actionable stop condition.</include>
	<rule>Keep the report concise.</rule>
	<rule>Use PASS, FAIL, NOT RUN, or BLOCKED only when supported by evidence.</rule>
</report>

<stop_condition>Stop after H0/G0 reporting when harness validation is incomplete, a required human approval is missing, or the task would cross into application implementation.</stop_condition>
```
