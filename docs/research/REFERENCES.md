# 226os — Références externes à étudier

**Usage :** comparaison de comportements, UX, architecture, licences et tests. Ne pas forker une application complète et prétendre qu'une réécriture IA efface les droits. Déclarer toute dépendance intégrée et vérifier ses licences transitives à version épinglée.

| Projet | Référence | But de l'étude | Licence du dépôt (à revérifier sur commit choisi) |
|---|---|---|---|
| Apache Answer | https://github.com/apache/answer | Solve : Q&A, votes, accepted answer, réputation, workflow | Apache-2.0 |
| Apache Answer Plugins | https://github.com/apache/answer-plugins | extensibilité | par sous-repo |
| Flarum Framework | https://github.com/flarum/framework | discussion, profils, permissions, UX | MIT |
| Nuxt HackerNews | https://github.com/nuxt/hackernews | navigation rapide, feed sobre | vérifier |
| 225 Open Source | https://www.225os.com/ | découverte des projets et positionnement | site tiers (non source à réutiliser) |
| TanStack Router | https://github.com/TanStack/router | route typée | MIT |
| TanStack Query | https://github.com/TanStack/query | synchronisation / cache serveur | MIT |
| TanStack Form | https://github.com/TanStack/form | validations formulaire | voir THIRD_PARTY_NOTICES.md |
| Astryx (Meta) | https://github.com/meta/astryx | UI principale (dépendance `@astryxdesign/core`) | MIT (voir THIRD_PARTY_NOTICES.md) |
| shadcn/ui | https://github.com/shadcn-ui/ui | UI accessible et composable | vérifier |
| Chi | https://github.com/go-chi/chi | API/middlewares Go | voir THIRD_PARTY_NOTICES.md |
| pgx | https://github.com/jackc/pgx | PostgreSQL Go | voir THIRD_PARTY_NOTICES.md |
| sqlc | https://github.com/sqlc-dev/sqlc | SQL Go typé | voir THIRD_PARTY_NOTICES.md |
| Traefik | https://github.com/traefik/traefik | TLS/edge/routing | voir THIRD_PARTY_NOTICES.md |
| GitHub Actions | https://github.com/actions/starter-workflows | CI workflows | workflows à adapter |

## Format obligatoire d'une étude
```yaml
repository: "https://github.com/owner/repo"
commit_sha: "..."
license_file: "..."
license_spdx: "..."
features_examined: []
source_files_and_lines: []
tradeoffs: []
behaviors_for_226os: []
third_party_code_integrated: false
compliance_review: "CLOSED par le propriétaire le 2026-11-13 (sans revue indépendante)"
```
Chaque proposition d'amélioration issue d'une référence inclut : (1) le problème **observé**, (2) ses sources et références précises, (3) une expérience ou métrique capable de prouver l'amélioration.

## Risques CADEV
Le concours exige une création pour l'essentiel originale pendant la période. Les composants sous GPL/AGPL/LGPL sont prohibés par son règlement tel que communiqué. L'utilisation d'un projet complet comme base, même réécrit en Go/TS, nécessite vérification d'admissibilité et de titularité. Les crédits et obligations de licence restent applicables.

## Étude : 225 Open Source (référence d'inspiration)

**Statut :** étude du 2026-11-12, **à compléter par l'équipe**. Source : le texte public de la page d'accueil, copié par le propriétaire le 2026-11-12 (cette copie n'est **pas** conservée dans le dépôt). Les pages « CGU » et « Confidentialité » du site n'ont pas été lues : à vérifier. Le contenu ci-dessous est paraphrasé ; aucun texte du site n'est repris.

```yaml
repository: "https://www.225os.com/ (site tiers ; aucun dépôt de code consulté)"
commit_sha: "sans objet (aucun code consulté ni intégré)"
license_file: "aucune licence de code indiquée sur la page d'accueil ; CGU non lues"
license_spdx: "sans objet : aucun code, texte, visuel ni donnée du site n'est repris"
date_consultation: "2026-11-12"
features_examined:
  - "Répertoire national de projets open source ivoiriens hébergés sur GitHub, filtre par pays, statistiques synchronisées depuis l'URL GitHub déclarée par l'auteur."
  - "Tableau de lancement de produits (SaaS, API, applications mobiles, modèles) avec votes de la communauté."
  - "Mise en relation entre une idée et un cofondateur technique, un mentor ou un investisseur."
  - "Articles, tutoriels et retours d'expérience en français, écrits par la communauté."
  - "Communauté via Discord et WhatsApp ; comptes (« Connexion »)."
  - "Positionnement : souveraineté numérique ivoirienne, gratuit et sans commission ; le site renvoie vers les dépôts des auteurs."
source_files_and_lines: []
tradeoffs:
  - "Leur périmètre est national et centré sur la visibilité des projets locaux. Le nôtre commence par l'Afrique de l'Ouest, puis s'ouvre à toute l'Afrique, et couvre l'entraide, la découverte d'API, l'apprentissage et les personnes."
behaviors_for_226os:
  - "Inspiration déclarée dans docs/PRD.md : découverte de projets open source et identité communautaire."
third_party_code_integrated: false
third_party_text_visuals_or_data_copied: false
compliance_review: "CLOSED par le propriétaire le 2026-11-13 : confirmation consignée ci-dessous, sans revue d'admissibilité indépendante"
```

### Comparaison honnête (observé sur la page d'accueil et notre périmètre)

« Non mentionné » signifie : absent de la page d'accueil lue, pas absent du site.

| Sujet | 225 Open Source (observé) | 226os |
|---|---|---|
| Portée | Côte d'Ivoire | Afrique de l'Ouest d'abord, puis toute l'Afrique ; interface FR/EN |
| Projets GitHub | Répertoire par pays ; l'auteur déclare son URL GitHub | **Ressemblance reconnue** : Build permet aussi de publier un projet par URL GitHub, avec un pays. Notre apport : besoins de contributeurs, candidatures et liens vers questions, ressources et articles |
| Tableau de lancement de produits | Oui | Hors périmètre du concours ; pas une priorité |
| Cofondateur, mentor, investisseur | Oui | Non au P0 : le mentorat reste P1/P2 et s'affiche « bientôt » |
| Articles et tutoriels | Oui, en français | **Ressemblance reconnue** : Learn publie aussi des articles et une série ; notre apport : sources, liens vers questions, API et projets, FR/EN, un assistant d'apprentissage |
| Entraide technique avec réponse acceptée | Non mentionné | Solve (P0) |
| Annuaire d'API et d'outils avec sources et pays couverts | Non mentionné | Discover (P0) |
| Assistant qui cite ses sources | Non mentionné | P0 |
| Annuaire de personnes avec visibilité par champ | Non mentionné | Connect (P0 minimal) |
| Chat communautaire | Discord et WhatsApp | Non prévu (canaux en direct : conditionnel, sinon P2) |

### Garde-fous pour rester original

- **Noms et formulations :** ne pas reprendre les noms de leurs rubriques ni leurs slogans ; ne pas reprendre l'argumentaire « GitHub n'a pas de filtre par pays » comme promesse centrale.
- **Aucune reprise** de texte, de structure de page, de visuels, de palette, de marque ni de données. Pas de capture de leur site dans le dépôt.
- **Pas d'ingestion** de leurs données ni de leur contenu.
- **GitHub :** l'import par URL repose sur l'API publique de GitHub et sur l'URL fournie par l'auteur du projet ; il ne dépend pas du site de référence.
- **Règlement du concours :** l'article 9 précise que les idées, concepts et fonctionnalités ne sont pas appropriables en tant que tels, et l'article 10 vise le plagiat, c'est-à-dire la copie de code ou de contenu. Notre démarche porte sur les comportements, pas sur la copie.
- **À faire par l'équipe :** lire leurs CGU et compléter cette étude avec les pages vues.

### Constats du propriétaire (consignés le 2026-11-12)

- Le nom **226os** se lit **« 226 Open Stack »** ; 226 est l'indicatif téléphonique **+226 du Burkina Faso**.
- Positionnement : **Afrique de l'Ouest d'abord, puis toute l'Afrique**.
- Les références servent d'**inspiration** : comportements et principes étudiés, jamais de code, de texte, de visuels, de marque ni de données copiés.
- Aucun contact avec 225 Open Source n'est jugé nécessaire.
- Le propriétaire a approuvé cette démarche et confirme que le projet respecte la règle de création pendant la période du concours (à joindre à la revue CADEV, **non vérifié par le dépôt**).

## Déclaration d'originalité

*Confirmation : déclaration confirmée par le propriétaire du dépôt (`Y4NN777`) le 2026-11-13.*

**Français.** 226os est un produit original, développé par notre équipe pendant la période du concours, sous licence Apache-2.0. Nous citons ouvertement 225 Open Source comme une source d'inspiration pour la découverte de projets et l'identité communautaire, comme les autres projets open source que nous avons étudiés. Nous avons étudié des comportements et des principes de conception, sans copier de code, de texte, de visuels ni de données. Chaque dépendance tierce est déclarée, épinglée et vérifiée côté licence. Notre orientation produit nous est propre : une plateforme qui part de l'Afrique de l'Ouest et s'ouvre à toute l'Afrique, avec un assistant d'apprentissage qui cite ses sources, un parcours pour apprendre et progresser à l'ère des agents, un annuaire d'API et d'intégrations africaines, et une mise en relation qui respecte la vie privée.

**English.** 226os is an original product, built by our team during the competition period, under the Apache-2.0 licence. We openly cite 225 Open Source as an inspiration for project discovery and community identity, as we cite other open source projects we studied. We studied behaviour and design principles only. We did not copy code, text, visual assets or data, and every third-party dependency is declared, pinned and licence-checked. Our product direction is our own: a platform that starts from West Africa and opens to all of Africa, with a grounded learning assistant that cites its sources, a learn-and-grow path for the agentic era, an African API and integration directory, and privacy-respecting people discovery.
