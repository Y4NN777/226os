import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// 127.0.0.1, not localhost: Node 24 resolves localhost to ::1 first and the API binds IPv4.
const apiTarget = 'http://127.0.0.1:8080';

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': apiTarget,
      '/healthz': apiTarget,
    },
  },
  test: {
    environment: 'jsdom',
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
