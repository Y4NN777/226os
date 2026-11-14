/** Connect filters kept in the URL (shareable, back-button friendly). Mirrors the listProfiles query. */
export interface ConnectSearch {
  country?: string;
  language?: string;
  skill?: string;
}

const COUNTRY_RE = /^[A-Z]{2}$/;
const LANGUAGE_RE = /^[a-z]{2,3}$/;
const SKILL_RE = /^[a-z0-9][a-z0-9.+#-]{0,31}$/;

/** Turns what a person typed into a valid tag, or undefined when it cannot be one. */
export function toTag(value: string, pattern: RegExp): string | undefined {
  const tag = value.trim().toLowerCase();
  return pattern.test(tag) ? tag : undefined;
}

export const languageTag = (value: string) => toTag(value, LANGUAGE_RE);
export const skillTag = (value: string) => toTag(value, SKILL_RE);

/** Keeps only values the API would accept; anything else is dropped. */
export function validateConnectSearch(raw: Record<string, unknown>): ConnectSearch {
  const out: ConnectSearch = {};
  if (typeof raw.country === 'string' && COUNTRY_RE.test(raw.country)) out.country = raw.country;
  if (typeof raw.language === 'string' && LANGUAGE_RE.test(raw.language)) {
    out.language = raw.language;
  }
  if (typeof raw.skill === 'string' && SKILL_RE.test(raw.skill)) out.skill = raw.skill;
  return out;
}
