-- Shared fixed-window rate limiter (platform/ratelimit). The table is rate_limit_buckets
-- (migration 00004). Identity keeps its own HitRateLimit query in identity.sql.

-- name: HitRateBucket :one
INSERT INTO rate_limit_buckets (scope, key_hash, window_start, hits)
VALUES ($1, $2, $3, 1)
ON CONFLICT (scope, key_hash, window_start)
DO UPDATE SET hits = rate_limit_buckets.hits + 1
RETURNING hits;

-- name: PurgeRateBuckets :exec
-- Bounded: at most 500 expired buckets per call, so one hit never runs an unbounded delete.
DELETE FROM rate_limit_buckets
WHERE ctid IN (
  SELECT b.ctid FROM rate_limit_buckets b WHERE b.window_start < $1 LIMIT 500
);
