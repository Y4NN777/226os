# 226OS — Design System & Frontend Experience Specification

**Status:** Approved direction; implementation subject to validation gates  
**Version:** 1.0 · 11 November 2026  
**Owner:** 7's · 226OS  
**Applies to:** `apps/web`, shared UI packages, documentation and agent-authored UI  
**Related:** `docs/PRD.md`, `docs/SPEC.md`, `docs/architecture/ARCHITECTURE.md`, `docs/DEV.md`, `docs/TEST.md`, `docs/RULES.md`, `AGENTS.md`

> **Decision record.** The team selected the **Graphite & Lime** direction shown in the 226OS dashboard exploration and chose **Meta Astryx** as the primary UI design system for **React 19 + TypeScript + Vite**. This document is the implementation contract, not a request to recreate the mockup pixel by pixel. Further visual polish must follow this system and explicit review. The illustrated dashboard, its people, numbers, imagery and events are illustrative—not shipped data or approved brand assets.

## 1. Product context and design intent

226OS connects African developers through technical questions, shared knowledge, African APIs and tools, open-source projects, collaboration, communities, mentorship, careers, multilingual access, and context-aware assistance. The interface must make these experiences feel like one product, not eight unrelated applications.

**Design attributes:** capable, clear, welcoming, technically credible, contemporary, purposeful.  
**Avoid:** generic admin-template aesthetics, dense decorative graphics, culture-as-decoration, neon overload, exaggerated shadows, fake activity, inaccessible brand-colored text, and layouts that collapse on mobile.

The **226** acknowledges the project's Burkinabè origin, while 226OS serves developers across Africa. Design and copy must not imply that one country represents the continent. The interface is **French/English**, responsive, keyboard-usable, and mindful of low-bandwidth users from day one.

### Experience principles

1. **Tasks before decoration.** Solving a question, finding a relevant API, contributing to a project and contacting another developer are the primary journeys.
2. **Knowledge stays readable.** Code, long-form answers, metadata, documentation and search results take precedence over dashboard ornaments.
3. **Connection across domains.** Surface meaningful links between a question, an API, a project and its contributors.
4. **Progressive disclosure.** Advanced options should not obscure core tasks.
5. **Accessible by default.** Semantic controls, contrast, focus, touch targets and reduced motion are mandatory.
6. **Performance is a product feature.** No heavy media or animation is necessary for core flows.
7. **Honest UI.** Do not manufacture metrics, endorsements, events, contributions, AI answers or content availability.

## 2. Approved color direction — Graphite & Lime

The following are **226OS brand foundation colors**. Semantic CSS tokens are the actual API for application components.

| Role | Token | Hex | Use |
| --- | --- | --- | --- |
| Graphite | `brand.graphite` | `#1D2420` | Primary text, dark ink and technical tone |
| Deep green | `brand.forest` | `#2B440E` | Promotional surfaces, dark-green hero treatments |
| Primary green | `brand.primary` | `#4D721D` | Primary actions, navigation selection, interactive emphasis |
| Lime accent | `brand.lime` | `#BDE874` | Limited dark-surface highlights, not body text on white |
| Warm orange | `brand.orange` | `#C67C35` | Secondary highlights, occasional labels and illustrations |
| Warm paper | `brand.canvas` | `#F6F7F4` | Light application background |
| White | `brand.white` | `#FFFFFF` | Cards and high-contrast content surfaces |

**Contrast observations, not full certification:** `#4D721D` against white is approximately **5.61:1** (acceptable for normal text); `#C67C35` against white is **3.31:1**, so orange must **not** be used for ordinary white-text buttons. Use graphite-on-orange, or an approved darker orange for white text. Every actual pairing and state must still be measured automatically.

### 2.1 Semantic tokens — reference design values

These are **226OS target values** to map onto Astryx token slots. They are not a claim that the names below are built-in Astryx variables. Do not replace Astryx internal token definitions or introduce two competing systems.

| Semantic role | Light mode | Dark mode | Notes |
| --- | --- | --- | --- |
| App background | `#F6F7F4` | `#111710` | Quiet canvas |
| Primary surface | `#FFFFFF` | `#1B241A` | Cards, panels and dialogs |
| Raised surface | `#FFFFFF` | `#243021` | Menus and elevated surfaces |
| Subtle surface | `#EDF2E7` | `#2C3926` | Selected backgrounds and tags |
| Primary text | `#1D2420` | `#F4F7EF` | Body and headings |
| Secondary text | `#576454` | `#C1CCB9` | Metadata and secondary copy |
| Border | `#DCE3D6` | `#43503C` | Containers and dividers |
| Brand action | `#4D721D` | `#BDE874` | Action foreground/background only in contrast-safe pairings |
| On brand action | `#FFFFFF` | `#16220E` | Paired with brand action |
| Secondary accent | `#C67C35` | `#E6AA66` | Use sparingly; test text contrast |
| On secondary accent | `#1D2420` | `#1D2420` | Never default to white |
| Focus indicator | `#275DCA` | `#9CC4FF` | Visibly distinct from brand green |
| Success | `#28683F` | `#8CDEA0` | Status, not a second primary brand |
| Warning | `#8A540D` | `#F3C46F` | Status plus icon/text |
| Error | `#B42318` | `#FFB0A8` | Error, not orange brand accent |

**Dark mode** is an approved design requirement, but the exact dark token set is a **proposed implementation baseline** until contrast, component and usability checks pass. The selected brand direction itself is final; refinements to semantic values are allowed only with documented evidence and design review.

### 2.2 Color usage rules

- Main primary action: solid green; one dominant CTA per region.
- Orange: complementary emphasis, never a substitute for warning or error.
- Lime: never used as ordinary text against the light canvas; reserve for highlights on deep green / dark surfaces where contrast is proven.
- White / neutral surfaces should remain predominant. Do not fill every card with colored backgrounds.
- Focus ring must be evident for keyboard users even on green buttons.
- Never rely on green/orange alone to indicate state; include label, icon or shape.
- Gradients are optional and restricted to limited promotional or onboarding areas—not data tables, code samples or forms.
- Never infer country identity solely from color or flags.

## 3. Official component framework: Meta Astryx

**Decision:** Astryx is the **primary** accessible component foundation. Do not scaffold a parallel shadcn/ui library or recreate Astryx components unless a documented capability gap is demonstrated. This supersedes the previous unresolved Astryx-vs-shadcn comparison. Keep Tailwind only for layout and site-specific styling when useful, not as a second component system.

**Compatibility:** React **19+**, Vite, TypeScript. Use precompiled Astryx CSS first; **no dedicated StyleX compiler or Vite plugin is required** for this consumer path. Astryx itself is built using StyleX; that does not require 226OS to compile all its component source. Astryx is MIT-licensed; the final dependency inventory and transitive licenses must still pass the CADEV policy checks.

### 3.1 Installation and implementation contract

```bash
# Run within apps/web (or via workspace tooling).
pnpm add @astryxdesign/core @astryxdesign/theme-neutral @stylexjs/stylex
pnpm add -D @astryxdesign/cli
```

Use a **pinned, reviewed** release recorded in the lockfile. Verify the specific component API with its installed CLI before generating code, because Astryx is currently in beta and APIs may change.

```css
/* apps/web/src/styles/index.css — conceptual import order */
@import '@astryxdesign/core/reset.css';
@import '@astryxdesign/core/astryx.css';
@import '@astryxdesign/theme-neutral/theme.css';
/* Then import the generated 226OS theme overrides and app styles. */
```

For a Tailwind integration, follow **Astryx's official cascade-layer order** and use its optional `tailwind-theme.css` token bridge. Never guess the order: the upstream docs demonstrate that a wrong cascade can silently override components. The Astryx `Theme` provider should wrap the application once, using a 226OS theme created with the supported `defineTheme()` API and an explicit light/dark/system mode strategy.

**Theme authoring guidance (illustrative; verify against installed release):**

```tsx
import {defineTheme} from '@astryxdesign/core/theme';

export const theme226OS = defineTheme({
  name: '226os',
  tokens: {
    '--color-accent': ['#4D721D', '#BDE874'],
    '--color-background-surface': ['#FFFFFF', '#1B241A'],
    '--color-text-primary': ['#1D2420', '#F4F7EF'],
  },
});
```

The tokens above illustrate how to map an approved palette; **confirm token support and the generated output** using `astryx docs theme` and `astryx docs tokens` from the pinned release. Compile a production static theme when supported using `astryx theme build`. Do not paste guessed component props or fabricate component names.

### 3.2 Agent-first Astryx workflow

Before an agent authors a UI component:

1. Read this file plus the relevant requirement in `docs/SPEC.md` and the existing UI patterns.
2. Inspect the **installed** Astryx CLI documentation: `pnpm exec astryx component --brief-all`, `pnpm exec astryx docs tokens`, `pnpm exec astryx docs theme`, and a targeted `pnpm exec astryx component <ComponentName>`.
3. Prefer verified Astryx primitives and compositions; use typed, composable props and CSS variable overrides.
4. Build reusable **226OS product components** out of Astryx primitives instead of patching vendor code.
5. If an Astryx component is missing or deficient, record the gap, evaluation and outcome before a custom accessible component is approved. `swizzle` is an exceptional escape hatch that brings ongoing maintenance and attribution responsibilities.
6. Add interaction, keyboard and visual tests; do not call a component complete because it renders.
7. Keep version-specific examples synchronized with the installed lockfile.

Useful upstream commands (verify current syntax against the installed CLI): `astryx component`, `astryx docs`, `astryx init`, `astryx theme build`, `astryx gap-report`, `astryx upgrade`.

## 4. Product foundations

### Typography

- **Default:** system sans-serif (`system-ui`, `-apple-system`, `Segoe UI`, `Roboto`, sans-serif) for immediate rendering and bilingual support.
- **Code:** system monospace, e.g. `ui-monospace`, `SFMono-Regular`, `Consolas`, `Liberation Mono`, monospace.
- No remote font downloads on critical paths. Any replacement font needs licensing, glyph coverage, bandwidth and accessibility review.
- Body: 16px baseline, 1.5–1.6 line-height; secondary UI text 14px minimum where practical. Use 12px only for ancillary captions that remain readable.
- Indicative type scale: 12 / 14 / 16 / 20 / 24 / 32 / 40; use semantic headings and meaningful hierarchy. Do not use all-caps on long text.

### Spacing, sizing and shape

- Base spacing grid: **4px**; common steps **4, 8, 12, 16, 24, 32, 48**.
- Card radii: approximately 12–16px; controls: approximately 8–12px. Final implementation uses Astryx radius tokens rather than arbitrary per-page radii.
- Prefer low-elevation borders over heavy shadows; reserve shadows for overlays and key raised surfaces.
- Interactive targets: aim for 44 × 44px where feasible on touch devices; never below WCAG 2.2's minimum target requirements without an applicable exception.
- Content widths: generous for dashboards, constrained for reading and discussion detail pages; max reading line length ~70–85 characters.
- Layering: app shell < sticky navigation < menus < dialogs < notifications; avoid arbitrary `z-index` values.

### Iconography and illustrations

- Use a coherent icon family via Astryx-supported patterns (Lucide-style icons are acceptable). Consistent stroke and optical sizing.
- No AI-generated or decorative images are required for implementation. The previous mockup's African-map hero and illustrations are **optional references only**, not mandated assets.
- SVG and CSS should cover most decorative needs. Avoid enormous hero images, stock photos and culturally reductive symbolism.
- All meaningful images need alt text; purely decorative images need empty alt text and must not distract screen readers.

## 5. Layout and responsive behavior

### Navigation model

- Desktop: stable left navigation plus global search; top-right account controls and an intentional creation action.
- Tablet: compact or collapsible rail; search remains accessible.
- Mobile: simplified navigation or accessible drawer; preserve direct access to Solve, Discover and Build.
- Group pages by user intent, not internal backend module names.
- Unimplemented areas must be clearly labeled as planned/preview or hidden. No dead navigation or fake buttons.

### Product surfaces

| Surface | Required pattern | Avoid |
| --- | --- | --- |
| Dashboard / Explore | Relevant feed, active projects, recent solutions, meaningful next actions | Arbitrary counters, repeated promo boxes |
| Solve | Search and filters → question → answer/code → accepted resolution | Tiny code typography, overflow traps |
| Discover / API Hub | Country/service/technology filters → resource detail with sources and supported regions | Unverified provider claims |
| Build / Open Source | Project listing → detail → contribution needs → application/interest | Project cards with no concrete next step |
| Profile / Community | Skills, contributions, country, visible privacy settings | Reputation without a defined meaning |
| Mentorship / Careers | Opportunity criteria and honest availability | Fictional jobs, mentors or matching |
| Intelligence | Source-linked assistance and transparent limitations | Unattributed claims, invented answers |

### Responsive checkpoints

Test at **360px**, **768px**, **1024px**, **1440px**, and at 200% browser zoom. No unintended horizontal document scrolling; code blocks may horizontally scroll within their own labeled region. Dialogs, tables and menus must remain usable at 360px.

## 6. Interaction, content and accessibility

Each actionable component must consider: **default, hover, focus-visible, active, disabled, loading, success, error, empty** (where relevant).

- Use accessible field labels, descriptions, validation summaries and error recovery.
- Favor native semantics; preserve keyboard navigation, escape behavior, dialog focus management and screen-reader announcements.
- Communicate status through text + semantics, not color alone.
- Respect `prefers-reduced-motion`, `prefers-contrast` when feasible, and OS text scaling.
- Do not auto-translate public user content without disclosure; original text remains accessible.
- Language switch must not discard form edits or context.
- Code snippets: readable syntax, copy controls, language label, horizontal scroll where needed, and sanitization for untrusted Markdown.
- User-generated avatars, external logos and links require privacy/security and performance review.
- Localized empty/loading/error states must be authored, not generic placeholders.

**WCAG 2.2 AA targets:** normal text **≥ 4.5:1**, large text **≥ 3:1**, and meaningful non-text UI states **≥ 3:1**. These are **release gates**, not an assertion that the entire design is already compliant.

## 7. Frontend implementation conventions

Approved frontend stack: **React 19 + TypeScript + Vite, TanStack Router, Query and Form, Astryx, Tailwind only where complementary**.

Suggested layout (adapt to the repo's accepted architecture, do not duplicate packages):

```text
apps/web/src/
  app/                  # app shell, providers, route setup
  routes/               # TanStack Router route definitions
  features/             # solve, discover, build, community, etc.
  components/           # cross-domain product compositions
  design-system/        # 226OS theme, token mapping, approved wrappers
  styles/               # application CSS and imports
  i18n/                 # French and English message catalogs
```

- Prefer Astryx source of truth for component behavior and CSS properties. UI compositions live in app code, not in a private fork of Astryx.
- Product components are domain-agnostic only when reused meaningfully; avoid abstraction solely for aesthetic consistency.
- Use TanStack Router for navigation, TanStack Query for server state and TanStack Form for forms; do not independently invent parallel state-management layers.
- Never bypass established API contracts for convenience.
- No arbitrary per-component hex colors; use semantic 226OS theme mappings. Local visualization colors require deliberate exceptions.
- Avoid ad hoc `!important`, inline spacing hacks and copying identical markup across pages.
- Build dark mode from the same semantic system, not a parallel stylesheet maintained by hand.

## 8. Quality gates and measurable acceptance

A design/UX pull request is eligible for review only when:

- [ ] The UI satisfies an identified PRD/SPEC requirement and shows a usable user journey.
- [ ] Existing Astryx components were evaluated before custom primitives were added.
- [ ] No unauthorized change to shared palette, token roles, Astryx dependency or global layout foundations occurred.
- [ ] Light and dark variants render correctly; no contrast violations in mandatory user journeys.
- [ ] Keyboard-only navigation, screen-reader labeling, focus and responsive behavior were verified.
- [ ] French and English strings are present and layout is checked for text expansion.
- [ ] Loading, empty, success and error states are tested as relevant.
- [ ] Automated component/accessibility checks and a human review of representative flows are recorded.
- [ ] Frontend coverage **exceeds 70%** under `docs/TEST.md` (including approved metric definitions) and no critical flow is untested.
- [ ] No critical or high vulnerabilities, unsafe content rendering or exposed secrets remain.
- [ ] Measured page weight, LCP/INP/CLS and relevant API latency conform to **approved** budgets; do not fabricate measurements or silently invent thresholds.
- [ ] Existing approved visual baselines, if available, are reviewed for regressions.

Quality gates must run through the documented `make` targets and GitHub Actions as defined in `docs/TEST.md` and `docs/SHIP.md`. Evidence includes command, environment, exit status, coverage/a11y reports and screenshots only when needed for review. Agents may propose exceptions but cannot approve their own.

## 9. Governance and change control

| Decision | Status | Scope |
| --- | --- | --- |
| DES-01 Graphite & Lime brand direction | **Approved** | Palette family and visual character |
| DES-02 Meta Astryx as primary component system | **Approved** | React/Vite UI foundations |
| DES-03 Detailed light semantic palette | **Baseline; validate in implementation** | Contrast and component-state refinements |
| DES-04 Detailed dark semantic palette | **Proposed baseline; approval after checks** | Contrast and usability |
| DES-05 System typography | **Approved default** | Change via evidence-backed ADR |
| DES-06 Final logo and illustration assets | **Not approved** | No invented identity assets |
| DES-07 Content and performance budgets | **Pending measured baselines** | TEST/SHIP alignment |

Changes to **approved** brand color roles or the primary component library require a human-approved ADR. Routine fixes for failing accessibility tests may adjust mapped token values with documented before/after evidence and review, without re-opening the entire brand direction.

### Required validation sequence

1. Review official Astryx docs and lock compatible versions.
2. Implement 226OS theme mapping and minimal UI foundations.
3. Test three representative flows: **Solve question detail**, **Discover API detail**, **Build contribution request** (light and dark, desktop and mobile).
4. Run automated accessibility/contrast and frontend test gates, plus keyboard and mobile manual checks.
5. Record deviations, supported token mappings, measured performance and design decision evidence.
6. Proceed to wider implementation only after the shared components and flows pass review.

## 10. Authoritative references

- Meta Astryx repository, license and project overview: https://github.com/meta/astryx
- Astryx core quick start (React, Vite, CSS import order): https://github.com/meta/astryx/blob/main/packages/core/README.md
- Astryx system architecture and theming: https://github.com/meta/astryx/wiki/System-Architecture
- Astryx distribution and agent-readable docs: https://github.com/meta/astryx/wiki/Distribution
- Astryx API conventions and component patterns: https://github.com/meta/astryx/wiki/API-Conventions
- Astryx docs and component catalog: https://astryx.atmeta.com/
- WCAG 2.2: https://www.w3.org/TR/WCAG22/
- TanStack Router: https://tanstack.com/router/latest

**Agent directive:** This file governs the 226OS UI. Read it before making frontend changes; implement the selected direction through the installed Astryx version, not by copying the mockup or inventing UI APIs. On conflict, flag the disagreement with file evidence and request a design decision rather than silently changing architecture or product scope.
