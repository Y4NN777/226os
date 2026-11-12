---
name: security-authz
description: 226os authentication and authorization (first-party email/password, Argon2id, opaque hashed sessions in PostgreSQL, cookie flags, synchronizer CSRF, PostgreSQL rate limits, anonymous/member/moderator/owner + resource ownership/IDOR, negative authz tests, error redaction, OWASP ASVS/Top 10). Use when touching src/api/internal/identity/**, src/api/**/auth*/**, src/web/src/features/identity/**, or any endpoint's permission check.
---
# Security, Authentication & Authorization

## Scope and triggers
- `src/api/internal/identity/**`, `src/api/**/auth*/**`, session/CSRF/rate-limit middleware, every service-level permission check, `src/web/src/features/identity/**`.
- Status: G0-AUTH-001..004 and G0-AUTHZ-001 are ACCEPTED WITH CONDITIONS (`docs/decisions/0001-g0-conditional-baseline.md:16`); the matrix is PROPOSED (`docs/spec-kit/AUTHORIZATION_MATRIX.md:3`); ADR-OPEN-003 must close before TASK-SEC-001 (`docs/decisions/README.md:32`, `docs/spec-kit/AUTHORIZATION_MATRIX.md:55`). Timeouts, Argon2 profile and rate windows are proposals, not accepted values.

## Project rules (binding)
- Threat model before any sensitive surface; least privilege, server auth, CSRF, rate limit, anti-IDOR, XSS/SSRF, input validation, non-revealing errors (`docs/RULES.md:24`).
- No key, credential, token or password in Git, prompts, logs, traces, screenshots, fixtures (`docs/RULES.md:25`); never log secrets (`docs/DEV.md:48`).
- Auth: first-party email/password, opaque server sessions in PostgreSQL, no third-party IdP in P0; Argon2id; token random, hashed at rest, rotated on sign-in/privilege change, revoked on logout/security events (G0-AUTH-001 `docs/spec-kit/G0_DECISIONS.md:9`; `docs/spec-kit/AUTHORIZATION_MATRIX.md:8-11`).
- Session row stores only hash + user, created, last seen, expiry, revoked, UA/device summary, security-event reason (`docs/spec-kit/AUTHORIZATION_MATRIX.md:10`); session hash unique; revoked/expired unusable (`docs/spec-kit/DATABASE_DESIGN.md:13`).
- Cookie: `HttpOnly`, `Secure` in deployed envs, `SameSite=Lax`, narrow host/path, no secrets in URLs; synchronizer CSRF token on every state-changing request incl. JSON (G0-AUTH-002 `docs/spec-kit/G0_DECISIONS.md:10`, `docs/spec-kit/AUTHORIZATION_MATRIX.md:12`). No bearer tokens in P0 (`docs/spec-kit/API_CONTRACT.md:11`).
- Proposed lifetimes: idle 24h, absolute 30d, reset token 30min, verification 24h — defaults for approval only (`docs/spec-kit/AUTHORIZATION_MATRIX.md:13`).
- Generic responses for sign-in/registration/reset (no enumeration); IP + account limits on auth endpoints (`docs/spec-kit/AUTHORIZATION_MATRIX.md:14`). Recovery + email verification required before production; reset tokens single-use, short-lived, hashed, never logged (G0-AUTH-003 `docs/spec-kit/G0_DECISIONS.md:11`).
- Rate limits PostgreSQL-backed (no Valkey without ADR), fail closed for sensitive auth actions (G0-AUTH-004 `docs/spec-kit/G0_DECISIONS.md:12`; G0-OPS-001 `:22`).
- Roles: `anonymous`, `member`, `moderator`, `owner`; owner = owner of that specific resource, no global admin (`docs/spec-kit/AUTHORIZATION_MATRIX.md:4`; G0-AUTHZ-001 `docs/spec-kit/G0_DECISIONS.md:13`).
- Subject always from server session, never client-supplied owner ID (`docs/spec-kit/AUTHORIZATION_MATRIX.md:46`); ownership/visibility/status/relation-target checked in the same transaction/boundary as the mutation/read (`:47`); denials don't reveal private-object existence (`:48`); moderator access narrow, audited, never from a frontend flag (`:49`); CSRF + same-origin checks on all browser mutations (`:50`); retries safe via DB uniqueness + idempotency keys (`:51`).
- Per-endpoint allow/deny + required negative cases: `docs/spec-kit/AUTHORIZATION_MATRIX.md:21-42`.
- Auth/authz centralized, access verified closest to the business operation (`docs/architecture/ARCHITECTURE.md:23`). Log forbidden attempts without secrets or unneeded personal data (`docs/SPEC.md:104`).
- Mandatory security tests: session expiry/revocation/rotation, cookie flags, CSRF on mutations, rate limit, validation; per-route ownership (author vs non-author, collaborator vs third party, anonymous); XSS on Markdown/snippets; SQLi on search; SSRF on GitHub import (`docs/TEST.md:35-37`). G4 gate (`AGENTS.md:50`).
- Auth/security exceptions need human approval (`AGENTS.md:130`); no exception for exposed secrets (`docs/TEST.md:52`).
- Doc conflicts: "collaborator" appears in TASK-SEC-001 AC (`docs/spec-kit/TASKS.md:73`), `docs/TEST.md:36`, `docs/spec-kit/API_CONTRACT.md:62` but is not a role in the matrix (`docs/spec-kit/AUTHORIZATION_MATRIX.md:4`) — UNVERIFIED; treat as a resource relation (e.g. accepted applicant) only once ADR defines it. `docs/spec-kit/AUTHORIZATION_MATRIX.md:14` says rate windows are "in G0_DECISIONS.md" but `docs/spec-kit/G0_DECISIONS.md:12` only says "documented windows" — dangling reference, windows are OPEN. 403 (`docs/SPEC.md:27`) vs non-disclosure (`docs/spec-kit/AUTHORIZATION_MATRIX.md:48`): 403 for visible objects, 404 for private ones.

## Expert practices
- Argon2id via `golang.org/x/crypto/argon2.IDKey`; OWASP floor m=19 MiB, t=2, p=1 — final profile pin per ADR after benchmark on target host; 16-byte random salt; store PHC-style string with params so they can be upgraded on login. [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html), [x/crypto/argon2](https://pkg.go.dev/golang.org/x/crypto/argon2)
- Compare derived keys with `crypto/subtle.ConstantTimeCompare`; on unknown email still run one Argon2id against a dummy hash to equalize timing. [crypto/subtle](https://pkg.go.dev/crypto/subtle), [OWASP Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- Bound password length (max pin per ADR) to cap hashing cost; check against breached-password list if approved. [OWASP Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html), [NIST SP 800-63B](https://pages.nist.gov/800-63-4/sp800-63b.html)
- Session token: ≥128 bits from `crypto/rand` (32 bytes, base64url); store SHA-256(token) only; look up by hash; check `revoked_at IS NULL AND expires_at > now() AND idle limit`; issue new token on login and privilege change, revoke the old row (fixation). [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html), [crypto/rand](https://pkg.go.dev/crypto/rand)
- Prefer `__Host-` cookie prefix (forces `Secure`, `Path=/`, no `Domain`); tension: `Secure` only "in deployed environments" (`docs/spec-kit/AUTHORIZATION_MATRIX.md:12`) — browsers treat `localhost` as secure for `Secure` cookies (MDN), so `__Host-` may work locally too; otherwise choose name by config. [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html#cookies), [MDN Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie)
- Logout and timeouts enforced server-side (revoke row), client clears cookie. Note: proposed 24h idle exceeds OWASP's 15–30 min low-risk guidance — surface in ADR-OPEN-003. [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html#session-expiration)
- CSRF: per-session synchronizer token stored server-side (hashed with session), sent by the SPA in a custom header, compared constant-time; plus `Sec-Fetch-Site`/`Origin` verification (reject `cross-site`, never trust `Origin: null`). SameSite=Lax alone is not sufficient. No state-changing `GET`. [OWASP CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- Go `net/http.CrossOriginProtection` (Fetch-Metadata based) may serve the origin layer if the pinned Go version ships it — UNVERIFIED for chosen toolchain; it does not replace the synchronizer token mandated by G0-AUTH-002. [net/http](https://pkg.go.dev/net/http#CrossOriginProtection)
- Authorization: deny by default; check on every request at the service layer with the resource loaded in the same tx; centralize policy functions (`policy.CanEditQuestion(subj, q)`); test them exhaustively. [OWASP Authorization](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html), [OWASP IDOR Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html)
- Scope queries by owner (`WHERE id=$1 AND owner_id=$2`) or load-then-check inside the tx; UUIDs are not authorization. [OWASP API1:2023 BOLA](https://api-security.owasp.org/editions/2023/en/0xa1-broken-object-level-authorization/)
- Rate limits: fixed/sliding window counters in a PG table keyed `(scope, key_hash, window_start)` with `INSERT ... ON CONFLICT DO UPDATE SET count = count+1 RETURNING count`; key on account and client IP from trusted proxy only; 429 + `Retry-After`; purge old windows. [OWASP Credential Stuffing](https://cheatsheetseries.owasp.org/cheatsheets/Credential_Stuffing_Prevention_Cheat_Sheet.html), [PG INSERT ON CONFLICT](https://www.postgresql.org/docs/current/sql-insert.html)
- Errors: generic auth failures; problem+json without internals; log `request_id`, subject ID, route, decision — never password, token, cookie, CSRF value, reset link, full email. [OWASP Logging](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html), [OWASP Error Handling](https://cheatsheetseries.owasp.org/cheatsheets/Error_Handling_Cheat_Sheet.html)
- Verify against OWASP ASVS (version pin per ADR) chapters on authentication, session management, access control; map findings to OWASP Top 10 (A01 Broken Access Control, A07 Identification & Authentication Failures in the 2021 list). [ASVS](https://owasp.org/www-project-application-security-verification-standard/), [OWASP Top 10](https://owasp.org/Top10/)
- Frontend: never store tokens in `localStorage`; rely on HttpOnly cookie; obtain CSRF token from an API response (endpoint UNVERIFIED, define in OpenAPI) and attach via generated client; UI role flags only hide controls, never authorize. [OWASP HTML5 Security](https://cheatsheetseries.owasp.org/cheatsheets/HTML5_Security_Cheat_Sheet.html)

## SOLID / design notes for this layer
- `identity` owns users, credentials, sessions, rate limits; exports a narrow `Subject{UserID, Roles}` + `SubjectFrom(ctx)`; other modules never read sessions tables.
- Middleware only authenticates (session -> Subject) and enforces CSRF/rate limits; authorization lives in each module's service via pure policy functions (SRP, testable without HTTP).
- `PasswordHasher`, `TokenGenerator`, `Clock` are interfaces owned by identity service so tests inject deterministic fakes (DIP).
- Moderator actions go through an explicit `moderation` path that writes an audit record in the same tx.

## TDD & tests for this layer
- RED first: the SPEC negatives — Bob cannot accept on Alice's question (403, no change; `docs/SPEC.md:24-28`), Charlie cannot list applications (`docs/SPEC.md:52-55`).
- Table-driven matrix test per endpoint: rows = {anonymous, member, owner, other member, moderator} × {allowed, denied, status, DB unchanged}; derive rows from `docs/spec-kit/AUTHORIZATION_MATRIX.md:21-42`.
- Session: login rotates ID (old token rejected); logout revokes; idle and absolute expiry via fake clock; revoked token rejected; cannot revoke another user's session.
- Cookie inspection: `HttpOnly`, `Secure` (deployed config), `SameSite=Lax`, `Path`, no `Domain`.
- CSRF: missing/wrong/other-session token -> 403; cross-site `Sec-Fetch-Site`/foreign `Origin` -> rejected; GET never mutates.
- Enumeration: identical status/body/timing class for unknown email vs wrong password vs reset of unknown email.
- Rate limit: N+1 attempt -> 429 + `Retry-After`; per-account and per-IP; DB failure -> fail closed on auth.
- Redaction: grep captured logs in tests for password/token/cookie values -> none.
- Real PG integration for session and limit tables; `-race` for concurrent logins/limits.

## Anti-patterns (reject in review)
- Owner ID from request body/query; authorization only in frontend or only in middleware by route pattern.
- Plaintext or reversible session/reset tokens in DB; tokens in URLs or logs; JWT/bearer in P0.
- Fast hashes (SHA-*, MD5) or bcrypt for new passwords; params hard-coded without upgrade path.
- Distinct messages "user not found" vs "wrong password"; reset endpoint revealing account existence.
- CSRF exempt JSON endpoints; relying on SameSite or CORS alone; `Origin: null` allowlisted.
- Trusting `X-Forwarded-For` from anywhere (chi `RealIP` is spoofable; deprecation status UNVERIFIED, check the chi changelog for the pinned version).
- Hidden moderator bypass; global admin role; unaudited moderator action.
- Rate limiter failing open on DB error for sign-in/reset.

## Done checklist
- [ ] Threat model (STRIDE-light) for the surface recorded in the task evidence.
- [ ] ADR-OPEN-003 values used (timeouts, Argon2 profile, rate windows); none invented.
- [ ] Matrix rows implemented as tests incl. every "Required negative case".
- [ ] Session lifecycle, cookie, CSRF, enumeration, rate-limit, redaction tests green on real PG.
- [ ] SAST/SCA/secret scan run; results attached (G4, `docs/TEST.md:47`).
- [ ] No secret in code, fixtures, logs, or screenshots.

## References
- Deep dive: [references/session-csrf-ratelimit.md](references/session-csrf-ratelimit.md)
- https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html · https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html · https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
- https://owasp.org/www-project-application-security-verification-standard/ · https://owasp.org/Top10/ · https://api-security.owasp.org/
