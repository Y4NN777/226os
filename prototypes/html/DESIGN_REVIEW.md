# Design review: 226os HTML prototype

**Status:** self-review by the author agents, 2026-11-11. This is NOT an independent design or accessibility audit. What was verified and what was not is stated per section. Authority on conflict: `DESIGN.md` > `CONVENTIONS.md`.

## 1. Design decisions and rationale

| Decision | Choice | Rationale / source |
|---|---|---|
| Direction | Graphite & Lime, quiet neutral canvas (#F6F7F4 / white), green primary, orange sparing, lime only on deep-green/dark surfaces | DESIGN.md 2, 2.2 |
| Density | Technical, dense **list rows** (`.list > .row`), no stat tiles, no invented counters, cards sparingly | Developer audience, scannability on slow/mobile networks; CONVENTIONS 2 |
| Navigation model | One shell, injected by `shell.js` (single source of truth for 9 domain entries). >=1024px full sidebar 232px; 768-1023px rail 92px; <768px bottom bar (4 domains + "More" opening the drawer) | PRD 11 (shared nav across domains); thumb reach on phones (PRD 7F) |
| P0 / P1 / P2 representation | P0 = normal links. Connect = `badge-preview` (page exists, content simulated). Language, Access, Intelligence = non-link item with dashed `badge-later` ("Bientot/Soon") plus a visually-hidden reason. Later features never styled as working | PRD 8 "no fake UI"; TRACEABILITY "P1/P2 not represented by fake UI" |
| FR/EN rule | French default; UI strings in dictionaries (`I18N.register`); **user content never translated**, carries `lang` + `data-user-content` and an original-language badge; switching language does not reset forms | SPEC US-IDENTITY-001, PRD 5.6/7E; `lang` visible per post |
| Evidence / verification display | Accepted answer, verified resource, sources and verification date shown as badge **with text/icon** plus `<time>`; source links `link-external`; no claim of official endorsement; all data fictional | DIFF-01/04 in TRACEABILITY |
| Authorization "view as" | Persona toggle (author / other member / anonymous) on question, profile, project pages to **demonstrate** server-enforced denials (403) and field visibility (profile). Clearly a prototype demonstrator | SPEC US-SOLVE-001 / US-BUILD-001 deny scenarios; PRD journeys A, C |
| Simulation honesty | `App.Sim.load` adds a "Simule/Simulated" badge; `?state=loading|empty|error` forces states; `data-sim` actions toast "simulated" | CONVENTIONS 6 |
| Typography / assets | System font stacks, inline CSS icon sprite, no remote fonts/images/flags, country as text code | DESIGN.md 4; US-ACCESS-001 (no heavy media) |
| No colour-only state | Every state has label/icon/shape | DESIGN.md 2.2 |

## 2. Coverage matrix

Pages: `index.html` (Explore), `pages/{solve,question,discover,resource,build,project,profile,connect}.html`. States column: L loading, E empty, X error (all forced via `?state=`), plus form validation and simulated success where noted. "Present" means a page and markup exist, not that it was tested.

| PRD domain | CADEV scope (PRD 8) | Screens in prototype | States / notes |
|---|---|---|---|
| Explore | accueil, recherche unifiee | `index.html` | client-side text match over mock arrays; L/E/X via `App.Sim` |
| Solve | Q&A, snippets, accepted answer, votes, search | `solve.html` (list, filters, `#ask` form), `question.html` (detail, answers, vote, accept, tabs, linked items, answer form + preview) | L/E/X, form-summary, vote toggle (not persisted), accepted mark, code blocks |
| Discover | catalogue, sourced fiches, country/category | `discover.html` (filter bar, rows), `resource.html` (detail, tabs, dialog, sources, linked items) | filters; provenance links; stale/unknown honesty: shown only if mock data models it (not verified per record) |
| Build | projets GitHub, besoins, candidatures | `build.html` (directory, filters, publish form + dialog), `project.html` (needs, roles, application dialog, duplicate-submission simulation) | form-summary, dialog, duplicate application demo |
| Identity & Community | profils, identite liee aux contenus | `profile.html` (tabs, view-as persona, privacy table, dialog, form) | owner / member / anonymous field visibility |
| Connect | perimetre documente | `connect.html` (preview; rows tagged `badge-later` + `badge-sim`) | honest "coming later" |
| Language | interface FR/EN | shell toggle on all pages | persisted in `localStorage`, forms not reset (author statement; not re-tested after the last CSS/JS edits) |
| Access | responsive + reseau | all pages, bottom bar, container-query rows | see sections 4-5 |
| Intelligence | recherche PostgreSQL + liens contextuels | header search, `?q=` on Explore, `.linked-items` | AI/semantic NOT shown (nav item `later`) |

| SPEC story | Where represented | Gap |
|---|---|---|
| US-SOLVE-001 (publish, answer, accept once, 403 for non-author, search) | `solve.html#ask`, `question.html` answer form, accept control, view-as 403 callout, search | "one accepted answer only" enforced client-side in mock only; real invariant is server/DB |
| US-DISCOVER-001 (filter by country + category, provenance, related discussions) | `discover.html`, `resource.html` | provenance depends on mock fields |
| US-BUILD-001 (interest in a need, no duplicate on second click, private applications denied) | `project.html` (dialog, duplicate simulation, view-as) | private list denial shown as simulated state |
| US-IDENTITY-001 (linked resource/project on question; FR/EN switch keeps content) | `.linked-items` on detail pages; language toggle; `lang` + `data-user-content` | none known; not machine-tested |
| US-ACCESS-001 (slow network, loading/empty/error, no heavy media) | `App.Sim` with skeletons; no media; byte counts below | no throttled-network test run |

Not represented: Identity sign-up/sign-in and settings screens listed in PRD 11 (only profile), account/session flows, publish-resource screen, offline drafts (P2).

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
| 6 | Nav shows Language, Access, Intelligence as "later", but PRD 8 / TRACEABILITY list P0 slices for them (FR/EN UI, responsive, PostgreSQL search + context links) | PRD 8 vs `shell.js` DOMAINS | cross-cutting P0 parts are delivered (toggle, responsive, search); the nav label means the dedicated domain screens, which could mislead. Needs wording decision |
| 7 | Astryx catalogue size: README "150+", earlier second-hand notes ">90"/">150" | README vs ASTRYX_NOTES | official index used; no count relied on |
| 8 | ASTRYX_NOTES.md "Not verified" list is now partly obsolete (Dialog, Tabs->TabList, Toast, Table etc. confirmed) | official index | superseded in ASTRYX_MAPPING section 1 |
| 9 | DESIGN.md says dark token set is a *proposed* baseline; prototype ships it as if final | DESIGN.md 2.1 | labelled proposed here; re-measure after Astryx mapping |
| 10 | PRD 11 lists Inscription/Connexion/Parametres under Identity; prototype has none | PRD 11 | coverage gap (section 2) |

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
3. Run axe + manual keyboard + one screen-reader pass on the 8 pages; fix; record NOT RUN items as run.
4. Throttled-network and gzip measurement on Explore, Solve, Build (US-ACCESS-001).
5. Decide the wording for nav items Language/Access/Intelligence (conflict 6) and add Identity auth/settings screens or mark them out of the prototype.
6. Re-measure contrast on Astryx-rendered components (hover/active/disabled/focus) in light, dark and system.
7. Ask product/legal review before any "verified"/"accepted" labels use real data.
