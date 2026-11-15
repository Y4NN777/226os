# Prototype fidelity, 2026-11-15

Owner decision (2026-11-15, 05:10 GMT): every page must match its HTML prototype in `prototypes/html/pages/` (layout, density, hierarchy, badges, states), in light and dark, on desktop and phone. This report records how that was checked and what is left.

## Method
- Captures: `src/web/scripts/capture-fidelity.sh` takes each app route next to its prototype page, in French, at 1280×800 and 375×812, in light theme, and also in dark for Découvrir, Résoudre and Construire. The app was the branch `feature/web/prototype-fidelity` served by Vite against the local-dev API with the demo seed.
- Judge: an independent reviewer that had not built the pages compared each pair. It gave a verdict per page (MATCH, CLOSE or OFF) and the visible differences.
- Not counted as differences, per `prototypes/html/DISCLAIMER.md`:
  - different data or people (the prototype is fictional, the app shows the demo seed);
  - star ratings, counters, sort and view toggles, tabs and fields the API does not serve;
  - navigation order and labels (O7);
  - wording, shortened on purpose ([`COPY_STYLE.md`](../team/COPY_STYLE.md));
  - prototype scaffolding;
  - the privacy and code-of-conduct links, because those documents do not exist.
- Not run: a pixel diff, signed-in captures (member home, own profile, settings forms, the ask view), and dark mode for detail pages.

## Verdicts

| Page | First pass, 1280 / 375 | Final pass, 1280 / 375 | Final, dark 1280 / 375 |
|---|---|---|---|
| Découvrir | CLOSE / CLOSE | CLOSE / CLOSE | CLOSE / CLOSE |
| Fiche | CLOSE / CLOSE | CLOSE / CLOSE | not captured |
| Résoudre | CLOSE / OFF | CLOSE / CLOSE | CLOSE / CLOSE |
| Question | CLOSE / OFF | CLOSE / CLOSE | not captured |
| Construire | CLOSE / CLOSE | CLOSE / CLOSE | CLOSE / CLOSE |
| Projet | CLOSE / CLOSE | CLOSE / CLOSE | not captured |
| Apprendre | CLOSE / CLOSE | CLOSE / CLOSE | not captured |
| Recherche | OFF / OFF | MATCH / MATCH | not captured |
| Connexion | CLOSE / CLOSE | MATCH / MATCH | not captured |
| Inscription | CLOSE / CLOSE | CLOSE / CLOSE | not captured |
| Connect | not judged | not comparable: the repository prototype is the outdated « Mentorat bientôt » page | not captured |
| Landing | no prototype | the judge found it finished at both widths | not captured |

The first pass ran before the last round of fixes, which covered:
- control heights of 44 px, and 52 px for search fields;
- the search icon inside the field;
- smaller field labels;
- Résoudre and the question page at 375 px;
- the search page empty state;
- the phone header.

## What is left, and why
- **Data the API does not serve yet.** These come with #238, followed by the web wiring:
  - project stage, licence and country;
  - result counts;
  - sort controls;
  - profile activity tabs.
- **Data or features out of scope.**
  - The Discover open-source strip and SDK/language filter.
  - The fiche's source table with publication dates, and its tabs.
  - The project's « Contexte lié » cards and repository card.
  - Build's skill filter chips.
- **Kept on purpose.**
  - Breadcrumbs on every page; on a phone, a page three levels deep shows only a « ‹ Section » back link.
  - The region and type filters on Découvrir.
  - The display name and username fields on sign-up.
- **Small visual differences.**
  - Solve tags are mono chips, where the prototype uses plain `#tag`.
  - Build monogram tiles are 48 px instead of 72 px, to match Découvrir.
  - Auth controls are 44 px instead of 48.
  - Auth links are not underlined.
  - On a phone, Construire has no bottom-bar item, so nothing is highlighted on that page.
  - The Apprendre « Langue » select is clipped at 1280 px.
- **Vote column.** The signed-out capture shows an up arrow only. The contract supports up and down votes, and members get both buttons.

## Evidence
- Captures, 58 files: [`evidence/fidelity/2026-11-15/`](evidence/fidelity/2026-11-15/), named `<page>-<app|proto>-<width>-<theme>.png`. The landing has app captures only.
- 2026-11-15, after #256: the 30 app captures were taken again from the branch `fix/web/landing-shell`, with the same script, routes and sizes, served by Vite against the local-dev API. They show the positioning badge removed, the anonymous landing without a sidebar, and the sidebar collapse button. The prototype captures and the verdicts above are unchanged. Later the same day, after `APP_ENV=local-dev make seed` refreshed the demo text (#241), the app captures were taken again from `develop`. Discover now shows the current names and short summaries; Build and the demo people keep « (démo) » because those pages show no Démo badge.
- Web gates on the branch at merge time: typecheck, lint and format:check PASS; Vitest PASS, 48 files and 646 tests.
- ADR-0007 section 8 of [`ASTRYX_REPORT.md`](ASTRYX_REPORT.md) links these captures for its theme and flow rows.
