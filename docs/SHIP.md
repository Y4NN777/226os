# 226os — Delivery, Operations & Release Runbook

**Cible :** Datacloud/Systalink, démonstration accessible avant 2026-11-15 17:00 GMT.

## 1. Stratégie
- Aucun workflow de déploiement n'existe encore : G7 est manuel et NOT RUN ; rien n'est déployé.
- Un seul arbre Compose à profils `local-dev` (machine locale), `staging-test` (branche `develop`) et `production-live` (branche `main`), avec fichiers d'environnement et secrets distincts par profil : cela satisfait l'intention « Compose local et prod séparés » (ADR-0006) ; Datacloud héberge `staging-test` et `production-live` ; `migrate-gated` est un quatrième profil technique de Compose (profil de job pour les migrations, pas un environnement) ; un seul hôte pour MVP : Traefik en edge, conteneur web, conteneur API, PostgreSQL avec volume persistant.
- Traefik gère TLS Let's Encrypt (challenge selon DNS et firewall vérifiés) et routes ; DB jamais exposée publiquement.
- Domaines proposés : `226os.tech` pour web, `/api/v1` routé vers API ; possibilité `api.226os.tech` après confirmation DNS et CORS. Ne pas supposer DNS opérationnel.
- Images Docker non-root si possible, multi-stage, versions et digests épinglés, healthchecks, limites ressources, logs structurés, secrets fournis via configuration sécurisée.

## 2. Pipeline livraison
```text
work branch -> PR develop -> develop staging/preview smoke
  -> production PR main -> human approval -> DB backup -> migrations
  -> deploy image digest tested on the develop SHA (same tree as main) -> production smoke -> monitor -> close G7
```
- Branches, promotions et protections : voir [`GIT_WORKFLOW.md`](GIT_WORKFLOW.md). `develop` correspond à `staging-test`, `main` à `production-live`.
- La production déploie l'image (par digest) testée sur le SHA `develop` validé, dont l'arbre est celui de `main` ; ne jamais reconstruire une autre révision lors de la promotion.
- GitHub Actions avec `permissions` minimales et OIDC si un fournisseur le permet ; ne jamais exposer SSH keys ou tokens aux PR de forks.
- Conditions d'environnement `production` et approval humaine ; pas d'auto-déploiement aveugle depuis PR.

## 3. Preflight avant déploiement
- Compte Datacloud et **produit commercial actif** confirmés (condition CADEV ; l'hébergement Datacloud est décidé, la preuve du compte reste à fournir), budget, promo, IP, accès SSH par clé, ports 80/443, DNS, stockage.
- Variables d'environnement définies par `.env.example` (sans valeurs sensibles) ; secrets chiffrés avec SOPS et age dans `secrets/*.enc.env` (ADR-0009), jamais en clair dans Git.
- DB backup cohérent, espace disque disponible, migration dry-run sur base éphémère ; image web et API identifiées par digest.
- Sauvegarde : `scripts/backup.sh <profil>` (`pg_dump -Fc`, dump en mode 600, fichier de preuve à passer en `BACKUP_EVIDENCE` à `migrate-up`). Restauration vérifiée : `scripts/restore-check.sh <profil> <dump>` (base jetable `restore_check_*`, comparaison du nombre de tables et de la version de migration, base supprimée ensuite ; `RESTORE_CHECK_CONFIRM=<profil>` hors local-dev). Exercice local-dev fait le 2026-11-12 ; staging et production restent à faire.
- Rapports quality gates sur SHA cible disponibles, migration reversible ou plan de restauration explicite.

## 4. Smoke post-déploiement
1. HTTPS cert valide, redirections, headers et Traefik routage.
2. `/healthz` (liveness) vérifié depuis l'extérieur. `/readyz` (base de données et dépendances nécessaires) vérifié depuis l'hôte, dans le conteneur : `docker compose exec api /busybox/busybox wget -q -O- http://127.0.0.1:8080/readyz`. Jamais par l'adresse publique : le routeur n'envoie pas `/readyz` à l'API, et le site web répond `200` avec sa page d'accueil (décision B2, ROADMAP section 10).
3. Auth, session, question/réponse/vote, recherche, fiche API, projet, candidature.
4. Lecture anonyme autorisée ; mutations non autorisées refusées ; FR/EN, mobile.
5. Logs sans secrets ; pas d'erreur 5xx significative sur le parcours ; métriques de latence collectées.

## 5. Repli / incident
- Redéployer l'image précédente par digest si le code est en cause ; pour migration incompatible, appliquer le plan de restauration explicitement testé. Ne jamais supposer qu'un rollback applicatif annule une migration de données.
- Suspendre les écritures si migration destructive ou risque d'intégrité. Capturer heure, erreur, SHA, décision, owner et postmortem court.
- Backups PostgreSQL chiffrés, rotation et restauration **vérifiée** avant présentation de la production comme récupérable.

## 6. CADEV — audit de soumission
- Équipe 2–5 participants, représentant et pays déclarés, clé de répartition et membre(s) inscrits.
- URL Git, URL démo publique, produit Datacloud actif, dépôt/documentation, compte de démonstration sécurisé si nécessaire.
- Projet original construit pendant le concours, registre `THIRD_PARTY_NOTICES.md`, `AI_USAGE.md`, provenance des dépendances, absence de licences interdites par le règlement.
- Soumission avant **15 novembre 2026, 17:00 GMT**. Chaque membre doit voter entre **15 novembre 18:00 et 16 novembre 23:59 GMT** sous peine d'exclusion. Prévoir responsable de suivi.
- Pointeurs : déclaration d'originalité dans [`docs/research/REFERENCES.md`](research/REFERENCES.md) (section « Déclaration d'originalité »), inventaire des dépendances dans [`THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md), usage de l'IA dans [`AI_USAGE.md`](../AI_USAGE.md).

## 7. Go/no-go
`GO` uniquement si G0–G7 passés, URL publique utilisable et obligations de participation remplies. Tout résultat non exécuté vaut gate non satisfait, sauf dispense humaine conforme à [TEST](TEST.md). Présenter honnêtement les fonctionnalités P1/P2 inachevées.
