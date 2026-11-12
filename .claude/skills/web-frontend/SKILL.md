---
name: web-frontend
description: 226os frontend rules for React 19 + strict TypeScript + Vite + TanStack Router/Query/Form on a pinned Meta Astryx foundation with Graphite & Lime tokens, FR/EN i18n, WCAG 2.2 AA, mobile/low-bandwidth, typed OpenAPI client and CSRF-protected mutations. Use when touching src/web/**.
---
# 226os Web Frontend

## Scope and triggers
- Any create/edit under `src/web/**`: routes, features, product components, theme/token mapping, i18n catalogs, API client wiring, frontend tests.
- Reading `api/openapi.yaml` to consume it. The frontend slice lists it as read-only in its allowed paths (slice block in `TODO.md`), and contract changes belong to the API owner and need the frontend owner's review (ADR-0008) — never change the contract to fit the UI (`DESIGN.md:226`, `docs/RULES.md:13`).
- Precondition: an open slice in `TODO.md` (S7 or the slice you were given) with named owner and reviewer and its allowed paths; the Astryx pin is decided by ADR-0007. Otherwise stop and ask the user to run `/start-slice`.
- Load `engineering-principles` too. Follow the root `DESIGN.md` on every UI change (`DESIGN.md:284`).

## Project rules (binding)
**Stack and foundation**
- Stack: React 19, strict TS, Vite, TanStack Router/Query/Form, Astryx primary, Tailwind for layout only and never a second component system (`DESIGN.md:83`, `DESIGN.md:208`, `docs/DEV.md:6`, `docs/archive/spec-kit/CONTRACTS.md:39`).
- No shadcn/ui and no re-creating Astryx components. A custom primitive needs a recorded gap, an accessible plan, provenance and human review (`DESIGN.md:83`, `DESIGN.md:132`, `docs/spec-kit/ASTRYX_VERIFICATION.md:25`).
- Astryx is pinned by ADR-0007 (exact version in `src/web/package.json`, lockfile committed); the verification record is `docs/spec-kit/ASTRYX_VERIFICATION.md` (section on version verification). CLI syntax, component APIs, token names and CSS import order must still be verified against the installed package. The snippets in DESIGN.md are illustrative, not normative (`docs/archive/spec-kit/CONTRACTS.md:44`, `DESIGN.md:122`). Never write an Astryx component name, prop or token from memory. Read the installed package's docs, types and CLI output first.
- Use the pinned release, no floating range. Package manager is pnpm with a committed lockfile (ADR-0006, ADR-0007).
- Use precompiled Astryx CSS first. Follow the upstream cascade-layer order exactly and never guess it (`DESIGN.md:85`, `DESIGN.md:105`).
- Wrap the app in one theme provider with an explicit light/dark/system strategy (`DESIGN.md:105`).

**Tokens and visual rules**
- Map Graphite & Lime onto Astryx token slots. Don't replace Astryx internals and don't add competing token systems (`DESIGN.md:48`, `docs/spec-kit/ASTRYX_VERIFICATION.md:22`).
- No per-component hex values. Avoid ad hoc `!important`, inline spacing hacks and duplicated markup (`DESIGN.md:227-228`).
- Derive dark mode from the same semantic tokens, never from a hand-maintained stylesheet (`DESIGN.md:229`). Dark values are a proposed baseline (DES-04, `DESIGN.md:257`).
- Colour rules:
  - Lime is never body text on the light canvas (`DESIGN.md:74`).
  - Never put white text on `#C67C35` orange (3.31:1) (`DESIGN.md:44`).
  - Orange never means warning or error (`DESIGN.md:73`).
  - The focus ring (`#275DCA` / `#9CC4FF`) must stay distinct from brand green (`DESIGN.md:63`, `DESIGN.md:76`).
  - Never use colour alone to signal state (`DESIGN.md:77`, `DESIGN.md:196`).
- Typography and layout:
  - System font stacks only. No remote fonts on critical paths (`DESIGN.md:142-144`).
  - 16px body text (`DESIGN.md:145`).
  - 4px spacing grid (`DESIGN.md:150`).
  - Radii come from Astryx tokens (`DESIGN.md:151`).
  - Use the layering scale, not arbitrary `z-index` values (`DESIGN.md:155`).
- No dead navigation or fake buttons. Unbuilt areas are hidden or labelled preview (`DESIGN.md:172`). Honest UI means no invented metrics, activity or AI answers (`DESIGN.md:28`).

**States, accessibility and responsiveness**
- Every actionable component handles default, hover, focus-visible, active, disabled, loading, success, error and empty states (`DESIGN.md:192`).
- Write localized loading, empty and error copy. Generic placeholders are not allowed (`DESIGN.md:202`, `docs/SPEC.md:80`).
- WCAG 2.2 AA is a release gate: text ≥4.5:1, large text ≥3:1, non-text UI ≥3:1 (`DESIGN.md:204`).
- Targets aim for 44×44px (`DESIGN.md:153`).
- Respect reduced motion and OS text scaling (`DESIGN.md:197`).
- Responsive checkpoints: 360, 768, 1024 and 1440px plus 200% zoom. No page-level horizontal scroll. Code blocks scroll inside their own labelled region (`DESIGN.md:188`, `docs/archive/spec-kit/CONTRACTS.md:43`).
- Low bandwidth: core flows must not depend on heavy media (`docs/SPEC.md:76-81`, `DESIGN.md:27`).
- Performance budgets are pending (DES-07). Don't invent thresholds and don't fabricate measurements (`DESIGN.md:245`, `DESIGN.md:260`).

**Internationalization**
- FR/EN UI. No hard-coded text on critical screens (`docs/DEV.md:54`).
- Switching locale changes UI labels only. Original user content stays unchanged and unsaved form edits survive (`docs/SPEC.md:66-70`, `DESIGN.md:198-199`, `docs/archive/spec-kit/CONTRACTS.md:42`).
- Locale persistence and fallback are `OPEN`.

**Data, contract and security**
- Status: the CSRF, idempotency, problem-field and cookie rules below are fixed by ADR-0003 and ADR-0004 and by `api/openapi.yaml`. When in doubt, the contract wins.
- Server state goes through TanStack Query with deterministic keys and documented invalidation. Don't duplicate it in a global store (`docs/DEV.md:52`, `DESIGN.md:225`).
- Use Router for typed navigation, with loaders and errors handled (`docs/DEV.md:53`).
- Forms use TanStack Form (`DESIGN.md:225`).
- The API client is typed and generated from OpenAPI 3.1. No hand-written DTOs, no invented endpoints, drift is checked (`docs/DEV.md:51`, `docs/RULES.md:13`, `docs/RULES.md:15`).
- `api/openapi.yaml` exists; the typed client is generated from it by `make generate` (openapi-typescript, ADR-0004). Never hand-write DTOs.
- Errors are `application/problem+json` (RFC 9457, `docs/SPEC.md:99`):
  - Branch on the stable `code`/`status` (`docs/spec-kit/API_CONTRACT.md`, section on errors).
  - `detail` may be localized on the client.
  - Map `violations[].field` to form fields.
  - Show `request_id` as a support reference.
  - Honour `Retry-After` on 429 and 503 responses (`docs/spec-kit/API_CONTRACT.md:58`).
- Sessions use HttpOnly cookies, so the frontend never reads or stores a session token.
- Every state-changing request sends the synchronizer CSRF token (`docs/spec-kit/AUTHORIZATION_MATRIX.md:12`, `docs/spec-kit/AUTHORIZATION_MATRIX.md:50`, `docs/archive/spec-kit/G0_DECISIONS.md:10`). How the token is delivered and its header name are `OPEN`.
- No bearer tokens (`docs/spec-kit/API_CONTRACT.md:11`).
- Retried creates (applications, votes, relations) send an `Idempotency-Key` that is reused across retries (`docs/spec-kit/API_CONTRACT.md:12`).
- Authorization is enforced on the server only. UI hiding is cosmetic (`docs/architecture/ARCHITECTURE.md:19`, `docs/spec-kit/AUTHORIZATION_MATRIX.md:46`).
- Sanitize untrusted Markdown and snippets. Never execute user code (`DESIGN.md:200`, `docs/SPEC.md:126`, `docs/TEST.md:37`).
- External avatars, logos and links need privacy and performance review (`DESIGN.md:201`).

## Expert practices
**React and TypeScript**
- Derive values during render. Keep effects for synchronizing with external systems only ([react.dev: You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect)).
- Use the React 19 `ref` prop instead of `forwardRef`. `use`, `useOptimistic` and document metadata are available ([react.dev React 19](https://react.dev/blog/2024/12/05/react-19)).
- Don't build a second form-state system with `useActionState` or `<form action>`. TanStack Form is mandated (see Tensions).
- Enable `strict`. Consider `noUncheckedIndexedAccess` and `exactOptionalPropertyTypes` (team decision). Ban `any` in favour of `unknown` plus narrowing ([TSConfig strict](https://www.typescriptlang.org/tsconfig/#strict)).

**TanStack**
- Define one `queryOptions` per resource, built by a key factory such as `['questions', 'list', filters]`. Every variable used by the query function goes in the key ([TanStack Query keys](https://tanstack.com/query/latest/docs/framework/react/guides/query-keys), [Effective keys](https://tkdodo.eu/blog/effective-react-query-keys)).
- Prefetch in route loaders with the same options object the component reads. Pass `queryClient` through router context ([Router external data loading](https://tanstack.com/router/latest/docs/framework/react/guide/external-data-loading)). The exact loader/prefetch method names are UNVERIFIED for the pinned version.
- After a mutation, invalidate the narrowest affected keys. Only use optimistic updates when rollback is defined and the server stays authoritative. Vote counts always come from the server (`docs/SPEC.md:111`).
- Forms:
  - Run sync validators on blur or submit, and async validators debounced.
  - Map server `violations` back to fields through the form-level submit validator.
  - Gate submission with `canSubmit` and disable it while pending to prevent double submits ([TanStack Form validation](https://tanstack.com/form/latest/docs/framework/react/guides/validation)).
  - This backs "second click creates no duplicate" (`docs/SPEC.md:50`).

**Vite and CSRF**
- Never put secrets in `VITE_*` variables, because they are inlined into the bundle ([Vite env](https://vite.dev/guide/env-and-mode)).
- Lazy-load heavy routes and code-highlighting (226os convention; `docs/DEV.md:53`).
- Send the CSRF token in a custom header on mutations to trusted same-origin endpoints only. Never send it in a URL, log or global default, and never store it in a cookie for the synchronizer pattern ([OWASP CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)).

**Accessibility**
- Prefer native elements to ARIA. When a custom widget is unavoidable, follow its APG pattern exactly: roles, keyboard and focus ([WAI-ARIA APG patterns](https://www.w3.org/WAI/ARIA/apg/patterns/)).
- WCAG 2.2 additions to check ([WCAG 2.2](https://www.w3.org/TR/WCAG22/)):
  - 2.4.11 Focus Not Obscured: sticky headers must not hide focus.
  - 2.5.7 Dragging Movements: provide a single-pointer alternative.
  - 2.5.8 Target Size: ≥24×24 CSS px ([Understanding 2.5.8](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html)). The project aims for 44.
  - 3.3.7 Redundant Entry.
  - 3.3.8 Accessible Authentication: allow paste and password managers on sign-in.
- Set `lang` on user-content blocks whose language differs from the UI (WCAG 3.1.2). Whether the API exposes content language is `OPEN`.
- Move focus to the page `h1` or main region on route change, and announce async results through a polite live region ([APG practices](https://www.w3.org/WAI/ARIA/apg/practices/); 226os convention).
- Measure LCP, INP and CLS on throttled mobile. Report measured values only ([web.dev Web Vitals](https://web.dev/articles/vitals)).

## SOLID / design notes
- Follow the folder layout in `DESIGN.md:212-221`. It conflicts with `docs/DEV.md:23` (see References):
  - `features/<domain>` owns its routes, queries and UI.
  - `components/` holds compositions reused across domains.
  - `design-system/` is the only place that imports Astryx theme internals and token mapping.
- SRP: keep data hooks (`useQuestion`), presentational components (props in, JSX out) and route glue separate.
- Use composition through children and slots instead of boolean-prop explosions.
- Extract a product component only on real reuse (`DESIGN.md:224`).
- DIP: components depend on hooks, hooks depend on the generated client, and nothing below `features/` calls `fetch` directly.

## TDD & tests for this layer
- Work RED first: a failing Vitest/Testing Library test or Playwright scenario traced to a `US-*` story (`docs/TEST.md:24-31`, `docs/RULES.md:18`).
- Query by role and label first. Use `getByTestId` only as a last resort ([TL query priority](https://testing-library.com/docs/queries/about), [guiding principles](https://testing-library.com/docs/guiding-principles)).
- Snapshots are never the only proof (`docs/DEV.md:55`).
- For each component, cover:
  - keyboard and focus behaviour
  - loading, empty, error and success states
  - FR and EN strings
  - problem+json error mapping
  - the CSRF header present on mutations
  - the double-submit guard
- Mock at the network boundary with fixtures that conform to the contract, never by hand-mocking internals.
- Playwright P0 flows: Solve question detail, Discover API detail and Build contribution request. Run each in light and dark at 360 and 1440px, plus a throttled-network run (`DESIGN.md:268`; slice block S7 in `TODO.md`).
- Automated accessibility checks run per page with axe, for example `@axe-core/playwright` (candidate, pin per ADR). Keyboard and screen-reader review stays manual (`docs/TEST.md:21`, `DESIGN.md:242`).
- Coverage must be strictly above 70% for statements, branches, functions and lines, with thresholds set at ≥71. Exclusions are allowed only through `docs/quality/COVERAGE_SCOPE.md` (`docs/TEST.md:7-8`, `docs/quality/COVERAGE_SCOPE.md:5`).
- Never edit a test just to make it pass (`docs/TEST.md:32`).

## Anti-patterns (reject in review)
- Astryx components, props or tokens written from memory. A parallel component kit. Vendor code patched or swizzled without a gap record.
- Hex literals, `!important` or arbitrary `z-index` in feature code. Dark mode as a separate stylesheet. Lime or orange text contrast violations.
- Server data copied into `useState` or a global store. `useEffect` fetching. Non-deterministic query keys.
- Hand-written DTOs or `fetch` calls outside the client layer. Invented endpoints. Session or CSRF tokens in `localStorage` or URLs.
- `dangerouslySetInnerHTML` on unsanitized content. Machine-translating user content without disclosure.
- Hard-coded UI strings. A locale switch that resets forms. Placeholders used as labels. A `div` with `onClick` where a button belongs.
- Fake counters, mentors, jobs or "AI" answers. Dead nav links.
- Coverage exclusions or skipped tests used to reach a threshold. Claiming PASS for something that was NOT RUN.

## Done checklist
- [ ] Slice opened with `/start-slice`. Astryx pin and lockfile recorded. Every Astryx API used was checked against installed docs.
- [ ] Every DESIGN.md §8 box ticked (`DESIGN.md:233-246`), with evidence.
- [ ] RED evidence, then GREEN. Coverage strictly above 70% on all four metrics, with report attached.
- [ ] axe clean on the touched pages. Keyboard and screen-reader pass recorded. Light and dark checked. All 4 checkpoints and 200% zoom checked.
- [ ] FR/EN strings present and expansion checked. Original content preserved on locale switch.
- [ ] Mutations send CSRF and idempotency headers where required. problem+json is mapped. No contract drift.
- [ ] Gates run through the reviewed `make` targets. `/verify-slice` reports PASS, FAIL, NOT RUN or BLOCKED.

## References
- Project: `DESIGN.md`, `docs/DEV.md`, `docs/RULES.md`, `docs/TEST.md`, `docs/SPEC.md`, `docs/spec-kit/{API_CONTRACT,AUTHORIZATION_MATRIX,ASTRYX_VERIFICATION}.md`, archived `docs/archive/spec-kit/{CONTRACTS,G0_DECISIONS}.md`, the slice block in `TODO.md`.
- Tensions to escalate rather than resolve silently:
  1. `src/web/src` layout: `app/`, `design-system/` and `styles/` (`DESIGN.md:212-221`) vs `lib/` (`docs/DEV.md:23`).
  2. StyleX: "no compiler required" (`DESIGN.md:85`) vs "verify" (`docs/spec-kit/ASTRYX_VERIFICATION.md:21`).
  3. Astryx CLI commands: given as workflow steps (`DESIGN.md:129`) vs "do not assume they exist" (`docs/spec-kit/ASTRYX_VERIFICATION.md:19`).
  4. React 19 form Actions vs mandated TanStack Form (`DESIGN.md:225`). Use TanStack Form.
- Web: [react.dev](https://react.dev/blog/2024/12/05/react-19), [TanStack Query](https://tanstack.com/query/latest/docs/framework/react/guides/query-keys), [TanStack Router](https://tanstack.com/router/latest/docs/framework/react/guide/external-data-loading), [TanStack Form](https://tanstack.com/form/latest/docs/framework/react/guides/validation), [Vite env](https://vite.dev/guide/env-and-mode), [WCAG 2.2](https://www.w3.org/TR/WCAG22/), [APG](https://www.w3.org/WAI/ARIA/apg/patterns/), [OWASP CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html), [Testing Library](https://testing-library.com/docs/queries/about), [Astryx (upstream)](https://github.com/meta/astryx).
