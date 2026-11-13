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

-- VerifyEmail consumes the token and marks the account verified in one
-- statement: a single statement is its own transaction, so a failure cannot
-- burn the token while leaving the account unverified.
-- name: VerifyEmail :one
WITH c AS (
  UPDATE email_verifications SET consumed_at = $2
  WHERE email_verifications.token_hash = $1 AND consumed_at IS NULL AND expires_at > $2
  RETURNING user_id
), u AS (
  UPDATE users SET email_verified_at = $2, updated_at = $2
  FROM c WHERE users.id = c.user_id AND users.email_verified_at IS NULL
)
SELECT user_id FROM c;

-- name: CreatePasswordReset :exec
INSERT INTO password_resets (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- ResetPassword consumes the token, stores the new hash, revokes every
-- session of the account and burns every other open reset token for that
-- account in one statement: a single statement is its own transaction, so a
-- failure cannot leave a consumed token with the old password still active,
-- and an intercepted older mail cannot reset the account again.
-- name: ResetPassword :one
WITH c AS (
  UPDATE password_resets SET consumed_at = $2
  WHERE password_resets.token_hash = $1 AND consumed_at IS NULL AND expires_at > $2
  RETURNING user_id
), u AS (
  UPDATE user_credentials SET password_hash = $3, password_changed_at = $2
  FROM c WHERE user_credentials.user_id = c.user_id
), s AS (
  UPDATE sessions SET revoked_at = $2, revoked_reason = 'password_reset'
  FROM c WHERE sessions.user_id = c.user_id AND sessions.revoked_at IS NULL
), t AS (
  UPDATE password_resets pr SET consumed_at = $2
  FROM c WHERE pr.user_id = c.user_id AND pr.token_hash <> $1 AND pr.consumed_at IS NULL
)
SELECT user_id FROM c;

-- name: HitRateLimit :one
INSERT INTO rate_limit_buckets (scope, key_hash, window_start, hits)
VALUES ($1, $2, $3, 1)
ON CONFLICT (scope, key_hash, window_start)
DO UPDATE SET hits = rate_limit_buckets.hits + 1
RETURNING hits;

-- name: PurgeRateLimitBuckets :exec
-- Bounded purge of windows that ended long ago (fixed windows never exceed 1 h);
-- uses rate_limit_buckets_window_idx.
DELETE FROM rate_limit_buckets
WHERE ctid IN (
  SELECT b.ctid FROM rate_limit_buckets b WHERE b.window_start < $1 LIMIT 500
);
