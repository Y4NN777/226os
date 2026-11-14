-- Solve questions and answers (S3, US-SOLVE-001). Only published questions are public; a removed
-- question is invisible to everyone (read, answer and edit all see "not found").
-- The byline shows username and display name whatever profile_visibility says (ADR-0005 point 15); the
-- country is shown only when both the profile and the country are public ('' otherwise).
-- The solve state is derived (never stored): open = no answer, answered = answers but none accepted,
-- resolved = an accepted answer. Lists are newest first with a (created_at, id) keyset.
-- The schema keeps ONE snippet per question and none per answer (migration 00005): the service refuses
-- more with a 422 instead of dropping data.

-- name: ListSolveQuestions :many
-- With q: most relevant first (ts_rank over the stored search_document, the S2 search tsquery:
-- websearch_to_tsquery('simple', f_unaccent(q)), a query with no positive term reads none), then newest,
-- then id. Bounded like S2 search (ADR-0005): only the 500 newest matches are ranked, so ts_rank never
-- scores every row of a very common term and the candidate set is the same on every page. Without q the
-- rank is 0, there is no cap, and the order is newest first. The keyset is (rank, created_at, id), all
-- descending; rank is the same real the cursor carried, so the comparison is exact.
WITH i AS NOT MATERIALIZED (
  SELECT CASE WHEN sqlc.narg('q')::text IS NULL THEN NULL
         ELSE websearch_to_tsquery('simple', f_unaccent(sqlc.narg('q')::text)) END AS tsq
),
candidates AS (
  SELECT q.id, q.title, q.body_md, q.language, q.snippet, q.snippet_language, q.country_code,
         q.answer_count, q.vote_score, q.accepted_answer_id, q.created_at, q.updated_at, q.author_id,
         (CASE WHEN i.tsq IS NULL THEN 0 ELSE ts_rank(q.search_document, i.tsq) END)::real AS rank
  FROM questions q
  CROSS JOIN i
  WHERE q.status = 'published'
    AND (i.tsq IS NULL OR (querytree(i.tsq) NOT IN ('', 'T') AND q.search_document @@ i.tsq))
    AND (sqlc.narg('tag')::text IS NULL OR EXISTS (
          SELECT 1 FROM question_tags qt JOIN tags t ON t.id = qt.tag_id AND t.status = 'active'
          WHERE qt.question_id = q.id AND t.slug = sqlc.narg('tag')::text
            -- only a tag the question displays (SolveQuestionTags: the first 5 active by slug): fewer than
            -- 5 active tags sort before it
            AND (SELECT count(*) FROM (
                   SELECT 1 FROM question_tags qt2
                   JOIN tags t2 ON t2.id = qt2.tag_id AND t2.status = 'active'
                   WHERE qt2.question_id = q.id AND t2.slug < t.slug LIMIT 5) lower_slugs) < 5))
    AND (sqlc.narg('state')::text IS NULL
         OR (CASE WHEN q.accepted_answer_id IS NOT NULL THEN 'resolved'
                  WHEN q.answer_count > 0 THEN 'answered' ELSE 'open' END) = sqlc.narg('state')::text)
    AND (sqlc.narg('country')::text IS NULL OR q.country_code = sqlc.narg('country')::text)
    AND (sqlc.narg('lang')::text IS NULL OR q.language = sqlc.narg('lang')::text)
  ORDER BY q.created_at DESC, q.id DESC
  LIMIT (SELECT CASE WHEN tsq IS NULL THEN NULL ELSE 500 END FROM i)
)
SELECT c.id, c.title, c.body_md, c.language, COALESCE(c.snippet, '')::text AS snippet,
       COALESCE(c.snippet_language, '')::text AS snippet_language, COALESCE(c.country_code, '')::text AS country_code,
       c.answer_count, c.vote_score, c.accepted_answer_id, c.created_at, c.updated_at,
       c.author_id, pr.username AS author_username, pr.display_name AS author_display_name,
       COALESCE(CASE WHEN pr.profile_visibility = 'public' AND pr.country_visibility = 'public'
             THEN pr.country_code END, '')::text AS author_country,
       c.rank
FROM candidates c
JOIN profiles pr ON pr.user_id = c.author_id
WHERE (sqlc.narg('after_created')::timestamptz IS NULL
       OR (c.rank, c.created_at, c.id)
          < (sqlc.arg('after_rank')::real, sqlc.narg('after_created')::timestamptz, sqlc.narg('after_id')::uuid))
ORDER BY c.rank DESC, c.created_at DESC, c.id DESC
LIMIT sqlc.arg('row_limit');

-- name: GetSolveQuestion :one
SELECT q.id, q.title, q.body_md, q.language, COALESCE(q.snippet, '')::text AS snippet,
       COALESCE(q.snippet_language, '')::text AS snippet_language, COALESCE(q.country_code, '')::text AS country_code,
       q.answer_count, q.vote_score, q.accepted_answer_id, q.created_at, q.updated_at,
       q.author_id, pr.username AS author_username, pr.display_name AS author_display_name,
       COALESCE(CASE WHEN pr.profile_visibility = 'public' AND pr.country_visibility = 'public'
             THEN pr.country_code END, '')::text AS author_country
FROM questions q
JOIN profiles pr ON pr.user_id = q.author_id
WHERE q.id = sqlc.arg('id')::uuid AND q.status = 'published';

-- name: SolveQuestionTags :many
-- Active tags only: a hidden tag is the moderator tool (DATABASE_DESIGN tags). At most 5 per question,
-- the first 5 by slug (Question.tags maxItems 5, issue #167): an edit keeps links to hidden tags, so a
-- restored tag can make a sixth active link, and no constraint caps the rows. The author's 422 counts
-- only the tags they can see (service validateTags); the bound lives on the read.
SELECT question_id, slug
FROM (
  SELECT qt.question_id, t.slug,
         row_number() OVER (PARTITION BY qt.question_id ORDER BY t.slug) AS rn
  FROM question_tags qt
  JOIN tags t ON t.id = qt.tag_id AND t.status = 'active'
  WHERE qt.question_id = ANY (sqlc.arg('ids')::uuid[])
) ranked
WHERE rn <= 5
ORDER BY question_id, slug
LIMIT 300;

-- name: SolveQuestionRelated :many
-- The visible links of a page of questions, at most 20 per question: a question always points to a
-- resource, a project or an article (content_relations stores it as the source). Only active relations
-- to published, public targets are returned.
SELECT x.question_id::uuid AS question_id, x.kind::text AS kind, x.target_id::uuid AS target_id,
       left(x.title, 200)::text AS title, x.lang::text AS lang
FROM (
  SELECT r.question_id, r.kind, r.target_id, r.title, r.lang,
         row_number() OVER (PARTITION BY r.question_id ORDER BY r.created_at, r.rel_id) AS n
  FROM (
    SELECT cr.src_question_id AS question_id, 'resource' AS kind, res.id AS target_id, res.name AS title,
           res.summary_language AS lang, cr.created_at, cr.id AS rel_id
    FROM content_relations cr
    JOIN resources res ON res.id = cr.tgt_resource_id AND res.status = 'published'
    JOIN providers pv ON pv.id = res.provider_id AND pv.status = 'published' AND pv.removed_at IS NULL
    WHERE cr.src_question_id = ANY (sqlc.arg('question_ids')::uuid[]) AND cr.status = 'active'
    UNION ALL
    SELECT cr.src_question_id, 'project', pj.id, pj.name, pj.language, cr.created_at, cr.id
    FROM content_relations cr
    JOIN projects pj ON pj.id = cr.tgt_project_id AND pj.status = 'published' AND pj.visibility = 'public'
    WHERE cr.src_question_id = ANY (sqlc.arg('question_ids')::uuid[]) AND cr.status = 'active'
    UNION ALL
    SELECT cr.src_question_id, 'article', a.id, a.title, a.language, cr.created_at, cr.id
    FROM content_relations cr
    JOIN learn_articles a ON a.id = cr.tgt_article_id AND a.status = 'published'
    WHERE cr.src_question_id = ANY (sqlc.arg('question_ids')::uuid[]) AND cr.status = 'active'
  ) r
) x
WHERE x.n <= 20
ORDER BY x.question_id, x.n
LIMIT 1020; -- 51 questions (a page plus one) x 20 links

-- name: InsertSolveQuestion :one
INSERT INTO questions (author_id, title, body_md, language, snippet, snippet_language, country_code)
VALUES (sqlc.arg('author_id')::uuid, sqlc.arg('title')::text, sqlc.arg('body_md')::text, sqlc.arg('language')::text,
        sqlc.narg('snippet')::text, sqlc.narg('snippet_language')::text, sqlc.narg('country_code')::text)
RETURNING id;

-- name: LockSolveQuestion :one
-- The row lock of an edit: the precondition (ownership, ETag) is decided on this locked row, so two
-- edits with the same ETag cannot both win. Removed is invisible.
SELECT id, author_id, answer_count, vote_score, accepted_answer_id, created_at, updated_at
FROM questions
WHERE id = sqlc.arg('id')::uuid AND status = 'published'
FOR UPDATE;

-- name: UpdateSolveQuestion :exec
-- JSON Merge Patch: an unset field keeps its value. The snippet and the country are cleared by
-- their set_* flag with a NULL value. lang is immutable and is never written here.
UPDATE questions SET
  title = COALESCE(sqlc.narg('title')::text, title),
  body_md = COALESCE(sqlc.narg('body_md')::text, body_md),
  snippet = CASE WHEN sqlc.arg('set_snippet')::boolean THEN sqlc.narg('snippet')::text ELSE snippet END,
  snippet_language = CASE WHEN sqlc.arg('set_snippet')::boolean THEN sqlc.narg('snippet_language')::text ELSE snippet_language END,
  country_code = CASE WHEN sqlc.arg('set_country')::boolean THEN sqlc.narg('country_code')::text ELSE country_code END,
  updated_at = now()
WHERE id = sqlc.arg('id')::uuid;

-- name: DeleteSolveQuestionTags :exec
-- Tags are replaced as a set; links to hidden tags are kept (moderator state, issue #156).
DELETE FROM question_tags qt
USING tags t
WHERE qt.question_id = sqlc.arg('question_id')::uuid AND qt.tag_id = t.id AND t.status = 'active';

-- name: InsertSolveQuestionTags :exec
INSERT INTO question_tags (question_id, tag_id)
SELECT sqlc.arg('question_id')::uuid, t.id FROM tags t WHERE t.slug = ANY (sqlc.arg('slugs')::text[])
ON CONFLICT DO NOTHING;

-- name: DeleteSolveQuestionLinks :exec
-- An edit replaces the active links the author wrote (created_by); a link someone else created stays, and
-- a hidden one is a moderator decision: it stays hidden, and re-inserting its pair conflicts and does nothing.
DELETE FROM content_relations
WHERE source_type = 'question' AND src_question_id = sqlc.arg('question_id')::uuid
  AND created_by = sqlc.arg('created_by')::uuid AND status = 'active';

-- The three link writers lock the target (FOR SHARE) and return whether it is usable, so an already
-- existing link (ON CONFLICT DO NOTHING) is not mistaken for a bad target.

-- name: SolveLinkResource :one
WITH target AS (
  SELECT r.id FROM resources r
  JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
  WHERE r.id = sqlc.arg('target_id')::uuid AND r.status = 'published'
  FOR SHARE OF r
), ins AS (
  INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_resource_id, created_by)
  SELECT 'question', sqlc.arg('question_id')::uuid, 'resource', target.id, sqlc.arg('created_by')::uuid FROM target
  ON CONFLICT DO NOTHING
)
SELECT EXISTS (SELECT 1 FROM target)::boolean;

-- name: SolveLinkProject :one
WITH target AS (
  SELECT pj.id FROM projects pj
  WHERE pj.id = sqlc.arg('target_id')::uuid AND pj.status = 'published' AND pj.visibility = 'public'
  FOR SHARE
), ins AS (
  INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_project_id, created_by)
  SELECT 'question', sqlc.arg('question_id')::uuid, 'project', target.id, sqlc.arg('created_by')::uuid FROM target
  ON CONFLICT DO NOTHING
)
SELECT EXISTS (SELECT 1 FROM target)::boolean;

-- name: SolveLinkArticle :one
WITH target AS (
  SELECT a.id FROM learn_articles a WHERE a.id = sqlc.arg('target_id')::uuid AND a.status = 'published' FOR SHARE
), ins AS (
  INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_article_id, created_by)
  SELECT 'question', sqlc.arg('question_id')::uuid, 'article', target.id, sqlc.arg('created_by')::uuid FROM target
  ON CONFLICT DO NOTHING
)
SELECT EXISTS (SELECT 1 FROM target)::boolean;

-- name: SolveQuestionExists :one
SELECT EXISTS (SELECT 1 FROM questions WHERE id = sqlc.arg('id')::uuid AND status = 'published')::boolean;

-- name: BumpSolveAnswerCount :one
-- Counts the answer and row-locks the question in one statement: it is the visibility check (zero rows
-- means unknown or removed) and keeps answer_count exact under concurrent answers.
UPDATE questions SET answer_count = answer_count + 1
WHERE id = sqlc.arg('id')::uuid AND status = 'published'
RETURNING id;

-- name: InsertSolveAnswer :one
INSERT INTO answers (question_id, author_id, body_md, language)
VALUES (sqlc.arg('question_id')::uuid, sqlc.arg('author_id')::uuid, sqlc.arg('body_md')::text, sqlc.arg('language')::text)
RETURNING id;

-- name: InsertSolveAnswerSources :exec
INSERT INTO answer_sources (answer_id, url)
SELECT sqlc.arg('answer_id')::uuid, u FROM unnest(sqlc.arg('urls')::text[]) AS u
ON CONFLICT DO NOTHING;

-- name: GetSolveAnswer :one
SELECT a.id, a.question_id, a.body_md, a.language, a.vote_score, a.created_at, a.updated_at,
       (q.accepted_answer_id IS NOT DISTINCT FROM a.id)::boolean AS is_accepted,
       a.author_id, pr.username AS author_username, pr.display_name AS author_display_name,
       COALESCE(CASE WHEN pr.profile_visibility = 'public' AND pr.country_visibility = 'public'
             THEN pr.country_code END, '')::text AS author_country
FROM answers a
JOIN questions q ON q.id = a.question_id
JOIN profiles pr ON pr.user_id = a.author_id
WHERE a.id = sqlc.arg('id')::uuid AND a.status = 'published' AND q.status = 'published';

-- name: ListSolveAnswers :many
-- Accepted first, then vote_score descending, then oldest first, then id. The keyset is the same
-- order written all-ascending: (rank, -vote_score, created_at, id) with rank 0 for the accepted answer.
SELECT a.id, a.question_id, a.body_md, a.language, a.vote_score, a.created_at, a.updated_at,
       (q.accepted_answer_id IS NOT DISTINCT FROM a.id)::boolean AS is_accepted,
       a.author_id, pr.username AS author_username, pr.display_name AS author_display_name,
       COALESCE(CASE WHEN pr.profile_visibility = 'public' AND pr.country_visibility = 'public'
             THEN pr.country_code END, '')::text AS author_country
FROM answers a
JOIN questions q ON q.id = a.question_id
JOIN profiles pr ON pr.user_id = a.author_id
WHERE a.question_id = sqlc.arg('question_id')::uuid AND a.status = 'published' AND q.status = 'published'
  AND (sqlc.narg('after_id')::uuid IS NULL
       OR ((CASE WHEN q.accepted_answer_id IS NOT DISTINCT FROM a.id THEN 0 ELSE 1 END), -a.vote_score, a.created_at, a.id)
          > (sqlc.arg('after_rank')::int, sqlc.arg('after_neg_score')::int, sqlc.narg('after_created')::timestamptz, sqlc.narg('after_id')::uuid))
ORDER BY (CASE WHEN q.accepted_answer_id IS NOT DISTINCT FROM a.id THEN 0 ELSE 1 END), a.vote_score DESC, a.created_at, a.id
LIMIT sqlc.arg('row_limit');

-- name: SolveAnswerSources :many
SELECT s.answer_id, s.url
FROM answer_sources s
WHERE s.answer_id = ANY (sqlc.arg('ids')::uuid[])
ORDER BY s.answer_id, s.created_at, s.id
LIMIT 300;

-- Votes (US-SOLVE-001). A vote is two statements in one transaction: the lock of the target row, then the
-- vote write together with the score delta. The lock serialises every vote and every retraction of one
-- target, so the second statement (which starts after the lock is held, with a fresh snapshot) reads the
-- caller's previous vote exactly; two concurrent first votes of one user cannot both add their value.
-- vote_score is never read into Go and written back: the UPDATE adds new value minus previous value.
-- NO KEY UPDATE is the lock an ordinary UPDATE takes: it blocks other voters but not the foreign-key
-- checks of inserts that point at the row.

-- name: LockSolveQuestionForVote :one
-- A removed or unknown question is no row. The author comes back for the self-vote rule.
SELECT author_id
FROM questions
WHERE id = sqlc.arg('id')::uuid AND status = 'published'
FOR NO KEY UPDATE;

-- name: LockSolveAnswerForVote :one
-- A removed or unknown answer, or an answer of a removed question, is no row.
SELECT a.author_id
FROM answers a
JOIN questions q ON q.id = a.question_id
WHERE a.id = sqlc.arg('id')::uuid AND a.status = 'published' AND q.status = 'published'
FOR NO KEY UPDATE OF a;

-- name: SetSolveQuestionVote :one
-- value 1 or -1 inserts or replaces the vote (ON CONFLICT: a repeat creates no row), value 0 deletes it.
-- Returns the new vote_score. Run only after LockSolveQuestionForVote.
WITH old AS (
  SELECT v.value FROM votes v
  WHERE v.user_id = sqlc.arg('user_id')::uuid AND v.target_type = 'question' AND v.target_id = sqlc.arg('question_id')::uuid
), del AS (
  DELETE FROM votes
  WHERE sqlc.arg('value')::smallint = 0
    AND user_id = sqlc.arg('user_id')::uuid AND target_type = 'question' AND target_id = sqlc.arg('question_id')::uuid
), up AS (
  INSERT INTO votes (user_id, target_type, question_id, value)
  SELECT sqlc.arg('user_id')::uuid, 'question', sqlc.arg('question_id')::uuid, sqlc.arg('value')::smallint
  WHERE sqlc.arg('value')::smallint <> 0
  ON CONFLICT ON CONSTRAINT votes_pk DO UPDATE SET value = EXCLUDED.value
)
UPDATE questions
SET vote_score = vote_score + sqlc.arg('value')::smallint - COALESCE((SELECT value FROM old), 0)::smallint
WHERE id = sqlc.arg('question_id')::uuid
RETURNING vote_score;

-- name: SetSolveAnswerVote :one
-- As SetSolveQuestionVote, for an answer. Run only after LockSolveAnswerForVote.
WITH old AS (
  SELECT v.value FROM votes v
  WHERE v.user_id = sqlc.arg('user_id')::uuid AND v.target_type = 'answer' AND v.target_id = sqlc.arg('answer_id')::uuid
), del AS (
  DELETE FROM votes
  WHERE sqlc.arg('value')::smallint = 0
    AND user_id = sqlc.arg('user_id')::uuid AND target_type = 'answer' AND target_id = sqlc.arg('answer_id')::uuid
), up AS (
  INSERT INTO votes (user_id, target_type, answer_id, value)
  SELECT sqlc.arg('user_id')::uuid, 'answer', sqlc.arg('answer_id')::uuid, sqlc.arg('value')::smallint
  WHERE sqlc.arg('value')::smallint <> 0
  ON CONFLICT ON CONSTRAINT votes_pk DO UPDATE SET value = EXCLUDED.value
)
UPDATE answers
SET vote_score = vote_score + sqlc.arg('value')::smallint - COALESCE((SELECT value FROM old), 0)::smallint
WHERE id = sqlc.arg('answer_id')::uuid
RETURNING vote_score;

-- name: LockSolveAnswerForAccept :one
-- The answer must be a published answer of THIS question (an id of another question, an unknown or a
-- removed answer is no row: one answer for all three). FOR SHARE keeps a moderator from removing it
-- between this check and the write.
SELECT a.id
FROM answers a
WHERE a.id = sqlc.arg('id')::uuid AND a.question_id = sqlc.arg('question_id')::uuid AND a.status = 'published'
FOR SHARE;

-- name: SetSolveAcceptedAnswer :exec
-- One UPDATE replaces the previous accepted answer or, with a NULL answer_id, clears it. It runs under the
-- question's row lock (LockSolveQuestion), so two concurrent accepts end with exactly one accepted answer;
-- the composite foreign key (accepted_answer_id, id) -> answers (id, question_id) is the backstop for the
-- answer belonging to the question. updated_at is not touched: acceptance does not change the ETag.
UPDATE questions
SET accepted_answer_id = sqlc.narg('answer_id')::uuid
WHERE id = sqlc.arg('id')::uuid;
