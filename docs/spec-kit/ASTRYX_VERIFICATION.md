# 226os Astryx Verification Protocol

**Status:** ACCEPTED (ADR-0007, 2026-11-11). The pin-and-verify decision is recorded in ADR-0007 and not repeated here. Executed at B0: see [docs/quality/ASTRYX_REPORT.md](../quality/ASTRYX_REPORT.md) (human review of its open gaps pending).

This document defines what is verified before implementation relies on a particular Astryx release or API. The three representative flows are Solve question detail, Discover API detail and Build contribution request. Record package manager, exact versions, lockfile hash, licence metadata and known beta limitations in the bootstrap PR.

## Verification sequence at bootstrap

1. Inspect the package release metadata and license from the official Astryx source; record exact version and integrity in the dependency inventory.
2. Install only within the approved `src/web` workspace using the approved package manager and lockfile policy.
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

- **Latest floating Astryx range:** faster updates but reproducibility and CSS/API drift risk; rejected.
- **shadcn/ui:** composable alternative but conflicts with the approved Astryx decision; requires a new ADR to replace it.
- **Custom primitives:** maximum control but highest accessibility, maintenance, and provenance risk; allowed only for evidenced gaps.

## Ownership

The beta-mitigation owner is `Y4NN777` (ADR-0002). No agent may claim Astryx integration is verified beyond what the report evidences.
