/**
 * True only for a same-origin path such as `/discover?q=x`: it must start with one
 * slash and contain no backslash (raw or %5C), whitespace or control character, so
 * neither the browser nor a router can turn it into another origin (`/\t/evil`, `//evil`).
 */
export function isSafeInternalPath(value: string): boolean {
  if (!value.startsWith('/') || value.startsWith('//')) return false;
  // eslint-disable-next-line no-control-regex -- the point is to reject control characters
  if (/[\u0000- \u007f\\]/.test(value) || /%5c/i.test(value)) return false;
  return new URL(value, 'http://origin.invalid').origin === 'http://origin.invalid';
}
