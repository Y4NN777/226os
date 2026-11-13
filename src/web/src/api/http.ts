import createClient, { type Middleware } from 'openapi-fetch';
import type { components, paths } from './schema';

type ErrorCode = components['schemas']['ErrorCode'];
type Violation = components['schemas']['Violation'];

/** Contract codes plus two client-side outcomes: no response at all, or a body that is not problem+json. */
export type ApiErrorCode = ErrorCode | 'NETWORK' | 'UNKNOWN';

/** One error type for every API failure; pages branch on `code` (ADR-0004), never on `detail`. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: ApiErrorCode;
  readonly requestId: string | undefined;
  readonly violations: readonly Violation[];
  /** Seconds, from Retry-After on 429 and 503. */
  readonly retryAfter: number | undefined;

  constructor(init: {
    status: number;
    code: ApiErrorCode;
    requestId?: string | undefined;
    violations?: readonly Violation[] | undefined;
    retryAfter?: number | undefined;
  }) {
    super(init.code);
    this.name = 'ApiError';
    this.status = init.status;
    this.code = init.code;
    this.requestId = init.requestId;
    this.violations = init.violations ?? [];
    this.retryAfter = init.retryAfter;
  }
}

const UNSAFE = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

/** A hostile or buggy Retry-After must not lock a form for hours. */
const MAX_RETRY_AFTER_SECONDS = 300;

function parseRetryAfter(response: Response): number | undefined {
  const raw = response.headers.get('Retry-After');
  if (raw === null) return undefined;
  const seconds = Number(raw);
  if (!Number.isFinite(seconds) || seconds < 0) return undefined;
  return Math.min(Math.ceil(seconds), MAX_RETRY_AFTER_SECONDS);
}

function toApiError(body: unknown, response: Response): ApiError {
  const retryAfter = parseRetryAfter(response);
  // The API always sends request_id; an edge or proxy problem may not, but its code still counts.
  if (isRecord(body) && typeof body.code === 'string') {
    return new ApiError({
      status: response.status,
      code: body.code as ErrorCode,
      requestId: typeof body.request_id === 'string' ? body.request_id : undefined,
      violations: Array.isArray(body.violations) ? (body.violations as Violation[]) : [],
      retryAfter,
    });
  }
  const code = response.status === 429 ? 'RATE_LIMITED' : 'UNKNOWN';
  return new ApiError({ status: response.status, code, retryAfter });
}

/**
 * Synchronizer CSRF token (ADR-0003), kept in memory only: never in storage, a URL or a log.
 * It comes from GET /auth/session, the sign-in response, or GET /auth/csrf when missing.
 */
export interface CsrfStore {
  set: (token: string | null) => void;
  get: () => Promise<string>;
}

function createCsrfStore(fetchImpl: typeof fetch, baseUrl: string): CsrfStore {
  let token: string | null = null;
  let pending: Promise<string> | null = null;
  // Bumped on every set(): a fetch that started earlier must not overwrite a newer token.
  let generation = 0;
  async function fetchToken(): Promise<string> {
    const response = await fetchImpl(
      new Request(`${baseUrl}/auth/csrf`, { credentials: 'include' }),
    );
    const body: unknown = await response.json().catch(() => null);
    if (!response.ok || !isRecord(body) || typeof body.csrf_token !== 'string') {
      throw toApiError(body, response);
    }
    return body.csrf_token;
  }
  return {
    set: (next) => {
      generation += 1;
      token = next;
    },
    get: () => {
      if (token !== null) return Promise.resolve(token);
      const started = generation;
      pending ??= fetchToken()
        .then((fresh) => {
          if (generation === started) token = fresh;
          return token ?? fresh;
        })
        .finally(() => {
          pending = null;
        });
      return pending;
    },
  };
}

interface FetchResult<T> {
  data?: T;
  error?: unknown;
  response: Response;
}

/**
 * The single API helper: a typed openapi-fetch client whose unsafe calls always carry
 * X-CSRF-Token, and `call`, which turns every failure into an ApiError and retries a
 * mutation once with a fresh token after 403 CSRF_FAILED (expired or rotated token).
 */
export function createApi(fetchImpl?: typeof fetch) {
  // Late-bound global fetch, so tests can stub it after this module is imported.
  const doFetch: typeof fetch = fetchImpl ?? ((input, init) => globalThis.fetch(input, init));
  const baseUrl = new URL('/api/v1', globalThis.location.origin).href;
  const csrf = createCsrfStore(doFetch, baseUrl);
  const client = createClient<paths>({ baseUrl, credentials: 'include', fetch: doFetch });

  const csrfMiddleware: Middleware = {
    async onRequest({ request }) {
      if (!UNSAFE.has(request.method)) return undefined;
      request.headers.set('X-CSRF-Token', await csrf.get());
      return request;
    },
  };
  client.use(csrfMiddleware);

  async function attempt<T>(run: () => Promise<FetchResult<T>>): Promise<T> {
    let result: FetchResult<T>;
    try {
      result = await run();
    } catch (error) {
      if (error instanceof ApiError) throw error;
      throw new ApiError({ status: 0, code: 'NETWORK' });
    }
    if (result.response.ok) return result.data as T;
    throw toApiError(result.error, result.response);
  }

  async function call<T>(run: () => Promise<FetchResult<T>>): Promise<T> {
    try {
      return await attempt(run);
    } catch (error) {
      if (error instanceof ApiError && error.code === 'CSRF_FAILED') {
        csrf.set(null);
        return attempt(run);
      }
      throw error;
    }
  }

  return { client, csrf, call };
}

export type Api = ReturnType<typeof createApi>;

/** The app-wide instance. Features import this, never fetch directly. */
export const api = createApi();
