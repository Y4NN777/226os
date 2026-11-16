import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// 127.0.0.1, not localhost: Node 24 resolves localhost to ::1 first and the API binds IPv4.
const apiTarget = 'http://127.0.0.1:8080';

export default defineConfig({
  plugins: [react()],
  server: {
    // The API's CSRF check compares Origin with Host (ADR-0003), so the proxy must pass the
    // browser's Host through unchanged; otherwise every sign-in from the dev server is a 403.
    proxy: {
      '/api': { target: apiTarget, changeOrigin: false },
      '/healthz': { target: apiTarget, changeOrigin: false },
    },
  },
  test: {
    environment: 'jsdom',
    // Owner decision 2026-11-14: jsdom page tests passed alone but hit the 5 s default when the
    // machine also ran integration suites; 15 s keeps a hung test visible without flaking.
    testTimeout: 15_000,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
    css: false,
    coverage: {
      provider: 'v8',
      // Vitest 5 has no `all` option: `include` defines the measured universe,
      // so untested files under src still count as 0%.
      include: ['src/**/*.{ts,tsx}'],
      // Only test files and the type-only generated declarations are excluded.
      exclude: ['src/**/*.test.{ts,tsx}', 'src/**/*.d.ts'],
      reporter: ['text', 'html', 'json-summary'],
      thresholds: { statements: 71, branches: 71, functions: 71, lines: 71 },
    },
  },
});
