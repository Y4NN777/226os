# api-contract — OpenAPI 3.1 patterns

Illustrative fragments; names, limits and paths are PROPOSED until ADR-OPEN-004.

## Shared components
```yaml
openapi: 3.1.0
info: { title: 226os API, version: 1.0.0 }
servers: [{ url: /api/v1 }]
components:
  securitySchemes:
    session: { type: apiKey, in: cookie, name: __Host-session }   # cookie name pin per ADR-OPEN-003
  parameters:
    CsrfToken: { name: X-CSRF-Token, in: header, required: true, schema: { type: string, maxLength: 128 } }
    IdempotencyKey: { name: Idempotency-Key, in: header, required: false, schema: { type: string, maxLength: 255 } }
    Limit: { name: limit, in: query, schema: { type: integer, minimum: 1, maximum: 50, default: 20 } }  # max 50 proposed, API_CONTRACT.md:9
    Cursor: { name: cursor, in: query, schema: { type: string, maxLength: 512 } }
  schemas:
    Problem:
      type: object
      required: [type, title, status, code]
      properties:
        type: { type: string, format: uri-reference }
        title: { type: string }
        status: { type: integer, minimum: 400, maximum: 599 }
        detail: { type: string, maxLength: 1000 }
        instance: { type: string, format: uri-reference }
        code: { type: string, pattern: '^[A-Z_]+$' }
        request_id: { type: string, maxLength: 64 }
        violations:
          type: array
          maxItems: 50
          items: { type: object, required: [field, code], additionalProperties: false,
                   properties: { field: { type: string }, code: { type: string } } }
    PageMeta:
      type: object
      required: [next_cursor]
      properties: { next_cursor: { type: [string, "null"] } }    # 3.1 null style
  responses:
    Problem400: { description: Malformed, content: { application/problem+json: { schema: { $ref: '#/components/schemas/Problem' } } } }
    # ... 401, 403, 404, 409, 413, 422, 429 (with Retry-After header), 500, 503
```
Header names (`X-CSRF-Token`) and cookie name are UNVERIFIED placeholders.

## Action endpoint example
```yaml
/questions/{id}/accept:
  post:
    operationId: acceptAnswer
    security: [{ session: [] }]
    parameters: [{ $ref: '#/components/parameters/CsrfToken' }]
    requestBody:
      required: true
      content:
        application/json:
          schema: { type: object, required: [answer_id], additionalProperties: false,
                    properties: { answer_id: { type: string, format: uuid } } }
    responses:
      '200': { description: Accepted answer recorded }
      '401': { $ref: '#/components/responses/Problem401' }
      '403': { $ref: '#/components/responses/Problem403' }   # SPEC.md:27 (public question)
      '404': { $ref: '#/components/responses/Problem404' }
      '409': { $ref: '#/components/responses/Problem409' }   # another answer already accepted (policy OPEN: replace vs conflict)
      '422': { $ref: '#/components/responses/Problem422' }
```

## Idempotency store (server side)
- Table keyed by `(subject_id, route, key)`; columns: request_hash, status (in_progress/completed), response status+body, created_at, expires_at.
- Flow: insert `in_progress` with `ON CONFLICT DO NOTHING`; if conflict: hash differs -> 422; still in progress -> 409; completed -> replay stored response.
- Purge by `expires_at` (retention OPEN, `docs/spec-kit/API_CONTRACT.md:70`). Never store secrets in the cached body.
- Source: https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/ (Internet-Draft, not RFC).

## Drift gate (proposed Make target, recipe to be reviewed)
1. Lint `api/openapi.yaml`.
2. Regenerate Go server types and TS client.
3. `git diff --exit-code` on generated paths.
4. Run HTTP integration tests that validate responses against the spec.
Codegen tools (e.g. oapi-codegen for Go, openapi-typescript for web) are UNVERIFIED choices — pin per ADR and license-check (`docs/research/REFERENCES.md:3`).
