# 226os — Development Playbook

**Statut :** règles d'exécution approuvées par direction produit · **Voir :** [RULES](RULES.md), [SPEC](SPEC.md), [TEST](TEST.md), [AGENTS](../AGENTS.md)

## 1. Architecture figée
- **Web :** React 19 + TypeScript strict + Vite, TanStack Router / Query / Form, Meta Astryx comme système de composants principal, Tailwind uniquement en complément de mise en page.
- **API :** Go, Chi, `pgx`, `sqlc`; monolithe modulaire, HTTP REST, contrat OpenAPI 3.1.
- **Persistance :** PostgreSQL 18, `tsvector`/GIN/`pg_trgm`/`unaccent`, migrations SQL goose forward-only dans `db/` (ADR-0005).
- **Infra :** Docker multi-stage, Compose (un arbre à profils `local-dev`, `staging-test`, `production-live`, ADR-0006), Traefik HTTPS, GitHub Actions, GHCR.
- **Contraintes :** pas d'ORM généraliste, de moteur de recherche externe, microservices, Redis ou Kubernetes sans ADR approuvée. Le code métier ne doit pas résider dans les handlers HTTP.

## 2. Topologie cible
```text
226os/
├── AGENTS.md
├── README.md
├── {PRD,SPEC,DESIGN,DEV,TEST,SHIP,RULES}.md
├── docs/architecture/ARCHITECTURE.md
├── docs/research/REFERENCES.md
├── docs/decisions/
├── docs/quality/
├── src/
│   ├── web/src/{routes,features,components,lib,i18n}/
│   └── api/{cmd,internal}/
├── api/openapi.yaml
├── db/{migrations,queries,seeds}/
├── infra/{compose,traefik}/
├── .github/workflows/
└── THIRD_PARTY_NOTICES.md
```
Les dossiers `src`, `api`, `db`, `infra` sont **cibles**, à créer lors de l'initialisation du code. Le paquet livré ici constitue les documents de pilotage, pas une application déjà compilable.

## 3. Organisation du travail : vertical slices
1. Décrire le problème et le critère d'acceptation (PRD + SPEC, ID stable).
2. Écrire les scénarios BDD, y compris échec/autorisations et régression.
3. Implémenter un test échouant (TDD RED) reproduisant le comportement.
4. Écrire le code minimal (GREEN), refactorer sans changer l'observable.
5. Vérifier les invariants SQL / permission / sécurité / performance.
6. Relier un écran au véritable contrat API, sans fausses données en production.
7. Joindre preuves (commandes, résultats, coverage, perf) dans la PR.
8. Merger seulement après gates CI et revue humaine.

## 4. Go backend
- Packages par domaine : `domain`, `service`, `repository`, `transport` ; interfaces à la frontière, dépendances orientées vers la logique métier.
- Context avec deadline sur accès DB, requêtes HTTP et intégrations externes ; erreurs typées et wrapping ; SQL paramétré uniquement.
- `go fmt`, `go vet`, `staticcheck`, `golangci-lint` sous versions épinglées ; `go test -race`, tests intégration PG sur Docker.
- Maintenir `sqlc` généré en phase avec `db/queries` et migrations ; test de non-drift.
- Séparer configuration, application et handlers ; éviter la journalisation de secrets.

## 5. React frontend
- TypeScript strict ; validations de formulaire explicites ; client OpenAPI typé, sans duplication manuelle des DTO.
- TanStack Query pour état serveur ; pas de duplication arbitraire dans store global ; clés de cache déterministes et invalidation documentée.
- TanStack Router pour navigation typée ; route-level lazy loading si utile ; erreurs et loaders traités systématiquement.
- Composants accessibles (labels, focus, aria, clavier), contraste, responsive, traduction UI FR/EN sans texte en dur dans les écrans critiques.
- Tests Vitest + Testing Library ; Playwright pour parcours réels ; pas de snapshots comme unique preuve.

## 6. Collaboration à 3–5 personnes
- 4 membres, inscrits à CADEV. Les zones de travail, les aides et les binômes de relecture sont fixés par l'[ADR-0008](decisions/0008-team-work-areas-and-ownership.md) (accepté) ; le détail par tranche est dans `TODO.md` et `docs/team/CONTRIBUTORS.md`. `Y4NN77` garde les décisions d'architecture, les risques et la mise en production.
- Tout ticket a un `owner`, un reviewer distinct si disponible, un domaine, un ID story et une preuve de vérification.
- Branches courtes, petites PRs, merge fréquents ; feature flags ou masquage explicite des fonctions inachevées.
- Les renforts 4–5 peuvent reprendre des chantiers sans changer l'architecture.

## 7. Changement d'architecture
Une modification de stack, contrat, dépendance importante ou frontière métier demande un ADR (`docs/decisions/NNNN-titre.md`) avec contexte, alternatives, décision, coût de migration, impact sécurité/performance et approbation. Refuser toute dérive silencieuse.

## 8. Respect des références et de l'originalité
Étudier le comportement, l'architecture et les tests des repos existants ; documenter URL, licence, commit, observations. Ne pas prendre la traduction de code par une IA pour une œuvre automatiquement indépendante. Déclarer les composants incorporés et les outils d'IA ; vérifier les licences directes/transitives. Voir [REFERENCES](research/REFERENCES.md).
