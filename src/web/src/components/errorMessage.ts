import { ApiError } from '../api/http';
import type { MessageKey } from '../i18n/messages';
import type { Params } from '../i18n/translate';

const known = {
  NETWORK: 'error.NETWORK',
  NOT_FOUND: 'error.NOT_FOUND',
  UNAUTHENTICATED: 'error.UNAUTHENTICATED',
  FORBIDDEN: 'error.FORBIDDEN',
  EMAIL_NOT_VERIFIED: 'error.EMAIL_NOT_VERIFIED',
  CSRF_FAILED: 'error.CSRF_FAILED',
  VALIDATION_ERROR: 'error.VALIDATION_ERROR',
  SERVICE_UNAVAILABLE: 'error.SERVICE_UNAVAILABLE',
} as const satisfies Partial<Record<string, MessageKey>>;

/**
 * Localized, user-safe message for any failure: branches on the stable problem code
 * (ADR-0004), never shows the server `detail`.
 */
export function errorMessage(error: unknown): { key: MessageKey; params?: Params } {
  if (!(error instanceof ApiError)) return { key: 'error.generic' };
  if (error.code === 'RATE_LIMITED') {
    return error.retryAfter
      ? { key: 'error.RATE_LIMITED', params: { seconds: error.retryAfter } }
      : { key: 'error.RATE_LIMITED.now' };
  }
  const key = (known as Partial<Record<string, MessageKey>>)[error.code];
  return { key: key ?? 'error.generic' };
}
