-- Demo fixtures 226os: identity (users, profiles, tags).
-- Contract: db/seeds/README.md. Deterministic, idempotent, local-dev only.
-- No BEGIN/COMMIT: make seed wraps every file in one psql -1 transaction.

-- 1. users
INSERT INTO users (id, email, email_verified_at, role, status, provenance, created_at, updated_at, deleted_at) VALUES
  ('01900000-0000-7000-8000-000000000001', 'demo-aicha@demo.226os.invalid', '2026-11-14T08:05:00Z', 'member', 'active', 'demo', '2026-11-14T08:00:00Z', '2026-11-14T08:00:00Z', NULL),
  ('01900000-0000-7000-8000-000000000002', 'demo-kofi@demo.226os.invalid', '2026-11-14T08:05:00Z', 'member', 'active', 'demo', '2026-11-14T08:00:00Z', '2026-11-14T08:00:00Z', NULL),
  ('01900000-0000-7000-8000-000000000003', 'demo-nadia@demo.226os.invalid', '2026-11-14T08:05:00Z', 'member', 'active', 'demo', '2026-11-14T08:00:00Z', '2026-11-14T08:00:00Z', NULL)
ON CONFLICT (id) DO NOTHING;

-- 2. profiles (user_id is the primary key)
INSERT INTO profiles (user_id, username, display_name, intent, country_code, languages, github_username, bio, bio_language, open_to_collab, country_visibility, languages_visibility, skills_visibility, github_visibility, bio_visibility, open_to_collab_visibility, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000001', 'demo-aicha', 'Aïcha (démo)', 'learn', 'BF', '{fr,mos}', NULL, NULL, NULL, true, 'public', 'public', 'public', 'public', 'public', 'public', '2026-11-14T08:10:00Z', '2026-11-14T08:10:00Z'),
  ('01900000-0000-7000-8000-000000000002', 'demo-kofi', 'Kofi (démo)', 'contribute', 'GH', '{fr,en}', NULL, NULL, NULL, true, 'public', 'public', 'public', 'public', 'public', 'public', '2026-11-14T08:10:00Z', '2026-11-14T08:10:00Z'),
  ('01900000-0000-7000-8000-000000000003', 'demo-nadia', 'Nadia (démo)', 'help', 'SN', '{fr,wo}', NULL, NULL, NULL, true, 'public', 'public', 'public', 'public', 'public', 'public', '2026-11-14T08:10:00Z', '2026-11-14T08:10:00Z')
ON CONFLICT (user_id) DO NOTHING;

-- 3. tags
INSERT INTO tags (id, slug, label, status, provenance, created_by, created_at) VALUES
  ('01900000-0000-7000-8000-000000000101', 'mobile-money', 'Mobile money', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-000000000102', 'paiements', 'Paiements', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-000000000103', 'sms', 'SMS', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-000000000104', 'ussd', 'USSD', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-000000000105', 'cartographie', 'Cartographie', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-000000000106', 'geodonnees', 'Géodonnées', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-000000000107', 'api', 'API', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-000000000108', 'webhooks', 'Webhooks', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-000000000109', 'securite', 'Sécurité', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-00000000010A', 'python', 'Python', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-00000000010B', 'go', 'Go', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-00000000010C', 'typescript', 'TypeScript', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-00000000010D', 'open-source', 'Open source', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-00000000010E', 'agriculture', 'Agriculture', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z'),
  ('01900000-0000-7000-8000-00000000010F', 'logistique', 'Logistique', 'active', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:15:00Z')
ON CONFLICT (id) DO NOTHING;
