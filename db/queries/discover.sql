-- Catalogue reads (S2). Published resources with a published provider only.
-- Removed rows and unpublished providers are absent, which is the public 404.
-- A source is verified only when it is the resource's current verified source.
-- The age that turns a source into `stale` is not decided, so these queries never
-- label a source stale; the service maps the stored facts to verified, claimed
-- or unverified.

-- name: ListCatalogue :many
SELECT r.id, r.slug, r.name, r.type, r.category_code, r.summary, r.doc_languages,
       r.docs_url, r.verification_status, r.verified_at, r.updated_at, r.verified_source_id,
       p.slug AS provider_slug, p.name AS provider_name, p.website_url AS provider_url,
       (SELECT count(*)::int FROM resource_sources s WHERE s.resource_id = r.id) AS source_count
FROM resources r
JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
WHERE r.status = 'published'
  AND (sqlc.narg('category')::text IS NULL OR r.category_code = sqlc.narg('category'))
  AND (sqlc.narg('resource_type')::text IS NULL OR r.type = sqlc.narg('resource_type'))
  AND (sqlc.narg('provider')::text IS NULL OR p.slug = sqlc.narg('provider'))
  AND (sqlc.narg('country')::text IS NULL OR EXISTS (
        SELECT 1 FROM resource_countries c
        WHERE c.resource_id = r.id AND c.country_code = sqlc.narg('country')))
  AND (sqlc.narg('sdk')::text IS NULL OR EXISTS (
        SELECT 1 FROM resource_sdk_languages k
        WHERE k.resource_id = r.id AND k.language = sqlc.narg('sdk')))
  AND (sqlc.narg('doclang')::text IS NULL OR sqlc.narg('doclang') = ANY (r.doc_languages))
  AND (sqlc.narg('q')::text IS NULL OR r.search_document @@ websearch_to_tsquery('simple', f_unaccent(sqlc.narg('q')::text)))
  AND (sqlc.narg('after_name')::text IS NULL
       OR (r.name, r.id) > (sqlc.narg('after_name')::text, sqlc.narg('after_id')::uuid))
ORDER BY r.name, r.id
LIMIT sqlc.arg('row_limit');

-- name: GetCatalogueEntry :one
SELECT r.id, r.slug, r.name, r.type, r.category_code, r.summary, r.doc_languages,
       r.docs_url, r.verification_status, r.verified_at, r.updated_at, r.verified_source_id,
       p.slug AS provider_slug, p.name AS provider_name, p.website_url AS provider_url,
       (SELECT count(*)::int FROM resource_sources s WHERE s.resource_id = r.id) AS source_count
FROM resources r
JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
WHERE r.id = $1 AND r.status = 'published';

-- name: ListCatalogueSources :many
SELECT s.id, s.kind, s.title, s.url, s.checked_at,
       r.verification_status, r.verified_source_id
FROM resource_sources s
JOIN resources r ON r.id = s.resource_id AND r.status = 'published'
JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
WHERE s.resource_id = $1
ORDER BY s.checked_at DESC NULLS LAST, s.id
LIMIT 50;

-- name: CatalogueCountries :many
SELECT resource_id, country_code
FROM resource_countries
WHERE resource_id = ANY (sqlc.arg('ids')::uuid[])
ORDER BY resource_id, country_code
LIMIT 4000;

-- name: CatalogueSDKs :many
SELECT resource_id, language, name, COALESCE(package_url, repo_url) AS url
FROM resource_sdk_languages
WHERE resource_id = ANY (sqlc.arg('ids')::uuid[])
  AND COALESCE(package_url, repo_url) IS NOT NULL
ORDER BY resource_id, language, name
LIMIT 1000;

-- name: CatalogueRelated :many
SELECT cr.src_resource_id, cr.target_type, cr.target_id,
       COALESCE(q.title, res.name, proj.name, art.title) AS title,
       COALESCE(q.language, res.summary_language, proj.language, art.language) AS lang
FROM content_relations cr
LEFT JOIN questions q ON q.id = cr.tgt_question_id AND q.status = 'published'
LEFT JOIN resources res ON res.id = cr.tgt_resource_id AND res.status = 'published'
LEFT JOIN projects proj ON proj.id = cr.tgt_project_id AND proj.status = 'published'
LEFT JOIN learn_articles art ON art.id = cr.tgt_article_id AND art.status = 'published'
WHERE cr.src_resource_id = ANY (sqlc.arg('ids')::uuid[])
  AND cr.status = 'active'
  AND COALESCE(q.id, res.id, proj.id, art.id) IS NOT NULL
ORDER BY cr.src_resource_id, cr.created_at, cr.id
LIMIT 1000;

-- name: CatalogueDiscussions :many
SELECT cr.tgt_resource_id AS resource_id, q.id AS question_id, q.title, q.language
FROM content_relations cr
JOIN questions q ON q.id = cr.src_question_id AND q.status = 'published'
WHERE cr.tgt_resource_id = ANY (sqlc.arg('ids')::uuid[])
  AND cr.source_type = 'question'
  AND cr.target_type = 'resource'
  AND cr.status = 'active'
ORDER BY cr.tgt_resource_id, cr.created_at, cr.id
LIMIT 1000;
