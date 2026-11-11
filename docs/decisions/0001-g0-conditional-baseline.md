# ADR-0001 — G0 conditional implementation baseline

- **Status:** ACCEPTED WITH CONDITIONS
- **Date:** 2026-11-11
- **Approver:** Yanis Axel DABO (`Y4NN777`), explicit repository-owner approval
- **Scope:** G0 planning baseline for implementation preparation; not a release approval

## Context

The G0 register contains proposed defaults for identity, authorization, API contracts, PostgreSQL, UI foundation, governance, operations, and CADEV compliance. The repository is documentation-only; application code, executable contracts, deployment artifacts, and gate evidence do not yet exist.

## Decision

Approve the proposed defaults in `docs/spec-kit/G0_DECISIONS.md` as the implementation planning baseline, with the following explicit choices:

- `ACCEPT_WITH_CONDITION`: G0-AUTH-001, G0-AUTH-002, G0-AUTH-003, G0-AUTH-004, G0-AUTHZ-001.
- `ACCEPT_WITH_CONDITION`: G0-API-001, G0-API-002.
- `ACCEPT_WITH_CONDITION`: G0-DATA-001, G0-DATA-002, G0-DATA-003.
- `ACCEPT_WITH_CONDITION`: G0-UI-001.
- `ACCEPT_WITH_CONDITION`: G0-GOV-001, G0-GOV-002.
- `ACCEPT_WITH_CONDITION`: G0-OPS-001.
- `DEFER`: G0-OPS-002 until Datacloud product/account, DNS, budget, registry, backup/restore, TLS, and production approver evidence exists.
- `DEFER`: G0-COMP-001 until the original-code license and CADEV dependency/license interpretation receive human/legal review.

This approval authorizes specification closure and implementation planning only. It does not authorize deployment, production release, secrets, migrations against live data, new paid services, or a claim that G1-G7 passed.

## Conditions

1. Assign named product, backend, frontend, security, QA, and release owners before implementation begins.
2. Convert each accepted-with-condition proposal into a detailed ADR or approved contract before its implementation slice changes behavior.
3. Verify the installed Astryx release/API and pin the package manager lockfile before frontend implementation.
4. Approve OpenAPI, DTO/error, database migration, authorization, and search contracts before parallel implementation.
5. Complete AI provenance, third-party notices, SPDX/license review, and CADEV eligibility review before G4/G7.
6. Keep G1-G7 `NOT RUN` until reproducible evidence exists on the target SHA.
7. Revisit deferred operations and licensing decisions before production release.

## Consequences

The team may prepare implementation tasks and RED tests against the stated defaults without reopening the product direction. Security, legal, infrastructure, and release evidence remain explicit blockers; conditional G0 approval cannot be used to bypass them.

## Evidence

- `docs/spec-kit/G0_DECISIONS.md`
- `docs/spec-kit/G0_APPROVAL_TEMPLATE.md`
- `docs/spec-kit/QUALITY_RELEASE.md`
- `docs/SHIP.md`
