# 226os P0 REST and Error Contract

**Status:** ACCEPTED (ADR-0004, ADR-0003, ADR-0005; 2026-11-11, approver `Y4NN777`)
**Source of truth:** [`api/openapi.yaml`](../../api/openapi.yaml) (OpenAPI 3.1.0, `servers: /api/v1`). Verbs, paths, schemas, status codes and per-operation `x-authz` live there; this note summarises the accepted decisions. If they disagree, the YAML wins and this file is a bug.

## Accepted conventions

General conventions (base path, UUIDs, timestamps, cookie session plus `X-CSRF-Token`, Problem type, pagination, status policy, `lang`, JSON Merge Patch) are stated once in `api/openapi.yaml` `info.description`; they are not repeated here.

- Extensions: every operation has `operationId`, a tag, `x-priority` (P0|P1), `x-spec` (US-* refs) and `x-authz` with `allow` and `notes` (subjects allowed, 403/404 choice). The `matrix` label (a row of `AUTHORIZATION_MATRIX.md`) is present on only some operations; labels follow ROADMAP C9 and do not always match the matrix names.
- Resources: `/auth/*`, `/me`, `/me/profile`, `/profiles/{username}`, questions, answers, `/votes`, `accept`/`unaccept`, resources (sources, error-reports, reviews P1), projects, contribution needs, applications, `/learn/articles`, `/learn/series` (lesson progress P1), `/search`, `/moderation/*`.
- Error catalogue: `ErrorCode` enum with `x-error-catalogue` is closed; `type` is `urn:226os:problem:<code>`, non-dereferenceable at P0.
- `/search` is the pagination exception (grouped, bounded `per_group`, no cursor); it spans questions, resources, projects and Learn.
- Idempotency: `Idempotency-Key` on the six create operations that declare it in the YAML (question, answer, project, contribution need, application, Learn article), 24 h retention, our own convention (the IETF draft is not an RFC). Votes are idempotent by database uniqueness `(user, target_type, target_id)`.
- Concurrency: `ETag` / `If-Match` is implemented on the question `PATCH` only. Answers and Learn articles have no P0 edit endpoint; they follow the same rule when one is added.
- Security: no bearer tokens; anonymous unsafe methods (sign-in, sign-up, reset) use the pre-session token from `GET /auth/csrf` plus an Origin/Fetch-Metadata check.
- Schemas: `additionalProperties: false` on request bodies, bounded strings and arrays, enums for statuses. Tags are lowercase slugs of at most 32 characters, at most 5 per item.
- Status model: publication `status` is internal (published/removed; articles also draft/archived); a question exposes derived `state`; a project exposes `visibility`.
- Public profile never returns email; each optional field has an owner-controlled visibility (`public|members|private`).
- Generation: oapi-codegen (Go), ogen as the documented fallback; openapi-typescript (web); `make generate && git diff --exit-code`.

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
