# 226os — Functional & Technical Specification

**Version :** 1.1 (US-LEARN-001, PRD 3.1) · **Statut :** contrat initial de développement · **Source produit :** [PRD.md](PRD.md)

## 1. Règles de spécification
- Toute fonctionnalité part d'une user story `US-*`, d'un scénario Given/When/Then, et d'un contrat d'API ou d'interface observables.
- P0, P1, P2 sont des **priorités de livraison**, non des fonctions supprimées. La vision exacte reste dans le PRD.
- La spécification est évolutive par PR et doit précéder le changement de comportement. Un agent n'invente pas de fonctionnalités.
- Les hypothèses non validées sont marquées `OPEN` ; les contrats approuvés sont marqués `ACCEPTED`.

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
    Then le serveur refuse avec 403
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
    Then le serveur refuse l'accès
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
| career / connect | Mentorat, opportunités, étude | identité |
| language | Traductions UI et contenus | services contractuels |
| discovery / intelligence | Recherche, recommandations, références | vues de lecture contrôlées |
| access | Performance réseau, UX hors ligne | transversal frontend |

Les modules P1/P2 ne doivent pas être artificiellement matérialisés en code vide. Leur place est décrite dans les contrats et la roadmap ; les dossiers apparaissent à mesure que les fonctionnalités commencent.

## 4. Contrat REST v1 — bornes de conception
- Préfixe `/api/v1`, JSON UTF-8 ; horodatages RFC3339 UTC ; IDs opaques UUID ; erreurs uniformes `application/problem+json` (RFC 9457).
- Pagination par curseur stable (taille maximale documentée), limite de taille des corps, filtres validés strictement, listes ordonnées de façon déterministe.
- Ressources initiales : `sessions`, `users/me`, `profiles`, `questions`, `answers`, `votes`, `articles` (`OPEN` : nom et schéma à fixer dans l'OpenAPI), `resources`, `projects`, `contribution-needs`, `applications`, `search`.
- Les verbes HTTP et schémas sont figés dans `api/openapi.yaml` avant développement parallèle ; ce document ne prétend pas qu'ils sont déjà implémentés.
- Auth via session cookie sécurisée, protections CSRF pour mutations, autorisation par propriétaire/role **sur serveur**.
- Journaliser les tentatives d'accès interdit sans secrets ni données personnelles inutiles.

## 5. Modèle conceptuel et contraintes
| Entité | Invariants à assurer en base |
|---|---|
| user/profile | identifiant unique ; email normalisé unique ; langue et pays validés |
| question/answer | propriétaire obligatoire ; max une réponse acceptée par question ; état contrôlé |
| vote | unique `(user_id, target_type, target_id)` ; pas de confiance dans les compteurs client |
| resource | source + URL + date de vérification si déclarée ; pays liés |
| project | URL GitHub normalisée ; propriétaire interne ; visibilité explicite |
| contribution_need | projet existant ; état borné ; compétences liées |
| application | unique `(user_id, contribution_need_id)` ; accès restreint |
| content_relation | types source/cible autorisés ; pas de liens pendants |

## 6. PostgreSQL et recherche
- PostgreSQL est l'unique moteur initial ; migrations SQL up/down contrôlées ; accès `pgx` + `sqlc` ; transaction pour mutations multi-tables.
- Full-text : `tsvector` alimenté explicitement ou via colonne générée selon vérification SQL ; index GIN ; `websearch_to_tsquery` avec dictionnaires adaptés ; `pg_trgm` pour tolérance aux fautes ; index et plans vérifiés par `EXPLAIN (ANALYZE, BUFFERS)` sur données représentatives.
- Recherche combinée des `questions`, `resources`, `projects` via requêtes explicitement bornées ; filtres pays, tags, type et langues ; tri stable ; protection contre requêtes coûteuses.
- Français et anglais : stratégies de tokenisation documentées ; aucune promesse de traduction automatique du contenu au P0.

## 7. Sécurité & performance comme critères fonctionnels
- Aucun endpoint sensible sans tests authorization négatifs, validation, rate-limit approprié, contrôle CSRF et journalisation non sensible.
- Aucun snippet utilisateur n'est exécuté ; HTML dangereux neutralisé ; URLs externes validées ; GitHub API à quotas, timeouts, et restrictions anti-SSRF.
- Budgets initiaux *à mesurer et ajuster après baseline* : endpoint lecture P95 <= 300 ms et mutations P95 <= 500 ms dans l'environnement de test défini ; poids JS et octets transférés surveillés. Ce ne sont pas des promesses de SLA.
- SLO de démonstration : santé API, navigation mobile et flux critiques prouvés sur l'environnement de déploiement.

## 8. Traceability map
| User story | API / module | Test BDD minimum | Gate |
|---|---|---|---|
| US-SOLVE-001 | questions/answers/votes | publication, acceptation, contrôle d'accès | test/quality/security |
| US-DISCOVER-001 | resources/search | filtres pays, sources, recherche | test/quality |
| US-BUILD-001 | projects/applications | candidature unique, lecture privée interdite | test/security |
| US-LEARN-001 | articles/tutorials (`OPEN`) | lecture publique, publication authentifiée, 401 sans session, langue d'origine préservée | test/security |
| US-IDENTITY-001 | sessions/profiles/relations | connexion, langue, liens | test/security |
| US-ACCESS-001 | frontend/infra | mobile + réseau simulé | perf/accessibility |

## 9. À compléter avant développement concurrent
La référence normative détaillée `api/openapi.yaml`, les migrations `db/migrations`, les schémas JSON/DTO, les cas d'erreur et la stratégie de seed doivent être produits sous revue. Un agent ne doit pas inventer leur statut de validation.
