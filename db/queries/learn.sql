-- Learn reads and article creation (S4, US-LEARN-001). Only published articles and series are public.
-- A draft or archived article is visible to its author only (viewer_id); removed content never leaks.
-- The byline shows username and display name whatever profile_visibility says (ADR-0005 point 15);
-- the country is shown only when both the profile and the country are public ('' otherwise).
-- Lists are newest first with a (published_at, id) keyset.

-- name: ListLearnArticles :many
SELECT a.id, a.kind, a.title, COALESCE(a.summary, '')::text AS summary, a.language, a.level, a.published_at,
       a.author_id, pr.username AS author_username, pr.display_name AS author_display_name,
       COALESCE(CASE WHEN pr.profile_visibility = 'public' AND pr.country_visibility = 'public'
             THEN pr.country_code END, '')::text AS author_country
FROM learn_articles a
JOIN profiles pr ON pr.user_id = a.author_id
WHERE a.status = 'published'
  AND (sqlc.narg('kind')::text IS NULL OR a.kind = sqlc.narg('kind'))
  AND (sqlc.narg('topic')::text IS NULL OR a.topic_code = sqlc.narg('topic'))
  AND (sqlc.narg('level')::text IS NULL OR a.level = sqlc.narg('level'))
  AND (sqlc.narg('lang')::text IS NULL OR a.language = sqlc.narg('lang'))
  AND (sqlc.narg('q')::text IS NULL OR a.search_document @@ websearch_to_tsquery('simple', f_unaccent(sqlc.narg('q')::text)))
  AND (sqlc.narg('after_published')::timestamptz IS NULL
       OR (a.published_at, a.id) < (sqlc.narg('after_published')::timestamptz, sqlc.narg('after_id')::uuid))
ORDER BY a.published_at DESC, a.id DESC
LIMIT sqlc.arg('row_limit');

-- name: GetLearnArticle :one
-- A published article, or the viewer's own draft or archived article. Removed is never returned.
SELECT a.id, a.kind, a.title, COALESCE(a.summary, '')::text AS summary, a.body_md, a.language, a.level,
       a.status, a.published_at, a.updated_at,
       a.author_id, pr.username AS author_username, pr.display_name AS author_display_name,
       COALESCE(CASE WHEN pr.profile_visibility = 'public' AND pr.country_visibility = 'public'
             THEN pr.country_code END, '')::text AS author_country
FROM learn_articles a
JOIN profiles pr ON pr.user_id = a.author_id
WHERE a.id = sqlc.arg('id')::uuid
  AND (a.status = 'published'
       OR (a.status IN ('draft', 'archived') AND a.author_id = sqlc.narg('viewer_id')::uuid));

-- name: LearnArticleTags :many
-- Active tags only: a hidden tag is the moderator tool (DATABASE_DESIGN tags).
SELECT lat.article_id, t.slug
FROM learn_article_tags lat
JOIN tags t ON t.id = lat.tag_id AND t.status = 'active'
WHERE lat.article_id = ANY (sqlc.arg('ids')::uuid[])
ORDER BY lat.article_id, t.slug
LIMIT 250;

-- name: LearnArticleRelated :many
-- Relations are stored canonically with the article as the target (question, resource or project ->
-- article). Only active relations to published, public items are returned.
SELECT cr.source_type AS link_type, cr.source_id AS link_id,
       COALESCE(q.title, res.name, proj.name)::text AS title,
       COALESCE(q.language, res.summary_language, proj.language)::text AS lang
FROM content_relations cr
LEFT JOIN questions q ON q.id = cr.src_question_id AND q.status = 'published'
LEFT JOIN resources res ON res.id = cr.src_resource_id AND res.status = 'published'
  AND EXISTS (SELECT 1 FROM providers p WHERE p.id = res.provider_id AND p.status = 'published' AND p.removed_at IS NULL)
LEFT JOIN projects proj ON proj.id = cr.src_project_id AND proj.status = 'published' AND proj.visibility = 'public'
WHERE cr.tgt_article_id = sqlc.arg('article_id')::uuid
  AND cr.status = 'active'
  AND COALESCE(q.id, res.id, proj.id) IS NOT NULL
ORDER BY cr.created_at, cr.id
LIMIT 20;

-- name: InsertLearnArticle :one
-- published_at is set only when the article is published; a draft keeps it NULL (learn_articles_published_ck).
INSERT INTO learn_articles (author_id, kind, title, summary, body_md, language, level, topic_code, status, published_at)
VALUES (sqlc.arg('author_id')::uuid, sqlc.arg('kind')::text, sqlc.arg('title')::text, sqlc.narg('summary')::text,
        sqlc.arg('body_md')::text, sqlc.arg('language')::text, sqlc.arg('level')::text, sqlc.narg('topic')::text,
        sqlc.arg('status')::text, CASE WHEN sqlc.arg('status')::text = 'published' THEN now() END)
RETURNING id;

-- name: EnsureTags :exec
-- Any member may create a tag when publishing (DATABASE_DESIGN tags); the label starts as the slug.
INSERT INTO tags (slug, label, created_by)
SELECT s, s, sqlc.arg('created_by')::uuid FROM unnest(sqlc.arg('slugs')::text[]) AS s
ON CONFLICT (slug) DO NOTHING;

-- name: InsertLearnArticleTags :exec
INSERT INTO learn_article_tags (article_id, tag_id)
SELECT sqlc.arg('article_id')::uuid, t.id FROM tags t WHERE t.slug = ANY (sqlc.arg('slugs')::text[])
ON CONFLICT DO NOTHING;

-- The three link writers below lock the target (FOR SHARE) and insert only when it is published and
-- public, in the creating transaction (DATABASE_DESIGN 2.2). Zero rows means the target is not linkable.

-- name: LinkQuestionToArticle :execrows
INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_article_id, created_by)
SELECT 'question', q.id, 'article', sqlc.arg('article_id')::uuid, sqlc.arg('created_by')::uuid
FROM questions q
WHERE q.id = sqlc.arg('target_id')::uuid AND q.status = 'published'
FOR SHARE
ON CONFLICT DO NOTHING;

-- name: LinkResourceToArticle :execrows
INSERT INTO content_relations (source_type, src_resource_id, target_type, tgt_article_id, created_by)
SELECT 'resource', r.id, 'article', sqlc.arg('article_id')::uuid, sqlc.arg('created_by')::uuid
FROM resources r
JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
WHERE r.id = sqlc.arg('target_id')::uuid AND r.status = 'published'
FOR SHARE OF r
ON CONFLICT DO NOTHING;

-- name: LinkProjectToArticle :execrows
INSERT INTO content_relations (source_type, src_project_id, target_type, tgt_article_id, created_by)
SELECT 'project', pj.id, 'article', sqlc.arg('article_id')::uuid, sqlc.arg('created_by')::uuid
FROM projects pj
WHERE pj.id = sqlc.arg('target_id')::uuid AND pj.status = 'published' AND pj.visibility = 'public'
FOR SHARE
ON CONFLICT DO NOTHING;

-- name: ListTrainingSeries :many
SELECT s.id, s.title, s.summary, s.language, s.level, s.published_at,
       s.author_id, pr.username AS author_username, pr.display_name AS author_display_name,
       COALESCE(CASE WHEN pr.profile_visibility = 'public' AND pr.country_visibility = 'public'
             THEN pr.country_code END, '')::text AS author_country,
       (SELECT count(*)::int FROM training_lessons l
        JOIN learn_articles la ON la.id = l.article_id AND la.status = 'published'
        WHERE l.series_id = s.id AND l.status = 'published') AS lesson_count
FROM training_series s
JOIN profiles pr ON pr.user_id = s.author_id
WHERE s.status = 'published'
  AND (sqlc.narg('topic')::text IS NULL OR s.topic_code = sqlc.narg('topic'))
  AND (sqlc.narg('level')::text IS NULL OR s.level = sqlc.narg('level'))
  AND (sqlc.narg('lang')::text IS NULL OR s.language = sqlc.narg('lang'))
  AND (sqlc.narg('after_published')::timestamptz IS NULL
       OR (s.published_at, s.id) < (sqlc.narg('after_published')::timestamptz, sqlc.narg('after_id')::uuid))
ORDER BY s.published_at DESC, s.id DESC
LIMIT sqlc.arg('row_limit');

-- name: GetTrainingSeries :one
SELECT s.id, s.title, s.summary, s.language, s.level, s.published_at,
       s.author_id, pr.username AS author_username, pr.display_name AS author_display_name,
       COALESCE(CASE WHEN pr.profile_visibility = 'public' AND pr.country_visibility = 'public'
             THEN pr.country_code END, '')::text AS author_country
FROM training_series s
JOIN profiles pr ON pr.user_id = s.author_id
WHERE s.id = $1 AND s.status = 'published';

-- name: TrainingSeriesLessons :many
-- A lesson is listed only when it and its article are published (the contract requires article_id).
SELECT l.id, l.position, l.title, l.article_id
FROM training_lessons l
JOIN learn_articles la ON la.id = l.article_id AND la.status = 'published'
WHERE l.series_id = $1 AND l.status = 'published'
ORDER BY l.position, l.id
LIMIT 200;
