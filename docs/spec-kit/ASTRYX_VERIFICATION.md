# 226OS Astryx Verification Protocol

**Status:** PROPOSED verification plan; no package installation or CLI execution performed

## Approved decision preserved

Meta Astryx is the approved primary component foundation under `DESIGN.md` and the existing architecture baseline. This document does not reopen that decision. It defines what must be verified before implementation relies on a particular release or API.

## Proposed default

Pin the first release that satisfies React 19 compatibility, documented CSS distribution, theme support, keyboard/focus behavior, and required components for the three representative flows: Solve question detail, Discover API detail, and Build contribution request. Do not use a floating version range. Record package manager, exact versions, lockfile hash, license metadata, and known beta limitations.

This is a proposal, not a version approval.

## Verification sequence after human G0 authorization

1. Inspect the package release metadata and license from the official Astryx source; record exact version and integrity in the dependency inventory.
2. Install only within the approved `apps/web` workspace using the approved package manager and lockfile policy.
3. Run the installed CLI's own help/brief documentation commands, or the release-equivalent documented commands. Do not assume the illustrative commands in `DESIGN.md` exist.
4. Verify the actual component catalog for buttons, inputs, forms, dialogs, menus, navigation, status, code/content, and theme/provider primitives needed by P0.
5. Verify CSS reset/core/theme import order and whether the consumer path requires StyleX compilation or only precompiled CSS.
6. Verify theme token APIs against installed types/docs. Map Graphite & Lime semantic roles without replacing Astryx internals or adding arbitrary per-component hex colors.
7. Render the three representative flows in light, dark, and system modes at 360px, 768px, 1024px, and 1440px plus 200% zoom.
8. Check keyboard navigation, focus-visible states, dialog escape/focus return, labels, reduced motion, contrast, loading/error/empty states, and FR/EN text expansion.
9. Record capability gaps. A custom primitive requires a documented gap, accessible implementation plan, license/provenance record, and human review.

## Decision criteria

| Criterion | Required result | Evidence |
|---|---|---|
| React/Vite compatibility | compatible with approved React 19/Vite baseline | package metadata and build/type evidence |
| License | direct/transitive inventory passes CADEV policy review | SPDX/SCA report |
| CSS distribution | documented import order works without hidden compiler requirement | minimal render and build evidence |
| Theme API | installed API supports semantic light/dark/system mapping | typed example and rendered states |
| Accessibility | representative components meet WCAG 2.2 AA targets | automated and manual evidence |
| Stability | no blocking beta issue in P0 primitives, or approved mitigation | gap report and human decision |
| Performance | no unnecessary media/JS cost for core flows | page-weight and throttled-network baseline |

## Alternatives

- **Latest floating Astryx range:** faster updates but reproducibility and CSS/API drift risk; not proposed.
- **shadcn/ui:** composable alternative but conflicts with the approved Astryx decision; requires a human ADR to replace it.
- **Custom primitives:** maximum control but highest accessibility, maintenance, and provenance risk; allowed only for evidenced gaps.

## G0 blocker

The exact Astryx version, package manager, lockfile policy, verified component/API set, and beta mitigation owner remain `OPEN`. No agent may claim Astryx integration is verified before this protocol produces evidence.
