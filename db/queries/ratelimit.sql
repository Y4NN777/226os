-- Shared fixed-window rate limiter (platform/ratelimit). The table is rate_limit_buckets
-- (migration 00004). Identity keeps its own HitRateLimit query in identity.sql.

-- name: HitRateBucket :one
INSERT INTO rate_limit_buckets (scope, key_hash, window_start, hits)
VALUES ($1, $2, $3, 1)
ON CONFLICT (scope, key_hash, window_start)
DO UPDATE SET hits = rate_limit_buckets.hits + 1
RETURNING hits;

-- name: GiveBackRateBucket :exec
-- Releases one hit of a fixed-window bucket in a single statement (the assistant gives back the fair-share
-- hit of a call the global guard refused). rate_limit_buckets_hits_ck requires hits > 0, so the last hit
-- deletes the row; the two branches touch disjoint rows. Under a concurrent change a give-back can be skipped,
-- never doubled: a stale count only errs on the strict side. Same statement as identity's GiveBackRateLimit.
WITH dec AS (
  UPDATE rate_limit_buckets SET hits = hits - 1
  WHERE scope = $1 AND key_hash = $2 AND window_start = $3 AND hits > 1
  RETURNING 1
)
DELETE FROM rate_limit_buckets AS b
WHERE b.scope = $1 AND b.key_hash = $2 AND b.window_start = $3 AND b.hits = 1;

-- name: PurgeRateBuckets :exec
-- Bounded: at most 500 expired buckets per call, so one hit never runs an unbounded delete.
DELETE FROM rate_limit_buckets
WHERE ctid IN (
  SELECT b.ctid FROM rate_limit_buckets b WHERE b.window_start < $1 LIMIT 500
);
