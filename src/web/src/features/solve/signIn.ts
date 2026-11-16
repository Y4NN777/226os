import { isSafeInternalPath } from '../../components/safePath';

/** Sign-in address that brings the visitor back to `path`; only a safe internal path is kept. */
export function signInHref(path: string): string {
  return isSafeInternalPath(path) ? `/connexion?next=${encodeURIComponent(path)}` : '/connexion';
}
