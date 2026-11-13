import { useEffect, useState } from 'react';

function readToken(): string | null {
  return new URLSearchParams(window.location.hash.replace(/^#/, '')).get('token');
}

/**
 * A future mail link may carry the token in the fragment (`#token=...`). It is read on the
 * first render and removed from the address bar at once, so it leaves no trace in history
 * or Referer. The read is pure (safe to repeat); only the removal is an effect.
 */
export function useHashToken(): string | null {
  const [token] = useState(readToken);
  useEffect(() => {
    if (readToken() === null) return;
    window.history.replaceState(
      window.history.state,
      '',
      window.location.pathname + window.location.search,
    );
  }, []);
  return token;
}
