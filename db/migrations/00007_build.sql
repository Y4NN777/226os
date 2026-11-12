-- Migration: 00007_build
-- Purpose: Create the Build domain (projects, maintainers, tags, contribution needs, need skills, applications) including the normalized-GitHub-URL partial unique index.
-- Tables/indexes: projects (generated STORED search_document), project_maintainers, project_tags, contribution_needs, need_skills, applications; indexes projects_github_url_uq (partial unique), projects_owner_idx, projects_browse_idx, projects_country_idx, project_maintainers_user_idx, project_maintainers_added_by_idx, project_tags_tag_idx, contribution_needs_project_idx, contribution_needs_open_idx, contribution_needs_created_by_idx, need_skills_tag_idx, applications_need_idx, applications_user_idx, applications_decided_by_idx.
-- Lock risk: none (new tables on an empty database).
-- Data volume: empty at creation.
-- Transaction: single transaction (goose default). The partial unique index is created inline on the empty table.
-- Compatibility window: none (new tables). Depends on 00001, 00002 and 00003.
-- Restore plan: forward-only. On an empty database, recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.5. Forward-only: no Down section.
-- +goose Up
CREATE TABLE projects (
  id           uuid PRIMARY KEY DEFAULT uuidv7(),
  owner_id     uuid NOT NULL,
  name         text NOT NULL,
  summary      text NOT NULL,
  description_md text,
  language     text NOT NULL,
  github_url   text,
  license_spdx text,
  stage        text NOT NULL DEFAULT 'idea',
  country_code text,
  visibility   text NOT NULL DEFAULT 'public',         -- public | private (independent of status)
  status       text NOT NULL DEFAULT 'published',      -- publication lifecycle: published | removed
  provenance   text NOT NULL DEFAULT 'live',
  published_at timestamptz NOT NULL DEFAULT now(),
  removed_at   timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  search_document tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', f_unaccent(name)), 'A') || setweight(to_tsvector('simple', f_unaccent(summary)), 'B')
    || setweight(to_tsvector('simple', f_unaccent(coalesce(description_md, ''))), 'C')) STORED,
  CONSTRAINT projects_owner_fk FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT projects_country_fk FOREIGN KEY (country_code) REFERENCES countries (code) ON DELETE RESTRICT,
  CONSTRAINT projects_name_ck CHECK (char_length(name) BETWEEN 1 AND 100),
  CONSTRAINT projects_summary_ck CHECK (char_length(summary) BETWEEN 1 AND 500),
  CONSTRAINT projects_description_ck CHECK (description_md IS NULL OR char_length(description_md) <= 20000),
  CONSTRAINT projects_language_ck CHECK (language IN ('fr','en')),
  CONSTRAINT projects_github_url_ck CHECK (github_url IS NULL OR
    (github_url ~ '^https://github\.com/[a-z0-9][a-z0-9-]{0,38}/[a-z0-9._-]{1,100}$' AND github_url !~ '\.git$')),
  CONSTRAINT projects_github_required_ck CHECK (github_url IS NOT NULL OR stage = 'idea'),
  CONSTRAINT projects_license_ck CHECK (license_spdx IS NULL OR license_spdx ~ '^[A-Za-z0-9.+-]{1,64}$'),
  CONSTRAINT projects_stage_ck CHECK (stage IN ('idea','prototype','active','maintained')),
  CONSTRAINT projects_visibility_ck CHECK (visibility IN ('public','private')),
  CONSTRAINT projects_status_ck CHECK (status IN ('published','removed')),
  CONSTRAINT projects_removed_ck CHECK ((status = 'removed') = (removed_at IS NOT NULL)),
  CONSTRAINT projects_provenance_ck CHECK (provenance IN ('live','demo'))
);
CREATE UNIQUE INDEX projects_github_url_uq ON projects (github_url) WHERE github_url IS NOT NULL AND status <> 'removed';
CREATE INDEX projects_owner_idx ON projects (owner_id, created_at DESC, id DESC);
CREATE INDEX projects_browse_idx ON projects (published_at DESC, id DESC) WHERE status = 'published' AND visibility = 'public';
CREATE INDEX projects_country_idx ON projects (country_code, published_at DESC, id DESC) WHERE status = 'published' AND visibility = 'public' AND country_code IS NOT NULL;

CREATE TABLE project_maintainers (
  project_id uuid NOT NULL,
  user_id    uuid NOT NULL,
  added_by   uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT project_maintainers_pk PRIMARY KEY (project_id, user_id),
  CONSTRAINT project_maintainers_project_fk FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE RESTRICT,
  CONSTRAINT project_maintainers_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT project_maintainers_added_by_fk FOREIGN KEY (added_by) REFERENCES users (id) ON DELETE RESTRICT
);
CREATE INDEX project_maintainers_user_idx ON project_maintainers (user_id, project_id);
CREATE INDEX project_maintainers_added_by_idx ON project_maintainers (added_by);

CREATE TABLE project_tags (
  project_id uuid NOT NULL,
  tag_id     uuid NOT NULL,
  CONSTRAINT project_tags_pk PRIMARY KEY (project_id, tag_id),
  CONSTRAINT project_tags_project_fk FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
  CONSTRAINT project_tags_tag_fk FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE RESTRICT
);
CREATE INDEX project_tags_tag_idx ON project_tags (tag_id, project_id);

CREATE TABLE contribution_needs (
  id              uuid PRIMARY KEY DEFAULT uuidv7(),
  project_id      uuid NOT NULL,
  title           text NOT NULL,
  description_md  text NOT NULL,
  language        text NOT NULL,
  level           text NOT NULL DEFAULT 'beginner',
  time_commitment text,
  status          text NOT NULL DEFAULT 'open',
  created_by      uuid NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  closed_at       timestamptz,
  CONSTRAINT contribution_needs_project_fk FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE RESTRICT,
  CONSTRAINT contribution_needs_created_by_fk FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT contribution_needs_title_ck CHECK (char_length(title) BETWEEN 1 AND 150),
  CONSTRAINT contribution_needs_description_ck CHECK (char_length(description_md) BETWEEN 1 AND 5000),
  CONSTRAINT contribution_needs_language_ck CHECK (language IN ('fr','en')),
  CONSTRAINT contribution_needs_level_ck CHECK (level IN ('beginner','intermediate','advanced')),
  CONSTRAINT contribution_needs_time_ck CHECK (time_commitment IS NULL OR time_commitment IN ('lt2','2to5','5to10','gt10')),
  CONSTRAINT contribution_needs_status_ck CHECK (status IN ('open','paused','filled','closed')),
  CONSTRAINT contribution_needs_closed_ck CHECK ((status IN ('filled','closed')) = (closed_at IS NOT NULL))
);
CREATE INDEX contribution_needs_project_idx ON contribution_needs (project_id, status, created_at DESC, id DESC);
CREATE INDEX contribution_needs_open_idx ON contribution_needs (created_at DESC, id DESC) WHERE status = 'open';
CREATE INDEX contribution_needs_created_by_idx ON contribution_needs (created_by);

CREATE TABLE need_skills (
  need_id uuid NOT NULL,
  tag_id  uuid NOT NULL,
  CONSTRAINT need_skills_pk PRIMARY KEY (need_id, tag_id),
  CONSTRAINT need_skills_need_fk FOREIGN KEY (need_id) REFERENCES contribution_needs (id) ON DELETE CASCADE,
  CONSTRAINT need_skills_tag_fk FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE RESTRICT
);
CREATE INDEX need_skills_tag_idx ON need_skills (tag_id, need_id);

CREATE TABLE applications (
  id                   uuid PRIMARY KEY DEFAULT uuidv7(),
  contribution_need_id uuid NOT NULL,
  user_id              uuid NOT NULL,
  message              text NOT NULL,
  language             text NOT NULL,
  status               text NOT NULL DEFAULT 'pending',
  decided_by           uuid,
  decided_at           timestamptz,
  withdrawn_at         timestamptz,
  created_at           timestamptz NOT NULL DEFAULT now(),
  updated_at           timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT applications_need_fk FOREIGN KEY (contribution_need_id) REFERENCES contribution_needs (id) ON DELETE RESTRICT,
  CONSTRAINT applications_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT applications_decided_by_fk FOREIGN KEY (decided_by) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT applications_user_need_uq UNIQUE (user_id, contribution_need_id),
  CONSTRAINT applications_message_ck CHECK (char_length(message) BETWEEN 1 AND 2000),
  CONSTRAINT applications_language_ck CHECK (language IN ('fr','en')),
  CONSTRAINT applications_status_ck CHECK (status IN ('pending','accepted','declined','withdrawn')),
  CONSTRAINT applications_decision_ck CHECK (
    (status IN ('accepted','declined')) = (decided_at IS NOT NULL AND decided_by IS NOT NULL)),
  CONSTRAINT applications_withdrawn_ck CHECK ((status = 'withdrawn') = (withdrawn_at IS NOT NULL))
);
CREATE INDEX applications_need_idx ON applications (contribution_need_id, status, created_at DESC, id DESC);
CREATE INDEX applications_user_idx ON applications (user_id, created_at DESC, id DESC);
CREATE INDEX applications_decided_by_idx ON applications (decided_by) WHERE decided_by IS NOT NULL;
