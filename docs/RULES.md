# 226os — Engineering Rules & Non-Negotiables

**Portée :** humains, agents IA et CI. **Priorité :** sécurité/intégrité > PRD > SPEC > ADR acceptés > règles techniques > préférences locales (les permissions, gates et preuves des agents sont dans [AGENTS](../AGENTS.md)). Aucune règle de ce fichier ne permet d'ignorer des contraintes supérieures.

## Produit
1. Préserver les **neuf domaines** du [PRD](PRD.md) ; P0/P1/P2 = ordre, jamais suppression. Pas de nouvelle fonctionnalité implicite.
2. Tout changement public a une user story et un test d'acceptation ; SPEC et OpenAPI suivent le comportement.
3. Stack et frontières : [ARCHITECTURE](architecture/ARCHITECTURE.md), toute dérogation par ADR signé.
4. Ne pas déclarer « done » sans produit déployé, preuve de test et critère d'acceptation.

## Code et données
5. Pas de logique métier dans les handlers ; SQL paramétré ; context avec deadline ; erreurs explicites ([DEV](DEV.md)).
6. OpenAPI 3.1 est le contrat partagé : clients générés ou validés, aucun drift toléré.
7. Migrations SQL traçables, testées, en avant seulement ; contraintes et index documentés.

## Tests et preuves
8. BDD avant code, TDD RED, GREEN, REFACTOR ; refus et régressions testés ([TEST](TEST.md)).
9. Couverture : backend 100 % des statements first-party mesurables, frontend strictement >70 % sur quatre métriques ; [TEST](TEST.md) en est le propriétaire.
10. Les gates G0 à G7 se prouvent par des artefacts du SHA concerné. Ne jamais inventer un test exécuté, une source lue, une mesure ou un audit. Sans preuve : `NOT RUN` ; si une condition bloque : `BLOCKED`.

## Sécurité et vie privée
11. Modèle de menace avant toute surface sensible ; moindre privilège, auth serveur, CSRF, rate-limit, anti-IDOR, XSS/SSRF, erreurs non révélatrices.
12. Aucun secret ni donnée privée dans Git, prompts, logs, captures ou fixtures ; secrets par les mécanismes approuvés ([ADR-0009](decisions/0009-secrets-with-sops-and-age.md)).
13. SAST, SCA, scan de secrets et licences automatisés ; vulnérabilité critique non résolue = blocage de release.
14. IA : sorties non fiables, code revu et testé, usage déclaré dans [AI_USAGE](../AI_USAGE.md).

## Exploitation
15. DB privée, TLS, images immuables, healthchecks, sauvegarde et rollback testés ; release protégée et traçable commit/digest ([SHIP](SHIP.md)).
16. PRs petites ; revue humaine pour choix produit, sécurité, migrations, dépôts tiers et architecture.

## Exceptions
L'agent ne peut pas s'accorder seul une dérogation. Une exception exige `owner`, justification, risque, durée, compensations et décision humaine versionnée. Aucune exception pour divulgation de secret ou preuve falsifiée.
