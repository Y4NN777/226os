-- Search reads (S2). One bounded, ranked query per group over the stored generated search_document
-- (migrations 00005-00008) and the partial GIN indexes of migration 00010: each query repeats the
-- index predicate (status = 'published', plus visibility = 'public' for projects) so the index is usable.
-- Order is ts_rank DESC then id, a stable UUID tie-breaker. websearch_to_tsquery never raises on
-- user input; the service caps its length first (C4 DECIDED option 1, ADR-0005 point 6).
-- The excerpt is the first 600 characters of the summary or body; the Go side turns it into a plain
-- 300-rune snippet. Answers have no search group, so their GIN index is not queried here.
-- Bounded ranking (#118): the tsquery is built once (input); a query with no positive term (only
-- negations, or nothing left after parsing: querytree is 'T' or empty) would match every row, so it
-- reads none; at most 500 matching rows (candidates) are ranked, so ts_rank never scores every row of
-- a very common term. Above 500 matches the ranked set is the first 500 the index returns.

-- name: SearchQuestions :many
WITH i AS NOT MATERIALIZED (SELECT websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text)) AS tsq),
candidates AS (
  SELECT q.id, q.title, q.body_md AS excerpt_src, q.language, q.search_document
  FROM questions q
  CROSS JOIN i
  WHERE querytree(i.tsq) NOT IN ('', 'T')
    AND q.status = 'published'
    AND q.search_document @@ i.tsq
    AND (sqlc.narg('country')::text IS NULL OR q.country_code = sqlc.narg('country'))
    AND (sqlc.narg('lang')::text IS NULL OR q.language = sqlc.narg('lang'))
  LIMIT 500
)
SELECT c.id, c.title, left(c.excerpt_src, 600)::text AS excerpt, c.language
FROM candidates c CROSS JOIN i
ORDER BY ts_rank(c.search_document, i.tsq) DESC, c.id
LIMIT sqlc.arg('row_limit');

-- A resource is visible only with a published, not removed provider (same rule as the catalogue).
-- Country is the resource_countries membership; lang is the language of the summary.
-- name: SearchResources :many
WITH i AS NOT MATERIALIZED (SELECT websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text)) AS tsq),
candidates AS (
  SELECT r.id, r.name AS title, r.summary AS excerpt_src, r.summary_language AS language, r.search_document
  FROM resources r
  JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
  CROSS JOIN i
  WHERE querytree(i.tsq) NOT IN ('', 'T')
    AND r.status = 'published'
    AND r.search_document @@ i.tsq
    AND (sqlc.narg('country')::text IS NULL OR EXISTS (
          SELECT 1 FROM resource_countries c
          WHERE c.resource_id = r.id AND c.country_code = sqlc.narg('country')))
    AND (sqlc.narg('lang')::text IS NULL OR r.summary_language = sqlc.narg('lang'))
  LIMIT 500
)
SELECT c.id, c.title, left(c.excerpt_src, 600)::text AS excerpt, c.language
FROM candidates c CROSS JOIN i
ORDER BY ts_rank(c.search_document, i.tsq) DESC, c.id
LIMIT sqlc.arg('row_limit');

-- Private projects never appear, even for their owner.
-- name: SearchProjects :many
WITH i AS NOT MATERIALIZED (SELECT websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text)) AS tsq),
candidates AS (
  SELECT p.id, p.name AS title, p.summary AS excerpt_src, p.language, p.search_document
  FROM projects p
  CROSS JOIN i
  WHERE querytree(i.tsq) NOT IN ('', 'T')
    AND p.status = 'published' AND p.visibility = 'public'
    AND p.search_document @@ i.tsq
    AND (sqlc.narg('country')::text IS NULL OR p.country_code = sqlc.narg('country'))
    AND (sqlc.narg('lang')::text IS NULL OR p.language = sqlc.narg('lang'))
  LIMIT 500
)
SELECT c.id, c.title, left(c.excerpt_src, 600)::text AS excerpt, c.language
FROM candidates c CROSS JOIN i
ORDER BY ts_rank(c.search_document, i.tsq) DESC, c.id
LIMIT sqlc.arg('row_limit');

-- Learn articles have no country column, so the country filter does not apply to this group.
-- Only published articles appear (draft, archived and removed never do).
-- name: SearchLearn :many
WITH i AS NOT MATERIALIZED (SELECT websearch_to_tsquery('simple', f_unaccent(sqlc.arg('q')::text)) AS tsq),
candidates AS (
  SELECT a.id, a.title, COALESCE(NULLIF(a.summary, ''), a.body_md) AS excerpt_src, a.language, a.search_document
  FROM learn_articles a
  CROSS JOIN i
  WHERE querytree(i.tsq) NOT IN ('', 'T')
    AND a.status = 'published'
    AND a.search_document @@ i.tsq
    AND (sqlc.narg('lang')::text IS NULL OR a.language = sqlc.narg('lang'))
  LIMIT 500
)
SELECT c.id, c.title, left(c.excerpt_src, 600)::text AS excerpt, c.language
FROM candidates c CROSS JOIN i
ORDER BY ts_rank(c.search_document, i.tsq) DESC, c.id
LIMIT sqlc.arg('row_limit');
