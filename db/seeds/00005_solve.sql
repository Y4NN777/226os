-- Demo fixtures 226os: solve (questions, answers, question_tags).
-- Contract: db/seeds/README.md. Deterministic, idempotent, local-dev only.
-- Optional Solve block: two questions, two answers, no votes, no accepted answer.
-- No BEGIN/COMMIT: make seed wraps every file in one psql -1 transaction.

-- 19. questions
INSERT INTO questions (id, author_id, title, body_md, language, country_code, status, accepted_answer_id, answer_count, vote_score, provenance, published_at, removed_at, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000901', '01900000-0000-7000-8000-000000000002', 'Un webhook peut-il arriver dans le désordre ?', $q1$J'intègre un service de paiement qui m'envoie des notifications par webhook. Pendant un test, deux événements sont arrivés avec des horodatages inversés : le plus ancien a été reçu après le plus récent. J'ai donc appliqué un statut périmé et l'affichage est resté faux.

Est-ce normal de recevoir les événements dans le désordre ? Faut-il se fier à l'ordre d'arrivée, ou comparer un numéro de séquence fourni dans la charge utile ?$q1$, 'fr', 'GH', 'published', NULL, 1, 0, 'demo', '2026-11-14T09:30:00Z', NULL, '2026-11-14T09:30:00Z', '2026-11-14T09:30:00Z'),
  ('01900000-0000-7000-8000-000000000902', '01900000-0000-7000-8000-000000000001', 'Faut-il enregistrer le corps brut avant de vérifier la signature ?', $q2$Pour valider une signature, je calcule une empreinte sur le corps reçu. Si je désérialise d'abord le JSON puis que je reconstruis le corps, l'empreinte ne correspond plus : l'ordre des clés et les espaces changent.

Faut-il donc conserver le corps brut avant tout traitement ? Où le garder, et pendant combien de temps ?$q2$, 'fr', 'BF', 'published', NULL, 1, 0, 'demo', '2026-11-14T09:30:00Z', NULL, '2026-11-14T09:30:00Z', '2026-11-14T09:30:00Z')
ON CONFLICT (id) DO NOTHING;

-- 20. answers
INSERT INTO answers (id, question_id, author_id, body_md, language, status, vote_score, provenance, removed_at, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000911', '01900000-0000-7000-8000-000000000901', '01900000-0000-7000-8000-000000000003', $a1$Oui, l'ordre d'arrivée n'est pas garanti. Un émetteur réessaie souvent les notifications en échec : un événement ancien peut être renvoyé plus tard et dépasser un événement récent.

Ne traitez pas l'ordre d'arrivée comme la vérité. Si la charge utile contient un identifiant d'événement ou un numéro de séquence, mémorisez le dernier numéro appliqué et ignorez tout numéro inférieur ou déjà vu. Sinon, comparez l'horodatage de l'événement, pas celui de la réception.

Dans tous les cas, rendez le traitement idempotent : rejouer deux fois le même événement doit produire le même état final.$a1$, 'fr', 'published', 0, 'demo', NULL, '2026-11-14T09:45:00Z', '2026-11-14T09:45:00Z'),
  ('01900000-0000-7000-8000-000000000912', '01900000-0000-7000-8000-000000000902', '01900000-0000-7000-8000-000000000002', $a2$Oui. La signature porte sur les octets exacts reçus, pas sur l'objet que vous en tirez. Lisez le corps brut une seule fois, vérifiez l'empreinte dessus, puis seulement désérialisez pour traiter les données.

Gardez le corps brut en mémoire le temps de la vérification. Si vous devez le rejouer plus tard, stockez-le dans un espace à accès restreint, avec une durée de conservation courte et une purge automatique.

Comparez les empreintes en temps constant, sans fractionner la comparaison ni sortir plus tôt : c'est la seule façon d'éviter de révéler une signature partielle.$a2$, 'fr', 'published', 0, 'demo', NULL, '2026-11-14T09:45:00Z', '2026-11-14T09:45:00Z')
ON CONFLICT (id) DO NOTHING;

-- 21. question_tags (join table, no id column)
INSERT INTO question_tags (question_id, tag_id) VALUES
  ('01900000-0000-7000-8000-000000000901', '01900000-0000-7000-8000-000000000108'),
  ('01900000-0000-7000-8000-000000000901', '01900000-0000-7000-8000-000000000107'),
  ('01900000-0000-7000-8000-000000000902', '01900000-0000-7000-8000-000000000108'),
  ('01900000-0000-7000-8000-000000000902', '01900000-0000-7000-8000-000000000109')
ON CONFLICT DO NOTHING;
