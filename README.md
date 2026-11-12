<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/brand/226os-logo-white.png">
    <img src="docs/brand/226os-logo.png" alt="226os" height="72">
  </picture>
</p>

<p align="center"><strong>Apprendre, se connecter, résoudre, découvrir et construire</strong><br>Entre développeurs africains</p>

226os (« 226 Open Stack », 226 étant l'indicatif téléphonique du Burkina Faso) est une plateforme qui relie l'entraide technique, les ressources utiles aux développeurs africains, les projets open source et l'apprentissage. Le but n'est pas d'ajouter un forum, un annuaire d'API ou un réseau professionnel de plus, mais de raccourcir le chemin entre un problème rencontré, une solution documentée et une contribution que d'autres peuvent réutiliser.

Positionnement : **Burkina Faso et Afrique de l'Ouest d'abord, puis toute l'Afrique.**

## Statut

| Sujet | État |
| --- | --- |
| Produit | Implémentation P0 en cours, suivie dans [ROADMAP.md](ROADMAP.md) et [TODO.md](TODO.md) |
| API et socle | Socle Go, contrat OpenAPI, migrations, secrets chiffrés et squelette web fusionnés ; les opérations métier répondent 501 tant que leur tranche n'est pas livrée |
| Gates G1 à G4 | Exécutées par la CI en groupes parallèles ; `make quality` en est la définition locale |
| Gates G5 (e2e) et G6 (performance) | NOT IMPLEMENTED |
| Gate G7 (livraison) | NOT RUN |
| Déploiement | Rien n'est déployé ; le domaine [226os.tech](https://226os.tech) est réservé |
| Contexte de livraison | CADEV 2026, soumission avant le 15 novembre 2026 à 17:00 GMT ; équipe 7's, quatre membres |

## Ce que c'est

Un parcours en quatre temps : découvrir des ressources, résoudre un problème, construire avec d'autres, puis rendre à la communauté. La navigation s'organise autour de ces piliers :

- **Apprendre** : un parcours et un assistant qui cite ses sources.
- **Connect** : mise en relation entre développeurs, dans le respect de la vie privée.
- **Solve** : questions et réponses techniques.
- **Discover** : annuaire d'API et d'outils africains.
- **Build** : projets open source et collaborations.

Les neuf domaines, les priorités P0, P1 et P2 et les amendements sont dans le [PRD](docs/PRD.md).

## Démarrage rapide

Prérequis, vérifiés par `make doctor` : Go, Node, pnpm (versions fixées dans `go.mod` et `package.json`), Docker avec le plugin Compose, Git et Python 3.

```sh
git clone https://github.com/Y4NN777/226os.git && cd 226os
make doctor        # vérifie la chaîne d'outils, sans rien modifier
make setup         # go mod download et pnpm install --frozen-lockfile
make up-local-dev  # démarre le profil local-dev (images construites, migrations appliquées)
make down-local-dev
make help          # liste toutes les cibles ; `make -n <cible>` affiche une recette sans l'exécuter
```

Les secrets sont chiffrés dans Git (SOPS et age) et déchiffrés à l'exécution. Avant le premier démarrage, créez votre clé puis le fichier `local-dev` : tout est expliqué dans [secrets/README.md](secrets/README.md), y compris l'accès des autres membres.

Le tableau des cibles Make est dans [docs/DEV.md](docs/DEV.md).

## Architecture

Un monolithe Go modulaire (Chi, pgx, sqlc) expose une API REST décrite en OpenAPI. Une application React 19 indépendante (TypeScript, Vite, TanStack, Astryx) la consomme. PostgreSQL sert de source de vérité et de moteur de recherche (plein texte, GIN, `pg_trgm`). Docker Compose et Traefik assurent l'exécution, GitHub Actions la CI, et la cible d'hébergement est Datacloud.

Définition de « fini » : `make quality` passe, avec des gates déterministes (G1 à G4) et une revue humaine. Le détail est dans [ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md) et [TEST.md](docs/TEST.md).

## Structure du dépôt

```text
226os/
├── src/api/          # API Go (cmd/, internal/)
├── src/web/          # application React + TypeScript + Vite
├── api/              # contrat REST openapi.yaml
├── db/               # migrations goose et requêtes sqlc
├── deploy/           # configuration Traefik
├── compose.yaml      # profils local-dev, staging-test, production-live
├── Makefile, mk/     # point d'entrée unique des gates et des opérations
├── scripts/          # scripts appelés par Make et la CI
├── secrets/          # secrets chiffrés et clés publiques
├── docs/             # PRD, SPEC, architecture, ADR, tests, livraison
├── prototypes/html/  # aperçu de design statique, pas le produit
├── ROADMAP.md        # plan jusqu'à la soumission CADEV
└── TODO.md           # suivi en direct des tâches
```

## Contribuer

Lisez d'abord [CONTRIBUTING.md](CONTRIBUTING.md) et la [page d'accueil de l'équipe](docs/team/KICKOFF.md). Les branches partent de `develop` ([GIT_WORKFLOW](docs/GIT_WORKFLOW.md)), et chaque PR met à jour [TODO.md](TODO.md) dans le même changement. Chaque contribution est signée par la personne qui en répond ; aucun agent IA n'apparaît comme auteur, et l'aide d'une IA se déclare dans [AI_USAGE.md](AI_USAGE.md).

## Documentation

| Document | Rôle |
| --- | --- |
| [PRD](docs/PRD.md) | Vision, neuf domaines, priorités |
| [SPEC](docs/SPEC.md) | User stories, scénarios BDD, traçabilité |
| [Architecture](docs/architecture/ARCHITECTURE.md) | Architecture approuvée et frontières techniques |
| [Développement](docs/DEV.md) | Organisation du code, cibles Make, workflow |
| [Tests](docs/TEST.md) | Stratégie de tests, gates et CI |
| [Livraison](docs/SHIP.md) | Déploiement, vérification, rollback |
| [Décisions](docs/decisions/README.md) | ADR : toute dérogation passe par un nouvel ADR |
| [Design](DESIGN.md) | Direction visuelle, tokens et composants |

## Originalité, IA et dépendances

226os est un produit original, développé par notre équipe sous licence Apache-2.0. Nous citons nos sources d'inspiration, dont 225 Open Source, sans copier de code, de texte, de visuels ni de données : voir la [déclaration d'originalité](docs/research/REFERENCES.md#déclaration-doriginalité). L'aide d'une IA est déclarée dans [AI_USAGE.md](AI_USAGE.md), les dépendances et leurs licences dans [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md), et l'audit de soumission dans [docs/SHIP.md](docs/SHIP.md).

## Licence

Le code original de 226os est publié sous licence [Apache 2.0](LICENSE) (voir aussi [NOTICE](NOTICE)). Les dépendances tierces conservent leurs licences.
