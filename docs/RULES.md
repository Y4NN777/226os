# 226os — Engineering Rules & Non-Negotiables

**Portée :** humains, Claude Code, autres agents et CI. **Priorité :** sécurité/intégrité > PRD > SPEC > ADR acceptés > règles techniques > préférences locales. Aucune règle dans ce fichier ne permet d'ignorer des contraintes supérieures.

## Produit et fidélité
1. Préserver les **neuf domaines** définis dans PRD (amendement 3.1 : Apprendre) ; P0/P1/P2 = ordre, jamais suppression. Pas de nouvelles features implicites.
2. Tout changement public a une user story et un test d'acceptation ; modifier SPEC/OpenAPI si comportement change.
3. Respecter le monorepo, le monolithe Go modulaire et la stack approuvée ; toute dérogation par ADR signé.
4. Ne pas déclarer « done » si le produit déployé, la preuve de test ou un critère d'acceptation manque.

## Code et données
5. Go : `gofmt`, statique, context/deadline, erreurs explicites, SQL paramétré, aucune logique métier dans les handlers.
6. React : TypeScript strict, TanStack Router/Query/Form, composants accessibles, i18n FR/EN, aucun endpoint inventé ou DTO divergent.
7. PostgreSQL : migrations traçables et testées, contraintes et index documentés ; recherche FTS/GIN/pg_trgm, limites requêtes et plans EXPLAIN.
8. OpenAPI 3.1 est le contrat partagé ; les clients sont générés ou validés automatiquement ; pas de drift toléré.

## Test et evidence
9. BDD avant code (AC observables) ; TDD RED -> GREEN -> REFACTOR ; régressions et cas de refus testés.
10. Couverture **backend 100 % statements mesurables first-party**, frontend **strictement >70 % statements/branches/functions/lines**, scopes documentés et pas d'exclusions opportunistes.
11. Des tests d'intégration DB et E2E sont indispensables ; 100 % coverage seul ne suffit pas. Toutes les gates G0–G7 doivent être vérifiées par artefacts produits à partir du SHA concerné.
12. Ne jamais inventer un test exécuté, une source lue, une performance mesurée ou un audit fait. Utiliser `NOT RUN`/`UNKNOWN` si preuve absente.

## Security & privacy by design
13. Modèle de menace avant implémentation de toute surface sensible ; moindre privilège, auth serveur, CSRF, rate-limit, anti-IDOR, XSS/SSRF, validation input, gestion d'erreurs non révélatrices.
14. Aucune clé, credential, donnée privée, token ou mot de passe dans Git, prompt, logs, traces, screenshots ou fixtures partagées. Secrets seulement via mécanismes approuvés.
15. Contrôles automatisés SAST, SCA, secret scan et licence ; vulnérabilité critique non résolue = blocage release.
16. IA : sorties traitées comme non fiables, code revu/testé, usage déclaré. Les droits d'auteur ne disparaissent pas après réécriture ou traduction du code.

## Performance et exploitation
17. Baseline perf avant optimisation ; budgets et régressions mesurés avec dataset et configuration reproductibles. Les réseaux mobiles sont un cas de test obligatoire.
18. Docker Compose/Traefik : DB privée, TLS, non-root quand possible, ressources limitées, images immuables, healthchecks, backups et rollback testés.
19. CI GitHub Actions exécute les gates ; release protégée par approbation et traçabilité commit/digest.
20. PRs petites, revue humaine pour choix produit, sécurité, migrations, dépôts tiers et changement d'architecture.

## Politique d'exception
L'agent ne peut pas s'accorder seul une dérogation. Une exception nécessite `owner`, justification, risque, durée, compensations et décision humaine versionnée. Pas d'exception pour divulgation de secrets ou preuve falsifiée.
