# 226os — Development Playbook

**Statut :** règles d'exécution approuvées par direction produit · **Voir :** [RULES](RULES.md), [SPEC](SPEC.md), [TEST](TEST.md), [ARCHITECTURE](architecture/ARCHITECTURE.md), [AGENTS](../AGENTS.md)

## 1. Stack et versions
La stack, les frontières et la règle « tout écart passe par un ADR » sont dans [ARCHITECTURE](architecture/ARCHITECTURE.md). Les versions d'outils sont fixées par [ADR-0006](decisions/0006-toolchain-environments-quality-commands.md) et [ADR-0007](decisions/0007-astryx-version-pinning-and-verification.md), dans `go.mod`, `package.json` et `compose.yaml`.

Node : le plancher `engines.node` (`>=24`, vérifié par `make doctor`) et la version épinglée par `.nvmrc` (26, utilisée pour l'image web) diffèrent volontairement. Le plancher dit ce que le code supporte, `.nvmrc` ce que nous exécutons.

## 2. Topologie du dépôt
```text
226os/
├── src/api/{cmd,internal}/        # API Go : cmd/{api,aicheck,mailcheck,healthcheck}, internal/{app,platform}
├── src/web/src/{app,api,features,i18n,styles,test,theme}/
├── api/openapi.yaml               # contrat, source des types Go et TypeScript générés
├── db/{migrations,queries}/ + db/sqlc.yaml
├── deploy/                        # configuration Traefik
├── compose.yaml                   # profils local-dev, staging-test, production-live
├── Makefile + mk/                 # gates et opérations
├── scripts/                       # appelés par Make et la CI
├── secrets/                       # fichiers chiffrés (SOPS, age), voir secrets/README.md
├── docs/                          # PRD, SPEC, DEV, TEST, SHIP, RULES, architecture, decisions, team
├── .github/workflows/
└── DESIGN.md, AGENTS.md, THIRD_PARTY_NOTICES.md, AI_USAGE.md
```

## 3. Cibles Make
`make help` est la source de vérité ; lisez une recette avec `make -n <cible>` avant de la lancer.

| Cible | Rôle |
| --- | --- |
| `doctor`, `setup` | vérifier la chaîne d'outils ; installer depuis les lockfiles uniquement |
| `quality` | G1 à G4, toutes les étapes, résumé PASS/FAIL par étape ; définition locale de « fini » |
| `generate`, `lint`, `typecheck`, `docs-check`, `test-unit`, `coverage` | G1 et G2 ; `lint-*` et `test-integration-*` sont les sous-cibles des groupes CI |
| `test-integration`, `security`, `licenses` | G3 et G4 ; Docker requis pour les deux premières |
| `ci-groups`, `ci-steps` | groupes CI et étapes attendues par la tâche `quality` ([ADR-0010](decisions/0010-ci-parallel-gate-groups.md)) |
| `smoke PROFILE=<profil> [BASE_URL=https://<hôte>]` | fumée post-déploiement en lecture seule : `/healthz` depuis l'extérieur, redirection HTTPS, `/readyz` non routé au bord, `/readyz` dans le conteneur |
| `build` | images `api`, `web`, `migrate` ; empreintes dans `reports/image-digests.txt` |
| `up-local-dev`, `down-local-dev` | profil local ; `up-staging-test` et `down-staging-test` pour la préproduction |
| `up-production-live` | refusé hors du contexte de livraison documenté |
| `migrate-up`, `migrate-status`, `migrate-new NAME=<slug>` | migrations goose, en avant seulement ; confirmation et preuve de sauvegarde hors `local-dev` |
| `db-reset` | recrée la base, `APP_ENV=local-dev` uniquement |
| `review`, `review-login` | revue IA locale et consultative ([open-code-review](https://github.com/alibaba/open-code-review), version fixée dans `scripts/review.sh`, lancée par `npx`) ; envoie le diff au fournisseur LLM configuré ; pas un gate. Sans clé : la commande `/delegate-review` de Claude Code |
| `test-e2e` (G5), `perf` (G6), `seed` | **NOT IMPLEMENTED** ; `quality-full` échoue tant que G5 et G6 le sont |

### Revue IA consultative (optionnelle)

La revue humaine de l'[ADR-0008](decisions/0008-team-work-areas-and-ownership.md) reste obligatoire ; la revue IA n'est qu'une aide, jamais un gate. Règles du projet : `.opencodereview/rule.json`.

- **Triage :** la revue classe le changement en CRITICAL, HIGH, MEDIUM ou NONE avec son périmètre d'impact (cette PR, une tranche, le travail des autres, toute l'équipe, la production), sur six axes (architecture, règles, planning et délai, sécurité, performance, cohérence). Elle ne mentionne le propriétaire que pour CRITICAL et ne bloque jamais.
- **Review gate (un check requis, jamais pour attendre une relecture) :** `scripts/review-gate.py`, exécuté depuis la branche de base. Il ne bloque que (1) un verdict CRITICAL avec preuve OBSERVED (`fichier:ligne`) que le propriétaire n'a pas levé avec le label `owner-override` (posé par lui seul), ou (2) une PR qui touche un chemin sensible (`.github/sensitive-paths.txt`, lu sur la branche de base) sans l'approbation du propriétaire (ses propres PR passent). Il ne bloque jamais si la revue n'a pas tourné ou si GitHub est illisible. Mode : variable `CODE_REVIEW_GATE` = `off`, `report` (défaut, ne bloque rien) ou `enforce`.
- **Avec Claude Code, sans clé d'API (recommandé) :** `/delegate-review` dans Claude Code (fichier `.claude/commands/delegate-review.md`). open-code-review choisit les fichiers et fournit les règles, Claude Code fait la revue avec votre connexion. Il propose des corrections et ne modifie rien.
- **Avec l'outil seul :** `make review-login` une fois (connexion ChatGPT, aperçu réservé à un usage local par une seule personne), puis `make review`. Le diff part chez le fournisseur choisi : seulement du code que vous avez le droit d'envoyer.
- **En CI :** le workflow `.github/workflows/code-review.yml` est désactivé tant que `CODE_REVIEW_ENABLED` n'est pas à `true`. Claude Code (action officielle d'Anthropic, jeton OAuth `CLAUDE_CODE_OAUTH_TOKEN` issu de `claude setup-token`) fait la revue en suivant `/delegate-review` ; consultatif, jamais un gate. Vérifiez les conditions d'Anthropic avant d'y mettre un jeton d'abonnement. La connexion ChatGPT ne doit pas servir en CI partagée : l'outil l'interdit.

## 4. Organisation du travail : vertical slices
1. Décrire le problème et le critère d'acceptation (PRD + SPEC, ID stable).
2. Écrire les scénarios BDD, y compris échec, autorisations et régression.
3. Implémenter un test échouant (RED), puis le code minimal (GREEN), puis refactorer.
4. Vérifier les invariants SQL, permissions, sécurité, performance.
5. Relier un écran au vrai contrat API, sans fausses données en production.
6. Joindre les preuves (commandes, résultats) dans la PR ; merger après gates CI et revue humaine.

## 5. Go et React
- **Go :** packages `domain`, `service`, `repository`, `transport` ; context avec deadline sur tout accès externe ; erreurs typées ; SQL paramétré ; `sqlc` sans drift (vérifié par `make generate`).
- **React :** TypeScript strict ; client OpenAPI typé, sans DTO manuel ; TanStack Query pour l'état serveur ; composants accessibles ; i18n FR/EN.
- **Tests :** Vitest et Testing Library côté web. Playwright (G5, parcours réels) : **NOT IMPLEMENTED**, voir [TEST](TEST.md).

## 6. Équipe
Zones de travail, binômes de relecture et rôles : [ADR-0008](decisions/0008-team-work-areas-and-ownership.md) et [docs/team/CONTRIBUTORS.md](team/CONTRIBUTORS.md). Le propriétaire est `Y4NN777` (login GitHub).

Tout ticket a un `owner`, un reviewer distinct si possible, un domaine, un ID de story et une preuve de vérification. Petites PRs, merges fréquents, fonctions inachevées masquées.

## 7. Originalité et dépendances
Étudier le comportement, pas copier le code ; documenter URL, licence, commit et observations dans [REFERENCES](research/REFERENCES.md). Les dépendances et leurs licences sont dans [THIRD_PARTY_NOTICES](../THIRD_PARTY_NOTICES.md), l'aide d'une IA dans [AI_USAGE](../AI_USAGE.md).
