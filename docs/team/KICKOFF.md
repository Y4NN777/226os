# 226os : page d'accueil de l'équipe et réunion de lancement

> **Lis ceci en premier.** Ce fichier est notre point d'entrée : le contexte en dix lignes, où trouver chaque document, ta mission, les règles et qui débloque quoi. Il sert aussi de script à `Y4NN777` pour la réunion de lancement de 30 minutes ([section 9](#9-script-de-la-réunion-30-minutes)).
>
> **Maintenu par :** `Y4NN777`. **Dernière mise à jour :** jeudi 2026-11-12. Toutes les heures sont en GMT.
>
> Ce fichier ne remplace rien : le plan reste dans [ROADMAP.md](../../ROADMAP.md), le suivi vivant dans [TODO.md](../../TODO.md). En cas de désaccord, ce sont eux qui font foi, et on corrige ce fichier.

**Sommaire :** [1. L'essentiel](#1-lessentiel-en-dix-lignes) | [2. Décisions prises](#2-les-décisions-déjà-prises-sans-débat) | [3. Carte des documents](#3-carte-des-documents) | [4. Glossaire](#4-glossaire) | [5. Une page par personne](#5-une-page-par-personne) | [6. Règles](#6-règles-de-travail) | [7. Calendrier et CADEV](#7-calendrier-et-règlement-cadev) | [8. Blocages](#8-blocages-connus-et-points-ouverts) | [9. Script](#9-script-de-la-réunion-30-minutes) | [10. FAQ](#10-faq)

## 1. L'essentiel en dix lignes

1. **226os** relie l'entraide technique, les API et outils utiles aux développeurs africains, les projets open source, l'apprentissage et les personnes ([README](../../README.md), [PRD](../PRD.md)).
2. **Les neuf domaines restent tous** dans le produit. Rien n'est retiré ni rétrogradé.
3. **Apprendre** (Learn & Grow) et **Connect** sont mis en avant, mais seulement dans la navigation, sur l'accueil et dans les textes (amendement 3.2 du PRD, proposé).
4. **Positionnement public : « Afrique de l'Ouest d'abord, puis toute l'Afrique ».** 226os se lit « 226 Open Stack » ; 226 est l'indicatif téléphonique du Burkina Faso. C'est l'origine du nom, pas un positionnement.
5. **Ce qui existe :** le bootstrap B0 est fusionné (PR #24). Une API Go qui ne sert que `GET /healthz`, les migrations `00001` à `00010`, le code généré depuis le contrat (pas encore monté), un squelette web React et la CI ([ROADMAP section 2](../../ROADMAP.md#2-starting-point-2026-11-12)).
6. **La CI lance `make quality` (G1 à G4) sur chaque PR**, et elle est verte d'après [ROADMAP section 2](../../ROADMAP.md#2-starting-point-2026-11-12).
7. **Rien n'est déployé.** Le domaine `226os.tech` est à nous, mais il n'affiche que la page d'attente de l'hébergeur.
8. **Ce qui reste à construire :** les 43 opérations P0 du contrat ([api/openapi.yaml](../../api/openapi.yaml)) et toutes les pages React. Avec l'annuaire E5, on passera à 44.
9. **G5 (e2e), G6 (performance) et G7 (livraison) sont NOT RUN.** `make test-e2e`, `make perf` et `make seed` répondent `NOT IMPLEMENTED`.
10. **Objectif :** soumettre `https://226os.tech` dimanche 15 vers 12:00, avant la limite de 17:00. Ensuite, **chaque membre vote**, sinon l'équipe est exclue.

## 2. Les décisions déjà prises, sans débat

On ne rouvre pas ces points en réunion. Pour en changer un, on passe par le propriétaire (et par un nouvel ADR si c'est une décision d'architecture). Sources : [journal des décisions](../../ROADMAP.md#14-decisions-log), [section 10 de ROADMAP](../../ROADMAP.md#10-blockers-contradictions-open-decisions), [amendement 3.2 du PRD](../PRD.md#amendement-32--apprendre-et-connect-mis-en-avant-assistant-sourcé-au-p0).

1. **Neuf domaines conservés**, Apprendre et Connect mis en avant dans la navigation, l'accueil et les textes uniquement (ROADMAP O7, PRD 3.2 point 1).
2. **L'assistant IA sourcé (S9) est P0, mais désactivé par configuration** tant que le propriétaire n'a pas tranché V3 (offre gratuite ou projet facturé). Fournisseur : Google AI Studio, modèle Gemma 4, derrière un port interchangeable (journal, TODO S9).
3. **Contenu original uniquement.** Aucun contenu tiers importé ni aspiré ; les sources externes ne sont que des références et des liens sur les fiches Discover (journal, « No third-party content »).
4. **Données de démonstration en `local-dev` seulement** : comptes `demo-` qui ne peuvent pas se connecter, `provenance='demo'`, badge visible (correction B du propriétaire, C15).
5. **`staging-test` et `production-live` : uniquement du vrai contenu**, créé via l'application par nos vrais comptes. Zéro ligne `demo` sur les deux (correction B).
6. **URL de soumission : `https://226os.tech`**, c'est-à-dire `production-live` hébergé sur Datacloud (O1, correction C).
7. **La série gratuite arrive en production par un script de contenu** revu en PR, écrit par `mibienpanjoe`, lancé par `Y4NN777` après une sauvegarde (B5, H19).
8. **Gate d'autorisation : trois listes de routes** (B1). Toute route montée est dans `scripts/public-routes.txt`, ou dans `scripts/unimplemented-routes.txt` (501, la liste rétrécit à chaque tranche), ou dans `scripts/authz-tests.txt` (route associée à un test négatif qui doit exister).
9. **C1 :** les actions sur un projet appartiennent au **propriétaire du projet seul** (pas de mainteneur délégué au P0) ; le candidat peut retirer sa candidature.
10. **C2 :** le filtre de thème Learn utilise le **code de thème**. PR de contrat d'abord, avant les handlers S4 et les fils E4.
11. **C14a :** un profil n'apparaît dans l'annuaire que si `open_to_collab` vaut vrai **et** que l'appelant a le droit de voir ce drapeau. Les filtres ne portent que sur des champs visibles.
12. **C14b :** nouvelle colonne `profiles.profile_visibility` (`public`, `members`, `private`). Un profil masqué renvoie exactement la même 404 qu'un profil inconnu ; la signature d'auteur reste visible.
13. **C19 (et C11) :** `00011` n'est plus réservé. On prend le numéro que donne `make migrate-new`.
14. **Utilisateurs suspendus : hors périmètre P0.**
15. **Région Afrique de l'Ouest = constante d'interface dans `src/web`**, sans migration ni colonne : `BJ, BF, CV, CI, GM, GH, GN, GW, LR, ML, MR, NE, NG, SN, SL, TG` (16 codes du géoschéma ONU, **à vérifier sur la liste M49**, ajustables par le propriétaire). Filtre par défaut « Afrique de l'Ouest », « Toute l'Afrique » en un clic.
16. **ADR-0008 accepté** : trois zones de travail, binômes de relecture, aides nommées ([ADR-0008](../decisions/0008-team-work-areas-and-ownership.md)).
17. **S10 (salons en direct) est conditionnel** au point d'étape du samedi 14 à 12:00. Sans feu vert, il reste P2 et on le dit.
18. Également décidé : e-mail via Gmail SMTP en staging et production, Mailpit en local ; Idempotency-Key sur six opérations ; la modération est construite en dernier, sans modérateur pré-créé (journal).

## 3. Carte des documents

| Je cherche… | Je vais voir… |
|---|---|
| Le produit et le pourquoi | [README](../../README.md), [PRD](../PRD.md) (neuf domaines, amendements 3.1 et 3.2), [ce qu'on ne promet pas](../PRD.md#non-objectifs-de-cadev) |
| Les comportements attendus (scénarios) | [SPEC](../SPEC.md) : US-SOLVE-001, US-DISCOVER-001, US-BUILD-001, US-LEARN-001, US-IDENTITY-001, US-ACCESS-001 ; proposées par la PR #29 : US-LEARN-002 (E4), US-CONNECT-001 (E5), US-INTEL-001 (S9) |
| Ce que la démo doit montrer dimanche | [ROADMAP section 3](../../ROADMAP.md#3-what-must-be-true-on-sunday) |
| Le contrat de l'API | [api/openapi.yaml](../../api/openapi.yaml) (`x-priority`, `x-authz`), [API_CONTRACT](../spec-kit/API_CONTRACT.md), [ADR-0004](../decisions/0004-api-contract-conventions.md) ; types web générés : [schema.d.ts](../../src/web/src/api/schema.d.ts) |
| Le modèle de données | [DATABASE_DESIGN](../spec-kit/DATABASE_DESIGN.md), [db/migrations](../../db/migrations/), [ADR-0005](../decisions/0005-data-model-migrations-search.md) |
| Qui a le droit de faire quoi | [AUTHORIZATION_MATRIX](../spec-kit/AUTHORIZATION_MATRIX.md) ([la matrice](../spec-kit/AUTHORIZATION_MATRIX.md#matrix), [modération](../spec-kit/AUTHORIZATION_MATRIX.md#moderator-scope-at-p0)), [ADR-0003](../decisions/0003-authentication-sessions-csrf-rate-limits.md) |
| Les contrats d'implémentation (API, données, recherche, front, sécurité, ops) | [CONTRACTS](../spec-kit/CONTRACTS.md), [TRACEABILITY](../spec-kit/TRACEABILITY.md), [Spec Kit](../spec-kit/README.md) |
| L'architecture | [ARCHITECTURE](../architecture/ARCHITECTURE.md), [registre des ADR](../decisions/README.md) |
| Le design et la charte | [DESIGN.md](../../DESIGN.md) (Graphite & Lime, Astryx, accessibilité), [ADR-0007](../decisions/0007-astryx-version-pinning-and-verification.md) |
| Les règles de test et de couverture | [TEST](../TEST.md) (section 6 pour les gates), [COVERAGE_SCOPE](../quality/COVERAGE_SCOPE.md), [RULES](../RULES.md), [DEV](../DEV.md) |
| Le workflow Git et les noms de branche | [GIT_WORKFLOW](../GIT_WORKFLOW.md#work-branch-naming), [règles de commit et de PR](../GIT_WORKFLOW.md#commit-and-pr-rules), [CONTRIBUTING](../../CONTRIBUTING.md), [CONTRIBUTORS](CONTRIBUTORS.md) |
| Comment lancer en local | [README, « Lancer le projet en local »](../../README.md#lancer-le-projet-en-local), [Makefile](../../Makefile) et `mk/*.mk` (voir ci-dessous) |
| Le plan | [ROADMAP](../../ROADMAP.md) : [tranches et personnes](../../ROADMAP.md#4-people-and-slices), [calendrier](../../ROADMAP.md#6-timeline-gmt), [chemin critique](../../ROADMAP.md#7-dependencies-and-critical-path) |
| Le suivi vivant | [TODO](../../TODO.md) : [Now](../../TODO.md#now), une section par tranche, [checklist de déploiement](../../TODO.md#deployment-checklist-owner-y4nn777-target-sha-recorded-on-each-line), [Done log](../../TODO.md#done-log) |
| Les risques | [Registre des risques](../../ROADMAP.md#11-risk-register), [ce qui saute en premier](../../ROADMAP.md#121-the-plan-is-bound-by-capacity) |
| Les références et l'étude 225 Open Source | [REFERENCES](../research/REFERENCES.md), [étude 225 Open Source](../research/REFERENCES.md#étude--225-open-source-référence-dinspiration) |
| Les licences et la déclaration de l'IA | [LICENSE](../../LICENSE) (Apache-2.0), [THIRD_PARTY_NOTICES](../../THIRD_PARTY_NOTICES.md), [AI_USAGE](../../AI_USAGE.md#required-record-for-each-material-ai-assisted-contribution) |
| Les prototypes et leur avertissement | [prototypes/html/DISCLAIMER.md](../../prototypes/html/DISCLAIMER.md) (à lire avant d'ouvrir une page), [prototypes/html](../../prototypes/html/README.md) |
| La mise en production et les règles du concours | [SHIP](../SHIP.md) ([section 6, CADEV](../SHIP.md#6-cadev--audit-de-soumission)), [TODO S8](../../TODO.md#s8--ops-release-and-cadev), [obligations CADEV](../../ROADMAP.md#9-cadev-obligations-docsshipmd-section-6) |
| Les règles des agents IA | [AGENTS.md](../../AGENTS.md), [CLAUDE.md](../../CLAUDE.md) |

**Les ADR, une ligne chacun** ([registre](../decisions/README.md)) :

| ADR | Ce qu'il fixe |
|---|---|
| [0001](../decisions/0001-g0-conditional-baseline.md) | Socle G0 : hébergement Datacloud, licence Apache-2.0, organisation des docs, gouvernance |
| [0002](../decisions/0002-interim-single-maintainer-ownership.md) | Propriété intérimaire par un seul mainteneur (amendé par 0008) |
| [0003](../decisions/0003-authentication-sessions-csrf-rate-limits.md) | Authentification, sessions, CSRF, limites de débit |
| [0004](../decisions/0004-api-contract-conventions.md) | Conventions du contrat d'API : erreurs problem+json, curseurs, idempotence |
| [0005](../decisions/0005-data-model-migrations-search.md) | Modèle de données, migrations, recherche PostgreSQL, seeds |
| [0006](../decisions/0006-toolchain-environments-quality-commands.md) | Outillage, environnements `local-dev` / `staging-test` / `production-live`, commandes Make ; Valkey et S3 exclus au P0 |
| [0007](../decisions/0007-astryx-version-pinning-and-verification.md) | Épinglage et vérification de la version d'Astryx |
| [0008](../decisions/0008-team-work-areas-and-ownership.md) | Zones de travail de l'équipe à quatre, relecteurs, fichiers partagés |

**Cibles Make réelles** ([mk/10-env.mk](../../mk/10-env.mk), [mk/40-quality.mk](../../mk/40-quality.mk), [mk/50-ops.mk](../../mk/50-ops.mk)). Lis une recette avec `make -n <cible>` avant de la lancer la première fois.

| Cible | Rôle | État |
|---|---|---|
| `make doctor` | Vérifie Go, Node, pnpm, Docker et Compose, sans rien modifier | disponible |
| `make setup` | `go mod download` et `pnpm install --frozen-lockfile` | disponible |
| `make quality` | G1 à G4 : `doctor generate lint typecheck test-unit coverage test-integration security licenses` | disponible (Docker requis) |
| `make up-local-dev` / `make down-local-dev` | Démarre ou arrête le profil local (lit `.env.local-dev`, créé à partir de `.env.example`, jamais commité) | disponible |
| `make migrate-new NAME=<slug>` | Crée la migration suivante (réservé à `mibienpanjoe`) | disponible |
| `make test-e2e`, `make perf`, `make seed` | G5, G6, données de démo | **NOT IMPLEMENTED** |
| `make quality-full` | `quality` plus `test-e2e` et `perf` | échoue tant que G5 et G6 ne sont pas implémentés |

## 4. Glossaire

**Priorités, gates et outils**

| Code | Sens |
|---|---|
| P0 / P1 / P2 | Ordre de livraison. P0 = ce qu'on livre pour CADEV ; P1 et P2 = plus tard, jamais présentés comme faits |
| G0 | Périmètre : story, critères d'acceptation, risques ([TEST section 6](../TEST.md), [AGENTS](../../AGENTS.md#deterministic-gate-protocol)) |
| G1 | Format, lint, types, génération reproductible (OpenAPI, sqlc) |
| G2 | Tests unitaires et couverture : Go à 100 %, web au-dessus de 70 % sur les quatre métriques |
| G3 | Intégration sur un vrai PostgreSQL, conformité au contrat, permissions |
| G4 | Sécurité : scans, tests négatifs, aucun secret, licences |
| G5 | E2E : parcours P0, FR/EN, mobile, accessibilité. **NOT RUN** |
| G6 | Performance : budgets et absence de régression. **NOT RUN** |
| G7 | Livraison : SHA, empreinte d'image, smoke, sauvegarde et rollback prouvés. **NOT RUN** |
| ADR | Architecture Decision Record, dans [docs/decisions](../decisions/README.md). Une décision acceptée ne change que par un nouvel ADR |
| CI, `make quality` | GitHub Actions lance `make quality` (G1 à G4) sur chaque PR. Vert obligatoire avant fusion |
| DoD-common | Définition de « fini » commune à toutes les tranches ([TODO](../../TODO.md), en tête de fichier) |
| US-XXX-NNN/n | Story de la SPEC ; `/2` = deuxième scénario de la story, dans l'ordre |
| `live` / `demo` | Provenance d'un contenu : réel, ou fixture de démo (`local-dev` uniquement) |

**Tranches et extensions** ([ROADMAP section 4](../../ROADMAP.md#4-people-and-slices))

| Code | Sens | Code | Sens |
|---|---|---|---|
| S0 | Cœur de l'API (pool DB, `/readyz`, montage, validation, curseurs, idempotence, mailer) | S6 | Modération, construite en dernier |
| S1 | Identité : auth, `me`, profils (11 opérations) | S7 | Front : les prototypes deviennent des pages React |
| S2 | Discover et recherche (5) | S8 | Ops, livraison, soumission CADEV |
| S3 | Solve (10) | S9 | Assistant IA sourcé (P0 par décision du propriétaire) |
| S4 | Learn, avec le script de contenu (5) | S10 | Salons en direct, **conditionnel** |
| S5 | Build (9) | SEED | Fixtures de démo et `make seed`, `local-dev` seulement |
| E1 | Règles pédagogiques de S9 (AI-1 à AI-5, AI-9, AI-10) devenues tests ; AI-6 et AI-7 sont P1 | E2 | Taxonomie : ajouts et renommages, une migration |
| E4 | Bouton « Poser une question sur cette leçon » et fils par thème | E5 | Connect : annuaire des membres (nouvelle opération) |
| IA (architecture de l'information) | Mise en avant des piliers : navigation, accueil, textes (`PANK4SS`) | E3, S11, S12 | Remplacés, plus d'actualité |

**Blocages (B) et contradictions (C)** ([ROADMAP section 10](../../ROADMAP.md#10-blockers-contradictions-open-decisions))

| Code | Sens | Code | Sens |
|---|---|---|---|
| B1 | Gate d'autorisation : DÉCIDÉ, trois listes de routes | C8 | Domaine de branche `learn` absent de GIT_WORKFLOW : ouvert |
| B2 | `/readyz` jamais routé vers l'extérieur : DÉCIDÉ (la sonde se lance dans le conteneur) | C9 | Libellés `x-authz` périmés : ouvert |
| B3 | Timeout de 30 s du routeur : DÉCIDÉ (30 s par défaut, une table d'exceptions par route quand S9 ou S10 en a besoin) | C10 | Priorité de S9 (P2 dans le PRD, P0 décidé) : amendement 3.2 |
| B4 | Aucun chemin de déploiement en production : planifié (S8) | C11 | `00011` réservé : DÉCIDÉ avec C19 |
| B5 | Série gratuite en production : DÉCIDÉ, script de contenu | C12 | Qui peut créer un lien article vers question : ouvert |
| C1 | Droits de mainteneur : DÉCIDÉ, propriétaire seul | C13 | `LinkRef` accepte `series`, la base non : ouvert |
| C2 | `topic` de Learn : DÉCIDÉ, code de thème | C14 | Profil privé sans colonne : DÉCIDÉ (a et b) |
| C3 | `my_progress` sans table de progression : ouvert | C15 | Auteurs des fixtures : DÉCIDÉ, comptes `demo-` |
| C4 | Longueur de `q` : 200 dans le contrat, 100 dans ADR-0005 : ouvert | C16 à C18 | Retirés |
| C5 | Codes d'erreur manquants (`METHOD_NOT_ALLOWED`, `Content-Type`) : ouvert | C19 | `00011` plus réservé : DÉCIDÉ |
| C6 | Docs périmées sur la propriété (trois membres) : ouvert | | |
| C7 | DATABASE_DESIGN dit que `db/` n'existe pas : ouvert | | |

**Tâches humaines (H)** ([ROADMAP section 8](../../ROADMAP.md#8-human-tasks-and-lead-time))

| Code | Sens | Code | Sens |
|---|---|---|---|
| H1 | Identités Git confirmées : FAIT | H10 | Rappels de vote et responsable de suivi |
| H2 | `PANK4SS` accepte l'invitation : en attente | H11 | URL de démo : DÉCIDÉ (production) |
| H3 | ADR-0008 accepté : FAIT | H12 | Clé AI Studio et vérifications V1 à V4 |
| H4 | Protection des branches `develop` et `main` | H13 | Appliquer l'amendement 3.2 et la story SPEC de S9 |
| H5 | Décisions encore ouvertes (liste dans ROADMAP section 8) | H14 | Retiré (H15 et H17 n'existent pas) |
| H6 | Compte et produit actif Datacloud | H16 | Amendement des piliers (O7), approuvé sur la PR docs |
| H7 | DNS de `226os.tech` et du sous-domaine de staging | H18 | O1 et responsable de SEED : clos |
| H8 | Mot de passe d'application Gmail SMTP | H19 | Lancer le script de contenu en production |
| H9 | Inscription CADEV : FAIT | | |

**Points ouverts (O), vérifications (V), risques (K)** ([décisions ouvertes](../../ROADMAP.md#open-decisions), [risques](../../ROADMAP.md#11-risk-register))

| Code | Sens | Code | Sens |
|---|---|---|---|
| O1 | URL de démo : DÉCIDÉ, `https://226os.tech` | O9 | Backend E5 : `NICE-DEV226` |
| O2 | Où vivent les pièces partagées : décidé par le lead au lancement | O10 | SEED : `PANK4SS` |
| O3 | Les 4 routes P1 : CLOS par B1 (elles répondent 501) | V1 | Identifiant exact du modèle Gemma 4 |
| O4 | Renvoi de l'e-mail de vérification : ouvert | V2 | Limites réelles du projet AI Studio |
| O5 | Questions de l'ADR S9 (données envoyées, quota, accès anonyme) | V3 | Projet gratuit ou facturé, vu la règle EEE |
| O6 | Qui construit S10 et ce que ça déplace | V4 | Latence et qualité FR/EN sur le jeu de référence |
| O7 | Piliers mis en avant : amendement 3.2 proposé | K1 à K39 | Risques ; K30 à K35 retirés. À surveiller : K1 capacité, K2 goulot de relecture, K21 règle EEE, K38 fenêtre de contenu courte, K39 pas d'écriture Discover |
| O8 | Retiré | R1 à R11 | Règles de chaque tranche, voir la [section 6](#6-règles-de-travail) |

**Les neuf domaines** ([PRD section 5](../PRD.md))

| Domaine | En une ligne |
|---|---|
| Solve | Questions, réponses, snippets, votes, réponse acceptée |
| Discover | Catalogue d'API, SDK et outils, par pays et catégorie, avec sources |
| Build | Projets open source, besoins de contribution, candidatures |
| Learn (Apprendre) | Articles, tutoriels, série gratuite lisible sans compte ; **mis en avant** |
| Connect | Annuaire des membres au P0 ; mentorat et carrière « bientôt » ; **mis en avant** |
| Identity & Community | Comptes, profils, visibilité du profil, vie communautaire |
| Language | Interface FR/EN ; le contenu garde sa langue d'origine |
| Access | Mobile d'abord, réseau lent, états de chargement clairs |
| Intelligence | Recherche, liens contextuels, assistant sourcé (P0 par l'amendement 3.2) |

## 5. Une page par personne

Les relecteurs viennent de [ROADMAP section 4](../../ROADMAP.md#4-people-and-slices) et de [CONTRIBUTORS](CONTRIBUTORS.md). Ta ligne dans [TODO, Now](../../TODO.md#now) est la seule que tu modifies.

### `Y4NN777`

- **Mission :** propriétaire, lead backend et ops. Tu poses le cœur de l'API dont tout le monde dépend, tu ouvres les comptes externes et tu mènes la livraison jusqu'à la soumission.
- **Tranches :** S0, S6, S8, S9 (avec `NICE-DEV226` en second développeur). Aide sur S2, S4 et le script de contenu (que tu lances), S5, E2. Relis le front (S7, E4, SEED pour `mk/`) et, avec `PANK4SS`, chaque PR de contrat.
- **Les prochaines 24 heures, dans l'ordre :**
  1. H4 : protection de `develop` et `main`, check `quality` obligatoire.
  2. H6 et H7 : compte et produit Datacloud, puis enregistrements A dès que l'IP est connue.
  3. S0 sur `feature/platform/authz-gate` : le gate B1 en RED puis GREEN.
  4. S0 sur `feature/platform/api-mount` : `pgxpool`, `DATABASE_URL`, `/readyz`, montage `HandlerWithOptions`. PR vers `NICE-DEV226`.
  5. Relire la première PR de contrat de `mibienpanjoe` (avec `PANK4SS`).
  6. Vendredi matin : fusionner S0, puis Idempotency-Key, mailer et Mailpit ; H8 (Gmail).
  7. Vendredi après-midi : staging avec TLS, DNS et SMTP ; H12 et V1 à V4 avant le soir ; benchmark Argon2id sur l'hôte, transmis à S1 ; brouillon de l'ADR S9.
- **Premier test RED :** une route parasite, un nom de test absent et une route non listée font chacun passer le gate B1 au rouge ([TODO S0](../../TODO.md#s0--api-core)).
- **Chemins autorisés (S0) :** `src/api/cmd/**`, `src/api/internal/app/**`, `src/api/internal/platform/**` (apigen généré seulement), `scripts/**`, `compose.yaml`, `deploy/**`, `.env.example`, `mk/**` (un changement par PR).
- **Qui t'aide, qui te relit :** `NICE-DEV226` relit S0 ; `mibienpanjoe` relit S6 et S9 ; S8 est relu par `NICE-DEV226` ou `mibienpanjoe`. `PANK4SS` aide à la demande.
- **Sections TODO :** [S0](../../TODO.md#s0--api-core), [S6](../../TODO.md#s6--moderation-built-last), [S8](../../TODO.md#s8--ops-release-and-cadev), [S9](../../TODO.md#s9--grounded-ai-assistant-p0-by-owner-decision-amends-prd-58), [pièces partagées](../../TODO.md#shared-pieces-o2-where-each-lives-is-decided-by-the-backend-lead-at-kickoff), [checklist de déploiement](../../TODO.md#deployment-checklist-owner-y4nn777-target-sha-recorded-on-each-line).
- **Points ouverts qui te concernent :** O2, B4, O4, O5 et V1 à V4, O6 (S10), K39 (contenu Discover en production), interface de modération (S6), question GPL des images de base, approbation de l'amendement 3.2, sous-domaine de staging à confirmer.
- **Fini quand :** `make quality` vert en CI sur le SHA de la PR (DoD-common), plus `make up-local-dev` avec `/readyz` à 200 et un mail de test reçu dans Mailpit (preuve dans la PR).
- **Ce qui te bloque :** le délai Datacloud (K5) et la relecture de S0 par `NICE-DEV226`. Si B1 ne peut pas se résoudre sans affaiblir le gate, tu t'arrêtes et tu tranches en propriétaire.

### `NICE-DEV226`

- **Mission :** développeur backend, responsable de l'identité puis de Solve. Tout ce qui demande une session passe par ton code, donc S1 est sur le chemin critique.
- **Tranches :** S1, puis S3, puis le backend E5 si le temps le permet. Second développeur de S9 après Solve. Aide sur S2, S4, S5 et E2 (par défaut S5 après Solve, décidé au point du vendredi 23:00). Relis S0 et les PR de `mibienpanjoe`.
- **Les prochaines 24 heures, dans l'ordre :**
  1. Relire S0 dès son ouverture : tout le monde l'attend.
  2. S1 sur `feature/identity/auth-sessions` : tests RED du domaine et du service avec des fakes (hachage, jetons, CSRF, sessions). **Aucune route montée avant la fusion de S0.**
  3. Relire la PR de migrations E2 et `profile_visibility` de `mibienpanjoe`.
  4. Vendredi matin : les endpoints d'auth et le limiteur de débit PostgreSQL.
  5. Vendredi après-midi : `me` et profils sur `feature/identity/profiles`, puis fusion de S1 avant la fin de journée.
- **Premier test RED :** à la connexion, la session tourne et l'ancien jeton est refusé ; une session révoquée ou expirée est inutilisable (ADR-0003, « Evidence », repris dans [TODO S1](../../TODO.md#s1--identity)).
- **Chemins autorisés (S1) :** `src/api/internal/identity/**`, `db/queries/identity.sql`, `db/queries/platform.sql`, `src/api/internal/platform/dbgen/**` (généré par `make generate` seulement), `scripts/public-routes.txt` (routes d'auth anonymes, avec leur raison).
- **Qui t'aide, qui te relit :** `mibienpanjoe` relit S1 et S3 ; le lead peut co-relire la surface d'auth. `PANK4SS` aide à la demande.
- **Sections TODO :** [S1](../../TODO.md#s1--identity), [S3](../../TODO.md#s3--solve), [E5](../../TODO.md#e5--connect-people-directory), [S9](../../TODO.md#s9--grounded-ai-assistant-p0-by-owner-decision-amends-prd-58).
- **Points ouverts qui te concernent :** O4 (pas d'opération pour renvoyer l'e-mail de vérification), K6 (Argon2id à 64 Mio contre la limite de 128M du conteneur), C14b (tests `getProfile` par visibilité et par appelant).
- **Fini quand :** `make quality` vert en CI sur le SHA de la PR (DoD-common), plus l'inspection des cookies par profil et un journal du parcours e-mail via Mailpit, sans données sensibles.
- **Ce qui te bloque :** la fusion de S0 (`Y4NN777`) ; le benchmark Argon2id sur l'hôte avant d'épingler les paramètres (`Y4NN777`, S8) ; la migration `profile_visibility` et la première PR de contrat avant la visibilité des profils (`mibienpanjoe`). Si une valeur d'ADR-0003 doit changer, tu t'arrêtes.

### `mibienpanjoe`

- **Mission :** développeur backend et **seul auteur des migrations**. Tu portes la plus grosse charge (Discover, Learn, Build, environ 19 opérations), avec deux aides nommées.
- **Tranches :** S2, S4 (avec le script de contenu), S5, la migration E2 et la migration `profile_visibility`. Responsable même quand on t'aide. Relis S1, S3, SEED, S6 et S9.
- **Les prochaines 24 heures, dans l'ordre :**
  1. **La première PR de contrat**, relue par `Y4NN777` et `PANK4SS`, puis `make generate`. Elle porte :
     - C2 : `topic` en code de thème dans `listLearnArticles`, ajouté dans `listTrainingSeries`, champ optionnel dans `ArticleCreate` ;
     - C1 : `x-authz` de `decideApplication` en propriétaire seul ;
     - C14b : `profile_visibility` dans les schémas de profil, `deny_status 404` sur `getProfile`.
  2. En parallèle, la PR de migrations sur `feature/platform/taxonomy-ai-topics` : E2 et `profile_visibility`, deux fichiers séparés créés avec `make migrate-new`, les tests pgtest mis à jour, et la correction de la réservation de `00011` dans DATABASE_DESIGN.
  3. Les requêtes sqlc de S2 avec `EXPLAIN` en local.
  4. Relire S1 au fil de l'eau.
  5. Vendredi midi : lectures publiques de S2 fusionnées (`listResources`, `getResource`, `listResourceSources`, `search`), puis lectures de S4.
  6. Vendredi après-midi : lectures S4 fusionnées ; signalement d'erreur S2 et publication S4 dès que S1 est fusionné.
- **Premier test RED :** mettre à jour d'abord `migrate_integration_test.go` (14 thèmes, 18 catégories) et `TestFreshMigrationToHead` (plus de 10 migrations) pour qu'ils échouent, puis écrire les migrations ([TODO E2](../../TODO.md#e2--taxonomy-add-and-relabel-only)).
- **Chemins autorisés :** PR de contrat : `api/openapi.yaml`, puis `src/api/internal/platform/apigen/**` et `src/web/src/api/schema.d.ts` régénérés par `make generate` uniquement. E2 : `db/migrations/<nouveaux fichiers>`, `src/api/internal/platform/pgtest/**` (et la ligne `00011` de `DATABASE_DESIGN.md` que TODO demande dans la même PR). S2 : `src/api/internal/discover/**`, `src/api/internal/search/**`, `db/queries/**` (fichiers discover et search), `dbgen/**` (généré), `scripts/public-routes.txt`. S4 ajoute `src/api/internal/learn/**` et `scripts/content/**`.
- **Qui t'aide, qui te relit :** aides `Y4NN777` et `NICE-DEV226` (jamais sur une migration). Relecteur `NICE-DEV226` ; s'il a co-écrit la PR, c'est `Y4NN777`. `Y4NN777` co-relit les changements pgtest. Les PR de contrat : `Y4NN777` et `PANK4SS`.
- **Sections TODO :** [S2](../../TODO.md#s2--discover-and-search), [S4](../../TODO.md#s4--learn), [S5](../../TODO.md#s5--build), [E2](../../TODO.md#e2--taxonomy-add-and-relabel-only).
- **Points ouverts qui te concernent :** C3, C4, C12, C13, K25 (sens du thème `ai`), K26 (listes pgtest), K39 (aucune écriture Discover en production, à voir avec le propriétaire), suivis de doc DATABASE_DESIGN (C1, C7, C19).
- **Fini quand :** `make quality` vert en CI sur le SHA de la PR (DoD-common) ; pour E2, `make test-integration` passe sur une base neuve et en montée de version, et `make lint` valide l'immuabilité des migrations ; pour S2, les plans `EXPLAIN` sont dans la PR.
- **Ce qui te bloque :** les relectures de la PR de contrat (`Y4NN777`, `PANK4SS`) ; S0 pour monter S2 ; S1 pour le signalement d'erreur et la publication ; la fusion de la PR C2 avant les handlers S4.

### `PANK4SS`

- **Mission :** responsable du front. Tu transformes les prototypes en vraies pages React avec Astryx, en mettant Apprendre et Connect en avant, et tu épaules n'importe qui à la demande.
- **Tranches :** S7 (coque, i18n, navigation des piliers, toutes les pages), E4 (avec `NICE-DEV226`), l'interface E5, SEED, l'interface S9. Co-relis chaque PR de contrat avec `Y4NN777`.
- **Les prochaines 24 heures, dans l'ordre :**
  1. **Accepter l'invitation GitHub (H2) avant tout push.**
  2. Relire la première PR de contrat de `mibienpanjoe` (avec `Y4NN777`).
  3. S7 : coque, i18n, composants partagés et navigation des piliers sur fixtures typées tirées de `schema.d.ts`.
  4. Conception de SEED : comptes de démo selon C15.
  5. Vendredi matin : pages de lecture publiques `accueil`, `discover`, `resource`, `recherche`, avec la région Afrique de l'Ouest par défaut.
  6. Vendredi après-midi : fixtures SEED et `make seed` qui refuse tout sauf `local-dev` ; textes FR/EN des piliers ; pages `apprendre`, `article`, `formation`, `solve`, `question` sur fixtures, puis les pages d'identité ; relecture du contrat S9.
- **Premier test RED :** dans la coque, changer de langue modifie les libellés mais jamais le contenu d'origine (Vitest et Testing Library, par rôle et par libellé). Pour SEED : `make seed` refuse avec `APP_ENV=staging-test` et `production-live` ([TODO S7](../../TODO.md#s7--frontend-prototypes-become-react-pages), [TODO SEED](../../TODO.md#seed--original-demo-fixtures-and-make-seed)).
- **Chemins autorisés :** S7 : `src/web/**` et `DESIGN.md` (`pnpm-lock.yaml` : un changement par PR). SEED : `db/seeds/**` et `mk/50-ops.mk` (cible `seed`). **`prototypes/html/**` est hors de ton périmètre** tant que le propriétaire ne l'a pas confirmé (TODO S7, « Prototype alignment »).
- **Qui t'aide, qui te relit :** `Y4NN777` relit S7, E4 et l'interface E5 ; `mibienpanjoe` relit SEED, et `Y4NN777` aussi pour `mk/50-ops.mk`. Les propriétaires de tranche te répondent sur le branchement des endpoints.
- **Sections TODO :** [S7](../../TODO.md#s7--frontend-prototypes-become-react-pages), [SEED](../../TODO.md#seed--original-demo-fixtures-and-make-seed), [E4](../../TODO.md#e4--zero-migration-interaction-modes), [E5](../../TODO.md#e5--connect-people-directory). Avant d'ouvrir un prototype : [DISCLAIMER](../../prototypes/html/DISCLAIMER.md).
- **Points ouverts qui te concernent :** vérifier les 16 codes de la région sur la liste M49 ; confirmation du périmètre « Prototype alignment » ; interface de modération si le propriétaire la demande ; correction du texte d'aide de `make seed` qui mentionne encore le staging.
- **Fini quand :** `make quality` vert en CI sur le SHA de la PR (DoD-common), avec le web au-dessus de 70 % sur les quatre métriques, tests par rôle et libellé ; pour SEED, preuve de `make seed` en `local-dev` et tests de refus. Les preuves visuelles, axe et responsive d'ADR-0007 restent NOT RUN tant qu'elles ne sont pas produites.
- **Ce qui te bloque :** H2 (toi, `Y4NN777` renvoie l'invitation si besoin) ; chaque endpoint fusionné par son propriétaire de tranche ; la migration E2 avant SEED (`mibienpanjoe`). Si une page a besoin d'un champ absent de `schema.d.ts`, PR de contrat d'abord.

## 6. Règles de travail

Les règles complètes R1 à R11 sont dans [ROADMAP section 5](../../ROADMAP.md#5-rules-for-every-slice-r1-to-r11-referenced-from-todomd). L'essentiel :

| Règle | En clair |
|---|---|
| R1 Test d'abord | Un test qui échoue (RED), le code qui le fait passer (GREEN), puis la refacto sans changer le comportement. Invariants et autorisations testés sur un vrai PostgreSQL |
| R2 Couverture Go | 100,00 % des instructions Go du projet. **Une seule exclusion** : `apigen.gen.go` ([COVERAGE_SCOPE](../quality/COVERAGE_SCOPE.md)). SMTP et LLM derrière des ports avec fakes, pas de réseau en test |
| R3 Couverture web | Strictement au-dessus de 70 % sur instructions, branches, fonctions et lignes |
| R4 CI verte | `make quality` vert dans la CI avant toute fusion. On n'affaiblit jamais un gate, on ne saute pas de test, on ne baisse pas un seuil |
| R5 Autorisations | Une route publique va dans `scripts/public-routes.txt` avec sa raison ; toute autre route a ses tests négatifs tirés de la matrice |
| R6 Migrations | **Seul `mibienpanjoe` en écrit**, une à la fois, en avant uniquement, immuables après fusion |
| R7 Fichiers partagés | `api/openapi.yaml`, `go.mod`/`go.sum`, `pnpm-lock.yaml`, `mk/**` : un seul éditeur à la fois, un changement par PR. Le contrat fusionne d'abord, seul, relu par `Y4NN777` et `PANK4SS` |
| R8 Branches | Une tranche, une branche : `feature/<domaine>/<nom>` (ou `fix/`, `chore/`…) depuis `develop`, PR vers `develop`. **Personne ne relit sa propre PR**, ni une PR qu'il a co-écrite |
| R9 Commits | Ton identité Git confirmée. Aucune attribution IA : pas de `Co-authored-by` d'agent, pas de pied « Generated with ». L'aide d'une IA se déclare dans [AI_USAGE.md](../../AI_USAGE.md) |
| R10 Données et secrets | Aucun secret dans Git, les logs, les fixtures, les prompts ou les captures (SMTP, clé de curseur, clé LLM). Les données de démo sont marquées et ne tournent jamais hors `local-dev` |
| R11 Statut honnête | On ne présente jamais du P1 ou du P2 comme fait. Un contrôle non lancé est `NOT RUN`, jamais `PASS` |

Et trois habitudes d'équipe :

- **Aide :** un aide est nommé comme tel sur la PR, peut coder sur la branche, et le propriétaire de la tranche reste responsable.
- **`TODO.md` se met à jour dans la même PR que le travail**, une ligne par changement, sans réécrire la case d'un autre.
- **Bloqué plus de 30 minutes : tu le dis** sur le canal de l'équipe, avec ce que tu as essayé.

**Comment faire une PR**

- [ ] Branche créée depuis `develop`, nommée selon [GIT_WORKFLOW](../GIT_WORKFLOW.md#work-branch-naming) (`learn` n'y figure pas encore mais la CI l'accepte, C8).
- [ ] Le test RED est écrit et nommé dans la description, puis GREEN.
- [ ] `make quality` passe en local ; la CI est verte sur le SHA de la PR.
- [ ] Couverture Go à 100,00 %, web au-dessus de 70 % si tu y as touché.
- [ ] Nouvelles routes : dans `public-routes.txt` avec leur raison, ou couvertes par des tests négatifs ; sortie de la table des routes de `make security` collée.
- [ ] Code généré jamais modifié à la main ; `make generate` relancé après un rebase si une autre PR a touché les mêmes fichiers.
- [ ] Description : story, relecteur, aides éventuelles, « Commands executed » avec les codes de sortie.
- [ ] `TODO.md` mis à jour dans la même PR.
- [ ] Aucun secret, aucun `Co-authored-by` d'agent, aide IA déclarée dans `AI_USAGE.md`.
- [ ] Le bon relecteur est demandé (voir ta page en section 5).

## 7. Calendrier et règlement CADEV

| Quand (GMT) | Quoi |
|---|---|
| Jeudi 12, fin de journée | S0 fusionné ou en relecture |
| **Vendredi 13, 12:00** | Lectures publiques fusionnées (S2 : catalogue et recherche) ; lectures S4 en relecture |
| Vendredi 13, fin de journée | S1 fusionné ; clé AI Studio en place, V1 à V4 vérifiés |
| **Vendredi 13, 23:00** | Point sur les aides : l'avancement de S4 et S5 décide qui aide qui en premier |
| **Samedi 14, 12:00** | Point d'étape : ce qu'on laisse tomber, feu vert ou non pour S10 |
| **Samedi 14, 18:00** | Gel des fonctionnalités : après, seulement des corrections |
| **Dimanche 15, matin** | Livraison en production, contenu réel, soumission **avant 12:00** |
| Dimanche 15, 17:00 | Limite officielle de soumission |
| **Dimanche 15, 18:00 à lundi 16, 23:59** | **Vote obligatoire pour chaque membre**, sinon l'équipe est exclue |

**Obligations du concours**, suivies dans [TODO S8](../../TODO.md#s8--ops-release-and-cadev) (« CADEV rules checklist ») et [SHIP section 6](../SHIP.md#6-cadev--audit-de-soumission). On ne recopie pas le règlement ici.

- Un produit Datacloud réellement acheté et actif à la soumission (`Y4NN777`).
- Aucun composant GPL, AGPL ou LGPL. Question ouverte sur busybox et les paquets Alpine des images (`Y4NN777`).
- Déclaration de titularité par chaque membre à la soumission.
- Chaque composant et sa licence déclarés : [THIRD_PARTY_NOTICES](../../THIRD_PARTY_NOTICES.md) n'a pas encore d'inventaire.
- Usage de l'IA déclaré dans [AI_USAGE.md](../../AI_USAGE.md), chacun pour son propre travail.
- Revue de licence des visuels de marque générés par IA (`Y4NN777`).
- Une présentation du code avec les quatre membres avant la soumission : la qualité technique compte pour 30 % et inclut notre capacité à expliquer le code.

## 8. Blocages connus et points ouverts

**Qui débloque quoi**

| Blocage | État | Qui débloque |
|---|---|---|
| Identités Git (H1) | Confirmées le 2026-11-12 | fait |
| Invitation GitHub de `PANK4SS` (H2) | En attente | `PANK4SS` accepte ; `Y4NN777` renvoie si besoin |
| Protection des branches (H4) | À faire | `Y4NN777` |
| S1 ne monte aucune route avant S0 | S0 en cours | `Y4NN777`, relu par `NICE-DEV226` |
| Handlers Learn (S4) et fils E4 attendent la PR de contrat C2 | PR à ouvrir | `mibienpanjoe`, relu par `Y4NN777` et `PANK4SS` |
| Visibilité des profils (S1, E5) attend la migration `profile_visibility` | À écrire | `mibienpanjoe` |
| Comptes du propriétaire : Datacloud (H6), DNS (H7), mot de passe d'application Gmail (H8), clé AI Studio (H12) | À faire | `Y4NN777` |
| `/readyz` externe (B2), timeout du routeur (B3) | Décidés le 2026-11-12 (voir le journal des décisions) | fait |
| Workflow de livraison (B4) | Planifié dans S8, à fermer avant dimanche | `Y4NN777` |

**Décisions encore ouvertes côté propriétaire** ([liste complète, H5](../../ROADMAP.md#8-human-tasks-and-lead-time)) :

- **V3 :** projet AI Studio gratuit ou lié à la facturation (règle EEE, K21). Tant que ce n'est pas tranché, l'assistant reste désactivé.
- **Images de base et GPL :** demander à l'organisateur si les paquets OS comptent, produire un rapport de licences des images, envisager de retirer busybox de l'image API ([TODO S8](../../TODO.md#s8--ops-release-and-cadev)).
- **Approbation de l'amendement 3.2 du PRD** sur la PR #29 (H16, O7), puis H13 (story SPEC de S9).
- Restent aussi ouverts : C3 à C10, C12, C13, O4 à O6, et K39 (comment remplir Discover en production).

## 9. Script de la réunion (30 minutes)

Animation : `Y4NN777`. But : à la fin, chacun sait **quoi faire en premier**, **qui l'aide** et **qui le relit**. On ne rediscute pas la section 2.

**Étape 1 : où on en est (3 min)**

- « Le socle est fusionné : API de départ, base de données, application web de départ, et la CI qui lance `make quality` sur chaque PR. »
- « Aucune fonctionnalité n'est construite : les 43 opérations P0 et les pages, c'est notre travail à partir de ce soir. »
- « Rien n'est en ligne. `226os.tech` est à nous, mais il n'y a encore rien dessus. »

**Étape 2 : ce qui est décidé (4 min)**

- « Tout ce qui est en section 2 de KICKOFF est décidé. Je lis les titres, on ne débat pas. Si quelqu'un voit un vrai problème, on en parle après, à deux. »
- Lire les points 1 à 8 de la [section 2](#2-les-décisions-déjà-prises-sans-débat), puis : « Les autres sont dans la liste, lisez-les ce soir. »

**Étape 3 : la première tâche de chacun (10 min)**

- « Chacun ouvre sa page en section 5 et lit à voix haute ses deux premières tâches, son premier test RED et son relecteur. »
- Ordre : `Y4NN777`, `NICE-DEV226`, `mibienpanjoe`, `PANK4SS`.
- `Y4NN777` tranche O2, que TODO confie au lead pour ce lancement : « Les pièces partagées (limiteur de débit, résolveur de session, liens, journal d'audit) : je confirme les propositions de TODO.md, ou je dis maintenant lesquelles je réattribue. » Les propositions sont dans la section [pièces partagées](../../TODO.md#shared-pieces-o2-where-each-lives-is-decided-by-the-backend-lead-at-kickoff).

**Étape 4 : les cinq règles (5 min)**

- « Un : le test d'abord, puis le code. »
- « Deux : la CI doit être verte avant de fusionner, et on ne touche jamais à un seuil. »
- « Trois : personne ne relit sa propre PR. »
- « Quatre : une tranche, une branche, des petites PR, et `TODO.md` mis à jour dans la même PR. »
- « Cinq : bloqué plus de 30 minutes, on le dit tout de suite. »

**Étape 5 : les dates (4 min)**

- « Vendredi midi, les lectures publiques sont fusionnées. Vendredi 23 h, on décide qui aide qui. »
- « Samedi midi, point d'étape : on décide ce qui saute et si on fait les salons en direct. Samedi 18 h, gel. »
- « Dimanche matin, mise en ligne et soumission avant midi. Et de dimanche 18 h à lundi 23 h 59, chacun vote, sinon on est exclus. »

**Étape 6 : trois choses à décider ensemble (4 min)**

1. « Quel canal pour dire qu'on est bloqué ? »
2. « À quelle heure le point quotidien de 10 minutes ? »
3. « En combien de temps on relit une PR ? Je propose 2 heures en journée. »

Juste après la réunion, `Y4NN777` consigne les trois réponses et sa décision sur O2 dans le [journal des décisions](../../ROADMAP.md#14-decisions-log), une ligne datée chacune (TODO « Now » ne contient que la ligne de chaque personne).

**Pour conclure :** « Chacun a sa première tâche, son relecteur et sa page ici. Tout ce qui bouge passe par TODO.md. On se retrouve au point de demain, et samedi midi on décide ce qui saute. »

## 10. FAQ

| Question | Réponse |
|---|---|
| Où je mets mon code ? | Dans les chemins autorisés de ta tranche : ta page en [section 5](#5-une-page-par-personne) et ta section de [TODO](../../TODO.md) |
| Je suis bloqué, je fais quoi ? | Au bout de 30 minutes, tu le dis sur le canal de l'équipe ; la personne qui débloque est en [section 8](#8-blocages-connus-et-points-ouverts) |
| Je veux ajouter une fonctionnalité | Non sans accord du propriétaire : rien de nouveau hors du [PRD](../PRD.md) et de la [SPEC](../SPEC.md), et le P1/P2 ne se présente pas comme fait ([non-objectifs](../PRD.md#non-objectifs-de-cadev)) |
| Comment je lance les tests ? | `make quality` lance tout G1 à G4 ; les étapes isolées sont `make test-unit`, `make coverage`, `make test-integration` ([README](../../README.md#lancer-le-projet-en-local)) |
| Comment j'ajoute une migration ? | Tu ne le fais pas : tu le demandes à `mibienpanjoe`, seul auteur, qui utilise `make migrate-new NAME=<slug>` (R6) |
| Comment je change le contrat d'API ? | Une PR dédiée sur [api/openapi.yaml](../../api/openapi.yaml), relue par `Y4NN777` et `PANK4SS`, fusionnée avant le code, puis `make generate` (R7) |
| Où sont les données de démo ? | Nulle part pour l'instant : `make seed` n'est pas implémenté. Elles viendront dans `db/seeds`, pour `local-dev` seulement ([TODO SEED](../../TODO.md#seed--original-demo-fixtures-and-make-seed)) |
| La CI est rouge, je fais quoi ? | Tu lis l'étape en échec dans le résumé de `make quality`, tu corriges sur ta branche. On ne baisse aucun seuil, on ne saute aucun test (R4) |
| Qui approuve quoi ? | Le relecteur de ta zone approuve ta PR ([ADR-0008](../decisions/0008-team-work-areas-and-ownership.md)) ; le propriétaire décide des ADR, dépendances, migrations sur données réelles, secrets, coûts et production |
| Ma route est refusée par le gate d'autorisation | Elle doit être dans `scripts/public-routes.txt` avec une raison, ou couverte par un test négatif listé (B1, [matrice](../spec-kit/AUTHORIZATION_MATRIX.md#matrix)) |
| Un prototype dit autre chose que la SPEC | Le prototype a tort ([DISCLAIMER](../../prototypes/html/DISCLAIMER.md)) |
| Je peux utiliser une IA pour coder ? | Oui, mais tu restes l'auteur, sans `Co-authored-by` d'agent, et tu le déclares dans [AI_USAGE.md](../../AI_USAGE.md) |
| Qu'est-ce que je ne fais jamais ? | Commiter un secret, pousser sur `develop` ou `main`, relire ma propre PR, éditer une migration fusionnée ou du code généré, lancer quelque chose en production, présenter un `NOT RUN` comme `PASS` |
