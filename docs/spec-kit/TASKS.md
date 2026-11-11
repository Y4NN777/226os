# 226OS Implementation Tasks

**Status:** PLANNED; owners and reviewers require human assignment

Every task follows the governance work-item contract. No task is `DONE` until its evidence paths contain real results for the target SHA.

## TASK-SPEC-001 — Close product and contract gaps

```yaml
goal: "Approve the P0 slice, route boundaries, open API/data decisions, and evidence ownership."
story_id: "PRD-TRACEABILITY"
priority: "P0"
owner: "OPEN"
reviewer: "OPEN"
allowed_paths: ["docs/spec-kit/**", "SPEC.md", "PRD.md", "docs/decisions/**"]
acceptance_criteria: ["All six PRD journeys have nominal and relevant negative cases", "All eight domains and four differentiators remain traceable", "Open decisions have an owner and due decision"]
threats_and_risks: ["scope drift", "false readiness claim"]
planned_tests: ["documentation traceability review", "link and orphan requirement review"]
required_gates: [G0]
evidence_paths: ["docs/spec-kit/TRACEABILITY.md", "docs/spec-kit/CLAUDE_HANDOFF.md"]
status: PLANNED
```

Dependencies: human checkpoint.

## TASK-DATA-001 — PostgreSQL schema, migrations, seeds, and search

```yaml
goal: "Implement reviewed P0 entities, invariants, migrations, deterministic seeds, and bounded PostgreSQL search."
story_id: "US-SOLVE-001 / US-DISCOVER-001 / US-BUILD-001"
priority: "P0"
owner: "OPEN"
reviewer: "OPEN"
allowed_paths: ["db/**", "apps/api/internal/**", "docs/decisions/**"]
acceptance_criteria: ["All listed invariants are enforced in the database", "Fresh and upgrade migration paths are tested", "Search filters and stable ordering work on representative data"]
threats_and_risks: ["data corruption", "expensive search", "seed data misrepresented as real"]
planned_tests: ["migration tests", "constraint tests", "EXPLAIN ANALYZE evidence", "real PostgreSQL integration"]
required_gates: [G0,G1,G2,G3,G4,G6]
evidence_paths: ["db/migrations/**", "docs/evidence/TASK-DATA-001/**"]
status: PLANNED
```

Dependencies: TASK-SPEC-001 and approved data/search contract.

## TASK-SEC-001 — Sessions, authentication, and authorization

```yaml
goal: "Implement secure session identity and server-side ownership/collaborator authorization for P0 routes."
story_id: "US-IDENTITY-001 / US-SOLVE-001 / US-BUILD-001"
priority: "P0"
owner: "OPEN"
reviewer: "OPEN"
allowed_paths: ["apps/api/**", "apps/web/src/features/identity/**", "api/openapi.yaml", "docs/spec-kit/**"]
acceptance_criteria: ["Anonymous, owner, collaborator, and third-party access cases are explicit", "CSRF/session controls are tested", "Sensitive failures reveal no unnecessary data"]
threats_and_risks: ["IDOR", "session theft", "CSRF", "privacy leakage"]
planned_tests: ["negative authorization matrix", "session expiry/revocation", "CSRF", "rate-limit tests"]
required_gates: [G0,G1,G2,G3,G4,G5]
evidence_paths: ["docs/evidence/TASK-SEC-001/**"]
status: PLANNED
```

Dependencies: TASK-SPEC-001 and approved auth decisions.

## TASK-API-001 — OpenAPI, Go modules, and P0 HTTP vertical slices

```yaml
goal: "Create the modular Go API and implement P0 endpoints strictly from the approved OpenAPI contract."
story_id: "US-SOLVE-001 / US-DISCOVER-001 / US-BUILD-001"
priority: "P0"
owner: "OPEN"
reviewer: "OPEN"
allowed_paths: ["apps/api/**", "api/openapi.yaml", "db/queries/**"]
acceptance_criteria: ["Handlers contain transport concerns only", "Business invariants live in application/domain boundaries", "Generated client/API contract has no drift"]
threats_and_risks: ["contract drift", "authorization bypass", "unbounded request cost"]
planned_tests: ["RED/GREEN unit tests", "HTTP integration", "OpenAPI validation/generation diff", "race tests"]
required_gates: [G0,G1,G2,G3,G4]
evidence_paths: ["api/openapi.yaml", "docs/evidence/TASK-API-001/**"]
status: PLANNED
```

Dependencies: TASK-DATA-001, TASK-SEC-001, and approved API contract.

## TASK-WEB-001 — Astryx shell and P0 frontend flows

```yaml
goal: "Build accessible FR/EN Explore, Solve, Discover, Build, identity, and relation flows using the pinned Astryx foundation."
story_id: "US-SOLVE-001 / US-DISCOVER-001 / US-BUILD-001 / US-IDENTITY-001 / US-ACCESS-001"
priority: "P0"
owner: "OPEN"
reviewer: "OPEN"
allowed_paths: ["apps/web/**", "api/openapi.yaml", "docs/spec-kit/**"]
acceptance_criteria: ["Core routes work at required responsive checkpoints", "Keyboard/focus/loading/error/empty states are covered", "FR/EN preserves original content and form edits"]
threats_and_risks: ["accessibility failure", "contract drift", "false UI state", "low-bandwidth regression"]
planned_tests: ["Vitest/Testing Library", "Playwright P0 flows", "axe/contrast", "throttled network", "frontend coverage"]
required_gates: [G0,G1,G2,G3,G5,G6]
evidence_paths: ["docs/evidence/TASK-WEB-001/**"]
status: PLANNED
```

Dependencies: TASK-API-001, TASK-SEC-001, approved Astryx release and route/locale decisions.

## TASK-OPS-001 — Make, CI, Compose, Traefik, and release evidence

```yaml
goal: "Provide reproducible local and CI gates plus a staging deployment and rollback-ready release path."
story_id: "PRD-CADEV-DELIVERY"
priority: "P0"
owner: "OPEN"
reviewer: "OPEN"
allowed_paths: ["Makefile", ".github/workflows/**", "infra/**", "docs/quality/**", "SHIP.md"]
acceptance_criteria: ["G1-G7 commands are explicit and fail closed", "DB is private and edge routing is TLS-protected", "Backup/restore and rollback are tested before release"]
threats_and_risks: ["secret exposure", "data loss", "unreviewed deployment", "non-reproducible gate"]
planned_tests: ["make doctor/quality/build", "Compose validation", "security/license scans", "staging smoke", "backup restore and rollback drill"]
required_gates: [G0,G1,G2,G3,G4,G5,G6,G7]
evidence_paths: ["docs/evidence/TASK-OPS-001/**"]
status: PLANNED
```

Dependencies: all previous tasks; human release approval; Datacloud readiness.
