-- Migration: 00008_learn
-- Purpose: Create the Learn domain (articles and tutorials, article tags, the free training series and its ordered lessons).
-- Tables/indexes: learn_articles (generated STORED search_document), learn_article_tags, training_series, training_lessons (training_lessons_position_uq is DEFERRABLE INITIALLY DEFERRED); indexes learn_articles_author_idx, learn_articles_browse_idx, learn_articles_topic_idx, learn_articles_topic_fk_idx, learn_article_tags_tag_idx, training_series_author_idx, training_series_browse_idx, training_series_topic_idx, training_lessons_article_idx.
-- Lock risk: none (new tables on an empty database).
-- Data volume: empty at creation (learn_topics reference rows come from 00002).
-- Transaction: single transaction (goose default).
-- Compatibility window: none (new tables). Depends on 00001, 00002 and 00003. learn_progress (P1) is created later in 00011, not here.
-- Restore plan: forward-only. On an empty database, recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.6. Forward-only: no Down section.
-- +goose Up
CREATE TABLE learn_articles (
  id           uuid PRIMARY KEY DEFAULT uuidv7(),
  author_id    uuid NOT NULL,
  kind         text NOT NULL,
  title        text NOT NULL,
  summary      text,
  body_md      text NOT NULL,
  language     text NOT NULL,
  level        text NOT NULL,
  topic_code   text,
  status       text NOT NULL DEFAULT 'draft',     -- draft | published | archived | removed; public iff published
  provenance   text NOT NULL DEFAULT 'live',
  published_at timestamptz,
  removed_at   timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  search_document tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', f_unaccent(title)), 'A') || setweight(to_tsvector('simple', f_unaccent(coalesce(summary, ''))), 'B')
    || setweight(to_tsvector('simple', f_unaccent(body_md)), 'C')) STORED,
  CONSTRAINT learn_articles_author_fk FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT learn_articles_topic_fk FOREIGN KEY (topic_code) REFERENCES learn_topics (code) ON DELETE RESTRICT,
  CONSTRAINT learn_articles_kind_ck CHECK (kind IN ('article','tutorial')),
  CONSTRAINT learn_articles_title_ck CHECK (char_length(title) BETWEEN 1 AND 200),
  CONSTRAINT learn_articles_summary_ck CHECK (summary IS NULL OR char_length(summary) <= 500),
  CONSTRAINT learn_articles_body_ck CHECK (char_length(body_md) BETWEEN 1 AND 50000),
  CONSTRAINT learn_articles_language_ck CHECK (language IN ('fr','en')),
  CONSTRAINT learn_articles_level_ck CHECK (level IN ('beginner','intermediate','advanced')),
  CONSTRAINT learn_articles_status_ck CHECK (status IN ('draft','published','archived','removed')),
  CONSTRAINT learn_articles_published_ck CHECK (status = 'draft' OR status = 'removed' OR published_at IS NOT NULL),
  CONSTRAINT learn_articles_removed_ck CHECK ((status = 'removed') = (removed_at IS NOT NULL)),
  CONSTRAINT learn_articles_provenance_ck CHECK (provenance IN ('live','demo'))
);
CREATE INDEX learn_articles_author_idx ON learn_articles (author_id, created_at DESC, id DESC);
CREATE INDEX learn_articles_browse_idx ON learn_articles (published_at DESC, id DESC) WHERE status = 'published';
CREATE INDEX learn_articles_topic_idx ON learn_articles (topic_code, published_at DESC, id DESC) WHERE status = 'published' AND topic_code IS NOT NULL;
CREATE INDEX learn_articles_topic_fk_idx ON learn_articles (topic_code) WHERE topic_code IS NOT NULL;

CREATE TABLE learn_article_tags (
  article_id uuid NOT NULL,
  tag_id     uuid NOT NULL,
  CONSTRAINT learn_article_tags_pk PRIMARY KEY (article_id, tag_id),
  CONSTRAINT learn_article_tags_article_fk FOREIGN KEY (article_id) REFERENCES learn_articles (id) ON DELETE CASCADE,
  CONSTRAINT learn_article_tags_tag_fk FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE RESTRICT
);
CREATE INDEX learn_article_tags_tag_idx ON learn_article_tags (tag_id, article_id);

CREATE TABLE training_series (
  id           uuid PRIMARY KEY DEFAULT uuidv7(),
  author_id    uuid NOT NULL,
  origin       text NOT NULL DEFAULT 'member',
  title        text NOT NULL,
  summary      text NOT NULL,
  language     text NOT NULL,
  level        text NOT NULL,
  topic_code   text,
  status       text NOT NULL DEFAULT 'published',   -- published | removed
  provenance   text NOT NULL DEFAULT 'live',
  published_at timestamptz NOT NULL DEFAULT now(),
  removed_at   timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT training_series_author_fk FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT training_series_topic_fk FOREIGN KEY (topic_code) REFERENCES learn_topics (code) ON DELETE RESTRICT,
  CONSTRAINT training_series_origin_ck CHECK (origin IN ('member','platform')),
  CONSTRAINT training_series_title_ck CHECK (char_length(title) BETWEEN 1 AND 200),
  CONSTRAINT training_series_summary_ck CHECK (char_length(summary) BETWEEN 1 AND 1000),
  CONSTRAINT training_series_language_ck CHECK (language IN ('fr','en')),
  CONSTRAINT training_series_level_ck CHECK (level IN ('beginner','intermediate','advanced')),
  CONSTRAINT training_series_status_ck CHECK (status IN ('published','removed')),
  CONSTRAINT training_series_removed_ck CHECK ((status = 'removed') = (removed_at IS NOT NULL)),
  CONSTRAINT training_series_provenance_ck CHECK (provenance IN ('live','demo'))
);
CREATE INDEX training_series_author_idx ON training_series (author_id);
CREATE INDEX training_series_browse_idx ON training_series (published_at DESC, id DESC) WHERE status = 'published';
CREATE INDEX training_series_topic_idx ON training_series (topic_code) WHERE topic_code IS NOT NULL;

-- Lesson order is unique per series; DEFERRABLE INITIALLY DEFERRED so a reorder
-- can swap positions inside one transaction (not usable as an ON CONFLICT arbiter).
CREATE TABLE training_lessons (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  series_id   uuid NOT NULL,
  position    integer NOT NULL,
  title       text NOT NULL,
  body_md     text,
  language    text NOT NULL,
  article_id  uuid,
  status      text NOT NULL DEFAULT 'published',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  removed_at  timestamptz,
  CONSTRAINT training_lessons_series_fk FOREIGN KEY (series_id) REFERENCES training_series (id) ON DELETE RESTRICT,
  CONSTRAINT training_lessons_article_fk FOREIGN KEY (article_id) REFERENCES learn_articles (id) ON DELETE RESTRICT,
  CONSTRAINT training_lessons_position_uq UNIQUE (series_id, position) DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT training_lessons_position_ck CHECK (position BETWEEN 1 AND 200),
  CONSTRAINT training_lessons_title_ck CHECK (char_length(title) BETWEEN 1 AND 200),
  CONSTRAINT training_lessons_body_ck CHECK (body_md IS NULL OR char_length(body_md) <= 100000),
  CONSTRAINT training_lessons_content_ck CHECK (body_md IS NOT NULL OR article_id IS NOT NULL),
  CONSTRAINT training_lessons_language_ck CHECK (language IN ('fr','en')),
  CONSTRAINT training_lessons_status_ck CHECK (status IN ('published','removed')),
  CONSTRAINT training_lessons_removed_ck CHECK ((status = 'removed') = (removed_at IS NOT NULL))
);
CREATE INDEX training_lessons_article_idx ON training_lessons (article_id) WHERE article_id IS NOT NULL;
