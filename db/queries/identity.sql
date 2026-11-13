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

-- ---------- profiles (S1 profiles slice) ----------

-- ProfileColumns is the shared projection: every profiles column plus the
-- owner's skill slugs, ordered for stable output.
-- name: GetProfileByUserID :one
SELECT p.user_id, p.username, p.display_name, p.intent, p.country_code,
       p.languages, p.github_username, p.bio, p.bio_language, p.open_to_collab,
       p.country_visibility, p.languages_visibility, p.skills_visibility,
       p.github_visibility, p.bio_visibility, p.open_to_collab_visibility,
       p.profile_visibility, p.created_at, p.updated_at,
       COALESCE((SELECT array_agg(t.slug ORDER BY t.slug)
                 FROM user_skills us JOIN tags t ON t.id = us.tag_id
                 WHERE us.user_id = p.user_id AND t.status = 'active'), '{}'::text[])::text[] AS skills
FROM profiles p
WHERE p.user_id = $1;

-- GetProfileByUsername adds the account status and join date the visibility
-- decision needs: non-active accounts must answer like unknown usernames.
-- name: GetProfileByUsername :one
SELECT p.user_id, p.username, p.display_name, p.intent, p.country_code,
       p.languages, p.github_username, p.bio, p.bio_language, p.open_to_collab,
       p.country_visibility, p.languages_visibility, p.skills_visibility,
       p.github_visibility, p.bio_visibility, p.open_to_collab_visibility,
       p.profile_visibility, p.created_at, p.updated_at,
       u.status AS account_status, u.created_at AS joined_at,
       COALESCE((SELECT array_agg(t.slug ORDER BY t.slug)
                 FROM user_skills us JOIN tags t ON t.id = us.tag_id
                 WHERE us.user_id = p.user_id AND t.status = 'active'), '{}'::text[])::text[] AS skills
FROM profiles p
JOIN users u ON u.id = p.user_id
WHERE p.username = $1;

-- ResolveSkillSlugs maps contract skill slugs to tag ids. Only active tags
-- resolve; the service rejects the patch when the count differs.
-- name: ResolveSkillSlugs :many
-- The input array is contract-bounded (maxItems 30); the LIMIT satisfies the
-- bounded-list rule and can only truncate an already-invalid oversized input,
-- which the service count check then rejects.
SELECT id, slug FROM tags WHERE slug = ANY($1::text[]) AND status = 'active'
LIMIT 100;

-- UpdateProfile applies a resolved JSON Merge Patch: every column is guarded
-- by a set-flag so omitted members keep their stored value and explicit nulls
-- clear nullable columns. bio_language follows bio (the pair CHECK requires
-- them set or cleared together; the service derives the language, 'en' by
-- contract decision). Skills are replaced by DeleteUserSkills +
-- InsertUserSkills inside the same transaction, after this UPDATE holds the
-- profile row lock, so concurrent patches serialize on fresh snapshots.
-- name: UpdateProfile :one
UPDATE profiles SET
    display_name          = CASE WHEN $2::boolean THEN $3  ELSE display_name END,
    intent                = CASE WHEN $4::boolean THEN $5  ELSE intent END,
    country_code          = CASE WHEN $6::boolean THEN $7  ELSE country_code END,
    languages             = CASE WHEN $8::boolean THEN $9  ELSE languages END,
    github_username       = CASE WHEN $10::boolean THEN $11 ELSE github_username END,
    bio                   = CASE WHEN $12::boolean THEN $13 ELSE bio END,
    bio_language          = CASE WHEN $14::boolean THEN $15 ELSE bio_language END,
    open_to_collab        = CASE WHEN $16::boolean THEN $17 ELSE open_to_collab END,
    country_visibility    = CASE WHEN $18::boolean THEN $19 ELSE country_visibility END,
    languages_visibility  = CASE WHEN $20::boolean THEN $21 ELSE languages_visibility END,
    skills_visibility     = CASE WHEN $22::boolean THEN $23 ELSE skills_visibility END,
    github_visibility     = CASE WHEN $24::boolean THEN $25 ELSE github_visibility END,
    bio_visibility        = CASE WHEN $26::boolean THEN $27 ELSE bio_visibility END,
    profile_visibility    = CASE WHEN $28::boolean THEN $29 ELSE profile_visibility END,
    updated_at            = now()
  WHERE profiles.user_id = $1
RETURNING profiles.user_id;

-- Skill replacement pair: DeleteUserSkills wipes the user's skill rows and
-- InsertUserSkills adds the resolved set. The repository runs both inside
-- the UpdateProfile transaction, after the profile row lock is held, so
-- each statement sees a fresh snapshot and concurrent patches serialize
-- instead of merging.
-- name: DeleteUserSkills :exec
DELETE FROM user_skills WHERE user_id = $1;

-- name: InsertUserSkills :exec
INSERT INTO user_skills (user_id, tag_id)
SELECT $1, unnest($2::uuid[])
ON CONFLICT DO NOTHING;

-- name: PurgeRateLimitBuckets :exec
-- Bounded purge of windows that ended long ago (fixed windows never exceed 1 h);
-- uses rate_limit_buckets_window_idx.
DELETE FROM rate_limit_buckets
WHERE ctid IN (
  SELECT b.ctid FROM rate_limit_buckets b WHERE b.window_start < $1 LIMIT 500 FOR UPDATE SKIP LOCKED
);

-- name: CountRateLimit :one
SELECT hits FROM rate_limit_buckets
WHERE scope = $1 AND key_hash = $2 AND window_start = $3;
