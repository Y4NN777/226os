# ADR-0007 — Astryx version pinning and verification

- **Status:** ACCEPTED
- **Date:** 2026-11-11
- **Approver:** `Y4NN77`
- **Closes:** ADR-OPEN-001
- **Refines:** G0-UI-001
- **Related:** `DESIGN.md`, `docs/spec-kit/ASTRYX_VERIFICATION.md`, [ADR-0006](0006-toolchain-environments-quality-commands.md)

## Context

Meta Astryx is the approved primary component foundation (BASE-002). It is a young, beta-stage package, and the repository has never installed it, so release numbers, CLI commands and component APIs written in `DESIGN.md` are illustrative until checked against the installed package.

## Decision

1. **Pin the exact `@astryxdesign/core` version installed at bootstrap** (no range, no `latest`), recorded in the bootstrap PR and held by the committed pnpm lockfile (ADR-0006).
2. **Verify the installed API via its own CLI and documentation before using it.** Run the installed package's help/docs commands, inspect its exported components and types, and use only what that installed release documents. Illustrative snippets in `DESIGN.md` are not normative.
3. **Record capability gaps** for any P0 need the package does not cover (component, accessibility behaviour, theme token). A custom component is allowed only against a recorded gap, with an accessible implementation plan and provenance entry (`THIRD_PARTY_NOTICES.md`, `AI_USAGE.md`).
4. Tailwind stays complementary (layout) only. Replacing Astryx needs a new ADR.
5. An Astryx upgrade is a separate, reviewed change that re-runs the verification.

The verification sequence and decision criteria are in `docs/spec-kit/ASTRYX_VERIFICATION.md`; running it is bootstrap/TASK-WEB-001 evidence, not a pending decision.

## Consequences

- **Security/supply chain:** a pinned, lockfile-backed package with a recorded licence and transitive inventory.
- **Cost:** a small verification effort at bootstrap; a gap record may add custom components later.
- **Risk:** beta API drift is contained by the pin; mitigation is the gap record plus the verification re-run on upgrade.

## Evidence / verification required

Exact version and lockfile integrity recorded; installed CLI/docs output captured; the three representative flows (Solve question detail, Discover API detail, Build contribution request) rendered in light, dark and system modes at 360, 768, 1024 and 1440 px and at 200% zoom; keyboard and axe results; licence inventory; capability-gap record. Gates G1, G5, G6.
