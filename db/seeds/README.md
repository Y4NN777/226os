# Fixtures de démo 226os : contrat

Ce document est le contrat des fixtures de démo du slice SEED #75. Il fixe les identifiants, les comptes, les tags, le contenu et le découpage SQL. Il ne contient aucun SQL : les fichiers sont écrits par les agents T3a à T3e.

Les fixtures sont fictives, déterministes et idempotentes. Elles ne tournent qu'en `local-dev`, jamais en staging ni en production. Toute ligne racine porte `provenance = 'demo'` et l'interface affiche un badge de démo.

Les chiffres de ce document sont contractuels. Une fois la PR mergée, les UUID et les valeurs naturelles listées ici ne changent plus.

## 1. Convention des UUID

Tous les identifiants utilisent la plage :

```text
01900000-0000-7000-8000-000000000NNN
```

Les trois derniers caractères hexadécimaux identifient l'objet. Le préfixe ne dépend ni de la machine, ni de la date, ni d'un générateur aléatoire. Aucun appel à `uuidv7()` n'est autorisé dans `db/seeds/`.

Les blocs sont réservés par table :

| Bloc | Table |
|---|---|
| `001` à `0FF` | `users` et `profiles` |
| `101` à `1FF` | `tags` |
| `201` à `2FF` | `providers` |
| `301` à `3FF` | `resources` |
| `401` à `4FF` | `resource_sources` |
| `451` à `4FF` | `resource_sdk_languages` |
| `501` à `5FF` | `projects` |
| `601` à `6FF` | `contribution_needs` |
| `701` à `7FF` | `learn_articles` |
| `801` à `8FF` | `training_series` |
| `811` à `8FF` | `training_lessons` |
| `901` à `9FF` | `questions` |
| `911` à `9FF` | `answers` |

Deux règles de détail :

- `profiles.user_id` est la clé primaire et vaut l'identifiant du `users` correspondant. Les trois profils réutilisent donc `001`, `002` et `003`.
- Les tables de jointure (`resource_countries`, `resource_tags`, `project_maintainers`, `project_tags`, `need_skills`, `learn_article_tags`, `question_tags`) n'ont pas de colonne `id`. Elles n'ont pas d'UUID propre.

### 1.1 Utilisateurs et profils

| UUID | Username | Nom affiché | Email | Pays | Rôle | Statut |
|---|---|---|---|---|---|---|
| `01900000-0000-7000-8000-000000000001` | `demo-aicha` | `Aïcha (démo)` | `demo-aicha@demo.226os.invalid` | `BF` | `member` | `active` |
| `01900000-0000-7000-8000-000000000002` | `demo-kofi` | `Kofi (démo)` | `demo-kofi@demo.226os.invalid` | `GH` | `member` | `active` |
| `01900000-0000-7000-8000-000000000003` | `demo-nadia` | `Nadia (démo)` | `demo-nadia@demo.226os.invalid` | `SN` | `member` | `active` |

Valeurs communes aux trois lignes `users` :

- `email_verified_at = '2026-11-14T08:05:00Z'`, valeur littérale.
- `deleted_at = NULL`.
- `provenance = 'demo'`.
- `created_at = updated_at = '2026-11-14T08:00:00Z'`.
- Aucune ligne dans `user_credentials` pour ces trois comptes.

Valeurs communes aux trois lignes `profiles` :

- `intent` : `learn` pour Aïcha, `contribute` pour Kofi, `help` pour Nadia.
- `languages` : `{fr,mos}` pour Aïcha, `{fr,en}` pour Kofi, `{fr,wo}` pour Nadia.
- `github_username = NULL`, `bio = NULL`, `bio_language = NULL`.
- `open_to_collab = true`.
- Toutes les colonnes de visibilité à `public`.
- `created_at = updated_at = '2026-11-14T08:10:00Z'`.

### 1.2 Tags

Tous les tags ont `status = 'active'`, `provenance = 'demo'`, `created_by = '01900000-0000-7000-8000-000000000001'` et `created_at = '2026-11-14T08:15:00Z'`.

| UUID | Slug | Label |
|---|---|---|
| `01900000-0000-7000-8000-000000000101` | `mobile-money` | `Mobile money` |
| `01900000-0000-7000-8000-000000000102` | `paiements` | `Paiements` |
| `01900000-0000-7000-8000-000000000103` | `sms` | `SMS` |
| `01900000-0000-7000-8000-000000000104` | `ussd` | `USSD` |
| `01900000-0000-7000-8000-000000000105` | `cartographie` | `Cartographie` |
| `01900000-0000-7000-8000-000000000106` | `geodonnees` | `Géodonnées` |
| `01900000-0000-7000-8000-000000000107` | `api` | `API` |
| `01900000-0000-7000-8000-000000000108` | `webhooks` | `Webhooks` |
| `01900000-0000-7000-8000-000000000109` | `securite` | `Sécurité` |
| `01900000-0000-7000-8000-00000000010A` | `python` | `Python` |
| `01900000-0000-7000-8000-00000000010B` | `go` | `Go` |
| `01900000-0000-7000-8000-00000000010C` | `typescript` | `TypeScript` |
| `01900000-0000-7000-8000-00000000010D` | `open-source` | `Open source` |
| `01900000-0000-7000-8000-00000000010E` | `agriculture` | `Agriculture` |
| `01900000-0000-7000-8000-00000000010F` | `logistique` | `Logistique` |

## 2. Discover

### 2.1 Providers

Tous les providers ont `status = 'published'`, `provenance = 'demo'`, `summary_language = 'fr'` et `removed_at = NULL`. Le résumé court est rédigé par T3c, dans la limite de 1000 caractères.

| UUID | Slug | Nom | Site | Créé par |
|---|---|---|---|---|
| `01900000-0000-7000-8000-000000000201` | `sahel-pay-demo` | `Sahel Pay (démo)` | `https://example.org/sahel-pay` | `001` |
| `01900000-0000-7000-8000-000000000202` | `baobab-sms-demo` | `Baobab SMS (démo)` | `https://example.org/baobab-sms` | `002` |
| `01900000-0000-7000-8000-000000000203` | `nimba-maps-demo` | `Nimba Maps (démo)` | `https://example.org/nimba-maps` | `003` |

`created_at = updated_at = '2026-11-14T08:20:00Z'`.

### 2.2 Resources

Toutes les resources ont `verification_status = 'unverified'`, `verified_at = NULL`, `verified_source_id = NULL`, `status = 'published'`, `provenance = 'demo'`, `summary_language = 'fr'`, `removed_at = NULL`, `created_at = updated_at = '2026-11-14T08:25:00Z'` et `published_at = '2026-11-14T09:00:00Z'`. Le résumé est rédigé par T3c, dans la limite de 2000 caractères.

| UUID | Slug | Nom | Type | Catégorie | Provider | Créé par |
|---|---|---|---|---|---|---|
| `01900000-0000-7000-8000-000000000301` | `sahel-pay-mobile-money-demo` | `Sahel Pay Mobile Money (démo)` | `api` | `mobile-money` | `201` | `001` |
| `01900000-0000-7000-8000-000000000302` | `baobab-sms-sdk-demo` | `SDK SMS Baobab (démo)` | `sdk` | `sms-ussd` | `202` | `002` |
| `01900000-0000-7000-8000-000000000303` | `nimba-maps-tool-demo` | `Nimba Maps Studio (démo)` | `tool` | `maps-addressing` | `203` | `003` |
| `01900000-0000-7000-8000-000000000304` | `sahel-pay-devkit-demo` | `Sahel Pay Devkit (démo)` | `tool` | `dev-tools` | `201` | `001` |

URL et langues de documentation :

| Resource | `website_url` | `docs_url` | `doc_languages` |
|---|---|---|---|
| `301` | `https://example.org/sahel-pay` | `https://example.org/sahel-pay/docs` | `{fr,en}` |
| `302` | `https://example.org/baobab-sms` | `https://example.org/baobab-sms/docs` | `{fr,en}` |
| `303` | `https://example.org/nimba-maps` | `https://example.org/nimba-maps/docs` | `{fr}` |
| `304` | `https://example.org/sahel-pay/devkit` | `https://example.org/sahel-pay/devkit/docs` | `{fr,en}` |

Pays :

| Resource | Codes pays |
|---|---|
| `301` | `BJ`, `BF`, `CI`, `SN` |
| `302` | `GH`, `NG`, `TG` |
| `303` | `CI`, `SN` |
| `304` | `BF`, `GH` |

Tags :

| Resource | Tags |
|---|---|
| `301` | `101`, `102`, `107` |
| `302` | `103`, `104`, `107` |
| `303` | `105`, `106`, `10F` |
| `304` | `107`, `108`, `10A` |

Sources. Toutes les sources ont `kind = 'claim'`, `checked_at = NULL`, `checked_by = NULL`, `published_at = NULL` et `created_at = '2026-11-14T08:30:00Z'`. Le titre fait entre 1 et 200 caractères.

| UUID | Resource | Titre | URL |
|---|---|---|---|
| `01900000-0000-7000-8000-000000000401` | `301` | `Documentation publique Sahel Pay (démo)` | `https://example.org/sahel-pay/docs` |
| `01900000-0000-7000-8000-000000000402` | `302` | `Documentation publique Baobab SMS (démo)` | `https://example.org/baobab-sms/docs` |
| `01900000-0000-7000-8000-000000000403` | `303` | `Documentation publique Nimba Maps (démo)` | `https://example.org/nimba-maps/docs` |
| `01900000-0000-7000-8000-000000000404` | `304` | `Documentation publique Sahel Pay Devkit (démo)` | `https://example.org/sahel-pay/devkit/docs` |

Langages de SDK. Chaque ligne a `created_at = '2026-11-14T08:35:00Z'` et des URLs `example.org`.

| UUID | Resource | `language` | `name` | `package_url` |
|---|---|---|---|---|
| `01900000-0000-7000-8000-000000000451` | `302` | `go` | `Baobab SMS pour Go` | `https://example.org/baobab-sms/sdk/go` |
| `01900000-0000-7000-8000-000000000452` | `302` | `python` | `Baobab SMS pour Python` | `https://example.org/baobab-sms/sdk/python` |
| `01900000-0000-7000-8000-000000000453` | `301` | `python` | `Sahel Pay pour Python` | `https://example.org/sahel-pay/sdk/python` |

`repo_url` et `version_note` restent `NULL` pour les trois lignes.

## 3. Build

### 3.1 Projects

Les deux projets sont en `stage = 'idea'`, donc `github_url = NULL` et `license_spdx = NULL`. Ils sont `visibility = 'public'`, `status = 'published'`, `provenance = 'demo'`, `language = 'fr'`, `removed_at = NULL`, `created_at = updated_at = '2026-11-14T08:35:00Z'` et `published_at = '2026-11-14T09:10:00Z'`.

| UUID | Nom | Owner | Pays | Résumé |
|---|---|---|---|---|
| `01900000-0000-7000-8000-000000000501` | `Suivi des prix agricoles (démo)` | `001` | `BF` | Court, rédigé par T3d, 500 caractères au plus |
| `01900000-0000-7000-8000-000000000502` | `Adressage des quartiers (démo)` | `003` | `SN` | Court, rédigé par T3d, 500 caractères au plus |

Mainteneurs. Chaque ligne a `created_at = '2026-11-14T08:40:00Z'`.

| Project | User | `added_by` |
|---|---|---|
| `501` | `001` | `001` |
| `501` | `002` | `001` |
| `502` | `003` | `003` |
| `502` | `001` | `003` |

Tags :

| Project | Tags |
|---|---|
| `501` | `10E`, `102`, `107` |
| `502` | `105`, `106`, `10D` |

### 3.2 Besoins de contribution

Les quatre besoins sont `status = 'open'`, `closed_at = NULL`, `language = 'fr'`, `provenance` portée par le projet, `created_at = updated_at = '2026-11-14T08:45:00Z'`. `description_md` est rédigé par T3d, entre 1 et 5000 caractères.

| UUID | Project | Titre | Niveau | Temps | Créé par |
|---|---|---|---|---|---|
| `01900000-0000-7000-8000-000000000601` | `501` | `Brancher une API de prix agricoles` | `intermediate` | `2to5` | `001` |
| `01900000-0000-7000-8000-000000000602` | `501` | `Rédiger la documentation en français` | `beginner` | `lt2` | `002` |
| `01900000-0000-7000-8000-000000000603` | `502` | `Numériser un plan de quartier` | `intermediate` | `2to5` | `003` |
| `01900000-0000-7000-8000-000000000604` | `502` | `Tester l'application sur mobile` | `beginner` | `lt2` | `001` |

Compétences demandées :

| Besoin | Tags |
|---|---|
| `601` | `107`, `10A` |
| `602` | `10D` |
| `603` | `105`, `106` |
| `604` | `10C`, `10D` |

Aucune ligne dans `applications`.

## 4. Learn

### 4.1 Articles

Les quatre articles sont `kind = 'tutorial'`, `language = 'fr'`, `topic_code = 'api-integration'`, `status = 'published'`, `provenance = 'demo'`, `author_id = '01900000-0000-7000-8000-000000000003'`, `removed_at = NULL`. Le `body_md` est le texte long original écrit par T3b, entre 1 et 50000 caractères. Le `summary` éventuel est aussi rédigé par T3b.

| UUID | Titre | Niveau | `published_at` | `created_at` et `updated_at` |
|---|---|---|---|---|
| `01900000-0000-7000-8000-000000000701` | `Comprendre un webhook` | `beginner` | `2026-11-10T09:00:00Z` | `2026-11-10T09:00:00Z` |
| `01900000-0000-7000-8000-000000000702` | `Vérifier une signature` | `intermediate` | `2026-11-11T09:00:00Z` | `2026-11-11T09:00:00Z` |
| `01900000-0000-7000-8000-000000000703` | `Rendre le traitement idempotent` | `intermediate` | `2026-11-12T09:00:00Z` | `2026-11-12T09:00:00Z` |
| `01900000-0000-7000-8000-000000000704` | `Tester et rejouer` | `intermediate` | `2026-11-13T09:00:00Z` | `2026-11-13T09:00:00Z` |

Tags des articles :

| Article | Tags |
|---|---|
| `701` | `108`, `107` |
| `702` | `108`, `109` |
| `703` | `108`, `109`, `107` |
| `704` | `108`, `107` |

### 4.2 Série et leçons

Une seule série. Elle est `origin = 'platform'`, `provenance = 'demo'`, `author_id = '01900000-0000-7000-8000-000000000003'`, `title = 'Premiers pas avec les webhooks'`, `summary` rédigé par T3b (1000 caractères au plus), `language = 'fr'`, `level = 'beginner'`, `topic_code = 'api-integration'`, `status = 'published'`, `removed_at = NULL`, `published_at = created_at = updated_at = '2026-11-13T18:00:00Z'`.

| UUID | Type |
|---|---|
| `01900000-0000-7000-8000-000000000801` | `training_series` |

Les quatre leçons pointent vers les quatre articles. Elles ont `language = 'fr'`, `status = 'published'`, `body_md = NULL`, `removed_at = NULL`, et `created_at = updated_at` à la minute indiquée.

| UUID | `series_id` | `position` | Titre | `article_id` | `created_at` |
|---|---|---|---|---|---|
| `01900000-0000-7000-8000-000000000811` | `801` | `1` | `Comprendre un webhook` | `701` | `2026-11-13T18:10:00Z` |
| `01900000-0000-7000-8000-000000000812` | `801` | `2` | `Vérifier une signature` | `702` | `2026-11-13T18:11:00Z` |
| `01900000-0000-7000-8000-000000000813` | `801` | `3` | `Rendre le traitement idempotent` | `703` | `2026-11-13T18:12:00Z` |
| `01900000-0000-7000-8000-000000000814` | `801` | `4` | `Tester et rejouer` | `704` | `2026-11-13T18:13:00Z` |

## 5. Solve (optionnel)

Le bloc Solve ne fait partie de la PR que si T3e a le temps. S'il est présent, il respecte exactement les valeurs ci-dessous.

### 5.1 Questions

Les deux questions ont `language = 'fr'`, `status = 'published'`, `provenance = 'demo'`, `answer_count = 1`, `vote_score = 0`, `accepted_answer_id = NULL`, `removed_at = NULL`, `published_at = created_at = updated_at = '2026-11-14T09:30:00Z'`. `body_md` est rédigé par T3e.

| UUID | Auteur | Titre | Pays |
|---|---|---|---|
| `01900000-0000-7000-8000-000000000901` | `002` | `Un webhook peut-il arriver dans le désordre ?` | `GH` |
| `01900000-0000-7000-8000-000000000902` | `001` | `Faut-il enregistrer le corps brut avant de vérifier la signature ?` | `BF` |

Tags :

| Question | Tags |
|---|---|
| `901` | `108`, `107` |
| `902` | `108`, `109` |

### 5.2 Réponses

Les deux réponses ont `language = 'fr'`, `status = 'published'`, `provenance = 'demo'`, `vote_score = 0`, `removed_at = NULL`, `created_at = updated_at = '2026-11-14T09:45:00Z'`. `body_md` est rédigé par T3e. Aucune réponse n'est acceptée.

| UUID | Question | Auteur |
|---|---|---|
| `01900000-0000-7000-8000-000000000911` | `901` | `003` |
| `01900000-0000-7000-8000-000000000912` | `902` | `002` |

Aucune ligne dans `votes`.

## 6. Découpage des fichiers SQL

Les fichiers vivent dans `db/seeds/` et sont exécutés par `make seed` :

| Fichier | Contenu |
|---|---|
| `00001_identity.sql` | `users`, `profiles`, `tags` |
| `00002_learn.sql` | `learn_articles`, `learn_article_tags`, `training_series`, `training_lessons` |
| `00003_discover.sql` | `providers`, `resources`, `resource_sources`, `resource_countries`, `resource_tags`, `resource_sdk_languages` |
| `00004_build.sql` | `projects`, `project_maintainers`, `project_tags`, `contribution_needs`, `need_skills` |
| `00005_solve.sql` | `questions`, `answers`, `question_tags`, uniquement si le bloc Solve est livré |

`make seed` liste les fichiers de `db/seeds/*.sql` et les exécute dans l'ordre lexical. Les préfixes `00001` à `00005` garantissent donc l'ordre. L'exécution passe par `psql -X -v ON_ERROR_STOP=1 -1` et une seule transaction : si un fichier échoue, rien n'est écrit. Les fichiers eux-mêmes ne contiennent ni `BEGIN` ni `COMMIT`, c'est `psql -1` qui porte l'atomicité.

## 7. Ordre d'insertion FK exact

L'ordre global, dans l'ordre des fichiers, est :

```text
00001_identity.sql
  1. users
  2. profiles
  3. tags

00002_learn.sql
  4. learn_articles
  5. learn_article_tags
  6. training_series
  7. training_lessons

00003_discover.sql
  8. providers
  9. resources
  10. resource_sources
  11. resource_countries
  12. resource_tags
  13. resource_sdk_languages

00004_build.sql
  14. projects
  15. project_maintainers
  16. project_tags
  17. contribution_needs
  18. need_skills

00005_solve.sql (optionnel)
  19. questions
  20. answers
  21. question_tags
```

Chaque table ne référence que des lignes insérées plus haut. `question_sources` et `answer_sources` ne sont pas alimentées. Si un agent doit en ajouter, il insère ces lignes après `questions` et `answers`, jamais avant.

## 8. Règles

### 8.1 Identifiants et temps

- UUID fixes uniquement, selon la plage et les tableaux ci-dessus.
- Timestamps littéraux, au format `'2026-11-14T08:00:00Z'`.
- Jamais `now()`, `CURRENT_TIMESTAMP`, `uuidv7()`, `random()` ou une valeur calculée à l'exécution.

Tous les `created_at`, `updated_at`, `published_at`, `email_verified_at` et `created_at` des tables de jointure sont explicitement fournis, même quand la colonne a une valeur par défaut. Deux exécutions sur une base vide produisent les mêmes lignes.

### 8.2 Idempotence

Sur chaque table qui a une colonne `id`, l'insertion se termine par :

```sql
ON CONFLICT (id) DO NOTHING
```

`profiles` fait exception : sa clé primaire s'appelle `user_id`, donc la clause est :

```sql
ON CONFLICT (user_id) DO NOTHING
```

Sur les tables de jointure sans colonne `id`, la forme correcte est :

```sql
ON CONFLICT DO NOTHING
```

Une deuxième exécution ne crée aucune ligne et ne modifie aucune ligne existante.

### 8.3 Provenance

`provenance = 'demo'` est obligatoire sur chaque ligne racine :

- `users`
- `tags`
- `providers`
- `resources`
- `questions`
- `answers`
- `projects`
- `learn_articles`
- `training_series`

Les tables filles (`profiles`, `resource_sources`, `resource_countries`, `resource_tags`, `resource_sdk_languages`, `project_maintainers`, `project_tags`, `contribution_needs`, `need_skills`, `learn_article_tags`, `training_lessons`, `question_tags`) n'ont pas de colonne `provenance` et héritent de leur racine.

### 8.4 Interdits absolus

- Aucune ligne dans `user_credentials`.
- Aucune ligne dans `votes`, ni sur une question, ni sur une réponse.
- Aucune ligne dans `applications`.
- Aucune ligne dans `sessions`, `email_verifications` ou `password_resets`.
- Aucun `resources.verification_status` autre que `unverified`.
- Aucun `verified_source_id` et aucun `verified_at`.
- Aucun `questions.accepted_answer_id`.
- Aucun `questions.answer_count` incohérent : il vaut exactement le nombre de réponses insérées, soit `1` pour `901` et `902`.
- Aucune URL hors `https://example.org/...`.
- Aucun email hors du domaine `@demo.226os.invalid`.
- Aucune entreprise, personne, marque ou donnée réelle.
- Aucun contenu copié depuis un site externe.

## 9. Inventaire du contenu

Le contenu est original et court. T3b, T3c, T3d et T3e écrivent les textes longs dans le respect de cet inventaire.

Comptes : Aïcha (Burkina Faso), Kofi (Ghana), Nadia (Sénégal), tous en démo.

Learn, série « Premiers pas avec les webhooks », quatre leçons :

1. Comprendre un webhook
2. Vérifier une signature
3. Rendre le traitement idempotent
4. Tester et rejouer

Discover, trois providers :

- Sahel Pay (démo), paiements mobile money
- Baobab SMS (démo), SMS et USSD
- Nimba Maps (démo), adressage et géodonnées

Discover, quatre resources :

- Sahel Pay Mobile Money (démo), API mobile money
- SDK SMS Baobab (démo), SDK SMS et USSD
- Nimba Maps Studio (démo), outil de cartographie
- Sahel Pay Devkit (démo), outil pour développeurs

Build, deux projets :

- Suivi des prix agricoles (démo)
- Adressage des quartiers (démo)

Build, quatre besoins ouverts :

- Brancher une API de prix agricoles
- Rédiger la documentation en français
- Numériser un plan de quartier
- Tester l'application sur mobile

Solve, deux questions optionnelles :

- Un webhook peut-il arriver dans le désordre ?
- Faut-il enregistrer le corps brut avant de vérifier la signature ?

## 10. À ne pas faire

- Ne pas modifier les migrations, `mk/50-ops.mk`, `TODO.md`, `api/openapi.yaml` ou le web depuis ce contrat.
- Ne pas créer de fichier `.sql` dans ce slice T1.
- Ne pas changer un UUID après le merge.
- Ne pas ajouter de credentials, de votes, de candidatures, de vérifications, de sessions ou de réponse acceptée.
- Ne pas utiliser `now()`, `CURRENT_TIMESTAMP` ou `uuidv7()`.
- Ne pas utiliser de vrais emails, de vraies entreprises, de vrais dépôts GitHub ou de vraies URLs.
- Ne pas copier du texte externe.
- Ne pas écrire `BEGIN` ou `COMMIT` dans les fichiers de seed.
- Ne pas lancer les seeds hors `APP_ENV=local-dev`.
- Ne pas contourner le hook Git avec `--no-verify`.
- Ne pas ajouter de trailer `Co-authored-by` ni de mention « Generated with ».

## 11. Fichiers produits par les autres agents

Ce README est le seul livrable de T1. Les autres fichiers du slice SEED #75 sont hors de ce contrat :

| Agent | Fichier |
|---|---|
| T2 | `mk/50-ops.mk`, cible `seed` uniquement |
| T3a | `db/seeds/00001_identity.sql` |
| T3b | `db/seeds/00002_learn.sql` |
| T3c | `db/seeds/00003_discover.sql` |
| T3d | `db/seeds/00004_build.sql` |
| T3e | `db/seeds/00005_solve.sql` (optionnel) |
| T4 | `src/api/internal/platform/pgtest/seed_integration_test.go` |
| T5 | `scripts/tests/seed_test.sh` |
| T6 et T8 | preuves `local-dev`, `TODO.md`, `docs/ai-usage/` |

Tout écart avec ce contrat doit être signalé au orchestrateur avant écriture du SQL, et non corrigé silencieusement dans un fichier de seed.

## 12. Lancer et vérifier les seeds

Sur une base `local-dev` démarrée et migrée (`make up-local-dev`), neuve ou réinitialisée (`make db-reset`, réservé à `local-dev`). L'idempotence ne porte que sur les clés primaires : si la base contient déjà un tag, un nom d'utilisateur ou un email qu'utilisent les fixtures (par exemple le tag `api` ou `go`), une contrainte d'unicité échoue et tout le lot est annulé, sans écriture partielle.

```sh
APP_ENV=local-dev make seed   # 107 lignes la première fois
APP_ENV=local-dev make seed   # 0 ligne ensuite : idempotent
```

Toute autre valeur de `APP_ENV`, ou un `COMPOSE_PROFILES` différent de `local-dev`, est refusée avec le code 2 avant le moindre appel à Docker.

Les preuves automatiques :

| Commande | Ce qu'elle vérifie |
|---|---|
| `scripts/tests/seed_test.sh` | les refus de `make seed` et l'appel `psql -1` en `local-dev`, sans Docker |
| `go test -race -tags integration -run Seed ./src/api/internal/platform/pgtest/` | le contenu de ce contrat sur une base PostgreSQL neuve : nombres de lignes, provenance, absence de credentials, de votes, de candidatures et de vérifications, URLs `example.org`, idempotence |

Une ligne ajoutée ou modifiée dans un fichier de seed doit d'abord l'être dans ce contrat, puis dans `seed_integration_test.go` si un nombre change.
