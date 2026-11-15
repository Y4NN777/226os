# 226os — Architecture Reference

**Décision acceptée :** React/TanStack + Go + PostgreSQL + Docker Compose/Traefik + GitHub Actions, monorepo, API REST/OpenAPI, monolithe modulaire. Les versions sont verrouillées : voir [ADR-0006](../decisions/0006-toolchain-environments-quality-commands.md), [ADR-0007](../decisions/0007-astryx-version-pinning-and-verification.md) et les fichiers qui les épinglent (`go.mod`, `package.json`, `compose.yaml`).

```mermaid
flowchart LR
  User[Browser mobile / desktop] --> T[Traefik TLS / Routing]
  T --> W[React Vite / TanStack]
  T --> A[Go Chi / REST API]
  A --> M[Modules métier: identity, solve, discover, build, community, connect, language, intelligence, learn]
  M --> Q[pgx + sqlc]
  Q --> P[(PostgreSQL: FTS + GIN + pg_trgm)]
  A -. limited read only .-> GH[GitHub public API]
  A --> SMTP[SMTP server]
  A -. grounded assistance .-> LLM[LLM provider]
  CI[GitHub Actions] --> IMG[Container Registry]
  IMG --> T
```

## Boundaries
- Web ne parle qu'aux API explicites, sans accès DB ; API est propriétaire des invariants et permissions.
- Modules métier n'importent pas directement transport HTTP de leurs voisins ; partage via interfaces explicites et modèles limités.
- PostgreSQL transactionnel est la source de vérité ; FTS dans le même cluster pour MVP ; aucun index secondaire externe.
- Build/Discover/Solve partagent des *relations de contenu*, pas une table géante mélangeant leurs modèles.
- Profils et identité réutilisés ; auth et authorization centralisés, vérification d'accès au plus près des opérations métier.
- Traefik seul exposé publiquement, DB en réseau privé ; secrets via variables sécurisées et volumes maîtrisés.

## Critères de changement
Cette section est la seule référence pour la règle de dérogation. Tout écart à la stack ou toute fragmentation microservices, nouvelle DB ou search externe requiert ADR avec justification mesurée, coût, risques et approbation humaine.

## Couche plateforme de l'API
Les modules métier `identity` (dont l'annuaire Connect), `discover`, `search`, `solve`, `learn`, `build`, `assistant` (intelligence) et `moderation` existent dans `src/api/internal` ; les autres modules du diagramme (community, language) sont **planifiés**. Chaque module suit les couches `domain`, `service`, `repository` et `transport`, sauf `assistant`, qui n'a pas de `repository` : il récupère ses passages par le service de `search` et appelle le modèle par `platform/llm`. Deux packages sont à part : `identity/password` (hachage Argon2id) et `build/github` (client de l'API GitHub, contrôles anti-SSRF). La couche transverse ci-dessous est consommée par ces modules par des interfaces.
- `internal/app` : racine de composition (configuration, journalisation, transport, pool pgx, purge périodique des clés d'idempotence, câblage des modules métier).
- `platform/apigen` (code généré depuis `api/openapi.yaml`) et `platform/apispec` (contrat embarqué pour valider les requêtes) ; une opération non livrée répond 501.
- `platform/httpx` : routeur, middleware, cycle de vie du serveur ; `/healthz`, `/readyz` et l'API sous `/api/v1`.
- `platform/apperr` : erreur applicative typée, rendue en problem+json ; `platform/config` : configuration validée depuis l'environnement.
- `platform/cursor`, `platform/etag`, `platform/idempotency` : conventions de l'[ADR-0004](../decisions/0004-api-contract-conventions.md) (curseurs signés, ETag fort et `If-Match`, `Idempotency-Key`).
- `platform/db` : exécuteur de transactions ; `platform/dbgen` : code sqlc généré ; `platform/pgtest` : tests d'intégration PostgreSQL réel (build tag `integration`).
- `platform/subject` : l'appelant authentifié, ou anonyme par défaut ; la session est résolue par `identity/transport` (`NewResolver`).
- `platform/ratelimit` : limiteur à fenêtre fixe sur la table PostgreSQL `rate_limit_buckets` ([ADR-0003](../decisions/0003-authentication-sessions-csrf-rate-limits.md)), clés en HMAC (ni IP ni identifiant stockés) ; l'appelant choisit d'échouer fermé (authentification, écritures) ou ouvert (lectures anonymes comme la recherche).
- `platform/mailer` : port d'envoi de mail, adaptateur SMTP (authentification seulement sous TLS, STARTTLS exigé) ; Mailpit en `local-dev`. Aucun code métier n'importe `net/smtp`.
- `platform/llm` : port vers un fournisseur de modèle de langage, indépendant du fournisseur ; un adaptateur AI Studio existe. Il est branché sur l'API par le module `assistant` (PR #161, [ADR-0012](../decisions/0012-ai-assistant-integration.md)), désactivé tant que `AI_ENABLED` vaut `false`.
- `cmd/api` : point d'entrée HTTP ; `cmd/aicheck` et `cmd/mailcheck` : vérifications manuelles de la configuration IA et SMTP, avec de vrais appels ; `cmd/healthcheck` : sonde `/healthz` ou `/readyz` embarquée dans l'image (pas de shell dans l'image distroless).

## Threat boundaries
Entrées externes : navigateur, URL GitHub, contenu Markdown, snippets, recherche, candidatures, uploads futurs. Appliquer validation des inputs, contrôle des rôles, anti-XSS, SSRF, CSRF et logs minimaux ; jamais exécuter le code déposé.
