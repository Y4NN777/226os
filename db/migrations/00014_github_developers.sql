-- Migration: 00014_github_developers
-- Purpose: Store the committers.top snapshot of public GitHub developers per country, listed by the Connect directory as labelled external entries (ADR-0016), and the opt-out list. No account is created for them.
-- Tables/indexes: CREATE TABLE github_developers (PRIMARY KEY (country_code, login), FK to countries, checks on login, rank, name, avatar host, contributions, organizations); github_developers_rank_idx (country_code, rank, login) for one country, github_developers_global_rank_idx (rank, country_code, login) for all countries, github_developers_login_lower_idx (lower(login)) for the profile rank; CREATE TABLE github_developer_optouts (login PRIMARY KEY, lower case only); profiles_github_lower_idx on profiles (lower(github_username)) for the member dedup.
-- Lock risk: two new tables; CREATE INDEX on profiles takes a SHARE lock that blocks profile writes for its duration (a small table at P0, milliseconds); the FK takes a short SHARE ROW EXCLUSIVE lock on countries (a small reference table, no rows are rewritten).
-- Data volume: at most 500 rows per country (10 countries at the MVP, 5000 rows), replaced daily by the API import job; the opt-out list is tiny.
-- Transaction: single transaction (goose default).
-- Compatibility window: additive; the previous image never reads these tables, and the new image lists no external entry while the tables are empty.
-- Restore plan: forward-only. Restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest; github_developers is a refreshable cache of public data (no backup required), github_developer_optouts must be kept.
-- Owner: Y4NN777
-- Design: docs/decisions/0016-connect-github-developers.md. Forward-only: no Down section.
-- +goose Up
CREATE TABLE github_developers (
  country_code  text        NOT NULL,
  login         text        NOT NULL,
  rank          integer     NOT NULL,
  name          text        NOT NULL,
  avatar_url    text        NOT NULL,
  contributions integer     NOT NULL,
  organizations text[]      NOT NULL DEFAULT '{}',
  generated_at  timestamptz NOT NULL,
  imported_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT github_developers_pkey PRIMARY KEY (country_code, login),
  CONSTRAINT github_developers_country_fk FOREIGN KEY (country_code) REFERENCES countries (code) ON DELETE RESTRICT,
  CONSTRAINT github_developers_login_ck CHECK (login ~ '^[A-Za-z0-9][A-Za-z0-9-]{0,38}$'),
  CONSTRAINT github_developers_rank_ck CHECK (rank > 0),
  CONSTRAINT github_developers_name_ck CHECK (char_length(name) BETWEEN 1 AND 200),
  CONSTRAINT github_developers_avatar_ck CHECK (avatar_url ~ '^https://avatars\.githubusercontent\.com/' AND char_length(avatar_url) <= 500),
  CONSTRAINT github_developers_contributions_ck CHECK (contributions >= 0),
  CONSTRAINT github_developers_organizations_ck CHECK (cardinality(organizations) <= 50)
);
CREATE INDEX github_developers_rank_idx ON github_developers (country_code, rank, login);
CREATE INDEX github_developers_global_rank_idx ON github_developers (rank, country_code, login);
CREATE INDEX github_developers_login_lower_idx ON github_developers (lower(login));
-- The directory drops a ranked login that a listable member shows (ADR-0016 point 3): one probe per external row.
CREATE INDEX profiles_github_lower_idx ON profiles (lower(github_username));

CREATE TABLE github_developer_optouts (
  login      text        NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT github_developer_optouts_pkey PRIMARY KEY (login),
  CONSTRAINT github_developer_optouts_login_ck CHECK (login = lower(login) AND login ~ '^[a-z0-9][a-z0-9-]{0,38}$')
);
