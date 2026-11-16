import { isSafeInternalPath } from '../../components/safePath';

export interface SignInSearch {
  /** Same-origin path to return to after sign-in. */
  next?: string;
}

/** Only a same-origin path is kept, so `next` can never become an open redirect. */
export function safeNext(value: unknown): string | undefined {
  return typeof value === 'string' && isSafeInternalPath(value) ? value : undefined;
}

export function validateSignInSearch(raw: Record<string, unknown>): SignInSearch {
  const next = safeNext(raw.next);
  return next ? { next } : {};
}
