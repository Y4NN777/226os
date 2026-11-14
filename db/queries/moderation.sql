-- Moderator writes. Callers execute the state change and audit insert through the same transaction.

-- name: LockModerationQuestion :one
SELECT status FROM questions WHERE id = $1 FOR UPDATE;
-- name: HideModerationQuestion :exec
UPDATE questions SET status = 'removed', removed_at = now(), updated_at = now() WHERE id = $1;
-- name: RestoreModerationQuestion :exec
UPDATE questions SET status = 'published', removed_at = NULL, updated_at = now() WHERE id = $1;

-- name: LockModerationAnswer :one
SELECT status FROM answers WHERE id = $1 FOR UPDATE;
-- name: HideModerationAnswer :exec
UPDATE answers SET status = 'removed', removed_at = now(), updated_at = now() WHERE id = $1;
-- name: RestoreModerationAnswer :exec
UPDATE answers SET status = 'published', removed_at = NULL, updated_at = now() WHERE id = $1;

-- name: LockModerationResource :one
SELECT status FROM resources WHERE id = $1 FOR UPDATE;
-- name: HideModerationResource :exec
UPDATE resources SET status = 'removed', removed_at = now(), updated_at = now() WHERE id = $1;
-- name: RestoreModerationResource :exec
UPDATE resources SET status = 'published', removed_at = NULL, updated_at = now() WHERE id = $1;

-- name: LockModerationProject :one
SELECT status FROM projects WHERE id = $1 FOR UPDATE;
-- name: HideModerationProject :exec
UPDATE projects SET status = 'removed', removed_at = now(), updated_at = now() WHERE id = $1;
-- name: RestoreModerationProject :exec
UPDATE projects SET status = 'published', removed_at = NULL, updated_at = now() WHERE id = $1;

-- name: LockModerationArticle :one
SELECT status FROM learn_articles WHERE id = $1 FOR UPDATE;
-- name: HideModerationArticle :exec
UPDATE learn_articles SET status = 'removed', removed_at = now(), updated_at = now() WHERE id = $1;
-- name: RestoreModerationArticle :exec
UPDATE learn_articles SET status = 'published', removed_at = NULL, updated_at = now() WHERE id = $1;

-- name: LockModerationRelation :one
SELECT status FROM content_relations WHERE id = $1 FOR UPDATE;
-- name: HideModerationRelation :exec
UPDATE content_relations SET status = 'hidden' WHERE id = $1;
-- name: RestoreModerationRelation :exec
UPDATE content_relations SET status = 'active' WHERE id = $1;

-- name: LockModerationErrorReport :one
SELECT id, resource_id, kind, status, created_at FROM error_reports WHERE id = $1 FOR UPDATE;

-- name: ResolveModerationErrorReport :one
UPDATE error_reports
SET status = sqlc.arg('status'), handled_by = sqlc.arg('actor_id'), handled_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, resource_id, kind, status, created_at;

-- name: InsertModerationAudit :exec
INSERT INTO audit_log (actor_id, actor_role, action, target_type, target_id, reason, request_id)
VALUES (sqlc.arg('actor_id'), 'moderator', sqlc.arg('action'), sqlc.arg('target_type'),
        sqlc.arg('target_id'), NULLIF(sqlc.arg('reason')::text, ''), NULLIF(sqlc.arg('request_id')::text, ''));
