# security-authz — session, CSRF, rate-limit patterns

Illustrative shapes. The numeric values are fixed by `docs/decisions/0003-authentication-sessions-csrf-rate-limits.md`; take them from there, not from this file.

## Sessions table (shape per `docs/spec-kit/AUTHORIZATION_MATRIX.md:10`)
```sql
CREATE TABLE sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  token_hash bytea NOT NULL CONSTRAINT sessions_token_hash_uq UNIQUE,   -- sha256(token)
  csrf_hash bytea NOT NULL,                                               -- sha256(csrf token)
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,                                        -- absolute lifetime
  revoked_at timestamptz,
  revoked_reason text CONSTRAINT sessions_revoked_reason_ck CHECK (revoked_reason IN ('logout','rotation','security_event','password_reset')),
  ua_summary text CONSTRAINT sessions_ua_len_ck CHECK (length(ua_summary) <= 200)
);
```
Lookup: `WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now() AND last_seen_at > now() - $idle`. Throttle `last_seen_at` updates (e.g. once per minute) to limit write load — UNVERIFIED tuning.

## Token issue (Go sketch)
```go
func newToken() (raw string, hash [32]byte, err error) {
    b := make([]byte, 32)                       // 256 bits
    if _, err = rand.Read(b); err != nil { return "", hash, fmt.Errorf("token: %w", err) }
    raw = base64.RawURLEncoding.EncodeToString(b)
    return raw, sha256.Sum256([]byte(raw)), nil
}
```
Sign-in: verify password -> in one tx revoke any pre-auth session, insert new session row -> set cookie -> return CSRF token in body. A fast hash (SHA-256) is fine for 256-bit random tokens; Argon2id is for low-entropy passwords only.

## Argon2id (Go sketch)
```go
// params pin per ADR; OWASP floor: m=19 MiB (19456 KiB), t=2, p=1
key := argon2.IDKey([]byte(pw), salt, p.Time, p.MemoryKiB, p.Threads, 32)
// store: $argon2id$v=19$m=19456,t=2,p=1$<b64 salt>$<b64 key>
```
On successful login, if stored params < current params, rehash and update in the same tx.
Source: https://pkg.go.dev/golang.org/x/crypto/argon2 · https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html

## CSRF middleware order
1. Safe methods (GET/HEAD/OPTIONS) pass, and must not mutate.
2. Reject if `Sec-Fetch-Site` = `cross-site`; if absent, require `Origin` (or `Referer` host) to exactly equal the configured public origin; never accept `null`.
3. Require session; compare `sha256(header token)` with `csrf_hash` via `subtle.ConstantTimeCompare`.
4. Failure -> 403 problem `csrf-failed` (`code: CSRF_FAILED`), log route + request_id only.
Pre-session endpoints (sign-in, registration, reset request) have no session token: ADR-0003 decides they use a pre-session CSRF token (Origin/Fetch-Metadata checks and rate limits still apply).
Source: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html

## PostgreSQL rate limit (G0-AUTH-004)
```sql
CREATE TABLE rate_limit_counters (
  scope text NOT NULL,            -- 'signin_ip', 'signin_account', 'reset_ip', 'write_user', 'github_import'
  key_hash bytea NOT NULL,        -- sha256(ip or normalized email) — no raw PII
  window_start timestamptz NOT NULL,
  count integer NOT NULL,
  PRIMARY KEY (scope, key_hash, window_start)
);
-- name: HitRateLimit :one
INSERT INTO rate_limit_counters (scope, key_hash, window_start, count)
VALUES ($1, $2, date_bin($3::interval, now(), 'epoch'), 1)
ON CONFLICT (scope, key_hash, window_start) DO UPDATE SET count = rate_limit_counters.count + 1
RETURNING count;
```
- `date_bin` requires PostgreSQL 14+ — UNVERIFIED against pinned major.
- Fail closed for auth scopes on DB error (503/429 per ADR). Purge rows older than the longest window periodically.
- Hot-row contention under attack is the known cost (`docs/archive/spec-kit/G0_DECISIONS.md:12`); measure in G6.
- Client IP only from trusted-proxy-aware extraction (Traefik), never raw `X-Forwarded-For`.

## Authorization matrix test skeleton
```go
cases := []struct{ name string; actor actor; method, path string; want int; mustNotChange bool }{
    {"anon accept", anon, "POST", "/api/v1/questions/{q}/accept", 401, true},
    {"bob accept alice q", bob, "POST", "/api/v1/questions/{q}/accept", 403, true},
    {"alice accept own q", alice, "POST", "/api/v1/questions/{q}/accept", 200, false},
    {"charlie lists apps", charlie, "GET", "/api/v1/projects/{p}/applications", 403, true}, // or 404 per non-disclosure ADR
}
```
Each case: fresh fixtures on real PG, valid CSRF for authenticated actors, assert status + problem `code` + DB state unchanged when denied.
