# Mise en place du déploiement (propriétaire)

À faire une fois, dans cet ordre. Le workflow est décrit dans [RELEASE_DRAFT](../RELEASE_DRAFT.md) ; la stratégie dans [SHIP](../SHIP.md).

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
bash server-bootstrap.sh "ssh-ed25519 AAAA... ci-226os"
```

L'option `--dry-run` montre les étapes sans rien changer. Le script installe Docker et le plugin compose, crée l'utilisateur `deploy` (groupe `docker`, clé CI seule), règle `ufw` (22, 80, 443), active les mises à jour de sécurité, prépare `/opt/226os` et désactive le mot de passe SSH (il demande confirmation ; gardez votre session ouverte et testez une connexion par clé dans un autre terminal). À la fin il affiche la ligne à mettre dans `DEPLOY_KNOWN_HOSTS`.

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
3. Sans eux, le job `e2e` échoue avec un message clair ; le groupe connecté n'est jamais ignoré. Chaque exécution publie une question et une réponse avec ce compte.

## 8. Après coup

- Demandez l'amendement du point 8 de l'ADR-0009 (la clé de la CI est stockée dans GitHub, par décision du 2026-11-15).
- `ufw` ne filtre pas les ports publiés par Docker : ici seul Traefik publie 80 et 443.
- Retour arrière : voir [SHIP](../SHIP.md) section 5.
