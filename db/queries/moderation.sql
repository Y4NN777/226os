-- Moderator writes. Callers execute the state change and audit insert through the same transaction.

-- name: LockModerationQuestion :one
SELECT status FROM questions WHERE id = $1 FOR UPDATE;
-- name: HideModerationQuestion :exec
UPDATE questions SET status = 'removed', removed_at = now(), updated_at = now() WHERE id = $1;
-- name: RestoreModerationQuestion :exec
UPDATE questions SET status = 'published', removed_at = NULL, updated_at = now() WHERE id = $1;

-- name: LockModerationAnswer :one
SELECT status FROM answers WHERE id = $1 FOR UPDATE;
-- name: LockModerationAnswerQuestion :one
-- Match the accept path's question-before-answer lock order to prevent a deadlock.
SELECT q.id
FROM questions AS q
JOIN answers AS a ON a.question_id = q.id
WHERE a.id = $1
FOR UPDATE OF q;
-- name: HideModerationAnswer :exec
WITH hidden AS (
  UPDATE answers SET status = 'removed', removed_at = now(), updated_at = now()
  WHERE answers.id = $1
  RETURNING question_id
)
UPDATE questions AS q
SET answer_count = GREATEST(q.answer_count - 1, 0),
    accepted_answer_id = CASE WHEN q.accepted_answer_id = $1 THEN NULL ELSE q.accepted_answer_id END,
    updated_at = now()
FROM hidden
WHERE q.id = hidden.question_id;
-- name: RestoreModerationAnswer :exec
WITH restored AS (
  UPDATE answers SET status = 'published', removed_at = NULL, updated_at = now()
  WHERE answers.id = $1
  RETURNING question_id
)
UPDATE questions AS q
SET answer_count = q.answer_count + 1, updated_at = now()
FROM restored
WHERE q.id = restored.question_id;

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
