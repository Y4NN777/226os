import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { I18nProvider } from '../i18n/I18nProvider';
import { ThemePreferenceProvider } from '../theme/ThemeProvider';

export function renderWithProviders(ui: ReactElement) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <I18nProvider>
        <ThemePreferenceProvider>{ui}</ThemePreferenceProvider>
      </I18nProvider>
    </QueryClientProvider>,
  );
}

export function setLanguages(languages: readonly string[]) {
  Object.defineProperty(window.navigator, 'languages', { value: languages, configurable: true });
}

export function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  });
}

/** The whole app (shell, router, providers) at `path`, on an in-memory history. */
export async function renderApp(path: string) {
  const [{ createMemoryHistory }, { App }, { createAppRouter }] = await Promise.all([
    import('@tanstack/react-router'),
    import('../app/App'),
    import('../app/router'),
  ]);
  const router = createAppRouter({ history: createMemoryHistory({ initialEntries: [path] }) });
  const view = render(<App router={router} />);
  return { ...view, router };
}
