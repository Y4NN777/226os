-- Migration: 00003_identity
-- Purpose: Create the identity tables (accounts, credentials, profiles, skills, sessions, email verification and password reset tokens) and the deferred tags.created_by FK.
-- Tables/indexes: users, user_credentials, profiles, user_skills, sessions, email_verifications, password_resets; ALTER tags ADD tags_created_by_fk; indexes tags_created_by_idx, profiles_country_idx, user_skills_tag_idx, sessions_user_active_idx, sessions_absolute_expiry_idx, email_verifications_user_open_idx, password_resets_user_open_idx.
-- Lock risk: new tables only; the ALTER TABLE tags takes a brief ACCESS EXCLUSIVE lock on a table that is empty at this point (tags_created_by_fk validates zero rows).
-- Data volume: empty at creation. users is the most referenced table: later changes to it must be expand/contract.
-- Transaction: single transaction (goose default).
-- Compatibility window: none (new tables). 00004 and later migrations reference users.
-- Restore plan: forward-only. On an empty database, recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.2. Forward-only: no Down section.
-- +goose Up
CREATE TABLE users (
  id                uuid PRIMARY KEY DEFAULT uuidv7(),
  email             text NOT NULL,
  email_verified_at timestamptz,
  role              text NOT NULL DEFAULT 'member',
  status            text NOT NULL DEFAULT 'active',
  provenance        text NOT NULL DEFAULT 'live',
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  deleted_at        timestamptz,
  CONSTRAINT users_email_uq UNIQUE (email),
  CONSTRAINT users_email_normalized_ck CHECK (
    email = lower(btrim(email)) AND char_length(email) BETWEEN 3 AND 254 AND position('@' IN email) > 1),
  CONSTRAINT users_role_ck CHECK (role IN ('member','moderator')),
  CONSTRAINT users_status_ck CHECK (status IN ('active','suspended','deleted')),   -- unverified email is email_verified_at IS NULL, not a status
  CONSTRAINT users_deleted_ck CHECK ((status = 'deleted') = (deleted_at IS NOT NULL)),
  CONSTRAINT users_provenance_ck CHECK (provenance IN ('live','demo'))
);
ALTER TABLE tags ADD CONSTRAINT tags_created_by_fk FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT;
CREATE INDEX tags_created_by_idx ON tags (created_by) WHERE created_by IS NOT NULL;

CREATE TABLE user_credentials (
  user_id             uuid PRIMARY KEY,
  password_hash       text NOT NULL,           -- PHC string, includes salt and parameters
  hash_params_version smallint NOT NULL DEFAULT 1,
  password_changed_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT user_credentials_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT user_credentials_hash_ck CHECK (left(password_hash, 10) = '$argon2id$' AND char_length(password_hash) <= 255)
);

CREATE TABLE profiles (
  user_id        uuid PRIMARY KEY,
  username       text NOT NULL,
  display_name   text NOT NULL,
  intent         text,
  country_code   text,
  languages      text[] NOT NULL DEFAULT '{}',   -- spoken languages (BCP 47 primary tags)
  github_username text,
  bio            text,
  bio_language   text,
  open_to_collab boolean NOT NULL DEFAULT false,
  country_visibility        text NOT NULL DEFAULT 'public',
  languages_visibility      text NOT NULL DEFAULT 'public',
  skills_visibility         text NOT NULL DEFAULT 'public',
  github_visibility         text NOT NULL DEFAULT 'public',
  bio_visibility            text NOT NULL DEFAULT 'public',
  open_to_collab_visibility text NOT NULL DEFAULT 'public',
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT profiles_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT profiles_country_fk FOREIGN KEY (country_code) REFERENCES countries (code) ON DELETE RESTRICT,
  CONSTRAINT profiles_username_uq UNIQUE (username),
  CONSTRAINT profiles_username_ck CHECK (username ~ '^[a-z0-9][a-z0-9_-]{2,29}$'),
  CONSTRAINT profiles_display_name_ck CHECK (char_length(display_name) BETWEEN 1 AND 80),
  CONSTRAINT profiles_intent_ck CHECK (intent IS NULL OR intent IN ('ask','help','integrate','contribute','learn')),
  CONSTRAINT profiles_languages_ck CHECK (cardinality(languages) <= 10 AND array_to_string(languages, ',') ~ '^([a-z]{2,3}(,[a-z]{2,3})*)?$'),
  CONSTRAINT profiles_github_ck CHECK (github_username IS NULL OR github_username ~ '^[A-Za-z0-9][A-Za-z0-9-]{0,38}$'),
  CONSTRAINT profiles_bio_ck CHECK (bio IS NULL OR char_length(bio) <= 500),
  CONSTRAINT profiles_bio_language_ck CHECK ((bio IS NULL) = (bio_language IS NULL) AND (bio_language IS NULL OR bio_language IN ('fr','en'))),
  CONSTRAINT profiles_visibility_ck CHECK (
    country_visibility IN ('public','members','private') AND languages_visibility IN ('public','members','private')
    AND skills_visibility IN ('public','members','private') AND github_visibility IN ('public','members','private')
    AND bio_visibility IN ('public','members','private') AND open_to_collab_visibility IN ('public','members','private'))
);
CREATE INDEX profiles_country_idx ON profiles (country_code) WHERE country_code IS NOT NULL;

CREATE TABLE user_skills (
  user_id    uuid NOT NULL,
  tag_id     uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT user_skills_pk PRIMARY KEY (user_id, tag_id),
  CONSTRAINT user_skills_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT user_skills_tag_fk FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE RESTRICT
);
CREATE INDEX user_skills_tag_idx ON user_skills (tag_id, user_id);

CREATE TABLE sessions (
  id                  uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id             uuid NOT NULL,
  token_hash          bytea NOT NULL,
  csrf_token_hash     bytea NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  last_seen_at        timestamptz NOT NULL DEFAULT now(),   -- updated at most every few minutes (write amplification)
  authenticated_at    timestamptz NOT NULL DEFAULT now(),   -- last sign-in or re-authentication; sensitive actions need it within 15 min
  idle_expires_at     timestamptz NOT NULL,                 -- last_seen + 2 h (ADR-0003)
  absolute_expires_at timestamptz NOT NULL,                 -- created + 7 d (ADR-0003)
  revoked_at          timestamptz,
  revoked_reason      text,
  ip_prefix           cidr,                                 -- /24 (IPv4) or /48 (IPv6) only
  user_agent_summary  text,
  CONSTRAINT sessions_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT sessions_token_hash_uq UNIQUE (token_hash),
  CONSTRAINT sessions_token_hash_ck CHECK (octet_length(token_hash) = 32),
  CONSTRAINT sessions_csrf_hash_ck CHECK (octet_length(csrf_token_hash) = 32),
  CONSTRAINT sessions_expiry_ck CHECK (idle_expires_at <= absolute_expires_at AND absolute_expires_at > created_at),
  CONSTRAINT sessions_revoked_ck CHECK ((revoked_at IS NULL) = (revoked_reason IS NULL)),
  CONSTRAINT sessions_revoked_reason_ck CHECK (revoked_reason IS NULL OR revoked_reason IN
    ('logout','rotated','password_reset','password_change','security_event','moderation','expired')),
  CONSTRAINT sessions_ua_ck CHECK (user_agent_summary IS NULL OR char_length(user_agent_summary) <= 120)
);
CREATE INDEX sessions_user_active_idx ON sessions (user_id) WHERE revoked_at IS NULL;
CREATE INDEX sessions_absolute_expiry_idx ON sessions (absolute_expires_at);      -- purge job

CREATE TABLE email_verifications (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id     uuid NOT NULL,
  email       text NOT NULL,                  -- address being verified (supports future email change)
  token_hash  bytea NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  expires_at  timestamptz NOT NULL,
  consumed_at timestamptz,
  CONSTRAINT email_verifications_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT email_verifications_token_uq UNIQUE (token_hash),
  CONSTRAINT email_verifications_hash_ck CHECK (octet_length(token_hash) = 32),
  CONSTRAINT email_verifications_email_ck CHECK (email = lower(btrim(email)) AND char_length(email) <= 254),
  CONSTRAINT email_verifications_expiry_ck CHECK (expires_at > created_at)
);
CREATE INDEX email_verifications_user_open_idx ON email_verifications (user_id) WHERE consumed_at IS NULL;

CREATE TABLE password_resets (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id     uuid NOT NULL,
  token_hash  bytea NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  expires_at  timestamptz NOT NULL,
  consumed_at timestamptz,
  CONSTRAINT password_resets_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT password_resets_token_uq UNIQUE (token_hash),
  CONSTRAINT password_resets_hash_ck CHECK (octet_length(token_hash) = 32),
  CONSTRAINT password_resets_expiry_ck CHECK (expires_at > created_at)
);
CREATE INDEX password_resets_user_open_idx ON password_resets (user_id) WHERE consumed_at IS NULL;
