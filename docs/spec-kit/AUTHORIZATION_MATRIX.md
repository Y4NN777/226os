# 226os P0 Authentication and Authorization Matrix

**Status:** ACCEPTED (ADR-0003, 2026-11-11, approver `Y4NN77`)
**Subject roles:** anonymous, member, moderator, owner. `owner` always means ownership of the specific resource, not a global administrator. "Collaborator" is **not** a role: it is a resource-scoped relationship (an applicant whose application was accepted on a project, ADR-0005). Owner-added `project_maintainers` have explicit rights on their project (edit project, manage needs and applications).
**Paths** are the `api/openapi.yaml` paths under `/api/v1`.

## Authentication and session policy (ADR-0003)

- Email/password; no external identity provider at P0.
- Argon2id m=64 MiB, t=3, p=1, benchmarked on the target host, never below the OWASP floor m=19 MiB, t=2, p=1. Password length 12-128, no composition rules.
- Sessions: opaque random token, only its hash stored; **idle 2 h, absolute 7 d**. Re-authentication is required when the last sign-in is more than **15 min** old for password change, email change and session revocation (403 `REAUTH_REQUIRED`). Rotate on sign-in, password change/reset, email change and privilege change; revoke on sign-out and security events.
- Cookie `__Host-226os_session` in `staging-test`/`production-live`, `226os_session` in `local-dev`; `HttpOnly`, `Secure` (deployed), `SameSite=Lax`; no secrets in URLs.
- CSRF: synchronizer token in `X-CSRF-Token` on every unsafe method. Anonymous auth endpoints use a pre-session token from `GET /auth/csrf` plus an Origin/Fetch-Metadata check.
- Reset token 30 min, verification token 24 h. Generic responses prevent enumeration; sign-in, sign-up and reset are rate limited by IP and account (windows in ADR-0003).
- **Status codes:** 401 unauthenticated; 403 authenticated but not allowed on an object visible to the caller; 404 when a private object's existence must not leak (rule: 403 if visible to the caller, 404 if not).
- **Unverified email:** may sign in, read and draft. Publishing a question, answer, article, project, need, application or review needs a verified email, else 403 `EMAIL_NOT_VERIFIED`.

## Matrix

Legend: `A` allowed, `D` denied (401 for anonymous, 403 or 404 per the rule above), `O` allowed only for the resource owner, `M` moderator-scoped, `C` authenticated member (verified email when the action publishes), `V` visibility-dependent public read, `P1` not in P0.

| Capability / endpoint (OpenAPI) | Anonymous | Member | Owner | Moderator | Required negative cases |
|---|---:|---:|---:|---:|---|
| `GET /search`, public `GET /questions`, `/resources`, `/projects`, `/learn/*` | V | V | V | V | private, unpublished and removed content never leaks |
| `GET /auth/csrf`, `POST /auth/sign-up`, `/auth/sign-in`, `/auth/password-reset/*`, `/auth/verify-email` | A | A | A | A | rate limit, generic errors, no enumeration, CSRF |
| `POST /auth/sign-out`, `GET /auth/session` | D / A (session state) | A | A | A | cannot revoke another user's session |
| `GET /me`, `PATCH /me/profile` | D | C | C | C | anonymous and cross-user access denied |
| `GET /profiles/{username}` | V | V | V | V | field visibility respected; email never returned |
| `POST /questions` | D | C (verified) | C | C | unauthenticated, unverified, CSRF, invalid body |
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
| `PATCH /projects/{projectId}` | D | O or maintainer | O | D | non-owner cannot mutate; private project 404 |
| `DELETE /projects/{id}` | not in P0 | not in P0 | not in P0 | hide only (below) | no delete endpoint at P0 |
| `POST /projects/{projectId}/contribution-needs` | D | O or maintainer | O | D | only owner or maintainer publishes a need |
| `PATCH /contribution-needs/{id}` | not in P0 | not in P0 | not in P0 | not in P0 | need status changes are out of P0 |
| `POST /contribution-needs/{needId}/applications` | D | C (verified, not the project owner) | C | C | duplicate application, closed need, re-apply only after own withdrawal |
| `GET /projects/{projectId}/applications` | D | O or maintainer | O | D | third party denied; no count or data leak |
| `PATCH /applications/{applicationId}` | D | applicant (withdraw) or project owner/maintainer (accept, decline) | O | D | cross-project IDOR, invalid state transition, no re-apply after decline |
| `GET /learn/series/{seriesId}/lessons/{lessonId}/progress`, `PUT` (lesson progress) | P1 | P1 | P1 | P1 | table deferred to migration 00011 |
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
