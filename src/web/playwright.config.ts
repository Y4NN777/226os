import { defineConfig } from '@playwright/test';

// End-to-end suite run against a deployed site (see docs/TEST.md, section 6 ter).
// It is never part of `pnpm test` (vitest only includes src/**/*.test.*).
const baseURL = process.env['E2E_BASE_URL'];
if (!baseURL) {
  throw new Error(
    'E2E_BASE_URL is required, for example E2E_BASE_URL=https://staging.example pnpm --filter web test:e2e',
  );
}

const isCI = Boolean(process.env['CI']);

export default defineConfig({
  testDir: './e2e',
  // Writes on a shared staging are unique per run; the suite itself is order independent.
  fullyParallel: true,
  forbidOnly: isCI,
  retries: isCI ? 1 : 0,
  // CI (a public repository uploads the report): no HTML report, trace, video or screenshot. A trace records
  // fill(password), the sign-in body and Set-Cookie. Only the junit XML (names, durations, messages) is uploaded.
  reporter: isCI
    ? [['list'], ['junit', { outputFile: 'test-results/junit.xml' }]]
    : [['list'], ['html', { outputFolder: 'playwright-report', open: 'never' }]],
  outputDir: 'test-results',
  use: {
    baseURL,
    locale: 'fr-FR',
    trace: isCI ? 'off' : 'on-first-retry',
    video: 'off',
    screenshot: 'off',
  },
  projects: [
    {
      name: 'desktop',
      testIgnore: /mobile\.spec\.ts/,
      use: { browserName: 'chromium', viewport: { width: 1280, height: 800 } },
    },
    {
      name: 'mobile',
      testMatch: /mobile\.spec\.ts/,
      use: {
        browserName: 'chromium',
        viewport: { width: 375, height: 667 },
        deviceScaleFactor: 2,
        isMobile: true,
        hasTouch: true,
      },
    },
  ],
});
