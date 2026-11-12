-- Migration: 00006_discover
-- Purpose: Create the Discover domain (providers, resources, sources, countries, tags, SDK languages, error reports) with the verified-source composite FK.
-- Tables/indexes: providers, resources, resource_sources, resource_countries, resource_tags, resource_sdk_languages, error_reports; ALTER resources ADD resources_verified_source_fk; generated STORED search_document on resources; indexes providers_created_by_idx, resources_provider_idx, resources_created_by_idx, resources_catalogue_idx, resource_sources_resource_idx, resource_sources_checked_by_idx, resource_countries_country_idx, resource_tags_tag_idx, resource_sdk_languages_language_idx, error_reports_open_uq (partial unique), error_reports_reporter_idx, error_reports_status_idx.
-- Lock risk: none (new tables on an empty database).
-- Data volume: empty at creation (resource_categories reference rows come from 00002).
-- Transaction: single transaction (goose default).
-- Compatibility window: none (new tables). Depends on 00001, 00002 and 00003.
-- Restore plan: forward-only. On an empty database, recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.4. Forward-only: no Down section. resource_reviews (P1) is not created here.
-- +goose Up
CREATE TABLE providers (
  id             uuid PRIMARY KEY DEFAULT uuidv7(),
  slug           text NOT NULL,
  name           text NOT NULL,
  website_url    text,
  summary        text,
  summary_language text,
  status         text NOT NULL DEFAULT 'published',
  provenance     text NOT NULL DEFAULT 'live',
  created_by     uuid NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  removed_at     timestamptz,
  CONSTRAINT providers_created_by_fk FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT providers_slug_uq UNIQUE (slug),
  CONSTRAINT providers_slug_ck CHECK (slug ~ '^[a-z0-9][a-z0-9-]{0,59}$'),
  CONSTRAINT providers_name_ck CHECK (char_length(name) BETWEEN 1 AND 120),
  CONSTRAINT providers_website_ck CHECK (website_url IS NULL OR (website_url ~ '^https://' AND char_length(website_url) <= 2048)),
  CONSTRAINT providers_summary_ck CHECK (summary IS NULL OR char_length(summary) <= 1000),
  CONSTRAINT providers_summary_language_ck CHECK ((summary IS NULL) = (summary_language IS NULL) AND (summary_language IS NULL OR summary_language IN ('fr','en'))),
  CONSTRAINT providers_status_ck CHECK (status IN ('published','removed')),
  CONSTRAINT providers_removed_ck CHECK ((status = 'removed') = (removed_at IS NOT NULL)),
  CONSTRAINT providers_provenance_ck CHECK (provenance IN ('live','demo'))
);
CREATE INDEX providers_created_by_idx ON providers (created_by);

CREATE TABLE resources (
  id                  uuid PRIMARY KEY DEFAULT uuidv7(),
  provider_id         uuid,
  category_code       text NOT NULL,
  type                text NOT NULL,                -- what the item is: api | sdk | tool
  slug                text NOT NULL,
  name                text NOT NULL,
  summary             text NOT NULL,
  summary_language    text NOT NULL,
  doc_languages       text[] NOT NULL DEFAULT '{}',
  website_url         text,
  docs_url            text,
  auth_note           text,                    -- free text, original language (summary_language)
  version_note        text,
  verification_status text NOT NULL DEFAULT 'unverified',
  verified_at         timestamptz,
  verified_source_id  uuid,
  status              text NOT NULL DEFAULT 'published',
  provenance          text NOT NULL DEFAULT 'live',
  created_by          uuid NOT NULL,
  published_at        timestamptz NOT NULL DEFAULT now(),
  removed_at          timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  search_document     tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', f_unaccent(name)), 'A') || setweight(to_tsvector('simple', f_unaccent(summary)), 'B')) STORED,
  CONSTRAINT resources_provider_fk FOREIGN KEY (provider_id) REFERENCES providers (id) ON DELETE RESTRICT,
  CONSTRAINT resources_category_fk FOREIGN KEY (category_code) REFERENCES resource_categories (code) ON DELETE RESTRICT,
  CONSTRAINT resources_created_by_fk FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT resources_slug_uq UNIQUE (slug),
  CONSTRAINT resources_slug_ck CHECK (slug ~ '^[a-z0-9][a-z0-9-]{0,79}$'),
  CONSTRAINT resources_type_ck CHECK (type IN ('api','sdk','tool')),
  CONSTRAINT resources_name_ck CHECK (char_length(name) BETWEEN 1 AND 120),
  CONSTRAINT resources_summary_ck CHECK (char_length(summary) BETWEEN 1 AND 2000),
  CONSTRAINT resources_summary_language_ck CHECK (summary_language IN ('fr','en')),
  CONSTRAINT resources_doc_languages_ck CHECK (doc_languages <@ ARRAY['fr','en']::text[]),
  CONSTRAINT resources_urls_ck CHECK (
    (website_url IS NULL OR (website_url ~ '^https://' AND char_length(website_url) <= 2048))
    AND (docs_url IS NULL OR (docs_url ~ '^https://' AND char_length(docs_url) <= 2048))),
  CONSTRAINT resources_notes_ck CHECK (
    (auth_note IS NULL OR char_length(auth_note) <= 1000) AND (version_note IS NULL OR char_length(version_note) <= 200)),
  CONSTRAINT resources_verification_status_ck CHECK (verification_status IN ('unverified','claimed','verified','disputed')),
  CONSTRAINT resources_verified_ck CHECK (
    verification_status <> 'verified' OR (verified_source_id IS NOT NULL AND verified_at IS NOT NULL)),
  CONSTRAINT resources_status_ck CHECK (status IN ('published','removed')),
  CONSTRAINT resources_removed_ck CHECK ((status = 'removed') = (removed_at IS NOT NULL)),
  CONSTRAINT resources_provenance_ck CHECK (provenance IN ('live','demo'))
);
CREATE INDEX resources_provider_idx ON resources (provider_id) WHERE provider_id IS NOT NULL;
CREATE INDEX resources_created_by_idx ON resources (created_by);
CREATE INDEX resources_catalogue_idx ON resources (category_code, type, name, id) WHERE status = 'published';

CREATE TABLE resource_sources (
  id           uuid PRIMARY KEY DEFAULT uuidv7(),
  resource_id  uuid NOT NULL,
  kind         text NOT NULL,
  title        text NOT NULL,
  url          text NOT NULL,
  published_at timestamptz,
  checked_at   timestamptz,
  checked_by   uuid,
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT resource_sources_resource_fk FOREIGN KEY (resource_id) REFERENCES resources (id) ON DELETE RESTRICT,
  CONSTRAINT resource_sources_checked_by_fk FOREIGN KEY (checked_by) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT resource_sources_id_resource_uq UNIQUE (id, resource_id),
  CONSTRAINT resource_sources_url_uq UNIQUE (resource_id, url),
  CONSTRAINT resource_sources_kind_ck CHECK (kind IN ('official_docs','changelog','pricing','community','other','claim')),
  CONSTRAINT resource_sources_url_ck CHECK (url ~ '^https://' AND char_length(url) <= 2048),
  CONSTRAINT resource_sources_title_ck CHECK (char_length(title) BETWEEN 1 AND 200),
  CONSTRAINT resource_sources_checked_ck CHECK (kind = 'claim' OR checked_at IS NOT NULL)
);
CREATE INDEX resource_sources_resource_idx ON resource_sources (resource_id, checked_at DESC);
CREATE INDEX resource_sources_checked_by_idx ON resource_sources (checked_by) WHERE checked_by IS NOT NULL;

-- Circular reference broken with a nullable column plus ALTER (no DEFERRABLE needed).
ALTER TABLE resources
  ADD CONSTRAINT resources_verified_source_fk
  FOREIGN KEY (verified_source_id, id) REFERENCES resource_sources (id, resource_id) ON DELETE RESTRICT;

CREATE TABLE resource_countries (
  resource_id  uuid NOT NULL,
  country_code text NOT NULL,
  CONSTRAINT resource_countries_pk PRIMARY KEY (resource_id, country_code),
  CONSTRAINT resource_countries_resource_fk FOREIGN KEY (resource_id) REFERENCES resources (id) ON DELETE CASCADE,
  CONSTRAINT resource_countries_country_fk FOREIGN KEY (country_code) REFERENCES countries (code) ON DELETE RESTRICT
);
CREATE INDEX resource_countries_country_idx ON resource_countries (country_code, resource_id);

CREATE TABLE resource_tags (
  resource_id uuid NOT NULL,
  tag_id      uuid NOT NULL,
  CONSTRAINT resource_tags_pk PRIMARY KEY (resource_id, tag_id),
  CONSTRAINT resource_tags_resource_fk FOREIGN KEY (resource_id) REFERENCES resources (id) ON DELETE CASCADE,
  CONSTRAINT resource_tags_tag_fk FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE RESTRICT
);
CREATE INDEX resource_tags_tag_idx ON resource_tags (tag_id, resource_id);

CREATE TABLE resource_sdk_languages (
  id           uuid PRIMARY KEY DEFAULT uuidv7(),
  resource_id  uuid NOT NULL,
  language     text NOT NULL,         -- programming language slug: 'go', 'python', 'typescript'
  name         text NOT NULL,
  package_url  text,
  repo_url     text,
  version_note text,
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT resource_sdk_languages_resource_fk FOREIGN KEY (resource_id) REFERENCES resources (id) ON DELETE CASCADE,
  CONSTRAINT resource_sdk_languages_uq UNIQUE (resource_id, language, name),
  CONSTRAINT resource_sdk_languages_language_ck CHECK (language ~ '^[a-z0-9][a-z0-9+#.-]{0,29}$'),
  CONSTRAINT resource_sdk_languages_name_ck CHECK (char_length(name) BETWEEN 1 AND 100),
  CONSTRAINT resource_sdk_languages_urls_ck CHECK (
    (package_url IS NULL OR (package_url ~ '^https://' AND char_length(package_url) <= 2048))
    AND (repo_url IS NULL OR (repo_url ~ '^https://' AND char_length(repo_url) <= 2048))),
  CONSTRAINT resource_sdk_languages_version_ck CHECK (version_note IS NULL OR char_length(version_note) <= 100)
);
CREATE INDEX resource_sdk_languages_language_idx ON resource_sdk_languages (language, resource_id);

CREATE TABLE error_reports (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  resource_id uuid NOT NULL,
  reporter_id uuid NOT NULL,
  kind        text NOT NULL,
  message     text,
  status      text NOT NULL DEFAULT 'received',
  handled_by  uuid,
  handled_at  timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT error_reports_resource_fk FOREIGN KEY (resource_id) REFERENCES resources (id) ON DELETE RESTRICT,
  CONSTRAINT error_reports_reporter_fk FOREIGN KEY (reporter_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT error_reports_handled_by_fk FOREIGN KEY (handled_by) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT error_reports_kind_ck CHECK (kind IN ('broken_link','outdated','incorrect','unsupported_country','other')),
  CONSTRAINT error_reports_message_ck CHECK (message IS NULL OR char_length(message) <= 2000),
  CONSTRAINT error_reports_status_ck CHECK (status IN ('received','triaged','resolved','rejected')),
  CONSTRAINT error_reports_handled_ck CHECK ((status IN ('resolved','rejected')) = (handled_at IS NOT NULL AND handled_by IS NOT NULL))
);
CREATE UNIQUE INDEX error_reports_open_uq ON error_reports (resource_id, reporter_id, kind) WHERE status IN ('received','triaged');
CREATE INDEX error_reports_reporter_idx ON error_reports (reporter_id);
CREATE INDEX error_reports_status_idx ON error_reports (status, created_at, id);
