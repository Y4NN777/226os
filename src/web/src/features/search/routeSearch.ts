export interface SearchSearch {
  q?: string;
}

export function validateSearchSearch(raw: Record<string, unknown>): SearchSearch {
  return typeof raw.q === 'string' && raw.q.trim() ? { q: raw.q.slice(0, 100) } : {};
}
