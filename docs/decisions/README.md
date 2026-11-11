# 226OS Architecture Decision Registry

**Status:** G0 planning baseline conditionally approved; detailed implementation and release decisions remain gated

The detailed proposal register is [G0_DECISIONS.md](../spec-kit/G0_DECISIONS.md). Entries there are not accepted decisions.

The recorded conditional approval is [ADR-0001](0001-g0-conditional-baseline.md). It authorizes implementation planning only; it does not close execution, security, legal, infrastructure, or release gates.

## Conditionally accepted G0 baseline

The proposed defaults for authentication, authorization, API shape, PostgreSQL, Astryx usage, documentation layout, governance artifacts, and the PostgreSQL-only P0 service baseline are `ACCEPTED WITH CONDITIONS` under ADR-0001. Each implementation slice must still produce its detailed contract and evidence.

`G0-OPS-002` (Datacloud and release readiness) and `G0-COMP-001` (original-code license and CADEV policy) are explicitly `DEFERRED` until their external and legal evidence exists.

## Accepted baseline

| ID | Decision | Source | Status |
|---|---|---|---|
| BASE-001 | React 19 + TypeScript + Vite + TanStack Router/Query/Form | AGENTS.md, DESIGN.md, DEV.md | ACCEPTED |
| BASE-002 | Meta Astryx is the primary UI component foundation; Tailwind is complementary only | AGENTS.md, DESIGN.md | ACCEPTED; release/version validation pending |
| BASE-003 | Graphite & Lime is the approved brand direction | AGENTS.md, DESIGN.md | ACCEPTED; detailed token validation pending |
| BASE-004 | Go modular monolith with Chi, pgx, sqlc, REST/OpenAPI | AGENTS.md, ARCHITECTURE.md, DEV.md | ACCEPTED |
| BASE-005 | PostgreSQL with FTS, GIN, and pg_trgm is the initial data/search system | AGENTS.md, ARCHITECTURE.md, SPEC.md | ACCEPTED |
| BASE-006 | Docker/Compose, Traefik, GitHub Actions, and Make are the delivery baseline | AGENTS.md, DEV.md, SHIP.md | ACCEPTED; artifacts pending |

## Pending human decisions

| ID | Decision request | Required before | Owner |
|---|---|---|---|
| ADR-OPEN-001 | Exact Astryx release, CLI verification, package manager, and lockfile policy | TASK-WEB-001 | OPEN |
| ADR-OPEN-002 | Canonical documentation move from root to `docs/`, or retain current root layout | broad implementation and link normalization | OPEN |
| ADR-OPEN-003 | P0 authentication provider, recovery, verification, session lifecycle, CSRF, and rate limits | TASK-SEC-001 | OPEN |
| ADR-OPEN-004 | OpenAPI verbs/schemas/errors/idempotency, pagination, and DTO generation | TASK-API-001 | OPEN |
| ADR-OPEN-005 | PostgreSQL schema, migration/restore, deletion, FTS strategy, and bilingual relevance policy | TASK-DATA-001 | OPEN |
| ADR-OPEN-010 | Governance document approval, AI provenance, and third-party notice ownership | G4/G7 | OPEN |
| ADR-OPEN-006 | Whether Valkey or S3-compatible storage is needed for P0 | TASK-OPS-001 | OPEN |
| ADR-OPEN-007 | Datacloud product, DNS, budget, registry, and release approver | TASK-OPS-001 / G7 | OPEN |
| ADR-OPEN-008 | Original-code license and CADEV dependency/license policy interpretation | G4/G7 | OPEN |

No pending decision is approved by this registry. A numbered ADR with context, alternatives, decision, migration/security/performance impact, approver, and date is required to close one.
