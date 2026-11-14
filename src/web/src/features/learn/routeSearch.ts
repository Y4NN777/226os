/**
 * Link kinds an article can carry. The API also names `article` (and once `series`), but it answers
 * 422 for those pairs for now, so they are never offered (issue #124).
 */
export const linkTypes = ['question', 'resource', 'project'] as const;
export type LinkType = (typeof linkTypes)[number];

export interface LinkDraft {
  type: LinkType;
  id: string;
}

export interface WriteSearch {
  /** "question:<uuid>": the content the article is about, set by the page that links here. */
  link?: string;
}

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/** "question:<uuid>" becomes a typed link; anything else (kind, id, shape) is dropped. */
export function parseLink(value: string | undefined): LinkDraft | undefined {
  const [type, id, ...rest] = (value ?? '').split(':');
  if (rest.length > 0 || id === undefined) return undefined;
  const known = linkTypes.find((candidate) => candidate === type);
  return known && UUID_RE.test(id) ? { type: known, id: id.toLowerCase() } : undefined;
}

export function validateWriteSearch(raw: Record<string, unknown>): WriteSearch {
  return typeof raw.link === 'string' && parseLink(raw.link)
    ? { link: raw.link.toLowerCase() }
    : {};
}
