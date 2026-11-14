# Moderation operations

This runbook is the operational fallback for the P0 moderation API. ADR-0003 limits moderators to hiding or restoring content and resolving error reports. Every successful moderator action, including the one-time role grant below, must write `audit_log` in the same transaction as the state change. Do not edit another member's content.

Run these statements only through an authenticated administrative PostgreSQL session. Set the values explicitly in `psql`; never put an e-mail address, password, token or connection string in the repository or shell history.

## Promote the owner account once

No moderator is seeded. Before the production demo, set the owner's public username and a change-ticket or deployment identifier:

```sql
\set owner_username 'replace-with-public-username'
\set request_id 'replace-with-change-id'

BEGIN;

WITH promoted AS (
  UPDATE users AS u
     SET role = 'moderator', updated_at = now()
    FROM profiles AS p
   WHERE p.user_id = u.id
     AND p.username = :'owner_username'
     AND u.status = 'active'
     AND u.role = 'member'
  RETURNING u.id
)
INSERT INTO audit_log (
  actor_id, actor_role, action, target_type, target_id, reason, request_id
)
SELECT id, 'moderator', 'user.promote', 'user', id,
       'One-time owner moderator promotion', :'request_id'
  FROM promoted;

-- Both counts must be exactly one before committing.
SELECT u.id, p.username, u.role
  FROM users AS u
  JOIN profiles AS p ON p.user_id = u.id
 WHERE p.username = :'owner_username'
 FOR UPDATE;

SELECT count(*) = 1 AS audit_written
  FROM audit_log
 WHERE action = 'user.promote'
   AND target_type = 'user'
   AND request_id = :'request_id';
```

The first check must show the intended username with role `moderator`, and `audit_written` must be true. Otherwise issue `ROLLBACK;` and investigate. Only after those checks, type `COMMIT;` separately. Re-running the step is intentionally a no-op, because it accepts only a current `member`; do not manufacture a second audit row.

## List open error reports

This query omits `reporter_id` and any profile or account join. The reporter's identity must not appear in operator notes, screenshots or responses to the reported member.

```sql
SELECT id, resource_id, kind, message, status, created_at
  FROM error_reports
 WHERE status IN ('received', 'triaged')
 ORDER BY created_at, id;
```

Resolve or reject reports through the moderation API when it is available so authorization, rate limiting and the audit transaction remain centralized.

## Emergency hide

Prefer the API. Use SQL only when the API is unavailable and leaving the item visible is a greater risk. Record the incident identifier as `request_id` and a specific reason of at most 500 characters. Never interpolate a table name: select the fixed statement for the target type.

For `question`, `answer`, `resource`, `project` or `article`, use the matching table (`questions`, `answers`, `resources`, `projects`, `learn_articles`) and matching audit target type. This question example is deliberately guarded so an already removed item does not create an audit row:

```sql
\set moderator_id '00000000-0000-0000-0000-000000000000'
\set target_id '00000000-0000-0000-0000-000000000000'
\set reason 'replace-with-incident-reason'
\set request_id 'replace-with-incident-id'

BEGIN;

WITH hidden AS (
  UPDATE questions
     SET status = 'removed', removed_at = now(), updated_at = now()
   WHERE id = :'target_id'::uuid
     AND status = 'published'
  RETURNING id
)
INSERT INTO audit_log (
  actor_id, actor_role, action, target_type, target_id, reason, request_id
)
SELECT :'moderator_id'::uuid, 'moderator', 'question.hide', 'question', id,
       :'reason', :'request_id'
  FROM hidden;

SELECT status = 'removed' AND removed_at IS NOT NULL AS target_hidden
  FROM questions
 WHERE id = :'target_id'::uuid;

SELECT count(*) = 1 AS audit_written
  FROM audit_log
 WHERE target_type = 'question'
   AND target_id = :'target_id'::uuid
   AND action = 'question.hide'
   AND request_id = :'request_id';
```

For a relation, use its `hidden` state and corresponding audit action:

```sql
BEGIN;

WITH hidden AS (
  UPDATE content_relations
     SET status = 'hidden'
   WHERE id = :'target_id'::uuid
     AND status = 'active'
  RETURNING id
)
INSERT INTO audit_log (
  actor_id, actor_role, action, target_type, target_id, reason, request_id
)
SELECT :'moderator_id'::uuid, 'moderator', 'relation.hide', 'relation', id,
       :'reason', :'request_id'
  FROM hidden;

SELECT status = 'hidden' AS target_hidden
  FROM content_relations
 WHERE id = :'target_id'::uuid;

SELECT count(*) = 1 AS audit_written
  FROM audit_log
 WHERE target_type = 'relation'
   AND target_id = :'target_id'::uuid
   AND action = 'relation.hide'
   AND request_id = :'request_id';
```

In either transaction, `target_hidden` and `audit_written` must both be true. Otherwise issue `ROLLBACK;`. Only after both checks, type `COMMIT;` separately.
