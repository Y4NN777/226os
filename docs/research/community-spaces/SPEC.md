# Espaces et salons de développeurs 226os (brouillon de spécification, 2026-11-18)

## Goal
Donner aux membres de 226os des **espaces** (les « groupes ») par pays, par technologie et par centre d'intérêt (PRD 5.4), et un espace attaché à chaque projet de Construire et à chaque série d'Apprendre : ce sont les **salons** de la S10 de la ROADMAP (« un membre poste dans le salon d'un projet ou d'une série »). Un espace regroupe des **fils de discussion** : un membre ouvre un fil, les autres répondent, tout se lit sans compte, se retrouve par la recherche et se modère. Dans cette première version, les salons sont asynchrones ; le rafraîchissement en direct d'un fil ouvert (Server-Sent Events) est la phase 2, décrite dans le journal de décisions. Quand ça marche, une question rapide ou une annonce n'est plus perdue dans un groupe de messagerie : elle reste attachée à son contexte (pays, technologie, projet, formation) et se retrouve plus tard.

## Non-goals
- Pas de messagerie privée ni de message direct entre membres.
- Pas de direct dans la phase 1 (WebSocket, Server-Sent Events) : c'est la phase 2 (journal de décisions).
- Pas de notification (e-mail, push, cloche) dans la phase 1 ; seul un indicateur « nouvelles réponses » visible par l'auteur d'un fil est prévu (critère dédié).
- Pas d'import ni de miroir du contenu de WhatsApp, Telegram ou Discord : un espace peut seulement afficher le lien de son groupe externe.
- Pas d'espace privé ni sur invitation : tout espace se lit sans compte.
- Proposition de coupe, à valider par le propriétaire : pas de fil d'activité global ni de réputation dans cette version, bien que le PRD 5.4 les range dans la vie communautaire (question ouverte).
- Pas de création d'espace par n'importe quel membre : les espaces de pays, de technologie et d'intérêt sont créés par un modérateur ; ceux de projet et de série naissent avec leur projet ou leur série.

## Users
- **Membre** (connecté, e-mail vérifié) : lit, rejoint ou quitte un espace, ouvre un fil, répond, modifie ou supprime ses propres messages, signale.
- **Visiteur** (sans compte) : lit les espaces et les fils, est invité à créer un compte pour écrire.
- **Responsable de projet ou auteur de série** : son espace naît avec le projet ou la série ; il épingle un fil et déclare le lien du groupe externe.
- **Modérateur** : crée les espaces de pays, de technologie et d'intérêt ; masque ou rétablit un fil ou une réponse ; traite les signalements ; retire un lien de groupe externe.

## Core flows
1. Découvrir les espaces : un visiteur ouvre Connect → onglet « Espaces » → liste des espaces (pays, technologies, intérêts) avec leur nombre de fils et de membres → il ouvre l'espace « Burkina Faso ».
2. Rejoindre ou quitter : un membre ouvre un espace → « Rejoindre » → l'espace apparaît dans « Mes espaces » et son nombre de membres augmente → « Quitter » le retire.
3. Lire un fil : depuis un espace → liste des fils récents (épinglés en tête) → ouvre un fil → lit le message et les réponses, sans compte.
4. Ouvrir un fil : un membre qui a rejoint l'espace → « Nouveau fil » → titre et message (Markdown, extraits de code) → le fil apparaît en tête de l'espace.
5. Répondre : un membre qui a rejoint l'espace répond à un fil → la réponse s'affiche sous le fil ; l'auteur du fil voit l'indicateur « nouvelles réponses » à sa prochaine visite de l'espace.
6. Modifier ou supprimer son message : l'auteur modifie son fil ou sa réponse (mention « modifié ») ou le supprime (« message supprimé » à la place du texte, réponses conservées).
7. Salon de projet : le responsable d'un projet ouvre l'onglet « Discussion » du projet → c'est l'espace du projet → il épingle un fil « Comment contribuer ».
8. Salon de série : un apprenant d'une série d'Apprendre ouvre « Discussion » → pose une question sur la leçon en cours.
9. Groupe externe : le responsable d'un espace déclare le lien de son groupe WhatsApp, Telegram ou Discord → l'espace affiche « Rejoindre le groupe » ; un modérateur peut retirer ce lien.
10. Signaler et modérer : un membre signale un fil ou une réponse (nouvelle opération) → un modérateur le masque → il disparaît pour tous sauf les modérateurs ; il peut le rétablir.

## Acceptance criteria
- [ ] Un visiteur sans compte liste les espaces et lit un fil ; toute écriture sans session est refusée. verify: tests d'intégration HTTP sur PostgreSQL réel : lectures anonymes 200, écritures anonymes 401.
- [ ] Un membre rejoint (`POST /spaces/{id}/members`) et quitte (`DELETE /spaces/{id}/members`) un espace ; rejoindre deux fois ne change rien. verify: test d'intégration (idempotence, nombre de membres).
- [ ] Ouvrir un fil ou répondre exige d'avoir rejoint l'espace et un e-mail vérifié : sinon 403 avec un message qui dit pourquoi. verify: tests négatifs d'autorisation et ligne de la matrice d'autorisation.
- [ ] Le nombre de membres d'un espace compte les seuls membres qui l'ont rejoint ; la liste des membres, si elle est affichée, ne montre que les profils listables par la règle de l'annuaire (visibilité du profil, ouverture à la collaboration). verify: test d'intégration avec un membre non listable qui ne doit pas apparaître.
- [ ] Seul un modérateur crée un espace de pays, de technologie ou d'intérêt ; un membre reçoit 403. verify: test négatif d'autorisation.
- [ ] Créer un projet de Construire ou une série d'Apprendre crée son espace dans la même transaction ; supprimer le projet ou la série archive l'espace : il reste lisible, toute écriture y est refusée (409). verify: test d'intégration (transaction, archivage, écriture refusée).
- [ ] Seul le responsable de l'espace ou un modérateur épingle et désépingle un fil ; les autres reçoivent 403. verify: test négatif d'autorisation.
- [ ] L'auteur modifie ou supprime son propre fil ou sa propre réponse ; sur le message d'un autre, 403 ; un message supprimé disparaît des listes et de la recherche, ses réponses restent. verify: tests d'intégration et tests négatifs.
- [ ] L'auteur d'un fil voit sur l'espace un indicateur du nombre de réponses arrivées depuis sa dernière lecture du fil. verify: test d'intégration (lecture, nouvelle réponse, indicateur à 1, relecture, indicateur à 0).
- [ ] Un membre signale un fil ou une réponse (nouvelle opération, comme `createResourceErrorReport` : identité du signaleur jamais publique, limite de débit) ; le signalement apparaît dans la file des modérateurs. verify: tests d'intégration et test négatif (le signaleur n'est pas exposé).
- [ ] Un fil ou une réponse masqué par un modérateur disparaît des listes, de la recherche et des pages pour un non-modérateur, et réapparaît au rétablissement. verify: tests d'intégration couvrant liste, page et `GET /search`.
- [ ] Les fils publiés apparaissent dans la recherche globale (groupe « Espaces »), avec la tolérance aux fautes existante. verify: test d'intégration de `GET /search` avec une faute de frappe.
- [ ] La liste des fils d'un espace est paginée par curseur signé (codec existant) ; avec un jeu de 10 000 fils dans un espace et 5 réponses par fil, une page de 20 fils répond en moins de 300 ms côté serveur. verify: test d'intégration sur ce jeu et `EXPLAIN (ANALYZE, BUFFERS)` consigné.
- [ ] Le lien de groupe externe n'est accepté que vers WhatsApp, Telegram ou Discord (liste blanche d'hôtes), déclaré par un responsable ou un modérateur, et un modérateur peut le retirer. verify: tests unitaires de validation d'URL et tests d'intégration d'autorisation.
- [ ] Interface honnête : Connect garde le bloc « Espaces : bientôt » tant qu'aucun espace réel n'existe en production ; un espace sans fil affiche un état vide honnête (« Aucun fil pour l'instant », sans compteur inventé). verify: tests web des deux états et mise à jour du scénario « Annoncer honnêtement ce qui n'est pas construit » de `docs/SPEC.md`.
- [ ] Aucun espace ni fil de démo en production : les données de démo n'existent qu'en `local-dev` et `staging-test` (règle des seeds), et les espaces des comptes de démo sont masqués en `staging-test` comme le sont leurs profils. verify: test du seed et de la requête de liste avec `include_demo` faux.
- [ ] Les pages Espaces fonctionnent à 360 px sans défilement horizontal, en FR et en EN, au clavier, contrastes AA, et correspondent au prototype validé par le propriétaire. verify: tests web (Vitest), captures Playwright à 360 et 1440 px relues avec le prototype validé, contrôle axe sans violation.
- [ ] Le Markdown d'un fil est rendu sans HTML brut (pas de XSS), comme dans Résoudre. verify: test web avec une charge `<script>` et un lien `javascript:`.
- [ ] L'écriture (fils, réponses, signalements) est limitée par la limite de débit existante ; au-delà, 429 avec Retry-After. verify: test d'intégration de la limite.

## Risks and assumptions
- Les membres écriront dans des espaces plutôt que dans leurs groupes de messagerie (hypothèse la plus risquée) — cheapest test: ouvrir trois espaces pilotes (Burkina Faso, Go, mobile money) et un salon de projet ; réussite si, sur deux semaines, au moins 5 fils et 20 réponses par espace pilote ; échec sous 1 message par jour alors que le groupe externe est actif.
- Les groupes WhatsApp et Telegram sont le lieu actuel des communautés tech en Afrique de l'Ouest (hypothèse, non sourcée) — cheapest test: demander aux trois communautés partenaires de déclarer leur groupe ; compter les clics « Rejoindre le groupe » sur deux semaines.
- La charge de modération reste gérable sans direct (risque K17 de la ROADMAP) — cheapest test: sur les espaces pilotes, moins de 5 signalements par semaine et une décision en moins de 24 h ; au-delà, revoir avant d'ouvrir d'autres espaces.
- Le serveur de production suffit, sans connexion persistante (une machine virtuelle de 3,8 Go de RAM, mesurée par `free -h` sur le serveur le 2026-11-15) — cheapest test: test de charge `k6` sur la liste de fils et la publication, réussite sous le budget de 300 ms à 20 utilisateurs simultanés.

## Decision log
- 2026-11-18 — Proposition : phase 1, espaces asynchrones (fils) par pays, technologie et intérêt, et salons de projet et de série en fils asynchrones, avec un lien de groupe externe optionnel, parce que cela répond au problème des connaissances dispersées (PRD, problème 1), tient sur le serveur actuel et sur mobile, et réutilise Résoudre, la recherche et la modération. Revisit if les fils des espaces pilotes restent sous un message par jour alors que leur groupe externe est actif.
- 2026-11-18 — Proposition : phase 2, salons en direct de la S10 (ROADMAP), construits sur la phase 1 : un fil ouvert se rafraîchit par Server-Sent Events, sans nouveau modèle de données. Déclencheur : phase 1 en production, au moins un salon de projet actif (plus de 5 messages par jour), test de charge des connexions persistantes réussi et charge de modération sous le seuil ci-dessus.
- 2026-11-18 — Priorité : P2, comme le PRD 5.4 (`x-priority: P2` dans le contrat) ; l'avancer en P1 demande un amendement du PRD.

## Open questions
- Valider l'approche (phase 1 asynchrone, phase 2 en direct) et confirmer que des fils asynchrones comptent comme « salons » pour la phase 1 — owner: Y4NN777
- Garder la priorité P2 ou l'avancer en P1 par un amendement du PRD — owner: Y4NN777
- Valider la coupe du fil d'activité global et de la réputation, que le PRD 5.4 range dans la vie communautaire — owner: Y4NN777
- Liste des espaces de départ (pays : Burkina Faso puis la sous-région ; technologies ; intérêts) et qui les crée — owner: Y4NN777
- Un salon pour chaque projet et chaque série, ou seulement à la demande du responsable — owner: Y4NN777
- Domaine du code : nouveau domaine `community` (prévu dans `docs/SPEC.md`) ou extension de Résoudre — owner: Y4NN777, avec le lead backend
- Retour du propriétaire (2026-11-18) : l'application est « trop rigide et manque de fun, même côté frontend ». Le prototype des espaces (F11) doit proposer des réponses concrètes dans le cadre de `DESIGN.md` (réactions légères sur un fil, présence des membres, mouvements courts du design system, ton des textes) ; lesquelles retenir — owner: Y4NN777
- Relecteurs : migrations (`mibienpanjoe`, ADR-0008), interface (`PANK4SS`) — owner: Y4NN777
