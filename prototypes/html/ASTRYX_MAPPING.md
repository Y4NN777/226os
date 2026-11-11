# Astryx mapping: HTML prototype -> `@astryxdesign/core` components

**Status:** PROPOSED mapping, written 2026-11-11. Component **names** are verified against the official index; **props, slots, variants, token names and behaviour are UNVERIFIED** until the pinned package is installed and inspected (G0-UI-001, `docs/spec-kit/ASTRYX_VERIFICATION.md`). Nothing here is an integration claim. The prototype itself runs no Astryx code.

## 1. Status and sources

| Source | What was read | Used for |
|---|---|---|
| https://github.com/meta/astryx (README) | package names, install, "A theme is a set of CSS custom property overrides", theme list, beta status | theming approach, package set |
| https://astryx.atmeta.com/components (component index) | exact component names by category; doc path of each = `https://astryx.atmeta.com/components/<Name>` | every "Astryx component" cell below |
| Per-component pages | **NOT read** for this document (not fetched). Cell "doc path" gives the URL to read at G0-UI-001, not evidence of its content | n/a |

Verified facts (from the two sources above; fetched by the main session, relayed here):
- Package `@astryxdesign/core`; related `@astryxdesign/cli`, `@astryxdesign/build`, `@astryxdesign/theme-*` (themes: neutral, butter, chocolate, matcha, stone, gothic, y2k). `@astryxdesign/vega` and `@astryxdesign/charts` are canary-only; `@astryxdesign/lab` is not on npm. MIT. "Currently in Beta"; no version listed, no GitHub releases. README claims "150+ accessible components".
- README CLI mentions only `astryx docs global` and `npm run astryx -- component --list`. README install uses **npm**.
- README does **not** mention `defineTheme`.

**Supersedes `ASTRYX_NOTES.md`.** That file (second-hand, "not verified") said Dialog, Tabs, Table, Toast, Badge, Avatar, Breadcrumbs, Pagination and form inputs were unconfirmed, and listed `Stepper` as inferred. The official index now confirms: `Dialog`, `TabList` (tabs are named **TabList**, not "Tabs"), `Toast`, `Table`, `Badge`, `Avatar`, `Breadcrumbs`, `Pagination`, `Field`, `TextInput`, `TextArea`, `Selector`, `CheckboxInput`, `RadioList`, `Stepper`, `Markdown`, `CodeBlock`, `EmptyState`, `Skeleton`, `PowerSearch`, `SideNav`, `AppShell`. Still **not** confirmed by the official sources read: `defineTheme`, `xstyle`, `swizzle`, MCP server, cascade-layer names, `reset.css/astryx.css/theme.css` import paths (these appear only in second-hand notes and `DESIGN.md`).

## 2. Theming approach

Official statement: a theme is **CSS custom property overrides**. So the Graphite & Lime mapping is: take a base theme (`@astryxdesign/theme-neutral` is the one `DESIGN.md` names; the README lists seven), and override its custom properties in a 226os stylesheet scoped for light and dark.

How the prototype's `assets/css/tokens.css` maps (names on the left are 226os semantic tokens; names on the right are **unknown** until the installed theme's token docs are read):

| 226os token (tokens.css) | Light / dark | Astryx target |
|---|---|---|
| `--bg`, `--surface`, `--surface-raised`, `--surface-subtle` | #F6F7F4/#111710, #FFF/#1B241A, #FFF/#243021, #EDF2E7/#2C3926 | background/surface slots: UNVERIFIED names |
| `--text`, `--text-muted` | #1D2420/#F4F7EF, #576454/#C1CCB9 | text slots: UNVERIFIED |
| `--action`, `--on-action`, `--link` | #4D721D/#BDE874, #FFF/#16220E | primary/accent slots: UNVERIFIED (DESIGN.md's `--color-accent` is illustrative only) |
| `--accent`, `--on-accent` | #C67C35/#E6AA66, graphite | secondary accent; may have no slot (then a 226os-prefixed property) |
| `--success/--warning/--danger/--info` (+`-bg`) | see tokens.css | status slots: UNVERIFIED; Badge/Banner/StatusDot variants UNVERIFIED |
| `--focus` | #275DCA/#9CC4FF | focus ring slot: UNVERIFIED; must stay visibly distinct from green |
| `--border`, `--border-strong` | #DCE3D6/#43503C, #6F7C6B/#7E8E75 | border slots; keep `--border-strong` (>=3:1 non-text) for input borders |
| `--code-*`, `--hero-*`, `--skeleton`, `--overlay` | see tokens.css | likely 226os-only properties (CodeBlock/Skeleton token names UNVERIFIED) |
| `--radius-*`, `--space-*`, `--fs-*`, `--dur-*`, `--z-*` | 4px grid, 6/8/12px radii | Astryx radius/spacing/type tokens preferred; DESIGN.md forbids arbitrary per-page radii |

Rules: (1) override, never fork Astryx internals; (2) one Graphite & Lime stylesheet, no per-component hex; (3) dark is applied by the same mechanism (custom properties under a mode selector); how Astryx switches modes (class, attribute, `prefers-color-scheme`) is UNVERIFIED; (4) re-run the contrast audit at the top of `tokens.css` on the *rendered* Astryx components, because Astryx internal pairings (hover/active/disabled) differ from the prototype's; (5) **flag:** `DESIGN.md` section 3.1 prescribes `defineTheme()` from `@astryxdesign/core/theme`, `astryx docs theme|tokens`, `astryx theme build`, `astryx gap-report` and a `tailwind-theme.css` bridge. None of these appears in the README read. Treat as unverified until the installed CLI lists them (the verification protocol already says not to assume them).

## 3. Component mapping

Legend: **Match** = a verified component covers the role; **Compose** = built from several verified components; **Gap** = no verified match (see section 4). "Doc" = `https://astryx.atmeta.com/components/<Name>` (to be read at G0-UI-001).

| # | Prototype element / class | Astryx component(s) (names verified) | Fit | Notes |
|---|---|---|---|---|
| 1 | App shell (`body.app`, `.app-header`, `.page`, skip link) | `AppShell`, `Layout` | Match? | Whether AppShell provides header/side/main slots and a skip link: UNVERIFIED. Keep own skip link if absent. |
| 2 | Sidebar nav (`.side-nav`, `.nav-list`, `.nav-item`, 232px full / 92px rail) | `SideNav` | Match? | Rail (collapsed) mode and active state `aria-current`: UNVERIFIED. |
| 3 | Nav item "later/preview" (`.badge-later`, `aria-disabled`) | `SideNav` item + `Badge` | Compose | Disabled-but-announced items with a "Soon" badge: UNVERIFIED. |
| 4 | Mobile bottom bar (`.bottom-nav`, 4 domains + More) | none in index (`TopNav`, `SideNav`, `BottomSheet`, `Drawer` exist) | **Gap** | Bottom tab bar not in the index. "More" panel could be `Drawer` or `BottomSheet`. See section 4. |
| 5 | Mobile nav drawer / scrim (`.nav-scrim`, `.nav-drawer-head`) | `Drawer` (alt. `BottomSheet`) | Match? | Focus trap/Esc: UNVERIFIED. |
| 6 | Header (brand, search, create, toggles) | `AppShell`, `TopNav`, `Toolbar` | Compose | Header contents are product composition. |
| 7 | Header search (`.search-field`, `form[role=search]`) | `PowerSearch` or `TextInput` (+`Icon`) | Match? | PowerSearch capability (filters/tokens vs plain query) UNVERIFIED; pick after reading its page. Server-side FTS stays 226os (PostgreSQL). |
| 8 | Global quick search / jump | `CommandPalette` | Optional | Not in prototype; candidate, not needed for P0. |
| 9 | Filter bar (`.filter-bar`, `.filter-group`) | `Toolbar`, `Selector`, `MultiSelector`, `Field`, `Stack` | Compose | Country/category/sort selectors. |
| 10 | Sidebar facets (`.filter-panel` fieldset/legend, checkboxes) | `CheckboxInput`, `RadioList`, `Field`, `Collapsible` | Compose | Fieldset/legend semantics UNVERIFIED. |
| 11 | Filter chips (`.chip[aria-pressed]`) | `ToggleButton` / `ToggleButtonGroup` (alt. `SegmentedControl`) | Compose | Toggle semantics (`aria-pressed`) UNVERIFIED. |
| 12 | Tag / tech tags (`.tag`, `.tag-list`) | `Token` (alt. `Badge`) | Match? | `Token` listed under Content; its purpose is UNVERIFIED. Mono styling is custom. |
| 13 | Dense list row (`ul.list > li.row`, container-query collapse) | `List`, `ClickableCard` (alt. `Table`) | Compose | Stretched-title-link pattern and container queries are 226os layout. |
| 14 | Card (`.card`, `.card-hover`) | `Card`, `ClickableCard` | Match | Use sparingly per DESIGN.md. |
| 15 | Metadata (`.meta`, `.dl-meta`, `<time>`) | `MetadataList`, `Timestamp`, `Text` | Match? | Timestamp locale/format and FR/EN handling UNVERIFIED. |
| 16 | Linked items (`.linked-items > a.linked-item`) | `ClickableCard` / `Link` + `Text` + `Icon` in `Grid` | Compose | Product component `LinkedItems` (PRD journey D). |
| 17 | Vote control (`.vote`, `.vote-count`) | `ToggleButton` or `IconButton` + `Text`, in `ButtonGroup`/`Stack` | **Compose / Gap** | No Vote component. Needs `aria-pressed`, count live update, label. See section 4. |
| 18 | Accepted-answer mark (`.badge-accepted`) | `Badge` + `Icon` | Compose | Status never colour-only. |
| 19 | Badges (`-verified`, `-later`, `-preview`, `-warn`, `-danger`, `-sim`, `-accent`) | `Badge`, `StatusDot` | Match? | Variant names UNVERIFIED; dashed "later" style is custom CSS. |
| 20 | Language-original badge ("FR"/"EN" on user content) | `Badge` + `VisuallyHidden` | Compose | Plus `lang` attribute (not component-specific). |
| 21 | Country text code (`.country`) | `Text`, `Badge`, `VisuallyHidden` | Compose | No flags, per DESIGN.md. |
| 22 | Avatar (`.avatar`, initials) | `Avatar`, `AvatarGroup` | Match | Initials fallback / no remote image: UNVERIFIED. |
| 23 | Buttons (`btn-primary/accent/ghost/danger`, `-sm`, `-icon`) | `Button`, `IconButton`, `ButtonGroup` | Match | Orange accent variant (graphite text) likely has no variant; custom property override or a 226os variant. Busy/`aria-busy` UNVERIFIED. |
| 24 | Links (`.link-external`, `.link-plain`) | `Link` | Match | External-arrow icon and `rel` handling UNVERIFIED. |
| 25 | Tabs (`[data-tabs]`, tablist/tab/tabpanel) | `TabList` | Match | Panel part and arrow-key behaviour: UNVERIFIED (read page). |
| 26 | Dialog (native `<dialog>`, head/body/foot) | `Dialog` | Match | Focus return, Esc, backdrop: UNVERIFIED; Astryx implementation replaces native `<dialog>` (decide at G0-UI-001). Alt. for mobile: `BottomSheet`. |
| 27 | Toast (`App.Toast`, live region) | `Toast` | Match | Provider/imperative API and live-region politeness UNVERIFIED. |
| 28 | Code block (`.code-block`, lang label, copy, `tok-*`) | `CodeBlock`, `Code` (inline), `IconButton`, `Toast` | Match? | Syntax highlighter, language label, copy button, keyboard-focusable scroll region: UNVERIFIED. Language label required by prototype rules. |
| 29 | Markdown prose (`.prose`, max 74ch) | `Markdown` (alt. `RichTextView`), `Heading`, `Blockquote` | Match? | Sanitization/raw-HTML policy and per-element styling UNVERIFIED; XSS posture is a security-review item. |
| 30 | Inline code | `Code` | Match | |
| 31 | Keyboard hints | `Kbd` | Match | Not used in prototype. |
| 32 | Breadcrumbs (`nav.breadcrumbs`) | `Breadcrumbs` | Match | `aria-current="page"` UNVERIFIED. |
| 33 | Pagination (`nav.pagination`) | `Pagination` | Match | Server pagination contract is SPEC-OPEN; cursor vs page UNVERIFIED in UI. |
| 34 | Field (`.field`, label, hint, error) | `Field` + `TextInput`/`TextArea`/`Selector`, `FormLayout` | Match | `aria-describedby`/`aria-invalid` wiring UNVERIFIED. TanStack Form integration is 226os code. |
| 35 | Error summary (`.form-summary`, `role=alert`, links to fields) | `Banner` + `List` + `Link` | **Compose / Gap** | No FormErrorSummary in index; focus-on-submit behaviour is custom. |
| 36 | Checkbox / radio (`label.choice`) | `CheckboxInput`, `RadioList` | Match | |
| 37 | Selects (`.select`) | `Selector`, `MultiSelector`, `ComplexSelector` | Match | Native `<select>` vs custom listbox: a11y UNVERIFIED. |
| 38 | Inline callout (`.callout`, `-success/-warn/-danger`) | `Banner` | Match? | |
| 39 | Empty / error / success state (`.state`) | `EmptyState`, `Banner`, `Button` | Match? | Error-with-retry variant UNVERIFIED. |
| 40 | Loading skeleton (`.skeleton`, `.skeleton-row`) | `Skeleton` | Match | |
| 41 | Spinner (`.spinner`) | `Spinner`, `ProgressBar` | Match | |
| 42 | Table (`.table-wrap`, scroll region, caption) | `Table`, `ScrollableArea` | Match? | Caption, `scope`, horizontal scroll labelled region: UNVERIFIED. |
| 43 | Language toggle (FR/EN, `.segmented`) | `SegmentedControl` | Match? | Alt. `ToggleButtonGroup`. Semantics (radiogroup vs pressed buttons) UNVERIFIED. |
| 44 | Theme toggle (light/dark/system, `.theme-toggle`) | `IconButton` or `SegmentedControl`, `Icon` | Compose | Theme state is 226os logic (Astryx provider mode API UNVERIFIED). |
| 45 | "View as" persona toggle (`.viewas`, `.solve-viewas`) | `SegmentedControl` / `ToggleButtonGroup`, `Banner`, `Button` | Compose | **Prototype-only demonstrator**, not a product feature (see section 4). |
| 46 | Simulated marker (`.badge-sim`, `.sim-marker`) | `Badge`, `Banner` | Compose | Prototype-only; remove in React. |
| 47 | Hero band (`.hero`, forest + lime) | `Section`/`Card` + custom properties | Compose | One per page, restrained. |
| 48 | Layout helpers (`.stack`, `.cluster`, `.grid-*`, `.with-aside`, `.with-filters`, `.section`, `.page-header`) | `Stack`, `Grid`, `Layout`, `Section`, `Divider`, `AspectRatio` | Match? | Responsive props UNVERIFIED; Tailwind permitted only for layout per DESIGN.md. |
| 49 | Visually-hidden text | `VisuallyHidden` | Match | |
| 50 | Icons (`.icon .i-<name>`, inline CSS sprite) | `Icon` | Match? | Icon set and sizing UNVERIFIED; DESIGN.md says Lucide-style acceptable. |
| 51 | Tooltips / hints (not used, `title` avoided) | `Tooltip`, `InfoTip`, `HoverCard`, `Popover` | Optional | |
| 52 | Row/overflow menus (not in prototype; needed for Build/Identity) | `DropdownMenu`, `MoreMenu` | Optional | |
| 53 | Stepper (not in prototype) | `Stepper` | Optional | Possible for multi-step publish flows (P1). |

Count: **53 mapped rows**; **70 distinct verified component names** cited in this section (script-checked against the index list; `LinkedItems` in row 16 is a proposed 226os component, not Astryx).

## 4. Not covered by Astryx (custom components needed)

Per `.claude/skills/web-frontend/SKILL.md` and `ASTRYX_VERIFICATION.md` step 9, each custom primitive needs a **capability gap record** (gap, evaluation, accessible plan, provenance, human review). Below are the **candidate** gaps. None is confirmed: each must be re-checked against the installed `astryx component --list` first.

| Candidate gap | Evidence for gap (index only) | Proposed outcome | Accessible plan | Owner / review |
|---|---|---|---|---|
| Mobile bottom tab bar (`.bottom-nav`) | Index has `SideNav`, `TopNav`, `BottomSheet`, `Drawer`, no bottom bar | 226os product component composed of `Link`/`Icon`/`Text` (or `SideNav` in a horizontal layout if supported) | `<nav aria-label>`, `aria-current`, 44px targets, "More" toggle with `aria-expanded` opening `Drawer` | OPEN: web owner + design review |
| Vote control | No Vote/Rating component in index | Product component `VoteControl` built from `ToggleButton`/`IconButton` + `Text` | group with label, `aria-pressed`, count announced via live region only on change, duplicate-vote handled server side (SPEC) | OPEN |
| Form error summary | Not in index | Product component using `Banner` + `List` + `Link`; focus moves to summary on failed submit | `role="alert"`, links to field ids | OPEN |
| Linked items (question/API/project) | Domain-specific | Product composition over `ClickableCard`/`Link` | Entire card single link, kind label as text | OPEN |
| Country code display | Domain-specific | Product `CountryCode` (`Text` + `VisuallyHidden` full name), never flags | text, not colour | OPEN |
| Dense list row | `List` exists; row anatomy (lead/main/aside + stretched link + container query) UNVERIFIED | Product `ResultRow` over `List`/`ClickableCard` | one link per row, other links above overlay | OPEN |
| Orange secondary-accent variant | Variant set UNVERIFIED | Custom property override first; custom variant only if needed | graphite-on-orange 4.78:1 (light) / 7.78:1 (dark), computed | OPEN |
| "View as" authorization demonstrator | Not a product feature | **Do not port.** In React, authorization is enforced server-side; UI shows the resulting 401/403 state. Keep as storybook/test fixture only | n/a | closed by decision (proposed) |
| Simulated badge / `?state=` forcing | Prototype only | Do not port; replaced by real loading/empty/error states and test mocks | n/a | closed (proposed) |

No gap should be implemented by copying prototype CSS: the prototype is a visual/UX proof; shipped components must come from Astryx or a recorded gap.

## 5. Verification checklist before React implementation (G0-UI-001)

1. Record package manager decision (README uses npm; `DESIGN.md` uses pnpm; `ASTRYX_VERIFICATION.md` lists it OPEN).
2. Pin an exact release (README shows none; Beta, no GitHub releases); record integrity, MIT license and transitive licenses.
3. Install in `apps/web` only; confirm React 19 / Vite compatibility and whether `@stylexjs/stylex` (in the README install) requires any compile step, contradicting or confirming `DESIGN.md` "no StyleX compiler required".
4. Run the installed CLI help; confirm which of `docs global|theme|tokens`, `component --list|--brief-all`, `theme build`, `gap-report` exist. Do not assume.
5. Read each cited component page (`https://astryx.atmeta.com/components/<Name>`) for rows 1-50; fill props, slots, a11y notes in this table; mark rows Match / Compose / Gap definitively.
6. Confirm `defineTheme` (or the real theme API), the token names, mode switching and import/cascade order; map section 2 table; update `DESIGN.md` if its examples are wrong (needs design decision, not a silent edit).
7. Render Solve question detail, Discover API detail, Build contribution request in light/dark/system at 360/768/1024/1440 and 200% zoom (reuse prototype pages as the visual reference).
8. Keyboard/AT: `TabList`, `Dialog` (Esc, focus return), `Toast` (live region), `Selector`/`MultiSelector`, `Field` error association, `Pagination`, `SideNav`; run axe plus manual NVDA/VoiceOver. The prototype has had neither.
9. Re-run contrast on rendered pairs, including hover/active/disabled/focus.
10. Check FR text expansion (~20%) and FR/EN switching without form reset on Astryx components.
11. Check `Markdown`/`CodeBlock` sanitization and bundle weight (JS/CSS) against the performance baseline.
12. Write capability gap records for confirmed gaps (section 4) and obtain human review.
