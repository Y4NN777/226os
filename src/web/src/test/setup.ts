import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach, beforeEach, vi } from 'vitest';

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
