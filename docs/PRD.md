# 226os — Product Requirements Document

**Version :** 3.2 (amendement 3.2 proposé le 2026-11-12 ; version 3.1 approuvée le 2026-11-11). **Statut :** périmètre produit approuvé, plan de réalisation adaptable. **Date :** 2026-11-12.  
**Équipe :** 7's (4 membres, inscrits à CADEV). **Produit :** 226os. **Domaine :** 226os.tech (réservé par le propriétaire, rien n'y est encore déployé).

## Amendement 3.2 — Apprendre et Connect mis en avant, assistant sourcé au P0

**Date :** 2026-11-12. **Version :** 3.1 → 3.2. **Statut :** PROPOSÉ, par décision du propriétaire du dépôt (Y4NN77) du 2026-11-12, consignée dans le journal des décisions de [ROADMAP.md](../ROADMAP.md) (section 14). Le propriétaire approuve ce texte lors de la revue de la PR qui l'introduit ; tant que cette PR n'est pas fusionnée, l'amendement reste proposé. Comme 3.1, c'est un changement de périmètre au sens de la règle de traçabilité ci-dessous, consigné sans réécrire l'historique des versions précédentes.

**Principe :** rien n'est supprimé ni rétrogradé. Les neuf domaines (section 5) et les quatre expériences différenciantes (section 6) restent tels quels. Les seuls changements de priorité sont ceux listés ci-dessous, chacun de façon explicite.

**1. Apprendre et Connect mis en avant.** Apprendre (libellé de travail « Learn & Grow ») et Connect deviennent les deux piliers mis en avant dans la navigation, sur l'accueil et dans les textes. Solve, Discover et Build restent des piliers de premier plan, accessibles depuis la même navigation. C'est un choix d'architecture de l'information et de rédaction : il ne retire ni ne rétrograde aucun pilier. La question de mettre fin au « poids égal » de l'amendement 3.1 (« Learn as the heart », ROADMAP O7 et H16) reste ouverte et n'est pas tranchée ici. Le libellé français définitif de « Learn & Grow » est à fixer par le propriétaire ; d'ici là, le libellé utilisateur reste « Apprendre ». Tous les piliers forment un seul écosystème : une identité, une recherche, des tags et des thèmes partagés, des relations typées entre contenus, un assistant appuyé sur tous les contenus publiés et un annuaire qui relie les personnes à leurs contributions.

**2. Assistant IA sourcé : de P2 à P0 (section 5.8).** Décision du propriétaire. L'assistant :
- répond uniquement à partir des contenus publiés sur la plateforme et cite ses sources, que le serveur vérifie parmi les passages retrouvés ;
- refuse quand il ne trouve aucune source, sans appeler le fournisseur, et renvoie vers la recherche ;
- n'affirme jamais de garantie sur une API ni sur la vérification d'une solution ;
- signale son incertitude, avec les dates des sources, quand celles-ci divergent ou ne sont pas datées ;
- termine chaque réponse par une section « Comment vérifier » ;
- est étiqueté comme IA et ne peut jamais être voté ni accepté comme réponse ;
- n'envoie au fournisseur que le texte de la question et les contenus publics retrouvés, après masquage des secrets et des adresses e-mail évidents ; l'interface indique que la question part chez le fournisseur ;
- est limité en débit, et répond « assistant indisponible » sans faire échouer la page si le fournisseur est en panne ou le quota épuisé.

Fournisseur : Google AI Studio avec les modèles Gemma, derrière un port qui permet d'en changer par configuration. Restent à vérifier par le propriétaire : l'identifiant exact du modèle, les quotas réels et les conditions de l'offre gratuite. Les conditions d'AI Studio n'autorisent que les services payants pour des utilisateurs de l'Espace économique européen, de la Suisse et du Royaume-Uni ; choisir entre un projet gratuit et un projet lié à un compte de facturation reste une décision ouverte du propriétaire, et l'assistant reste désactivé par configuration tant qu'elle n'est pas prise. L'accès des visiteurs non connectés reste aussi à décider. Le mode pédagogique « indice d'abord » et la demande de la tentative de l'apprenant sont P1. La recherche sémantique et les résumés restent P2.

**3. Connect au P0, minimal et réel (section 5.5).** Un annuaire des membres en lecture seule, filtrable par pays, langue et compétence :
- il n'expose que les champs que l'appelant a le droit de voir, et jamais l'adresse e-mail ;
- un profil n'y apparaît pour un appelant que si le compte est actif, si `open_to_collab` vaut vrai et que l'appelant peut voir cette information, et si la visibilité du profil le permet ;
- les filtres ne portent que sur des champs visibles pour l'appelant.

Un nouveau réglage de visibilité du profil (public, membres, privé) est ajouté à Identity (section 5.4). Un profil masqué renvoie la même 404 qu'un profil inconnu. La signature d'auteur sur les contenus publiés reste visible quel que soit ce réglage. Cet annuaire est la forme minimale de la « découverte de membres » de la section 5.4 ; la découverte enrichie reste P1. Le mentorat, la relecture de CV, la préparation aux entretiens, les opportunités, les groupes d'étude et les espaces communautaires restent P1 ou P2 et sont présentés honnêtement comme « bientôt » ; aucune mise en relation n'est annoncée. Les permanences (office hours) sont un processus d'équipe, pas une fonctionnalité.

**4. Apprendre (section 5.9).**
- Thèmes réorientés vers le développement à l'ère des agents IA : 14 thèmes après une migration d'ajout et de renommage. Les 9 thèmes existants restent ; s'ajoutent `learn-to-learn`, `ai-assisted-dev`, `ai-agents`, `open-models` et `african-languages` ; `ai`, `mobile` et `open-source` sont renommés sans changer de code.
- Catégories de ressources (Discover, section 5.2) : ajout de `ai-models`, `ai-tooling` et `ai-datasets`, soit 18 catégories après la migration ; `cloud-hosting` est renommée. Le thème `payments` et les catégories `mobile-money` et `sms-ussd` restent, comme l'expérience différenciante 1.
- Un bouton « Poser une question sur cette leçon » et des fils par thème relient les discussions aux leçons, sans nouvelle table.
- Les salons en direct (rattachés à un projet ou à une série) sont CONDITIONNELS : ils ne sont construits que si le point d'étape du samedi 2026-11-14 à 12:00 GMT le permet ; sinon ils restent P2 et sont présentés comme non construits.

**5. Contenu original, sans contenu tiers.** Aucun contenu tiers n'est importé ni aspiré. Les sources externes servent uniquement de références qui nous inspirent et de liens documentaires sur les fiches Discover. En préproduction et en production, le contenu est réel : il est créé via l'application par les comptes réels de l'équipe. Les données de démonstration, avec des comptes de démonstration qui ne correspondent à personne, n'existent qu'en `local-dev`. La série gratuite publique (US-LEARN-001) est une courte série originale, chargée une seule fois par un script revu en PR et lancé par le propriétaire.

**6. Équipe et calendrier.** L'équipe compte quatre personnes (le propriétaire et responsable du projet, deux développeurs backend, une personne frontend), inscrites à CADEV. Soumission CADEV avant le 2026-11-15 à 17:00 GMT ; vote du 2026-11-15 18:00 au 2026-11-16 23:59 GMT. Le domaine 226os.tech est réservé par le propriétaire et n'est pas encore déployé.

**7. Positionnement : l'Afrique de l'Ouest d'abord, puis toute l'Afrique.** Ce libellé remplace la présentation du produit comme « plateforme panafricaine » ; l'ouverture panafricaine n'est pas exclue, elle vient ensuite. Le modèle de données, la liste des 54 pays et l'interface FR/EN ne changent pas. La région « Afrique de l'Ouest » du produit est celle de la classification des Nations unies, limitée aux pays de notre liste : BJ, BF, CV, CI, GM, GH, GN, GW, LR, ML, MR, NE, NG, SN, SL, TG (16 pays ; à vérifier sur la liste M49, le propriétaire peut l'ajuster). Elle vit comme une constante côté interface, sans migration ni colonne. Les propositions d'interface qui en découlent (filtre de région par défaut avec « Toute l'Afrique » en un clic, curation, mises en avant) sont listées dans `TODO.md` (S7) au statut PROPOSÉ. Le nom 226os vient de +226, l'indicatif téléphonique du Burkina Faso ; cela concerne le nom, pas le positionnement.
**Impact documentaire :** SPEC (US-INTEL-001, US-CONNECT-001, US-LEARN-002, extension de US-IDENTITY-001, traçabilité), DATABASE_DESIGN (colonne `profile_visibility` et ajouts de taxonomie prévus), README, ROADMAP et TODO. Les opérations de l'annuaire et de l'assistant entrent dans `api/openapi.yaml` par des PR de contrat séparées, avec les lignes correspondantes de la matrice d'autorisation.

## Amendement 3.1 — Pilier Apprendre (Learn)

**Date :** 2026-11-11 · **Version :** 3.0 → 3.1 · **Approbation :** décision explicite du propriétaire du dépôt (Y4NN77), donnée le 2026-11-11. Cet amendement est un changement de périmètre au sens de la règle de traçabilité ci-dessous ; il est consigné ici sans réécrire l'historique des versions précédentes.

**Décision :** le contenu d'apprentissage devient un pilier public pleinement visible, de poids égal aux autres. Il couvre les articles, les tutoriels et les formations, y compris des **formations gratuites à l'IA que le fondateur donne personnellement à travers ses articles et ses tutoriels**. Chaque membre est **invité à publier les siens**.

**Piliers publics :** Questions · API & outils · Projets open source · Apprendre (articles, tutoriels, formations). Libellé utilisateur FR « Apprendre », EN « Learn ».

**Structure :** Learn (identifiant de domaine `LEARN`) est un **neuvième domaine**, section 5.9, à côté des huit existants. Les contenus d'apprentissage cessent d'être une ligne « articles » du domaine Solve ; les questions/réponses, snippets, votes et revue de code restent dans Solve.

**Supersession :** la mention « articles » en P1 sous Solve (sections 5.1 et 8, version 3.0) est remplacée par cet amendement. Les « retours d'expérience » restent listés sous Solve.

**Statut des priorités :** la répartition P0/P1/P2 de la section 5.9 est **validée par le propriétaire le 2026-11-11** : Apprendre fait partie du P0 (lecture et publication d'articles et tutoriels, formation gratuite lisible publiquement). Une publication sans session est refusée avec 401 ; un refus pour un membre authentifié non autorisé reste 403. Les quatre expériences différenciantes (section 6) sont inchangées.

**Impact documentaire :** SPEC (US-LEARN-001, frontières de domaines), TRACEABILITY (DOM-LEARN), TASKS (notes de périmètre, sans nouvelle tâche), README.

## Objet et séparation des responsabilités
Ce document est la **source de vérité du QUOI et du POURQUOI**. Il ne prescrit ni l'implémentation détaillée ni les commandes d'outillage. Les domaines P1/P2 demeurent dans le produit ; une priorité est un ordre de livraison, jamais une suppression. **Aucune nouvelle fonctionnalité n'est ajoutée par ce découpage** ; les amendements 3.1 et 3.2 ci-dessus sont les seules extensions de périmètre, avec accord humain explicite (pour 3.2, l'approbation du propriétaire sur la PR qui l'introduit).

Documents liés : [SPEC](SPEC.md) (comportements et contrats), [DEV](DEV.md) (mise en œuvre), [TEST](TEST.md) (preuves), [SHIP](SHIP.md) (mise en production), [RULES](RULES.md) (normes), [AGENTS](../AGENTS.md) (orchestration).

## 1. Résumé exécutif

226os est une plateforme construite pour et par les développeurs, qui part de l'Afrique de l'Ouest et s'ouvre à toute l'Afrique (amendement 3.2). Elle réunit l'entraide technique, les connaissances vérifiables, la découverte d'API/SDK et de services locaux, les projets open source, la collaboration et les opportunités de carrière dans un écosystème commun. Elle vise à franchir les barrières géographiques, linguistiques et de connectivité.

**Positionnement :** « L'écosystème où les développeurs africains partagent leurs connaissances, résolvent leurs défis et construisent ensemble. »

**Différence recherchée :** relier une question technique à une réponse expliquée et sourcée, à des ressources pertinentes, puis à un projet ou une opportunité de collaboration, et à des articles, tutoriels et formations qui permettent d'apprendre et de transmettre. 226os n'est ni un clone de Stack Overflow, ni un simple annuaire open source.

**Référence de marché :** 225os (`https://www.225os.com/`) est une inspiration en matière de découverte de projets et d'identité communautaire ; 226os ne doit pas en reproduire le produit ni le code.

## 2. Problèmes utilisateurs et objectifs

### Problèmes

1. Les connaissances techniques sont dispersées entre forums, dépôts, messageries et réseaux sociaux ; les solutions manquent souvent de contexte vérifiable.
2. Les API et outils utiles dans les pays africains sont difficiles à découvrir, comparer et intégrer, particulièrement quand leur documentation est incomplète.
3. Les mainteneurs open source et porteurs de projets peinent à trouver des contributeurs et à rendre leur travail visible.
4. Les développeurs juniors trouvent difficilement mentorat, feedback sur leurs travaux et opportunités adaptées.
5. Les échanges transfrontaliers restent limités par la langue et les contraintes réseau.

### Objectifs produit

- Résoudre et capitaliser des problèmes techniques.
- Apprendre et transmettre : articles, tutoriels et formations, dont des formations gratuites à l'IA, ouverts à la publication par tous les membres.
- Rendre visibles et exploitables des ressources et projets africains.
- Permettre la collaboration et la contribution open source.
- Connecter personnes, connaissances, outils et opportunités.
- Proposer une expérience mobile, bilingue et économe en données.
- Favoriser des réponses accompagnées de preuves, sources et contexte.

### Indicateurs à suivre

Pour CADEV : parcours de démonstration fonctionnels, disponibilité HTTPS, aucun incident bloquant P0, utilisabilité sur mobile, preuve d'utilisation de Datacloud, traçabilité du code et des licences. Après lancement : taux de résolution des questions, ressources consultées et maintenues, besoins de contribution pourvus, candidatures à des projets, engagement des membres, participation FR/EN et performances en débit réduit. Ne pas inventer d'objectifs chiffrés d'adoption sans référence initiale.

## 3. Publics cibles

| Public | Besoin principal |
|---|---|
| Développeur débutant / étudiant | Comprendre les erreurs, apprendre, partager ses questions, recevoir un retour |
| Développeur expérimenté | Transmettre des réponses fiables, publier des outils, contribuer et faire de la revue de code |
| Mainteneur open source | Présenter un dépôt, identifier des compétences et trouver des contributeurs |
| Intégrateur d'API locales | Identifier fournisseurs, pays, documentation, SDK et limites d'intégration |
| Porteur de projet / cofondateur | Présenter une initiative et recruter des collaborateurs |
| Mentor / recruteur | Découvrir des profils, contributions et opportunités d'accompagnement |

## 4. Principes produit

- **Un seul écosystème connecté :** Solve, Discover, Build et les autres piliers partagent identité, recherche, tags, profils et liens contextuels.
- **L'utilité avant la quantité :** chaque capacité doit résoudre un problème réel et démontrable.
- **Evidence-based :** conserver les références, versions, contexte d'exécution et dates de mise à jour lorsque pertinentes.
- **Open source & attribution :** favoriser la contribution sans présenter des travaux tiers comme originaux.
- **Africa-first, not Africa-only :** contexte africain fort, accès possible au-delà des frontières, identité 226os conservée.
- **Mobile-first, FR/EN, faible bande passante :** expériences utilisables dans des conditions réelles.
- **Sécurité et modération intégrées :** protections de base dès la première livraison.

## 5. Périmètre fonctionnel complet : neuf domaines, quatre piliers publics

### 5.1 SOLVE — Entraide technique et connaissances

**Fonctionnalités conservées :**
- Questions et réponses de type Q&A ; publication, modification et consultation.
- Markdown, blocs de code et coloration syntaxique ; extraits réutilisables.
- Votes, réponse acceptée, réputation et badges.
- Revue de code collaborative.
- Retours d'expérience (les articles, tutoriels et formations relèvent du domaine LEARN, section 5.9, depuis l'amendement 3.1).
- Recherche par mots-clés, langage, technologie, difficulté, tags et statut.
- Discussions thématiques et liens vers ressources/projets.
- Contexte technique : langage, version, environnement, erreurs, description des tentatives.
- Solutions documentées avec sources, contexte et confirmations lorsque disponibles.

**Parcours minimal :** A pose une question avec un snippet ; B répond ; A accepte ; C retrouve la réponse via la recherche et suit le lien vers une ressource ou un projet lié.

**Priorités de réalisation :** P0 questions/réponses, snippets, tags, votes, réponse acceptée, recherche et liens contextuels ; P1 réputation, badges et discussions plus riches (la ligne « articles » en P1 de la version 3.0 est remplacée par l'amendement 3.1) ; P2 revue de code avancée.

### 5.2 DISCOVER — African API & Developer Tools Hub

**Fonctionnalités conservées :**
- Catalogue d'API, SDK, bibliothèques et outils africains ou utiles au contexte africain.
- Catégories : paiement/Mobile Money, SMS, identité/KYC, fintech, logistique et autres catégories utiles.
- Fiches fournisseurs, pays couverts, capacités, liens de documentation et SDK.
- Exemples d'intégration/code et informations de version.
- Recherche et filtres par pays, catégorie, langage et fournisseur.
- Comparaison de fournisseurs par couverture et capacités techniques.
- Avis/retours communautaires, signalement d'informations obsolètes ou d'API non fonctionnelles.
- Sources vérifiables, date de dernière vérification et liens vers discussions Solve.

**Parcours minimal :** filtrer les solutions de paiement pour un pays ; ouvrir une fiche documentée ; consulter un exemple ; accéder aux questions techniques liées.

**Priorités :** P0 catalogue, filtres, fiches sourcées et liens ; P1 comparaison et feedback ; P2 vérifications automatisées et exemples exécutables.

### 5.3 BUILD — Open Source & Project Collaboration

**Fonctionnalités conservées (centrales, non optionnelles pour la vision) :**
- Annuaire de projets open source africains et import/référencement de dépôts GitHub publics.
- Publication de projets, idées et initiatives en cours.
- Description, technologies, objectifs, état et liens vers le code/documentation.
- Recherche de contributeurs **et de cofondateurs** par rôle, compétence et niveau.
- Publication de besoins concrets de contribution.
- Candidatures, manifestations d'intérêt et état des demandes.
- Visibilité des contributions communautaires et des profils associés.
- Liens entre projets, problèmes Solve et ressources Discover.
- Découverte de projets par pays, technologies et centres d'intérêt.

**Parcours minimal :** A importe un dépôt public, publie un besoin « Go/backend » ; B découvre le projet, consulte le profil et manifeste son intérêt ; A peut voir la demande. Une question Solve ou fiche d'API peut être associée au projet.

**Priorités :** P0 publication/import par URL GitHub, fiche projet, besoin de contributeurs, demande de collaboration et relation inter-domaines ; P1 gestion plus riche des candidatures/contributions ; P2 synchronisation GitHub avancée et matching élaboré.

### 5.4 IDENTITY & COMMUNITY — Profils et vie communautaire

**Fonctionnalités conservées :**
- Profils : nom public, bio, pays, langues, technologies/compétences, GitHub, portfolio.
- Réputation, badges et contributions visibles.
- Découverte de membres selon compétences, intérêts et pays.
- Communautés/espaces par pays, technologies et centres d'intérêt.
- Fil d'activité, publications et échanges thématiques.

**Priorités :** P0 création de compte, connexion et profils minimaux utilisables dans tous les parcours ; P1 activité, réputation et découverte enrichie ; P2 espaces communautaires complets.

### 5.5 CONNECT — Mentorat et développement de carrière

**Fonctionnalités conservées :**
- Demandes/offres de mentorat et mise en relation juniors/seniors.
- Relecture de CV et portfolios.
- Préparation aux entretiens techniques.
- Offres de stages, emplois et missions à distance.
- Parcours d'apprentissage, groupes d'étude et sessions communautaires de programmation.
- Annuaire des membres en lecture seule, filtrable par pays, langue et compétence (amendement 3.2).

**Priorité :** conserver les parcours et entités dans la roadmap ; ne pas annoncer de mise en relation complète en CADEV si elle n'est pas réellement fonctionnelle. Un pilote est envisageable après validation du cœur de l'application.

**Priorités de réalisation (amendement 3.2, proposé) :** P0 annuaire des membres en lecture seule : seuls les champs visibles pour l'appelant, jamais l'adresse e-mail, filtres limités aux champs visibles, profil listé seulement si le compte est actif, si `open_to_collab` est vrai et visible pour l'appelant et si la visibilité du profil le permet. Les autres parcours de Connect (mentorat et mise en relation, relecture de CV et de portfolios, préparation aux entretiens, offres de stages, d'emplois et de missions, groupes d'étude, sessions et espaces communautaires) restent P1 ou P2 et sont affichés comme « bientôt », sans écran factice. Les permanences sont un processus d'équipe, pas une fonctionnalité.

### 5.6 LANGUAGE — Collaboration panafricaine multilingue

**Fonctionnalités conservées :**
- Interface bilingue français/anglais.
- Publication dans la langue choisie et conservation du texte original.
- Traduction assistée des discussions, clairement indiquée.
- Recherche translingue et découverte de contenus indépendamment de la langue.
- Possibilité de collaboration entre personnes de pays différents, en privilégiant compétences et intérêts.

**Priorités :** P0 interface FR/EN ; P1 traduction assistée des contenus ; P2 recherche translingue/recommandations enrichies. Ne pas confondre UI traduite et contenus traduits.

### 5.7 ACCESS — Accessibilité et connectivité

**Fonctionnalités conservées :**
- Mobile-first et responsive.
- Mode faible consommation de données ; réduction des médias et chargement progressif.
- Mise en cache pertinente, optimisation des ressources et images.
- Brouillons hors connexion et fonctionnement dégradé en réseau instable.
- Navigation accessible, états de chargement, erreurs et parcours clavier.

**Priorités :** P0 design responsive et optimisations réseau mesurées ; P1 mode léger explicite/cache amélioré ; P2 brouillons offline et mode dégradé avancé.

### 5.8 INTELLIGENCE — Assistance et découverte contextuelles

**Fonctionnalités conservées :**
- Recherche sémantique de solutions techniques.
- Suggestions de questions similaires et de ressources associées.
- Recommandations de projets selon compétences.
- Aide à la formulation des questions.
- Résumés de longues discussions.
- Suggestions d'API liées au problème.
- Assistance IA avec citations/références aux sources.
- Mode pédagogique « indice d'abord » et demande de la tentative de l'apprenant avant la réponse (amendement 3.2).

**Priorités :** P0 recherche textuelle et liens contextuels déterministes ; P1 suggestions simples ; P2 sémantique, résumés et assistant contextuel sourcé. L'IA ne doit pas inventer des garanties sur une API ou la vérification d'une solution.

**Priorités révisées (amendement 3.2, proposé) :** l'assistant sourcé passe de **P2 à P0**. Il répond uniquement à partir des contenus publiés sur la plateforme, cite ses sources, refuse sans source, n'affirme aucune garantie sur une API ni sur la vérification d'une solution, termine chaque réponse par « Comment vérifier », est étiqueté comme IA et ne peut être ni voté ni accepté comme réponse, n'envoie au fournisseur que la question et des contenus publics, et est limité en débit (règles complètes dans l'amendement 3.2). Il reste désactivé par configuration tant que le propriétaire n'a pas choisi entre un projet gratuit et un projet lié à la facturation. Le mode « indice d'abord » et la demande de la tentative de l'apprenant sont **P1**, avec les suggestions simples. La recherche sémantique et les résumés restent **P2**.

### 5.9 LEARN — Apprendre : articles, tutoriels et formations

Pilier public de même poids que Questions, API & outils et Projets open source (amendement 3.1). Libellé FR « Apprendre », EN « Learn ». Depuis l'amendement 3.2 (proposé), Apprendre est mis en avant avec Connect dans la navigation, sur l'accueil et dans les textes, sans changement de périmètre pour les autres piliers.

**Fonctionnalités conservées :**
- Articles et tutoriels : Markdown, blocs de code, tags, niveau et langue ; consultation publique.
- Formations gratuites à l'IA, données personnellement par le fondateur à travers ses articles et tutoriels, lisibles publiquement.
- Publication ouverte : chaque membre est invité à publier ses propres articles, tutoriels et formations.
- Liens explicites vers les questions, API/outils et projets concernés, et retour depuis ceux-ci.
- Séries de formation structurées avec progression, mises en avant d'auteurs, formats plus riches.
- Certificats et exercices interactifs.
- Thèmes tournés vers le développement à l'ère des agents IA : 14 thèmes après l'ajout de `learn-to-learn`, `ai-assisted-dev`, `ai-agents`, `open-models` et `african-languages`, et le renommage de `ai`, `mobile` et `open-source` (amendement 3.2).
- Bouton « Poser une question sur cette leçon » qui ouvre une question Solve liée à l'article, et fils par thème qui réunissent les articles d'un thème et les questions portant le tag du même nom (amendement 3.2).
- Salons en direct rattachés à une série (amendement 3.2, conditionnels).

**Parcours minimal :** un visiteur lit un tutoriel, suit la série gratuite sur l'IA, puis ouvre la question ou le projet lié ; un membre publie son propre article.

**Priorités de réalisation (validées par le propriétaire le 2026-11-11) :** P0 lire et publier des articles et tutoriels (Markdown, code, tags, niveau, langue) avec liens vers questions/API/projets, et série de formations gratuites lisible publiquement ; P1 séries de formation structurées avec progression, mises en avant d'auteurs, formats plus riches ; P2 certificats, exercices interactifs. Le contenu original reste identifiable selon sa langue (FR/EN), sans traduction automatique implicite.

**Ajouts au P0 (amendement 3.2, proposé) :** thèmes réorientés (14 thèmes), bouton « Poser une question sur cette leçon » et fils par thème, sans nouvelle table. La série gratuite est une courte série originale, chargée en préproduction et en production par un script unique revu en PR et lancé par le propriétaire ; aucun contenu tiers n'est importé. Les salons en direct sont **conditionnels** au point d'étape du samedi 2026-11-14 à 12:00 GMT ; sans feu vert, ils restent **P2** et sont présentés comme non construits.

## 6. Quatre expériences différenciantes à démontrer ou préparer

1. **African Integration Intelligence** — relier fournisseurs et SDK par pays aux exemples, limitations, sources et retours d'intégration.
2. **Problem-to-Project** — transformer une question en connaissance réutilisable puis en besoin de contribution/projet, avec liens explicites.
3. **Cross-Border Developer Collaboration** — faire découvrir une collaboration selon les compétences au-delà des langues et pays ; UI bilingue dès le début et traduction assistée ultérieure.
4. **Evidence-Based Technical Knowledge** — mettre en avant versions, sources, reproductibilité, confirmations et date de mise à jour plutôt que les votes seuls.

Ces expériences résultent des fonctionnalités ci-dessus ; ce ne sont **pas** quatre nouveaux produits à construire. Le partage des connaissances (domaine LEARN) les nourrit : un article ou un tutoriel s'appuie sur les mêmes sources, versions et liens contextuels.

## 7. Parcours utilisateurs et critères d'acceptation

### A — Poser et résoudre une question

**Étant donné** un utilisateur connecté, **quand** il publie une question avec titre, description et bloc de code, **alors** la question apparaît dans le flux et la recherche ; un autre utilisateur peut répondre et voter ; l'auteur peut accepter une réponse ; les droits de modification sont appliqués côté serveur.

### B — Trouver une API locale

**Étant donné** une fiche sourcée, **quand** l'utilisateur choisit un pays et une catégorie, **alors** il voit les ressources correspondantes, leurs pays supportés et les liens de documentation, sans fausse promesse d'intégration officielle.

### C — Collaborer sur un projet open source

**Étant donné** un projet référencé par URL GitHub, **quand** son responsable publie un besoin de contribution, **alors** un autre membre peut manifester son intérêt et le responsable peut consulter la demande. Une même candidature ne doit pas être créée plusieurs fois accidentellement.

### D — Naviguer entre les domaines

**Étant donné** une question liée à une API et un projet, **quand** un membre consulte l'un de ces contenus, **alors** il peut atteindre les autres sans chercher leur URL à la main.

### E — Participer au-delà des frontières

**Étant donné** deux membres de pays/langues différents, **quand** ils consultent 226os, **alors** l'interface peut être affichée en FR ou EN, tandis que le contenu original reste identifiable. Les traductions automatisées ne sont pas implicitement promises pour le MVP.

### G — Apprendre et publier un contenu (VALIDÉ par le propriétaire le 2026-11-11)

**Étant donné** un visiteur, **quand** il ouvre un article, un tutoriel ou la série de formations gratuites, **alors** il peut le lire sans compte ; **étant donné** un membre connecté, **quand** il publie un article avec titre, contenu Markdown, tags, niveau et langue, **alors** il apparaît dans Apprendre et la recherche, avec ses liens vers questions, API ou projets. Un visiteur non connecté ne peut pas publier.

### F — Réseau limité

**Étant donné** un téléphone et un réseau simulé lent, **quand** les pages principales chargent, **alors** elles restent utilisables, affichent les états d'attente et ne dépendent pas de médias lourds pour accomplir les tâches essentielles.

## 8. Priorisation de livraison CADEV (sans réduction de la vision)

| Domaine | Démonstration CADEV / P0 | Suite prévue (P1/P2) |
|---|---|---|
| Solve | Q&A, snippets, réponse acceptée, votes, recherche | Réputation, discussions et revue de code avancée |
| Learn (VALIDÉ par le propriétaire le 2026-11-11 ; mis en avant, amendement 3.2 proposé) | Lecture et publication d'articles et tutoriels, série de formations gratuites lisible publiquement ; thèmes réorientés, « Poser une question sur cette leçon » et fils par thème (3.2) | Séries structurées avec progression, mises en avant d'auteurs, formats riches ; certificats et exercices interactifs (P2) ; salons en direct conditionnels au point d'étape du 2026-11-14, sinon P2 (3.2) |
| Discover | Catalogue API/SDK, fiches sourcées, pays, catégories | Comparaison, feedback et vérifications |
| Build | Projets GitHub, besoins, candidatures et liens | Matching et synchronisation avancée |
| Identity & Community | Auth, profils, réglage de visibilité du profil (3.2) et identité reliée aux contenus | Réputation, fil, communautés, découverte enrichie des membres |
| Connect (mis en avant, amendement 3.2 proposé) | Annuaire des membres en lecture seule (3.2) ; autres parcours affichés « bientôt » | Mentorat, carrière, opportunités, groupes d'étude, espaces communautaires |
| Language | Interface FR/EN | Traduction des contenus, recherche translingue |
| Access | Responsive et optimisation réseau vérifiable | Cache avancé, brouillons offline |
| Intelligence | Recherche PostgreSQL + liens contextuels ; assistant sourcé, déplacé de P2 à P0 (3.2), désactivé par configuration tant que la décision sur l'offre du fournisseur n'est pas prise | Suggestions, mode « indice d'abord » et demande de la tentative (P1) ; sémantique et résumés (P2) |

**Règle de livraison :** viser des flux complets et fiables plutôt qu'une succession d'écrans non fonctionnels. L'équipe peut augmenter le périmètre selon l'arrivée des renforts sans réécrire ce PRD.

## 11. Écrans et navigation

| Zone | Pages/écrans attendus |
|---|---|
| Explore | Accueil, recherche unifiée (avec le panneau de l'assistant, 3.2), contenus et projets à découvrir |
| Solve | Questions, détail de question, poser une question (avec le panneau de l'assistant, 3.2), réponses |
| Learn | Apprendre : liste d'articles/tutoriels/formations, lecture, publier un article, fils par thème et « Poser une question sur cette leçon » (3.2) |
| Discover | Annuaire d'API/outils, filtres, fiche ressource, liens documentaires |
| Build | Annuaire de projets, détail, publier, besoin de contribution, candidature |
| Identity | Inscription, connexion, profil public, paramètres (dont la visibilité du profil, 3.2) |
| Community / Connect | Espaces, activité, mentorat et opportunités — interfaces à réaliser progressivement selon roadmap. Annuaire des membres au P0 (3.2) ; le reste est affiché « bientôt » |

**Navigation (amendement 3.2, proposé) :** Apprendre et Connect sont mis en avant dans la navigation et sur l'accueil. Explore, Solve, Discover, Build et Identity restent accessibles depuis la même navigation, sur ordinateur comme sur mobile. Aucune zone n'est retirée ; une zone non construite est signalée comme « bientôt », jamais présentée comme disponible.

Le design system doit être cohérent sur l'ensemble des domaines : navigation partagée, états chargement/erreur/vide, composants de code, tags, pays et technologies, mode mobile et FR/EN.

## 12. Organisation de l'équipe 7's

**Situation actuelle (mise à jour 2026-11-12, amendement 3.2) :** 4 personnes inscrites à CADEV : le propriétaire et responsable du projet, deux développeurs backend et une personne frontend. La répartition par zone de travail est dans [ADR-0008](decisions/0008-team-work-areas-and-ownership.md) et [docs/team/CONTRIBUTORS.md](team/CONTRIBUTORS.md) (logins GitHub uniquement). Les responsabilités ci-dessous sont des **chantiers**, non des affectations imposées aux individus.

| Chantier | Responsabilités | Mode à 3 | Mode à 4–5 |
|---|---|---|---|
| Architecture & Platform | Go, PostgreSQL, recherche, auth, Docker, Traefik, CI/CD | Un référent, soutien transversal | Peut se scinder backend / infra-QA |
| Frontend & Experience | React/TanStack, UI, design system, mobile, FR/EN | Un référent | Renfort UX/intégration |
| Community & Knowledge | Solve, contenu, réponses, recherche, qualité des parcours | Un référent ou partagé | Référent domaine dédié |
| Open Source & Collaboration | Build, GitHub, besoins, candidatures, projets | Partagé entre frontend/backend | Renfort domaine dédié |
| Developer Ecosystem | Discover, API Hub, données sourcées, Connect et Intelligence selon priorité | Partagé | Renfort contenu/intégrations |

**Rôles individuels non affectés :** Yanis Axel DABO est capitaine/porteur ; Abdoul Aziz KABORE est membre connu ; le troisième membre n'est pas nommé ici ; renforts non confirmés. Ne pas supposer de spécialités sans échange avec eux. Prévoir owner et reviewer pour chaque domaine, intégration régulière et décisions consignées dans le dépôt.

Note (2026-11-12) : l'équipe compte désormais quatre membres ; le propriétaire met à jour les noms de cette section.

## Traçabilité produit
Toute user story reçoit un identifiant `US-<DOMAINE>-NNN`, une priorité P0/P1/P2, un scénario BDD et un critère d'acceptation dans [SPEC.md](SPEC.md). Les décisions d'architecture ne changent jamais le périmètre sans un amendement versionné du PRD et un accord humain (exemple : amendement 3.1, pilier Apprendre).

## Non-objectifs de CADEV
Ne pas présenter comme terminés les fonctions P1/P2 non implémentées. Les autres domaines restent visibles dans la roadmap sans écrans factices. Le concours exige un produit original, une démo publique et un déploiement Datacloud : voir [SHIP.md](SHIP.md).
