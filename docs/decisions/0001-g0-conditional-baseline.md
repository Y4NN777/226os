# ADR-0001 — G0 implementation baseline

- **Status:** ACCEPTED
- **Date:** 2026-11-11
- **Approver:** `Y4NN777`, repository owner
- **Scope:** G0 decisions for implementation preparation and the approval of P0 implementation under ADR-0001 to ADR-0007; not a release approval
- **Amended:** 2026-11-11, same approver. The owner decided that no G0 decision may remain proposed or conditional. This ADR was first recorded as a conditional baseline (hence the file name, kept so links stay stable). It now records each original condition as met and how, and records the decisions that were deferred.
- **Metadata note 2026-11-12:** approver spelling normalised to the GitHub login (`Y4NN777`); the decision text is unchanged.

## Context

The G0 register contained proposed defaults for identity, authorization, API contracts, PostgreSQL, UI foundation, governance, operations, and licensing. At the first recording (2026-11-11) the repository was documentation-only. Application code, executable contracts, deployment artifacts and gate evidence did not exist yet; G1 to G7 are evidence, not decisions, and were all `NOT RUN` at that date. Current gate status is in [the decisions README](README.md).

## Decision

All rows of `docs/archive/spec-kit/G0_DECISIONS.md` (archived 2026-11-12) are `ACCEPTED`:

| Decisions | Recorded in |
|---|---|
| G0-AUTH-001, G0-AUTH-002, G0-AUTH-003, G0-AUTH-004, G0-AUTHZ-001 | [ADR-0003](0003-authentication-sessions-csrf-rate-limits.md) |
| G0-API-001, G0-API-002 | [ADR-0004](0004-api-contract-conventions.md) |
| G0-DATA-001, G0-DATA-002, G0-DATA-003 | [ADR-0005](0005-data-model-migrations-search.md) |
| G0-OPS-001 | [ADR-0006](0006-toolchain-environments-quality-commands.md) |
| G0-UI-001 | [ADR-0007](0007-astryx-version-pinning-and-verification.md) |
| G0-GOV-001 (documentation layout) | this ADR: the current layout is accepted. Root `README.md`, `DESIGN.md`, `AGENTS.md`, `CLAUDE.md` and governance files stay at the root; engineering documents (PRD, SPEC, DEV, TEST, SHIP, RULES, architecture, research, quality, decisions, spec-kit) stay under `docs/`. Closes ADR-OPEN-002. |
| G0-GOV-002 (governance artifacts) | this ADR: `CONTRIBUTING.md`, `AI_USAGE.md` and `THIRD_PARTY_NOTICES.md` are in force as project policy. Their project-specific completion (dependency inventory, provenance entries) is G4/G7 evidence work. Closes ADR-OPEN-010. |
| G0-OPS-002 (hosting) | this ADR plus ADR-0006: **Datacloud is the host for `staging-test` and `production-live`.** Actual deployment still needs account/product, DNS, budget and approver evidence; that is execution evidence for G7, not an open decision. Closes ADR-OPEN-007. |
| G0-COMP-001 (licence) | this ADR: **Apache-2.0** for original 226os code (`LICENSE` and `NOTICE` at the repository root). Third-party licences stay separate and are tracked in `THIRD_PARTY_NOTICES.md`. Closes ADR-OPEN-008. |
| Interim ownership | [ADR-0002](0002-interim-single-maintainer-ownership.md) |

Together with the baseline already recorded in `docs/decisions/README.md` (BASE-001 to BASE-006) this authorizes **implementation of P0** (bootstrap task B0 first) under ADR-0001 to ADR-0007. It does not authorize deployment, production release, secrets, migrations against live data, new paid services, or any claim that G1 to G7 passed.

## Original conditions and how each is met

| # | Condition (as first recorded) | Status | How |
|---|---|---|---|
| 1 | Name product, backend, frontend, security, QA, and release owners before implementation | Met | [ADR-0002](0002-interim-single-maintainer-ownership.md): `Y4NN777` is the interim owner of every role until P0 completes (amended by [ADR-0008](0008-team-work-areas-and-ownership.md): work areas for the four-person team). |
| 2 | Convert each proposal into a detailed ADR or approved contract before its slice changes behavior | Met | ADR-0003 (auth), ADR-0004 (API), ADR-0005 (data), ADR-0006 (toolchain), ADR-0007 (Astryx); `docs/spec-kit/AUTHORIZATION_MATRIX.md`, `API_CONTRACT.md`, `DATABASE_DESIGN.md`. |
| 3 | Verify the installed Astryx release/API and pin the package-manager lockfile before frontend implementation | Met as a decision | ADR-0006 fixes pnpm via `packageManager` and a committed lockfile; ADR-0007 fixes the Astryx pin-and-verify procedure. Running it is task B0/TASK-WEB-001 evidence. |
| 4 | Approve OpenAPI, DTO/error, migration, authorization, and search contracts before parallel implementation | Met | ADR-0004 and `api/openapi.yaml`; ADR-0005 and `DATABASE_DESIGN.md`; ADR-0003 and `AUTHORIZATION_MATRIX.md`. Executable verification (generator run, linter) happens at bootstrap. |
| 5 | Complete AI provenance, third-party notices, SPDX/licence review before G4/G7 | Decided; evidence at G4/G7 | G0-GOV-002 and G0-COMP-001 above. The dependency inventory and licence report are produced on the target SHA at G4/G7. |
| 6 | Keep G1 to G7 `NOT RUN` until reproducible evidence exists on the target SHA | Standing rule, met | At the first recording (2026-11-11) every gate was `NOT RUN`; see the decisions README for the current status. |
| 7 | Revisit deferred operations and licensing decisions before production release | Met | G0-OPS-002 and G0-COMP-001 are decided above. Datacloud account, DNS, budget and approver evidence remain G7 evidence. |

## Consequences

- The team may start B0 bootstrap and the P0 slices against ADR-0001 to ADR-0007 without reopening product direction.
- Security, legal, infrastructure and release evidence remain explicit gates; this approval cannot be used to bypass them (ADR-0002 does not waive G4/G7 reviews either).
- A change to any accepted decision needs a new or superseding ADR.

## Evidence

- `docs/archive/spec-kit/G0_DECISIONS.md` (archived 2026-11-12)
- `docs/spec-kit/G0_APPROVAL_TEMPLATE.md`
- `docs/spec-kit/QUALITY_RELEASE.md`
- `docs/SHIP.md`
