// Mirrors the `github` pattern of the API contract.
const GITHUB_LOGIN = /^[A-Za-z0-9-]{1,39}$/;

/** The GitHub profile URL for a valid login, or null so callers render no link. */
export function githubProfileUrl(login: string): string | null {
  return GITHUB_LOGIN.test(login) ? `https://github.com/${encodeURIComponent(login)}` : null;
}
