-- Migration: 00009_content_relations
-- Purpose: Create content_relations: typed links between questions, resources, projects and articles with one real FK column per entity per side (no polymorphic table).
-- Tables/indexes: content_relations (generated STORED source_id and target_id); indexes content_relations_target_idx, content_relations_created_by_idx and eight per-FK partial indexes (src_* and tgt_*). Overlap between content_relations_target_idx and the per-FK indexes is to be reviewed against EXPLAIN evidence.
-- Lock risk: none (new table on an empty database).
-- Data volume: empty at creation.
-- Transaction: single transaction (goose default).
-- Compatibility window: none (new table). Must follow 00005 to 00008 (FK targets questions, resources, projects, learn_articles).
-- Restore plan: forward-only. On an empty database, recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 2.2 and 3.7. Forward-only: no Down section. content_reports (P1) is not created here.
-- +goose Up
CREATE TABLE content_relations (
  id             uuid PRIMARY KEY DEFAULT uuidv7(),
  source_type    text NOT NULL,
  src_question_id uuid,
  src_resource_id uuid,
  src_project_id  uuid,
  src_article_id  uuid,
  target_type    text NOT NULL,
  tgt_question_id uuid,
  tgt_resource_id uuid,
  tgt_project_id  uuid,
  tgt_article_id  uuid,
  source_id uuid GENERATED ALWAYS AS (COALESCE(src_question_id, src_resource_id, src_project_id, src_article_id)) STORED NOT NULL,
  target_id uuid GENERATED ALWAYS AS (COALESCE(tgt_question_id, tgt_resource_id, tgt_project_id, tgt_article_id)) STORED NOT NULL,
  relation_kind  text NOT NULL DEFAULT 'related',
  status         text NOT NULL DEFAULT 'active',
  created_by     uuid NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT content_relations_src_question_fk FOREIGN KEY (src_question_id) REFERENCES questions (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_src_resource_fk FOREIGN KEY (src_resource_id) REFERENCES resources (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_src_project_fk  FOREIGN KEY (src_project_id)  REFERENCES projects (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_src_article_fk  FOREIGN KEY (src_article_id)  REFERENCES learn_articles (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_tgt_question_fk FOREIGN KEY (tgt_question_id) REFERENCES questions (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_tgt_resource_fk FOREIGN KEY (tgt_resource_id) REFERENCES resources (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_tgt_project_fk  FOREIGN KEY (tgt_project_id)  REFERENCES projects (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_tgt_article_fk  FOREIGN KEY (tgt_article_id)  REFERENCES learn_articles (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_created_by_fk   FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT content_relations_source_ck CHECK (
    num_nonnulls(src_question_id, src_resource_id, src_project_id, src_article_id) = 1
    AND (source_type = 'question') = (src_question_id IS NOT NULL)
    AND (source_type = 'resource') = (src_resource_id IS NOT NULL)
    AND (source_type = 'project')  = (src_project_id IS NOT NULL)
    AND (source_type = 'article')  = (src_article_id IS NOT NULL)),
  CONSTRAINT content_relations_target_ck CHECK (
    num_nonnulls(tgt_question_id, tgt_resource_id, tgt_project_id, tgt_article_id) = 1
    AND (target_type = 'question') = (tgt_question_id IS NOT NULL)
    AND (target_type = 'resource') = (tgt_resource_id IS NOT NULL)
    AND (target_type = 'project')  = (tgt_project_id IS NOT NULL)
    AND (target_type = 'article')  = (tgt_article_id IS NOT NULL)),
  CONSTRAINT content_relations_pair_ck CHECK ((source_type, target_type) IN (
    ('question','resource'), ('question','project'), ('question','article'),
    ('resource','project'), ('resource','article'), ('project','article'))),
  CONSTRAINT content_relations_kind_ck CHECK (relation_kind IN ('related','uses','explains','implements')),
  CONSTRAINT content_relations_status_ck CHECK (status IN ('active','hidden')),
  CONSTRAINT content_relations_uq UNIQUE (source_type, source_id, target_type, target_id, relation_kind)
);
CREATE INDEX content_relations_target_idx ON content_relations (target_type, target_id) WHERE status = 'active';
CREATE INDEX content_relations_created_by_idx ON content_relations (created_by);
-- one index per nullable FK so RESTRICT checks on parents stay cheap
CREATE INDEX content_relations_src_question_idx ON content_relations (src_question_id) WHERE src_question_id IS NOT NULL;
CREATE INDEX content_relations_src_resource_idx ON content_relations (src_resource_id) WHERE src_resource_id IS NOT NULL;
CREATE INDEX content_relations_src_project_idx  ON content_relations (src_project_id)  WHERE src_project_id IS NOT NULL;
CREATE INDEX content_relations_src_article_idx  ON content_relations (src_article_id)  WHERE src_article_id IS NOT NULL;
CREATE INDEX content_relations_tgt_question_idx ON content_relations (tgt_question_id) WHERE tgt_question_id IS NOT NULL;
CREATE INDEX content_relations_tgt_resource_idx ON content_relations (tgt_resource_id) WHERE tgt_resource_id IS NOT NULL;
CREATE INDEX content_relations_tgt_project_idx  ON content_relations (tgt_project_id)  WHERE tgt_project_id IS NOT NULL;
CREATE INDEX content_relations_tgt_article_idx  ON content_relations (tgt_article_id)  WHERE tgt_article_id IS NOT NULL;
