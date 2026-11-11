# Conventions — contract for lane agents

Foundation files are **frozen** for lanes: `assets/css/{tokens,base,layout,icons,components}.css`, `assets/js/{app,shell,theme-init,i18n-common}.js`, `assets/data/{people,common}.js`, `index.html`, `pages/_template.html`. If you need a change, write it down in your report instead of editing (parallel edits of shared files are forbidden). Authority on conflict: `/DESIGN.md` > this file.

## 1. File layout and ownership

| Lane (owner) | Page(s) | JS (dictionary + logic) | Mock data | Optional CSS |
|---|---|---|---|---|
| Solve | `pages/solve.html`, `pages/question.html` | `assets/js/pages/solve.js`, `question.js` | `assets/data/solve.js` | `assets/css/pages/solve.css` |
| Discover | `pages/discover.html`, `pages/resource.html` | `assets/js/pages/discover.js`, `resource.js` | `assets/data/discover.js` | `assets/css/pages/discover.css` |
| Build | `pages/build.html`, `pages/project.html` | `assets/js/pages/build.js`, `project.js` | `assets/data/build.js` | `assets/css/pages/build.css` |
| Identity / Connect | `pages/profile.html`, `pages/connect.html` | `assets/js/pages/profile.js`, `connect.js` | `assets/data/connect.js` | `assets/css/pages/connect.css` |

Only touch your own files. Shared mock data is read-only (`window.MOCK.{people,questions,resources,projects,countries,categories,kinds}`, helpers `MOCK.byId`, `MOCK.person(id)`, `MOCK.countryName(code)`, `MOCK.label({fr,en})`, `MOCK.fmtDate(iso)`). Your extra data goes in `window.MOCK.<lane> = {...}` from `assets/data/<lane>.js`. All content must stay clearly fictional.

### Starting a page
1. `cp pages/_template.html pages/<name>.html` (and `assets/js/pages/template-demo.js` -> `assets/js/pages/<name>.js`).
2. Set `<title data-i18n=...>`, `data-active` (`solve|discover|build|identity|connect|explore`), `data-root="../"` (already correct for `pages/`).
3. Write content inside `<main id="main" tabindex="-1">` only. **Do not hand-write header, nav, footer, skip target or toasts**: `shell.js` injects the shell (single source of truth for the 8-domain navigation). Keep the `skip-link` as first child of `<body>` and the script order of the template (`shell.js` sync, then `defer`: app, i18n-common, people, common, your data, your page js).
4. Exactly one `<h1>`; headings do not skip levels. Use `<main>`, `<section aria-labelledby>`, `<nav aria-label>`, `<article>`, `<form>` etc.

### Route contract (links between lanes; all relative, `?id=` selects the mock record)
`index.html` · `pages/solve.html` (list; `#ask` = ask-question form) · `pages/question.html?id=q-...` · `pages/discover.html` · `pages/resource.html?id=r-...` · `pages/build.html` · `pages/project.html?id=pr-...` · `pages/profile.html?id=p-...` (no id = own profile) · `pages/connect.html`. Cross-links (question <-> API <-> project) are mandatory on detail pages: use `.linked-items` (PRD journey D).

## 2. Design rules (from DESIGN.md)

- Tokens only: no hex/rgb in pages, no `!important`, no inline spacing hacks (existing inline `style` is allowed only for one-off flex shorthand; prefer utilities). Orange = secondary accent, never warning/error; lime only on deep-green/dark surfaces (`.hero`); one dominant `.btn-primary` per region; never colour-only state (always label/icon/shape). Never use flags: country is text (`.country`/`.country-code`).
- Density: lists use `.list > .row`, not large cards. No stat tiles, no invented counters. Accepted/verified/votes come from the mock data only.
- Every actionable control needs default/hover/focus-visible/active/disabled; add loading/empty/error/success where relevant. Targets >= 44px (`--target`), never < 24px.
- Responsive 360-430 / 768 / 1024-1440, no horizontal page scroll (code blocks and tables scroll inside their own labelled region).
- Gradients, glass, decorative illustration: do not use. Icons: only `.icon .i-<name>` (see `icons.css`), `aria-hidden="true"`.

## 3. Component catalogue (classes in `components.css` / `layout.css`)

| Need | Markup / classes | Notes |
|---|---|---|
| Button | `<button class="btn btn-primary|btn-accent|btn-ghost|btn-danger [btn-sm|btn-icon]">`; busy: `aria-busy="true"`; disabled: `disabled`; toggle: `aria-pressed` | Icon-only needs `aria-label`. Orange `btn-accent` uses graphite text. |
| Links | `<a>` (styled), `.link-external` (adds arrow), `.link-plain` | External links: `rel="noopener noreferrer"`; example.org only. |
| Field | `.field > label.label + .hint + input.input|select.select|textarea.textarea + .error-msg`; errors: `.field.has-error`, `aria-invalid="true"`, `aria-describedby` pointing to hint/error ids; required: `.req` + `required` | Summary on submit: `.form-summary` (`role="alert"`, heading + list of links to fields). Checkbox/radio: `label.choice > input`. |
| Search | `.search-field > .icon.i-search + input.input[type=search]` (`.search-field-lg`) | Wrap in `<form role="search">`. |
| Chips / tags | `.chip` (button, `aria-pressed` = filter toggle), `.tag-list > li > .tag` (tech tag, mono) | Tags may be `<a class="tag">`. |
| Filter bar | `.filter-bar > .filter-group`, `.filter-chips`; sidebar facets: `.with-filters > .filter-panel (fieldset/legend)` + results; `.results-meta` for count/sort | Drive with `App.Filter.run`. |
| Tabs | `<div data-tabs><div class="tablist" role="tablist"><button class="tab" role="tab" id aria-controls aria-selected>` + `.tabpanel[role=tabpanel][aria-labelledby]` | Arrow/Home/End handled by `app.js`; first selected tab wins. |
| List row | `ul.list.list-reset > li.row > .row-lead + .row-main (h3.row-title>a, .row-excerpt, .tag-list, .meta) + .row-aside` | Title link covers the row; put other links in `.row-above`. Narrow containers collapse automatically. |
| Card | `.card` (+ `.card-hover`) | Use sparingly; no stacks of big cards. |
| Metadata | `.meta > .meta-item`, `.dl-meta` (dl in two columns), `<time datetime>` | No stat tiles. |
| Linked items | `.linked-items > a.linked-item (.kind + strong)` | Question/API/project cross-links. |
| Badges | `.badge` + `-accepted` `-verified` `-later` (P1/P2 present-but-later, dashed) `-preview` `-warn` `-danger` `-sim` `-accent` | Include icon or text; never colour only. |
| Country | `<span class="country"><span class="country-code">BF</span><span class="visually-hidden">Burkina Faso</span></span>` | Text code, no flag. |
| Avatar | `.avatar` (`-sm`, `-lg`) with initials, `aria-hidden` when name is adjacent; `.person` wrapper | No remote images. |
| Vote | `.vote > button[aria-pressed] + .vote-count + button` | Buttons need `aria-label`; for display only use `.vote` with `role="img"` + `aria-label` (see index). Simulate with `data-sim` or local state. |
| Code | `.code-block > .code-head (lang label + button[data-copy="#id"]) + pre[tabindex=0][aria-label] > code#id`; syntax spans `.tok-kw .tok-str .tok-fn .tok-com`; inline `<code>` | Copy uses clipboard + toast. Always show language label. |
| Prose | `.prose` around rendered Markdown (h2/h3, lists, blockquote, code) | Max 74ch. Mark user text with `lang` + `data-user-content`. |
| Breadcrumbs | `nav.breadcrumbs[aria-label] > ol > li` (last `aria-current="page"`) | See template. |
| Pagination | `nav.pagination[aria-label] > a.btn / .btn[aria-current="page"]` | Static is fine. |
| Dialog | `<dialog class="dialog" id aria-labelledby>` with `.dialog-head/.dialog-body/.dialog-foot`; open: `[data-dialog-open="id"]`; close: `[data-dialog-close]` | Native modal: Esc, focus containment, backdrop click, focus returns to opener. Use `App.Dialog.open(id)` from JS. |
| Toast | `App.Toast.show({kind:'success|error|info', title, msg, ttl})` | Live region injected. Strings via `I18N.t`. |
| States | `.state` (+ `-error`, `-success`) with icon, `h3`, `p`, optional button; `.skeleton` / `.skeleton-row`; `.spinner`; inline `.callout` (+ `-success|-warn|-danger`) | Use `App.Sim.load` for list areas. Author lane-specific empty/error copy (`emptyKey: '<domain>.xxx.empty'` with `.title`/`.body`). |
| Table | `.table-wrap[tabindex=0][role=region][aria-label] > table.table (caption, thead th[scope=col], tbody th[scope=row])` | Scrolls inside its own region at 360px. |
| Language / theme | provided by shell (`.segmented`, `.theme-toggle`) | Do not duplicate. |
| Layout helpers | `.page-header(-row)`, `.lede`, `.section(-head)`, `.stack(-sm|-lg)`, `.cluster`, `.grid-auto`, `.grid-2`, `.grid-2-lg`, `.grid-3`, `.with-aside`, `.with-filters`, `.scroll-x`, `.hero` (restrained forest band; max one per page), `.visually-hidden`, `.text-muted`, `.text-sm`, `.mono` | |

## 4. i18n (French default; FR/EN must both exist)

- Key naming: `<domain>.<area>.<key>`, domain = your lane (`solve`, `discover`, `build`, `profile`, `connect`), e.g. `solve.list.empty.title`. Shared keys live in `i18n-common.js` (`nav.*`, `common.*`, `state.*`, `sim.*`, `header.*`, `footer.*`): reuse them, never redefine.
- Register in your page JS: `I18N.register({ fr: {...}, en: {...} });`. Use `data-i18n="key"` (text), `data-i18n-attr="aria-label:key;placeholder:key2"` (attributes), `I18N.t(key, {n: 3})` (JS, `{n}` placeholders). Set `document.title = I18N.t(...)`. JS-rendered markup must re-render on the `i18n:change` document event (`App.Sim.load` areas redraw automatically).
- UI strings only. **User content never changes with the language switch**: it keeps its original language, carries `lang="xx" data-user-content`, shows the original-language badge (`.badge` "FR"/"EN"), and is never auto-translated. A switch must not reset forms (do not rebuild `<form>`s on `i18n:change`).
- French typography: use a no-break space before `: ? ! ; »` and after `«`. Check text expansion (FR is ~20% longer) at 360px.

## 5. data-testid

Put `data-testid` on **every element that has text, an icon, an image or an interaction**, including shell elements you add. Names: kebab-case, meaningful, unique per page, `<area>-<thing>[-<qualifier>]` (`question-title`, `answer-accept-button`, `resource-country-filter`). Repeats get `-1`, `-2` suffixes from 1 (`question-row-1`, `question-tag-1-2` for nested repeats). Form controls: `<form>-<field>-input|error|hint`. Never reuse a testid on two elements.

## 6. Simulated behaviour

Anything not backed by real logic must be visibly simulated: async areas via `App.Sim.load(container, {load, render, delay, skeleton, emptyKey})` (adds the `Simulé/Simulated` badge once per page; `?state=loading|empty|error` forces a state; returns `{run, refresh}`); fake actions via `data-sim` (toast "Action simulée") or an explicit `.badge-sim` plus `.callout` explaining it. Client-side demos that do work (filter, tabs, vote toggle) must not claim persistence. P1/P2 features: show as `.badge-later` ("Bientôt/Soon") with honest wording, never as working. No invented metrics, endorsements or AI answers.

## 7. Page-specific CSS

Default: no new CSS. Reuse components and utilities first. Only for a **genuinely new pattern**, create `assets/css/pages/<lane>.css`, link it after `components.css`, prefix classes with the lane (`.solve-answer`), use tokens only, and add a mobile-first rule set. Never override foundation classes.

## 8. Keyboard and accessibility checklist

Skip link works; logical tab order; visible focus (never remove outlines); labels for every control; errors announced (`role="alert"`/`aria-live`) and linked by `aria-describedby`; status never by colour alone; AA contrast (use semantic token pairs only; see audit in `tokens.css`); dialogs return focus; `prefers-reduced-motion` respected (tokens collapse durations); content readable at 200% zoom.

## 9. Screenshot review command

```bash
S=/tmp/claude-1000/-home-aiobi6-Personal-Temp-226os/e447cffd-f11d-4447-8c7d-816101a3a029/scratchpad/shots; mkdir -p $S
for w in 390 768 1280; do for th in light dark; do
google-chrome --headless=new --disable-gpu --no-sandbox --hide-scrollbars --virtual-time-budget=4000 \
  --screenshot=$S/<page>-$w-$th.png --window-size=$w,2400 \
  "file:///home/aiobi6/Personal_Temp/226os/prototypes/html/pages/<page>.html?theme=$th&lang=fr"
done; done
# English once; forced states:  ?lang=en   ?state=error   ?state=empty   ?state=loading
```
Read the PNGs, fix overflow/overlap/contrast issues, and re-shoot. A virtual-time budget is needed so the simulated delay finishes. (Tall `--window-size` heights capture below-the-fold content.)
