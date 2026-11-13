-- Migration: 00012_profile_visibility
-- Purpose: Add the whole-profile visibility column decided in C14b. New and existing profiles default to public. This is separate from the per-field profiles_visibility_ck check.
-- Tables/indexes: profiles.profile_visibility text NOT NULL DEFAULT public, named check profiles_profile_visibility_ck. No new index.
-- Lock risk: ACCESS EXCLUSIVE on profiles. A constant default does not rewrite the table. The check then reads the existing rows.
-- Data volume: one column on profiles. The table is empty at bootstrap and small at P0. Milliseconds.
-- Transaction: single transaction (goose default).
-- Compatibility window: the previous application omits the column on insert, so the default applies. It does not read the column. Values are public, members or private.
-- Restore plan: forward-only. Restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest. Do not edit 00003.
-- Owner: mibienpanjoe
-- Design: TODO.md E2, contract C14b. Forward-only: no Down section.
-- +goose Up
ALTER TABLE profiles
  ADD COLUMN profile_visibility text NOT NULL DEFAULT 'public';

ALTER TABLE profiles
  ADD CONSTRAINT profiles_profile_visibility_ck
  CHECK (profile_visibility IN ('public', 'members', 'private'));
