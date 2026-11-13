# Brouillon du workflow de release (B4, S8)

**Statut : BROUILLON, non actif.** Rien dans ce fichier ne déploie quoi que ce soit : aucun workflow n'existe dans `.github/workflows/` pour la production (ADR-0006, TASK-OPS-001), et `make up-production-live` reste refusé hors d'un contexte de livraison automatisé. Ce document propose la forme du workflow pour relecture par le propriétaire (`Y4NN777`) ; il ne devient un fichier de workflow qu'après sa validation et une PR relue. G7 reste **NOT RUN**.

## Ce que le workflow doit garantir

1. **Le SHA testé, rien d'autre.** On livre l'image construite depuis le SHA de `develop` dont `quality` est vert, par empreinte (`@sha256:`), jamais par étiquette. Pas de reconstruction sur `main` ([SHIP](SHIP.md) §2, [GIT_WORKFLOW](GIT_WORKFLOW.md)).
2. **Une porte humaine.** Un environnement GitHub `production` protégé (relecteur requis : `Y4NN777`) ; le job de déploiement ne démarre pas sans approbation.
3. **Sauvegarde avant migration.** `scripts/backup.sh` produit la preuve `BACKUP_EVIDENCE` exigée par `make migrate-up` ; une migration sans preuve est refusée.
4. **Fumée après déploiement.** `scripts/smoke.sh production-live https://<hôte>` doit passer ; sinon retour arrière.
5. **Retour arrière par empreinte.** Redéployer l'empreinte précédente ; un retour applicatif n'annule pas une migration (plan de restauration, [SHIP](SHIP.md) §5).

## Étapes proposées

| # | Étape | Outil existant | Manque aujourd'hui |
|---|---|---|---|
| 1 | Vérifier que le SHA candidat est le sommet de `develop` et que `quality` y est vert | `gh api` (lecture) | rien |
| 2 | Construire `api`, `web`, `migrate` pour ce SHA ; écrire les empreintes | `scripts/build-images.sh`, `reports/image-digests.txt` | publication dans GHCR (ADR-0006), jeton `packages: write` limité à ce job |
| 3 | Scanner les images | `scripts/image-scan.sh` | rien |
| 4 | Approbation de l'environnement `production` | GitHub | créer l'environnement et ses relecteurs (H4, propriétaire) |
| 5 | Sauvegarde, puis `migrate-up` avec `BACKUP_EVIDENCE` | `scripts/backup.sh`, `scripts/migrate-up.sh` | hôte Datacloud et accès (H6, H7) |
| 6 | Déployer par empreinte | `scripts/compose.sh production-live up -d --wait` avec `CI=true`, `PRODUCTION_UP_CONFIRM=production-live` | canal d'accès à l'hôte (SSH ou OIDC), variables d'image par empreinte |
| 7 | Fumée | `scripts/smoke.sh production-live https://<hôte>` | `PUBLIC_HOST`, DNS (H7) |
| 8 | Journal | ligne dans TODO (SHA, empreintes, approbateur, résultat de la fumée) | rien |

## Squelette (à valider, ne pas activer tel quel)

```yaml
# .github/workflows/release.yml (BROUILLON)
name: Release
on:
  workflow_dispatch:
    inputs:
      sha: { description: "SHA de develop à livrer", required: true }
permissions: { contents: read }
concurrency: { group: release, cancel-in-progress: false }
jobs:
  build:
    runs-on: ubuntu-latest
    permissions: { contents: read, packages: write }
    outputs: { digests: ${{ steps.build.outputs.digests }} }
    steps:
      - uses: actions/checkout@<sha de 40 caractères> # vX.Y.Z
        with: { ref: ${{ inputs.sha }}, persist-credentials: false }
      # 1. vérifier que inputs.sha est le sommet de develop et que `quality` y est vert (échec sinon)
      # 2. make build, puis publication par empreinte dans GHCR
      - id: build
        run: make build
  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: production        # approbation humaine requise
    steps:
      - run: echo "NOT IMPLEMENTED: sauvegarde, migrate-up, déploiement par empreinte, fumée" && exit 1
```

Le job `deploy` échoue volontairement tant qu'il n'est pas écrit : il ne peut pas déployer par accident.

## Questions pour le propriétaire

- **Registre :** GHCR comme le dit l'ADR-0006 ? Les images doivent-elles être publiques ou privées ?
- **Accès à l'hôte :** SSH avec une clé dédiée stockée comme secret de l'environnement `production`, ou OIDC si Datacloud le permet.
- **Qui approuve :** le propriétaire seul, ou un second relecteur (règle de relecture de `main` ouverte).
- **Promotion :** la PR `develop` vers `main` précède-t-elle le déploiement (recommandé, merge commit) ? Le contrôle « même arbre » reste manuel aujourd'hui (`git diff develop main` vide).
- **Délai :** sans compte Datacloud et sans DNS (H6, H7), seules les étapes 1 à 3 et 8 sont exécutables.
