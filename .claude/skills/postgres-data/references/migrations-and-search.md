# postgres-data — migration and search patterns

Illustrative only. Table/column names are placeholders until ADR-OPEN-005 approves the schema.

## Migration header (required fields from `docs/spec-kit/DATABASE_DESIGN.md:41`)
```sql
-- migration: 0003_questions_search.sql
-- purpose: add search_document + GIN index to questions
-- affects: questions (column, index)
-- lock risk: ADD COLUMN ... GENERATED STORED rewrites table -> ACCESS EXCLUSIVE; size estimate: <N rows>
-- tx mode: transactional | non-transactional (CONCURRENTLY)
-- compatibility window: app vN and vN+1 both work
-- rollback/restore: redeploy previous image; column is additive; restore from backup if data changed
-- owner: <named human>
SET lock_timeout = '5s';   -- value pin per ADR
```
Note: adding a STORED generated column rewrites the table (UNVERIFIED for the pinned major — confirm on ALTER TABLE docs before running against data).

## Expand / contract for incompatible change
1. Expand: add new nullable column / table; app writes both.
2. Backfill in bounded batches (`WHERE id > $last ORDER BY id LIMIT 1000`), outside the DDL migration.
3. Add constraint `NOT VALID`, then `VALIDATE CONSTRAINT` in a later migration.
4. Switch reads; deploy.
5. Contract (drop old) only in a later release, with backup + human-approved ADR (`docs/spec-kit/DATABASE_DESIGN.md:40`).

## Invariant examples
```sql
CREATE TABLE answers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id uuid NOT NULL REFERENCES questions(id) ON DELETE RESTRICT,
  author_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  accepted boolean NOT NULL DEFAULT false,
  status text NOT NULL CONSTRAINT answers_status_ck CHECK (status IN ('published','archived','removed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX answers_one_accepted_per_question_uq ON answers(question_id) WHERE accepted;

CREATE TABLE votes (
  user_id uuid NOT NULL REFERENCES users(id),
  target_type text NOT NULL CONSTRAINT votes_target_type_ck CHECK (target_type IN ('question','answer')),
  target_id uuid NOT NULL,
  value smallint NOT NULL CONSTRAINT votes_value_ck CHECK (value IN (-1, 1)),
  CONSTRAINT votes_one_per_target_uq UNIQUE (user_id, target_type, target_id)
);
```
`target_id` here has no FK — the existence check must be transactional or split into typed columns (`docs/spec-kit/DATABASE_DESIGN.md:17` "foreign keys or equivalent existence checks"). Prefer typed FK columns + CHECK exactly-one-not-null for auditability.

## Search query shape (bounded)
```sql
-- name: SearchQuestions :many
SELECT id, title, created_at,
       ts_rank(search_document, q) AS rank
FROM questions, websearch_to_tsquery('simple', @query::text) AS q
WHERE status = 'published' AND visibility = 'public'
  AND search_document @@ q
  AND (sqlc.narg('tag')::text IS NULL OR id IN (SELECT qt.question_id FROM question_tags qt JOIN tags t ON t.id = qt.tag_id WHERE t.code = sqlc.narg('tag')))
  -- AND <keyset cursor predicate on (rank, created_at, id)>  -- placeholder
ORDER BY rank DESC, created_at DESC, id DESC
LIMIT @page_limit;          -- server clamps to approved max (proposed 50, API_CONTRACT.md:9)
```
- Validate `len(query)` in Go before SQL (max length pin per ADR).
- Cursor encodes `(rank, created_at, id)`; float rank equality across pages is fragile — UNVERIFIED approach; consider ordering by `(rank rounded, id)` or snapshotting.
- Typo tolerance: secondary `title % @query` on `gin_trgm_ops` index only when FTS returns too few rows; require minimum query length (trigram-less patterns scan the whole index — https://www.postgresql.org/docs/current/pgtrgm.html).

## EXPLAIN evidence template
```sql
BEGIN;
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) <query with representative params>;
ROLLBACK;
```
Record: dataset size + seed version, PG version, params, plan, execution time, buffers hit/read. Confirm `Bitmap Index Scan` on the GIN index and no `Seq Scan` on large tables.

## sqlc.yaml shape (pin sqlc version per ADR)
```yaml
version: "2"
sql:
  - engine: postgresql
    schema: db/migrations
    queries: db/queries/solve
    gen:
      go:
        package: solvedb
        out: src/api/internal/solve/repository/solvedb
        sql_package: pgx/v5
```
Sources: https://docs.sqlc.dev/en/latest/reference/config.html · https://www.postgresql.org/docs/current/sql-altertable.html · https://www.postgresql.org/docs/current/textsearch-tables.html
