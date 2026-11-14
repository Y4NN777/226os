-- Demo fixtures 226os: discover (providers, resources, sources, countries, tags, SDK languages).
-- Contract: db/seeds/README.md. Deterministic, idempotent, local-dev only.
-- No BEGIN/COMMIT: make seed wraps every file in one psql -1 transaction.

-- 8. providers
INSERT INTO providers (id, slug, name, website_url, summary, summary_language, status, provenance, created_by, created_at, updated_at, removed_at) VALUES
  ('01900000-0000-7000-8000-000000000201', 'sahel-pay-demo', 'Sahel Pay (démo)', 'https://example.org/sahel-pay', 'Sahel Pay (démo) est un fournisseur fictif de paiements mobile money pour l''Afrique de l''Ouest. Cette fiche ne décrit aucun service réel et sert uniquement à illustrer l''annuaire Discover en local.', 'fr', 'published', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:20:00Z', '2026-11-14T08:20:00Z', NULL),
  ('01900000-0000-7000-8000-000000000202', 'baobab-sms-demo', 'Baobab SMS (démo)', 'https://example.org/baobab-sms', 'Baobab SMS (démo) est un fournisseur fictif de messagerie SMS et USSD. Ses services et ses exemples sont entièrement inventés pour la démonstration locale de 226os.', 'fr', 'published', 'demo', '01900000-0000-7000-8000-000000000002', '2026-11-14T08:20:00Z', '2026-11-14T08:20:00Z', NULL),
  ('01900000-0000-7000-8000-000000000203', 'nimba-maps-demo', 'Nimba Maps (démo)', 'https://example.org/nimba-maps', 'Nimba Maps (démo) est un fournisseur fictif de cartographie et de géodonnées. Cette fiche de démonstration ne correspond à aucune entreprise ni à aucun produit existant.', 'fr', 'published', 'demo', '01900000-0000-7000-8000-000000000003', '2026-11-14T08:20:00Z', '2026-11-14T08:20:00Z', NULL)
ON CONFLICT (id) DO NOTHING;

-- 9. resources
INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, doc_languages, website_url, docs_url, auth_note, version_note, verification_status, verified_at, verified_source_id, status, provenance, created_by, published_at, removed_at, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000301', '01900000-0000-7000-8000-000000000201', 'mobile-money', 'api', 'sahel-pay-mobile-money-demo', 'Sahel Pay Mobile Money (démo)', 'API fictive de paiement mobile money destinée aux équipes qui veulent simuler un encaissement et un remboursement. La documentation de démonstration présente des appels REST, des webhooks signés et des exemples de codes d''erreur. Aucun service réel n''est joignable.', 'fr', '{fr,en}', 'https://example.org/sahel-pay', 'https://example.org/sahel-pay/docs', NULL, NULL, 'unverified', NULL, NULL, 'published', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T09:00:00Z', NULL, '2026-11-14T08:25:00Z', '2026-11-14T08:25:00Z'),
  ('01900000-0000-7000-8000-000000000302', '01900000-0000-7000-8000-000000000202', 'sms-ussd', 'sdk', 'baobab-sms-sdk-demo', 'SDK SMS Baobab (démo)', 'SDK fictif pour envoyer des SMS et des messages USSD. La démonstration montre l''initialisation du client, l''envoi d''un message et la lecture d''un accusé de réception. Les paquets proposés pour Go et Python sont eux aussi fictifs.', 'fr', '{fr,en}', 'https://example.org/baobab-sms', 'https://example.org/baobab-sms/docs', NULL, NULL, 'unverified', NULL, NULL, 'published', 'demo', '01900000-0000-7000-8000-000000000002', '2026-11-14T09:00:00Z', NULL, '2026-11-14T08:25:00Z', '2026-11-14T08:25:00Z'),
  ('01900000-0000-7000-8000-000000000303', '01900000-0000-7000-8000-000000000203', 'maps-addressing', 'tool', 'nimba-maps-tool-demo', 'Nimba Maps Studio (démo)', 'Outil fictif de cartographie pour dessiner et annoter des plans de quartier. La documentation de démonstration décrit l''import d''un fond de carte, la saisie d''adresses et l''export des géodonnées. Aucune donnée réelle n''est utilisée.', 'fr', '{fr}', 'https://example.org/nimba-maps', 'https://example.org/nimba-maps/docs', NULL, NULL, 'unverified', NULL, NULL, 'published', 'demo', '01900000-0000-7000-8000-000000000003', '2026-11-14T09:00:00Z', NULL, '2026-11-14T08:25:00Z', '2026-11-14T08:25:00Z'),
  ('01900000-0000-7000-8000-000000000304', '01900000-0000-7000-8000-000000000201', 'dev-tools', 'tool', 'sahel-pay-devkit-demo', 'Sahel Pay Devkit (démo)', 'Trousse fictive pour développeurs qui rassemble des exemples d''intégration et un jeu de données de test. Elle illustre la configuration d''une clé d''API, la gestion des webhooks et la relecture des événements. Tout le contenu est simulé.', 'fr', '{fr,en}', 'https://example.org/sahel-pay/devkit', 'https://example.org/sahel-pay/devkit/docs', NULL, NULL, 'unverified', NULL, NULL, 'published', 'demo', '01900000-0000-7000-8000-000000000001', '2026-11-14T09:00:00Z', NULL, '2026-11-14T08:25:00Z', '2026-11-14T08:25:00Z')
ON CONFLICT (id) DO NOTHING;

-- 10. resource_sources
INSERT INTO resource_sources (id, resource_id, kind, title, url, published_at, checked_at, checked_by, created_at) VALUES
  ('01900000-0000-7000-8000-000000000401', '01900000-0000-7000-8000-000000000301', 'claim', 'Documentation publique Sahel Pay (démo)', 'https://example.org/sahel-pay/docs', NULL, NULL, NULL, '2026-11-14T08:30:00Z'),
  ('01900000-0000-7000-8000-000000000402', '01900000-0000-7000-8000-000000000302', 'claim', 'Documentation publique Baobab SMS (démo)', 'https://example.org/baobab-sms/docs', NULL, NULL, NULL, '2026-11-14T08:30:00Z'),
  ('01900000-0000-7000-8000-000000000403', '01900000-0000-7000-8000-000000000303', 'claim', 'Documentation publique Nimba Maps (démo)', 'https://example.org/nimba-maps/docs', NULL, NULL, NULL, '2026-11-14T08:30:00Z'),
  ('01900000-0000-7000-8000-000000000404', '01900000-0000-7000-8000-000000000304', 'claim', 'Documentation publique Sahel Pay Devkit (démo)', 'https://example.org/sahel-pay/devkit/docs', NULL, NULL, NULL, '2026-11-14T08:30:00Z')
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
INSERT INTO resource_tags (resource_id, tag_id) VALUES
  ('01900000-0000-7000-8000-000000000301', '01900000-0000-7000-8000-000000000101'),
  ('01900000-0000-7000-8000-000000000301', '01900000-0000-7000-8000-000000000102'),
  ('01900000-0000-7000-8000-000000000301', '01900000-0000-7000-8000-000000000107'),
  ('01900000-0000-7000-8000-000000000302', '01900000-0000-7000-8000-000000000103'),
  ('01900000-0000-7000-8000-000000000302', '01900000-0000-7000-8000-000000000104'),
  ('01900000-0000-7000-8000-000000000302', '01900000-0000-7000-8000-000000000107'),
  ('01900000-0000-7000-8000-000000000303', '01900000-0000-7000-8000-000000000105'),
  ('01900000-0000-7000-8000-000000000303', '01900000-0000-7000-8000-000000000106'),
  ('01900000-0000-7000-8000-000000000303', '01900000-0000-7000-8000-00000000010F'),
  ('01900000-0000-7000-8000-000000000304', '01900000-0000-7000-8000-000000000107'),
  ('01900000-0000-7000-8000-000000000304', '01900000-0000-7000-8000-000000000108'),
  ('01900000-0000-7000-8000-000000000304', '01900000-0000-7000-8000-00000000010A')
ON CONFLICT DO NOTHING;

-- 13. resource_sdk_languages
INSERT INTO resource_sdk_languages (id, resource_id, language, name, package_url, repo_url, version_note, created_at) VALUES
  ('01900000-0000-7000-8000-000000000451', '01900000-0000-7000-8000-000000000302', 'go', 'Baobab SMS pour Go', 'https://example.org/baobab-sms/sdk/go', NULL, NULL, '2026-11-14T08:35:00Z'),
  ('01900000-0000-7000-8000-000000000452', '01900000-0000-7000-8000-000000000302', 'python', 'Baobab SMS pour Python', 'https://example.org/baobab-sms/sdk/python', NULL, NULL, '2026-11-14T08:35:00Z'),
  ('01900000-0000-7000-8000-000000000453', '01900000-0000-7000-8000-000000000301', 'python', 'Sahel Pay pour Python', 'https://example.org/sahel-pay/sdk/python', NULL, NULL, '2026-11-14T08:35:00Z')
ON CONFLICT (id) DO NOTHING;
