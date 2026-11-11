**De la connaissance au logiciel : construire ensemble la tech africaine.**

226os est une plateforme pensée pour relier l'entraide technique, les ressources utiles aux développeurs africains, les projets open source et les personnes qui les font avancer. L'objectif n'est pas de créer un forum, un annuaire d'API ou un réseau professionnel supplémentaire. C'est de rendre plus direct le passage d'un problème rencontré à une solution documentée, puis à une contribution réutilisable.

il s'adresse aux développeurs de toute l'Afrique**, quels que soient leur pays, leur langue ou leur niveau d'expérience.

> **État du projet :** cadrage et initialisation du dépôt. Les fonctionnalités présentées ci-dessous constituent la vision et le périmètre envisagés ; elles ne sont pas annoncées comme déjà livrées.
>
> **Domaine prévu :** [226os.tech](https://226os.tech) 
> **Équipe :** 7's — trois membres actuellement, un ou deux renforts envisagés.  
> **Premier contexte de livraison :** CADEV 2026.

## Pourquoi 226os ?

Intégrer un service de paiement, retrouver un exemple fiable, comprendre une erreur de déploiement ou rechercher des contributeurs implique souvent de passer d'un groupe de discussion à une documentation incomplète, puis à plusieurs dépôts GitHub. Une solution trouvée disparaît parfois dans un fil de messages ; un projet intéressant peine à rencontrer les bonnes compétences.

Les plateformes existantes répondent chacune à une partie du problème. Ce qui manque souvent, c'est la **continuité du contexte** : les services disponibles selon les pays, les contraintes techniques réellement rencontrées, les solutions vérifiables et les personnes qui peuvent faire progresser un projet.

226os veut relier ces éléments plutôt que les juxtaposer.

## Un parcours continu

**Discover → Solve → Build → Share back**

Prenons un développeur qui intègre un service de paiement au Burkina Faso :

1. **Discover :** il identifie les fournisseurs, les pays couverts, les SDK, les exemples et les limites documentées.
2. **Solve :** il pose une question avec son contexte technique, partage son code, échange avec la communauté et documente la résolution.
3. **Build :** il découvre un projet open source lié à son besoin, propose une contribution ou recherche d'autres développeurs.
4. **Share back :** le correctif, l'exemple ou le guide devient une ressource que d'autres peuvent retrouver et évaluer.

C'est cette boucle de contribution qui constitue le cœur de 226os.

## Les huit domaines du produit

| Domaine | Fonctionnalités envisagées |
| --- | --- |
| **Solve — Entraide et connaissances** | Questions/réponses, snippets, solutions acceptées, votes et réputation, revue de code, articles, tutoriels et discussions thématiques. |
| **Discover — Ressources pour développeurs africains** | Catalogue d'API, SDK et outils, couverture par pays, documentation, comparaisons, exemples d'intégration et retours communautaires. |
| **Build — Open source et collaboration** | Découverte et import de projets GitHub, publication de projets, besoins en compétences, contributeurs et cofondateurs, candidatures et reconnaissance des contributions. |
| **Connect — Mentorat et carrière** | Mentorat, relecture de CV et de portfolios, préparation aux entretiens, groupes d'étude, stages et opportunités professionnelles. |
| **Identity & Community** | Profils développeurs, compétences, contributions, réputation, communautés, activités et découverte de personnes ou d'équipes. |
| **Language — Collaboration panafricaine** | Interface français/anglais, publications multilingues, recherche bilingue, traduction assistée avec préservation des textes originaux. |
| **Access — Connexions limitées** | Interface mobile-first, chargement progressif, réduction des transferts, cache, brouillons hors ligne et dégradation contrôlée. |
| **Intelligence — Assistance contextualisée** | Questions similaires, ressources et projets liés, aide à la rédaction, résumés et recommandations assistées par IA avec sources explicites. |

Ces domaines définissent la **vision complète**, et non une promesse de disponibilité immédiate.

### Ce qui doit nous distinguer

**Le contexte des intégrations africaines.** Une fiche d'API doit indiquer les pays couverts, les options techniques disponibles, les limites documentées et les échanges qui étayent ces informations.

**Du problème à la contribution.** Une bonne réponse peut alimenter un guide, une issue GitHub, un correctif ou un nouveau projet collaboratif, sans perdre son contexte.

**La collaboration au-delà des frontières.** Compétences et besoins guident la découverte de contenus et de contributeurs, au-delà de la langue et du pays.

**La connaissance technique appuyée par des preuves.** Versions, étapes de reproduction, sources documentaires et confirmations communautaires doivent permettre de juger une solution autrement qu'au nombre de votes.

## Première livraison : CADEV 2026

Le premier périmètre s'organise autour de trois parcours fonctionnels, appuyés par des profils, la recherche, une interface accessible et une navigation bilingue :

- **Solve :** publier une question technique avec code et contexte, répondre et identifier une résolution.
- **Discover :** rechercher une API ou une ressource pertinente selon le pays et consulter une fiche documentée.
- **Build :** publier ou référencer un projet open source, annoncer des besoins en compétences et manifester son intérêt pour contribuer.

**La collaboration open source fait partie du premier périmètre**, et ne constitue pas une option repoussée à plus tard. Les autres domaines restent dans la vision du produit ; leur profondeur de livraison dépendra des capacités réelles de l'équipe et des contrôles qualité.

La participation à CADEV implique un projet original, un dépôt Git et une démonstration utilisable en ligne, déployée avec un service Datacloud éligible. Les dépendances, leurs licences et l'assistance IA doivent être déclarées. Réécrire ou traduire le code d'un tiers ne suffit pas à faire disparaître les obligations de droit d'auteur et d'originalité.

## Architecture et méthode d'ingénierie

226os adopte un **monolithe Go modulaire** exposant une API REST/OpenAPI, une application React indépendante et PostgreSQL comme source de vérité et moteur de recherche initial. Les composants d'infrastructure supplémentaires sont introduits en réponse à des besoins justifiés, pas par défaut.

| Couche | Choix et orientations |
| --- | --- |
| **Frontend** | React 19, TypeScript, Vite, TanStack Router / Query / Form, Meta Astryx comme système de composants principal, Tailwind CSS uniquement en complément de mise en page. |
| **Backend** | Go, Chi, REST/OpenAPI, pgx, sqlc ; domaines métier explicitement séparés. |
| **Base et recherche** | PostgreSQL, recherche full-text, index GIN, `pg_trgm`, migrations SQL versionnées. |
| **Cache** | Valkey, compatible Redis et sous BSD-3, uniquement pour cache, limitation de débit ou coordination justifiés. |
| **Objets** | Interface S3-compatible ; service Datacloud à évaluer en production et SeaweedFS à étudier pour l'auto-hébergement. |
| **Infrastructure** | Docker, Docker Compose, Traefik et HTTPS. |
| **Automatisation** | Make comme interface stable des tâches locales ; GitHub Actions pour CI/CD et contrôles de livraison. |

Meta Astryx est le système de composants principal approuvé par [DESIGN.md](DESIGN.md). Sa version, ses APIs installées et son intégration restent à vérifier avant implémentation.

La méthode de développement repose sur des objectifs explicites, des contrats BDD, le cycle TDD, des tests d'intégration, des vérifications de sécurité, des preuves de performance et des **gates déterministes** avant livraison. Les objectifs de couverture documentés sont **100 % des statements backend Go mesurables** et **plus de 70 % pour les métriques frontend définies** ; la couverture ne remplace ni les tests d'autorisation ni la vérification des parcours complets.

### Organisation envisagée du dépôt

```text
226os/
├── apps/
│   ├── web/                   # React / TanStack
│   └── api/                   # Monolithe Go modulaire
├── db/                        # Migrations et requêtes SQL
├── infra/                     # Compose, Traefik, cache et stockage objet
├── docs/
│   ├── PRD.md
│   ├── SPEC.md
│   ├── DESIGN.md
│   ├── architecture/ARCHITECTURE.md
│   ├── DEV.md
│   ├── TEST.md
│   ├── SHIP.md
│   ├── RULES.md
│   └── research/REFERENCES.md
├── Makefile                   # Prévu ; commandes à implémenter et vérifier
├── .github/                   # Instructions Copilot et workflows de CI/CD
├── AGENTS.md
├── CONTRIBUTING.md
├── AI_USAGE.md
└── THIRD_PARTY_NOTICES.md
```

**Il n'existe pas encore de procédure d'installation en une commande vérifiée.** Les commandes `make bootstrap`, `make up`, `make test`, `make quality` et `make down` seront documentées comme fonctionnelles seulement après implémentation et exécution réelle.

## Références étudiées

226os s'inspire d'approches existantes pour comprendre les bons compromis, sans présenter leurs implémentations comme du code original :

- **[225 Open Source](https://github.com/kouame09/225_os)** — découverte de projets, lancements, pitches, articles et visibilité des développeurs. Une référence régionale qui montre qu'un annuaire seul ne suffirait pas à nous différencier.
- **[Apache Answer](https://github.com/apache/answer)** — structure des questions/réponses, recherche, réputation, historique et extensibilité.
- **[Flarum](https://github.com/flarum/framework)** — discussions, modération, permissions et expérience communautaire adaptable.
- **[Nuxt Hacker News](https://github.com/nuxt/hackernews)** — navigation sobre, lecture de contenus et attention portée aux performances.
- **[Meta Astryx](https://github.com/meta/astryx)** — système de design React 19+ retenu comme fondation principale, sous réserve de vérification de la version et des APIs installées. shadcn/ui reste une référence comparative, pas une dépendance approuvée.

Ces projets sont **des références de recherche, pas des dépendances déclarées**. Les analyses, attributions et décisions de réutilisation appartiennent à [`docs/research/REFERENCES.md`](docs/research/REFERENCES.md).

## Documentation

| Document | Rôle |
| --- | --- |
| [PRD](PRD.md) | Vision, utilisateurs, huit domaines, besoins et priorités |
| [SPEC](SPEC.md) | Comportements, critères d'acceptation et contrats |
| [Design](DESIGN.md) | Direction visuelle, tokens, composants et validation de la palette |
| [Architecture](docs/architecture/ARCHITECTURE.md) | Architecture approuvée et frontières techniques |
| [Development](DEV.md) | Conventions et workflow de développement |
| [Testing](TEST.md) | Stratégie de tests et quality gates |
| [Shipping](SHIP.md) | Déploiement, vérification et rollback |
| [Rules](RULES.md) | Contraintes d'ingénierie |
| [Agent instructions](AGENTS.md) | Orchestration par objectifs, permissions et preuves |

Le projet est porté par **7's**, actuellement trois membres avec un ou deux renforts envisagés. L'affectation des responsabilités dépendra des compétences et des besoins de livraison.

## Licence

La licence du **code original de 226os** n'est pas encore définie. Aucune autorisation de réutilisation illimitée ne doit donc être présumée. Les dépendances tierces conservent leurs licences et obligations de notification.

---

**226os prend racine au Burkina Faso. Son terrain de collaboration dépasse les frontières.**
