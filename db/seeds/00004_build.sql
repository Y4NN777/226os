-- Demo fixtures 226os: build (projects, maintainers, tags, contribution needs, need skills).
-- Contract: db/seeds/README.md. Deterministic, idempotent, local-dev only.
-- No BEGIN/COMMIT: make seed wraps every file in one psql -1 transaction.

-- 14. projects
INSERT INTO projects (id, owner_id, name, summary, description_md, language, github_url, license_spdx, stage, country_code, visibility, status, provenance, published_at, removed_at, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000501', '01900000-0000-7000-8000-000000000001', 'Suivi des prix agricoles (démo)', 'Un projet de suivi des prix agricoles pour les marchés locaux. L''objectif est de collecter des relevés de prix, de les rendre lisibles et de faciliter leur consultation par les producteurs et les commerçants. La première étape consiste à brancher une source de données et à produire un tableau de bord simple.', NULL, 'fr', NULL, NULL, 'idea', 'BF', 'public', 'published', 'demo', '2026-11-14T09:10:00Z', NULL, '2026-11-14T08:35:00Z', '2026-11-14T08:35:00Z'),
  ('01900000-0000-7000-8000-000000000502', '01900000-0000-7000-8000-000000000003', 'Adressage des quartiers (démo)', 'Un projet d''adressage des quartiers à partir de plans papier. Il vise à numériser les rues, les points utiles et les repères afin de faciliter les livraisons et l''accès aux services. La première étape consiste à saisir un plan de quartier et à vérifier la cohérence des données.', NULL, 'fr', NULL, NULL, 'idea', 'SN', 'public', 'published', 'demo', '2026-11-14T09:10:00Z', NULL, '2026-11-14T08:35:00Z', '2026-11-14T08:35:00Z')
ON CONFLICT (id) DO NOTHING;

-- 15. project_maintainers
INSERT INTO project_maintainers (project_id, user_id, added_by, created_at) VALUES
  ('01900000-0000-7000-8000-000000000501', '01900000-0000-7000-8000-000000000001', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:40:00Z'),
  ('01900000-0000-7000-8000-000000000501', '01900000-0000-7000-8000-000000000002', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:40:00Z'),
  ('01900000-0000-7000-8000-000000000502', '01900000-0000-7000-8000-000000000003', '01900000-0000-7000-8000-000000000003', '2026-11-14T08:40:00Z'),
  ('01900000-0000-7000-8000-000000000502', '01900000-0000-7000-8000-000000000001', '01900000-0000-7000-8000-000000000003', '2026-11-14T08:40:00Z')
ON CONFLICT DO NOTHING;

-- 16. project_tags
INSERT INTO project_tags (project_id, tag_id) VALUES
  ('01900000-0000-7000-8000-000000000501', '01900000-0000-7000-8000-00000000010E'),
  ('01900000-0000-7000-8000-000000000501', '01900000-0000-7000-8000-000000000102'),
  ('01900000-0000-7000-8000-000000000501', '01900000-0000-7000-8000-000000000107'),
  ('01900000-0000-7000-8000-000000000502', '01900000-0000-7000-8000-000000000105'),
  ('01900000-0000-7000-8000-000000000502', '01900000-0000-7000-8000-000000000106'),
  ('01900000-0000-7000-8000-000000000502', '01900000-0000-7000-8000-00000000010D')
ON CONFLICT DO NOTHING;

-- 17. contribution_needs
INSERT INTO contribution_needs (id, project_id, title, description_md, language, level, time_commitment, status, created_by, created_at, updated_at, closed_at) VALUES
  ('01900000-0000-7000-8000-000000000601', '01900000-0000-7000-8000-000000000501', 'Brancher une API de prix agricoles', 'Le projet collecte déjà des prix de marché dans un tableur. Nous cherchons une personne à l''aise avec les API REST pour brancher une source de prix agricoles, normaliser les réponses et enregistrer les relevés dans la base du projet. Le travail couvre l''authentification, la gestion des erreurs et un petit script de synchronisation documenté. Une première version lisible et testée suffit.', 'fr', 'intermediate', '2to5', 'open', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:45:00Z', '2026-11-14T08:45:00Z', NULL),
  ('01900000-0000-7000-8000-000000000602', '01900000-0000-7000-8000-000000000501', 'Rédiger la documentation en français', 'Nous voulons expliquer le projet à des personnes qui ne lisent pas l''anglais technique. La mission consiste à écrire une page de présentation, un guide d''installation pas à pas et une courte FAQ en français clair. Les textes doivent rester simples, éviter le jargon et s''appuyer sur le code existant. Aucune connaissance avancée du développement n''est nécessaire, mais une relecture attentive du dépôt est attendue.', 'fr', 'beginner', 'lt2', 'open', '01900000-0000-7000-8000-000000000002', '2026-11-14T08:45:00Z', '2026-11-14T08:45:00Z', NULL),
  ('01900000-0000-7000-8000-000000000603', '01900000-0000-7000-8000-000000000502', 'Numériser un plan de quartier', 'Le projet travaille sur l''adressage des quartiers à partir de plans papier. La mission consiste à numériser un plan fourni, à relever les rues et les points utiles, puis à saisir ces informations dans un format cartographique simple. Il faut être méthodique, vérifier les doublons et documenter les zones incertaines. Une courte formation sur les outils est prévue.', 'fr', 'intermediate', '2to5', 'open', '01900000-0000-7000-8000-000000000003', '2026-11-14T08:45:00Z', '2026-11-14T08:45:00Z', NULL),
  ('01900000-0000-7000-8000-000000000604', '01900000-0000-7000-8000-000000000502', 'Tester l''application sur mobile', 'L''application vient d''être adaptée aux petits écrans. Nous cherchons des personnes pour la tester sur un téléphone Android récent et sur un appareil plus ancien. Le travail consiste à suivre un scénario simple, noter les problèmes d''affichage, de lenteur ou de compréhension, puis décrire chaque défaut avec une capture d''écran. Aucun outil de développement n''est requis.', 'fr', 'beginner', 'lt2', 'open', '01900000-0000-7000-8000-000000000001', '2026-11-14T08:45:00Z', '2026-11-14T08:45:00Z', NULL)
ON CONFLICT (id) DO NOTHING;

-- 18. need_skills
INSERT INTO need_skills (need_id, tag_id) VALUES
  ('01900000-0000-7000-8000-000000000601', '01900000-0000-7000-8000-000000000107'),
  ('01900000-0000-7000-8000-000000000601', '01900000-0000-7000-8000-00000000010A'),
  ('01900000-0000-7000-8000-000000000602', '01900000-0000-7000-8000-00000000010D'),
  ('01900000-0000-7000-8000-000000000603', '01900000-0000-7000-8000-000000000105'),
  ('01900000-0000-7000-8000-000000000603', '01900000-0000-7000-8000-000000000106'),
  ('01900000-0000-7000-8000-000000000604', '01900000-0000-7000-8000-00000000010C'),
  ('01900000-0000-7000-8000-000000000604', '01900000-0000-7000-8000-00000000010D')
ON CONFLICT DO NOTHING;
