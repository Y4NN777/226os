# 226os P0 Authentication and Authorization Matrix

**Status:** ACCEPTED (ADR-0003, 2026-11-11, approver `Y4NN777`)
**Subject roles:** anonymous, member, moderator, owner. `owner` always means ownership of the specific resource, not a global administrator. "Collaborator" is **not** a role: it is a resource-scoped relationship (an applicant whose application was accepted on a project, ADR-0005). Owner-added `project_maintainers` have explicit rights on their project (edit project, manage needs and applications).
**Paths** are the `api/openapi.yaml` paths under `/api/v1`.

## Authentication and session policy

The policy (Argon2id parameters, session lifetimes, cookie and CSRF rules, token lifetimes, 401/403/404 rule, unverified-email limits, rate limits) is owned by [ADR-0003](../decisions/0003-authentication-sessions-csrf-rate-limits.md) and is not repeated here. Rule used below: 403 if the object is visible to the caller, 404 if its existence must not leak. Publishing needs a verified email (403 `EMAIL_NOT_VERIFIED`).

**Status:** negative authorization tests mapped: none yet; `scripts/authz-tests.txt` is empty and slices add their lines with their routes (checked by `scripts/check-authz-routes.sh`).

## Matrix

Legend: `A` allowed, `D` denied (401 for anonymous, 403 or 404 per the rule above), `O` allowed only for the resource owner, `M` moderator-scoped, `C` authenticated member (verified email when the action publishes), `V` visibility-dependent public read, `P1` not in P0.

| Capability / endpoint (OpenAPI) | Anonymous | Member | Owner | Moderator | Required negative cases |
|---|---:|---:|---:|---:|---|
| `GET /search`, public `GET /questions`, `/resources`, `/projects`, `/learn/*` | V | V | V | V | private, unpublished and removed content never leaks |
| `GET /questions/{questionId}`, `GET /questions/{questionId}/answers` (getQuestion, listAnswers) | V | V | V | V | non-published question is 404 |
| `GET /resources/{resourceId}/sources` (listResourceSources) | V | V | V | V | a stale source is never shown as verified |
| `GET /resources/{resourceId}/reviews` (listResourceReviews) | P1 | P1 | P1 | P1 | P1 read, minimal |
| `GET /projects/{projectId}/contribution-needs` (listContributionNeeds) | V | V | V | V | follows project visibility; private project 404 |
| `GET /learn/series`, `/learn/series/{seriesId}` | V | V | V | V | public; `my_progress` only with a session |
| `GET /auth/csrf`, `POST /auth/sign-up`, `/auth/sign-in`, `/auth/password-reset/*`, `/auth/verify-email` | A | A | A | A | rate limit, generic errors, no enumeration, CSRF |
| `POST /auth/sign-out`, `GET /auth/session` | D / A (session state) | A | A | A | cannot revoke another user's session |
| `GET /me`, `PATCH /me/profile` | D | C | C | C | anonymous and cross-user access denied |
| `GET /profiles/{username}` | V | V | V | V | field visibility respected; email never returned |
| `POST /questions` | D | C (verified) | C | C | unauthenticated, unverified, CSRF, invalid body |
| `POST /assistant/answers` (askAssistant; ADR-0012) | D (401) | C (verified) | C | C | unauthenticated 401; unverified 403 `EMAIL_NOT_VERIFIED`; missing or wrong CSRF 403; malformed JSON 400, oversized body 413, unknown field or `question` outside 3 to 1000 characters 422; 11th call in the hour 429 `RATE_LIMITED` with `Retry-After` (`assistant:user`, 10 per member per hour); 503 `SERVICE_UNAVAILABLE` when disabled, on provider error, provider 429, timeout (20 s) or empty output (`Retry-After: 30`), when the global guard is spent (`assistant:global`, 12 provider calls per hour, with `Retry-After`) or when the limiter fails (fail closed); the 400, 401, 403, 413, 422 and 429 answers, the disabled and spent-guard 503 and `no_source` never call the provider, and a provider failure is one call, never retried; private, removed, draft and archived content is never retrieved or cited; no `questions`, `answers` or `votes` row |
| `PATCH /questions/{questionId}` | D | O | O | D (moderators never edit others' content) | non-owner 403 on a visible object; `If-Match` required |
| `DELETE /questions/{id}` | not in P0 | not in P0 | not in P0 | hide only (below) | no delete endpoint at P0 |
| `POST /questions/{questionId}/answers` | D | C (verified) | C | C | unauthenticated, unverified, CSRF, closed question |
| `PATCH/DELETE /answers/{id}` | not in P0 | not in P0 | not in P0 | hide only (below) | no answer edit or delete endpoint at P0 |
| `POST /questions/{questionId}/accept`, `/unaccept` | D | O (question author) | O | D | third party 403; one accepted answer invariant |
| `POST /votes`, `DELETE /votes` | D | C, one per target; not own content | C | C | duplicate vote, self-vote 403 `SELF_VOTE_FORBIDDEN`, CSRF |
| `POST /resources/{resourceId}/error-reports` | D | C | C | C | reporter privacy, rate limit |
| `POST /resources`, `PATCH /resources/{id}` | not in P0 | not in P0 | not in P0 | not in P0 | no P0 story needs a member-facing create/edit; the owner loads catalogue content outside the API; moderators only hide |
| `POST /resources/{resourceId}/reviews` | P1 | P1 | P1 | P1 | one review per user and resource |
| `POST /projects` | D | C (verified) | C | C | invalid GitHub URL, SSRF, rate limit |
| `PATCH /projects/{projectId}` | D | O (owner only, C1) | O | D | non-owner cannot mutate (403 public, 404 private); a `project_maintainers` row grants nothing |
| `DELETE /projects/{id}` | not in P0 | not in P0 | not in P0 | hide only (below) | no delete endpoint at P0 |
| `POST /projects/{projectId}/contribution-needs` | D | O (owner only, C1) | O | D | only the owner publishes a need (403 public, 404 private) |
| `PATCH /contribution-needs/{id}` | not in P0 | not in P0 | not in P0 | not in P0 | need status changes are out of P0 |
| `POST /contribution-needs/{needId}/applications` | D | C (verified, not the project owner) | C | C | duplicate application, closed need, re-apply only after own withdrawal |
| `GET /projects/{projectId}/applications` | D | O (owner only, C1) | O | D | third party denied; no count or data leak |
| `PATCH /applications/{applicationId}` | D | applicant (withdraw) or project owner (accept, decline; C1) | O | D | cross-project IDOR, invalid state transition, no re-apply after decline |
| `GET /learn/series/{seriesId}/lessons/{lessonId}/progress`, `PUT` (lesson progress) | P1 | P1 | P1 | P1 | table deferred to a later migration |
| `POST /learn/articles` | D (401) | C (verified) | C | C | unauthenticated 401, unverified 403, CSRF |
| content relations via `links[]` on create/update | D | C: only the author of the source item, only to published public targets | O of source | remove only (below) | dangling/private target, non-author link |
| `POST /moderation/{targetType}/{targetId}/hide`, `/restore` | D | D | D | M | audit row in the same transaction; non-moderator 403 |
| `POST /moderation/error-reports/{reportId}/resolve` | D | D | D | M | audit row in the same transaction; non-moderator 403 |
| tag merge/rename | not in P0 | not in P0 | not in P0 | P1 (policy: moderators) | not exposed at P0 |

## Moderator scope at P0

Hide and restore content (questions, answers, resources, projects, articles, content relations) and resolve error reports. Nothing else. Every action writes an `audit_log` row in the same transaction. Moderators never edit others' content. The role is granted by the repository owner (ADR-0003).

## Required authorization properties

1. Every sensitive endpoint identifies the acting subject on the server from the session, never from a client-supplied owner ID.
2. Resource ownership, visibility, status and relation target are checked in the same transaction or authorization boundary as the mutation or read.
3. Denied access follows the rule: 403 if the object is visible to the caller, 404 if its existence must not leak.
4. Moderator access is narrow, audited and never inferred from a frontend role flag.
5. All state-changing browser requests require CSRF validation and same-origin checks.
6. Duplicate writes are safe under retries through database uniqueness and, where needed, idempotency keys.
