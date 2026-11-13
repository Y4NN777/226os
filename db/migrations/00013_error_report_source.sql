-- Migration: 00013_error_report_source
-- Purpose: Let an error report name the resource source it is about (createResourceErrorReport source_id), without allowing a source of another resource.
-- Tables/indexes: ALTER error_reports ADD COLUMN source_id (nullable uuid) and error_reports_source_fk, a composite FOREIGN KEY (source_id, resource_id) REFERENCES resource_sources (id, resource_id) using resource_sources_id_resource_uq from 00006; error_reports_source_idx (partial, source_id IS NOT NULL).
-- Lock risk: ALTER TABLE takes a short SHARE ROW EXCLUSIVE lock to add the FK; the new column is NULL for every existing row, so validation scans trivially and nothing is rewritten.
-- Data volume: error_reports is small (one row per member report); existing rows keep source_id NULL.
-- Transaction: single transaction (goose default).
-- Compatibility window: additive and nullable; the previous image ignores the column. The FK uses MATCH SIMPLE, so a NULL source_id is never checked.
-- Restore plan: forward-only. Restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest; the column holds no data the previous image needs.
-- Owner: mibienpanjoe
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.4. Forward-only: no Down section.
-- +goose Up
ALTER TABLE error_reports ADD COLUMN source_id uuid;
ALTER TABLE error_reports
  ADD CONSTRAINT error_reports_source_fk FOREIGN KEY (source_id, resource_id)
  REFERENCES resource_sources (id, resource_id) ON DELETE RESTRICT;
CREATE INDEX error_reports_source_idx ON error_reports (source_id) WHERE source_id IS NOT NULL;
