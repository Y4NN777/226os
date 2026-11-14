-- Build (S5, US-BUILD-001): projects and contribution needs. Reads only return published projects that
-- are public or owned by the viewer; the owner is always the session subject.

-- name: ProjectURLTaken :one
-- The URL is unique among projects that are not removed (projects_github_url_uq); checked before the
-- GitHub call so a duplicate costs no quota. The insert stays the backstop.
SELECT EXISTS (
  SELECT 1 FROM projects WHERE github_url = sqlc.arg('github_url')::text AND status <> 'removed'
)::boolean;

-- name: InsertProject :one
INSERT INTO projects (owner_id, name, summary, description_md, language, github_url, visibility)
VALUES (sqlc.arg('owner_id')::uuid, sqlc.arg('name')::text, sqlc.arg('summary')::text,
        sqlc.narg('description_md')::text, sqlc.arg('language')::text, sqlc.arg('github_url')::text,
        sqlc.arg('visibility')::text)
RETURNING id;

-- name: InsertProjectTags :exec
INSERT INTO project_tags (project_id, tag_id)
SELECT sqlc.arg('project_id')::uuid, t.id FROM tags t
WHERE t.slug = ANY (sqlc.arg('slugs')::text[]) AND t.status = 'active'
ON CONFLICT DO NOTHING;

-- name: DeleteProjectTags :exec
-- Technologies are replaced as a set; links to hidden tags are kept (moderator state, issue #116).
DELETE FROM project_tags pt
USING tags t
WHERE pt.project_id = sqlc.arg('project_id')::uuid AND pt.tag_id = t.id AND t.status = 'active';

-- name: LockProjectAccess :one
-- The owner and visibility of a published project, locked for the writing transaction. Removed or
-- missing is no row.
SELECT owner_id, visibility FROM projects
WHERE id = sqlc.arg('id')::uuid AND status = 'published'
FOR UPDATE;

-- name: ProjectAccess :one
SELECT owner_id, visibility FROM projects
WHERE id = sqlc.arg('id')::uuid AND status = 'published';

-- name: UpdateProject :exec
UPDATE projects SET
  description_md = CASE WHEN sqlc.arg('set_description')::boolean THEN sqlc.narg('description_md')::text ELSE description_md END,
  visibility = COALESCE(sqlc.narg('visibility')::text, visibility),
  updated_at = now()
WHERE id = sqlc.arg('id')::uuid;

-- name: GetProject :one
-- description falls back to the summary taken from GitHub when the owner wrote none.
SELECT p.id, p.name, COALESCE(p.github_url, '')::text AS github_url,
       left(COALESCE(p.description_md, p.summary), 1000)::text AS description,
       p.language, p.visibility, p.owner_id, pr.username, pr.display_name,
       ARRAY(SELECT t.slug FROM project_tags pt JOIN tags t ON t.id = pt.tag_id
             WHERE pt.project_id = p.id AND t.status = 'active' ORDER BY t.slug)::text[] AS technologies,
       (SELECT count(*) FROM contribution_needs n WHERE n.project_id = p.id AND n.status = 'open')::int AS open_needs,
       p.published_at, p.created_at, p.updated_at
FROM projects p
JOIN profiles pr ON pr.user_id = p.owner_id
WHERE p.id = sqlc.arg('id')::uuid AND p.status = 'published'
  AND (p.visibility = 'public' OR p.owner_id = sqlc.narg('viewer_id')::uuid);

-- name: ListProjects :many
-- Newest first with a UUID tie-breaker (keyset). Public projects use projects_browse_idx; the viewer's
-- own private projects come from projects_owner_idx.
SELECT p.id, p.name, COALESCE(p.github_url, '')::text AS github_url,
       left(COALESCE(p.description_md, p.summary), 1000)::text AS description,
       p.language, p.visibility, p.owner_id, pr.username, pr.display_name,
       ARRAY(SELECT t.slug FROM project_tags pt JOIN tags t ON t.id = pt.tag_id
             WHERE pt.project_id = p.id AND t.status = 'active' ORDER BY t.slug)::text[] AS technologies,
       (SELECT count(*) FROM contribution_needs n WHERE n.project_id = p.id AND n.status = 'open')::int AS open_needs,
       p.published_at, p.created_at, p.updated_at
FROM projects p
JOIN profiles pr ON pr.user_id = p.owner_id
WHERE p.status = 'published'
  AND (p.visibility = 'public' OR p.owner_id = sqlc.narg('viewer_id')::uuid)
  AND (sqlc.narg('q')::text IS NULL
       OR p.search_document @@ websearch_to_tsquery('simple', f_unaccent(sqlc.narg('q')::text)))
  AND (sqlc.narg('technology')::text IS NULL OR EXISTS (
       SELECT 1 FROM project_tags pt JOIN tags t ON t.id = pt.tag_id
       WHERE pt.project_id = p.id AND t.slug = sqlc.narg('technology')::text AND t.status = 'active'))
  AND (sqlc.narg('skill')::text IS NULL OR EXISTS (
       SELECT 1 FROM contribution_needs n
       JOIN need_skills ns ON ns.need_id = n.id
       JOIN tags t ON t.id = ns.tag_id
       WHERE n.project_id = p.id AND n.status = 'open' AND t.slug = sqlc.narg('skill')::text AND t.status = 'active'))
  AND (NOT sqlc.arg('open_needs')::boolean OR EXISTS (
       SELECT 1 FROM contribution_needs n WHERE n.project_id = p.id AND n.status = 'open'))
  AND (sqlc.narg('after_published')::timestamptz IS NULL
       OR (p.published_at, p.id) < (sqlc.narg('after_published')::timestamptz, sqlc.narg('after_id')::uuid))
ORDER BY p.published_at DESC, p.id DESC
LIMIT sqlc.arg('row_limit');

-- name: ProjectRelated :many
-- The visible links of a page of projects, at most 20 per project (a window per project, so one
-- project cannot starve the others): questions and resources that point to it, articles it points to.
SELECT x.project_id::uuid AS project_id, x.kind::text AS kind, x.target_id::uuid AS target_id,
       left(x.title, 200)::text AS title, x.lang::text AS lang
FROM (
  SELECT r.project_id, r.kind, r.target_id, r.title, r.lang,
         row_number() OVER (PARTITION BY r.project_id ORDER BY r.created_at, r.rel_id) AS n
  FROM (
    SELECT cr.tgt_project_id AS project_id, 'question' AS kind, q.id AS target_id, q.title, q.language AS lang,
           cr.created_at, cr.id AS rel_id
    FROM content_relations cr
    JOIN questions q ON q.id = cr.src_question_id AND q.status = 'published'
    WHERE cr.tgt_project_id = ANY (sqlc.arg('project_ids')::uuid[]) AND cr.status = 'active'
    UNION ALL
    SELECT cr.tgt_project_id, 'resource', res.id, res.name, res.summary_language, cr.created_at, cr.id
    FROM content_relations cr
    JOIN resources res ON res.id = cr.src_resource_id AND res.status = 'published'
    JOIN providers pv ON pv.id = res.provider_id AND pv.status = 'published' AND pv.removed_at IS NULL
    WHERE cr.tgt_project_id = ANY (sqlc.arg('project_ids')::uuid[]) AND cr.status = 'active'
    UNION ALL
    SELECT cr.src_project_id, 'article', a.id, a.title, a.language, cr.created_at, cr.id
    FROM content_relations cr
    JOIN learn_articles a ON a.id = cr.tgt_article_id AND a.status = 'published'
    WHERE cr.src_project_id = ANY (sqlc.arg('project_ids')::uuid[]) AND cr.target_type = 'article'
      AND cr.status = 'active'
  ) r
) x
WHERE x.n <= 20
ORDER BY x.project_id, x.n
LIMIT 1020; -- 51 projects (a page plus one) x 20 links

-- The three link writers lock the target (FOR SHARE) and return whether it is usable, so an already
-- existing link (ON CONFLICT DO NOTHING) is not mistaken for a bad target.

-- name: LinkQuestionToProject :one
WITH target AS (
  SELECT q.id FROM questions q WHERE q.id = sqlc.arg('target_id')::uuid AND q.status = 'published' FOR SHARE
), ins AS (
  INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_project_id, created_by)
  SELECT 'question', target.id, 'project', sqlc.arg('project_id')::uuid, sqlc.arg('created_by')::uuid FROM target
  ON CONFLICT DO NOTHING
)
SELECT EXISTS (SELECT 1 FROM target)::boolean;

-- name: LinkResourceToProject :one
WITH target AS (
  SELECT r.id FROM resources r
  JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
  WHERE r.id = sqlc.arg('target_id')::uuid AND r.status = 'published'
  FOR SHARE OF r
), ins AS (
  INSERT INTO content_relations (source_type, src_resource_id, target_type, tgt_project_id, created_by)
  SELECT 'resource', target.id, 'project', sqlc.arg('project_id')::uuid, sqlc.arg('created_by')::uuid FROM target
  ON CONFLICT DO NOTHING
)
SELECT EXISTS (SELECT 1 FROM target)::boolean;

-- name: LinkProjectToArticleTarget :one
WITH target AS (
  SELECT a.id FROM learn_articles a WHERE a.id = sqlc.arg('target_id')::uuid AND a.status = 'published' FOR SHARE
), ins AS (
  INSERT INTO content_relations (source_type, src_project_id, target_type, tgt_article_id, created_by)
  SELECT 'project', sqlc.arg('project_id')::uuid, 'article', target.id, sqlc.arg('created_by')::uuid FROM target
  ON CONFLICT DO NOTHING
)
SELECT EXISTS (SELECT 1 FROM target)::boolean;

-- name: InsertNeed :one
INSERT INTO contribution_needs (project_id, title, description_md, language, created_by)
VALUES (sqlc.arg('project_id')::uuid, sqlc.arg('title')::text, sqlc.arg('description_md')::text,
        sqlc.arg('language')::text, sqlc.arg('created_by')::uuid)
RETURNING id, status, created_at;

-- name: InsertNeedSkills :exec
INSERT INTO need_skills (need_id, tag_id)
SELECT sqlc.arg('need_id')::uuid, t.id FROM tags t
WHERE t.slug = ANY (sqlc.arg('slugs')::text[]) AND t.status = 'active'
ON CONFLICT DO NOTHING;

-- name: ListNeeds :many
-- Newest first with a UUID tie-breaker (contribution_needs_project_idx covers the project filter).
-- viewer_application_* is the caller's OWN application to the need (applications_user_need_uq: at most one
-- row, so the join never multiplies a need); no caller or no application gives NULLs, and no other
-- applicant's row can match.
SELECT n.id, n.project_id, n.title, n.description_md, n.language, n.status, n.created_at,
       ARRAY(SELECT t.slug FROM need_skills ns JOIN tags t ON t.id = ns.tag_id
             WHERE ns.need_id = n.id AND t.status = 'active' ORDER BY t.slug)::text[] AS skills,
       va.id AS viewer_application_id, va.status AS viewer_application_status
FROM contribution_needs n
LEFT JOIN applications va ON va.contribution_need_id = n.id AND va.user_id = sqlc.narg('viewer_id')::uuid
WHERE n.project_id = sqlc.arg('project_id')::uuid
  AND (sqlc.narg('after_created')::timestamptz IS NULL
       OR (n.created_at, n.id) < (sqlc.narg('after_created')::timestamptz, sqlc.narg('after_id')::uuid))
ORDER BY n.created_at DESC, n.id DESC
LIMIT sqlc.arg('row_limit');

-- Applications (US-BUILD-001). Transitions are guarded on the locked row (DATABASE_DESIGN 3.5):
-- pending -> accepted | declined | withdrawn; withdrawn -> pending (the applicant re-applies, the row is
-- reopened); accepted and declined are terminal.

-- name: LockNeedForApply :one
-- The need, its project's owner and visibility; the need row is locked so it cannot change under the
-- application. A removed project or a missing need is no row.
SELECT n.status, n.language, p.owner_id, p.visibility
FROM contribution_needs n
JOIN projects p ON p.id = n.project_id AND p.status = 'published'
WHERE n.id = sqlc.arg('need_id')::uuid
FOR SHARE OF n;

-- name: InsertApplication :one
-- No row returned means the (user, need) pair already exists (applications_user_need_uq).
INSERT INTO applications (contribution_need_id, user_id, message, language)
VALUES (sqlc.arg('need_id')::uuid, sqlc.arg('user_id')::uuid, sqlc.arg('message')::text, sqlc.arg('language')::text)
ON CONFLICT (user_id, contribution_need_id) DO NOTHING
RETURNING id;

-- name: LockOwnApplication :one
SELECT id, status FROM applications
WHERE user_id = sqlc.arg('user_id')::uuid AND contribution_need_id = sqlc.arg('need_id')::uuid
FOR UPDATE;

-- name: ReopenApplication :execrows
UPDATE applications SET status = 'pending', withdrawn_at = NULL, message = sqlc.arg('message')::text,
  language = sqlc.arg('language')::text, updated_at = now()
WHERE id = sqlc.arg('id')::uuid AND status = 'withdrawn';

-- name: LockApplicationForDecision :one
-- The applicant, the status and the project owner, read on the locked application row.
SELECT a.user_id, a.status, p.owner_id
FROM applications a
JOIN contribution_needs n ON n.id = a.contribution_need_id
JOIN projects p ON p.id = n.project_id AND p.status = 'published'
WHERE a.id = sqlc.arg('id')::uuid
FOR UPDATE OF a;

-- name: DecideApplication :execrows
UPDATE applications SET
  status = sqlc.arg('status')::text,
  decided_by = CASE WHEN sqlc.arg('status')::text IN ('accepted', 'declined') THEN sqlc.arg('actor_id')::uuid END,
  decided_at = CASE WHEN sqlc.arg('status')::text IN ('accepted', 'declined') THEN now() END,
  withdrawn_at = CASE WHEN sqlc.arg('status')::text = 'withdrawn' THEN now() END,
  updated_at = now()
WHERE id = sqlc.arg('id')::uuid AND status = 'pending';

-- name: GetApplication :one
SELECT a.id, a.contribution_need_id, n.project_id, a.user_id, pr.username, pr.display_name, a.message,
       a.status, a.created_at, a.decided_at
FROM applications a
JOIN contribution_needs n ON n.id = a.contribution_need_id
JOIN profiles pr ON pr.user_id = a.user_id
WHERE a.id = sqlc.arg('id')::uuid;

-- name: ListProjectApplications :many
-- The owner predicate is part of the query (DATABASE_DESIGN section 7), never a post-filter.
SELECT a.id, a.contribution_need_id, n.project_id, a.user_id, pr.username, pr.display_name, a.message,
       a.status, a.created_at, a.decided_at
FROM applications a
JOIN contribution_needs n ON n.id = a.contribution_need_id
JOIN projects p ON p.id = n.project_id
JOIN profiles pr ON pr.user_id = a.user_id
WHERE n.project_id = sqlc.arg('project_id')::uuid AND p.owner_id = sqlc.arg('owner_id')::uuid
  AND (sqlc.narg('status')::text IS NULL OR a.status = sqlc.narg('status')::text)
  AND (sqlc.narg('after_created')::timestamptz IS NULL
       OR (a.created_at, a.id) < (sqlc.narg('after_created')::timestamptz, sqlc.narg('after_id')::uuid))
ORDER BY a.created_at DESC, a.id DESC
LIMIT sqlc.arg('row_limit');
