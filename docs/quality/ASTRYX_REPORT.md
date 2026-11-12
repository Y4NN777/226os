# Astryx verification report (B0 unit 4, ADR-0007)

> Verification snapshot 2026-11-11. Environment: Node 24.21.0 at that time; the current pin is `.nvmrc` 26 (`engines` floor 24 in `src/web/package.json`). Evidence below is not re-run; re-run it on every Astryx bump. Moved from `src/web/ASTRYX_REPORT.md` on 2026-11-12. Protocol: `docs/spec-kit/ASTRYX_VERIFICATION.md`.

Scope: evidence produced while bootstrapping `src/web`. Environment: Node 24.21.0, pnpm 10.32.1, React 19.3.1, Vite 8.3.8. Status words follow AGENTS.md: PASS / FAIL / NOT RUN.

## 1. Package identity

| Item                                 | Value                                                                                                                                                                                                                                                                                                                                                                               |
| ------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Package name                         | `@astryxdesign/core` (correct; repository `github.com/meta/astryx`, directory `packages/core`)                                                                                                                                                                                                                                                                                      |
| Pinned version                       | `0.7.14` (exact, no range) in `src/web/package.json`                                                                                                                                                                                                                                                                                                                                |
| Lockfile integrity                   | `sha512-MSRNyl8/t+0Uo3dmdPZfQi7WNktRZS1UBTicHbQdVIXX3DuCmQ5pPTOHi0OW4jPKszcifV5b6v/mFArjs9oVbA==` (`pnpm-lock.yaml`, identical to `pnpm view` registry value)                                                                                                                                                                                                                       |
| CLI                                  | `@astryxdesign/cli@0.7.14` (devDependency), integrity `sha512-gOTskkep8h0IyeFKCl/vSs/H8CGF8zusiK3hEPWu2naxUzt4fU50OCEhsNpkyJi/0QcOMp9PjfRHlT7wn17mrA==`                                                                                                                                                                                                                             |
| Peer dependency                      | `@stylexjs/stylex@^0.19.0`, pinned `0.19.2` (integrity `sha512-l0fRhOBZ4rFiZ0GntYSlp4lYO5dPT11F3mLb7O4ShK6eE1WOjHtv9GHGQZuzHJbQkjcEaRjPSpBn6itI7yXc3Q==`)                                                                                                                                                                                                                           |
| Peers                                | `react >=19.0.0`, `react-dom >=19.0.0` (React 19.3.1 satisfies)                                                                                                                                                                                                                                                                                                                     |
| Licence                              | MIT (registry metadata and `pnpm licenses`)                                                                                                                                                                                                                                                                                                                                         |
| `@astryxdesign/theme-neutral@0.7.14` | NOT installed. It was installed first, then removed: it brings Figtree and `fonts.css` (remote font, conflicts with DESIGN.md "system fonts only") and its own source is "package theme source" the docs discourage importing into runtime. Core defaults plus a local `defineTheme` are used instead. Consequence: `astryx doctor` reports a `warn` "No theme packages installed". |
| Build scripts                        | pnpm ignored the `postinstall` of `@astryxdesign/core` and `@astryxdesign/cli` (pnpm 10 default; not approved via `pnpm approve-builds`). The core script only prints an `astryx init` nudge. `pnpm install --frozen-lockfile` still exits 0.                                                                                                                                       |

## 2. Transitive licence inventory (production dependencies of `web`, 30 packages)

`pnpm licenses list --filter web --prod`:

- MIT: 28 (includes `@astryxdesign/core`, `@stylexjs/stylex`, TanStack packages, `react`, `react-dom`, `openapi-fetch`, `@formatjs/*`, `styleq`, `seroval*`, `cookie-es`, `scheduler`, `use-sync-external-store`, `csstype`, `@types/react`).
- BSD-3-Clause: 1 (`intl-messageformat@11.2.15`, pulled by Astryx core).
- Unlicense: 1 (`isbot@5.2.3`, pulled by TanStack Router). Needs a CADEV allowed-list decision; not decided here.
- Dev-only tools (ESLint, Vitest, Prettier, jsdom, ...) are not in this inventory. `make licenses` (G4/G7) must formalise it. A `licenses` target now exists (`mk/30-security.mk`, `scripts/licenses.sh`, part of `make quality`); its result is not part of this snapshot: NOT RUN here (no output captured in this report).

## 3. Installed CLI output (captured)

`pnpm exec astryx --version` -> `0.7.14`. `pnpm exec astryx --help` lists: `init`, `component`, `docs`, `blog`, `swizzle`, `gap-report`, `template`, `upgrade`, `theme`, `integration`, `hook`, `discover`, `search`, `build`, `doctor`, `experimental`, `manifest`, `help`.

Commands actually run (read-only; `init` and `theme add` were deliberately NOT run, they write project files):

- `pnpm exec astryx component --detail brief` -> "Components (174)".
- `pnpm exec astryx component Theme Button` and `--props` for Layout, LayoutHeader, LayoutContent, Heading, Text, TextInput, Badge, Banner, VStack, HStack, SegmentedControl, SegmentedControlItem, Link, LinkProvider, InternationalizationProvider.
- `pnpm exec astryx docs`, `docs styling-overview --full`, `docs use-a-theme --full`, `docs author-a-theme --full`, `docs tokens`, `docs tokens/tokens-color`, `tokens-focus`, `tokens-typography`, `docs internationalization/i18n-quick-start --full`.
- `pnpm exec astryx doctor`: node-version ok, core-installed ok, core/cli version alignment ok, themes warn (see section 1).

Differences from DESIGN.md snippets (DESIGN.md is illustrative):

- `pnpm exec astryx component --brief-all` does NOT exist (`error: unknown option '--brief-all'`). Real forms: `component --list`, `component --detail brief`, `component <Name>`, `component <Name> --props`.
- DESIGN.md install line lists `@astryxdesign/theme-neutral`; not used (section 1).
- `astryx docs theme` and `docs tokens` exist (as claimed). `astryx theme build`, `gap-report`, `upgrade` exist.
- DESIGN.md says token `--color-accent` etc. accepted by `defineTheme`: confirmed (`--color-accent`, `--color-background-body`, `--color-background-surface`, `--color-text-primary`, `--color-border`, `--focus-outline-color` all documented in `docs tokens`).

## 4. CSS distribution and compiler requirement

Documented plain-CSS path (`docs styling-overview`): `@import '@astryxdesign/core/reset.css'; @import '@astryxdesign/core/astryx.css';` then a theme. No StyleX compiler or Vite plugin is required for consumption (confirmed: `vite build` succeeds with only `@vitejs/plugin-react`). StyleX compilation is only needed for own `stylex.create()` code or swizzled components; none used. Layers declared by the shipped CSS: `reset`, `astryx-base`, `astryx-theme`. Tailwind is not installed (not needed yet).

Evidence: `pnpm --filter web build` exit 0; CSS bundle 276.26 kB (43.16 kB gzip). Astryx ships token defaults in `astryx.css` (e.g. `--color-background-body` default `#f1f4f7`/`#111112`, `--font-size-base: .875rem`).

## 5. Components used in the shell (all props checked against `astryx component <Name> --props`)

`Theme`, `InternationalizationProvider`, `Layout`/`LayoutHeader`/`LayoutContent` (`role="main"`), `Heading`, `Text`, `VStack`/`HStack`, `Button` (`isLoading`, `variant`), `Badge` (`variant` success/error + `label`), `SegmentedControl`/`SegmentedControlItem`, `TextInput` (`status`, `description`, `htmlName`, `onBlur` passes through to the input). Router links use TanStack `Link` (plain anchor) rather than Astryx `Link`/`LinkProvider`: `LinkProvider` forwards `href`, TanStack needs typed `to`; no adapter was written.

Exported components relevant to P0 (catalog of 174): Button, IconButton, TextInput, TextArea, Field, FormLayout, CheckboxInput, RadioList, Selector, SegmentedControl, Dialog, AlertDialog, Drawer, DropdownMenu, Popover, Tooltip, Toast, Banner, Badge, StatusDot, Skeleton, Spinner, EmptyState, Tab/TabList, Table, Pagination, Breadcrumbs, TopNav, SideNav, MobileNav, AppShell, Layout, Card, Markdown, Code, CodeBlock, Link, Avatar, PowerSearch/CommandPalette, Timestamp, Theme, InternationalizationProvider.

## 6. Accessibility behaviour observed (jsdom, installed release)

- `Button isLoading`: sets `aria-busy` + `aria-disabled`, stays focusable, announces "Loading" in a live region (a second `role=status` appears; tests avoid `getByRole('status')` ambiguity).
- `TextInput status={{type:'error'}}`: input gets `aria-invalid="true"` and `aria-describedby`; the message is rendered visibly AND mirrored in an assertive live region (`role="alert"`), so text queries match twice. Label always rendered; description supported.
- `SegmentedControl`: `role="radiogroup"` with `role="radio"` items, `aria-checked`, roving `tabindex` (selected item `0`, others `-1`).
- `Theme`: the root Theme writes `data-theme` (`light`/`dark`/`system`) and `data-astryx-theme` on `<html>`; observed and asserted in tests.
- `InternationalizationProvider`: ships a `fr-FR` catalog (`@astryxdesign/core/locales/fr-FR.generated.js`) wired so Astryx's own strings follow the app locale. `dir` is not set by the provider (DOM `dir` must be set by the app; LTR only for FR/EN, so not needed now).

## 7. Capability gaps and risks vs DESIGN.md (not custom-built; recorded only)

1. Default body size is 14px; DESIGN.md requires 16px. Mitigated with `typography.scale {base: 16, ratio: 1.2}` in `defineTheme` (documented option; ratio 1.2 is the documented default). Test asserts the generated CSS. Remaining scale tokens (headings) follow the geometric scale and were not reviewed against DESIGN.md's indicative 12/14/16/20/24/32/40.
2. No Astryx slot is documented for a "secondary accent" (DESIGN.md orange `#C67C35`/`#E6AA66` and `on` colour). Orange is not mapped; the no-white-on-orange rule is satisfied by absence. Needs an Astryx-supported slot or a recorded gap before use.
3. Status fills (`--color-error`, `--color-success`, `--color-warning` and their `--color-on-*`) are Astryx defaults. Only status text tokens were mapped and contrast-tested. Badge/Banner fills are UNVERIFIED for WCAG AA with the Graphite & Lime palette.
4. Radius (cards 12-16px, controls 8-12px) and spacing are Astryx defaults, not mapped or checked.
5. Touch target size (aim 44x44px) of Button/SegmentedControl/TextInput at `md` is UNVERIFIED (no browser measurement).
6. Runtime `defineTheme` is used (style injection after mount). The docs recommend `astryx theme build` (static CSS) for SSR; this is an SPA with no component overrides, so no flash is expected, but the built path was not exercised. Revisit when component overrides appear.
7. Bundle: production JS 894.37 kB (279.53 kB gzip), CSS 276.26 kB (43.16 kB gzip); Vite warns about a chunk above 500 kB. No budget exists (DES-07 pending), so no threshold is claimed. Likely contributors: Astryx runtime + `intl-messageformat` + TanStack; not profiled. Route-level lazy loading is the natural follow-up.
8. Beta stability: `0.7.14`, the CLI and docs change between releases. Pin, `astryx upgrade`, and re-run this report on every bump.
9. `astryx doctor` warns that no theme package is installed (by design here).

## 8. ADR-0007 evidence checklist

| Required evidence                                                                                                       | Status                                                                                                            |
| ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Exact version + lockfile integrity                                                                                      | PASS (section 1)                                                                                                  |
| Installed CLI/docs output captured                                                                                      | PASS (section 3)                                                                                                  |
| Licence inventory                                                                                                       | PASS for production inventory (snapshot); Unlicense entry needs policy decision (OPEN); `make licenses` target exists, NOT RUN in this report                     |
| React 19 / Vite compatibility                                                                                           | PASS at the snapshot (typecheck, tests, production build; the test count "66" recorded then is not re-verified)                                                                      |
| Documented CSS import path works without a compiler                                                                     | PASS (build)                                                                                                      |
| Theme API light/dark/system                                                                                             | PASS in jsdom (`data-theme` asserted); rendered visual check NOT RUN                                              |
| Three representative flows (Solve, Discover, Build) rendered in light/dark/system at 360/768/1024/1440 px and 200% zoom | NOT RUN (flows not built; no browser)                                                                             |
| Keyboard and axe results                                                                                                | NOT RUN (Playwright/axe belong to a later unit)                                                                   |
| Contrast of mapped tokens                                                                                               | PASS for the pairs listed in `theme226os.test.tsx` (computed WCAG ratios, light and dark); other pairings NOT RUN |
| Capability-gap record                                                                                                   | Section 7 (human review pending)                                                                                  |
