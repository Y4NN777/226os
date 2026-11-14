import { useEffect } from 'react';

/** Keep WCAG 2.4.2 page titles synchronized with the localized visible page name. */
export function usePageTitle(page: string) {
  useEffect(() => {
    document.title = `${page} · 226os`;
  }, [page]);
}
