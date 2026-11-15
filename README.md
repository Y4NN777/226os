<p align="center">
  <picture><source media="(prefers-color-scheme: dark)" srcset="docs/brand/226os-logo-white.png"><img src="docs/brand/226os-logo.png" alt="226os" height="72"></picture>
</p>

<p align="center"><strong>Apprendre, se connecter, résoudre, découvrir et construire</strong><br>Entre développeurs africains</p>

226os (« 226 Open Stack », 226 étant l'indicatif téléphonique du Burkina Faso) est une plateforme qui relie l'entraide technique, les ressources utiles aux développeurs africains, les projets open source et l'apprentissage. Le but n'est pas d'ajouter un forum, un annuaire d'API ou un réseau professionnel de plus, mais de raccourcir le chemin entre un problème rencontré, une solution documentée et une contribution que d'autres peuvent réutiliser.

Positionnement : **né au Burkina Faso, ouvert à l'Afrique de l'Ouest, puis à tout le continent.** 226os veut d'abord être la référence des développeurs, ingénieurs et techies burkinabè, puis s'étendre dans la sous-région par des partenariats.

**État** : le MVP tourne sur `develop` et la mise en production se prépare. Plan dans [ROADMAP.md](ROADMAP.md), suivi dans [TODO.md](TODO.md).

## Pourquoi 226os

Beaucoup de solutions techniques africaines existent déjà, mais restent peu connues de ceux qui pourraient s'en servir. 226os veut changer cela :

- **Faire connaître les solutions africaines** à la communauté tech du continent : API de paiement mobile, SMS et USSD, cartographie, outils et SDK conçus pour nos marchés.
- **Donner de la visibilité aux projets open source africains** et aider leurs mainteneurs à trouver des contributeurs.
- **Faire grandir la culture open source en Afrique** : contribuer, documenter, citer ses sources et partager ce qu'on sait, avec des formations ouvertes à tous.

## Fonctionnalités

- **Apprendre** : des articles et des tutoriels écrits par des développeurs, et des formations gratuites lisibles sans compte ; filtre par sujet, publication ouverte aux membres.
- **Résoudre** : posez une question avec votre code ; les autres répondent et votent, et vous marquez la réponse qui vous a débloqué.
- **Découvrir** : les API et outils africains (paiement mobile, SMS et USSD, cartographie…) avec leurs pays, leur catégorie et leurs sources, Afrique de l'Ouest par défaut.
- **Construire** : publiez votre projet open source depuis son dépôt GitHub, dites de quelle aide vous avez besoin et recevez des candidatures.
- **Connect** : trouvez des développeurs par pays, langue et compétence ; chacun choisit ce que son profil montre.
- **Recherche** : une seule recherche plein texte couvre les questions, les ressources, les projets et Apprendre.
- **Modération** : « Signaler une erreur » sur une fiche Découvrir ; les modérateurs masquent ou rétablissent un contenu et traitent les signalements par l'API ([runbook](docs/runbooks/moderation.md)).
- **Assistant sourcé** : il répond uniquement à partir des contenus publiés et cite ses sources ; sans source, il le dit. Désactivé par défaut.

## Ce qui distingue 226os

- **Des intégrations africaines documentées** : chaque API ou SDK est relié aux pays couverts, à ses sources et aux questions qui en parlent.
- **Du problème au projet** : une question renvoie vers la ressource et le projet liés, et un besoin de contribution peut en naître.
- **Des collaborations au-delà des frontières** : on trouve des contributeurs par compétence et par pays, dans une interface en français et en anglais qui ne touche jamais au contenu d'origine.
- **Des connaissances vérifiables** : sources, versions et dates passent avant le nombre de votes.

## Démarrage rapide

Prérequis : Go, Node et pnpm (versions fixées dans `go.mod` et `package.json`), Docker avec Compose, Git, Python 3, `sops` et `age`.

```sh
git clone https://github.com/Y4NN777/226os.git && cd 226os
make doctor                        # vérifie la chaîne d'outils sans rien modifier
make setup                         # installe les dépendances depuis les lockfiles
scripts/secrets.sh keygen          # une fois : crée votre clé age et affiche la clé publique
# ajoutez cette clé publique à secrets/local-dev.recipients
scripts/secrets.sh init local-dev  # une fois : génère les secrets du profil local
make up-local-dev                  # construit les images, applique les migrations, démarre
APP_ENV=local-dev make seed        # facultatif : charge les données de démo
```

En local, l'application répond sur `http://127.0.0.1:8088` et Mailpit sur `http://127.0.0.1:8025`. Les secrets sont expliqués dans [secrets/README.md](secrets/README.md), les cibles Make et l'organisation du code dans [docs/DEV.md](docs/DEV.md).

## Architecture

Un monolithe Go modulaire (Chi, pgx, sqlc) expose une API REST décrite en OpenAPI ([api/openapi.yaml](api/openapi.yaml)), consommée par une application React 19 (TypeScript, Vite, TanStack, Astryx avec la direction Graphite & Lime). PostgreSQL sert à la fois de source de vérité et de moteur de recherche (plein texte, GIN, `pg_trgm`) ; l'assistant passe par un port LLM interchangeable, branché sur Google AI Studio. Docker Compose et Traefik font tourner l'ensemble, Make et GitHub Actions portent la CI, et l'hébergement visé est Datacloud. Le détail est dans [ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md), les décisions dans les [ADR](docs/decisions/README.md).

## Qualité

Chaque pull request passe les mêmes contrôles en CI, et `make quality` les lance en local :

- **Code propre** : formatage, lint, types, et code généré (client d'API, requêtes SQL) à jour.
- **Tests unitaires** : 100 % des instructions couvertes côté Go, plus de 70 % côté web.
- **Intégration** : tests sur une vraie base PostgreSQL, réponses de l'API conformes au contrat OpenAPI, contrôles d'accès testés rôle par rôle.
- **Sécurité** : analyse statique du code, failles connues des dépendances et des images Docker, recherche de secrets, et tests des accès refusés sur chaque route.

Les tests de bout en bout dans le navigateur et les tests de performance ne sont pas encore en place. Le détail est dans [docs/TEST.md](docs/TEST.md).

## Documentation

| Document | Rôle |
| --- | --- |
| [PRD](docs/PRD.md) | Vision, neuf domaines, priorités |
| [SPEC](docs/SPEC.md) | User stories, scénarios BDD, traçabilité |
| [Développement](docs/DEV.md) | Organisation du code, cibles Make |
| [Livraison](docs/SHIP.md) | Déploiement, vérification, rollback |
| [Design](DESIGN.md) | Direction visuelle, tokens et composants |

## Contribuer

Lisez [CONTRIBUTING.md](CONTRIBUTING.md) et le [workflow Git](docs/GIT_WORKFLOW.md) avant votre première branche ; la page d'accueil de l'équipe est [KICKOFF.md](docs/team/KICKOFF.md). La revue est automatisée : une revue IA à chaque push, une correction automatique des findings et une fusion automatique des PR propres ; le propriétaire intervient sur un CRITICAL, un chemin sensible ou la chaîne de revue ([ADR-0014](docs/decisions/0014-ai-review-merges-develop.md)).

## Licence et originalité

Le code original de 226os est publié sous licence [Apache 2.0](LICENSE) (voir [NOTICE](NOTICE)) ; les dépendances tierces gardent leurs licences ([THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)).
226os est un produit original : nous citons nos sources d'inspiration, dont 225 Open Source, sans copier de code, de texte ni de visuels ([déclaration d'originalité](docs/research/REFERENCES.md#déclaration-doriginalité)). L'aide d'une IA est déclarée dans [AI_USAGE.md](AI_USAGE.md).
