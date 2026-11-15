import { useEffect, useState } from 'react';

function readToken(): string | null {
  return new URLSearchParams(window.location.hash.replace(/^#/, '')).get('token');
}

function clearHash(): void {
  window.history.replaceState(
    window.history.state,
    '',
    window.location.pathname + window.location.search,
  );
}

/**
 * A mail link may carry the token in the fragment (`#token=...`). It is read on the first
 * render and removed from the address bar at once, so it leaves no trace in history or
 * Referer. A link opened in a tab already on the page only changes the fragment (no reload),
 * so a `hashchange` listener picks the new token up the same way. The read is pure (safe to
 * repeat); only the removal is an effect.
 */
export function useHashToken(): string | null {
  const [token, setToken] = useState(readToken);
  useEffect(() => {
    if (readToken() !== null) clearHash();
    const onHashChange = () => {
      const next = readToken();
      if (next === null) return;
      setToken(next);
      clearHash();
    };
    window.addEventListener('hashchange', onHashChange);
    return () => {
      window.removeEventListener('hashchange', onHashChange);
    };
  }, []);
  return token;
}
