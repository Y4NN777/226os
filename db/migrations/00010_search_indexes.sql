-- Migration: 00010_search_indexes
-- Purpose: Create the full-text (GIN on search_document) and bounded trigram (pg_trgm over f_unaccent) search indexes, all partial on published content.
-- Tables/indexes: questions_search_gin, questions_title_trgm, answers_search_gin, resources_search_gin, resources_name_trgm, providers_name_trgm, projects_search_gin, projects_name_trgm, learn_articles_search_gin, learn_articles_title_trgm. Optional query indexes of DATABASE_DESIGN.md section 7 are NOT created here: they wait for EXPLAIN evidence.
-- Lock risk: plain (transactional) CREATE INDEX takes a SHARE lock that blocks writes, acceptable only because every table is empty at this point. After data exists, every new index must use CREATE INDEX CONCURRENTLY in its own file that opts out of the transaction (goose NO TRANSACTION annotation), with lock_timeout set to fail fast.
-- Data volume: empty tables, so index builds are instantaneous.
-- Transaction: single transaction (goose default); this file must NOT be reused for concurrent builds on populated tables.
-- Compatibility window: none (new indexes). Queries must repeat the partial predicates (status = 'published', plus visibility = 'public' for projects) and the f_unaccent expression to use them.
-- Restore plan: forward-only. Indexes are rebuildable: restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest; a mistaken index is dropped by a new forward migration.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.9 and 9 (ADR-0005 point 6). Forward-only: no Down section.
-- +goose Up
CREATE INDEX questions_search_gin ON questions USING gin (search_document) WHERE status = 'published';
CREATE INDEX questions_title_trgm ON questions USING gin (f_unaccent(title) gin_trgm_ops) WHERE status = 'published';
CREATE INDEX answers_search_gin   ON answers   USING gin (search_document) WHERE status = 'published';
CREATE INDEX resources_search_gin ON resources USING gin (search_document) WHERE status = 'published';
CREATE INDEX resources_name_trgm  ON resources USING gin (f_unaccent(name) gin_trgm_ops) WHERE status = 'published';
CREATE INDEX providers_name_trgm  ON providers USING gin (f_unaccent(name) gin_trgm_ops) WHERE status = 'published';
CREATE INDEX projects_search_gin  ON projects  USING gin (search_document) WHERE status = 'published' AND visibility = 'public';
CREATE INDEX projects_name_trgm   ON projects  USING gin (f_unaccent(name) gin_trgm_ops) WHERE status = 'published' AND visibility = 'public';
CREATE INDEX learn_articles_search_gin ON learn_articles USING gin (search_document) WHERE status = 'published';
CREATE INDEX learn_articles_title_trgm ON learn_articles USING gin (f_unaccent(title) gin_trgm_ops) WHERE status = 'published';
