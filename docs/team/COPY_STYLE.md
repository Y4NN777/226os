# 226os copy style guide (FR first, EN second)

Owner-approved on 2026-11-15 (decisions: « Paramètres », shorter filter labels, no Connect spaces teaser text, « Déclaré », no « (démo) » suffix where the Démo badge shows). Every brief that writes UI copy points here. Related: [`OWNER_PREFERENCES.md`](OWNER_PREFERENCES.md).

The test of every string: would a francophone developer write it like this in a pull request? If a sentence can go and the screen still works, it goes.

## Voice
- **Vouvoiement, everywhere.** The UI already uses it (50 strings use « votre / vos »). The audience runs from students to seniors in a professional setting across francophone Africa, and the Microsoft fr-FR guide uses « vous » too. One form across the product (owner rule). Use fewer pronouns: infinitive buttons and nominal labels (« Langue », « Message ») remove most of the « vous / votre ».
- Direct and calm. No marketing (« une seule plateforme », « gratuitement »), no exclamation marks, no « simplement / facilement / rapidement », no « n’hésitez pas », « veuillez », « Soyez le premier », « Voici », « Ici ».
- Avoid « il faut », « il y a », « c’est » and « on » as sentence openers (Microsoft fr-FR §2.1.4). Avoid formal verbs: « nécessiter » becomes « devoir », « avoir la possibilité de » becomes « pouvoir ».
- Neutral without median dots: « Vous uniquement », not « Vous seul » ; « Connexion à 226os », not « Content de vous revoir ».

## Length limits (FR words; EN about the same)
| Element | Limit | Rule |
|---|---|---|
| Page title / h1 | ≤ 4 words | A noun or an infinitive. No final period. Landing hero only: ≤ 8 words. |
| Lede under a title | 0 by default | Allowed only where the title does not name the content: the landing hero, and the Découvrir, Construire and Connect pages. 1 sentence, ≤ 10 words. Never a lede that restates the title or the buttons. |
| Button | ≤ 3 words | Infinitive verb + object: « Publier la question », « Rechercher », « Effacer les filtres ». Sign-in gates may say « Se connecter pour + verbe ». |
| Link | ≤ 4 words | « Voir les projets », « Tous les articles », « Retour aux questions ». |
| Field label | ≤ 3 words | A noun. Drop the form’s context: « Langue », not « Langue de la question ». |
| Hint | ≤ 10 words | A format or an example: « De 1 à 5, séparées par des virgules. Ex. go, cors ». No advice prose. |
| Empty state | 1 line + 1 action | Title « Aucun X » or « Aucun X ne correspond ». Optional body ≤ 8 words that says what to do. No « pour le moment / pour l’instant ». |
| Error | ≤ 12 words | What happened + what to do: « Lien non valide. Retirez-le. » No parentheses (Microsoft §5.5.1). Ends with a period. |
| Failure title | ≤ 5 words | « Impossible de + infinitif » (« Impossible d’enregistrer le nom »). |
| Notice / toast | ≤ 4 words | « Candidature envoyée. », « Nom enregistré. » |
| aria-label | ≤ 5 words | A verb for an action (« Voter pour la question »), a noun for a region. |

## Rules
1. No explanatory text for self-evident UI. A title, a list and a button explain themselves.
2. A note never repeats a button. If the button says « Saisir un autre code », the body does not say « Réessayez avec un autre code ».
3. One idea per sentence and at most two sentences per string. A second sentence is allowed only for an error (what happened + fix), a hint ending in « Ex. … », or a safety or privacy notice. Safety notices (rule 6) may run to 14 words. On 2026-11-15 the rewrite brought the strings with two or more sentences from 48 to 34 and those of 12 or more words from 44 to 10; every remaining one is an error, a hint with an example, or a safety notice.
4. Sentence case in both languages: capitalize only the first word and proper nouns (« Créer un compte », « Create an account »). The same rule applies in Primer, the GC design system and Microsoft fr-FR.
5. Constraints are given as facts, not prose: « 2 000 caractères maximum », « De 1 à 5 étiquettes : minuscules, chiffres, - + . # ».
6. Safety and privacy facts stay, just shorter. These include the assistant disclosure (Google AI Studio; only the question and public content are sent), « jamais exécuté », « tous vos appareils seront déconnectés », the anti-enumeration conditionals (« Si un compte existe… », « Si l’inscription est valide… ») with « 30 minutes », and « Votre e-mail n’est jamais public ».
7. No « (démo) » in names. The `state.demo` badge says it. The suffix lives in `db/seeds/00003_discover.sql` and `00004_build.sql`.
8. Typography (FR, enforced by `translate.test.ts`): use U+00A0 before `: ; ? !`, inside « », between a number and its unit (`{seconds} s`), and in thousands (`2 000`). Use ’ for apostrophes and … for ellipses. EN: no NBSP; use “ ” and ’.
9. Calques to avoid: « en ligne » for *live*, « afin de », « au sein de », « assurez-vous de », « Lancer la recherche » (use « Rechercher »), « Tout apprendre », « Voir en tant que ». In EN, avoid « Search on 226os » and « Everything to learn ».
10. One word per concept (glossary below). Never two labels for the same thing.

## Glossary (FR / EN)
| Use | EN | Notes |
|---|---|---|
| Apprendre, Connect, Résoudre, Découvrir, Construire | Learn, Connect, Solve, Discover, Build | Pillar names, used in nav and titles only. Connect stays in English (proper noun). |
| fiche | listing | An entry in the Discover catalogue. Replaces « API ou outil » as a content kind. |
| catalogue | catalogue | The set of fiches. |
| question, réponse, réponse acceptée | question, answer, accepted answer | Question states: « Sans réponse », « Avec réponse », « Résolue ». |
| étiquette | tag | For questions and articles. Replaces « mot-clé ». |
| article, tutoriel, formation, leçon | article, tutorial, course, lesson | A formation is a series. Replaces « série » and « parcours ». |
| projet, besoin, candidature, postuler | project, need, application, apply | « Besoins » alone on the project page. |
| membre, profil, annuaire | member, profile, directory | Connect is the annuaire. |
| source, Vérifié, Déclaré, À revérifier, Contesté | source, Verified, Claimed, Needs recheck, Disputed | Status badges. |
| Paramètres | Settings | Replaces « Réglages » to match `/parametres`. |
| Effacer les filtres | Clear filters | Replaces « Réinitialiser les filtres ». |
| Afficher plus | Show more | Used in every list. |
| Contenus liés | Related | Replaces « À voir aussi », « Liens associés » and « Contenu lié ». |
| Se connecter, Créer un compte, Se déconnecter | Sign in, Create an account, Sign out | |

Words to avoid: pour le moment, pour l’instant, momentanément, simplement, facilement, gratuitement, n’hésitez pas, veuillez, Soyez le premier / la première personne, Voici, Ici, il faut, Lancer la recherche, Réinitialiser, Réglages, Série, Mot-clé, Prêt à…?

## Sources
Read directly:
- Microsoft, *French (France) Localization Style Guide*, §2.1 voice (« clear, friendly, and concise »), §2.1.4 words to avoid (« il faut », « il y a », « c’est », « on », formal verbs), §4.1.5 capitalization, §5.5.1 error messages (period at the end, avoid parentheses, standard « Impossible de… »). Use of « vous » is from the sample text, p. 13. https://download.microsoft.com/download/c/7/9/c7921dbd-4531-4a4e-8490-e4656d739f3d/fra-fra-StyleGuide.pdf. Note that §2.1.4 discourages « Impossible de » while §5.5.1 makes it the standard phrase. 226os keeps it for failure titles because it is shorter.
- GitHub Primer, Content: sentence case, no punctuation in labels and buttons, cut adjectives and adverbs, never say « easy », « quick » or « just ». https://primer.style/foundations/content
- Google developer documentation style guide, Voice and tone: no « please », « simply », « easy » or « quickly », and no exclamation points. https://developers.google.com/style/tone
- Atlassian, Voice and tone: « Tell people only what they need to know in the moment and nothing more. » https://atlassian.design/foundations/content/voice-tone
- Mailchimp, Writing for accessibility: short sentences, descriptive links, important information first. https://styleguide.mailchimp.com/writing-for-accessibility/
- Système de design GC (Canada), Bouton: short and precise labels, a verb alone or verb + noun (« Soumettre », « Obtenir une estimation »), lowercase after the first letter. https://systeme-design.canada.ca/fr/composants/bouton/design
- OQLF, Principes de rédaction pour le Web: short active sentences, one idea per sentence, verb forms over noun forms. https://vitrinelinguistique.oqlf.gouv.qc.ca/25035/banque-de-depannage-linguistique/la-redaction-et-la-communication/redaction-pour-le-web/principes-de-redaction-pour-le-web
- Real French product strings: Mastodon `fr.json` (« Se connecter », « Créer un compte », « Réessayer », « Charger plus », « Aucun résultat. », « Chargement… »). Its empty states are two-sentence examples of what not to do. https://github.com/mastodon/mastodon/blob/main/app/javascript/mastodon/locales/fr.json. Discourse `client.fr.yml` (« Mot de passe oublié ? », « Afficher plus », « Page introuvable », « Répondre », « Étiquettes »). https://github.com/discourse/discourse/blob/main/config/locales/client.fr.yml

From search results only (the pages now redirect to shopify.dev and were not read):
- Shopify Polaris error messages and empty state: say what is wrong and what to do, be specific, use one primary action. https://polaris.shopify.com/patterns/error-messages, https://polaris.shopify.com/components/layout-and-structure/empty-state

Not reached, so no rules are attributed to them:
- DSFR (systeme-de-design.gouv.fr) returned HTTP 403.
- The modernisation.gouv.fr forms kit returned HTTP 403.
- Apple HIG Writing returned only its title.
- The Scaleway, OVHcloud, Qonto, Alan and Doctolib consoles were not checked.
