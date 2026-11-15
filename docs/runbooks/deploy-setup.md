# Mise en place du déploiement

À faire une fois, dans cet ordre. Le workflow est décrit dans [RELEASE_DRAFT](../RELEASE_DRAFT.md) ; la stratégie dans [SHIP](../SHIP.md).

## Qui fait quoi

Rayan (`PANK4SS`) détient le serveur de **staging**. Le propriétaire (`Y4NN777`) détient le serveur de production, le dépôt GitHub, les clés et la zone DNS `226os.tech`.

| Étape | Qui | Remis à l'autre |
|---|---|---|
| 1. Clés age de la CI, valeurs réelles des fichiers chiffrés | propriétaire | — |
| 2. Clé SSH de déploiement (une par serveur) | propriétaire | la clé **publique** de staging à Rayan |
| 3. DNS (enregistrements `A` de staging et de production) | propriétaire | Rayan donne l'IP du serveur de staging |
| 4. Amorçage du serveur de staging | Rayan | au propriétaire : l'IP ou le nom du serveur et la ligne `DEPLOY_KNOWN_HOSTS` affichée à la fin |
| 4. Amorçage du serveur de production | propriétaire | — |
| 5. Environnements GitHub | propriétaire (administrateur du dépôt) | — |
| 6. Premier déploiement de staging | propriétaire ; Rayan suit le serveur (`docker compose ps`, `docker compose logs`) | — |
| 7. Compte des tests E2E | propriétaire | — |
| 8. Copie des sauvegardes hors du serveur | chacun pour le serveur qu'il détient | — |

Aucune clé privée ne circule : Rayan ne reçoit que la clé publique de déploiement de staging, le propriétaire ne reçoit que des informations publiques (IP, empreinte d'hôte).

## 1. Clés age de la CI (une par environnement)

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

L'option `--dry-run` montre les étapes sans rien changer. Le script installe Docker et le plugin compose, crée l'utilisateur `deploy` (groupe `docker`, clé CI seule), règle `ufw` (port SSH, 80, 443), active les mises à jour de sécurité, prépare `/opt/226os` et désactive le mot de passe SSH (gardez votre session ouverte et testez une connexion par clé dans un autre terminal). Un serveur est lié au profil de son premier déploiement : staging ne peut pas être déployé sur le serveur de production, ni l'inverse. À la fin il affiche la ligne à mettre dans `DEPLOY_KNOWN_HOSTS`.

## 5. Environnements GitHub

Réglages du dépôt, Environments, créer `staging` et `production`.

| | `staging` | `production` |
|---|---|---|
| Branche autorisée | `develop` | `main` |
| Relecteur requis | aucun | `Y4NN777` |
| Secret `DEPLOY_SSH_KEY` | clé privée du serveur de staging | clé privée du serveur de production |
| Secret `DEPLOY_KNOWN_HOSTS` | `<DEPLOY_HOST> ssh-ed25519 AAAA...` affichée par l'amorçage | idem |
| Secret `SOPS_AGE_KEY` | contenu de `ci-staging.key` | contenu de `ci-production.key` |
| Variable `DEPLOY_HOST` | IP ou nom du serveur | idem |
| Variable `DEPLOY_USER` | `deploy` | `deploy` |
| Variable `PUBLIC_HOST` | hôte public de staging (identique à celui du fichier chiffré) | `226os.tech` (idem) |

`DEPLOY_HOST` doit être écrit exactement comme dans `DEPLOY_KNOWN_HOSTS`. Supprimez ensuite les fichiers `*.key` locaux (gardez une copie hors ligne des clés age).

## 6. Premier déploiement

- `workflow_run` ne part que depuis la copie du workflow sur `develop` ; le déclencheur `push` sur `main` ne fonctionne qu'une fois le fichier arrivé sur `main` (premier passage de `develop` vers `main`, merge commit, sans autre changement).
- Après le premier push d'images, vérifiez dans GitHub, Packages, que les trois paquets `226os-*` sont liés au dépôt (sinon le serveur ne peut pas les tirer : liez-les, ou rendez-les publics).
- Le premier déploiement de staging crée la base, applique les migrations et charge les données de démonstration. Production : jamais de données de démonstration.

## 7. Compte des tests E2E (après le premier déploiement de staging)

1. Sur le site de staging, créez un vrai compte et vérifiez son e-mail.
2. Dans l'environnement `staging`, créez les secrets `E2E_USER` et `E2E_PASSWORD`.
3. Sans eux, le job `e2e` échoue avec un message clair ; le groupe connecté n'est jamais ignoré. Chaque exécution publie une question et une réponse avec ce compte. Le dépôt est public : en CI, Playwright n'écrit ni trace, ni vidéo, ni capture, ni rapport HTML (ils contiendraient le mot de passe et le cookie), et seul le XML junit est publié. Les étiquettes `tested-<arbre>` ne sont posées que si ce job a réellement tourné.

## 8. Sauvegardes

Avant chaque migration, `/opt/226os/backups` reçoit un dump chiffré avec age (`<profil>-<date>.dump.age`, pour les clés de `secrets/<profil>.recipients`) et son fichier de preuve ; le dump en clair est supprimé. Les dix derniers sont gardés. **Étape du détenteur de chaque serveur : aucune copie hors du serveur n'existe.** Copiez régulièrement ce dossier ailleurs (par exemple `rsync` depuis votre machine) : un disque perdu perd les sauvegardes avec la base. Déchiffrement : `age -d -i <votre clé age> fichier.dump.age > fichier.dump`, puis `scripts/restore-check.sh`.

## 9. Après coup

- La clé age de la CI est stockée dans GitHub : amendement de l'ADR-0009 du 2026-11-15.
- `ufw` ne filtre pas les ports publiés par Docker : ici seul Traefik publie 80 et 443.
- Retour arrière : voir [SHIP](../SHIP.md) section 5.
