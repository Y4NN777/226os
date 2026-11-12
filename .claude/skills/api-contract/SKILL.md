---
name: api-contract
description: 226os design-first REST contract (OpenAPI 3.1 in api/openapi.yaml, RFC 9457 problem+json, REST verbs + action endpoints, opaque cursor pagination, Idempotency-Key, codegen and drift checks). Use when touching api/openapi.yaml, api/**, generated server/client code, or HTTP status/error mapping.
---
# API Contract (OpenAPI 3.1, design-first)

## Scope and triggers
- `api/openapi.yaml`, `api/**`, generated Go server types / TS client, `src/api/internal/*/transport/**` status/error mapping, `src/web` API client usage.
- Status: `api/openapi.yaml` exists and is the source of truth; ADR-0004 (`docs/decisions/0004-api-contract-conventions.md`) is ACCEPTED and fixes the conventions below. `docs/spec-kit/API_CONTRACT.md` is the narrative companion. A change to `api/openapi.yaml` needs the backend lead and the frontend owner (ADR-0008).

## Project rules (binding)
- OpenAPI 3.1 is the shared contract; clients generated or automatically validated; no drift tolerated (`docs/RULES.md:15`, `docs/DEV.md:7`).
- Verbs and schemas frozen in `api/openapi.yaml` before parallel development (`docs/SPEC.md:102`); spec changes precede behaviour changes (`docs/SPEC.md:8`, `docs/RULES.md:7`). Agents never invent endpoints or divergent DTOs (`docs/RULES.md:13`, `docs/DEV.md:51`).
- Prefix `/api/v1`, JSON UTF-8, RFC3339 UTC timestamps, opaque UUIDs, errors `application/problem+json` RFC 9457 (`docs/SPEC.md:99`, `docs/spec-kit/API_CONTRACT.md`).
- Initial resources: sessions, users/me, profiles, questions, answers, votes, resources, projects, contribution-needs, applications, search (`docs/SPEC.md:101`). The path inventory is the one in `api/openapi.yaml` (ADR-0004).
- Conventional REST: collection `GET/POST`, member `GET/PATCH/DELETE` only where behaviour exists; action endpoints for accept, vote, interest, session ops (G0-API-001 `docs/archive/spec-kit/G0_DECISIONS.md:14`).
- Cursor pagination: `limit` default 20, max 50 (above 50 returns 422, never a silent clamp), cursors opaque and HMAC-signed, deterministic order + UUID tie-breaker (ADR-0004, section Pagination; `docs/SPEC.md:100`).
- Bodies and query filters schema-validated before authorization-sensitive work; unknown fields rejected on mutations unless marked extensible; body size limit (`docs/spec-kit/API_CONTRACT.md`, `docs/SPEC.md:100`).
- Browser mutations: session cookie + CSRF token; no bearer tokens in P0 (`docs/spec-kit/API_CONTRACT.md`).
- `Idempotency-Key` on retried writes that can duplicate (applications, votes, relation creation, selected publication); key bound to subject, route, request hash, bounded retention (`docs/spec-kit/API_CONTRACT.md`). Retention is 24 hours (ADR-0004).
- `ETag`/`If-Match` on questions, answers and Learn articles (412 on mismatch, 428 when missing; ADR-0004); last-write-wins is NOT an implicit contract (`docs/spec-kit/API_CONTRACT.md`).
- Problem fields: stable `type`, `title`, `status`, `code`; human-safe `detail`; `instance`; opaque `request_id`; `violations[{field,code}]` (`docs/spec-kit/API_CONTRACT.md`; G0-API-002 `docs/archive/spec-kit/G0_DECISIONS.md:15`).
- Never expose stack traces, SQL, credentials, tokens, internal hostnames, private-object existence (`docs/spec-kit/API_CONTRACT.md`).
- Status map: 400 malformed, 401 no/invalid session, 403 forbidden, 404 absent or intentionally undisclosed, 409 conflict, 413 too large, 422 semantic validation, 429 rate limit, 500, 503; `Retry-After` on 429/503; bounded error bodies (`docs/spec-kit/API_CONTRACT.md`).
- Required contract tests: CSRF rejection, session rotation/revocation, malformed JSON, unknown fields, oversized bodies, invalid cursors, replay, duplicate idempotency key same/different hash, redaction, OpenAPI lint, examples, generated-client drift (`docs/spec-kit/API_CONTRACT.md`). G1 requires reproducible OpenAPI generation (`AGENTS.md:47`); G3 requires OpenAPI diff (`docs/TEST.md:46`).
- Doc tension: US-SOLVE-001 demands 403 when Bob accepts (`docs/SPEC.md:27`) vs "don't reveal private-object existence" (`docs/spec-kit/AUTHORIZATION_MATRIX.md:48`). Rule: public/visible object -> 403; private or existence-sensitive -> 404. Record per operation in the spec.
- `type` is `urn:226os:problem:<code>`, a non-dereferenceable identifier (ADR-0004). Codegen: oapi-codegen for Go, openapi-typescript for web, via `make generate` (ADR-0004; ogen is the documented fallback).

## Expert practices
- Write OpenAPI first, review, then generate; spec lives in one file, `$ref` into `components/{schemas,parameters,responses}`. [OpenAPI 3.1.0](https://spec.openapis.org/oas/v3.1.0)
- OpenAPI 3.1 schemas are JSON Schema 2020-12: nullable via `type: [string, "null"]`, not `nullable: true` (3.0 only). [OAS 3.1 Schema Object](https://spec.openapis.org/oas/v3.1.0#schema-object)
- Set `additionalProperties: false` on mutation request bodies to encode "unknown fields rejected"; add `maxLength`, `maxItems`, `format: uuid`, `format: date-time`, `enum` everywhere. [JSON Schema 2020-12](https://json-schema.org/draft/2020-12/json-schema-core)
- Give every operation a stable `operationId`, `security` requirement, all error responses via shared `components/responses/Problem*`, and `examples`. [OAS Operation Object](https://spec.openapis.org/oas/v3.1.0#operation-object)
- Declare the session cookie as `securitySchemes: { type: apiKey, in: cookie, name: <cookie> }` and the CSRF header as a required header parameter on mutations. [OAS Security Scheme](https://spec.openapis.org/oas/v3.1.0#security-scheme-object)
- Problem details: `type` is the primary identifier (absent => `about:blank`); `title` stable per type; `status` must equal the HTTP status; extension members (`code`, `request_id`, `violations`) allowed, clients ignore unknown ones; vet content, no implementation details. RFC 9457 obsoletes RFC 7807. [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html)
- Use `422` + `violations` for semantic validation; `400` only for unparseable syntax. [RFC 9110 §15.5](https://www.rfc-editor.org/rfc/rfc9110.html#section-15.5)
- Action endpoints are `POST` (non-idempotent by HTTP semantics) — make them safe to retry via DB uniqueness + `Idempotency-Key`. `PATCH` uses JSON Merge Patch (ADR-0004). [RFC 9110 methods](https://www.rfc-editor.org/rfc/rfc9110.html#section-9.2.2), [RFC 7396](https://www.rfc-editor.org/rfc/rfc7396)
- `Idempotency-Key` is an IETF Internet-Draft (latest seen -07, Oct 2025, UNVERIFIED: re-check datatracker.ietf.org before relying on it), NOT an RFC: string value; replay with same key + same payload returns the stored response; same key + different payload -> 422; concurrent in-flight duplicate -> 409. Treat draft semantics as guidance; our own convention is pinned in ADR-0004. [draft-ietf-httpapi-idempotency-key-header](https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/)
- Cursor: encode `(sort_key, id)` server-side, base64url + HMAC (key is a secret, `docs/RULES.md:25`); reject tampered/expired cursors with 400 problem; return `next_cursor` (null at end). Never expose offsets. [RFC 8288 Web Linking (optional Link header)](https://www.rfc-editor.org/rfc/rfc8288)
- Conditional updates: `ETag` on GET, `If-Match` on PATCH -> `412` on mismatch, `428` if required and missing. [RFC 9110 §13.1.1](https://www.rfc-editor.org/rfc/rfc9110.html#section-13.1.1), [RFC 6585 428](https://www.rfc-editor.org/rfc/rfc6585#section-3)
- `429`/`503` include `Retry-After`. [RFC 6585 §4](https://www.rfc-editor.org/rfc/rfc6585#section-4), [RFC 9110 Retry-After](https://www.rfc-editor.org/rfc/rfc9110.html#section-10.2.3)
- Lint the spec in CI (the pinned linter run by `make lint`, see the `lint-contract` recipe in `mk/20-gates.mk`); fail on missing error responses, missing `operationId`, unbounded strings/arrays. [OWASP API Security Top 10 — API4 Unrestricted Resource Consumption](https://api-security.owasp.org/editions/2023/en/0xa4-unrestricted-resource-consumption/)
- Drift check: regenerate Go types + TS client from `api/openapi.yaml` and `git diff --exit-code`; additionally validate real responses against the schema in HTTP integration tests.
- Versioning: additive changes within `/api/v1`; removals/renames require a new version or deprecation window recorded in an ADR (`docs/DEV.md:64`).

## SOLID / design notes for this layer
- Contract is the port: transport DTOs come from codegen; domain types never carry JSON tags for the wire.
- One mapping function per direction (DTO->command, domain->DTO) per operation; one problem writer for all errors.
- Keep validation in two layers: schema (shape, bounds) in transport, business rules in service. Do not duplicate business rules in OpenAPI.

## TDD & tests for this layer
- RED: write the OpenAPI operation + a failing HTTP test asserting status, headers and body schema before the handler exists.
- Contract tests validate every response (2xx and problem) against the spec; fixtures for each problem `type` (`docs/spec-kit/API_CONTRACT.md`).
- Table tests per operation: malformed JSON (400), unknown field (400/422 per ADR), oversized (413), invalid cursor (400), limit > max (422, ADR-0004), unauthenticated (401), CSRF missing (403), non-owner (403/404), conflict (409), rate limit (429 + `Retry-After`).
- Idempotency: same key+same body -> identical response, one DB row; same key+different body -> 422; different subject reusing key -> independent.
- Pagination: walk all pages, assert no duplicates/gaps and stable order under concurrent inserts.
- Redaction: induce DB/upstream failure, assert 500/503 body has no SQL, hostnames, tokens (`docs/spec-kit/API_CONTRACT.md`).

## Anti-patterns (reject in review)
- Code-first spec, hand-written DTOs in web or API, endpoints not in `api/openapi.yaml`.
- `nullable: true` in a 3.1 document; schemas without bounds; `additionalProperties` open on mutations.
- Custom error envelopes, `{ "error": "..." }`, stack/SQL in `detail`, `status` member differing from HTTP status.
- Offset/page-number pagination; unsigned cursors; unbounded `limit`.
- State changes via `GET`; bearer tokens or tokens in URLs.
- Idempotency keys not bound to subject+route+body hash; unlimited retention.
- 403 vs 404 chosen ad hoc, leaking private object existence.
- Claiming "RFC" for the Idempotency-Key draft.

## Done checklist
- [ ] Operation in `api/openapi.yaml` reviewed before code; lint clean.
- [ ] All responses incl. problems documented with examples.
- [ ] Generated server/client regenerated; `git diff --exit-code` clean.
- [ ] Contract, negative, idempotency, pagination and redaction tests green on real PG.
- [ ] 403/404 choice and idempotency/ETag policy recorded per operation.
- [ ] No status code, limit or retention value differs from ADR-0004 without a new ADR.

## References
- Deep dive: [references/openapi-patterns.md](references/openapi-patterns.md)
- https://spec.openapis.org/oas/v3.1.0 · https://www.rfc-editor.org/rfc/rfc9457.html · https://www.rfc-editor.org/rfc/rfc9110.html
- https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/ · https://api-security.owasp.org/editions/2023/en/0x11-t10/
