import { linkTypes, isValidLinkId, type LinkType } from './validate';

/** Solve filters kept in the URL (shareable, back-button friendly), plus an optional pre-filled link. */
export interface SolveSearch {
  q?: string;
  tag?: string;
  /** "article:<uuid>": the content a question is about, set by the page that links here. */
  link?: string;
}

const TAG_RE = /^[a-z0-9][a-z0-9.+#-]{0,31}$/;

/** "article:<uuid>" becomes a typed link; anything else (type, id, shape) is dropped. */
export function parseLink(value: string | undefined): { type: LinkType; id: string } | undefined {
  const [type, id, ...rest] = (value ?? '').split(':');
  if (rest.length > 0 || type === undefined || id === undefined) return undefined;
  const known = linkTypes.find((candidate) => candidate === type);
  return known && isValidLinkId(id) ? { type: known, id: id.toLowerCase() } : undefined;
}

export function validateSolveSearch(raw: Record<string, unknown>): SolveSearch {
  const out: SolveSearch = {};
  if (typeof raw.q === 'string' && raw.q.trim()) out.q = raw.q.trim().slice(0, 200);
  if (typeof raw.tag === 'string' && TAG_RE.test(raw.tag)) out.tag = raw.tag;
  if (typeof raw.link === 'string' && parseLink(raw.link)) out.link = raw.link.toLowerCase();
  return out;
}
