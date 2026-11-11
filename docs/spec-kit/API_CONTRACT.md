# 226os P0 REST and Error Contract

**Status:** ACCEPTED (ADR-0004, ADR-0003, ADR-0005; 2026-11-11, approver `Y4NN77`)
**Source of truth:** [`api/openapi.yaml`](../../api/openapi.yaml) (OpenAPI 3.1.0, `servers: /api/v1`). Verbs, paths, schemas, status codes and per-operation `x-authz` live there; this note summarises the accepted decisions. If they disagree, the YAML wins and this file is a bug.

## Accepted conventions

- Every operation has `operationId`, a tag, `x-priority` (P0|P1), `x-spec` (US-* refs) and `x-authz` (subjects allowed, 403/404 choice, matrix row).
- Resources: `/auth/*` (csrf, sign-up, sign-in, sign-out, session, verify-email, password-reset), `/me`, `/me/profile`, `/profiles/{username}`, questions, answers, `/votes`, `accept`/`unaccept`, resources (sources, error-reports, reviews P1), projects, contribution needs, applications, `/learn/articles`, `/learn/series` (lesson progress P1), `/search`, `/moderation/*`.
- Errors: RFC 9457 `application/problem+json` through one `Problem` schema (`type`, `title`, `status`, `detail`, `instance`, `code`, `request_id`, `violations[]`). **`type` is `urn:226os:problem:<code>`**, a non-dereferenceable identifier at P0. The `ErrorCode` enum with `x-error-catalogue` is the closed catalogue.
- Status policy: 401 no or invalid session (including Learn publish without a session); **403 if the object is visible to the caller** (for example a non-author accepting an answer), CSRF failure, `EMAIL_NOT_VERIFIED`, `SELF_VOTE_FORBIDDEN`, `REAUTH_REQUIRED`; **404 if the object is absent or private and its existence must not leak** (private projects, private profiles); 409 duplicate application, state conflict, in-flight idempotent duplicate; 412 `If-Match` mismatch; 413; 422 validation, `limit` above 50 and `Idempotency-Key` reuse with a different body; 428 missing `If-Match`; 429 with `Retry-After`; 400 only for malformed syntax and an invalid cursor.
- Pagination: opaque integrity-protected `cursor` plus `limit` (default 20, 1..50, above 50 returns 422); envelope `{items, next_cursor}`; deterministic order with a UUID tie-breaker. `/search` is the exception (grouped, bounded `per_group`, no cursor); it spans questions, resources, projects and **Learn**.
- Idempotency: optional `Idempotency-Key` header on creates (application, answer, project, article, question, contribution need), **24 h retention**, our own convention (the IETF draft is not an RFC). Votes are idempotent by database uniqueness `(user, target_type, target_id)`.
- Concurrency: `ETag` / `If-Match` on questions, answers and Learn articles. The question `PATCH` is the only P0 edit endpoint of the three; the others follow the same rule when added.
- Security: `cookieAuth` (opaque session cookie `__Host-226os_session` deployed, `226os_session` local-dev) plus `csrfToken` (`X-CSRF-Token`) on every unsafe method, including anonymous ones (sign-in, sign-up, reset, using the pre-session token from `GET /auth/csrf` plus an Origin/Fetch-Metadata check); no bearer tokens.
- Schemas: `additionalProperties: false` on all request bodies, bounded strings and arrays, enums for statuses, UUID ids, RFC 3339 UTC, `lang` fr|en on user content (original language preserved, no auto-translation). PATCH is JSON Merge Patch. Tags are lowercase slugs of at most 32 characters, at most 5 per item.
- Status model in the API: publication `status` is internal (published/removed; articles also draft/archived and may be created as `draft`); a question exposes derived `state` open/answered/resolved; a project exposes `visibility` public/private.
- Public profile never returns email; fields default to public and each optional field has an owner-controlled visibility (`public|members|private`).
- Generation: oapi-codegen (Go), with a verification step at bootstrap and ogen as the documented fallback; openapi-typescript (web); `make generate && git diff --exit-code`.

## Decisions on points first listed as divergences

1. Paths: the YAML paths are canonical (`/auth/*`, `/me`, `/me/profile`, `/profiles/{username}`, `accept` plus `unaccept`); `AUTHORIZATION_MATRIX.md` uses them.
2. Matrix rows that the YAML lacks (`DELETE /questions/{id}`, `PATCH/DELETE /answers/{id}`, `POST/PATCH /resources`, `PATCH /contribution-needs/{id}`, `DELETE /projects/{id}`, a standalone `POST /content-relations`) are **out of P0**; moderators can only hide and restore. Relations are written through `links[]` on create/update and read as resolved `related[]`. The moderation operations that P0 needs are in the YAML under `/moderation/*`.
3. 403 versus 404 follows the rule above, per operation in `x-authz`.
4. Learn is `/learn/articles` (`type` article|tutorial) and `/learn/series`. Lesson progress and resource reviews are P1.
5. Self-vote: 403 `SELF_VOTE_FORBIDDEN`. Unverified email: 403 `EMAIL_NOT_VERIFIED` on publishing actions. Moderator scope and audit: ADR-0003.
6. Question solve state is derived (`state`); publication lifecycle and project visibility are separate (ADR-0005).
7. Intent enum (`ask|help|integrate|contribute|learn`) is taken from the prototype welcome step; resource `category` and SDK language are slugs backed by reference tables.

## Problem details

Content type: `application/problem+json`.

```json
{
  "type": "urn:226os:problem:VALIDATION_ERROR",
  "title": "Request validation failed",
  "status": 422,
  "detail": "One or more fields are invalid.",
  "instance": "/api/v1/questions",
  "code": "VALIDATION_ERROR",
  "request_id": "req_opaque_id",
  "violations": [
    {"field": "title", "code": "required"}
  ]
}
```

Rules:

- `type`, `title`, `status` and `code` are stable; `detail` is human-safe and may be localized at the client.
- `request_id` is opaque and safe to show to support staff; it must not contain credentials or personal data.
- `violations` contains field names and safe machine codes, never submitted secrets or database details.
- Do not expose stack traces, SQL, provider credentials, session tokens, internal hostnames, or private-object existence.
- `Retry-After` is required on applicable `429` and `503` responses. Error bodies stay bounded.

## Security and contract tests required (G3/G4)

- Anonymous, owner, accepted applicant ("collaborator"), maintainer, moderator and third-party cases for every sensitive operation.
- CSRF rejection, session rotation/revocation, malformed JSON, unknown fields, oversized bodies, invalid cursors, `limit` above 50, and request replay.
- Duplicate idempotency key with the same and a different request hash; `If-Match` mismatch and missing header.
- Redaction assertions for validation, authorization, database and upstream GitHub failures.
- OpenAPI lint, schema examples, generated-client drift check and contract fixtures.
