# 226os — standalone HTML prototype

High-fidelity, static prototype of the 226os UI (Graphite & Lime). It is a **design/UX proof**, not the product: no backend, no build step, no external resources, no Astryx runtime. Authority: `/DESIGN.md`, `/docs/PRD.md`, `/docs/SPEC.md`. Lane authors read `CONVENTIONS.md`.

## Preview

- Open `index.html` directly (`file://`), or serve it: `cd prototypes/html && python3 -m http.server 8080` then http://localhost:8080/.
- Query parameters (all pages): `?lang=fr|en` (default fr), `?theme=light|dark` (default: OS preference), `?state=loading|empty|error|notfound` (forces the simulated async state), `?q=` on `index.html` (search).
- Language and theme choices persist in `localStorage` (keys `226os.lang`, `226os.theme`); the page still works if storage is blocked.

## What is simulated

- All data is fictional (`assets/data/*.js`): people, providers, projects, dates and counts are invented and not real metrics or endorsements.
- Loading, empty and error states come from `App.Sim.load` after an artificial delay, always with a visible **Simulé / Simulated** badge.
- Buttons marked `data-sim` show a "simulated action" toast; nothing is saved.
- Search is a client-side text match over the mock arrays (accent-insensitive).
- Auth is simulated in `pages/auth.html` (sign-in, create account, generic errors, 429, session-expired, verification-pending, forgot-password); nothing is stored or verified. Not implemented here: real authentication, persistence, real GitHub import, translation of content, AI. Settings screens are absent.

## Structure

```
index.html                 Explore / Home
pages/_template.html       copy this to start a page (do not edit)
pages/auth.html            simulated sign-in / create account / recovery states
pages/*.html               9 lane pages (solve, question, discover, resource, build, project, profile, connect, auth)
assets/css/                tokens, base, layout, icons, components
assets/css/pages/          per-lane CSS (auth, build, discover, identity, solve)
assets/i18n/               solve.js (Solve/Question dictionary)
assets/js/                 theme-init (head), shell (injects nav), app (runtime, tabs, sim states), i18n-common, markdown.js (shared renderer)
assets/js/pages/           <name>.js per page, plus home-domains.js (Explore domain priority table)
assets/data/               people.js, common.js (shared), <lane>.js (lane-owned)
CONVENTIONS.md             contract for lane agents (data-testid rules)
ASTRYX_MAPPING.md          prototype -> Astryx component mapping; names verified against the official component index, props/behaviour NOT verified
ASTRYX_NOTES.md            SUPERSEDED second-hand notes (not verified); use ASTRYX_MAPPING.md
DESIGN_REVIEW.md           self-review: decisions, coverage, a11y, tokens added beyond DESIGN.md, document conflicts
```
