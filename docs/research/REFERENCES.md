# 226os — Références externes à étudier

**Usage :** comparaison de comportements, UX, architecture, licences et tests. Ne pas forker une application complète et prétendre qu'une réécriture IA efface les droits. Déclarer toute dépendance intégrée et vérifier ses licences transitives à version épinglée.

| Projet | Référence | But de l'étude | Licence du dépôt (à revérifier sur commit choisi) |
|---|---|---|---|
| Apache Answer | https://github.com/apache/answer | Solve : Q&A, votes, accepted answer, réputation, workflow | Apache-2.0 |
| Apache Answer Plugins | https://github.com/apache/answer-plugins | extensibilité | par sous-repo |
| Flarum Framework | https://github.com/flarum/framework | discussion, profils, permissions, UX | MIT |
| Nuxt HackerNews | https://github.com/nuxt/hackernews | navigation rapide, feed sobre | vérifier |
| 225 Open Source | https://www.225os.com/ | découverte des projets et positionnement | site tiers (non source à réutiliser) |
| TanStack Router | https://github.com/TanStack/router | route typée | MIT |
| TanStack Query | https://github.com/TanStack/query | synchronisation / cache serveur | MIT |
| TanStack Form | https://github.com/TanStack/form | validations formulaire | vérifier |
| shadcn/ui | https://github.com/shadcn-ui/ui | UI accessible et composable | vérifier |
| Chi | https://github.com/go-chi/chi | API/middlewares Go | vérifier |
| pgx | https://github.com/jackc/pgx | PostgreSQL Go | vérifier |
| sqlc | https://github.com/sqlc-dev/sqlc | SQL Go typé | vérifier |
| Traefik | https://github.com/traefik/traefik | TLS/edge/routing | vérifier |
| GitHub Actions | https://github.com/actions/starter-workflows | CI workflows | workflows à adapter |

## Format obligatoire d'une étude
```yaml
repository: "https://github.com/owner/repo"
commit_sha: "..."
license_file: "..."
license_spdx: "..."
features_examined: []
source_files_and_lines: []
tradeoffs: []
behaviors_for_226os: []
third_party_code_integrated: false
compliance_review: PENDING
```
Chaque proposition d'amélioration issue d'une référence inclut : (1) le problème **observé**, (2) ses sources et références précises, (3) une expérience ou métrique capable de prouver l'amélioration.

## Risques CADEV
Le concours exige une création pour l'essentiel originale pendant la période. Les composants sous GPL/AGPL/LGPL sont prohibés par son règlement tel que communiqué. L'utilisation d'un projet complet comme base, même réécrit en Go/TS, nécessite vérification d'admissibilité et de titularité. Les crédits et obligations de licence restent applicables.
