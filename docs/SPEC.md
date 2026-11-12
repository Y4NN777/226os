# 226os — Functional & Technical Specification

**Version :** 1.2 (US-INTEL-001, US-CONNECT-001, US-LEARN-002 et extension de US-IDENTITY-001, proposées avec l'amendement 3.2 du PRD ; 1.1 : US-LEARN-001, PRD 3.1). **Statut :** contrat initial de développement. **Source produit :** [PRD.md](PRD.md)

## 1. Règles de spécification
- Toute fonctionnalité part d'une user story `US-*`, d'un scénario Given/When/Then, et d'un contrat d'API ou d'interface observables.
- P0, P1, P2 sont des **priorités de livraison**, non des fonctions supprimées. La vision exacte reste dans le PRD.
- La spécification est évolutive par PR et doit précéder le changement de comportement. Un agent n'invente pas de fonctionnalités.
- Les contrats approuvés sont marqués `ACCEPTED` (voir l'[index des ADR](decisions/README.md)) ; toute nouvelle hypothèse non validée doit être signalée comme telle dans la PR qui l'introduit.

## 2. Parcours P0 et scénarios BDD

### US-SOLVE-001 — Poser une question et obtenir une réponse
```gherkin
Feature: Entraide technique
  Scenario: Publier puis résoudre une question
    Given Alice est authentifiée
    When elle publie une question avec titre, description, tags et snippet valide
    Then la question est persistée et visible dans l'exploration
    And Bob peut répondre en étant authentifié
    And Alice peut accepter une seule réponse à sa question
    And un tiers peut retrouver la question via la recherche

  Scenario: Interdire une acceptation non autorisée
    Given Bob n'est pas l'autrice de la question
    When Bob tente d'accepter une réponse via l'API
    Then le serveur refuse avec 403 (l'objet est visible pour Bob)
    And aucune réponse acceptée n'est modifiée
```

### US-DISCOVER-001 — Trouver une API documentée
```gherkin
Feature: Annuaire des ressources
  Scenario: Recherche par pays et catégorie
    Given une fiche API possède un fournisseur, des pays pris en charge et une source
    When un visiteur filtre par pays et catégorie
    Then les fiches correspondant aux filtres sont affichées
    And chaque fiche expose sa provenance documentaire
    And les relations vers les discussions associées sont consultables
```

### US-BUILD-001 — Collaborer sur un projet open source
```gherkin
Feature: Collaboration
  Scenario: Demander à contribuer à un projet
    Given Alice a publié un projet et un besoin de compétence Go
    And Bob est connecté
    When Bob manifeste son intérêt pour ce besoin
    Then Alice peut consulter la candidature
    And une nouvelle demande identique n'est pas créée par un second clic

  Scenario: Empêcher une lecture non autorisée des candidatures
    Given Charlie n'est pas responsable du projet
    When Charlie demande la liste privée des candidatures
    Then le serveur refuse l'accès (403 si le projet est visible pour lui, 404 s'il est privé)
```

### US-LEARN-001 — Apprendre et publier un article (VALIDÉ par le propriétaire le 2026-11-11, PRD amendement 3.1 ; 401 sans session confirmé)
```gherkin
Feature: Apprendre
  Scenario: Lire un tutoriel et suivre la formation gratuite
    Given un article, un tutoriel et une série de formations gratuites à l'IA sont publiés
    When un visiteur non authentifié ouvre Apprendre
    Then il voit la liste avec titre, niveau, langue et tags
    And il peut lire l'article, le tutoriel et la série sans compte
    And les liens vers les questions, API ou projets associés sont visibles et résolus

  Scenario: Publier un article
    Given Alice est authentifiée
    When elle publie un article avec titre, contenu Markdown, tags, niveau et langue valides
    Then l'article est persisté et visible dans Apprendre
    And un tiers peut le retrouver via la recherche

  Scenario: Interdire la publication sans authentification
    Given un visiteur n'est pas authentifié
    When il tente de publier un article via l'API
    Then le serveur refuse avec 401
    And aucun article n'est créé
    And la lecture des articles publiés reste publique

  Scenario: Préserver le contenu original FR/EN
    Given un article est rédigé en français
    When un visiteur choisit l'interface en anglais
    Then les libellés de navigation s'affichent en anglais
    And le texte original de l'article n'est pas altéré
    And sa langue d'origine est indiquée
```

### US-LEARN-002 — Poser une question sur une leçon et suivre un thème (PROPOSÉ, PRD amendement 3.2)

Aucune nouvelle table : la question réutilise `links[]` de la création de question, et la paire question vers article est déjà autorisée. Les fils par thème attendent une PR de contrat en attente, qui type le filtre `topic` comme code de thème ; les questions n'ont pas de thème et sont filtrées par le tag qui porte le même nom que le code du thème.

```gherkin
Feature: Relier les discussions aux leçons
  Scenario: Poser une question depuis une leçon
    Given un article est publié dans Apprendre
    And Alice est authentifiée avec un email vérifié
    When elle choisit « Poser une question sur cette leçon » depuis l'article
    Then le formulaire de question de Solve s'ouvre, prérempli à partir de l'article
    And la question publiée est créée avec un lien vers l'article
    And ce lien est visible et résolu sur la question

  Scenario: Demander une connexion avant de poser la question
    Given un visiteur non authentifié lit un article publié
    When il choisit « Poser une question sur cette leçon »
    Then il est redirigé vers la connexion
    And aucune question n'est créée

  Scenario: Suivre le fil d'un thème
    Given des articles publiés ont le thème ai-agents
    And des questions publiées portent le tag ai-agents
    When un visiteur ouvre le fil du thème ai-agents
    Then il voit les articles publiés de ce thème
    And il voit les questions publiées qui portent le tag ai-agents
    And aucun brouillon ni contenu retiré n'apparaît
    And le fil est en lecture seule
```

### US-IDENTITY-001 — Identité et collaboration inter-domaines
```gherkin
Feature: Profils et liens contextuels
  Scenario: Retrouver une ressource depuis une question
    Given une question est liée à une ressource et à un projet
    When un visiteur consulte la question
    Then les liens vers la ressource et le projet sont visibles et résolus

  Scenario: Changement de langue de l'interface
    Given l'interface existe en français et en anglais
    When un visiteur choisit anglais
    Then les libellés de navigation s'affichent en anglais
    And le texte original des contributions n'est pas altéré
```

**Extension : visibilité du profil (PROPOSÉ, PRD amendement 3.2).** Le profil reçoit un réglage `profile_visibility` (`public`, `members` ou `private`, `public` par défaut), modifiable par son propriétaire via `PATCH /me/profile` une fois la PR de contrat fusionnée. Les modérateurs n'ont pas d'accès particulier aux profils. Le traitement des comptes suspendus par `GET /profiles/{username}` reste une décision ouverte.

```gherkin
Feature: Visibilité du profil
  Scenario: Restreindre son profil aux membres
    Given Alice est authentifiée et n'a jamais modifié la visibilité de son profil
    When elle choisit la visibilité « membres » dans ses paramètres
    Then son profil n'est plus visible des visiteurs non authentifiés
    And il reste visible des membres authentifiés

  Scenario Outline: Lire un profil selon sa visibilité
    Given le profil d'Alice a la visibilité <visibilité>
    When <appelant> demande le profil d'Alice
    Then le serveur répond <statut>
    And la réponse ne contient aucune adresse e-mail

    Examples:
      | visibilité | appelant                        | statut |
      | public     | un visiteur non authentifié     | 200    |
      | members    | un visiteur non authentifié     | 404    |
      | members    | Bob, membre authentifié         | 200    |
      | private    | Bob, membre authentifié         | 404    |
      | private    | un modérateur                   | 404    |
      | private    | Alice elle-même                 | 200    |

  Scenario: Un profil masqué ne se distingue pas d'un profil inconnu
    Given le profil d'Alice est privé
    When Bob demande le profil d'Alice, puis celui d'un nom d'utilisateur qui n'existe pas
    Then les deux réponses sont des 404 au corps identique, à request_id et instance près

  Scenario: La signature d'auteur reste visible
    Given le profil d'Alice est privé
    And Alice a publié une question
    When un visiteur non authentifié lit cette question
    Then il voit le nom d'utilisateur et le nom affiché d'Alice comme autrice
```

### US-CONNECT-001 — Annuaire des membres (PROPOSÉ, PRD amendement 3.2)

Opération en lecture seule, suggérée `GET /profiles`, dont le nom sera fixé par une PR de contrat en attente ; elle n'est pas encore dans `api/openapi.yaml`. Un profil est listé pour un appelant seulement si le compte est actif, si `profile_visibility` l'autorise pour cet appelant (`private` jamais, `members` pour les membres seulement), si `open_to_collab` vaut vrai et si `open_to_collab_visibility` l'autorise pour cet appelant. Un champ masqué compte comme absent, y compris pour les filtres.

```gherkin
Feature: Annuaire des membres
  Background:
    Given les profils suivants existent :
      | membre | compte   | open_to_collab | visibilité de open_to_collab | visibilité du profil |
      | awa    | actif    | vrai           | public                       | public               |
      | koffi  | actif    | vrai           | public                       | members              |
      | sami   | actif    | vrai           | members                      | public               |
      | nadia  | actif    | vrai           | public                       | private              |
      | lina   | actif    | vrai           | private                      | public               |
      | omar   | actif    | faux           | public                       | public               |
      | issa   | suspendu | vrai           | public                       | public               |

  Scenario: Un visiteur ne voit que les profils publics et listables
    When un visiteur non authentifié ouvre l'annuaire
    Then il voit uniquement awa

  Scenario: Un membre voit aussi les profils réservés aux membres
    When Bob, membre authentifié, ouvre l'annuaire
    Then il voit uniquement awa, koffi et sami

  Scenario: N'afficher que les champs visibles, jamais l'e-mail
    Given le pays d'awa est réservé aux membres et ses compétences sont publiques
    When un visiteur non authentifié ouvre l'annuaire
    Then la fiche d'awa montre ses compétences mais pas son pays
    And aucune entrée de l'annuaire ne contient d'adresse e-mail

  Scenario: Un filtre ne révèle pas un champ masqué
    Given le pays d'awa est BF et réservé aux membres
    When un visiteur non authentifié filtre l'annuaire sur le pays BF
    Then awa n'apparaît pas dans les résultats

  Scenario: Un membre peut filtrer sur un champ qui lui est visible
    Given le pays d'awa est BF et réservé aux membres
    When Bob, membre authentifié, filtre l'annuaire sur le pays BF
    Then awa apparaît dans les résultats

  Scenario: Aucun résultat
    When un visiteur filtre l'annuaire sur une compétence qu'aucun profil visible ne déclare
    Then le serveur répond 200 avec une liste vide et sans curseur suivant
    And la page affiche un état vide avec une action pour réinitialiser les filtres

  Scenario: Taille de page bornée
    When un appelant demande l'annuaire avec limit=51
    Then le serveur répond 422

  Scenario: Limiter la collecte en masse
    Given un appelant a atteint la limite de débit de l'annuaire
    When il envoie une requête de plus
    Then le serveur répond 429 avec un en-tête Retry-After

  Scenario: Annoncer honnêtement ce qui n'est pas construit
    Given le mentorat, les opportunités et les espaces communautaires ne sont pas construits
    When un visiteur ouvre la page Connect
    Then ces parcours sont présentés comme « bientôt »
    And aucune mise en relation ni aucun espace fictif n'est affiché
```

### US-INTEL-001 — Assistant IA sourcé (PROPOSÉ, PRD amendement 3.2)

Les opérations de l'assistant ne sont pas encore dans `api/openapi.yaml` : elles arrivent par une PR de contrat en attente, après une décision d'architecture sur les délais, le plafond de coût et le comportement FR/EN (fournisseur : [ADR-0011](decisions/0011-ai-assistant-provider.md)). L'accès des visiteurs non authentifiés est une décision ouverte ; les scénarios ci-dessous utilisent donc un membre. Tant que la restriction géographique du fournisseur n'est pas tranchée par le propriétaire (voir l'ADR-0011), l'assistant reste désactivé par configuration. Ces scénarios et les règles de refus, de citation et de confidentialité forment aussi le jeu d'évaluation de référence, exécuté en CI contre un adaptateur factice ; l'exécution contre le vrai fournisseur est consignée à part et reste NOT RUN tant qu'elle n'a pas eu lieu.

```gherkin
Feature: Assistant sourcé
  Background:
    Given l'assistant est activé dans cet environnement
    And Alice est authentifiée

  Scenario: Répondre à partir des contenus publiés, avec citations
    Given des questions, des articles ou des fiches publiés correspondent à la question d'Alice
    When Alice pose sa question à l'assistant
    Then la réponse cite au moins une source, avec un lien vers chaque contenu cité
    And chaque source citée fait partie des passages retrouvés par la recherche de la plateforme
    And le serveur retire ou rejette toute citation extérieure à ces passages
    And la réponse se termine par une section « Comment vérifier »

  Scenario: Refuser faute de source
    Given aucun contenu publié ne correspond à la question d'Alice
    When Alice pose sa question à l'assistant
    Then l'assistant indique qu'il n'a trouvé aucune source
    And il propose un lien vers la recherche
    And le fournisseur n'est pas appelé

  Scenario: Ne promettre aucune garantie
    Given une fiche d'API publiée et des réponses à son sujet
    When Alice demande si cette API fonctionnera dans son projet
    Then la réponse n'affirme aucune garantie sur l'API
    And elle ne présente aucune solution comme vérifiée
    And elle renvoie aux sources et à la section « Comment vérifier »

  Scenario: Signaler l'incertitude
    Given les sources retrouvées se contredisent ou ne sont pas datées
    When Alice pose sa question à l'assistant
    Then la réponse signale explicitement son incertitude
    And elle indique la date de chaque source quand elle existe

  Scenario: Une réponse de l'IA n'est pas une réponse de la communauté
    When l'assistant répond à Alice
    Then la réponse est étiquetée comme générée par une IA
    And aucune ligne n'est créée dans answers ni dans votes
    And la réponse de l'assistant ne peut être ni votée ni acceptée

  Scenario: N'envoyer aucune donnée personnelle au fournisseur
    Given la question d'Alice contient une adresse e-mail et une clé d'API
    When l'assistant interroge le fournisseur
    Then la requête ne contient que le texte de la question, après masquage, et des contenus publics retrouvés
    And l'adresse e-mail et la clé sont masquées avant l'envoi
    And ni le profil d'Alice, ni l'adresse e-mail de son compte, ni aucun contenu privé ou retiré n'est envoyé
    And la question n'est pas journalisée avec ces données
    And l'interface indique que la question est envoyée au fournisseur

  Scenario: Ignorer une instruction cachée dans un contenu
    Given un contenu publié contient une instruction adressée à l'assistant
    When ce contenu fait partie des passages retrouvés pour la question d'Alice
    Then l'assistant le traite comme une donnée et non comme une instruction
    And la réponse reste sourcée, ou c'est un refus

  Scenario: Fournisseur indisponible ou quota épuisé
    Given le fournisseur est en panne ou son quota est épuisé
    When Alice pose sa question à l'assistant
    Then elle voit le message « assistant indisponible »
    And le reste de la page fonctionne normalement
    And le serveur répond par un problème 503 sans relancer d'appels en boucle

  Scenario: Limite de débit
    Given Alice a atteint la limite de requêtes de l'assistant
    When elle envoie une requête de plus
    Then le serveur répond 429 avec un en-tête Retry-After
```

### US-ACCESS-001 — Réseau limité
```gherkin
Feature: Accès mobile
  Scenario: Réseau lent
    Given un appareil mobile avec débit limité simulé
    When le visiteur ouvre Explore, Solve ou Build
    Then le contenu utile n'exige pas de média lourd
    And les états chargement, vide et erreur sont compréhensibles
```

## 3. Frontières de domaines (tout le produit)
| Domaine | Responsabilité | Dépendances permises |
|---|---|---|
| identity | Auth, sessions, profils et compétences | couche data/shared |
| knowledge / solve | Questions, réponses, votes, contenus, code review | identité via interfaces |
| resources / discover | Fournisseurs, API, SDK, pays, sources | identité, liens contextuels |
| projects / build | Référencement GitHub, projets, besoins, candidatures | identité, liens contextuels |
| knowledge / learn | Articles, tutoriels, formations, séries, tags, niveau, langue (amendement PRD 3.1) | identité, liens contextuels |
| community | Espaces, fil, discussions, réputation | identité |
| career / connect | Mentorat, opportunités, étude ; annuaire des membres en lecture seule (PRD 3.2) | identité |
| language | Traductions UI et contenus | services contractuels |
| discovery / intelligence | Recherche, recommandations, références ; assistant sourcé (PRD 3.2) | vues de lecture contrôlées ; fournisseur d'IA derrière un port |
| access | Performance réseau, UX hors ligne | transversal frontend |

Les modules P1/P2 ne doivent pas être artificiellement matérialisés en code vide. Leur place est décrite dans les contrats et la roadmap ; les dossiers apparaissent à mesure que les fonctionnalités commencent.

## 4. Contrat REST v1 — bornes de conception
- Préfixe `/api/v1`, JSON UTF-8 ; horodatages RFC3339 UTC ; IDs opaques UUID ; erreurs uniformes `application/problem+json` (RFC 9457).
- Pagination par curseur stable (taille maximale documentée), limite de taille des corps, filtres validés strictement, listes ordonnées de façon déterministe.
- Ressources initiales (chemins de `api/openapi.yaml`) : `/auth/*`, `/me`, `/profiles/{username}`, `questions`, `answers`, `votes`, `/learn/articles` et `/learn/series`, `resources`, `projects`, `contribution-needs`, `applications`, `/moderation/*`, `/search`.
- Les verbes HTTP et schémas sont figés dans `api/openapi.yaml` avant développement parallèle ; ce document ne prétend pas qu'ils sont déjà implémentés.
- L'annuaire des membres (US-CONNECT-001) et l'assistant (US-INTEL-001) entrent dans `api/openapi.yaml` par des PR de contrat séparées (PRD 3.2) ; ils n'y figurent pas encore.
- Auth via session cookie sécurisée, protections CSRF pour mutations, autorisation par propriétaire/role **sur serveur**. Codes : 401 non authentifié ; 403 authentifié mais non autorisé sur un objet visible ; 404 si l'existence d'un objet privé ne doit pas fuiter (règle : 403 si l'objet est visible pour l'appelant, 404 sinon). Publier exige un email vérifié (403 `EMAIL_NOT_VERIFIED`).
- Journaliser les tentatives d'accès interdit sans secrets ni données personnelles inutiles.

## 5. Modèle conceptuel et contraintes
| Entité | Invariants à assurer en base |
|---|---|
| user/profile | identifiant unique ; email normalisé unique ; langue et pays validés |
| question/answer | propriétaire obligatoire ; max une réponse acceptée par question ; cycle de publication published/removed ; état de résolution dérivé (open/answered/resolved) |
| vote | unique `(user_id, target_type, target_id)` ; up/down ; auto-vote interdit (403) ; pas de confiance dans les compteurs client |
| resource | source + URL + date de vérification si déclarée ; pays liés |
| project | URL GitHub normalisée ; propriétaire interne ; visibilité public/private |
| contribution_need | projet existant ; état borné ; compétences liées |
| application | unique `(user_id, contribution_need_id)` ; accès restreint ; ré-application seulement après retrait par le candidat |
| content_relation | types source/cible autorisés ; pas de liens pendants |

## 6. PostgreSQL et recherche
- PostgreSQL 18 est l'unique moteur initial ; migrations SQL goose **forward-only** (sans section Down), séquentielles, immuables après merge, avec restauration comme repli ; accès `pgx` + `sqlc` ; transaction pour mutations multi-tables.
- Full-text : `tsvector` alimenté explicitement ou via colonne générée selon vérification SQL ; index GIN ; `websearch_to_tsquery` en configuration `simple` avec `unaccent` (fonction wrapper immuable) ; `pg_trgm` pour tolérance aux fautes ; index et plans vérifiés par `EXPLAIN (ANALYZE, BUFFERS)` sur données représentatives.
- Recherche combinée des `questions`, `resources`, `projects` et des articles/tutoriels Learn via requêtes explicitement bornées ; filtres pays, tags, type et langues ; tri stable ; protection contre requêtes coûteuses.
- Français et anglais : stratégies de tokenisation documentées ; aucune promesse de traduction automatique du contenu au P0.

## 7. Sécurité & performance comme critères fonctionnels
- Aucun endpoint sensible sans tests authorization négatifs, validation, rate-limit approprié, contrôle CSRF et journalisation non sensible.
- Aucun snippet utilisateur n'est exécuté ; HTML dangereux neutralisé ; URLs externes validées ; GitHub API à quotas, timeouts, et restrictions anti-SSRF.
- Budgets initiaux *à mesurer et ajuster après baseline* : endpoint lecture P95 <= 300 ms et mutations P95 <= 500 ms dans l'environnement de test défini ; poids JS et octets transférés surveillés. Ce ne sont pas des promesses de SLA.
- SLO de démonstration : santé API, navigation mobile et flux critiques prouvés sur l'environnement de déploiement.

## 8. Traceability map
Ce tableau est la seule carte de traçabilité au niveau des user stories ; [TRACEABILITY.md](spec-kit/TRACEABILITY.md) ne couvre que les domaines, les expériences et les parcours.

| User story | API / module | Test BDD minimum | Gate |
|---|---|---|---|
| US-SOLVE-001 | questions/answers/votes | publication, acceptation, contrôle d'accès | test/quality/security |
| US-DISCOVER-001 | resources/search | filtres pays, sources, recherche | test/quality |
| US-BUILD-001 | projects/applications | candidature unique, lecture privée interdite | test/security |
| US-LEARN-001 | learn/articles, learn/series | lecture publique, publication authentifiée, 401 sans session, langue d'origine préservée | test/security |
| US-LEARN-002 (proposé, PRD 3.2) | questions (`links[]` vers l'article), learn/articles par code de thème, questions par tag | question liée à l'article, redirection vers la connexion sans session, fil par thème sans brouillon ni contenu retiré | test/e2e |
| US-IDENTITY-001 | sessions/profiles/relations | connexion, langue, liens ; visibilité du profil par appelant, 404 identique à un profil inconnu, signature visible, jamais d'e-mail (extension PRD 3.2) | test/security |
| US-CONNECT-001 (proposé, PRD 3.2) | annuaire, `GET /profiles` suggéré (PR de contrat en attente) | liste par appelant selon les visibilités, filtres sur champs visibles, jamais d'e-mail, état vide, `limit=51` → 422, 429 | test/security |
| US-INTEL-001 (proposé, PRD 3.2) | opérations de l'assistant à définir (PR de contrat en attente), search | citations limitées aux passages retrouvés, refus sans appel au fournisseur, aucune garantie, « Comment vérifier », étiquette IA non votable, aucune donnée personnelle envoyée, injection ignorée, 503, 429 ; jeu d'évaluation de référence | test/security/évaluation |
| US-ACCESS-001 | frontend/infra | mobile + réseau simulé | perf/accessibility |

`api/openapi.yaml` compte aujourd'hui 43 opérations `x-priority: P0`. Le total passe à 44 une fois la PR de contrat de l'annuaire acceptée ; le nombre d'opérations de l'assistant n'est pas encore fixé. Le calendrier de ces PR est suivi dans [ROADMAP.md](../ROADMAP.md).

## 9. Références normatives
`api/openapi.yaml` (accepté), `docs/spec-kit/DATABASE_DESIGN.md` (modèle accepté) et les ADR acceptés (voir l'[index des ADR](decisions/README.md)) sont normatifs. Les migrations `00001` à `00010` de `db/migrations` et le code généré depuis `api/openapi.yaml` existent depuis le bootstrap B0 (PR #24) ; les migrations suivantes, le code applicatif et les seeds sont à produire sous revue. Ce document ne présume aucun statut d'exécution des gates G1 à G7.
