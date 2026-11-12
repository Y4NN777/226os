<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/brand/226os-logo-white.png">
    <img src="docs/brand/226os-logo.png" alt="226os" height="72">
  </picture>
</p>

<p align="center"><strong>Apprendre, se connecter, résoudre, découvrir et construire</strong><br>Entre développeurs africains</p>

226os est une plateforme qui relie l'entraide technique, les ressources utiles aux développeurs africains, les projets open source, l'apprentissage et les personnes qui font avancer tout cela. Le but n'est pas d'ajouter un forum, un annuaire d'API ou un réseau professionnel de plus, mais de raccourcir le chemin entre un problème rencontré, une solution documentée et une contribution que d'autres peuvent réutiliser.

Son positionnement : **Burkina Faso et Afrique de l'Ouest d'abord, puis toute l'Afrique**. La plateforme part du contexte ouest-africain et s'adresse ensuite aux développeurs de tout le continent, quels que soient leur pays, leur langue ou leur niveau.

226os se lit « 226 Open Stack » : 226 est l'indicatif téléphonique du Burkina Faso.

> **Domaine :** [226os.tech](https://226os.tech), réservé, rien n'y est encore déployé.
> **Équipe :** 7's, quatre membres.
> **Premier contexte de livraison :** CADEV 2026, soumission avant le 15 novembre 2026 à 17:00 GMT.

## L'orientation actuelle

Depuis le 12 novembre 2026, la navigation, l'accueil et les textes mettent en avant deux piliers : **Apprendre** et **Connect**. Solve (questions), Discover (API et outils) et Build (projets open source) restent des piliers à part entière, accessibles depuis la même navigation. Rien n'est retiré : les neuf domaines et les quatre expériences différenciantes du [PRD](docs/PRD.md) restent en place. C'est un changement de mise en avant, pas de périmètre (amendement 3.2 du PRD, proposé).

Tout reste un seul écosystème : un compte, une recherche unique, des tags et des thèmes partagés, des liens typés entre contenus, un assistant qui s'appuie sur tout ce qui est publié, et un annuaire qui relie les personnes à leurs contributions.

Ce qui est prévu au P0 (rien n'est encore livré) :

- **Assistant IA sourcé :** il répond uniquement à partir des contenus publiés sur 226os, cite ses sources, refuse quand il n'en a pas et termine chaque réponse par une section « Comment vérifier ». Il reste désactivé par configuration tant que le propriétaire n'a pas tranché les conditions d'utilisation du fournisseur.
- **Connect :** un annuaire des membres en lecture seule, filtrable par pays, langue et compétence, qui respecte la visibilité choisie par chacun. Le mentorat, les opportunités et les espaces communautaires sont affichés comme « bientôt ».
- **Apprendre :** des thèmes tournés vers le développement à l'ère des agents IA, un bouton « Poser une question sur cette leçon » et des fils par thème.
- **Contenu original :** aucun contenu tiers n'est importé ni aspiré. Les sources externes apparaissent seulement comme références et comme liens documentaires sur les fiches Discover.

## Pourquoi 226os ?

Intégrer un service de paiement, retrouver un exemple fiable, comprendre une erreur de déploiement ou trouver des contributeurs oblige souvent à passer d'un groupe de discussion à une documentation incomplète, puis à plusieurs dépôts GitHub. Une solution trouvée se perd dans un fil de messages ; un projet intéressant ne croise pas les bonnes compétences.

Ce qui manque, c'est la **continuité du contexte** : quels services existent dans quels pays, quelles contraintes techniques se posent vraiment, quelles solutions sont vérifiables et qui peut faire avancer un projet. 226os veut relier ces éléments plutôt que les juxtaposer.

## Un parcours continu

**Discover → Solve → Build → Share back**

Prenons un développeur d'Afrique de l'Ouest qui intègre un service, un modèle ou une API dans un produit utilisé avec une connexion limitée :

1. **Discover :** il repère les fournisseurs, les pays couverts, les SDK et les modèles, les exemples et les limites documentées.
2. **Solve :** il pose une question avec son contexte technique et son code, échange avec la communauté et documente la résolution.
3. **Build :** il trouve un projet open source lié à son besoin, propose une contribution ou cherche d'autres développeurs.
4. **Share back :** le correctif, l'exemple ou le guide devient une ressource que d'autres retrouvent et évaluent.

## Les neuf domaines du produit

| Domaine | Fonctionnalités envisagées |
| --- | --- |
| **Solve, entraide et connaissances** | Questions et réponses, snippets, solutions acceptées, votes et réputation, revue de code et discussions thématiques. |
| **Discover, ressources pour développeurs africains** | Catalogue d'API, de SDK et d'outils, couverture par pays, documentation, comparaisons, exemples d'intégration et retours de la communauté. |
| **Build, open source et collaboration** | Découverte et import de projets GitHub, publication de projets, besoins en compétences, contributeurs et cofondateurs, candidatures et reconnaissance des contributions. |
| **Learn, apprendre** | Articles et tutoriels (Markdown, code, tags, niveau, langue) liés aux questions, aux API et aux projets ; série gratuite sur l'IA lisible sans compte ; publication ouverte à tous les membres. Les séries avec progression, les certificats et les exercices interactifs viennent ensuite (P1 et P2). |
| **Connect, mentorat et carrière** | Annuaire des membres (P0) ; mentorat, relecture de CV et de portfolios, préparation aux entretiens, groupes d'étude, stages et opportunités (plus tard). |
| **Identity & Community** | Profils développeurs, visibilité du profil, compétences, contributions, réputation, communautés et activité. |
| **Language, collaboration panafricaine** | Interface en français et en anglais, publications multilingues, recherche bilingue, traduction assistée qui conserve le texte original. |
| **Access, connexions limitées** | Interface mobile d'abord, chargement progressif, transferts réduits, cache, brouillons hors ligne et dégradation maîtrisée. |
| **Intelligence, assistance contextuelle** | Assistant qui répond à partir des contenus publiés avec ses sources (P0), questions similaires, ressources et projets liés, aide à la rédaction, résumés. |

Ces domaines décrivent la **vision complète**, pas ce qui est disponible aujourd'hui. Les priorités P0, P1 et P2 sont dans le [PRD](docs/PRD.md).

### Ce qui doit nous distinguer

- **Le contexte des intégrations africaines.** Une fiche d'API indique les pays couverts, les options techniques, les limites documentées et les échanges qui étayent ces informations.
- **Du problème à la contribution.** Une bonne réponse peut nourrir un guide, une issue GitHub, un correctif ou un nouveau projet, sans perdre son contexte.
- **La collaboration au-delà des frontières.** Les compétences et les besoins guident la découverte de contenus et de contributeurs, au-delà de la langue et du pays.
- **La connaissance technique appuyée par des preuves.** Versions, étapes de reproduction, sources et confirmations permettent de juger une solution autrement qu'au nombre de votes.

Apprendre irrigue ces quatre expériences : un article renvoie vers les questions, les API et les projets qu'il met en pratique, et inversement.

## État du projet

- **Bootstrap B0 et cœur de l'API (S0) fusionnés** (PR #24, #31 à #39) : API Go avec `GET /healthz`, `GET /readyz` et les 47 opérations du contrat montées sous `/api/v1` (encore en 501 tant que leur tranche n'est pas livrée), validation du contrat, erreurs problem+json, curseurs signés, idempotence, envoi de mail, migrations `00001` à `00010`, requêtes sqlc d'identité et de référence, squelette web (Vite, React 19, TanStack, Astryx 0.7.14), profils Compose, cibles Make et CI.
- **Implémentation P0 en cours**, découpée en tranches dans [ROADMAP.md](ROADMAP.md) et suivie dans [TODO.md](TODO.md). Le contrat compte aujourd'hui 43 opérations P0.
- **Gates :** la CI lance les gates G1 à G4 en groupes parallèles (une tâche par groupe, puis la tâche `quality` vérifie que chaque étape de `make quality` a réussi pour le commit testé, [ADR-0010](docs/decisions/0010-ci-parallel-gate-groups.md)) sur chaque PR vers `develop` ou `main` et à chaque push sur `develop`. `make quality` reste la définition locale, séquentielle. G5 (e2e), G6 (performance) et G7 (livraison) sont **NOT RUN**.
- **Rien n'est déployé**, ni en préproduction ni en production.

Aucune fonctionnalité P1 ou P2 n'est présentée comme livrée. La participation à CADEV impose un projet original, un dépôt Git, une démonstration en ligne sur un service Datacloud et la déclaration des dépendances, de leurs licences et de l'aide d'une IA.

## Lancer le projet en local

Prérequis vérifiés par `make doctor` : la version de Go fixée dans `go.mod`, Node (plancher `engines.node`), la version exacte de pnpm fixée dans `package.json`, Docker avec le plugin Compose, Git et Python 3.

```sh
make doctor        # vérifie la chaîne d'outils, sans rien modifier
make setup         # go mod download et pnpm install --frozen-lockfile
make up-local-dev  # démarre le profil local-dev (images construites, migrations appliquées)
make down-local-dev
```

`make up-local-dev` lit `.env.local-dev`, à créer à partir de `.env.example`. Ce fichier doit définir `APP_ENV=local-dev` une seule fois, sinon `scripts/compose-env-guard.sh` refuse de démarrer. Ne le committez pas.

Lisez une recette avec `make -n <cible>` avant de la lancer pour la première fois ; `make help` liste toutes les cibles.

| Cible | Rôle | État |
| --- | --- | --- |
| `make quality` | G1 à G4 : `doctor generate lint typecheck test-unit coverage test-integration security licenses` | disponible ; `test-integration` et `security` ont besoin de Docker |
| `make generate`, `lint`, `typecheck`, `test-unit`, `coverage`, `test-integration` | étapes de G1 à G3 ; `test-integration` lance un PostgreSQL 18 éphémère | disponible |
| `make security`, `make licenses` | G4 : routes et autorisations, vulnérabilités, secrets, SAST, images, licences | disponible |
| `make build` | images `api`, `web` et `migrate`, empreintes dans `reports/image-digests.txt` | disponible |
| `make migrate-up`, `migrate-status`, `migrate-new NAME=<slug>` | migrations goose, uniquement en avant ; hors `local-dev`, `migrate-up` exige une confirmation et une preuve de sauvegarde | disponible |
| `make db-reset` | recrée la base, avec `APP_ENV=local-dev` uniquement | disponible |
| `make up-staging-test`, `down-staging-test` | profil de préproduction, avec `.env.staging-test` hors Git | disponible ; aucune préproduction n'est encore déployée |
| `make up-production-live` | production | refusé hors du job de livraison automatisé, qui n'existe pas encore |
| `make test-e2e` (G5), `make perf` (G6), `make seed` | e2e, performance, données de démonstration | **NOT IMPLEMENTED** |
| `make quality-full` | `quality` plus `test-e2e` et `perf` | échoue tant que G5 et G6 ne sont pas implémentés |

## Organisation du dépôt

```text
226os/
├── src/api/               # API Go (Chi, pgx, sqlc) : cmd/api, internal/
├── src/web/               # application React 19 + TypeScript + Vite
├── api/openapi.yaml       # contrat REST, source des types Go et TypeScript générés
├── db/migrations/         # migrations SQL goose, immuables après fusion
├── db/queries/            # requêtes sqlc (db/sqlc.yaml)
├── scripts/               # scripts appelés par Make et la CI
├── deploy/                # configuration Traefik
├── compose.yaml           # profils local-dev, staging-test, production-live
├── Makefile, mk/          # point d'entrée unique des gates et des opérations
├── .github/               # workflows quality et branch-policy, CODEOWNERS
├── prototypes/html/       # aperçu de design, pas le produit
├── docs/                  # PRD, SPEC, architecture, ADR, tests, livraison
├── ROADMAP.md             # plan jusqu'à la soumission CADEV
└── TODO.md                # suivi en direct des tâches
```

`prototypes/html` est un aperçu de design statique, pas le produit : ses données sont fictives et ses actions simulées. Voir [`prototypes/html/DISCLAIMER.md`](prototypes/html/DISCLAIMER.md).

## Où l'équipe se synchronise

- [docs/team/KICKOFF.md](docs/team/KICKOFF.md) : **la page d'accueil de l'équipe**, à lire en premier : contexte, carte des documents, glossaire, une page par personne et FAQ.
- [ROADMAP.md](ROADMAP.md) : le plan, les tranches, les règles, les décisions datées et les points ouverts.
- [TODO.md](TODO.md) : le suivi en direct. Mettez-le à jour **dans la même PR que le travail**, une ligne par changement.
- [docs/team/CONTRIBUTORS.md](docs/team/CONTRIBUTORS.md) : qui fait quoi (logins GitHub uniquement) et les étapes avant le premier commit.
- [docs/GIT_WORKFLOW.md](docs/GIT_WORKFLOW.md) : branches `feature/<domaine>/<nom>` depuis `develop`, PR vers `develop`, promotion vers `main` par PR de production.
- [docs/decisions/](docs/decisions/) : les ADR. Une décision acceptée ne change que par un nouvel ADR.

Chaque contribution est signée par la personne qui en répond, avec son identité Git validée. Aucun agent IA n'apparaît comme auteur ni dans un `Co-authored-by` ; l'aide d'une IA se déclare dans [AI_USAGE.md](AI_USAGE.md).

## Architecture

Un **monolithe Go modulaire** expose une API REST décrite en OpenAPI, une application React indépendante la consomme, et PostgreSQL sert à la fois de source de vérité et de moteur de recherche. Aucun composant d'infrastructure n'est ajouté sans besoin justifié et sans ADR.

| Couche | Choix |
| --- | --- |
| **Frontend** | React 19, TypeScript, Vite, TanStack Router, Query et Form ; Meta Astryx 0.7.14 comme système de composants principal ([ADR-0007](docs/decisions/0007-astryx-version-pinning-and-verification.md), [DESIGN.md](DESIGN.md)). |
| **Backend** | Go, Chi, pgx, sqlc ; REST et OpenAPI ; domaines métier séparés. |
| **Base et recherche** | PostgreSQL 18, recherche plein texte, index GIN, `pg_trgm`, migrations SQL versionnées. |
| **Cache et stockage objet** | Valkey et le stockage compatible S3 sont exclus au P0 ([ADR-0006](docs/decisions/0006-toolchain-environments-quality-commands.md)) ; les adopter demandera un nouvel ADR. |
| **Infrastructure** | Docker, Docker Compose, Traefik et HTTPS, hébergement Datacloud. |
| **Automatisation** | Make comme interface unique ; GitHub Actions pour la CI. |

La méthode : contrats BDD, cycle TDD, tests d'intégration sur un vrai PostgreSQL, contrôles de sécurité et **gates déterministes** avant livraison. Les seuils de couverture sont **100 % des instructions Go du projet (hors code généré)** et **plus de 70 % sur les quatre métriques frontend** ; la couverture ne remplace ni les tests d'autorisation ni la vérification des parcours complets.

## Inspirations et originalité

226os est un produit original, développé par notre équipe pendant la période du concours, sous licence Apache-2.0. Nous citons ouvertement 225 Open Source comme une source d'inspiration pour la découverte de projets et l'identité communautaire, comme les autres projets open source que nous avons étudiés. Nous avons étudié des comportements et des principes de conception, sans copier de code, de texte, de visuels ni de données. Chaque dépendance tierce est déclarée, épinglée et vérifiée côté licence. Notre orientation produit nous est propre : une plateforme qui part de l'Afrique de l'Ouest et s'ouvre à toute l'Afrique, avec un assistant d'apprentissage qui cite ses sources, un parcours pour apprendre et progresser à l'ère des agents, un annuaire d'API et d'intégrations africaines, et une mise en relation qui respecte la vie privée.

L'étude détaillée de 225 Open Source (ce que nous avons observé, ce que nous faisons différemment, nos garde-fous) est dans [`docs/research/REFERENCES.md`](docs/research/REFERENCES.md#étude--225-open-source-référence-dinspiration).

Références étudiées, pour comprendre les bons compromis :

- **[225 Open Source](https://github.com/kouame09/225_os)** : découverte de projets, lancements, articles et visibilité des développeurs. Une référence régionale qui montre qu'un annuaire seul ne suffirait pas à nous distinguer.
- **[Apache Answer](https://github.com/apache/answer)** : structure des questions et réponses, recherche, réputation, historique.
- **[Flarum](https://github.com/flarum/framework)** : discussions, modération, permissions.
- **[Nuxt Hacker News](https://github.com/nuxt/hackernews)** : navigation sobre, lecture de contenus, attention aux performances.
- **[Meta Astryx](https://github.com/meta/astryx)** : système de design React retenu comme fondation principale.

Ce sont des **références de recherche, pas des dépendances** : elles inspirent des choix de comportement et d'ergonomie, rien de plus. Les analyses et les règles de réutilisation sont dans [`docs/research/REFERENCES.md`](docs/research/REFERENCES.md).

## Documentation

| Document | Rôle |
| --- | --- |
| [PRD](docs/PRD.md) | Vision, neuf domaines, priorités, amendements 3.1 (Apprendre) et 3.2 (proposé) |
| [SPEC](docs/SPEC.md) | User stories, scénarios BDD et traçabilité |
| [Design](DESIGN.md) | Direction visuelle, tokens et composants |
| [Architecture](docs/architecture/ARCHITECTURE.md) | Architecture approuvée et frontières techniques |
| [Development](docs/DEV.md) | Conventions et workflow de développement |
| [Testing](docs/TEST.md) | Stratégie de tests et gates |
| [Shipping](docs/SHIP.md) | Déploiement, vérification et rollback |
| [Rules](docs/RULES.md) | Contraintes d'ingénierie |
| [Git workflow](docs/GIT_WORKFLOW.md) | Branches, pull requests, préproduction et production |
| [Contributing](CONTRIBUTING.md) | Règles de contribution |
| [Agent instructions](AGENTS.md) | Orchestration des agents, permissions et preuves |

## Licence

Le **code original de 226os** est publié sous licence [Apache 2.0](LICENSE) (voir aussi [NOTICE](NOTICE)). Les dépendances tierces conservent leurs licences et leurs obligations de mention ([THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)).

---

**226os part du Burkina et de l'Afrique de l'Ouest. Son terrain de collaboration dépasse les frontières.**
