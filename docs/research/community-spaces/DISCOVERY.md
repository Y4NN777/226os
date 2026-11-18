# Espaces et salons de développeurs : cadrage (brouillon, 2026-11-18)

Statut : **PROPOSÉ**, à valider par le propriétaire. Rien n'est construit. Sources : `docs/PRD.md` 5.4 (« Communautés/espaces par pays, technologies et centres d'intérêt », « Fil d'activité, publications et échanges thématiques », P2), `ROADMAP.md` S10 (« salons en direct » par projet ou série, NON CONSTRUIT, P2) et risque K17 (charge de modération), `docs/SPEC.md` (Connect annonce les espaces « bientôt », aucun espace fictif), `DESIGN.md` (sobriété réseau, interface honnête).

## Cadrage

- **Pour qui.** Les membres de 226os, d'abord au Burkina Faso puis en Afrique de l'Ouest (PRD, amendement 3.3) ; les porteurs de projets open source et les auteurs de formations, qui veulent un lieu de discussion attaché à leur projet ou à leur série.
- **Le travail à faire.** Se retrouver entre pairs d'un même pays, d'une même technologie ou d'un même intérêt ; poser une question rapide ou partager une ressource sans ouvrir un fil formel dans Résoudre ; suivre la vie d'un projet ou d'une formation.
- **Ce qu'ils font aujourd'hui (hypothèse, à vérifier).** Des groupes WhatsApp et Telegram par communauté ou par école, des serveurs Discord pour certains projets. L'information s'y perd : c'est le problème 1 du PRD (« connaissances dispersées entre forums, dépôts, messageries »).
- **Pourquoi maintenant.** Le P0 est en production ; Connect a désormais un annuaire réel (membres et développeurs GitHub classés). Les espaces sont l'étape P2 annoncée « bientôt » dans l'interface.
- **Contraintes.** Un seul serveur (3,8 Go de RAM, mesuré par `free -h` sur le serveur de production le 2026-11-15) ; utilisateurs sur mobile et connexions lentes ; modération déjà construite (masquer / rétablir, S6) ; contrat d'API d'abord (OpenAPI) ; FR/EN ; pas de dépendance nouvelle sans revue de licence.

## Trois approches

| | A. Espaces asynchrones (fils de discussion) | B. Salons en direct (S10) | C. Annuaire de groupes externes |
|---|---|---|---|
| **Pari** | Les échanges utiles se lisent et se retrouvent plus tard : un espace = des fils, cherchables, modérés, légers | L'énergie vient de l'instantané : les gens restent si ça bouge en direct | Les communautés existent déjà sur WhatsApp, Telegram ou Discord : 226os doit les rendre visibles, pas les remplacer |
| **Test le moins cher** | Ouvrir 3 espaces (Burkina Faso, Go, mobile money) avec des fils, mesurer messages par membre actif sur 2 semaines | Un salon par projet sur un seul projet pilote, mesurer les messages en direct par jour | Lister 10 groupes réels avec leur lien et leur responsable, mesurer les clics sortants |
| **Risque principal** | Ressembler à un forum vide (« ville fantôme ») | Coût serveur (connexions persistantes), charge de modération en temps réel (K17), mauvais réseau mobile | L'information reste dispersée hors de la plateforme ; dépendance à des outils tiers |

## Critères

1. Répond au problème 1 du PRD (connaissance retrouvable, pas dispersée).
2. Tient sur le serveur actuel et sur un réseau mobile lent.
3. Modérable avec les outils existants (masquer / rétablir, signalements).
4. Effort de construction raisonnable (contrat, migration, API, web, tests) et réutilisation de ce qui existe (Résoudre, recherche, modération, profils).
5. Pas d'espace fictif ni d'activité inventée (interface honnête).

## Recommandation

**A d'abord, avec un lien externe optionnel de C, et B plus tard au-dessus de A.**
- A coche les critères 1 à 4 : un espace est une liste de fils (titre, message, réponses), cherchable et modérable comme Résoudre ; tout passe en requêtes HTTP courtes.
- C se greffe à moindre coût : un espace peut afficher le lien de son groupe WhatsApp, Telegram ou Discord officiel, déclaré par ses responsables, sans copier son contenu.
- B (le direct) devient une amélioration de A : rafraîchir un fil ouvert par Server-Sent Events quand le serveur et la modération le permettent. Il n'est pas dans la première version.
- **Ce qui ferait changer d'avis :** si, après 2 semaines, les fils d'un espace pilote restent sous un message par jour alors que son groupe externe est actif, privilégier C ; si les membres demandent explicitement le direct et que la charge le permet, avancer B.
