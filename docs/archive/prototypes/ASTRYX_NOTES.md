> Archived 2026-11-12. Historical record, not maintained.
> Current source: [docs/quality/ASTRYX_REPORT.md](../../quality/ASTRYX_REPORT.md).

# Astryx research notes (vocabulary only)

> **Superseded by ASTRYX_MAPPING.md — second-hand notes, not verified.** Kept only as a record of what the first, partly failed research pass saw.

The HTML prototype does **not** run Astryx. It uses Astryx as vocabulary and reference for naming and behaviour. Only items seen in second-hand sources are listed, with confidence marked per row; one entry (Stepper) was inferred from a type name, as flagged in its row. DESIGN.md and `docs/spec-kit/ASTRYX_VERIFICATION.md` remain the authority: the real component catalogue must be verified with the installed CLI after G0.

## Fetch status (2026-11-11)

| Required source | Result |
|---|---|
| https://github.com/meta/astryx | **Not fetchable** (WebFetch: "Unable to verify if domain github.com is safe to fetch"). Not read. |
| https://astryx.atmeta.com | **Not fetchable** (same error). Not read. |
| https://astryx.atmeta.com/storybook/ | **Not fetchable** (same error). Not read. |
| Secondary sources (read instead) | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md and https://blog.openreplay.com/first-look-astryx-meta-design-system/ were fetched; a web search summary also gave general facts. |

All rows below are therefore **second-hand**: confidence is capped at medium. Re-verify against the official docs/CLI before relying on any of them.

## Components and patterns seen

| Astryx component / pattern | What it is | Source URL | Confidence |
|---|---|---|---|
| `Button` (props `label`, `variant`; variants `primary`, `secondary`) | Clickable control. Prototype analogue: `.btn`, `.btn-primary`, `.btn-ghost` | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md ; https://blog.openreplay.com/first-look-astryx-meta-design-system/ | Medium (two secondary sources agree) |
| Markdown renderer | Formats model output in the AI chat template. Prototype analogue: `.prose` | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Low-medium (one source, generic name) |
| `Stepper` (inferred only from `StepperContextValue` / `useStepperContext`) | A stepper component with a context hook; compact-layout fields removed in 0.6.0 | https://blog.openreplay.com/first-look-astryx-meta-design-system/ | Low (component name itself is inferred from a type name; not used in the prototype) |
| Template: Issues tracker | Jira-style layout with grouped rows and dropdown menus (inspiration for dense list rows) | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| Template: E-commerce checkout form | Form with validation states (inspiration for `.field.has-error`) | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| Template: Kanban sprint board | Board with drag-and-drop cards (not used) | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| Template: IDE interface | VS Code-style layout with file tree and tabbed editor (inspiration for code/tab density) | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| Template: AI chat interface | Chat UI rendering formatted output (relevant to the future Intelligence domain) | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| `defineTheme` from `@astryxdesign/core/theme` | Type-safe theme definition (name, colors, typography, radius, motion, overrides). Prototype analogue: `tokens.css` | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| `tokens` object (e.g. `--color-neutral` with light and dark values) | Sets CSS variables per theme mode. Prototype analogue: semantic tokens with light/dark values | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| `components` object / `xstyle` prop | Per-component variant overrides / StyleX one-off overrides (no prototype analogue) | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| Tailwind bridge (`tailwind-theme.css` per DESIGN.md) | Maps Astryx tokens to Tailwind utilities | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| Cascade layers `@layer reset` / `@layer astryx-base`; stylesheets `reset.css`, `astryx.css`, `theme.css` | CSS distribution and import order | https://blog.openreplay.com/first-look-astryx-meta-design-system/ | Medium |
| Packages `@astryxdesign/core` (React 19+ peer), `theme-neutral`, `lab` (experimental components), `charts` / `vega` (canary), `build`, `cli` | Distribution layout | https://blog.openreplay.com/first-look-astryx-meta-design-system/ | Medium |
| CLI: `component <Name>`, `init`, `swizzle`, `upgrade --apply`, `theme list/add`, `docs styling --dense` | Agent-readable docs and ejecting | https://blog.openreplay.com/first-look-astryx-meta-design-system/ ; https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |
| MCP server | Gives agents structured access to components, templates, tokens | https://betterstack.com/community/guides/ai/astryx-meta-design-system.md | Medium |

## Not verified (do not assume)

- Whether Astryx ships Dialog, Tabs, Table, Toast, Chip/Badge, Avatar, Breadcrumbs, Pagination or form inputs. No fetched source named them. The prototype's classes for these are original CSS written to DESIGN.md rules, not Astryx copies.
- Component counts differ across sources (">90" in the repo per one summary, ">150" on the docs site per another). Not relied on.
- Theme/token names beyond `--color-neutral`, `--color-accent`, `--color-background-surface`, `--color-text-primary` (the last three appear only in DESIGN.md as illustrative).
