# 226os P0 REST and Error Contract

**Status:** PROPOSED contract; it becomes normative only after human approval and creation of `api/openapi.yaml`

## API defaults proposed for approval

- Base path: `/api/v1`.
- JSON UTF-8; RFC3339 UTC timestamps; opaque UUID identifiers.
- `GET` collection responses use cursor pagination with `limit` bounded to a proposed maximum of 50. Cursors are opaque and signed or integrity-protected; ordering is deterministic by a stable primary sort plus UUID tie-breaker.
- Request bodies and query filters are schema-validated before authorization-sensitive work. Unknown fields are rejected for mutations unless explicitly marked extensible.
- Browser mutations use session cookies plus CSRF tokens. The API does not expose bearer tokens in this P0 proposal.
- Retried writes accept an `Idempotency-Key` where the operation can create duplicates: applications, votes, relation creation, and selected publication endpoints. The server binds a key to subject, route, request hash, and bounded retention.
- `ETag`/`If-Match` is proposed for edits where lost updates are likely; absent approval, last-write-wins is not an implicit contract.

## Proposed operation inventory

| Resource | Proposed operations | Authorization reference |
|---|---|---|
| sessions | `POST /sessions`, `DELETE /sessions/current` | AUTHORIZATION_MATRIX.md |
| identity | `GET /users/me`, `PATCH /profiles/me`, `GET /profiles/{id}` | AUTHORIZATION_MATRIX.md |
| questions | `GET /questions`, `POST /questions`, `GET/PATCH/DELETE /questions/{id}` | AUTHORIZATION_MATRIX.md |
| answers | `GET /questions/{id}/answers`, `POST /questions/{id}/answers`, `PATCH/DELETE /answers/{id}` | AUTHORIZATION_MATRIX.md |
| question actions | `POST /questions/{id}/accept`, `POST/DELETE /votes` | AUTHORIZATION_MATRIX.md |
| resources | `GET /resources`, `POST /resources`, `GET/PATCH /resources/{id}` | AUTHORIZATION_MATRIX.md |
| projects | `GET /projects`, `POST /projects`, `GET/PATCH/DELETE /projects/{id}` | AUTHORIZATION_MATRIX.md |
| contribution needs | `POST /projects/{id}/contribution-needs`, `GET/PATCH /contribution-needs/{id}` | AUTHORIZATION_MATRIX.md |
| applications | `POST /contribution-needs/{id}/applications`, `GET /projects/{id}/applications`, `PATCH /applications/{id}` | AUTHORIZATION_MATRIX.md |
| relations/search | `GET /search`, `POST /content-relations`, relation reads through authorized representations | AUTHORIZATION_MATRIX.md |

Exact verbs, fields, status codes, filters, and response envelopes remain `OPEN` until `api/openapi.yaml` is reviewed.

## Problem details proposal

Content type: `application/problem+json`.

```json
{
  "type": "https://226os.tech/problems/validation-error",
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

Proposed rules:

- `type`, `title`, `status`, and `code` are stable; `detail` is human-safe and may be localized at the client.
- `request_id` is opaque and safe to show to support staff; it must not contain credentials or personal data.
- `violations` contains field names and safe machine codes, never submitted secrets or database details.
- Do not expose stack traces, SQL, provider credentials, session tokens, internal hostnames, or private-object existence.
- Proposed mappings: `400` malformed syntax, `401` missing/invalid session, `403` authenticated but forbidden, `404` public object absent or intentionally undisclosed, `409` uniqueness/state conflict, `413` body too large, `422` semantic validation, `429` rate limit, `500` unexpected server failure, `503` unavailable dependency.
- `Retry-After` is required on applicable `429` and `503` responses. Error bodies must remain bounded.

## Security and contract tests required before G0 closure

- Anonymous, owner, collaborator, moderator, and third-party cases for every sensitive operation.
- CSRF rejection, session rotation/revocation, malformed JSON, unknown fields, oversized bodies, invalid cursors, and request replay.
- Duplicate idempotency key with same and different request hashes.
- Redaction assertions for validation, authorization, database, and upstream GitHub failures.
- OpenAPI lint, schema examples, generated-client drift check, and contract fixtures.

## G0 approval record

Human approval must record exact status codes, request/response schemas, pagination limits, idempotency retention, optimistic concurrency policy, and whether `type` URLs are public documentation endpoints or identifiers only.
