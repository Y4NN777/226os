<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/brand/226os-logo-white.png">
    <img src="docs/brand/226os-logo.png" alt="226os" height="72">
  </picture>
</p>

<p align="center"><strong>Apprendre, se connecter, résoudre, découvrir et construire</strong><br>Entre développeurs africains</p>

226os (« 226 Open Stack », 226 étant l'indicatif téléphonique du Burkina Faso) est une plateforme qui relie l'entraide technique, les ressources utiles aux développeurs africains, les projets open source et l'apprentissage. Le but n'est pas d'ajouter un forum, un annuaire d'API ou un réseau professionnel de plus, mais de raccourcir le chemin entre un problème rencontré, une solution documentée et une contribution que d'autres peuvent réutiliser.

Positionnement : **Afrique de l'Ouest d'abord, puis toute l'Afrique.**

## Statut

| Sujet | État |
| --- | --- |
| Produit | P0 : identité, Apprendre, Résoudre, Découvrir, Construire et l'assistant sont livrés sur `develop`. Restent la modération, la page Connect, les données de démo locales et la livraison. Les canaux en direct sont reportés en P2. État par lot dans [TODO.md](TODO.md), plan dans [ROADMAP.md](ROADMAP.md) |
| Web et API | Pages en ligne détaillées ci-dessous. La modération et les quatre opérations P1 répondent encore 501 |
| Gates G1 à G4 | Exécutées par la CI en groupes parallèles ; `make quality` en est la définition locale |
| Gates G5 (e2e) et G6 (performance) | NOT IMPLEMENTED |
| Gate G7 (livraison) | NOT RUN |
| Déploiement | Rien n'est déployé ; le domaine [226os.tech](https://226os.tech) est réservé |
| Contexte de livraison | CADEV 2026, soumission avant le 15 novembre 2026 à 17:00 GMT ; équipe 7's, quatre membres |

## Ce qui tourne sur `develop`

La navigation a deux groupes : « Progresser » (Apprendre, Connect) et « Piliers » (Résoudre, Découvrir, Construire). Une zone pas encore construite affiche « Bientôt » et n'est pas un lien.

| Pilier | Pages | État |
| --- | --- | --- |
| Apprendre | `/apprendre` (articles et formations, filtre par sujet), `/apprendre/articles/<id>`, `/apprendre/formations/<id>`, `/ecrire` pour publier un article | En ligne |
| Résoudre | `/solve` (questions), `/questions/<id>` (réponses, votes, réponse acceptée) | En ligne |
| Découvrir | `/discover` (API et outils africains, Afrique de l'Ouest par défaut), `/discover/<id>` | En ligne |
| Construire | `/build` (projets), `/build/<id>` (besoins de contribution, candidatures) | En ligne |
| Connect | annuaire de développeurs | « Bientôt » : l'API `GET /api/v1/profiles` est prête, la page ne l'est pas |

Autour des piliers :

- `/recherche` : recherche plein texte et assistant IA qui répond seulement à partir de sources publiées et citées, sinon il refuse. Le même panneau existe dans le formulaire de question de Résoudre. L'interface prévient que la question part chez le fournisseur (Google AI Studio). L'assistant est désactivé par défaut.
- `/profils/<username>` : profil d'un membre ; `/parametres` : nom affiché et visibilité du profil.
- `/` : page d'accueil publique pour les visiteurs, accueil membre une fois connecté.
- Pages de compte sans chrome : `/connexion`, `/inscription`, `/verifier-email`, `/mot-de-passe-oublie`, `/reinitialiser`, `/bienvenue`.

Les neuf domaines, les priorités P0, P1 et P2 et les amendements sont dans le [PRD](docs/PRD.md).

## Démarrage rapide

Prérequis vérifiés par `make doctor` : Go, Node et pnpm (versions fixées dans `go.mod` et `package.json`), Docker avec le plugin Compose, Git et Python 3. Il faut aussi `sops` et `age` pour les secrets, que `make doctor` ne vérifie pas.

```sh
git clone https://github.com/Y4NN777/226os.git && cd 226os
make doctor                        # vérifie la chaîne d'outils, sans rien modifier
make setup                         # go mod download et pnpm install --frozen-lockfile
scripts/secrets.sh keygen          # une fois : crée votre clé age et affiche la clé publique
# ajoutez cette clé publique à secrets/local-dev.recipients (voir secrets/README.md)
scripts/secrets.sh init local-dev  # une fois : mot de passe de la base, DATABASE_URL et CURSOR_KEY générés
make up-local-dev                  # construit les images, applique les migrations, démarre local-dev
make down-local-dev                # arrête le profil (les volumes sont conservés)
make help                          # liste les cibles ; `make -n <cible>` affiche une recette sans l'exécuter
```

Les secrets sont chiffrés dans Git (SOPS et age, [ADR-0009](docs/decisions/0009-secrets-with-sops-and-age.md)). Vous ne déchiffrez rien à la main : `make up-local-dev` passe par `scripts/compose.sh`, qui déchiffre le fichier du profil en mémoire le temps de la commande, à condition que votre clé soit parmi ses destinataires ([secrets/README.md](secrets/README.md)).

En local, l'application répond sur `http://127.0.0.1:8088` (Traefik). Mailpit écoute sur `http://127.0.0.1:8025` et capture les e-mails quand les variables SMTP portent les valeurs local-dev indiquées dans [.env.example](.env.example).

L'assistant reste éteint tant que `AI_ENABLED` vaut `false` (valeur par défaut). Pour l'allumer, il faut `AI_ENABLED=true`, `AI_API_KEY` et `AI_MODEL` dans le fichier de secrets du profil (`scripts/secrets.sh edit local-dev`) ; aucun modèle n'est choisi par défaut.

L'enregistrement d'un projet vérifie le dépôt auprès de l'API GitHub. Sans jeton, cet appel partage la limite anonyme de 60 requêtes par heure et par adresse IP du serveur ; avec `GITHUB_TOKEN` (un jeton à grain fin en lecture seule sur les dépôts publics, dans le fichier de secrets du profil), elle passe à 5 000. Le jeton est facultatif : s'il est refusé, l'appel repart sans lui.

Le tableau complet des cibles Make est dans [docs/DEV.md](docs/DEV.md).

## Architecture

Un monolithe Go modulaire (Chi, pgx, sqlc) expose une API REST décrite en OpenAPI ([api/openapi.yaml](api/openapi.yaml)). Une application React 19 indépendante (TypeScript, Vite, TanStack Router, Query et Form, Astryx avec la direction Graphite & Lime) la consomme. PostgreSQL sert de source de vérité et de moteur de recherche (plein texte, GIN, `pg_trgm`). L'assistant passe par un port LLM interchangeable, branché sur Google AI Studio ([ADR-0011](docs/decisions/0011-ai-assistant-provider.md), [ADR-0012](docs/decisions/0012-ai-assistant-integration.md)). Docker Compose et Traefik assurent l'exécution, Make et GitHub Actions la CI, et la cible d'hébergement est Datacloud. Pas de Valkey ni de stockage S3 en P0 ([ADR-0006](docs/decisions/0006-toolchain-environments-quality-commands.md)).

Le détail est dans [ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md).

## Qualité

Les gates sont définies dans [TEST.md](docs/TEST.md) :

- **G1** : format, lint, typage, génération OpenAPI et sqlc reproductible.
- **G2** : tests unitaires et couverture. Go à 100,00 % des statements first-party ; web strictement au-dessus de 70 % sur les quatre métriques (statements, branches, functions, lines). Les exclusions vivent seulement dans [COVERAGE_SCOPE.md](docs/quality/COVERAGE_SCOPE.md).
- **G3** : intégration sur un PostgreSQL réel, contrat d'API et invariants.
- **G4** : sécurité (tests d'autorisation négatifs, SAST, dépendances, secrets, licences).
- **G5** (e2e et UX) et **G6** (performance) : NOT IMPLEMENTED. **G7** (livraison, manuelle) : NOT RUN.

`make quality` lance G1 à G4 ; la CI les exécute en groupes parallèles ([ADR-0010](docs/decisions/0010-ci-parallel-gate-groups.md)). Une PR qui ne touche que la documentation n'exécute que le groupe `repo`.

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

Lisez d'abord [CONTRIBUTING.md](CONTRIBUTING.md) et la [page d'accueil de l'équipe](docs/team/KICKOFF.md).

- Une branche `feature|fix|docs|chore/<domaine>/<nom>` part de `develop` et y revient par PR ; `main` ne reçoit que des promotions depuis `develop`, hotfix mis à part ([GIT_WORKFLOW](docs/GIT_WORKFLOW.md)).
- La PR suit le modèle [.github/PULL_REQUEST_TEMPLATE.md](.github/PULL_REQUEST_TEMPLATE.md) et met à jour [TODO.md](TODO.md) dans le même changement.
- Une PR vers `develop` fusionne quand la CI est verte et que la revue IA n'a plus de finding ouvert ; les simples notes ne bloquent jamais ([préférences du propriétaire](docs/team/OWNER_PREFERENCES.md)). Le périmètre, le contrat d'API, les migrations, les ADR, les exceptions de sécurité et les secrets restent des décisions du propriétaire.
- Chaque contribution est signée par la personne qui en répond. Aucun agent IA n'apparaît comme auteur ni en `Co-authored-by`. L'aide d'une IA se déclare dans un fichier de [docs/ai-usage/](docs/ai-usage/README.md), repris dans [AI_USAGE.md](AI_USAGE.md).

## Documentation

| Document | Rôle |
| --- | --- |
| [PRD](docs/PRD.md) | Vision, neuf domaines, priorités |
| [SPEC](docs/SPEC.md) | User stories, scénarios BDD, traçabilité |
| [Architecture](docs/architecture/ARCHITECTURE.md) | Architecture approuvée et frontières techniques |
| [Développement](docs/DEV.md) | Organisation du code, cibles Make, workflow |
| [Tests](docs/TEST.md) | Stratégie de tests, gates et CI |
| [Couverture](docs/quality/COVERAGE_SCOPE.md) | Périmètre de couverture et exclusions justifiées |
| [Livraison](docs/SHIP.md) | Déploiement, vérification, rollback |
| [Décisions](docs/decisions/README.md) | ADR : toute dérogation passe par un nouvel ADR |
| [Design](DESIGN.md) | Direction visuelle, tokens et composants |
| [Contribuer](CONTRIBUTING.md) et [Git](docs/GIT_WORKFLOW.md) | Identité, branches, PR, promotion |
| [Secrets](secrets/README.md) | Clés age, profils chiffrés, accès des membres |

## Originalité, IA et dépendances

226os est un produit original, développé par notre équipe sous licence Apache-2.0. Nous citons nos sources d'inspiration, dont 225 Open Source, sans copier de code, de texte, de visuels ni de données : voir la [déclaration d'originalité](docs/research/REFERENCES.md#déclaration-doriginalité). L'aide d'une IA est déclarée dans [AI_USAGE.md](AI_USAGE.md), les dépendances et leurs licences dans [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md), et l'audit de soumission dans [docs/SHIP.md](docs/SHIP.md).

## Licence

Le code original de 226os est publié sous licence [Apache 2.0](LICENSE) (voir aussi [NOTICE](NOTICE)). Les dépendances tierces conservent leurs licences.
