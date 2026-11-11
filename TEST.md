# 226os — Testing Strategy, BDD/TDD & Quality Gates

**Politique :** test-first + preuves reproductibles. Ce document fixe les critères; aucun résultat n'est présumé acquis.

## 1. Règles de couverture approuvées
- **Backend Go : objectif obligatoire = 100 % de couverture des statements instrumentables appartenant au code applicatif first-party.** Mesure CI : `go test -covermode=atomic -coverprofile=coverage.out ./...` puis `go tool cover -func=coverage.out` ; échec si total < 100.00 % (sans arrondi trompeur).
- **Frontend : seuil strictement > 70 %** pour **statements, branches, functions, lines**, calculé par Vitest/V8 (ou Istanbul si ADR), sur le périmètre source applicatif. Configurer des seuils permettant un strict supérieur (par exemple 71 % minimum). Une seule métrique en défaut = échec.
- Les fichiers generated, vendors, migrations SQL, types purs, configurations et mocks peuvent être exclus **seulement par liste versionnée et justifiée** dans `docs/quality/COVERAGE_SCOPE.md`. Ne jamais exclure une branche métier ou un endpoint difficile pour atteindre 100 %.
- Le seuil Go est une mesure de statements, pas une garantie de toutes les branches ni de tous les scénarios. Pour permissions, validation, transactions, auth et invariants critiques, exiger une **matrice de branches positives/négatives** et des tests d'intégration ; une couverture à 100 % sans ces preuves échoue quand même.
- Couverture reportée par commit SHA, rapport HTML/texte, périmètre et exclusions ; évolution contrôlée par PR. Aucun badge « 100 % » sans rapport réel.

## 2. Pyramide des tests
| Niveau | Responsabilité | Technologie / gate |
|---|---|---|
| Unit | domaines, validations, services, fonctions UI | Go `testing`/Vitest — obligatoire |
| Integration | SQL, transactions, permissions, contrats API | PostgreSQL réel éphémère Docker + tests HTTP — obligatoire |
| Contract | API réelle conforme à OpenAPI et client généré | validation schéma / génération diff — obligatoire |
| E2E | Q&A, API Hub, collaboration, connexion, langue | Playwright — parcours P0 obligatoires |
| Security | auth, IDOR, CSRF, XSS, SSRF, dépendances, secrets | tests négatifs + SAST/dependency scan |
| Performance | SQL plans, temps réponse, poids JS, slow-network | baseline et rapport ; gate selon seuil validé |
| Accessibility | clavier, labels, axe, contrastes | audit automatisé + revue manuelle ciblée |

## 3. BDD — specification by example
- Avant tout code, chaque story P0 a un fichier `.feature` ou un scénario Given/When/Then traçable dans `SPEC.md`.
- Le BDD décrit un résultat observable, pas la structure interne ; les scénarios sont exécutables via tests HTTP/UI ou tracés vers ceux-ci.
- Chaque parcours comprend : nominal, données invalides, acteur non autorisé, duplication/idempotence, erreurs DB/réseau, expérience FR/EN et mobile selon pertinence.

## 4. TDD — boucle rigoureuse
- **RED:** une preuve de test échoué pour comportement absent/bug.
- **GREEN:** correctif minimal, résultats enregistrés.
- **REFACTOR:** amélioration interne, tests restent verts, invariants toujours validés.
- IA/agents peuvent proposer les tests mais ne peuvent pas les supprimer/modifier uniquement pour cacher une panne ; changement de l'attendu = justification métier + reviewer humain.

## 5. Tests sécurité obligatoires
- Authentification/expiration/révocation session, rotation si applicable, cookie flags, CSRF pour mutations, rate-limit, validations.
- Tests de propriété par route : auteur vs non-auteur, collaborateur vs tiers, utilisateur anonyme.
- Test XSS sur Markdown et snippets ; aucun code utilisateur exécuté ; SQL injection contre requêtes de recherche ; contrôle SSRF/import GitHub.
- Scan secrets et dépendances à chaque PR ; rapport licence pour dépendances runtime ; traiter toute vulnérabilité élevée/critique exploitable avant release.

## 6. Quality gates (déterministes, bloquants)
| Gate | Condition d'acceptation | Preuve |
|---|---|---|
| G0 Scope | user story, AC et risques documentés | ticket + SPEC/PRD IDs |
| G1 Format/static | format, lint, types, génération propre | logs CI |
| G2 Unit + coverage | Go 100 % mesuré, frontend >70 % ×4 | coverage par SHA |
| G3 Integration + contract | PG réel, API conformes, tests permissions | logs suite + OpenAPI diff |
| G4 Security | scans + tests négatifs passés; pas de secret | rapports horodatés |
| G5 E2E/UX | parcours P0 fonctionnels, FR/EN, responsive, a11y | traces Playwright/captures |
| G6 Perf | budgets validés, baseline et absence de régression non acceptée | rapport + dataset/config |
| G7 Release | SHA, image digest, smoke, backup/rollback | checklist SHIP |

Aucune étape ne peut être sautée tacitement. Exception documentée uniquement par décision humaine datée, avec durée et risque ; **pas d'exception pour secret exposé, vulnérabilité critique avérée ou perte de données probable**.

## 7. Test evidence record
Pour chaque PR : `story`, `commit`, `environment`, `commands`, `result`, `coverage backend/frontend`, `security checks`, `performance measurements`, `reviewer`, `known risks`. Si l'exécution était impossible, écrire **NOT RUN** et bloquer le gate concerné.

## 8. Exit criteria CADEV
Les flux Solve, Discover, Build, profils, recherche, FR/EN, mobile et déploiement sont testés sur l'application réellement déployée. Le CI doit être vert sur le SHA déployé; erreurs non masquées, licences/IA déclarées. Voir [SHIP](SHIP.md).
