# 226os : page d'accueil de l'équipe

> **Lis ceci en premier.** Point d'entrée de l'équipe : contexte, carte des documents, ta mission, les règles. Ce fichier ne remplace rien : le plan fait foi dans [ROADMAP.md](../../ROADMAP.md), le suivi vivant dans [TODO.md](../../TODO.md). En cas de désaccord, on corrige ce fichier.
>
> **Maintenu par :** `Y4NN777`. Les heures sont en GMT.

## 1. Contexte

**226os** relie l'entraide technique, les API et outils utiles aux développeurs africains, les projets open source, l'apprentissage et les personnes. L'état du dépôt (ce qui existe, ce qui est construit, ce qui est déployé) est dans le [README](../../README.md) et dans [TODO](../../TODO.md#now) : on ne le recopie pas ici. Les décisions sont dans le [registre des ADR](../decisions/README.md) et le [journal des décisions](../../ROADMAP.md#14-decisions-log) de la ROADMAP.

Règles d'honnêteté : G5 et G6 ne sont pas implémentés, G7 est `NOT RUN`, rien n'est déployé tant que la checklist de déploiement de TODO ne le dit pas. S10 (canaux en direct) n'est pas construit : il reste P2, décision du point d'étape du 14 novembre à 12:00 ([ROADMAP, section 14](../../ROADMAP.md#14-decisions-log)).

**Jour 1 :** identité Git, `make doctor`, `make setup`, clé age pour les secrets, `make up-local-dev` : tout est dans [CONTRIBUTING, « Day one »](../../CONTRIBUTING.md#day-one).

## 2. Carte des documents

| Je cherche… | Je vais voir… |
|---|---|
| Le produit | [README](../../README.md), [PRD](../PRD.md) |
| Les comportements attendus | [SPEC](../SPEC.md), [Spec Kit](../spec-kit/README.md) |
| Le contrat de l'API | [api/openapi.yaml](../../api/openapi.yaml), [ADR-0004](../decisions/0004-api-contract-conventions.md) |
| Les données | [DATABASE_DESIGN](../spec-kit/DATABASE_DESIGN.md), [db/migrations](../../db/migrations/), [ADR-0005](../decisions/0005-data-model-migrations-search.md) |
| Qui a le droit de faire quoi | [AUTHORIZATION_MATRIX](../spec-kit/AUTHORIZATION_MATRIX.md), [ADR-0003](../decisions/0003-authentication-sessions-csrf-rate-limits.md) |
| L'architecture et les ADR | [ARCHITECTURE](../architecture/ARCHITECTURE.md), [registre des ADR](../decisions/README.md) |
| Le design | [DESIGN.md](../../DESIGN.md), [ADR-0007](../decisions/0007-astryx-version-pinning-and-verification.md) |
| Les tests et la couverture | [TEST](../TEST.md), [COVERAGE_SCOPE](../quality/COVERAGE_SCOPE.md), [RULES](../RULES.md), [DEV](../DEV.md) |
| Git, branches, PR | [GIT_WORKFLOW](../GIT_WORKFLOW.md), [CONTRIBUTING](../../CONTRIBUTING.md), [CONTRIBUTORS](CONTRIBUTORS.md) |
| Les commandes Make | [README](../../README.md), [DEV](../DEV.md), `make help` |
| Les secrets | [secrets/README](../../secrets/README.md), [ADR-0009](../decisions/0009-secrets-with-sops-and-age.md) |
| Le plan, le calendrier, les risques | [ROADMAP](../../ROADMAP.md) |
| Le suivi vivant | [TODO](../../TODO.md) |
| Livraison et règles CADEV | [SHIP](../SHIP.md), [TODO S8](../../TODO.md#s8--ops-release-and-cadev) |
| Licences et déclaration de l'IA | [THIRD_PARTY_NOTICES](../../THIRD_PARTY_NOTICES.md), [AI_USAGE](../../AI_USAGE.md) |
| Les prototypes | [DISCLAIMER](../../prototypes/html/DISCLAIMER.md) (à lire avant d'ouvrir une page) |
| Les règles des agents IA | [AGENTS.md](../../AGENTS.md), [CLAUDE.md](../../CLAUDE.md) |

## 3. Une page par personne

Les relecteurs viennent de [ADR-0008](../decisions/0008-team-work-areas-and-ownership.md) et de [TODO](../../TODO.md). Ta ligne dans [TODO, Now](../../TODO.md#now) est la seule que tu modifies. Nul ne relit une PR qu'il a écrite ou co-écrite.

### `Y4NN777`

- **Mission :** propriétaire, lead backend et ops. Tranches S0, S6, S8, S9 ; aide sur S2, S4, S5, E2.
- **Premier test RED (S0) :** une route parasite, un nom de test absent et une route non listée font chacun passer le gate B1 au rouge ([TODO S0](../../TODO.md#s0--api-core)).
- **Chemins autorisés (S0) :** `src/api/cmd/**`, `src/api/internal/app/**`, `src/api/internal/platform/**` (apigen généré seulement), `scripts/**`, `compose.yaml`, `deploy/**`, `.env.example`, `mk/**` (un changement par PR).
- **Relecteur :** `NICE-DEV226` pour S0 ; `mibienpanjoe` pour S6 et S9 ; `NICE-DEV226` ou `mibienpanjoe` pour S8. Tu relis le front et, avec `PANK4SS`, chaque PR de contrat.

### `NICE-DEV226`

- **Mission :** développeur backend. Tranches S1 (identité, sur le chemin critique), S3 (Solve), puis le backend E5 si le temps le permet ; second développeur de S9.
- **Premier test RED (S1) :** à la connexion, la session tourne et l'ancien jeton est refusé ; une session révoquée ou expirée est inutilisable ([TODO S1](../../TODO.md#s1--identity), ADR-0003).
- **Chemins autorisés (S1) :** `src/api/internal/identity/**`, `db/queries/identity.sql`, `db/queries/platform.sql`, `src/api/internal/platform/dbgen/**` (généré), `scripts/public-routes.txt`. Aucune route montée avant la fusion de S0.
- **Relecteur :** `mibienpanjoe` pour S1 et S3 (le lead peut co-relire la surface d'auth).

### `mibienpanjoe`

- **Mission :** développeur backend et **seul auteur des migrations**. Tranches S2 (Discover), S4 (Learn et script de contenu), S5 (Build), E2 et `profile_visibility`. Responsable même quand on t'aide.
- **Premier test RED (E2) :** mettre à jour d'abord les tests de migration pour qu'ils échouent, puis écrire les migrations ([TODO E2](../../TODO.md#e2--taxonomy-add-and-relabel-only)). La première PR de contrat (C2, C1, C14b) passe avant les handlers S4.
- **Chemins autorisés :** contrat : `api/openapi.yaml`, puis fichiers générés par `make generate` uniquement. E2 : nouveaux fichiers `db/migrations/**`, `src/api/internal/platform/pgtest/**`. S2 : `src/api/internal/discover/**`, `src/api/internal/search/**`, requêtes `db/queries/**` concernées. S4 : `src/api/internal/learn/**`, `scripts/content/**`.
- **Relecteur :** `NICE-DEV226` ; s'il a co-écrit la PR, `Y4NN777`. Contrat : `Y4NN777` et `PANK4SS`.

### `PANK4SS`

- **Mission :** responsable du front. Tranches S7 (prototypes vers pages React), E4, interface E5, SEED. **Accepte l'invitation GitHub avant tout push.**
- **Premier test RED (S7) :** changer de langue modifie les libellés, jamais le contenu d'origine. Pour SEED : `make seed` refuse hors `local-dev` ([TODO S7](../../TODO.md#s7--frontend-prototypes-become-react-pages), [TODO SEED](../../TODO.md#seed--original-demo-fixtures-and-make-seed)).
- **Chemins autorisés :** S7 : `src/web/**` et `DESIGN.md`. SEED : `db/seeds/**` et la cible `seed` de `mk/50-ops.mk`. `prototypes/html/**` est hors périmètre tant que le propriétaire ne l'a pas confirmé.
- **Relecteur :** `Y4NN777` pour S7, E4 et l'interface E5 ; `mibienpanjoe` pour SEED (et `Y4NN777` pour `mk/50-ops.mk`).
- **Depuis le 2026-11-13 (décision du propriétaire, TODO S7) :** `Y4NN777` porte S7 et en a construit les pages ; `PANK4SS` relit les PR S7 et suit l'issue #109.

## 4. Règles de travail

Texte complet : [ROADMAP, section 5](../../ROADMAP.md#5-rules-for-every-slice-r1-to-r11-referenced-from-todomd).

| Règle | En clair |
|---|---|
| R1 Test d'abord | Un test qui échoue (RED), le code qui le fait passer (GREEN), puis la refacto. Invariants et autorisations testés sur un vrai PostgreSQL |
| R2 Couverture Go | 100,00 % des instructions Go du projet ; une seule exclusion : `apigen.gen.go` ([COVERAGE_SCOPE](../quality/COVERAGE_SCOPE.md)). SMTP et LLM derrière des ports avec fakes |
| R3 Couverture web | Strictement au-dessus de 70 % sur instructions, branches, fonctions et lignes |
| R4 CI verte | `make quality` vert dans la CI avant toute fusion. On n'affaiblit jamais un gate et on ne baisse pas un seuil |
| R5 Autorisations | Une route publique va dans `scripts/public-routes.txt` avec sa raison ; toute autre route a ses tests négatifs |
| R6 Migrations | Seul `mibienpanjoe` en écrit, une à la fois, en avant uniquement, immuables après fusion |
| R7 Fichiers partagés | `api/openapi.yaml`, `go.mod`/`go.sum`, `pnpm-lock.yaml`, `mk/**` : un seul éditeur à la fois, un changement par PR. Le contrat fusionne d'abord, relu par `Y4NN777` et `PANK4SS` |
| R8 Branches | Une tranche, une branche depuis `develop`, PR vers `develop` ([GIT_WORKFLOW](../GIT_WORKFLOW.md#work-branch-naming)). Personne ne relit sa propre PR |
| R9 Commits | Ton identité Git confirmée. Aucune attribution IA (pas de `Co-authored-by` d'agent, pas de pied « Generated with ») ; l'aide d'une IA se déclare dans un fichier de [docs/ai-usage/](../ai-usage/README.md), jamais en modifiant [AI_USAGE.md](../../AI_USAGE.md) |
| R10 Données et secrets | Aucun secret dans Git, les logs, les fixtures, les prompts ou les captures. Les données de démo ne tournent jamais hors `local-dev` |
| R11 Statut honnête | Jamais de P1 ou P2 présenté comme fait. Un contrôle non lancé est `NOT RUN`, jamais `PASS` |

Habitudes : un aide est nommé sur la PR et le propriétaire de la tranche reste responsable ; bloqué plus de 30 minutes, tu le dis sur le canal de l'équipe avec ce que tu as essayé ; moins de PR, mieux regroupées (une tranche par PR, les changements de docs ou de chore liés dans la même).

## 5. Comment faire une PR

Le [modèle de PR](../../.github/PULL_REQUEST_TEMPLATE.md) s'ouvre tout seul ; la liste ci-dessous en est la version courte.

- [ ] Branche depuis `develop`, nommée selon [GIT_WORKFLOW](../GIT_WORKFLOW.md#work-branch-naming).
- [ ] Test RED écrit et nommé dans la description, puis GREEN.
- [ ] `make quality` passe en local ; la CI est verte sur le SHA de la PR.
- [ ] Couverture Go à 100,00 %, web au-dessus de 70 % si tu y as touché.
- [ ] Nouvelles routes : dans `public-routes.txt` avec leur raison, ou couvertes par des tests négatifs.
- [ ] Code généré jamais modifié à la main ; `make generate` relancé après un rebase si besoin.
- [ ] Description : story, relecteur, aides éventuelles, commandes lancées avec leur résultat (`PASS` / `FAIL` / `NOT RUN`).
- [ ] `TODO.md` mis à jour dans la même PR.
- [ ] Aucun secret, aucun `Co-authored-by` d'agent, aide IA déclarée dans un fichier de `docs/ai-usage/`.
- [ ] Fichiers CADEV (`THIRD_PARTY_NOTICES.md`, `AI_USAGE.md`) : intacts ou seulement complétés.
- [ ] Le bon relecteur est demandé (voir ta page en section 3).
