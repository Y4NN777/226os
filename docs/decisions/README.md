# 226os Architecture Decision Registry

**Status:** G0 is closed (2026-11-11): every G0 decision is `ACCEPTED` (approver `Y4NN777`), and P0 implementation is authorized under ADR-0001 to ADR-0010. ADR-0008 to ADR-0011 are dated 2026-11-12. Gates are evidence, not decisions: G1 to G4 run in CI through `make quality`; G5, G6 and G7 are `NOT RUN`.

The (archived) register of G0 decisions is [G0_DECISIONS.md](../archive/spec-kit/G0_DECISIONS.md); each row points to the ADR that records it.

## ADR registry

| ADR | Title | Status | Date |
|---|---|---|---|
| [0001](0001-g0-conditional-baseline.md) | G0 implementation baseline (original conditions recorded as met; hosting, licence, layout and governance decisions) | ACCEPTED | 2026-11-11 |
| [0002](0002-interim-single-maintainer-ownership.md) | Interim single-maintainer ownership (amended by ADR-0008) | ACCEPTED | 2026-11-11 |
| [0003](0003-authentication-sessions-csrf-rate-limits.md) | Authentication, sessions, CSRF, rate limits | ACCEPTED | 2026-11-11 |
| [0004](0004-api-contract-conventions.md) | API contract conventions | ACCEPTED | 2026-11-11 |
| [0005](0005-data-model-migrations-search.md) | Data model, migrations, search | ACCEPTED | 2026-11-11 |
| [0006](0006-toolchain-environments-quality-commands.md) | Toolchain, environments, quality commands | ACCEPTED | 2026-11-11 |
| [0007](0007-astryx-version-pinning-and-verification.md) | Astryx version pinning and verification | ACCEPTED | 2026-11-11 |
| [0008](0008-team-work-areas-and-ownership.md) | Team work areas and ownership for P0 (amends ADR-0002) | ACCEPTED | 2026-11-12 |
| [0009](0009-secrets-with-sops-and-age.md) | Secrets with SOPS and age (amends the "secrets never in Git" rule of ADR-0006) | ACCEPTED | 2026-11-12 |
| [0010](0010-ci-parallel-gate-groups.md) | CI runs the quality gates as parallel groups (amends the single-job CI of ADR-0006) | ACCEPTED | 2026-11-12 |
| [0011](0011-ai-assistant-provider.md) | AI assistant provider (Google AI Studio, Gemma 4) | ACCEPTED | 2026-11-12 |
| [0012](0012-ai-assistant-integration.md) | AI assistant integration (grounded answers, limits, fallback, golden set) | ACCEPTED | 2026-11-14 |

## Accepted baseline

| ID | Decision | Source | Status |
|---|---|---|---|
| BASE-001 | React 19 + TypeScript + Vite + TanStack Router/Query/Form | AGENTS.md, DESIGN.md, DEV.md | ACCEPTED |
| BASE-002 | Meta Astryx is the primary UI component foundation; Tailwind is complementary only | AGENTS.md, DESIGN.md, ADR-0007 | ACCEPTED |
| BASE-003 | Graphite & Lime is the approved brand direction | AGENTS.md, DESIGN.md | ACCEPTED |
| BASE-004 | Go modular monolith with Chi, pgx, sqlc, REST/OpenAPI | AGENTS.md, ARCHITECTURE.md, DEV.md, ADR-0004 | ACCEPTED |
| BASE-005 | PostgreSQL (full-text search, GIN, pg_trgm, unaccent) is the initial data and search system | AGENTS.md, ARCHITECTURE.md, SPEC.md, ADR-0005 | ACCEPTED |
| BASE-006 | Docker/Compose, Traefik, GitHub Actions, and Make are the delivery baseline | AGENTS.md, DEV.md, SHIP.md, ADR-0006 | ACCEPTED |

## Decision requests (all decided)

| ID | Decision request | Decided in | Status |
|---|---|---|---|
| ADR-OPEN-001 | Exact Astryx release, CLI verification, package manager, and lockfile policy | ADR-0007 (Astryx pin and verify), ADR-0006 (pnpm, lockfile) | ACCEPTED |
| ADR-OPEN-002 | Canonical documentation layout | ADR-0001 (current layout retained) | ACCEPTED |
| ADR-OPEN-003 | Authentication, recovery, verification, session lifecycle, CSRF, rate limits | ADR-0003 | ACCEPTED |
| ADR-OPEN-004 | OpenAPI verbs/schemas/errors/idempotency, pagination, DTO generation | ADR-0004 | ACCEPTED |
| ADR-OPEN-005 | PostgreSQL schema, migration/restore, deletion, search, bilingual policy | ADR-0005 | ACCEPTED |
| ADR-OPEN-006 | Valkey or S3-compatible storage for P0 | ADR-0006 (excluded at P0) | ACCEPTED |
| ADR-OPEN-007 | Hosting | ADR-0006 and ADR-0001 (Datacloud hosts `staging-test` and `production-live`) | ACCEPTED |
| ADR-OPEN-008 | Original-code licence | ADR-0001 (Apache-2.0; `LICENSE`, `NOTICE`) | ACCEPTED |
| ADR-OPEN-010 | Governance documents, AI provenance, third-party notices | ADR-0001 (policies in force) | ACCEPTED |

The owner roles in ADR-0002 are interim until P0 completes and were amended by ADR-0008 (work areas for the four-person team). The id ADR-OPEN-009 was never used (no trace in the Git history), hence the gap in the table above. Items that remain are execution evidence (Datacloud account/product, DNS, budget and approver evidence for G7; the dependency and licence inventory at G4/G7; the exact tool versions recorded at bootstrap), not open decisions. A change to an accepted decision needs a new or superseding ADR with context, alternatives, decision, migration/security/performance impact, approver and date.
