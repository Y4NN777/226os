import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from '@tanstack/react-router';
import { useState } from 'react';
import { I18nProvider } from '../i18n/I18nProvider';
import { ThemePreferenceProvider } from '../theme/ThemeProvider';
import { createAppRouter, type AppRouter } from './router';

export function App({ router }: { router?: AppRouter }) {
  const [queryClient] = useState(() => new QueryClient());
  const [appRouter] = useState(() => router ?? createAppRouter());
  return (
    <QueryClientProvider client={queryClient}>
      <I18nProvider>
        <ThemePreferenceProvider>
          <RouterProvider router={appRouter} />
        </ThemePreferenceProvider>
      </I18nProvider>
    </QueryClientProvider>
  );
}
