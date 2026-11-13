-- Migration: 00011_taxonomy_ai_topics
-- Purpose: Add the five Learn topics and three Discover categories from E2, and relabel four existing rows. Codes of existing rows stay unchanged. No rows are deleted.
-- Tables/indexes: learn_topics (5 inserts, 3 updates), resource_categories (3 inserts, 1 update). No new indexes or constraints.
-- Lock risk: RowExclusiveLock on two small reference tables. No rewrite and no index build.
-- Data volume: 9 topic rows and 15 category rows exist before this file; 8 writes. Milliseconds.
-- Transaction: single transaction (goose default).
-- Compatibility window: additive codes and label text only. The previous application keeps working; it does not need the new codes. Existing codes payments, ai, mobile-money and sms-ussd stay valid.
-- Restore plan: forward-only. Restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest. Do not edit 00002.
-- Owner: mibienpanjoe
-- Design: TODO.md E2. Forward-only: no Down section.
-- +goose Up
INSERT INTO learn_topics (code, name_fr, name_en) VALUES
  ('learn-to-learn', 'Apprendre et rester à jour', 'Learning & keeping up'),
  ('ai-assisted-dev', 'Coder avec l’IA sans perdre la main', 'Coding with AI, keeping your skills'),
  ('ai-agents', 'Construire et évaluer des agents', 'Building & evaluating agents'),
  ('open-models', 'Modèles ouverts et IA en local', 'Open-weight & local models'),
  ('african-languages', 'IA et langues africaines', 'AI & African languages');

UPDATE learn_topics SET name_fr = 'Comprendre les LLM', name_en = 'Understanding LLMs' WHERE code = 'ai';
UPDATE learn_topics SET name_fr = 'Mobile, USSD et faible débit', name_en = 'Mobile, USSD & low bandwidth' WHERE code = 'mobile';
UPDATE learn_topics SET name_fr = 'Contribuer à l’open source', name_en = 'Contributing to open source' WHERE code = 'open-source';

UPDATE resource_categories
SET name_fr = 'Hébergement, cloud et GPU en Afrique', name_en = 'Hosting, cloud & GPUs in Africa'
WHERE code = 'cloud-hosting';

INSERT INTO resource_categories (code, name_fr, name_en, sort_order) VALUES
  ('ai-models', 'Modèles et API d’IA', 'AI models & APIs', 160),
  ('ai-tooling', 'Agents et outillage IA', 'AI agents & tooling', 170),
  ('ai-datasets', 'Jeux de données et évaluation', 'Datasets & evaluation', 180);
