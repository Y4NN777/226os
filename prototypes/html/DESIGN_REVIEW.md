# Design review: 226os HTML prototype

**Status:** self-review by the author agents, 2026-11-11. This is NOT an independent design or accessibility audit. What was verified and what was not is stated per section. Authority on conflict: `DESIGN.md` > `CONVENTIONS.md`.

## 1. Design decisions and rationale

| Decision | Choice | Rationale / source |
|---|---|---|
| Direction | Graphite & Lime, quiet neutral canvas (#F6F7F4 / white), green primary, orange sparing, lime only on deep-green/dark surfaces | DESIGN.md 2, 2.2 |
| Density | Technical, dense **list rows** (`.list > .row`), no stat tiles, no invented counters, cards sparingly | Developer audience, scannability on slow/mobile networks; CONVENTIONS 2 |
| Navigation model | One shell, injected by `shell.js` (single source of truth for 9 domain entries). >=1024px full sidebar 232px; 768-1023px rail 92px; <768px bottom bar (4 domains + "More" opening the drawer) | PRD 11 (shared nav across domains); thumb reach on phones (PRD 7F) |
| P0 / P1 / P2 representation | P0 with own pages = normal links. Connect = `badge-preview` "Parcours conçu (P1/P2)" / "Designed journey (P1/P2)" (PRD 8: P0 = documented scope and designed journey, the rest planned P1/P2; page exists, non-functional, content simulated). Language, Access, Intelligence = P0 delivered across all pages: link to `index.html#domains` with `badge-builtin` "Intégré/Built in" and a visible P0/P1 note (drawer + sidebar; rail keeps it visually hidden). Later features never styled as working; `badge-later` stays for P1/P2 items inside pages |
| FR/EN rule | French default; UI strings in dictionaries (`I18N.register`); **user content never translated**, carries `lang` + `data-user-content` and an original-language badge; switching language does not reset forms | SPEC US-IDENTITY-001, PRD 5.6/7E; `lang` visible per post |
| Evidence / verification display | Accepted answer, verified resource, sources and verification date shown as badge **with text/icon** plus `<time>`; source links `link-external`; no claim of official endorsement; all data fictional | DIFF-01/04 in TRACEABILITY |
| Authorization "view as" | Persona toggle (author / other member / anonymous) on question, profile, project pages to **demonstrate** server-enforced denials (403) and field visibility (profile). Clearly a prototype demonstrator | SPEC US-SOLVE-001 / US-BUILD-001 deny scenarios; PRD journeys A, C |
| Simulation honesty | `App.Sim.load` adds a "Simule/Simulated" badge; `?state=loading|empty|error` forces states; `data-sim` actions toast "simulated" | CONVENTIONS 6 |
| Typography / assets | System font stacks, inline CSS icon sprite, no remote fonts/images/flags, country as text code | DESIGN.md 4; US-ACCESS-001 (no heavy media) |
| No colour-only state | Every state has label/icon/shape | DESIGN.md 2.2 |

## 2. Coverage matrix

Pages: `index.html` (Explore) plus 9 lane pages `pages/{solve,question,discover,resource,build,project,profile,connect,auth}.html` (10 screens in total; `_template.html` is a copy-source, not a screen). States column: L loading, E empty, X error (all forced via `?state=`), plus form validation and simulated success where noted. "Present" means a page and markup exist, not that it was tested.

| PRD domain | CADEV scope (PRD 8) | Screens in prototype | States / notes |
|---|---|---|---|
| Explore | accueil, recherche unifiee | `index.html` | client-side text match over mock arrays; L/E/X via `App.Sim` |
| Solve | Q&A, snippets, accepted answer, votes, search | `solve.html` (list, filters, `#ask` form), `question.html` (detail, answers, vote, accept, tabs, linked items, answer form + preview) | L/E/X, form-summary, vote toggle (not persisted), accepted mark, code blocks |
| Discover | catalogue, sourced fiches, country/category | `discover.html` (filter bar, rows), `resource.html` (detail, tabs, dialog, sources, linked items) | filters (country, category, type, SDK / language, provider, doc language; `?sdk=`/`?provider=`); provenance links; stale/unknown honesty: shown only if mock data models it (not verified per record) |
| Build | projets GitHub, besoins, candidatures | `build.html` (directory, filters, publish form + dialog), `project.html` (needs, roles, application dialog, duplicate-submission simulation) | form-summary, dialog, duplicate application demo |
| Identity & Community | profils, identite liee aux contenus | `auth.html` (sign-in / create account, generic errors, 429, session-expired, verification-pending, forgot-password; all simulated), `profile.html` (tabs, view-as persona, privacy table, dialog, form) | owner / member / anonymous field visibility |
| Connect | perimetre documente et parcours concu | `connect.html` (designed journey, non-functional; rows tagged `badge-later` + `badge-sim`) | honest "coming later" |
| Language | interface FR/EN | shell toggle on all pages | persisted in `localStorage`, forms not reset (author statement; not re-tested after the last CSS/JS edits) |
| Access | responsive + reseau | all pages, bottom bar, container-query rows | see sections 4-5 |
| Intelligence | recherche PostgreSQL + liens contextuels | header search, `?q=` on Explore, `.linked-items` | AI/semantic NOT shown; nav item = "Intégré" (P0), suggestions P1 in the note |

| SPEC story | Where represented | Gap |
|---|---|---|
| US-SOLVE-001 (publish, answer, accept once, 403 for non-author, search) | `solve.html#ask`, `question.html` answer form, accept control, view-as 403 callout, search | "one accepted answer only" enforced client-side in mock only; real invariant is server/DB |
| US-DISCOVER-001 (filter by country + category, provenance, related discussions) | `discover.html`, `resource.html` | provenance depends on mock fields |
| US-BUILD-001 (interest in a need, no duplicate on second click, private applications denied) | `project.html` (dialog, duplicate simulation, view-as) | private list denial shown as simulated state |
| US-IDENTITY-001 (linked resource/project on question; FR/EN switch keeps content) | `.linked-items` on detail pages; language toggle; `lang` + `data-user-content` | none known; not machine-tested |
| US-ACCESS-001 (slow network, loading/empty/error, no heavy media) | `App.Sim` with skeletons; no media; byte counts below | no throttled-network test run |

Not represented: Identity settings screen listed in PRD 11 (sign-up/sign-in exist, simulated, in `auth.html`), publish-resource screen, offline drafts (P2).

## 3. Accessibility

**Approach:** semantic landmarks (`header`, `nav aria-label`, `main#main`, skip link), exactly one `h1` per page, labelled controls, `aria-current`, `aria-pressed` for toggles, `role=alert`/live regions for errors and toasts, native `<dialog>` with focus return, labelled scrollable regions for tables/code, visible 2px focus ring (blue, offset, distinct from green), 44px target aim (never <24px), `prefers-reduced-motion` collapses durations, `lang` on user content, `data-testid` on interactive/text elements.

**Verified (by me, this session):**
- Contrast: 10 token pairs recomputed with the WCAG formula and matched the audit comment in `tokens.css`: #4D721D/white 5.61, #576454/#F6F7F4 5.82, graphite/#C67C35 4.78, lime/forest 7.73, #6F7C6B/white 4.40, lime/dark surface 11.39, #7E8E75/dark surface 4.58, #C1CCB9/#2C3926 7.33, #8A540D/#FBEFD9 5.50. The full audit in `tokens.css` covers ~70 pairs; I did not recompute all. Pairs only: **not** rendered-pixel measurement, and not hover/active/disabled states or text over arbitrary backgrounds.
- Class/markup inventory by grep (counts of dialogs, tabs, form-summary per page).

**NOT verified (do not claim):** no axe/Lighthouse/WAVE run; no NVDA/JAWS/VoiceOver/TalkBack test; no keyboard walkthrough recorded in this review; no 200%/400% zoom test; no forced-colors/high-contrast check; no real-device touch testing; no FR/EN text-expansion check at 360px beyond author statements; focus-return of the native `<dialog>` on every browser unverified; keyboard behaviour of `[data-tabs]` (arrows/Home/End in `app.js`) read but not executed here. Screenshots were taken by lane agents (see CONVENTIONS 9), not reviewed here.

Known risks to test first: the "view as" and simulated-state areas replacing content (live-region announcements), stretched row links with nested links, `aria-disabled` nav items for "later" domains, code blocks `tabindex=0` scroll regions, 92px rail labels at 768px.

## 4. Responsiveness

Breakpoints: <640/<768 mobile (bottom bar, single column), >=768 rail + 2-column grids, >=1024 full sidebar, `.with-filters` (240px facets), `.with-aside` (300px sticky aside). List rows adapt to their **container** width (`@container`, `max-width: 479px`), not the viewport. Tables and code scroll inside labelled regions. Target widths 360-430 / 768 / 1024-1440. Verified by authors via headless Chrome screenshots per CONVENTIONS 9 (390/768/1280); not independently re-run in this review.

## 5. Performance

Counted with `wc -c` on the files as committed (uncompressed bytes; no minification, no gzip measurement):

| Group | Bytes |
|---|---|
| Foundation CSS (tokens 8,844 + base 2,725 + layout 9,205 + icons 17,135 + components 26,264) | 64,173 |
| Page CSS (solve 5,463, build 1,341, discover 1,131, identity 1,020) | 8,955 |
| Foundation JS (app 12,547, shell 5,964, i18n-common 5,028, theme-init 417) | 23,956 |
| Page JS (all lanes, incl. template-demo) | 188,439 (largest: project 36,941; profile 28,933; resource 29,430) |
| Data + i18n dictionaries | 86,696 |
| HTML pages | 5,862 (index) + 2,924-15,929 each |

Per-page load example, `question.html`: CSS 69,636 B (6 files) + JS 89,822 B (9 files) + HTML 10,434 B = ~170 KB uncompressed, no images/fonts. Not measured: gzip/brotli size, request waterfall, throttled 3G timing, LCP/INP. Caveats: ~190 KB of page JS is hand-written DOM rendering with inline mock data; it is prototype weight, not a predictor for the React build. Astryx + React + TanStack runtime cost is unknown and must be baselined at G0-UI-001/G5.

## 6. Document conflicts and inconsistencies found

| # | Conflict | Evidence | Handling in prototype |
|---|---|---|---|
| 1 | Naming of #2B440E vs #4D721D: the brief called #4D721D "Forest"; DESIGN.md names #2B440E **Deep green** (`brand.forest`) and #4D721D **Primary green** (`brand.primary`) | DESIGN.md 2 table | prototype follows DESIGN.md (`--brand-forest` #2B440E, `--brand-primary` #4D721D) |
| 2 | `defineTheme()` and `astryx docs theme|tokens`, `theme build`, `gap-report`, `tailwind-theme.css` in DESIGN.md 3.1/3.2 are not in the Astryx README read; README says themes are "CSS custom property overrides" | DESIGN.md:105-136 vs README | not used; flagged in ASTRYX_MAPPING section 2; needs a design decision to amend DESIGN.md after install |
| 3 | Package manager: DESIGN.md uses `pnpm`; README uses `npm`; ASTRYX_VERIFICATION.md says package manager `OPEN` | three sources | not decided here; G0 item |
| 4 | "No StyleX compiler required" (DESIGN.md:85) vs README install including `@stylexjs/stylex` | DESIGN.md vs README | unverified; checklist item 3 |
| 5 | SPEC defines only **403**; the prototype shows **401** for the anonymous persona (`question.js` `persona === 'anon' ? '401' : '403'`) | SPEC.md:27 only 403 | 401 is **inferred**, not specified. Needs SPEC/OpenAPI decision (401 unauthenticated vs 403 forbidden) |
| 6 | RESOLVED: Language, Access, Intelligence were labelled "Bientôt/Soon" although PRD 5.6-5.8 give them a P0 delivered across all pages | PRD 5.6-5.8 vs `shell.js` | now "Intégré/Built in" + visible P0/P1 note; `index.html#domains` table has a Statut column (`home-domains.js`) |
| 7 | Astryx catalogue size: README "150+", earlier second-hand notes ">90"/">150" | README vs ASTRYX_NOTES | official index used; no count relied on |
| 8 | ASTRYX_NOTES.md "Not verified" list is now partly obsolete (Dialog, Tabs->TabList, Toast, Table etc. confirmed) | official index | superseded in ASTRYX_MAPPING section 1 |
| 9 | DESIGN.md says dark token set is a *proposed* baseline; prototype ships it as if final | DESIGN.md 2.1 | labelled proposed here; re-measure after Astryx mapping |
| 10 | PRD 11 lists Inscription/Connexion/Parametres under Identity; prototype has sign-up/sign-in (`auth.html`, simulated) but no settings screen | PRD 11 | partial coverage gap (section 2) |
| 11 | Connect label: earlier "Aperçu (P1)" had no doc basis; PRD 8 says P0 = "Périmètre documenté et parcours conçu", rest "Suite prévue (P1/P2)" | PRD 8 | relabelled "Parcours conçu (P1/P2)" in nav, page, badge, domain table |
| 12 | ASTRYX_NOTES.md was described as verified; it is second-hand | README/mapping | marked superseded; ASTRYX_MAPPING.md is the reference (names verified vs official index only) |

## 7. Gaps and trade-offs

- Static HTML with global-state JS and inline mock data: fast to review, but nothing is reusable in React; Astryx not exercised (ASTRYX_MAPPING lists unverified props).
- Own CSS components vs "no re-creating Astryx components" (DESIGN.md 3): acceptable only because this is a throwaway prototype; none of it should ship.
- Dense rows over cards: better scan, but less room for evidence/verification detail on mobile (`row-lead` hidden below 480px container).
- "View as" can be misread as client-side security; the page callouts say simulated, but the pattern must not be ported.
- Hero band lime-on-forest and orange accents are single-use; orange text on light surfaces is not used (3.31:1 white).
- No independent a11y/AT test, no throttled-network test, no cross-browser test (Chrome headless only per CONVENTIONS 9).
- Data realism: all fictional; "verified" badges are labels on mock data, not evidence.

## 8. Recommended next steps

1. Human G0 decision on package manager, Astryx pin, DESIGN.md amendments (defineTheme/CLI) and the 401 vs 403 contract.
2. Execute ASTRYX_MAPPING section 5 checklist (install in `apps/web`, read component pages, write gap records).
3. Run axe + manual keyboard + one screen-reader pass on the 10 screens; fix; record NOT RUN items as run.
4. Throttled-network and gzip measurement on Explore, Solve, Build (US-ACCESS-001).
5. Decide the wording for nav items Language/Access/Intelligence (conflict 6) and add the Identity settings screen (auth already exists in `auth.html`) or mark it out of the prototype.
6. Re-measure contrast on Astryx-rendered components (hover/active/disabled/focus) in light, dark and system.
7. Ask product/legal review before any "verified"/"accepted" labels use real data.

## 9. Tokens added beyond DESIGN.md

DESIGN.md 2.1 defines 15 semantic roles (background, surfaces, text, border, action, accent, focus, success, warning, error). `assets/css/tokens.css` adds the tokens below, needed for interactive states and components. Ratios are the WCAG values recorded in the `tokens.css` audit comment (recomputed for 10 of them, see section 3); "n/a" = decorative or exempt. Light / dark values given where they differ. These are additions for review; none changes an approved brand colour role.

| Token | Light / dark | Purpose | Contrast (light / dark) |
|---|---|---|---|
| `--surface-hover` | #F0F4EA / #243021 | hover background of rows and ghost buttons | link on it 5.04 / n/a |
| `--action-hover`, `--link-hover` | #3F5F16 / #D0F08E | hover of action and link | on-action on hover 7.34 / 13.03 |
| `--action-active` | #34500F / #A8D45E | pressed state | 9.13 / 9.65 |
| `--link` | = action | link colour (alias, kept separate for later divergence) | 5.61 on surface / 11.39 |
| `--text-disabled` | #7B8A74 both | disabled controls (WCAG exempt); never used for information | exempt |
| `--border-strong` | #6F7C6B / #7E8E75 | input and control borders (non-text 3:1) | 4.40 on surface / 4.58 |
| `--success-bg`, `--warning-bg`, `--danger-bg` | #E3F1E7, #FBEFD9, #FCE8E5 / #1B3324, #3A2D12, #3E1F1C | status backgrounds for badges and callouts (always with icon/text) | 5.73, 5.50, 5.58 / 8.46, 8.27, 8.48 |
| `--info`, `--info-bg` | #1F4FA8 on #E4ECFA / #9CC4FF on #1D2F4D | neutral informational callouts (no info role in DESIGN.md) | 7.66 / 6.45 / 8.95 / 7.51 |
| `--overlay` | rgba(29,36,32,.55) / rgba(0,0,0,.65) | dialog backdrop | n/a |
| `--hero-bg`, `--hero-text`, `--hero-accent` | #2B440E, #FFFFFF, #BDE874 / #1B241A, #F4F7EF, #BDE874 | Explore hero band (lime only on deep green/dark) | lime on forest 7.73, white 10.85 / lime on dark surface 11.39 |
| `--skeleton` | #E3EADC / #2C3926 | loading placeholder | n/a (decorative) |
| `--code-bg`, `--code-text`, `--code-border` | #1D2420, #E8EEE0, #1D2420 / #0C110B, #E8EEE0, #43503C | code blocks, dark in both themes | text 13.38 / 16.12 |
| `--code-kw`, `--code-str`, `--code-fn`, `--code-com` | #BDE874, #E6AA66, #9CC4FF, #A3B09B (both) | syntax highlighting (never colour-only: structure and language label remain) | 11.29, 7.78, 8.87, 6.97 / 13.59, 9.36, 10.68, 8.40 |
| `--inline-code-bg` | #EDF2E7 / #2C3926 | inline `<code>` | text on it 13.92 / 11.28 |
| `--shadow-raised`, `--shadow-overlay` | rgba graphite / black | elevation of raised surfaces and dialogs | n/a |

Also added: spacing, radius, motion, z-index, layout and font-size scales (DESIGN.md 4 only fixes system fonts), `--target` 44px and `--target-min` 24px. Limits: ratios come from the audit comment, not rendered-pixel measurement; hover/active/disabled rendering and dark values remain a proposed baseline (DESIGN.md 2.1). To be re-validated after the Astryx token mapping.


## 10. Final pass: naming, UX simplification, French copy

### 10.1 Navigation naming (FR UI labels vs PRD domains)

| PRD domain | FR menu | FR page title | EN menu |
|---|---|---|---|
| Explore | Explorer | Explorer | Explore |
| Solve | Résoudre | Résoudre · Questions techniques | Solve |
| Discover | Découvrir | Découvrir · API & outils | Discover |
| Build | Construire | Construire · Projets open source | Build |
| Identity & Community | Communauté | Communauté · Membres | Community |
| Connect | Mentorat | Mentorat & collaboration | Mentoring |

Menu labels are one word, page titles add the noun. English keeps the PRD names (Connect is shown as "Mentoring"). Code identifiers, file names, `data-active` values and `data-testid`s keep the PRD names, so specs and tests are unaffected. The English domain words never appear in French UI strings (nav, breadcrumbs, titles, H1, home table, toasts, aria labels).

### 10.2 UX simplification decisions

- **Annotations toggle.** Everything that explains the prototype (Simulé badges, notes, role explanations, P0/P1 labels) is `[data-proto]` and hidden by default; shown via the Prototype button or `?proto=1`.
- **Icon-first.** Secondary actions (theme, filters, reset, share, report, sort, edit) are icon buttons with `aria-label` and a tooltip; primary actions keep a text label. Tooltips auto-flip near viewport edges (WCAG 1.4.13: dismissible with Esc).
- **Borders policy.** No borders on cards, chips or panels (fill and elevation instead); borders only as list dividers, inputs and focus. Keeps the Graphite and Lime surfaces calm in both themes.
- **Auto language and account menu.** Language follows the browser (`fr*` -> French, else English) unless chosen in the account menu; sign-in, profile, language and theme live in one icon menu, no FR/EN switch in the header.
- **Thumbnails.** Project and catalogue rows carry a generated thumbnail (initials, deterministic colour) for scanning; kept on phones, only vote and stat columns collapse below 480px.
- **Reviews pulled forward from P1.** Star rating and written reviews (with "utile" votes) appear on API/tool sheets in the prototype because they carry the integration feedback that Discover needs; reputation and badges stay P1.
- **OSS projects surfaced in Discover, canonical in Build.** Discover lists open source projects as a catalogue type for visibility; each card links to the Build project page, nothing is duplicated.
- **Build rows keep the key facts.** Compact "Recherche : Go, UX FR/EN" (3 skills max, then "+n") and a link to the originating question, which serves the contributor need and the Problem-to-Project path.

### 10.3 French copy rules

- Short, direct labels in natural developer French; keep real jargon (webhook, API, SDK, pull request, open source, backend, sandbox, stack).
- No calques or stiff nominalisations ("Voir en tant que" -> "Rôle simulé", "Parcours conçu" -> "Maquette", "Droits d'accès" demo labels -> "Rôles (démo)"), no marketing filler, no exclamation marks, no stacked hedging.
- No inclusive median dots; rephrase or use the generic masculine ("Mainteneur", "Candidat").
- Typography: no-break space before `: ; ? !` and inside `« »`; French quotes; sentence case only; accents on capitals.
- Always "vous". Fixed vocabulary: question, réponse acceptée, besoin, candidature (verb "postuler"), avis, mainteneur, fiche.
- Dates in French format (`10 novembre 2026`), decimals with a comma.
- User content (questions, bios, reviews) keeps its original language and is never rewritten by the UI.
