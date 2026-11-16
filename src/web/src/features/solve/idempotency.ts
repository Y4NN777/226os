import { useCallback, useRef } from 'react';

/**
 * One Idempotency-Key per submission. A retry of the same content reuses the key (the server
 * then replays instead of duplicating); changed content gets a new one, since the server
 * refuses a key reused with a different body (422 IDEMPOTENCY_KEY_REUSED).
 */
export function useIdempotencyKey() {
  const current = useRef<{ fingerprint: string; key: string } | null>(null);
  const keyFor = useCallback((payload: unknown) => {
    const fingerprint = JSON.stringify(payload);
    if (current.current?.fingerprint !== fingerprint) {
      current.current = { fingerprint, key: crypto.randomUUID() };
    }
    return current.current.key;
  }, []);
  /** After a success: the next submission is a new one. */
  const settle = useCallback(() => {
    current.current = null;
  }, []);
  return { keyFor, settle };
}
