---
name: bdd-acceptance
description: Acceptance tests for 226os. Given/When/Then is taken verbatim from docs/SPEC.md, traced to US-* IDs, and run as Playwright E2E with FR/EN, mobile, slow network, accessibility and negative cases (G5). Use when touching tests/e2e/**, e2e/**, **/*.feature, apps/web/e2e/**.
---
# BDD Acceptance and E2E (G5)

## Scope and triggers
- Executable acceptance scenarios: `**/*.feature`, `tests/e2e/**`, `e2e/**`, `apps/web/e2e/**`, plus the Playwright config and fixtures.
- Any P0 story change (`US-SOLVE-001`, `US-DISCOVER-001`, `US-BUILD-001`, `US-IDENTITY-001`, `US-ACCESS-001`).
- HTTP-level acceptance against the API belongs to `integration-testing`. This skill owns the traceability and the browser journeys.

## Project rules (binding)
- BDD comes before code. Every P0 story has a `.feature` file or a traceable Given/When/Then in `SPEC.md` before implementation. `docs/TEST.md:24`, `docs/RULES.md:18`, `docs/DEV.md:35`.
- Scenarios describe observable outcomes, not internals, and are executable or traced to HTTP/UI tests. `docs/TEST.md:25`.
- Every journey covers: the nominal path, invalid data, an unauthorised actor, duplication/idempotency, DB/network errors, and FR/EN plus mobile where relevant. `docs/TEST.md:26`.
- The source scenarios are `docs/SPEC.md:13-29` (SOLVE), `:31-40` (DISCOVER), `:42-56` (BUILD), `:58-71` (IDENTITY) and `:73-81` (ACCESS). The traceability map is at `docs/SPEC.md:130-137`.
- Required cases per journey A-F, including 360px mobile, slow network and text preservation, are listed at `docs/spec-kit/TRACEABILITY.md:29-38`. No task goes GREEN without a stable scope ID, an acceptance example, a contract, a test-first plan and an evidence path (`:42`).
- G5 evidence is Playwright P0 traces, FR/EN, mobile, keyboard, axe/contrast, and a manual review. `docs/spec-kit/QUALITY_RELEASE.md:14`, `docs/TEST.md:48`, `AGENTS.md:51`.
- Responsive checkpoints are 360/768/1024/1440px and 200% zoom, with no horizontal document scroll. `DESIGN.md:188`. Accessibility, reduced motion and touch targets are mandatory. `DESIGN.md:26`, `DESIGN.md:153`.
- Frontend coverage needs companion evidence: P0 E2E, keyboard/a11y, locale and throttled-network tests. `docs/quality/COVERAGE_SCOPE.md:17`.
- Seeds and fixtures are deterministic, clearly synthetic, and never suggest real users, endorsements or verification. `docs/spec-kit/CONTRACTS.md:25`. No secrets or private data in fixtures, traces or screenshots. `docs/RULES.md:25`, `docs/spec-kit/QUALITY_RELEASE.md:41`.
- The CADEV exit requires the flows to be tested on the real deployed application. `docs/TEST.md:58`.
- The Make target list in `AGENTS.md:120` has no E2E target. Any `make test-e2e` is PROPOSED and needs review and approval before use (`AGENTS.md:124`).

## Expert practices
- Copy SPEC step text verbatim, in French, with the English `Given/When/Then/And` keywords exactly as SPEC writes them. Do not translate and do not add `# language: fr`, since the keywords in SPEC are English. [Gherkin reference](https://cucumber.io/docs/gherkin/reference/)
- Add new negative scenarios required by `TEST.md:26` as separate `Scenario`s tagged `@derived`. They need reviewer sign-off and a PR to `SPEC.md` (`docs/SPEC.md:8`). Never edit the canonical text in a test file.
- Tag each scenario with `@US-…` and `@P0` (or `@P1`). Keep `Background` short, and use `Scenario Outline` + `Examples` for locale and viewport matrices. [Gherkin reference](https://cucumber.io/docs/gherkin/reference/)
- Isolation: each test owns its context, data and session. Reuse an authenticated state only through a setup project. [Playwright best practices](https://playwright.dev/docs/best-practices)
- Locators: use `getByRole`, `getByLabel` or `getByText` with accessible names. No CSS/XPath tied to DOM structure. [Playwright best practices](https://playwright.dev/docs/best-practices)
- Assertions: use web-first `await expect(locator).toBeVisible()`. Never `waitForTimeout` or manual `isVisible()` polling. [Playwright best practices](https://playwright.dev/docs/best-practices)
- CI: set `trace: 'on-first-retry'`, not always on. Shard large suites. Upload traces as artifacts named by SHA. [Playwright best practices](https://playwright.dev/docs/best-practices)
- Locale and mobile: use the `locale` / `viewport` / `devices[...]` projects and `offline: true` for the offline state. [Playwright emulation](https://playwright.dev/docs/emulation)
- Slow network: Playwright has no native bandwidth throttling. Use Chromium CDP `Network.emulateNetworkConditions` (Chromium-only) or a throttling proxy. UNVERIFIED: the method needs an ADR before G6.
- Run axe through `@axe-core/playwright` with the WCAG A/AA tags and assert `violations` equals `[]`. Automated checks miss many issues, so add keyboard-only runs and a manual review. [Playwright accessibility testing](https://playwright.dev/docs/accessibility-testing)
- Seed data through the API or a test-only seed endpoint/DB fixture per worker, keyed by a run ID. Never depend on another test's leftovers or on execution order.
- Negative UI cases must also assert the server state. For example, after a forbidden accept, re-fetch through the API and confirm nothing changed (`docs/SPEC.md:28`).

## Patterns
```gherkin
# tests/e2e/features/solve.feature
@US-SOLVE-001 @P0
Feature: Entraide technique
  Scenario: Interdire une acceptation non autorisée
    Given Bob n'est pas l'autrice de la question
    When Bob tente d'accepter une réponse via l'API
    Then le serveur refuse avec 403
    And aucune réponse acceptée n'est modifiée

  @derived @needs-spec-review
  Scenario: Double vote refusé
    ...
```
```ts
// playwright.config.ts (excerpt) — FR/EN × desktop/mobile matrix
projects: [
  { name: 'fr-desktop', use: { locale: 'fr-FR', viewport: { width: 1440, height: 900 } } },
  { name: 'en-mobile',  use: { ...devices['Pixel 5'], locale: 'en-US' } }, // device preset choice UNVERIFIED
  { name: 'fr-360',     use: { locale: 'fr-FR', viewport: { width: 360, height: 740 } } },
],
use: { trace: 'on-first-retry', baseURL: process.env.E2E_BASE_URL },
```
```ts
test('@US-IDENTITY-001 switching to English keeps original content', async ({ page, seed }) => {
  const q = await seed.question({ title: 'Erreur CORS sur Go' }); // synthetic, per-test
  await page.goto(`/questions/${q.id}`);
  await page.getByRole('button', { name: /langue|language/i }).click();
  await page.getByRole('menuitem', { name: 'English' }).click();
  await expect(page.getByRole('navigation')).toContainText('Explore'); // label key UNVERIFIED
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Erreur CORS sur Go');
});
```
```ts
const { violations } = await new AxeBuilder({ page })
  .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']).analyze();
expect(violations).toEqual([]);
```

## Anti-patterns (reject in review)
- Paraphrased or translated SPEC steps. New behaviour hidden inside an existing scenario.
- Scenarios without a `@US-*` tag, or a story with no negative scenario.
- `page.waitForTimeout`, `sleep`, retries used to mask flakiness, or `test.fixme`/`skip` to get green.
- Locators like `.btn-primary`, `div > span:nth-child(2)`, or positional XPath.
- Shared accounts or data across tests, order-dependent suites, production or real personal data.
- Asserting only on UI text for an authorisation rule without checking server state.
- Checking only desktop/FR. Missing the 360px, EN, keyboard, axe or slow-network runs.
- Traces or screenshots that capture cookies, tokens or real emails.
- Claiming G5 from local runs against a dev server when the gate needs the deployed staging SHA.

## Done checklist
- [ ] Every scenario in the touched story is executable, with SPEC text unchanged and a `@US-*` tag.
- [ ] Negative cases from `TEST.md:26` and `TRACEABILITY.md:33-38` are present or explicitly `OPEN` with an owner.
- [ ] FR + EN, 360px + desktop, keyboard-only, axe A/AA and slow network have each run.
- [ ] Command (reviewed target), exit code, and target URL/SHA recorded. Traces/HTML report stored under `docs/evidence/<TASK-ID>/` or a CI artifact named by SHA.
- [ ] Manual a11y review noted, with reviewer name (axe alone is not G5).
- [ ] No secrets or personal data in artifacts.
- [ ] Gate status reported as `PASS / FAIL / NOT RUN / BLOCKED`. Unexecuted means `NOT RUN` (`AGENTS.md:31`).

## References
- `docs/SPEC.md`, `docs/TEST.md`, `docs/spec-kit/TRACEABILITY.md`, `docs/spec-kit/QUALITY_RELEASE.md`, `DESIGN.md`
- https://cucumber.io/docs/gherkin/reference/
- https://playwright.dev/docs/best-practices · https://playwright.dev/docs/emulation · https://playwright.dev/docs/accessibility-testing
