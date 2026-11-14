/** Solve filters kept in the URL (shareable, back-button friendly). */
export interface SolveSearch {
  q?: string;
  tag?: string;
}

const TAG_RE = /^[a-z0-9][a-z0-9.+#-]{0,31}$/;

export function validateSolveSearch(raw: Record<string, unknown>): SolveSearch {
  const out: SolveSearch = {};
  if (typeof raw.q === 'string' && raw.q.trim()) out.q = raw.q.trim().slice(0, 200);
  if (typeof raw.tag === 'string' && TAG_RE.test(raw.tag)) out.tag = raw.tag;
  return out;
}
