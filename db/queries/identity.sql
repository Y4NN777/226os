-- Identity queries for the S1 Store port (ADR-0003). Tokens and passwords are
-- only ever stored hashed; every lookup by token is by hash. All statements are
-- parameterized; the only dynamic value is the keep-id exclusion, which is a
-- bound uuid parameter.

-- CreateAccount writes user + credential + profile atomically: a single
-- statement is its own transaction, so no partial account can exist.
-- name: CreateAccount :one
WITH u AS (
  INSERT INTO users (email) VALUES ($1)
  RETURNING id, email, email_verified_at, role, status, created_at
), c AS (
  INSERT INTO user_credentials (user_id, password_hash)
  SELECT id, $2 FROM u
), p AS (
  INSERT INTO profiles (user_id, username, display_name)
  SELECT id, $3, $4 FROM u
)
SELECT id, email, email_verified_at, role, status, created_at FROM u;

-- name: GetAccountByEmail :one
SELECT u.id, u.email, u.email_verified_at, u.role, u.status, u.created_at,
       c.password_hash, p.username, p.display_name
FROM users u
JOIN user_credentials c ON c.user_id = u.id
JOIN profiles p ON p.user_id = u.id
WHERE u.email = $1;

-- name: GetAccountByID :one
SELECT u.id, u.email, u.email_verified_at, u.role, u.status, u.created_at,
       c.password_hash, p.username, p.display_name
FROM users u
JOIN user_credentials c ON c.user_id = u.id
JOIN profiles p ON p.user_id = u.id
WHERE u.id = $1;

-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, token_hash, csrf_token_hash, idle_expires_at, absolute_expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetSessionByTokenHash :one
SELECT id, user_id, token_hash, csrf_token_hash, created_at, last_seen_at,
       authenticated_at, idle_expires_at, absolute_expires_at, revoked_at, revoked_reason
FROM sessions
WHERE token_hash = $1;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = $3, revoked_reason = $2
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeUserSessions :exec
UPDATE sessions SET revoked_at = $3, revoked_reason = $2
WHERE user_id = $1 AND revoked_at IS NULL
  AND ($4::uuid IS NULL OR id <> $4::uuid);

-- name: TouchSession :exec
UPDATE sessions SET last_seen_at = $2, idle_expires_at = $3
WHERE id = $1 AND revoked_at IS NULL;

-- name: CreateEmailVerification :exec
INSERT INTO email_verifications (user_id, email, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: ConsumeEmailVerification :one
UPDATE email_verifications SET consumed_at = $2
WHERE token_hash = $1 AND consumed_at IS NULL AND expires_at > $2
RETURNING user_id;

-- name: MarkEmailVerified :exec
UPDATE users SET email_verified_at = $2, updated_at = $2
WHERE id = $1 AND email_verified_at IS NULL;

-- name: CreatePasswordReset :exec
INSERT INTO password_resets (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: ConsumePasswordReset :one
UPDATE password_resets SET consumed_at = $2
WHERE token_hash = $1 AND consumed_at IS NULL AND expires_at > $2
RETURNING user_id;

-- name: SetPasswordHash :exec
UPDATE user_credentials SET password_hash = $2, password_changed_at = $3
WHERE user_id = $1;

-- name: HitRateLimit :one
INSERT INTO rate_limit_buckets (scope, key_hash, window_start, hits)
VALUES ($1, $2, $3, 1)
ON CONFLICT (scope, key_hash, window_start)
DO UPDATE SET hits = rate_limit_buckets.hits + 1
RETURNING hits;
