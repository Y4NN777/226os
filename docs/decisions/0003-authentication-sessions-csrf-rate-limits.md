# ADR-0003 — Authentication, sessions, CSRF, and rate limits

- **Status:** ACCEPTED
- **Date:** 2026-11-11
- **Approver:** `Y4NN77`
- **Closes:** ADR-OPEN-003
- **Refines:** G0-AUTH-001, G0-AUTH-002, G0-AUTH-003, G0-AUTH-004, G0-AUTHZ-001
- **Related:** `docs/spec-kit/AUTHORIZATION_MATRIX.md`, `docs/spec-kit/API_CONTRACT.md`, `api/openapi.yaml`, [ADR-0004](0004-api-contract-conventions.md), [ADR-0005](0005-data-model-migrations-search.md)

## Context

The G0 register fixes the model (first-party email/password, opaque server sessions in PostgreSQL, Argon2id, `SameSite=Lax`, synchronizer CSRF, PostgreSQL rate limits). This ADR fixes the parameters, windows and lifetimes so TASK-SEC-001 can start. The earlier matrix value of a 24-hour idle timeout was far above OWASP Session Management guidance and is replaced.

## Options considered

| Topic | Options | Outcome |
|---|---|---|
| Identity | first-party (chosen), OIDC provider, stateless signed cookies | OIDC adds an external dependency for P0; stateless tokens make revocation hard |
| Session store | opaque hashed token in PostgreSQL (chosen), JWT | one DB lookup per request is acceptable at P0 scale |
| CSRF | synchronizer token (chosen), double-submit cookie, SameSite=Strict only | double-submit is weaker against subdomain cookie injection; Strict breaks inbound links |
| Rate limits | PostgreSQL buckets (chosen), Valkey, proxy-only | Valkey is excluded by ADR-0006; a proxy cannot enforce per-account limits |
| Idle timeout | 24 h, 30 min, 2 h (chosen) | 2 h idle with 7 d absolute and re-authentication for sensitive actions |

## Decision

**Passwords.** Argon2id via `golang.org/x/crypto/argon2`, 16-byte random salt, 32-byte output, PHC-format string with its parameters stored per hash so they can be raised later (rehash on successful sign-in). Profile **m=64 MiB, t=3, p=1**, benchmarked on the target host (target about 100 to 250 ms per hash) and **never below the OWASP floor m=19 MiB (19456 KiB), t=2, p=1**. Hashing concurrency is capped by a semaphore. Length 12 to 128 characters, no composition rules, Unicode NFKC normalization; a breached-password check is a later improvement (no external call at P0). An unknown email runs one Argon2id against a dummy hash to equalize timing; comparisons use `crypto/subtle`.

**Sessions.** 32 bytes from `crypto/rand` (base64url); only the SHA-256 of the token is stored. A session is valid iff `revoked_at IS NULL` and now is below both expiries. **Idle timeout 2 hours, absolute lifetime 7 days.** `sessions.authenticated_at` records the last sign-in or re-authentication. **Re-authentication (password prompt) is required when `authenticated_at` is older than 15 minutes** for password change, email change and session revocation; otherwise 403 `REAUTH_REQUIRED`. A new token is issued on sign-in, password change/reset, email change and privilege change, and the old one is revoked. Sign-out revokes. Password reset revokes all other sessions. `last_seen_at` is touched at most every 5 minutes. Reset token 30 minutes, email-verification token 24 hours, both single-use, 32 random bytes, hashed at rest, never logged.

**Cookie.** `HttpOnly; Secure; SameSite=Lax; Path=/`, no `Domain` attribute (host-only), `Max-Age` equal to the absolute lifetime. Name **`__Host-226os_session`** in `staging-test` and `production-live` (the prefix requires `Secure`, `Path=/`, no `Domain`) and **`226os_session`** in `local-dev` (plain HTTP, no `Secure`). Configuration, not code branches, selects the profile; a test asserts the deployed profile emits the prefixed `Secure` cookie. No tokens in URLs.

**CSRF.** Synchronizer token: 32 random bytes bound to the session, stored hashed with it, returned by `GET /api/v1/auth/csrf`, `GET /api/v1/auth/session`, `GET /api/v1/me` and the sign-in response, **sent in the `X-CSRF-Token` header on every unsafe method** (`POST`, `PUT`, `PATCH`, `DELETE`) and compared in constant time; rotated with the session. **Anonymous auth endpoints** (sign-up, sign-in, password reset) have no session to bind to: they use a **pre-session CSRF token** obtained from `GET /api/v1/auth/csrf` (signed, short-lived, bound to a host-only pre-session cookie named `__Host-226os_csrf` or `226os_csrf` in local-dev) **plus an `Origin` / `Sec-Fetch-Site` same-origin check**. The same `Origin` / Fetch-Metadata check applies to every mutation, and mutation bodies must be `application/json`. Failure returns 403 problem `CSRF_FAILED`.

**Status codes.** 401 `UNAUTHENTICATED` when there is no valid session. 403 `FORBIDDEN` when the caller is authenticated but not allowed on an object that is visible to them. 404 `NOT_FOUND` when a private object's existence must not leak. **Rule: 403 if the object is visible to the caller, 404 if not.** Policy decisions live in the service layer, never in handlers or the frontend.

**Unverified email.** A user with an unverified email may sign in, read, and keep drafts. **Publishing a question, answer, article, project, contribution need, application or review requires a verified email**, otherwise 403 `EMAIL_NOT_VERIFIED`. Votes, reports and relations follow the same rule as the content they act on.

**Generic responses.** Sign-in returns one `INVALID_CREDENTIALS` response for unknown email and wrong password; sign-up and reset request return the same success body and similar timing whether or not the email exists. A suspended or deleted account receives the same response as a wrong password.

**Rate limits.** PostgreSQL fixed-window buckets (`rate_limit_buckets`, **`UNLOGGED` accepted**, so counters reset after a crash; this is a documented residual risk) keyed by an HMAC of the IP or the normalized email. Sign-in: 5 failures per account per 15 min, 20 per IP per 15 min. Registration: 5 per IP per hour. Reset request: 3 per account and 10 per IP per hour. Verification resend: 3 per account per hour. Authenticated writes: 60 per user per minute. GitHub import: 10 per user per hour. Anonymous search: 60 per IP per minute. Response 429 `RATE_LIMITED` with `Retry-After`; auth limits fail closed (a database error means deny). Throttle with backoff, no permanent lockout (account-lockout DoS). Client IP only from the Traefik-trusted `X-Forwarded-For`. Expired buckets are removed by a bounded cleanup job. Valkey only by a new ADR.

**Roles and moderation.** Policy subjects are `anonymous`, `member`, `moderator`, and `owner` (ownership of a specific resource, not a global admin). The acting user always comes from the session. "Collaborator" is not a role (see ADR-0005). **Moderator scope at P0: hide and restore content, and resolve error reports. Nothing else.** Every moderator action writes an `audit_log` row (actor, action, target type/id, reason, `request_id`, UTC time) **in the same transaction** as the action. Moderators never edit others' content and get no implicit ownership. The `moderator` role is granted by the repository owner through an audited operational step (no self-service and no API at P0). Reporter identity is not exposed to the reported user.

**Email delivery.** Mailpit in `local-dev`. Deployed environments send through SMTP configured by environment secrets; the provider is chosen during the Datacloud setup. That is an operational detail, not an open decision.

## Consequences

- **Security:** server-side revocation works; theft of the database yields neither live tokens nor passwords. Argon2id plus the concurrency cap trades CPU/RAM for hash strength. The 2 h idle timeout and the 15 min re-authentication window reduce the hijack window at the price of re-sign-ins. UNLOGGED counters can reset after a crash.
- **Data:** `users`, `user_credentials`, `sessions` (with `authenticated_at`), `email_verifications`, `password_resets`, `rate_limit_buckets`, `audit_log`; see `docs/spec-kit/DATABASE_DESIGN.md`.
- **Ops:** secrets (cursor/HMAC keys, SMTP credentials) outside Git; rate-limit table contention is monitored; SMTP provider chosen with the Datacloud setup.
- **Cost:** no new service in `production-live`; Mailpit is local only.

## Evidence / verification required

Threat model; Argon2id benchmark on the target host recorded before pinning parameters; cookie inspection per environment; session expiry, rotation, revocation and re-authentication tests; CSRF negative tests (missing, wrong, cross-origin, anonymous mutation, pre-session token); enumeration and timing tests; 401/403/404 matrix tests from `AUTHORIZATION_MATRIX.md`; `EMAIL_NOT_VERIFIED` tests; 429 contract tests; token-never-logged test; moderation audit-in-same-transaction tests. Gates G3 and G4.
