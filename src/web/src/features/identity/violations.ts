import { ApiError } from '../../api/http';

/**
 * Violation fields reach the client as JSON pointers ("/username", from schema validation) or as
 * bare names ("new_password", from service rules). Both are reduced to the bare name.
 */
export function violationName(field: string): string {
  return field.replace(/^\//, '');
}

export function violationsByName(error: ApiError): Map<string, string> {
  const result = new Map<string, string>();
  for (const violation of error.violations) {
    const name = violationName(violation.field);
    if (!result.has(name)) result.set(name, violation.code);
  }
  return result;
}

/** The code is wrong, expired, used, or malformed: INVALID_TOKEN, or a violation on `token`. */
export function isRefusedToken(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    (error.code === 'INVALID_TOKEN' || violationsByName(error).has('token'))
  );
}
