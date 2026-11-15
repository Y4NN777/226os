-- Demo fixtures 226os: discover (providers, resources, sources, countries, tags, SDK languages).
-- Contract: db/seeds/README.md. Deterministic, idempotent, local-dev and staging-test only.
-- No BEGIN/COMMIT: make seed wraps every file in one psql -1 transaction.

-- 8. providers
INSERT INTO providers (id, slug, name, website_url, summary, summary_language, status, provenance, created_by, created_at, updated_at, removed_at) VALUES
  ('01900000-0000-7000-8000-000000000201', 'sahel-pay-demo', 'Sahel Pay', 'https://example.org/sahel-pay', 'Sahel Pay est un fournisseur fictif de paiements mobile money pour l''Afrique de l''Ouest.', 'fr', 'published', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:20:00Z', '2026-11-14T08:20:00Z', NULL),
  ('01900000-0000-7000-8000-000000000202', 'baobab-sms-demo', 'Baobab SMS', 'https://example.org/baobab-sms', 'Baobab SMS est un fournisseur fictif de messagerie SMS et USSD.', 'fr', 'published', 'demo', '01900000-0000-7000-8000-000000000002', '2026-11-14T08:20:00Z', '2026-11-14T08:20:00Z', NULL),
  ('01900000-0000-7000-8000-000000000203', 'nimba-maps-demo', 'Nimba Maps', 'https://example.org/nimba-maps', 'Nimba Maps est un fournisseur fictif de cartographie et de géodonnées.', 'fr', 'published', 'demo', '01900000-0000-7000-8000-000000000003', '2026-11-14T08:20:00Z', '2026-11-14T08:20:00Z', NULL)
ON CONFLICT (id) DO NOTHING;

-- 9. resources
INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, doc_languages, website_url, docs_url, auth_note, version_note, verification_status, verified_at, verified_source_id, status, provenance, created_by, published_at, removed_at, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000301', '01900000-0000-7000-8000-000000000201', 'mobile-money', 'api', 'sahel-pay-mobile-money-demo', 'Sahel Pay Mobile Money', 'API fictive de paiement mobile money pour simuler un encaissement et un remboursement : appels REST, webhooks signés et exemples de codes d''erreur.', 'fr', '{fr,en}', 'https://example.org/sahel-pay', 'https://example.org/sahel-pay/docs', NULL, NULL, 'unverified', NULL, NULL, 'published', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T09:00:00Z', NULL, '2026-11-14T08:25:00Z', '2026-11-14T08:25:00Z'),
  ('01900000-0000-7000-8000-000000000302', '01900000-0000-7000-8000-000000000202', 'sms-ussd', 'sdk', 'baobab-sms-sdk-demo', 'SDK SMS Baobab', 'SDK fictif pour envoyer des SMS et des messages USSD : initialisation du client, envoi d''un message et lecture d''un accusé de réception, en Go et en Python.', 'fr', '{fr,en}', 'https://example.org/baobab-sms', 'https://example.org/baobab-sms/docs', NULL, NULL, 'unverified', NULL, NULL, 'published', 'demo', '01900000-0000-7000-8000-000000000002', '2026-11-14T09:00:00Z', NULL, '2026-11-14T08:25:00Z', '2026-11-14T08:25:00Z'),
  ('01900000-0000-7000-8000-000000000303', '01900000-0000-7000-8000-000000000203', 'maps-addressing', 'tool', 'nimba-maps-tool-demo', 'Nimba Maps Studio', 'Outil fictif de cartographie pour dessiner et annoter des plans de quartier : import d''un fond de carte, saisie d''adresses et export des géodonnées.', 'fr', '{fr}', 'https://example.org/nimba-maps', 'https://example.org/nimba-maps/docs', NULL, NULL, 'unverified', NULL, NULL, 'published', 'demo', '01900000-0000-7000-8000-000000000003', '2026-11-14T09:00:00Z', NULL, '2026-11-14T08:25:00Z', '2026-11-14T08:25:00Z'),
  ('01900000-0000-7000-8000-000000000304', '01900000-0000-7000-8000-000000000201', 'dev-tools', 'tool', 'sahel-pay-devkit-demo', 'Sahel Pay Devkit', 'Trousse fictive pour développeurs : exemples d''intégration, jeu de données de test, configuration d''une clé d''API, gestion des webhooks et relecture des événements.', 'fr', '{fr,en}', 'https://example.org/sahel-pay/devkit', 'https://example.org/sahel-pay/devkit/docs', NULL, NULL, 'unverified', NULL, NULL, 'published', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T09:00:00Z', NULL, '2026-11-14T08:25:00Z', '2026-11-14T08:25:00Z')
ON CONFLICT (id) DO NOTHING;

-- 10. resource_sources
INSERT INTO resource_sources (id, resource_id, kind, title, url, published_at, checked_at, checked_by, created_at) VALUES
  ('01900000-0000-7000-8000-000000000401', '01900000-0000-7000-8000-000000000301', 'claim', 'Documentation publique Sahel Pay', 'https://example.org/sahel-pay/docs', NULL, NULL, NULL, '2026-11-14T08:30:00Z'),
  ('01900000-0000-7000-8000-000000000402', '01900000-0000-7000-8000-000000000302', 'claim', 'Documentation publique Baobab SMS', 'https://example.org/baobab-sms/docs', NULL, NULL, NULL, '2026-11-14T08:30:00Z'),
  ('01900000-0000-7000-8000-000000000403', '01900000-0000-7000-8000-000000000303', 'claim', 'Documentation publique Nimba Maps', 'https://example.org/nimba-maps/docs', NULL, NULL, NULL, '2026-11-14T08:30:00Z'),
  ('01900000-0000-7000-8000-000000000404', '01900000-0000-7000-8000-000000000304', 'claim', 'Documentation publique Sahel Pay Devkit', 'https://example.org/sahel-pay/devkit/docs', NULL, NULL, NULL, '2026-11-14T08:30:00Z')
ON CONFLICT (id) DO NOTHING;

-- 11. resource_countries
INSERT INTO resource_countries (resource_id, country_code) VALUES
  ('01900000-0000-7000-8000-000000000301', 'BJ'),
  ('01900000-0000-7000-8000-000000000301', 'BF'),
  ('01900000-0000-7000-8000-000000000301', 'CI'),
  ('01900000-0000-7000-8000-000000000301', 'SN'),
  ('01900000-0000-7000-8000-000000000302', 'GH'),
  ('01900000-0000-7000-8000-000000000302', 'NG'),
  ('01900000-0000-7000-8000-000000000302', 'TG'),
  ('01900000-0000-7000-8000-000000000303', 'CI'),
  ('01900000-0000-7000-8000-000000000303', 'SN'),
  ('01900000-0000-7000-8000-000000000304', 'BF'),
  ('01900000-0000-7000-8000-000000000304', 'GH')
ON CONFLICT DO NOTHING;

-- 12. resource_tags
INSERT INTO resource_tags (resource_id, tag_id)
SELECT v.resource_id, t.id FROM (VALUES
  ('01900000-0000-7000-8000-000000000301'::uuid, 'mobile-money'),
  ('01900000-0000-7000-8000-000000000301'::uuid, 'paiements'),
  ('01900000-0000-7000-8000-000000000301'::uuid, 'api'),
  ('01900000-0000-7000-8000-000000000302'::uuid, 'sms'),
  ('01900000-0000-7000-8000-000000000302'::uuid, 'ussd'),
  ('01900000-0000-7000-8000-000000000302'::uuid, 'api'),
  ('01900000-0000-7000-8000-000000000303'::uuid, 'cartographie'),
  ('01900000-0000-7000-8000-000000000303'::uuid, 'geodonnees'),
  ('01900000-0000-7000-8000-000000000303'::uuid, 'logistique'),
  ('01900000-0000-7000-8000-000000000304'::uuid, 'api'),
  ('01900000-0000-7000-8000-000000000304'::uuid, 'webhooks'),
  ('01900000-0000-7000-8000-000000000304'::uuid, 'python')
) AS v(resource_id, slug) JOIN tags t ON t.slug = v.slug
ON CONFLICT DO NOTHING;

-- 13. resource_sdk_languages
INSERT INTO resource_sdk_languages (id, resource_id, language, name, package_url, repo_url, version_note, created_at) VALUES
  ('01900000-0000-7000-8000-000000000451', '01900000-0000-7000-8000-000000000302', 'go', 'Baobab SMS pour Go', 'https://example.org/baobab-sms/sdk/go', NULL, NULL, '2026-11-14T08:35:00Z'),
  ('01900000-0000-7000-8000-000000000452', '01900000-0000-7000-8000-000000000302', 'python', 'Baobab SMS pour Python', 'https://example.org/baobab-sms/sdk/python', NULL, NULL, '2026-11-14T08:35:00Z'),
  ('01900000-0000-7000-8000-000000000453', '01900000-0000-7000-8000-000000000301', 'python', 'Sahel Pay pour Python', 'https://example.org/sahel-pay/sdk/python', NULL, NULL, '2026-11-14T08:35:00Z')
ON CONFLICT (id) DO NOTHING;
