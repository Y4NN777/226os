# 226os Spec Kit

**Status:** G0 closed (see [docs/decisions/README.md](../decisions/README.md)); the application is under implementation.
**Date:** 2026-11-12
**Owner:** Y4NN777

## Purpose

This directory binds the product and engineering documents into implementation contracts. It does not replace them and cannot approve an open decision or change product scope.

## Document precedence

1. `AGENTS.md`: agent permissions, evidence, human approval.
2. `docs/PRD.md`: product scope and rationale.
3. `docs/SPEC.md`: observable behavior, invariants, API boundaries.
4. `docs/architecture/ARCHITECTURE.md`: accepted technical boundaries.
5. `DESIGN.md`: UI direction and Astryx usage.
6. `docs/RULES.md`, `docs/TEST.md`, `docs/DEV.md`, `docs/SHIP.md`: engineering, quality, implementation and release procedure.
7. `docs/decisions/`: ADRs; a later accepted ADR supersedes an earlier one only when it says so.
8. This Spec Kit.

## Status vocabulary

`SPECIFIED` (documented, no execution evidence), `VERIFIED` (inspected evidence), `PASS`, `FAIL`, `NOT RUN` (never a success), `BLOCKED`. Gate statuses come only from commands that were run.

## Contents

- [Traceability](TRACEABILITY.md): PRD, SPEC, UI/API/data, BDD and gates.
- [Authentication and authorization](AUTHORIZATION_MATRIX.md): endpoint permission matrix.
- [API contract](API_CONTRACT.md): accepted REST, error and pagination decisions.
- [Database design](DATABASE_DESIGN.md): schema rationale, invariants, search and seed design.
- [Astryx verification](ASTRYX_VERIFICATION.md): verification protocol and criteria.
- [Quality and release](QUALITY_RELEASE.md): gate status and evidence record.
- Archive (historical, not maintained): [Tasks](../archive/spec-kit/TASKS.md), [G0 decisions](../archive/spec-kit/G0_DECISIONS.md), [G0 approval record](../archive/spec-kit/G0_APPROVAL_RECORD.md), [Claude handoff](../archive/spec-kit/CLAUDE_HANDOFF.md), [Contracts](../archive/spec-kit/CONTRACTS.md).

## Contract index

Detail lives in `api/openapi.yaml`, `db/migrations/`, the documents above and ADR-0003 to ADR-0007. Contracts not stated elsewhere:

- Search: questions, resources, projects and Learn articles; bounded, stable ordering; country/tag/type/language filters; PostgreSQL full-text (`simple` plus immutable `unaccent` wrapper), GIN and `pg_trgm`. Cost is bounded and plans are captured with `EXPLAIN (ANALYZE, BUFFERS)`. No semantic search or AI claims at P0; relevance dataset and latency budgets are G6 evidence.
- Frontend: React 19, TypeScript strict, Vite, TanStack Router/Query/Form, Astryx as the primary component foundation. Public reads need no session; every mutation does. Shared UI covers loading, empty, error, success, focus, keyboard and dialog states. The FR/EN choice preserves original user content and unsaved form state. Responsive checkpoints: 360, 768, 1024, 1440 px and 200% zoom.
- Security evidence: negative authorization, CSRF/session, XSS/SSRF and rate-limit tests, SAST, SCA, secret and licence scans. No code execution from snippets or user Markdown.
- Operations: Docker multi-stage images; one profile-based Compose tree (`local-dev`, `staging-test`, `production-live`) behind Traefik; PostgreSQL stays private; health/readiness endpoints; structured logs without secrets; encrypted backup/restore evidence. Valkey and S3-compatible storage are excluded at P0 (new ADR needed). Make and GitHub Actions are the only gate interfaces.
- Seeds are deterministic, synthetic, and never presented as real activity or provider verification.
