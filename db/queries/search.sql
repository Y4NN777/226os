-- Search reads (S2). One bounded, ranked query per group over the stored generated search_document
-- (migrations 00005-00008) and the partial GIN indexes of migration 00010: each query repeats the
-- index predicate (status = 'published', plus visibility = 'public' for projects) so the index is usable.
-- Order is ts_rank DESC then id, a stable UUID tie-breaker. websearch_to_tsquery never raises on
-- user input; the service caps its length first (C4 DECIDED option 1, ADR-0005 point 6).
-- The excerpt is the first 600 characters of the summary or body; the Go side turns it into a plain
-- 300-rune snippet. Answers have no search group, so their GIN index is not queried here.

-- name: SearchQuestions :many
SELECT q.id, q.title, left(q.body_md, 600)::text AS excerpt, q.language
FROM questions q
WHERE q.status = 'published'
  AND q.search_document @@ websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text))
  AND (sqlc.narg('country')::text IS NULL OR q.country_code = sqlc.narg('country'))
  AND (sqlc.narg('lang')::text IS NULL OR q.language = sqlc.narg('lang'))
ORDER BY ts_rank(q.search_document, websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text))) DESC, q.id
LIMIT sqlc.arg('row_limit');

-- A resource is visible only with a published, not removed provider (same rule as the catalogue).
-- Country is the resource_countries membership; lang is the language of the summary.
-- name: SearchResources :many
SELECT r.id, r.name AS title, left(r.summary, 600)::text AS excerpt, r.summary_language AS language
FROM resources r
JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
WHERE r.status = 'published'
  AND r.search_document @@ websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text))
  AND (sqlc.narg('country')::text IS NULL OR EXISTS (
        SELECT 1 FROM resource_countries c
        WHERE c.resource_id = r.id AND c.country_code = sqlc.narg('country')))
  AND (sqlc.narg('lang')::text IS NULL OR r.summary_language = sqlc.narg('lang'))
ORDER BY ts_rank(r.search_document, websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text))) DESC, r.id
LIMIT sqlc.arg('row_limit');

-- Private projects never appear, even for their owner.
-- name: SearchProjects :many
SELECT p.id, p.name AS title, left(p.summary, 600)::text AS excerpt, p.language
FROM projects p
WHERE p.status = 'published' AND p.visibility = 'public'
  AND p.search_document @@ websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text))
  AND (sqlc.narg('country')::text IS NULL OR p.country_code = sqlc.narg('country'))
  AND (sqlc.narg('lang')::text IS NULL OR p.language = sqlc.narg('lang'))
ORDER BY ts_rank(p.search_document, websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text))) DESC, p.id
LIMIT sqlc.arg('row_limit');

-- Learn articles have no country column, so the country filter does not apply to this group.
-- Only published articles appear (draft, archived and removed never do).
-- name: SearchLearn :many
SELECT a.id, a.title, left(COALESCE(NULLIF(a.summary, ''), a.body_md), 600)::text AS excerpt, a.language
FROM learn_articles a
WHERE a.status = 'published'
  AND a.search_document @@ websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text))
  AND (sqlc.narg('lang')::text IS NULL OR a.language = sqlc.narg('lang'))
ORDER BY ts_rank(a.search_document, websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text))) DESC, a.id
LIMIT sqlc.arg('row_limit');
