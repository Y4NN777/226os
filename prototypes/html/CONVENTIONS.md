# Conventions — contract for lane agents

Foundation files are **frozen** for lanes (UX pass: see section 11) (integration pass done; see section 10 for the shared APIs they now expose): `assets/css/{tokens,base,layout,icons,components}.css`, `assets/js/{app,shell,theme-init,i18n-common}.js`, `assets/data/{people,common}.js`, `index.html`, `pages/_template.html`. If you need a change, write it down in your report instead of editing (parallel edits of shared files are forbidden). Authority on conflict: `/DESIGN.md` > this file.

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
| Language / theme | provided by the shell account menu (see section 11) | Never add a language/theme control to a page. |
| Layout helpers | `.page-header(-row)`, `.lede`, `.section(-head)`, `.stack(-sm|-lg)`, `.cluster`, `.grid-auto`, `.grid-2`, `.grid-2-lg`, `.grid-3`, `.with-aside`, `.with-filters`, `.scroll-x`, `.hero` (restrained forest band; max one per page), `.visually-hidden`, `.text-muted`, `.text-sm`, `.mono` | |

## 4. i18n (French default; FR/EN must both exist)

- Key naming: `<domain>.<area>.<key>`, domain = your lane (`solve`, `discover`, `build`, `profile`, `connect`), e.g. `solve.list.empty.title`. Shared keys live in `i18n-common.js` (`nav.*`, `common.*`, `state.*`, `sim.*`, `header.*`, `footer.*`): reuse them, never redefine.
- Register in your page JS: `I18N.register({ fr: {...}, en: {...} });`. Use `data-i18n="key"` (text), `data-i18n-attr="aria-label:key;placeholder:key2"` (attributes), `I18N.t(key, {n: 3})` (JS, `{n}` placeholders). Set `document.title = I18N.t(...)`. JS-rendered markup must re-render on the `i18n:change` document event (`App.Sim.load` areas redraw automatically).
- UI strings only. **User content never changes with the language switch**: it keeps its original language, carries `lang="xx" data-user-content`, shows the original-language badge (`.badge` "FR"/"EN"), and is never auto-translated. A switch must not reset forms (do not rebuild `<form>`s on `i18n:change`).
- French typography: use a no-break space before `: ? ! ; »` and after `«`. Check text expansion (FR is ~20% longer) at 360px.

## 5. data-testid

Put `data-testid` on **every element that has text, an icon, an image or an interaction**, including shell elements you add. Names: kebab-case, meaningful, unique per page, `<area>-<thing>[-<qualifier>]` (`question-title`, `answer-accept-button`, `resource-country-filter`). Repeats get `-1`, `-2` suffixes from 1 (`question-row-1`, `question-tag-1-2` for nested repeats). Form controls: `<form>-<field>-input|error|hint`. Never reuse a testid on two elements.

## 6. Simulated behaviour

Anything not backed by real logic must be visibly simulated: async areas via `App.Sim.load(container, {load, render, delay, skeleton, emptyKey})` (adds the `Simulé/Simulated` marker once per page, hidden unless annotations are on: it is `[data-proto]`; `?state=loading|empty|error` forces a state; returns `{run, refresh}`); fake actions via `data-sim` (toast "Action simulée") or an explicit `.badge-sim` plus `.callout` explaining it. Client-side demos that do work (filter, tabs, vote toggle) must not claim persistence. P1/P2 features: show as `.badge-later` ("Bientôt/Soon") with honest wording, never as working. No invented metrics, endorsements or AI answers.

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

## 10. Shared runtime APIs added at integration (use these, do not re-implement)

- `App.Tabs.init(root)` wires `[data-tabs]` inside `root` (idempotent; use after rendering tabs dynamically). Arrow/Home/End move focus **and call `.click()`** on the target tab, so a `click` listener (or the bubbling `tabs:change` event, `detail: {index, tab, panel}`) is the single activation path.
- `App.Sim.load(box, {load, render, delay, skeleton, emptyKey, empty, error})`: `empty`/`error` take `{icon, key | titleKey+bodyKey, heading, action: {labelKey|label, href|onClick, icon, primary, testid}, retry:false}`. Actions render in `.state-actions`. `?state=` accepts `loading|empty|error|notfound`.
- Detail pages (tabs, header must not be rebuilt): `App.Sim.forced()` returns the forced state or null; `App.Sim.show(box, kind, cfg)` renders loading/empty/error/notfound into any box (language-aware, adds the Simulated marker, Retry reloads without `?state`).
- `App.avatar(personOrName, {size:'sm|lg', testid})` and `App.initials(name)`: initials avatar (`.avatar`).
- `window.Markdown.render(md, {id, headStart, copyText, copyLabel, codeLabel})` from `assets/js/markdown.js` (load before page scripts). `headStart: 2` maps `##` to `h3`. `Markdown.unbalancedFence(md)`.
- `App.TestIds` (in `app.js`, auto-runs on every page): assigns `<nearest-testid-ancestor>-<icon-name|tag>-N` to any text/icon/control still lacking a `data-testid`. Explicit test ids always win; still name important elements yourself.
- CSS: `.h1`-`.h4` heading-size utilities; `.filter-bar.filter-bar-grid` (2 cols on phones, auto-fit columns from 640px); `.tag-list` resets list styling itself; `.row` works with or without `.row-lead`; `.dialog-body > .callout|.form-summary` get bottom spacing; `.tablist` shows an edge shadow on the side with hidden tabs (set `--tab-bg` if it sits on a non-`--bg` surface; cards, dialogs and lists are handled); `.pagination .btn` never wraps its label.
- Shell: the header create action ("Ask a question", `pages/solve.html#ask`) is now visible at every width (icon-only below 1024px, label kept for assistive tech).

## 11. UX pass (supersedes anything above that conflicts). A real user does not need prototype scaffolding.

**Principle.** Default view = product only. Anything that exists to explain the prototype (Simulé/Simulated badges and notes, "view as" panels, P0/P1/Bientôt explanations, demo callouts, "this is fictional" text) must carry the bare attribute `data-proto`; it is `display:none` unless `<html class="proto-annotations">` (toggle in the floating Prototype panel, `?proto=1` for QA, persisted in `226os.proto.annotations`). Do not add a visible "Simulé" label to real-looking UI. Simulated actions still give feedback via `data-sim` toast. Capabilities (language, access, intelligence) are not menu items; they live in DESIGN_REVIEW.

**Shell (shell.js / app.js, do not duplicate).**
- Header: search, theme icon toggle, account menu (icon, `aria-haspopup=menu`: sign in, my profile, language Français/English radios, theme Clair/Sombre/Système radios), primary "Ask" (+). No FR/EN control, no sign-in text button. Test ids: `account-menu-button`, `account-menu-panel`, `account-menu-lang-fr|en`, `account-menu-theme-light|dark|system`, `theme-toggle`, `header-create`.
- Language: `?lang=` (QA override, not persisted) > persisted choice (`226os.lang`, set only from the menu) > `navigator.languages[0]` (`fr*` -> fr, else en). A QA browser is therefore English by default: pass `?lang=fr` in screenshot commands.
- Nav: Explore, Solve, Discover, Build, Identité & Communauté, Connect (neutral "Bientôt" label, dot on the rail). `home-domains.js` and the domain table were removed from `index.html`.
- Footer: just "226os"; the prototype note is `[data-proto]`.
- Floating **Prototype** button (bottom-left, `proto-toggle`) opens `proto-panel`: role switcher, screen state (Normal/Chargement/Vide/Erreur, reloads with `?state=`), annotations switch.

**`App.Proto`** (lanes: migrate your "view as" panel into it, then delete the panel and its `?as=` plumbing or keep `?as=` as an alias read by your own code).
- `App.Proto.register({roles:[{id:'author', label:{fr:'Auteur', en:'Author'} | labelKey:'q.role.author'}], default:'visitor', legend?:{fr,en}})` once at script load (before DOMContentLoaded).
- `App.Proto.role`, `.is(id)`, `.setRole(id)`; document event `proto:role` (`detail:{role, previous, roles}`; not fired for the initial value, read `Proto.role` at start). `?role=` overrides; last choice remembered per page. Re-render only the role-dependent parts, never rebuild forms.
- `App.Proto.state` ('default|loading|empty|error|notfound'), `.setState(s)`; `App.Proto.annotations`, `.setAnnotations(bool)`.

**Borders removed** (use fill/elevation): `.btn` (secondary = `--surface-subtle` fill; ghost transparent; danger = `--danger-bg` fill), `.chip` (fill; pressed = `--selected-bg` + check), `.badge`, `.country-code`, `.tag`, `.avatar`, `.vote button`, `.callout`, `.state` (fill, no dashes), `.toast`, `.dialog`, `.hero`, `.segmented`, `.card`, `.list`, `.filter-bar`, `.filter-panel`, `.linked-item`, `.review` (`--shadow-card`). Kept: inputs/select/textarea, `.table-wrap`, `.code-block`, `kbd`, row separators, tabs underline, focus rings. `prefers-contrast: more` / forced colors restore outlines. New tokens: `--surface-subtle-hover`, `--selected-bg`, `--rating`, `--shadow-card`, `--tone-0..4-bg/fg` (AA pairs documented in tokens.css). Pages must not re-add borders to these; use `.card` + spacing instead of bordered boxes.

**Icon buttons.** `.btn.btn-icon` is now borderless (44px; `.btn-sm` or `.btn-icon-sm` = 36px) with a CSS tooltip taken from `aria-label` (focus + hover, Esc dismisses; `.tip-top|.tip-end|.tip-start` position it; `.has-tip` enables it on non-icon buttons; `[data-tip="text"]` on any element). `aria-label` is mandatory and i18n (`data-i18n-attr="aria-label:key"`). Tooltips are off inside `.code-block`, `.dialog`, `.menu-panel`. Pressed = `aria-pressed="true"` (selected fill). Two-glyph buttons: `<span class="icon i-eye icon-main">` + `<span class="icon i-eye-off icon-alt">` (alt shown when `aria-pressed="true"` or `.is-done`). New icons: `eye eye-off more-vertical sort log-in log-out settings sliders star-fill star-half chevron-up image palette` (plus existing `copy check x filter dots`).
- Password reveal: `<button type="button" class="btn btn-icon" data-pw-toggle="#pw" aria-pressed="false">` + eye/eye-off glyphs; app.js toggles the input type and sets the aria-label (`common.showPassword|hidePassword`). No "Afficher" text.
- Copy: `<button class="btn btn-icon btn-sm" data-copy="#code-id" aria-label="Copier"><span class="icon i-copy icon-main"></span><span class="icon i-check icon-alt"></span></button>`; success adds `.is-done` (check) for 1.6 s + toast. `markdown.js` still renders a text "Copier" button: switch it to this markup (shared file, owned by the lane that integrates it).
- Row/card actions: `.btn-icon.btn-sm` with `i-more-vertical` + `App.Menu`; filter/sort/close likewise (`common.filter`, `common.sort`, `common.more`, `common.close`).

**`App.Menu.init(root)`** wires `.menu[data-menu] > button[data-menu-button][aria-haspopup=menu][aria-expanded][aria-controls] + .menu-panel[role=menu][data-menu-panel][hidden]`; items `a|button.menu-item[role=menuitem|menuitemradio|menuitemcheckbox]` (`aria-checked` + `<span class="icon i-check menu-check">` for radios), `.menu-sep[role=separator]`, `.menu-label`; `.menu-panel.menu-start` aligns left. Keys: Enter/Space/ArrowDown open on first, ArrowUp last, arrows/Home/End/type-ahead, Esc closes and refocuses the button, Tab closes, outside click closes. Selecting an item closes the menu. Give every item a testid.

**`App.thumb(name, {size:'sm|md|lg|wide', shape:'circle', icon:'code', img:'local/path.png', alt, tone:0-4, testid})`** returns `span.thumb.tone-N` (+ `.thumb-sm|-lg|-wide|-circle`). Tone and monogram derive deterministically from `name`; with `icon` shows a glyph; with `img` an image slot (local files only, `alt` required unless decorative). Sizes 32/48/72 px, `wide` = 16:9 cover for cards. Decorative: aria-hidden unless it has an image alt; always put the real name next to it. `App.avatar` now also uses `.tone-N`. Use thumbs for resource/API, project and card/row leads (`.row-lead`).

**`App.rating(value, {count, max:5, size:'lg', testid})`** read-only: `span.rating[role=img]` with a full text alternative ("4,5 sur 5, 12 avis"), stars + numeric value (+ count). Half stars. **`App.ratingInput({name, legend | legendKey, value, required, testid})`** returns `fieldset.rating-input` of native radios (`name`, values 1..5, arrow keys, `required` on the group); read the value with `form.elements[name].value`. Style hooks only: `.rating`, `.rating-lg`, `.rating-input`.

**`App.review({name | person, country:'BF', rating, date:'2026-03-01', text, lang:'fr', verified:true}, {testid:'review-1'})`** returns `article.review` (avatar, name, `.country` text code, date, rating, user text with `lang` + `data-user-content`, original-language badge, optional "Intégration vérifiée" badge). Wrap in `ul.review-list > li`. testids: `<testid>-name|-country|-date|-rating|-text|-lang|-verified`. Fictional data only; never claim real reviews. Re-render on `i18n:change` (labels and date format are language-aware).

**Shared keys added** (do not redefine): `header.account|profile|themeLabel|theme.light|dark|system`, `proto.*`, `common.showPassword|hidePassword|more|filter|rating.of|reviews|integrationVerified`. Removed: `nav.preview|builtin|*.why`.

**What each lane must adopt (phase 2).**
- All pages: wrap every Simulé badge, `.sim-marker`, "fictional/simulated" callout and demo note in `data-proto` (or delete); remove "view as" panels in favour of `App.Proto.register`; replace text buttons that mean icon actions (copy, show/hide password, close, more, filter, sort) with `.btn-icon`; drop borders added in page CSS on chips/cards/panels; add thumbs to list rows and headers; use `ul.list`/`.card` without bordered wrappers.
- `pages/solve.html`, `question.html` (Solve): migrate the view-as panel (question has `?as=author|other|moderator|visitor|admin`) and the Simulé badge on the ask form (`solve.html` ask section) to Proto/`data-proto`; icon copy in code blocks; reviews are not needed here.
- `pages/discover.html`, `resource.html` (Discover): resource thumbs in rows and header (`App.thumb`), rating + `App.review` list for the "retours" tab (replace the "Bientôt" feedback badge/notes with real reviews incl. `verified`), `nav.later` badges for compare only if truly non-functional; copy buttons to icons.
- `pages/build.html`, `project.html` (Build): thumbs for projects, `?as=owner` panel to Proto, remove sim notes, icon actions.
- `pages/profile.html` (own: `?as=self&edit=1`), `connect.html`, `auth.html` (Identity/Connect): auth password reveal -> `data-pw-toggle`; profile view-as/edit demo panel and its `.sim-marker` (line ~79) to Proto/`data-proto`; connect "preview" explanations to `data-proto`; avatars via `App.avatar`.
- `index.html`/`home.js`: done by this pass (domain table removed; `renderDomains` in home.js is now a no-op and its `home.dom.*` keys are dead: remove at the next integration).
