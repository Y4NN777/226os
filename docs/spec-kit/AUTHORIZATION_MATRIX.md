# 226OS P0 Authentication and Authorization Matrix

**Status:** PROPOSED contract; human approval required
**Subject roles:** anonymous, member, moderator, owner. `owner` always means ownership of the specific resource, not a global administrator.

## Authentication/session policy proposal

- Email/password is the proposed P0 authentication method; no external identity provider is assumed.
- Passwords are stored only as Argon2id hashes with an implementation-reviewed memory/time/parallelism profile.
- Sessions are opaque random identifiers. Store only a hash and metadata: user, created time, last seen, expiry, revoked time, user-agent/device summary, and security-event reason.
- Rotate the session after authentication, privilege changes, password reset, and sensitive account changes. Revoke on logout and security events.
- Cookie proposal: `HttpOnly`, `Secure` in deployed environments, `SameSite=Lax`, narrow domain/path, no secrets in URLs. Use a synchronizer CSRF token for every state-changing request.
- Proposed limits: idle timeout 24 hours, absolute lifetime 30 days, reset token lifetime 30 minutes, verification token lifetime 24 hours. These are defaults for approval, not accepted values.
- Generic responses for sign-in, registration, and reset prevent account enumeration. Auth endpoints require IP and account limits; exact windows are in `G0_DECISIONS.md`.
- Recovery/verification is required for production. A controlled demo account is not a substitute for the policy.

## Matrix

Legend: `A` allowed, `D` denied, `O` allowed only for the resource owner, `M` moderator-scoped, `C` authenticated member, `V` visibility-dependent public read.

| Capability / endpoint family | Anonymous | Member | Owner | Moderator | Required negative cases |
|---|---:|---:|---:|---:|---|
| `GET /search`, public questions/resources/projects | V | V | V | V | private/unpublished content never leaks |
| `POST /sessions`, registration, reset request | A | A | A | A | rate limit, generic errors, no enumeration |
| `DELETE /sessions/current` | D | A | A | A | cannot revoke another session |
| `GET /users/me`, `PATCH /profiles/me` | D | C | C | C | anonymous and cross-user access denied |
| `GET /profiles/{id}` | V | V | V | V | privacy settings respected |
| `POST /questions` | D | C | C | C | unauthenticated, CSRF, invalid body |
| `PATCH/DELETE /questions/{id}` | D | O | O | M only by documented moderation policy | non-owner cannot mutate; no hidden moderator bypass |
| `POST /questions/{id}/answers` | D | C | C | C | unauthenticated, CSRF, invalid question state |
| `PATCH/DELETE /answers/{id}` | D | O | O | M only by policy | answer owner and moderator boundaries |
| `POST /questions/{id}/accept` | D | O if question owner | O | M only by policy | Bob/third party cannot accept; one accepted answer invariant |
| `POST /votes` / `DELETE /votes` | D | C once per target | C | C | duplicate vote, self-vote policy, CSRF |
| `POST /resources` and `PATCH /resources/{id}` | D | C or O per source-ownership policy | O where applicable | M | provider claims require source/provenance; unauthorized edit |
| `POST /projects` | D | C | C | C | invalid GitHub URL, SSRF, rate limit |
| `PATCH/DELETE /projects/{id}` | D | O | O | M | non-owner cannot mutate; private project visibility |
| `POST /projects/{id}/contribution-needs` | D | O | O | M | only project owner can publish need unless explicitly delegated |
| `POST /contribution-needs/{id}/applications` | D | C | C | C | duplicate application, self/closed need policy |
| `GET /projects/{id}/applications` | D | O | O | M | Charlie/third party denied; no count/data leak |
| `PATCH /applications/{id}` | D | applicant or project owner | O | M | cross-project IDOR, invalid state transition |
| `POST /content-relations` | D | C subject to relation policy | O for owned source | M | dangling/private target and unauthorized relation |
| moderation/report endpoints | D | C to report | O for own report | M | reporter privacy and moderator audit |

## Required authorization properties

1. Every sensitive endpoint identifies the acting subject on the server from the session, never from a client-supplied owner ID.
2. Resource ownership, visibility, status, and relation target are checked in the same transaction or authorization boundary as the mutation/read.
3. Denied access returns the approved problem response without revealing whether a private object exists when enumeration is a risk.
4. Moderator access is narrow, audited, and never inferred from a frontend role flag.
5. All state-changing browser requests require CSRF validation and same-origin policy checks appropriate to the deployment.
6. Duplicate writes are safe under retries through database uniqueness and, where needed, idempotency keys.

## Approval blockers

The human must approve the auth method, session timeouts, recovery requirement, role semantics, moderator scope, self-voting policy, resource-edit ownership policy, and exact rate-limit windows before TASK-SEC-001 is implementation-ready.
