-- Migration: 00005_solve
-- Purpose: Create the Solve domain (questions, answers, votes, tags and sources) with the accepted-answer composite FK and weighted French-aware search columns.
-- Tables/indexes: questions, answers, votes, question_tags, question_sources, answer_sources; ALTER questions ADD questions_accepted_answer_fk; generated STORED search_document on questions and answers; indexes questions_author_idx, questions_country_idx, answers_question_idx, answers_author_idx, votes_question_idx, votes_answer_idx, question_tags_tag_idx.
-- Lock risk: none (new tables). Generated STORED columns on empty tables are free; adding one to a populated table later rewrites it.
-- Data volume: empty at creation.
-- Transaction: single transaction (goose default).
-- Compatibility window: none (new tables). Depends on 00001 (f_unaccent), 00002 (countries, tags) and 00003 (users).
-- Restore plan: forward-only. On an empty database, recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.3. Forward-only: no Down section.
-- +goose Up
CREATE TABLE questions (
  id                 uuid PRIMARY KEY DEFAULT uuidv7(),
  author_id          uuid NOT NULL,
  title              text NOT NULL,
  body_md            text NOT NULL,
  language           text NOT NULL,
  snippet            text,
  snippet_language   text,                   -- programming language label, display only (never executed: SPEC:158)
  environment        text,                   -- e.g. 'Go 1.23'
  country_code       text,
  status             text NOT NULL DEFAULT 'published',   -- publication lifecycle only; solve state is derived
  accepted_answer_id uuid,
  answer_count       integer NOT NULL DEFAULT 0,
  vote_score         integer NOT NULL DEFAULT 0,
  provenance         text NOT NULL DEFAULT 'live',
  published_at       timestamptz NOT NULL DEFAULT now(),
  removed_at         timestamptz,
  created_at         timestamptz NOT NULL DEFAULT now(),
  updated_at         timestamptz NOT NULL DEFAULT now(),
  search_document    tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', f_unaccent(title)), 'A') || setweight(to_tsvector('simple', f_unaccent(body_md)), 'B')) STORED,
  CONSTRAINT questions_author_fk FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT questions_country_fk FOREIGN KEY (country_code) REFERENCES countries (code) ON DELETE RESTRICT,
  CONSTRAINT questions_title_ck CHECK (char_length(title) BETWEEN 1 AND 200),
  CONSTRAINT questions_body_ck CHECK (char_length(body_md) BETWEEN 1 AND 20000),
  CONSTRAINT questions_language_ck CHECK (language IN ('fr','en')),
  CONSTRAINT questions_snippet_ck CHECK (snippet IS NULL OR char_length(snippet) <= 10000),
  CONSTRAINT questions_snippet_language_ck CHECK (snippet_language IS NULL OR char_length(snippet_language) <= 30),
  CONSTRAINT questions_environment_ck CHECK (environment IS NULL OR char_length(environment) <= 100),
  CONSTRAINT questions_status_ck CHECK (status IN ('published','removed')),
  CONSTRAINT questions_removed_ck CHECK ((status = 'removed') = (removed_at IS NOT NULL)),
  CONSTRAINT questions_counters_ck CHECK (answer_count >= 0),
  CONSTRAINT questions_provenance_ck CHECK (provenance IN ('live','demo'))
);
CREATE INDEX questions_author_idx ON questions (author_id, created_at DESC, id DESC);
CREATE INDEX questions_country_idx ON questions (country_code) WHERE country_code IS NOT NULL;

CREATE TABLE answers (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  question_id uuid NOT NULL,
  author_id   uuid NOT NULL,
  body_md     text NOT NULL,
  language    text NOT NULL,
  status      text NOT NULL DEFAULT 'published',
  vote_score  integer NOT NULL DEFAULT 0,
  provenance  text NOT NULL DEFAULT 'live',
  removed_at  timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  search_document tsvector GENERATED ALWAYS AS (to_tsvector('simple', f_unaccent(body_md))) STORED,
  CONSTRAINT answers_question_fk FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE RESTRICT,
  CONSTRAINT answers_author_fk FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT answers_id_question_uq UNIQUE (id, question_id),
  CONSTRAINT answers_body_ck CHECK (char_length(body_md) BETWEEN 1 AND 20000),
  CONSTRAINT answers_language_ck CHECK (language IN ('fr','en')),
  CONSTRAINT answers_status_ck CHECK (status IN ('published','removed')),
  CONSTRAINT answers_removed_ck CHECK ((status = 'removed') = (removed_at IS NOT NULL)),
  CONSTRAINT answers_provenance_ck CHECK (provenance IN ('live','demo'))
);
CREATE INDEX answers_question_idx ON answers (question_id, created_at, id) WHERE status = 'published';
CREATE INDEX answers_author_idx ON answers (author_id, created_at DESC, id DESC);

-- Circular reference broken with a nullable column plus ALTER (no DEFERRABLE needed).
-- MATCH SIMPLE: not checked while accepted_answer_id IS NULL.
ALTER TABLE questions
  ADD CONSTRAINT questions_accepted_answer_fk
  FOREIGN KEY (accepted_answer_id, id) REFERENCES answers (id, question_id) ON DELETE RESTRICT;

CREATE TABLE votes (
  user_id     uuid NOT NULL,
  target_type text NOT NULL,
  question_id uuid,
  answer_id   uuid,
  target_id   uuid GENERATED ALWAYS AS (COALESCE(question_id, answer_id)) STORED NOT NULL,
  value       smallint NOT NULL DEFAULT 1,
  created_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT votes_pk PRIMARY KEY (user_id, target_type, target_id),
  CONSTRAINT votes_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT votes_question_fk FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE RESTRICT,
  CONSTRAINT votes_answer_fk FOREIGN KEY (answer_id) REFERENCES answers (id) ON DELETE RESTRICT,
  CONSTRAINT votes_target_ck CHECK (
    (target_type = 'question' AND question_id IS NOT NULL AND answer_id IS NULL)
    OR (target_type = 'answer' AND answer_id IS NOT NULL AND question_id IS NULL)),
  CONSTRAINT votes_value_ck CHECK (value IN (1, -1))
);
CREATE INDEX votes_question_idx ON votes (question_id) WHERE question_id IS NOT NULL;
CREATE INDEX votes_answer_idx ON votes (answer_id) WHERE answer_id IS NOT NULL;

CREATE TABLE question_tags (
  question_id uuid NOT NULL,
  tag_id      uuid NOT NULL,
  CONSTRAINT question_tags_pk PRIMARY KEY (question_id, tag_id),
  CONSTRAINT question_tags_question_fk FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE,
  CONSTRAINT question_tags_tag_fk FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE RESTRICT
);
CREATE INDEX question_tags_tag_idx ON question_tags (tag_id, question_id);

CREATE TABLE question_sources (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  question_id uuid NOT NULL,
  title       text,
  url         text NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT question_sources_question_fk FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE,
  CONSTRAINT question_sources_url_uq UNIQUE (question_id, url),
  CONSTRAINT question_sources_url_ck CHECK (url ~ '^https?://' AND char_length(url) <= 2048),
  CONSTRAINT question_sources_title_ck CHECK (title IS NULL OR char_length(title) <= 200)
);

CREATE TABLE answer_sources (
  id         uuid PRIMARY KEY DEFAULT uuidv7(),
  answer_id  uuid NOT NULL,
  title      text,
  url        text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT answer_sources_answer_fk FOREIGN KEY (answer_id) REFERENCES answers (id) ON DELETE CASCADE,
  CONSTRAINT answer_sources_url_uq UNIQUE (answer_id, url),
  CONSTRAINT answer_sources_url_ck CHECK (url ~ '^https?://' AND char_length(url) <= 2048),
  CONSTRAINT answer_sources_title_ck CHECK (title IS NULL OR char_length(title) <= 200)
);
