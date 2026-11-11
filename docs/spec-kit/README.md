# 226OS Spec Kit

**Status:** implementation-ready documentation baseline, pending human approval
**Date:** 2026-11-11
**Owner:** 7's / human approver to assign

## Purpose

This directory is the canonical implementation handoff for 226OS. It does not replace the product or engineering source documents; it binds them into an executable plan and records what is accepted, open, specified, verified, or not run.

## Document precedence

1. `AGENTS.md` governs agent permissions, evidence, and human approval.
2. `docs/PRD.md` governs product scope and rationale, including all eight domains and four differentiators.
3. `docs/SPEC.md` governs observable behavior, invariants, and API boundaries.
4. `docs/architecture/ARCHITECTURE.md` governs accepted technical boundaries.
5. `DESIGN.md` governs UI direction and Astryx usage.
6. `docs/RULES.md`, `docs/TEST.md`, `docs/DEV.md`, and `docs/SHIP.md` govern engineering, quality, implementation, and release procedure.
7. `docs/decisions/` records accepted or pending changes; a later accepted ADR supersedes an earlier decision only when explicitly stated.
8. This Spec Kit binds the above documents into tasks. It cannot approve an open decision or change product scope.

The repository keeps the primary product and engineering documents under `docs/`; root-level `README.md`, `DESIGN.md`, and governance files remain entry points and control documents.

## Status vocabulary

- `SPECIFIED`: documented requirement or procedure; no execution evidence yet.
- `VERIFIED`: inspected evidence exists in the repository or a recorded review.
- `PASS`: an executable gate produced acceptable evidence for the target SHA.
- `FAIL`: an executable gate produced unacceptable evidence.
- `NOT RUN`: no executable check was run; never treat it as success.
- `BLOCKED`: cannot proceed until an explicit dependency or human decision is resolved.

## Contents

- [Traceability](TRACEABILITY.md): PRD, SPEC, UI/API/data, BDD, tasks, and gates.
- [Contracts](CONTRACTS.md): API, data, search, frontend, security, and operational contracts to approve before parallel implementation.
- [G0 decisions](G0_DECISIONS.md): proposed defaults and concise human decision register.
- [Authentication and authorization](AUTHORIZATION_MATRIX.md): proposed session policy and endpoint permission matrix.
- [API contract](API_CONTRACT.md): proposed REST, error, pagination, and idempotency contract.
- [Database design](DATABASE_DESIGN.md): proposed schema, migration, seed, and search design.
- [Astryx verification](ASTRYX_VERIFICATION.md): release/API verification protocol; no package is installed by this task.
- [G0 approval template](G0_APPROVAL_TEMPLATE.md): human sign-off record and checklist.
- [Tasks](TASKS.md): bounded implementation work items with owners, allowed paths, tests, evidence, and dependencies.
- [Quality and release](QUALITY_RELEASE.md): deterministic gates, evidence records, and release prerequisites.
- [Decisions](../decisions/README.md): accepted baseline and pending human decisions.
- [Claude handoff](CLAUDE_HANDOFF.md): standalone first-session instructions.

## Current readiness

Documentation review is `VERIFIED` for the findings recorded in the handoff. G0 is `CONDITIONAL GO` for specification planning, with implementation still blocked pending named owners, ADRs, evidence plans, governance/legal review, and the conditions in `G0_APPROVAL_TEMPLATE.md`. G1-G7 are `NOT RUN` because no application, infrastructure, CI, or test artifacts exist yet.
