# 226os — Delivery, Operations & Release Runbook

**Cible :** Datacloud/Systalink, démonstration accessible avant 2026-11-15 17:00 GMT.

## 1. Stratégie
- Compose local et Compose prod séparés ; un seul hôte pour MVP : Traefik en edge, conteneur web, conteneur API, PostgreSQL avec volume persistant.
- Traefik gère TLS Let's Encrypt (challenge selon DNS et firewall vérifiés) et routes ; DB jamais exposée publiquement.
- Domaines proposés : `226os.tech` pour web, `/api/v1` routé vers API ; possibilité `api.226os.tech` après confirmation DNS et CORS. Ne pas supposer DNS opérationnel.
- Images Docker non-root si possible, multi-stage, versions et digests épinglés, healthchecks, limites ressources, logs structurés, secrets fournis via configuration sécurisée.

## 2. Pipeline livraison
```text
PR -> G0/G1/G2/G3/G4/G5/G6 -> human review -> merge
  -> build immutable images (commit SHA + digest) -> staging smoke
  -> human approval production -> DB backup -> migrations
  -> deploy digest exact -> smoke E2E -> monitor -> close G7
```
- GitHub Actions avec `permissions` minimales et OIDC si un fournisseur le permet ; ne jamais exposer SSH keys ou tokens aux PR de forks.
- Conditions d'environnement `production` et approval humaine ; pas d'auto-déploiement aveugle depuis PR.

## 3. Preflight avant déploiement
- Compte Datacloud et **produit commercial actif** confirmés (condition CADEV), budget, promo, IP, accès SSH par clé, ports 80/443, DNS, stockage.
- Variables d'environnement définies par `.env.example` (sans valeurs sensibles), secrets hors Git.
- DB backup cohérent, espace disque disponible, migration dry-run sur base éphémère ; image web et API identifiées par digest.
- Rapports quality gates sur SHA cible disponibles, migration reversible ou plan de restauration explicite.

## 4. Smoke post-déploiement
1. HTTPS cert valide, redirections, headers et Traefik routage.
2. `/healthz` (liveness) et `/readyz` (DB et dépendances nécessaires) vérifiés.
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

## 7. Go/no-go
`GO` uniquement si G0–G7 passés, URL publique utilisable et obligations de participation remplies. Tout résultat non exécuté vaut gate non satisfait, sauf dispense humaine conforme à [TEST](TEST.md). Présenter honnêtement les fonctionnalités P1/P2 inachevées.
