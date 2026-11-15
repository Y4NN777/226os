# Workflow de déploiement (B4, S8)

**Statut : écrit, non encore exécuté.** Le workflow `.github/workflows/deploy.yml` existe depuis la PR `feature/ops/deploy-pipeline`. Il ne fait rien tant que le propriétaire (`Y4NN777`) n'a pas créé les environnements GitHub, les secrets, les variables, le DNS et préparé les deux serveurs : [`runbooks/deploy-setup.md`](runbooks/deploy-setup.md). G7 reste **NOT RUN** jusqu'au premier déploiement réel. (Le nom du fichier est conservé pour les liens existants.)

## Ce que le workflow garantit

1. **Le SHA testé, rien d'autre.** Staging construit les images du SHA de `develop` dont `quality` est vert et les pousse dans GHCR (`ghcr.io/y4nn777/226os-{api,web,migrate}`). Tout est déployé par empreinte (`@sha256:`), jamais par étiquette.
2. **Production ne reconstruit jamais.** Elle déploie les empreintes étiquetées `tested-<arbre>` : l'étiquette n'est posée qu'après un déploiement staging, une fumée et des tests E2E réussis pour cet arbre Git. Un commit de fusion `develop` vers `main` a le même arbre que le sommet de `develop` testé. Sans empreinte testée pour l'arbre, le job échoue avec un message, avant la demande d'approbation.
3. **Une porte humaine pour la production** : l'environnement `production` (relecteur requis, réglé par le propriétaire).
4. **Sauvegarde avant migration.** `scripts/backup.sh` puis `scripts/migrate-up.sh` avec `BACKUP_EVIDENCE` ; une migration sans preuve est refusée.
5. **Fumée après déploiement** (`scripts/smoke.sh`), avec reprises pour laisser le temps au premier certificat ; un échec fait échouer le déploiement.
6. **Sauvegardes chiffrées et serveur lié à un profil.** Le dump d'avant migration est chiffré avec age (clés publiques de `secrets/<profil>.recipients`) ; chaque serveur refuse un autre profil que celui de son premier déploiement ; un échec affiche l'étape, l'état des migrations, la sauvegarde et l'indication de retour arrière. Le jeton du registre n'est utilisé que dans un répertoire temporaire supprimé après le pull.
7. **Secrets** : la clé age de l'environnement (`SOPS_AGE_KEY`) déchiffre `secrets/<profil>.enc.env` dans la RAM du runner ; le fichier est copié par SSH en mode 600, jamais affiché ni archivé. Hôte SSH épinglé (`DEPLOY_KNOWN_HOSTS`), pas de `StrictHostKeyChecking=no`.

## Étapes

| Cible | Déclencheur | Étapes |
|---|---|---|
| staging (`staging-test`) | fin réussie de `Quality` sur un push de `develop` | `plan` (sommet de `develop` seulement) → `build` (build, scan Trivy, push par SHA ; `packages: write`, aucun secret) → `resolve` (empreintes lues dans le registre) → `deploy` (SSH : sauvegarde, migration, `up --wait` par empreinte, `make seed`, fumée) → `e2e` (Playwright contre l'URL de staging) → `promote` (étiquette `tested-<arbre>`) |
| production (`production-live`) | push sur `main` | `plan` → `resolve` (échoue sans empreinte testée) → `deploy` après approbation (SSH : sauvegarde, migration, `up --wait` par empreinte, fumée ; jamais de seeds) |

Concurrence : un déploiement à la fois par environnement, jamais annulé en cours. Les scripts testés hors ligne : `deploy-plan.sh`, `deploy-images.sh`, `deploy-env.sh`, `deploy-remote.sh`, `server-bootstrap.sh` (`scripts/tests/`).

## Limites connues

- Un correctif urgent poussé sur `main` dont l'arbre n'a pas été testé sur staging échoue volontairement : il passe aussi par `develop`.
- Le serveur reçoit un jeton GHCR court (celui du job) pour tirer les images, retiré à la fin ; l'accès du jeton au paquet est à vérifier après le premier push (runbook).
- La clé age vit dans les secrets d'environnement GitHub : c'est une décision du propriétaire du 2026-11-15 qui contredit le point 8 de l'ADR-0009 (« hors périmètre ») ; l'ADR demande un amendement.
- Le retour arrière reste manuel : redéployer l'empreinte précédente ([SHIP](SHIP.md) section 5).
