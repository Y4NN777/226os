-- Minimal bootstrap query proving :exec generation for the platform area (purge job shape).
-- name: DeleteExpiredIdempotencyKeys :exec
DELETE FROM idempotency_keys WHERE expires_at < $1;
