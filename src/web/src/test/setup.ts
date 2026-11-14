import '@testing-library/jest-dom/vitest';
import { cleanup, configure } from '@testing-library/react';
import { afterEach, beforeEach, vi } from 'vitest';

// Whole-app renders (router, Astryx shell, providers) can take over a second on a loaded
// CI runner; findBy* waits longer instead of failing on load. Assertions are unchanged.
configure({ asyncUtilTimeout: 4000 });

// Load the whole app once, before the first test starts. renderApp() imports it lazily, and on a
// loaded machine that transform and evaluation used to land inside the first test of a file and
// eat its 15 s budget. Top-level await in a setup file is not counted against any test.
await Promise.all([import('../app/App'), import('../app/router')]);

beforeEach(() => {
  // jsdom does not implement scrolling; the router restores scroll on navigation.
  vi.stubGlobal('scrollTo', vi.fn());
  // jsdom has no matchMedia; Astryx and the "system" theme mode read it.
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  );
});

afterEach(() => {
  cleanup();
  localStorage.clear();
  document.documentElement.removeAttribute('lang');
  document.documentElement.removeAttribute('data-theme');
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

afterEach(async () => {
  // The CSRF token lives in memory for the app's lifetime; tests must not share it.
  const { api } = await import('../api/http');
  api.csrf.set(null);
});
