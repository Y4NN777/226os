# ADR-0004 — API contract conventions

- **Status:** ACCEPTED
- **Date:** 2026-11-11
- **Approver:** `Y4NN777`
- **Closes:** ADR-OPEN-004
- **Refines:** G0-API-001, G0-API-002
- **Related:** `docs/spec-kit/API_CONTRACT.md`, `api/openapi.yaml`, [ADR-0003](0003-authentication-sessions-csrf-rate-limits.md), [ADR-0006](0006-toolchain-environments-quality-commands.md)
- **Metadata note 2026-11-12:** approver spelling normalised to the GitHub login (`Y4NN777`); the decision text is unchanged.

## Context

TASK-API-001 needs a reviewed contract before parallel Go and web work. `api/openapi.yaml` is the single source of truth; DTOs are never duplicated by hand (`docs/DEV.md` section 5). This ADR fixes how the contract is authored, enforced and consumed.

## Options considered

- **Design-first OpenAPI 3.1 (chosen)** vs code-first annotations: design-first lets frontend and backend start in parallel and makes review of the contract the checkpoint.
- **REST resources plus action endpoints (chosen)** vs RPC style: action endpoints (`accept`, `votes`, session operations) keep authorization and audit semantics explicit.
- **Go generation:** oapi-codegen (chosen), ogen, hand-written. oapi-codegen generates Chi-compatible server interfaces and types; its OpenAPI 3.1 support is partial. A **verification step at bootstrap** (run the generator on `api/openapi.yaml` and a spec linter) decides; **ogen is the documented fallback** if 3.1 support is insufficient. Switching to the fallback is recorded in the bootstrap PR and does not need a new ADR.
- **TypeScript:** openapi-typescript (chosen) with a typed fetch client over hand-written DTOs.

## Decision

1. **Format and location.** OpenAPI 3.1, design-first, **`api/openapi.yaml` is the source of truth**. A pinned OpenAPI linter runs in `make lint`.
2. **Base path and versioning.** Paths under `/api/v1`. Additive changes stay in v1; breaking changes need `/api/v2` and an ADR. Clients ignore unknown response fields.
3. **Resources and actions.** The path inventory is the one in `api/openapi.yaml` (`/auth/*`, `/me`, `/me/profile`, `/profiles/{username}`, questions, answers, `/votes`, `accept`/`unaccept`, resources, projects, contribution needs, applications, `/learn/*`, `/search`, `/moderation/*`). `PATCH` uses JSON Merge Patch (RFC 7396).
4. **Errors.** `application/problem+json` per **RFC 9457** with `type`, `title`, `status`, `detail`, `instance`, `code`, `request_id`, `violations[]{field, code}`. **`type` is `urn:226os:problem:<code>` (for example `urn:226os:problem:VALIDATION_ERROR`), a non-dereferenceable identifier at P0.** The error catalogue is the closed `ErrorCode` enum with `x-error-catalogue` in `api/openapi.yaml` (for example `MALFORMED_REQUEST` 400, `INVALID_CURSOR` 400, `UNAUTHENTICATED` 401, `CSRF_FAILED` 403, `FORBIDDEN` 403, `EMAIL_NOT_VERIFIED` 403, `SELF_VOTE_FORBIDDEN` 403, `REAUTH_REQUIRED` 403, `NOT_FOUND` 404, `STATE_CONFLICT` 409, `PRECONDITION_FAILED` 412, `PAYLOAD_TOO_LARGE` 413, `VALIDATION_ERROR` 422, `PRECONDITION_REQUIRED` 428, `RATE_LIMITED` 429 with `Retry-After`, `INTERNAL_ERROR` 500, `SERVICE_UNAVAILABLE` 503). The YAML wins over this list. Never return stack traces, SQL, tokens, internal hosts, or private-object existence. A Go constants file is generated or checked against the enum.
5. **Pagination.** Cursor only, no offset. `limit` **default 20, maximum 50; a value above 50 returns 422**, never a silent clamp. The cursor is opaque base64url, HMAC-signed with a server key, and encodes the sort key, a UUID tie-breaker and a hash of the filters; a tampered, expired or mismatched cursor returns 400 `INVALID_CURSOR`. Envelope `{ "items": [], "next_cursor": null | "..." }`. `/search` is the documented exception (grouped, bounded `per_group`).
6. **Idempotency.** `Idempotency-Key` header, **24-hour retention**, adopted as **our own convention** (the IETF httpapi draft is an Internet-Draft, not an RFC; we take its ideas and pin the behaviour in our spec). Bound to subject, route and request hash. Same key and hash replays the stored response; same key and different hash returns 422 `IDEMPOTENCY_KEY_REUSED`; a concurrent duplicate returns 409 `IDEMPOTENCY_IN_PROGRESS`. Database uniqueness remains the real safety net.
7. **Concurrency.** **`ETag` / `If-Match` on questions, answers and Learn articles** (412 `PRECONDITION_FAILED` on mismatch, 428 `PRECONDITION_REQUIRED` when the header is missing). Only the question `PATCH` exists in the P0 contract; answer and article edits follow the same convention when added. Other resources are explicitly last-write-wins.
8. **Validation.** Requests are schema-validated before authorization-sensitive work with the generated spec. Mutation bodies set `additionalProperties: false`, with `maxLength`, `maxItems`, `format` and `enum` everywhere; Go decoding also uses `DisallowUnknownFields`. Body limit 1 MiB (413). `Content-Type` must be `application/json` on mutations. Responses are validated against the spec in contract tests.
9. **Auth surface in the spec.** Session cookie as `cookieAuth` (`__Host-226os_session` deployed, `226os_session` local-dev, ADR-0003); `X-CSRF-Token` as `csrfToken` required on unsafe methods; 401/403 declared on every protected operation.
10. **Generation and drift.** `make generate` runs oapi-codegen (Go server interface and types under a generated package in `src/api/internal/`) and openapi-typescript (web types). Generated files are committed and never hand-edited. CI runs **`make generate && git diff --exit-code`**; any diff fails G1. Tool versions are pinned.

## Consequences

- **Security:** uniform redacted errors, validated inputs and explicit 401/403/404 semantics are contract-tested. Idempotency keys are subject-bound so they cannot become a cross-user replay vector.
- **Data:** idempotency records need a table with TTL cleanup; cursors depend on stable sort columns and UUID tie-breakers (ADR-0005).
- **Ops:** the generation toolchain is part of CI; the cursor-signing key is a secret outside Git.
- **Cost:** none beyond CI time. A generator limitation may force an ogen switch or a spec subset, recorded in the bootstrap PR.

## Evidence / verification required

OpenAPI lint pass; the generator accepts the spec (oapi-codegen, else ogen); `make generate && git diff --exit-code` clean; contract fixtures for each catalogue entry; tests for malformed JSON, unknown fields, oversized bodies, invalid and tampered cursors, limit bounds (including 422 above 50), idempotency same/different hash, ETag mismatch and missing header; redaction tests. Gates G1 and G3.
