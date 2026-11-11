-- Migration: 00001_extensions
-- Purpose: Install pg_trgm and unaccent and create the immutable f_unaccent(text) wrapper used by generated search columns and expression indexes.
-- Tables/indexes: none (extensions pg_trgm, unaccent; function public.f_unaccent(text)).
-- Lock risk: none beyond a brief catalog lock; no table exists yet.
-- Data volume: no data.
-- Transaction: single transaction (goose default). Needs a role allowed to create these trusted extensions; runs once.
-- Compatibility window: none (first migration); every later search migration depends on f_unaccent.
-- Restore plan: forward-only. On an empty database, drop and recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.0 (ADR-0005 points 1 and 6). Forward-only: no Down section.
-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- unaccent() is STABLE, so it cannot appear in generated columns or expression
-- indexes. This wrapper pins the dictionary schema-qualified and is IMMUTABLE.
-- +goose StatementBegin
CREATE FUNCTION f_unaccent(text) RETURNS text
  LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
  AS $$ SELECT public.unaccent('public.unaccent'::regdictionary, $1) $$;
-- +goose StatementEnd
