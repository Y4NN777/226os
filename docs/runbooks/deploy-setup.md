# Mise en place du déploiement

À faire une fois, dans cet ordre. Le workflow est décrit dans [RELEASE_DRAFT](../RELEASE_DRAFT.md) ; la stratégie dans [SHIP](../SHIP.md).

## Qui fait quoi

Rayan (`PANK4SS`) détient le serveur de **staging**. Azael (`NICE-DEV226`) achète le produit Datacloud et détient le serveur de **production** (décision du propriétaire, 2026-11-15). Le propriétaire (`Y4NN777`) détient le dépôt GitHub, les clés et la zone DNS `226os.tech`. Le dépôt est personnel : seul le propriétaire peut poser des secrets ou des variables d'environnement GitHub. Il n'en pose donc que **deux par environnement**, une fois : `DEPLOY_SSH_KEY` et `SOPS_AGE_KEY`. Tout le reste est dans le dépôt, et Rayan et Azael le modifient par PR avec leur droit d'écriture :

- `deploy/staging.conf` et `deploy/production.conf` : `DEPLOY_HOST`, `DEPLOY_USER`, `PUBLIC_HOST` et `DEPLOY_KNOWN_HOSTS` (la ligne de clé d'hôte du serveur, publique). Le workflow lit le fichier dans le commit déployé (`develop` pour staging, `main` pour production), jamais dans une PR, et refuse toute valeur mal formée ou encore à `FILL-ME`.
- Le compte des tests E2E (`E2E_USER`, `E2E_PASSWORD`) n'est pas dans le dépôt : ce sont les deux secrets de l'environnement GitHub `staging-e2e` (section 7).

| Étape | Qui | Remis à l'autre |
|---|---|---|
| 1. Clés age de la CI, valeurs réelles des fichiers chiffrés | propriétaire | — |
| 2. Clé SSH de déploiement (une par serveur) | propriétaire | la clé **publique** de staging à Rayan, celle de production à Azael |
| 3. DNS (enregistrements `A` de staging et de production) | propriétaire | Rayan donne l'IP du serveur de staging, Azael celle du serveur de production |
| 4. Amorçage du serveur de staging | Rayan | rien à transmettre : il met l'IP et la ligne de clé d'hôte affichée à la fin dans `deploy/staging.conf` (étape 5) |
| 4. Achat du produit Datacloud et amorçage du serveur de production | Azael | rien à transmettre : il met l'IP et la ligne de clé d'hôte dans `deploy/production.conf` (étape 5) |
| 5. `deploy/staging.conf` (hôte, ligne de clé d'hôte) | Rayan, par PR après l'amorçage | — |
| 5. `deploy/production.conf` (hôte, ligne de clé d'hôte) | Azael, par PR après l'amorçage | — |
| 6. Les 2 secrets de chaque environnement GitHub | propriétaire (administrateur du dépôt) | — |
| 7. Premier déploiement de staging | propriétaire ; Rayan suit le serveur (`docker compose ps`, `docker compose logs`) | — |
| 8. Environnement `staging-e2e` et compte des tests E2E (2 secrets) | propriétaire, avec un compte de staging vérifié créé par Rayan ou lui | l'identifiant et le mot de passe du compte, jamais par message |
| 9. Copie des sauvegardes hors du serveur | chacun pour le serveur qu'il détient | — |

Aucune clé privée ne circule : Rayan et Azael ne reçoivent que la clé publique de déploiement de leur serveur, le propriétaire ne reçoit que des informations publiques (clé age publique de Rayan). L'IP et la clé d'hôte passent par la PR.

## 1. Clés age de la CI (une par environnement)

> État au 2026-11-15 (décision du propriétaire, faute de temps avant la soumission) : la clé age personnelle du propriétaire sert de `SOPS_AGE_KEY` dans les deux environnements. Une fuite depuis la CI exposerait donc staging et production : à remplacer après la soumission par une clé par environnement, comme ci-dessous.

1. Sur votre machine : `age-keygen -o ci-staging.key` et `age-keygen -o ci-production.key`. Gardez les fichiers hors de Git.
2. Ajoutez la clé **publique** (ligne `age1...`) de chacune à `secrets/staging-test.recipients` et `secrets/production-live.recipients`, puis `scripts/secrets.sh rotate staging-test` et `scripts/secrets.sh rotate production-live`, et ouvrez la PR.
3. Remplissez les valeurs réelles avec `scripts/secrets.sh edit <profil>` : `PUBLIC_HOST`, `ACME_EMAIL`, `APP_BASE_URL=https://<hôte>`, SMTP, `AI_*` si besoin. Sans `PUBLIC_HOST` et `ACME_EMAIL`, `scripts/compose-env-guard.sh` refuse le déploiement.

## 2. Clé SSH de déploiement

`ssh-keygen -t ed25519 -f deploy_key -N "" -C ci-226os` : une clé par serveur. La partie publique va au script d'amorçage (étape 4), la privée dans le secret `DEPLOY_SSH_KEY`.

## 3. DNS

Un enregistrement `A` (et `AAAA` si IPv6) par hôte public, vers l'IP du serveur correspondant : l'hôte de staging, celui de production (`226os.tech`). Traefik obtient les certificats Let's Encrypt (HTTP-01) : les ports 80 et 443 doivent être joignables. Ne lancez pas le premier déploiement avant que le DNS réponde.

## 4. Amorcer chaque serveur (neuf, Ubuntu ou Debian, en root)

```bash
sudo -i
# copier scripts/server-bootstrap.sh sur le serveur, puis :
bash server-bootstrap.sh --ssh-fingerprint "SHA256:..." "ssh-ed25519 AAAA... ci-226os"
```

`--ssh-fingerprint` est l'empreinte (`ssh-keygen -lf` de votre clé publique) de la clé avec laquelle **vous** vous connectez en root : sans elle, ou si elle n'est pas dans `/root/.ssh/authorized_keys`, le mot de passe SSH reste activé (pas de verrouillage possible). Le script lit le port de sshd, le limite dans `ufw`, enregistre la clé CI en `restrict`, vérifie l'empreinte de la clé APT de Docker, installe `age` et contrôle avec `sshd -T` que le mot de passe est bien désactivé.

L'option `--dry-run` montre les étapes sans rien changer. Le script installe Docker et le plugin compose, crée l'utilisateur `deploy` (groupe `docker`, clé CI seule), règle `ufw` (port SSH, 80, 443), active les mises à jour de sécurité, prépare `/opt/226os` et désactive le mot de passe SSH (gardez votre session ouverte et testez une connexion par clé dans un autre terminal). Un serveur est lié au profil de son premier déploiement : staging ne peut pas être déployé sur le serveur de production, ni l'inverse. À la fin il affiche la ligne à mettre dans `DEPLOY_KNOWN_HOSTS` de `deploy/<env>.conf` (section 5).

## 5. Réglages de déploiement et environnements GitHub

**Dans le dépôt** (PR, relue comme le reste) : `deploy/staging.conf` et `deploy/production.conf`, au format `CLE=valeur`, rien d'évalué, ni guillemets ni espaces autour du `=`.

| Clé | Valeur |
|---|---|
| `DEPLOY_HOST` | IP ou nom du serveur, écrit exactement comme dans `DEPLOY_KNOWN_HOSTS` |
| `DEPLOY_USER` | `deploy` |
| `PUBLIC_HOST` | hôte public (staging : `staging.226os.tech`, production : `226os.tech`), identique à celui du fichier chiffré |
| `DEPLOY_KNOWN_HOSTS` | `<DEPLOY_HOST> ssh-ed25519 AAAA...`, la ligne publique affichée par l'amorçage |

Les fichiers livrés contiennent `FILL-ME` pour l'hôte et la clé d'hôte : le job `deploy` échoue avec « fill deploy/staging.conf » (ou `production.conf`) tant qu'il en reste. `scripts/deploy-conf.sh` refuse aussi toute valeur mal formée, une clé inconnue ou en double.

**Dans GitHub** (le propriétaire seul), Réglages, Environments, créer `staging` et `production`, avec **deux secrets chacun** et rien d'autre (ni variable, ni autre secret) :

| | `staging` | `production` |
|---|---|---|
| Branche autorisée | `develop` | `main` |
| Relecteurs requis (une seule approbation par déploiement) | `Y4NN777` ou `PANK4SS` (Rayan) | `Y4NN777` |
| Secret `DEPLOY_SSH_KEY` | clé privée du serveur de staging | clé privée du serveur de production |
| Secret `SOPS_AGE_KEY` | contenu de `ci-staging.key` | contenu de `ci-production.key` |

Supprimez ensuite les fichiers `*.key` locaux (gardez une copie hors ligne des clés age).

Un troisième environnement, `staging-e2e`, porte le compte des tests E2E (section 7) : branche autorisée `develop`, aucun relecteur requis (l'approbation a eu lieu pour le déploiement), et exactement deux secrets, `E2E_USER` et `E2E_PASSWORD`. Il ne contient ni `DEPLOY_SSH_KEY` ni `SOPS_AGE_KEY`.

## 6. Premier déploiement

- `workflow_run` ne part que depuis la copie du workflow sur `develop` ; le déclencheur `push` sur `main` ne fonctionne qu'une fois le fichier arrivé sur `main` (premier passage de `develop` vers `main`, merge commit, sans autre changement).
- Après le premier push d'images, vérifiez dans GitHub, Packages, que les trois paquets `226os-*` sont liés au dépôt (sinon le serveur ne peut pas les tirer : liez-les, ou rendez-les publics).
- Le premier déploiement de staging crée la base, applique les migrations et charge les données de démonstration. Production : jamais de données de démonstration.

## 7. Compte des tests E2E (après le premier déploiement de staging)

Le compte est dans l'environnement GitHub `staging-e2e`, pas dans le dépôt. Les tests E2E tournent dans un job `e2e` à part, après le déploiement de staging : ce job ne reçoit jamais `DEPLOY_SSH_KEY` ni `SOPS_AGE_KEY`, donc le code npm et Playwright (scripts d'installation des dépendances, specs) ne peut atteindre ni le serveur ni les secrets du profil. L'approbation de `staging` reste la seule par déploiement.

1. Sur le site de staging, créez un vrai compte et vérifiez son e-mail.
2. Le propriétaire crée l'environnement `staging-e2e` (Réglages, Environments) : branche autorisée `develop`, aucun relecteur requis.
3. Le propriétaire ajoute les deux secrets, sur une ligne chacun, la valeur étant saisie à l'invite : `gh secret set E2E_USER --env staging-e2e` puis `gh secret set E2E_PASSWORD --env staging-e2e`. **À faire avant le premier déploiement de staging qui exécute les E2E** : sans ces deux secrets, le job `e2e` échoue avec un message clair ; le groupe connecté n'est jamais ignoré. Ces valeurs ne vont ni dans le dépôt, ni sur le serveur.
4. Tant que le script `test:e2e` n'existe pas sur le commit, le job `e2e` affiche « E2E NOT RUN » et les étiquettes `tested-<arbre>` ne sont pas posées. Chaque exécution publie une question et une réponse avec ce compte. Le dépôt est public : en CI, Playwright n'écrit ni trace, ni vidéo, ni capture, ni rapport HTML (ils contiendraient le mot de passe et le cookie), et seul le XML junit est publié. Les étiquettes `tested-<arbre>` ne sont posées que si le job `e2e` a réellement tourné, et chaque empreinte reçoit alors une attestation signée (arbre et commit) : la production vérifie l'attestation avec `gh attestation verify` et refuse une empreinte seulement étiquetée ou attestée pour un autre arbre.

### Ordre réel pour staging (le compte E2E n'existe qu'après un premier déploiement)

`secrets/staging-test.enc.env` ne contient pas encore `E2E_USER` ni `E2E_PASSWORD`. Le premier déploiement de staging va donc jusqu'à la fumée puis échoue à l'étape E2E, et ne pose pas d'étiquette `tested-<arbre>` : c'est attendu. Commandes du propriétaire, dans l'ordre :

1. Sur le serveur de staging, neuf, en root, ports 80 et 443 libres (`ss -tlnp`) :
   ```bash
   bash server-bootstrap.sh --dry-run --ssh-fingerprint "SHA256:<empreinte de votre clé root>" "$(cat deploy_staging.pub)"
   bash server-bootstrap.sh --ssh-fingerprint "SHA256:<empreinte de votre clé root>" "$(cat deploy_staging.pub)"
   ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub   # doit afficher l'empreinte attendue
   cat /etc/ssh/ssh_host_ed25519_key.pub              # la ligne à épingler
   ```
   `deploy_staging.pub` est la clé publique `ci-226os-staging`, dont la privée est le secret `DEPLOY_SSH_KEY` de l'environnement `staging`.
2. DNS : `staging.226os.tech` en `A` vers l'IP du serveur ; vérifier avec `dig +short staging.226os.tech`.
3. PR depuis `develop` (`scripts/worktree.sh new fix/deploy/staging-conf`) : dans `deploy/staging.conf`, `DEPLOY_HOST=staging.226os.tech` et `DEPLOY_KNOWN_HOSTS=staging.226os.tech ssh-ed25519 <clé lue à l'étape 1>`, puis `bash scripts/tests/deploy-conf_test.sh` et `scripts/deploy-conf.sh staging`.
4. Avant la fusion, contrôler le fichier chiffré sans afficher de valeur : `SECRETS_OUT_DIR=$(mktemp -d /dev/shm/s.XXXX) scripts/secrets.sh decrypt staging-test >/dev/null` puis `scripts/compose-env-guard.sh staging-test <ce dossier>/.env.staging-test` (doit afficher OK ; `PUBLIC_HOST=staging.226os.tech`).
5. Fusion : le workflow `Deploy` part sur `develop` après `Quality`. Approuver l'environnement `staging`. Attendu : déploiement, seeds et fumée OK, puis échec de l'étape E2E (« compte manquant »).
6. Sur https://staging.226os.tech, créer le compte de test et vérifier son e-mail, puis `scripts/secrets.sh edit staging-test` pour ajouter `E2E_USER` et `E2E_PASSWORD`, et ouvrir la PR.
7. Après sa fusion, le déploiement suivant de staging exécute les E2E et pose `tested-<arbre>`. La production peut alors partir de `main` (fusion `develop` vers `main` sans autre changement, puis approbation de l'environnement `production`).

## 8. Sauvegardes

Avant chaque migration, `/opt/226os/backups` reçoit un dump chiffré avec age (`<profil>-<date>.dump.age`, pour les clés de `secrets/<profil>.recipients`) et son fichier de preuve ; le dump en clair est supprimé. Les dix derniers sont gardés. **Étape du détenteur de chaque serveur : aucune copie hors du serveur n'existe.** Copiez régulièrement ce dossier ailleurs (par exemple `rsync` depuis votre machine) : un disque perdu perd les sauvegardes avec la base. Déchiffrement : `age -d -i <votre clé age> fichier.dump.age > fichier.dump`, puis `scripts/restore-check.sh`.

## 9. Après coup

- La clé age de la CI est stockée dans GitHub : amendement de l'ADR-0009 du 2026-11-15.
- `ufw` ne filtre pas les ports publiés par Docker : ici seul Traefik publie 80 et 443.
- Retour arrière : voir [SHIP](../SHIP.md) section 5.
